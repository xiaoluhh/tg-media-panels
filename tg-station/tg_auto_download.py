#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
TG 私人中转站 v2 —— 自动下载 + 媒体展示 一体化
- 下载器：Telethon 监听转发消息 → SQLite 去重(原始频道ID+原始消息ID) → 并发下载
- 展示层：复用 media-browser（子进程，下载完成自动重建索引并重启，页面自动刷新）
- 登录：扫码(默认) / 手机号验证码；2FA 密码走 config.twofa_password 兜底
- 失败任务：记录 SQLite，后台按 5min→15min→30min→60min 退避自动重试，重启后续传
"""
import asyncio
import json
import os
import re
import shutil
import socket
import sqlite3
import subprocess
import sys
import threading
import time
import urllib.request
from datetime import datetime, timedelta

# pythonw 无窗口运行时 stdout 为 None，做兜底
if sys.stdout is None:
    sys.stdout = open(os.devnull, "w")

from telethon import TelegramClient, events
from telethon.errors import SessionPasswordNeededError
from telethon.tl.types import (
    MessageMediaPhoto, MessageMediaDocument,
    DocumentAttributeSticker, DocumentAttributeAudio,
    DocumentAttributeVideo, DocumentAttributeAnimated, DocumentAttributeFilename,
)

VIDEO_EXTS = {".mp4", ".mkv", ".webm", ".mov", ".avi", ".ts", ".m4v", ".flv", ".wmv", ".mpg", ".mpeg", ".3gp"}
IMAGE_EXTS = {".jpg", ".jpeg", ".png", ".gif", ".webp"}
# 媒体页面（tg_station）按扩展名白名单收录文件；名字里没有扩展名的文件会"下好了但页面看不见"
MEDIA_EXTS = IMAGE_EXTS | VIDEO_EXTS


def sniff_media_ext(path, media=None):
    """按文件头判断真实类型，返回扩展名（媒体页面兜底补名用）"""
    try:
        with open(path, "rb") as f:
            head = f.read(16)
    except Exception:
        head = b""
    if head.startswith(b"\xff\xd8\xff"):
        return ".jpg"
    if head.startswith(b"\x89PNG"):
        return ".png"
    if head.startswith(b"GIF8"):
        return ".gif"
    if head[:4] == b"RIFF" and head[8:12] == b"WEBP":
        return ".webp"
    if b"ftyp" in head[:12]:
        return ".mp4"
    if head[:4] == b"\x1a\x45\xdf\xa3":
        return ".mkv"
    if head[:4] == b"RIFF" and head[8:12] == b"AVI ":
        return ".avi"
    if isinstance(media, MessageMediaPhoto):
        return ".jpg"
    return None
RETRY_INTERVALS = [5, 15, 30, 60]  # 分钟，失败重试退避；超过则封顶 60 分钟


def app_dir():
    """运行数据目录：打包后 = exe 所在目录；开发 = 脚本目录"""
    if getattr(sys, "frozen", False):
        return os.path.dirname(sys.executable)
    return os.path.dirname(os.path.abspath(__file__))


def resource_dir():
    """内置资源目录（PyInstaller 打包的资源）：onefile/_internal；开发 = 脚本目录"""
    if getattr(sys, "frozen", False):
        return getattr(sys, "_MEIPASS", os.path.dirname(sys.executable))
    return app_dir()


BASE_DIR = app_dir()
# 整合模式（tg_station.exe 启动本脚本时）通过环境变量指定与 Go 后端共享的配置文件；
# 独立运行则用脚本目录下的 config.json。
CONFIG_FILE = os.environ.get("TG_STATION_CONFIG") or os.path.join(BASE_DIR, "config.json")
DB_FILE = os.path.join(BASE_DIR, "downloads.db")
LOG_FILE = os.path.join(BASE_DIR, "tg_station.log")
STATUS_FILE = os.path.join(BASE_DIR, "dl_status.json")
PROGRESS_FILE = os.path.join(BASE_DIR, "progress.json")
LOGIN_FILE = os.path.join(BASE_DIR, "login_status.json")
PAUSE_FILE = os.path.join(BASE_DIR, "pause.json")


def log(msg):
    """统一日志：窗口打印 + 写文件"""
    line = f"[{datetime.now().strftime('%H:%M:%S')}] {msg}"
    try:
        print(line, flush=True)
    except UnicodeEncodeError:
        # Windows 管道 GBK 编码兜底：不支持的字符替换为 ?，避免日志崩溃
        print(line.encode("gbk", errors="replace").decode("gbk", errors="replace"), flush=True)
    try:
        with open(LOG_FILE, "a", encoding="utf-8") as f:
            f.write(line + "\n")
    except Exception:
        pass


# ---------------- 暂停控制（保留原语义） ----------------

def init_pause():
    if not os.path.exists(PAUSE_FILE):
        try:
            with open(PAUSE_FILE, "w", encoding="utf-8") as f:
                json.dump({"paused": False}, f)
        except Exception:
            pass


def is_paused():
    try:
        with open(PAUSE_FILE, encoding="utf-8") as f:
            return bool(json.load(f).get("paused", True))
    except Exception:
        return True


# ---------------- 代理 ----------------
# 代理统一走用户常开的 Clash Verge（127.0.0.1:7897）。
# 2026-10-09 起不再内置 mihomo 兜底：既不探测也不启动 mihomo（用户暂停使用）。
CLASH_PROXY_PORT = 7897
ACTIVE_PROXY_PORT = None  # ensure_proxy() 探测后写入，build_proxy 优先使用


def port_open(port, host="127.0.0.1", timeout=1):
    """TCP 端口连通性探测（代理 7897 / media-browser 8080 共用）"""
    s = socket.socket()
    s.settimeout(timeout)
    try:
        s.connect((host, port))
        return True
    except Exception:
        return False
    finally:
        s.close()


def ensure_proxy():
    """确保 Clash Verge 代理可用（127.0.0.1:7897）。

    不可用时只记日志提示，等用户在电脑上打开 Clash Verge。
    """
    global ACTIVE_PROXY_PORT
    if port_open(CLASH_PROXY_PORT):
        if ACTIVE_PROXY_PORT != CLASH_PROXY_PORT:
            log(f"[proxy] 使用 Clash Verge 代理 127.0.0.1:{CLASH_PROXY_PORT}")
        ACTIVE_PROXY_PORT = CLASH_PROXY_PORT
        return True
    ACTIVE_PROXY_PORT = None
    log("[!] Clash Verge 代理不可用：请在电脑上打开 Clash Verge（127.0.0.1:7897）")
    return False


def ensure_resources():
    """打包后首次运行：把内置的 config.json / mz.txt 复制到 exe 旁边"""
    src = resource_dir()
    if src == app_dir():
        return
    try:
        c = os.path.join(src, "config.json")
        if os.path.exists(c) and not os.path.exists(CONFIG_FILE):
            shutil.copy2(c, CONFIG_FILE)
        mz = os.path.join(src, "mz.txt")
        if os.path.exists(mz) and not os.path.exists(os.path.join(app_dir(), "mz.txt")):
            shutil.copy2(mz, os.path.join(app_dir(), "mz.txt"))
    except Exception as e:
        log(f"[!] 资源初始化失败: {e}")


# ---------------- 登录状态（文件，供外部查看） ----------------

def write_login_status(required, qr_path=None, note=None):
    try:
        with open(LOGIN_FILE, "w", encoding="utf-8") as f:
            json.dump({
                "login_required": bool(required),
                "qr": os.path.basename(qr_path) if qr_path else None,
                "note": note or "",
                "time": datetime.now().strftime("%H:%M:%S"),
            }, f, ensure_ascii=False)
    except Exception:
        pass


# ---------------- SQLite 去重与状态 ----------------

DB_SCHEMA = """
CREATE TABLE IF NOT EXISTS downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fwd_channel_id INTEGER NOT NULL,   -- 原始频道ID（ForwardOrigin）
    fwd_msg_id INTEGER NOT NULL,       -- 原始消息ID（ForwardOrigin）
    channel_id INTEGER,                -- 实际收到消息的频道ID
    msg_id INTEGER,                    -- 实际消息ID
    filename TEXT,                     -- 文件名（下载完成后更新）
    size INTEGER,                      -- 期望大小（字节，photo 为 NULL）
    path TEXT,                         -- 本地绝对路径
    download_date TEXT,                -- YYYY-MM-DD
    status TEXT DEFAULT 'pending',     -- pending/downloading/done/failed/skipped
    error TEXT,
    retries INTEGER DEFAULT 0,         -- 已失败次数（决定退避间隔）
    next_retry_at TEXT,                -- 下次自动重试时间（ISO）
    created_at TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_fwd ON downloads(fwd_channel_id, fwd_msg_id);
"""

_conn = None


def db():
    global _conn
    if _conn is None:
        _conn = sqlite3.connect(DB_FILE)
        _conn.executescript(DB_SCHEMA)
        _conn.commit()
    return _conn


def db_try_claim(origin, channel_id, msg_id, filename, size):
    """尝试插入 pending 记录；唯一键(原始频道ID+原始消息ID)冲突 → 返回 False（已存在）"""
    cid, mid = origin
    try:
        cur = db().execute(
            "INSERT OR IGNORE INTO downloads (fwd_channel_id, fwd_msg_id, channel_id, msg_id, filename, size, download_date, status, created_at) "
            "VALUES (?,?,?,?,?,?,?,?,?)",
            (cid, mid, channel_id, msg_id, filename, size,
             datetime.now().strftime("%Y-%m-%d"), "pending",
             datetime.now().isoformat(timespec="seconds")))
        db().commit()
        return cur.rowcount > 0
    except Exception as e:
        log(f"[!] SQLite 写入失败: {e}")
        return False


def db_set_status(origin, status, path=None, error=None, retries=None, next_retry_at=None):
    cid, mid = origin
    try:
        if retries is None:
            db().execute("UPDATE downloads SET status=?, path=?, error=? WHERE fwd_channel_id=? AND fwd_msg_id=?",
                         (status, path, error, cid, mid))
        else:
            db().execute("UPDATE downloads SET status=?, path=?, error=?, retries=?, next_retry_at=? WHERE fwd_channel_id=? AND fwd_msg_id=?",
                         (status, path, error, retries, next_retry_at, cid, mid))
        db().commit()
    except Exception as e:
        log(f"[!] SQLite 更新失败: {e}")


def db_take_pending():
    """取出遗留 pending/downloading 任务（重启/中断后恢复下载），返回 [(fwd_channel_id, fwd_msg_id, channel_id, msg_id)]"""
    try:
        rows = db().execute(
            "SELECT fwd_channel_id, fwd_msg_id, channel_id, msg_id FROM downloads "
            "WHERE status IN ('pending','downloading')").fetchall()
        return [(r[0], r[1], r[2], r[3]) for r in rows]
    except Exception:
        return []


def db_take_failed_for_retry():
    """取出到期的失败任务并原子置为 pending（防止重复取），返回 [(fwd_channel_id, fwd_msg_id, channel_id, msg_id)]"""
    now = datetime.now().isoformat(timespec="seconds")
    rows = db().execute(
        "SELECT fwd_channel_id, fwd_msg_id, channel_id, msg_id FROM downloads "
        "WHERE status='failed' AND (next_retry_at IS NULL OR next_retry_at <= ?)", (now,)).fetchall()
    out = []
    for cid, mid, chid, mmsg in rows:
        cur = db().execute(
            "UPDATE downloads SET status='pending', next_retry_at=NULL "
            "WHERE fwd_channel_id=? AND fwd_msg_id=? AND status='failed'", (cid, mid))
        if cur.rowcount > 0:
            out.append((cid, mid, chid, mmsg))
    db().commit()
    return out


def db_failed_count():
    try:
        return db().execute("SELECT COUNT(*) FROM downloads WHERE status='failed'").fetchone()[0]
    except Exception:
        return 0


# ---------------- 命名（保留原逻辑） ----------------

DEFAULT_CONFIG = {
    "api_id": 0,
    "api_hash": "",
    "session": os.path.join(BASE_DIR, "tg_session"),
    "download_dir": r"H:\1\tg",
    "channels": [],
    "types": ["photo", "video", "document"],
    "include_forwarded": True,
    "skip_gif": True,
    "download_history": False,       # 启动锚点：默认不扫历史，只监听新消息；需要补扫时临时改 true
    "catchup_today": True,           # 启动监听时补扫转发的消息（下过的自动跳过，SQLite 去重）；False=只下启动后的新消息
    "catchup_since": "",             # 补扫起始时间 "2026-10-07"/"2026-10-07 18:30"；空=今天 0 点
    "only_forwarded": True,          # 仅下载转发来源非空的消息
    "max_file_size_mb": 8192,        # 超过此大小自动跳过（0 = 不限）
    "retries": 3,                    # 单次任务的断连重试次数
    "concurrency": 4,                # 同时下载数（3-5）
    "idle_stop_minutes": 30,         # 连续多少分钟无新消息自动停止监听（0=不启用，防封号/省资源）
    "proxy": None,                   # {"type": "socks5"/"http", "host": ..., "port": ...}
    "name_list_file": "mz.txt",
    "login_method": "qr",            # qr=扫码 / phone=手机号+验证码
    "twofa_password": "",            # 2FA 密码兜底（账号开启两步验证时填）
    "media_browser": {
        "exe": r"D:\media-browser\media-browser-windows-x64.exe",
        "workdir": r"D:\media-browser",
        "port": 8080,
        "index_path": r"D:\media-browser\media_index.json",
    },
}


def load_config():
    if not os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE, "w", encoding="utf-8") as f:
            json.dump(DEFAULT_CONFIG, f, ensure_ascii=False, indent=2)
        log(f"[!] 已生成配置文件 {CONFIG_FILE}")
        log("    请打开它，填入 my.telegram.org 申请的 api_id 和 api_hash")
        sys.exit(1)
    with open(CONFIG_FILE, encoding="utf-8") as f:
        cfg = json.load(f)
    if not cfg.get("api_id") or not cfg.get("api_hash"):
        log(f"[!] {CONFIG_FILE} 里的 api_id / api_hash 为空，请先填写")
        sys.exit(1)
    return cfg


def clean_name(name):
    return re.sub(r'[\\/:*?"<>|\r\n]', "_", str(name)).strip() or "unknown"


_name_list_cache = [None, None, None]


def _resolve_list_path(cfg):
    path = cfg.get("name_list_file", "")
    if not path:
        return ""
    if not os.path.isabs(path):
        path = os.path.join(app_dir(), path)
    return path if os.path.exists(path) else ""


_name_map_cache = [None, None, None, None, None]


def load_name_map(cfg):
    """解析「《作品》：角色1、角色2」格式清单 → (作品集, 角色→作品映射, 角色原名)，键为规范化名"""
    path = _resolve_list_path(cfg)
    if not path:
        return {}, {}, {}
    try:
        mtime = os.path.getmtime(path)
    except Exception:
        mtime = 0
    if _name_map_cache[0] == path and _name_map_cache[1] == mtime:
        return _name_map_cache[2], _name_map_cache[3], _name_map_cache[4]
    works, role2work, role_orig = {}, {}, {}
    try:
        with open(path, encoding="utf-8") as f:
            for ln in f:
                ln = ln.strip()
                if not ln:
                    continue
                m = re.match(r"《(.+?)》", ln)
                wname = m.group(1).strip() if m else ""
                if wname:
                    works[_norm(wname)] = wname
                body = ln.split("：", 1)[-1] if "：" in ln else (ln.split(":", 1)[-1] if ":" in ln else "")
                if body and body != ln:
                    for piece in re.split(r"[、，,，\s]+", body):
                        piece = piece.strip()
                        if piece and wname:
                            role2work[_norm(piece)] = wname
                            role_orig[_norm(piece)] = piece
    except Exception as e:
        log(f"[!] 名单映射读取失败: {e}")
    _name_map_cache[:] = [path, mtime, works, role2work, role_orig]
    return works, role2work, role_orig


def _norm(s):
    return re.sub(r'[\s\-_.,()\[\]【】（）\.]', "", str(s)).lower()


def _extract_candidates(text):
    cands = set()
    for m in re.finditer(r"#([^\s#，,。.！!?？]+)", text):
        cands.add(m.group(1).strip())
    rest = re.sub(r"#\S+", "", text)
    for piece in re.split(r"[\s、，,。.！!?？:：;；()（）\[\]【】\-]+", rest):
        piece = piece.strip()
        if len(piece) >= 2:
            cands.add(piece)
    return cands


def match_list_name(cfg, text):
    """名单匹配：候选词命中角色→「作品-角色」，命中作品→作品名；支持子串；无命中 None"""
    works, role2work, role_orig = load_name_map(cfg)
    if not works and not role2work:
        return None
    ncands = {_norm(c): c for c in _extract_candidates(text)}
    if not ncands:
        return None
    for nc, orig in ncands.items():
        if nc in role2work:
            w = role2work[nc]
            return f"{w}-{orig}" if w else orig
    for nc, orig in ncands.items():
        if nc in works:
            return works[nc]
    for nc, orig in ncands.items():
        for rn, ro in role_orig.items():
            if rn and rn in nc:
                w = role2work[rn]
                return f"{w}-{ro}" if w else ro
    for nc, orig in ncands.items():
        for wn, wo in works.items():
            if wn and wn in nc:
                return wo
    return None


def smart_name(cfg, msg):
    """智能命名：名单匹配 > None(调用方退回 caption)。不再调用任何 API"""
    text = (msg.message or "").strip()
    if not text:
        return None
    return match_list_name(cfg, text)


def should_download(cfg, msg):
    if not msg.media:
        return False
    if not cfg.get("include_forwarded", True) and msg.fwd_from:
        return False
    types = cfg.get("types", [])
    if msg.photo:
        return "photo" in types
    if msg.document:
        doc = msg.document
        attrs = {type(a) for a in doc.attributes}
        mime = (doc.mime_type or "").lower()
        if DocumentAttributeSticker in attrs or DocumentAttributeAudio in attrs:
            return False
        if DocumentAttributeVideo in attrs:
            round_msg = any(isinstance(a, DocumentAttributeVideo) and a.round_message for a in doc.attributes)
            if round_msg:
                return False
            return "video" in types
        if DocumentAttributeAnimated in attrs:
            return "video" in types and not cfg.get("skip_gif", True)
        if mime.startswith("video/"):
            return "video" in types
        return "document" in types
    return False


def build_filename(msg):
    """文件名：caption > 原始文件名 > photo/media 兜底。msg.id 前缀保证幂等"""
    name = ""
    if msg.message and msg.message.strip():
        name = msg.message.strip().replace("\r", " ").replace("\n", " ")[:40]
    if not name and msg.document:
        for a in msg.document.attributes:
            if isinstance(a, DocumentAttributeFilename):
                name = a.file_name
                break
    if not name and msg.photo:
        name = "photo.jpg"
    if not name:
        name = "media"
    name = clean_name(name)
    return f"{msg.id}_{name}"


# ---------------- 转发来源解析（去重关键） ----------------

def fwd_origin(msg):
    """
    解析转发来源 → (原始频道ID, 原始消息ID)。
    优先 saved_from_peer（完整转发链），回退 from_id + channel_post。
    无转发来源返回 None。
    """
    fwd = msg.fwd_from
    if not fwd:
        return None
    peer = None
    mid = 0
    if getattr(fwd, "saved_from_peer", None) is not None:
        peer = fwd.saved_from_peer
        mid = getattr(fwd, "saved_from_msg_id", 0) or 0
    if peer is None:
        peer = fwd.from_id
        mid = getattr(fwd, "channel_post", 0) or 0
    if peer is None:
        return None
    cid = getattr(peer, "channel_id", None)
    if cid is None:
        cid = getattr(peer, "user_id", None)
    if cid is None:
        cid = getattr(peer, "chat_id", None)
    if cid is None:
        try:
            cid = int(peer)
        except Exception:
            return None
    if not mid:
        return None  # 原始消息 ID 缺失，无法去重
    return (int(cid), int(mid))


# ---------------- 状态文件（进度/最近完成，兼容旧格式） ----------------

_progress = {}
# 活跃下载计数：>0 表示仍有文件在下载（含续传），空闲自动停止必须等它归零
_ACTIVE_DOWNLOADS = 0
# 最近一次"全部下载完成"的时间戳：下载全部结束后才重新开始空闲计时
_LAST_DOWNLOAD_DONE = 0.0


def _dec_active():
    """下载任务结束（成功/取消/失败）时调用；全部结束时刷新空闲计时起点"""
    global _ACTIVE_DOWNLOADS, _LAST_DOWNLOAD_DONE
    _ACTIVE_DOWNLOADS = max(0, _ACTIVE_DOWNLOADS - 1)
    if _ACTIVE_DOWNLOADS == 0:
        _LAST_DOWNLOAD_DONE = time.time()
_last_progress_write = 0.0


def write_progress():
    global _last_progress_write
    now = time.time()
    if now - _last_progress_write < 0.5:
        return
    _last_progress_write = now
    try:
        tasks = []
        for p, info in _progress.items():
            total = info["total"] or 0
            pct = int(info["done"] * 100 / total) if total else 0
            tasks.append({
                "file": info["file"],
                "pct": min(100, max(0, pct)),
                "done_mb": round(info["done"] / 1048576, 1),
                "total_mb": round(total / 1048576, 1),
            })
        with open(PROGRESS_FILE, "w", encoding="utf-8") as f:
            json.dump({"tasks": tasks, "time": datetime.now().strftime("%H:%M:%S")},
                      f, ensure_ascii=False)
    except Exception:
        pass


def write_status(chat_name, path):
    """下载成功记录（最近 10 条）"""
    try:
        with open(STATUS_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
        if not isinstance(data, list):
            data = []
    except Exception:
        data = []
    data.append({
        "time": datetime.now().strftime("%H:%M:%S"),
        "chat": chat_name,
        "file": os.path.basename(path),
    })
    try:
        with open(STATUS_FILE, "w", encoding="utf-8") as f:
            json.dump(data[-10:], f, ensure_ascii=False, indent=1)
    except Exception:
        pass


# ---------------- media-browser 联动 ----------------

MB_PROC = None


def mb_config(cfg):
    mb = cfg.get("media_browser") or {}
    exe = mb.get("exe") or r"D:\media-browser\media-browser-windows-x64.exe"
    return {
        "exe": exe,
        "workdir": mb.get("workdir") or os.path.dirname(exe),
        "port": int(mb.get("port") or 8080),
        # 未配置 index_path 时落到程序目录；不能是裸文件名，否则 dirname 为空串，
        # 重建索引时 os.makedirs("") 会抛 WinError 3（系统找不到指定的路径）
        "index_path": mb.get("index_path") or os.path.join(app_dir(), "media_index.json"),
    }


def kill_media_browser():
    """杀掉 media-browser（句柄优先，taskkill 兜底）"""
    global MB_PROC
    if MB_PROC and MB_PROC.poll() is None:
        try:
            MB_PROC.kill()
        except Exception:
            pass
    try:
        subprocess.run(["taskkill", "/IM", "media-browser-windows-x64.exe", "/F"],
                       capture_output=True, timeout=10)
    except Exception:
        pass
    MB_PROC = None
    time.sleep(1)


def spawn_media_browser(cfg):
    """拉起 media-browser 子进程"""
    global MB_PROC
    mb = mb_config(cfg)
    if port_open(mb["port"]):
        log(f"[OK] media-browser 已在 {mb['port']} 端口运行，复用")
        return
    if not os.path.exists(mb["exe"]):
        log(f"[!] 找不到 media-browser: {mb['exe']}")
        return
    args = ["-dir", cfg["download_dir"], "-index", "media_index.json",
            "-thumbs", "thumbs", "-port", str(mb["port"])]
    try:
        MB_PROC = subprocess.Popen([mb["exe"]] + args, cwd=mb["workdir"],
                                   creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
        log(f"[OK] media-browser 已启动: http://localhost:{mb['port']}")
    except Exception as e:
        log(f"[!] media-browser 启动失败: {e}")


def build_media_index(download_dir, index_path):
    """扫描下载目录，重建 media-browser 索引（相对路径/正斜杠，Date=mtime，IsVideo=扩展名）"""
    entries = []
    if os.path.isdir(download_dir):
        for root, dirs, files in os.walk(download_dir):
            for fn in files:
                full = os.path.join(root, fn)
                try:
                    rel = os.path.relpath(full, download_dir).replace("\\", "/")
                    mtime = os.path.getmtime(full)
                    dt = datetime.fromtimestamp(mtime).astimezone().isoformat(timespec="milliseconds")
                    is_video = os.path.splitext(fn)[1].lower() in VIDEO_EXTS
                    entries.append({"Path": rel, "Date": dt, "IsVideo": is_video})
                except Exception:
                    continue
    entries.sort(key=lambda e: e.get("Date", ""), reverse=True)
    try:
        # index_path 可能是裸文件名（未配置 media_browser.index_path 时），
        # 此时 dirname 为空串，os.makedirs("") 会抛 WinError 3
        d = os.path.dirname(index_path)
        if d:
            os.makedirs(d, exist_ok=True)
        tmp = index_path + ".tmp"
        with open(tmp, "w", encoding="utf-8") as f:
            json.dump(entries, f, ensure_ascii=False)
        os.replace(tmp, index_path)
        return len(entries)
    except Exception as e:
        log(f"[!] 索引写入失败: {e}")
        return -1


_refresh_pending = False


def notify_station_refresh(cfg):
    """整合模式：通知 Go 后端(tg_station)重新扫描媒体索引。

    POST http://127.0.0.1:<port>/api/refresh，成功返回 True；
    Go 后端不在运行（独立模式）时返回 False，调用方回退旧逻辑。
    """
    mb = mb_config(cfg)
    url = f"http://127.0.0.1:{mb['port']}/api/refresh"
    try:
        req = urllib.request.Request(url, data=b"{}", method="POST")
        urllib.request.urlopen(req, timeout=3)
        return True
    except Exception:
        return False


async def _debounced_refresh(cfg):
    """下载完成后 5 秒窗口合并，通知 tg_station 刷新媒体索引（页面自动刷新）。

    整合模式：POST /api/refresh 由 Go 后端重新扫描索引 + 生成缩略图；
    独立模式（Go 后端不在）：回退为旧逻辑 —— 重建索引文件 + 重启 media-browser。
    """
    global _refresh_pending
    await asyncio.sleep(5)
    _refresh_pending = False
    try:
        if notify_station_refresh(cfg):
            log("[OK] 已通知 tg_station 刷新媒体索引")
            return
        # 回退：独立 media-browser 模式
        mb = mb_config(cfg)
        n = build_media_index(cfg["download_dir"], mb["index_path"])
        if n >= 0:
            log(f"[OK] 媒体索引已重建（{n} 个文件），刷新 media-browser")
        kill_media_browser()
        spawn_media_browser(cfg)
    except Exception as e:
        log(f"[!] 刷新媒体索引失败: {e}")


def schedule_refresh(cfg):
    """下载完成时调用；5 秒内多次完成只刷新一次"""
    global _refresh_pending
    if _refresh_pending:
        return
    _refresh_pending = True
    try:
        asyncio.get_running_loop().create_task(_debounced_refresh(cfg))
    except RuntimeError:
        _refresh_pending = False


# ---------------- 下载执行器 ----------------

async def download_msg(client, cfg, chat, msg, origin):
    """下载单条消息；成功/失败/跳过都写 SQLite。断连类错误按 cfg.retries 重试"""
    path = None
    try:
        chat_name = getattr(chat, "title", None) or getattr(chat, "first_name", str(getattr(chat, "id", "?")))
        # 兜底：重试场景 chat 可能为 None，从 msg 恢复
        if not chat_name or chat_name == "?":
            try:
                chat_name = getattr(msg.chat, "title", None) or getattr(msg.chat, "first_name", "")
            except Exception:
                pass
        month = datetime.now().strftime("%Y-%m")
        folder = os.path.join(cfg["download_dir"], clean_name(chat_name), month)
        os.makedirs(folder, exist_ok=True)

        # 大小过滤（photo 无 size，不限制）
        max_mb = int(cfg.get("max_file_size_mb") or 0)
        if max_mb > 0 and msg.document and msg.document.size and msg.document.size > max_mb * 1048576:
            log(f"[--] 超过大小阈值({max_mb}MB)跳过: {msg.document.size // 1048576}MB")
            db_set_status(origin, "skipped", error=f"size>{max_mb}MB")
            return

        matched = await asyncio.to_thread(smart_name, cfg, msg)
        fname = f"{msg.id}_{clean_name(matched)}" if matched else build_filename(msg)
        path = os.path.join(folder, fname)
        expected = msg.document.size if msg.document else None

        # 幂等：磁盘上已有「消息ID_」前缀的完整文件则跳过（命名规则变化不重下）；残片则续传
        resume_from = None
        if os.path.isdir(folder):
            for fn in os.listdir(folder):
                if fn.startswith(f"{msg.id}_"):
                    p = os.path.join(folder, fn)
                    try:
                        sz = os.path.getsize(p)
                    except Exception:
                        sz = 0
                    complete = (expected is not None and sz == expected and sz > 0) or (expected is None and sz > 0)
                    if complete:
                        log(f"[--] 已存在(ID匹配): {fn}")
                        db_set_status(origin, "done", path=p)
                        return
                    log(f"[R] 发现残片(ID匹配): {fn}，续传")
                    resume_from = p
                    break
        if resume_from is None and os.path.exists(path):
            sz = os.path.getsize(path)
            if expected is not None and sz == expected and sz > 0:
                db_set_status(origin, "done", path=path)
                return
            if expected is None and sz > 0:
                db_set_status(origin, "done", path=path)
                return
            if expected is not None and sz < expected:
                log(f"[R] 发现残片({sz // 1048576}MB/{expected // 1048576}MB)，续传")
                resume_from = path
        if resume_from:
            path = resume_from

        db_set_status(origin, "downloading", path=path)
        log(f"[↓] 开始下载: {os.path.basename(path)} ({expected // 1048576 if expected else '?'}MB)")
        global _ACTIVE_DOWNLOADS
        _ACTIVE_DOWNLOADS += 1

        max_retries = max(1, int(cfg.get("retries") or 3))
        saved = None
        # 单任务取消：cancel.json 里记录的文件名集合，命中则放弃下载（删残片）
        def is_cancelled():
            try:
                cf = os.path.join(BASE_DIR, "cancel.json")
                if os.path.exists(cf):
                    with open(cf, encoding="utf-8") as f:
                        lst = json.load(f)
                    return os.path.basename(path) in lst
            except Exception:
                pass
            return False
        for attempt in range(1, max_retries + 1):
            try:
                def progress_cb(received, total):
                    if total:
                        _progress[path] = {"file": os.path.basename(path), "done": received, "total": total}
                        write_progress()
                        # 每下载一段检查取消标记，命中则中断（抛异常触发清理）
                        if is_cancelled():
                            raise asyncio.CancelledError()
                done = os.path.getsize(path) if os.path.exists(path) else 0
                if done > 0 and expected and done < expected:
                    # 断点续传：Telethon download_media 不支持 offset（字符串路径从头覆盖写），
                    # 改用 iter_download 从已有大小继续
                    log(f"[R] 断点续传: {os.path.basename(path)} ({done // 1048576}MB 起)")
                    received = done
                    with open(path, "r+b") as fd:
                        fd.seek(0, 2)
                        async for chunk in client.iter_download(msg.document, offset=fd.tell(),
                                                                request_size=1024 * 1024, file_size=expected):
                            fd.write(chunk)
                            received += len(chunk)
                            progress_cb(received, expected)
                    saved = path
                else:
                    saved = await msg.download_media(file=path, progress_callback=progress_cb)
                break
            except (ConnectionError, TimeoutError, OSError, RuntimeError) as e:
                if attempt >= max_retries:
                    raise
                wait = min(5 * attempt, 30)
                log(f"[!] 下载中断(第{attempt}次): {str(e)[:80]}，{wait}s 后重试（残片保留，自动续传）")
                await asyncio.sleep(wait)
                if not client.is_connected():
                    try:
                        await client.connect()
                    except Exception as ce:
                        log(f"[!] 重连失败: {ce}")

        if saved:
            # 关键：download_media 传无扩展名路径时，Telethon 会按媒体类型
            # 自动补扩展名（photo→.jpg 等），saved 返回的是实际保存路径。
            # 后续所有操作必须用 saved，否则 WinError 2（找不到无扩展名文件）
            real = saved if isinstance(saved, str) else path
            # 图包/长 caption 命名时 Telethon 补不出扩展名 → 文件无扩展名 → 页面不收录。
            # 这里按文件头兜底补上，避免"下好了但看不见"。
            if os.path.splitext(real)[1].lower() not in MEDIA_EXTS:
                ne = sniff_media_ext(real, msg.media)
                if ne:
                    newp = real + ne
                    try:
                        os.replace(real, newp)
                        log(f"[i] 补扩展名: {os.path.basename(newp)}")
                        real = newp
                    except Exception as e:
                        log(f"[!] 补扩展名失败: {e}")
            size_mb = os.path.getsize(real) / 1024 / 1024
            log(f"[OK] {chat_name} | {size_mb:.1f}MB | {os.path.basename(real)}")
            db_set_status(origin, "done", path=real)
            write_status(chat_name, real)
            _progress.pop(path, None)
            write_progress()
            schedule_refresh(cfg)  # 下载完成 → 刷新媒体浏览器
            _dec_active()
        else:
            raise RuntimeError("下载无内容")
    except asyncio.CancelledError:
        # 用户取消（cancel.json 命中）：清理残片，标记 cancelled，不安排重试
        _progress.pop(path, None) if path else None
        write_progress()
        if path:
            try:
                for fn in os.listdir(os.path.dirname(path)):
                    if fn.startswith(f"{msg.id}_"):
                        try:
                            os.remove(os.path.join(os.path.dirname(path), fn))
                        except Exception:
                            pass
            except Exception:
                pass
        db_set_status(origin, "cancelled", error="用户取消", retries=0, next_retry_at="")
        log(f"[X] 已取消下载: {os.path.basename(path) if path else ''}")
        _dec_active()
        return
    except Exception as e:
        _progress.pop(path, None) if path else None
        write_progress()
        # 清理残片：path 可能无扩展名（Telethon 自动补），按 msg.id 前缀删目录内文件
        if path:
            try:
                for fn in os.listdir(os.path.dirname(path)):
                    if fn.startswith(f"{msg.id}_"):
                        try:
                            os.remove(os.path.join(os.path.dirname(path), fn))
                        except Exception:
                            pass
            except Exception:
                pass
        # 记录失败 + 安排自动重试（退避 5→15→30→60 分钟）
        retries = 1
        try:
            row = db().execute("SELECT retries FROM downloads WHERE fwd_channel_id=? AND fwd_msg_id=?",
                               (origin[0], origin[1])).fetchone()
            retries = (row[0] if row else 0) + 1
        except Exception:
            pass
        interval = RETRY_INTERVALS[min(retries - 1, len(RETRY_INTERVALS) - 1)]
        next_at = (datetime.now() + timedelta(minutes=interval)).isoformat(timespec="seconds")
        db_set_status(origin, "failed", error=str(e)[:200], retries=retries, next_retry_at=next_at)
        log(f"[!] 下载失败(将自动重试): {str(e)[:100]} | 第{retries}次失败，{interval}分钟后再试")
        _dec_active()


async def worker(client, cfg, queue):
    """并发下载 worker"""
    while True:
        item = await queue.get()
        try:
            if item is None:
                return
            chat, msg, origin = item
            await download_msg(client, cfg, chat, msg, origin)
        except asyncio.CancelledError:
            raise
        except Exception as e:
            log(f"[!] worker 异常: {e}")
        finally:
            queue.task_done()


async def retry_loop(client, cfg, queue):
    """每 60s 检查失败任务，到期的重新从 TG 拉消息入队（重启后也能续重试）"""
    while True:
        await asyncio.sleep(60)
        try:
            tasks = db_take_failed_for_retry()
            for cid, mid, chid, mmsg in tasks:
                try:
                    msg = await client.get_messages(chid, ids=mmsg)
                    if msg and msg.media:
                        log(f"[R] 自动重试: 原始({cid},{mid}) 消息{mmsg}")
                        # 恢复频道实体：否则 chat=None 会导致目录名解析成 "_"
                        chat_ent = None
                        try:
                            chat_ent = await client.get_entity(chid)
                        except Exception:
                            chat_ent = getattr(msg, "chat", None)
                        await queue.put((chat_ent, msg, (cid, mid)))
                    else:
                        # 消息已不可用（被删等），标记失败终态
                        db_set_status((cid, mid), "failed", error="消息已不可用")
                except Exception as e:
                    log(f"[!] 重试取消息失败({cid},{mid}): {e}")
            if tasks and db_failed_count() > 0:
                log(f"[i] 当前失败任务 {db_failed_count()} 个，等待自动重试")
        except asyncio.CancelledError:
            raise
        except Exception as e:
            log(f"[!] 重试循环异常: {e}")


async def handle_message(client, cfg, queue, chat, msg):
    """新消息入口：仅转发 → 大小过滤 → SQLite 去重 → 入队"""
    if not msg.media:
        return
    if cfg.get("only_forwarded", True) and not msg.fwd_from:
        return  # 仅限转发来源非空的消息
    if not should_download(cfg, msg):
        return
    origin = fwd_origin(msg)
    if origin is None:
        if cfg.get("only_forwarded", True):
            return  # 转发但解析不到来源，跳过
        origin = (0, msg.id)  # 非转发模式：用自身消息 ID 去重
    size = msg.document.size if msg.document else None
    if not db_try_claim(origin, getattr(chat, "id", 0), msg.id, build_filename(msg), size):
        log(f"[--] 已存在(SQLite): 原始({origin[0]},{origin[1]})")
        return
    max_mb = int(cfg.get("max_file_size_mb") or 0)
    if max_mb > 0 and size and size > max_mb * 1048576:
        log(f"[--] 超过大小阈值({max_mb}MB)跳过: {size // 1048576}MB")
        db_set_status(origin, "skipped", error=f"size>{max_mb}MB")
        return
    log(f"[+] 检测到新转发: 原始({origin[0]},{origin[1]}) | {build_filename(msg)}")
    await queue.put((chat, msg, origin))


# ---------------- 历史全量（默认关闭，config.download_history=true 时启用） ----------------

async def scan_history(client, cfg, dialog, queue):
    """全量下载频道历史媒体（启动锚点模式默认不执行）"""
    log(f"[*] 开始全量下载历史: {dialog.title} ...")
    count = 0
    try:
        async for msg in client.iter_messages(dialog.id, wait_time=0.3):
            if msg.media:
                await handle_message(client, cfg, queue, dialog, msg)
                count += 1
    except Exception as e:
        log(f"[!] 历史遍历中断: {e}")
    log(f"[*] 历史媒体处理完毕，共 {count} 条")


# ---------------- 当天补扫（默认开启，config.catchup_today=true 时启用） ----------------

async def catchup_today(client, cfg, chat, queue):
    """补扫起始时间之后转发的消息（下过的自动跳过，SQLite 去重）。

    config.catchup_since 可指定起始时间，如 "2026-10-07" 或 "2026-10-07 18:30"；
    留空 = 今天 0 点。limit 5000 条消息（含非媒体）以覆盖长时间停机后的积压。
    """
    try:
        raw = str(cfg.get("catchup_since") or "").strip()
        since = None
        for fmt in ("%Y-%m-%d %H:%M", "%Y-%m-%dT%H:%M", "%Y-%m-%d"):
            if not raw:
                break
            try:
                since = datetime.strptime(raw, fmt)
                break
            except ValueError:
                continue
        if since is None:
            since = datetime.now().replace(hour=0, minute=0, second=0, microsecond=0)
        title = getattr(chat, "title", None) or chat.id
        log(f"[*] 历史补扫开始: {title}（{since:%Y-%m-%d %H:%M} 之后）...")
        count = 0
        async for msg in client.iter_messages(chat, offset_date=since, reverse=True, limit=5000):
            count += 1
            await handle_message(client, cfg, queue, chat, msg)
        log(f"[*] 历史补扫完成: {title}，检查 {count} 条（下过的自动跳过）")
    except Exception as e:
        log(f"[!] 当天补扫失败({getattr(chat, 'id', chat)}): {e}")


# ---------------- 登录 ----------------

async def do_login(client, cfg):
    """登录：session 有效直接过；否则按 login_method 扫码或手机号；2FA 用 twofa_password 兜底"""
    if not await client.is_user_authorized():
        if cfg.get("login_method", "qr") == "phone":
            # 手机号 + 验证码
            phone = input("请输入手机号（带国家码，如 +86138...）: ").strip()
            sent = await client.send_code_request(phone)
            code = input("请输入验证码: ").strip()
            try:
                await client.sign_in(phone, code)
            except SessionPasswordNeededError:
                pw = cfg.get("twofa_password") or ""
                if not pw:
                    log("[!] 账号需要两步验证：请把 twofa_password 填进 config.json 后重启")
                    write_login_status(True, note="需要两步验证密码")
                    raise
                await client.sign_in(password=pw)
        else:
            # 扫码登录（默认）：生成二维码图片并打开，过期自动重新生成
            import qrcode
            qr_path = os.path.join(BASE_DIR, "qr_login.png")
            log("[!] 需要登录，正在生成二维码并打开图片...")
            while True:
                try:
                    qr = await client.qr_login()
                    qrcode.make(qr.url, box_size=8, border=2).save(qr_path)
                    write_login_status(True, qr_path)
                    if os.name == "nt":
                        os.startfile(qr_path)  # 打开图片供手机扫
                    log(f"[!] 二维码已生成: {qr_path}（{qr.url[:60]}...）")
                    await qr.wait()
                    log("[OK] 扫码成功!")
                    break
                except Exception as e:
                    log(f"[!] 二维码过期或登录中断，重新生成: {str(e)[:60]}")
                    await asyncio.sleep(2)
            # 扫码后可能还需要 2FA
            try:
                await client.start(password=cfg.get("twofa_password") or None)
            except SessionPasswordNeededError:
                pw = cfg.get("twofa_password") or ""
                if not pw:
                    log("[!] 账号需要两步验证：请把 twofa_password 填进 config.json 后重启")
                    write_login_status(True, note="需要两步验证密码")
                    raise
                await client.sign_in(password=pw)
        # 登录成功 → 写授权标记（供 --check-session / tg_station 登录状态接口判定）
        try:
            me = await client.get_me()
            write_auth_marker(cfg, me.username or me.first_name or "")
        except Exception:
            pass
        write_login_status(False)
    me = await client.get_me()
    log(f"[OK] 已登录: {me.first_name} (@{me.username})")


# ---------------- 主流程 ----------------

def build_proxy(cfg):
    """从配置构建 Telethon proxy 参数；未配置代理返回 None"""
    if not cfg.get("proxy"):
        return None
    import socks
    pc = dict(cfg["proxy"])
    # 优先使用 ensure_proxy() 探测到的 Clash Verge 端口
    if ACTIVE_PROXY_PORT:
        pc["host"] = "127.0.0.1"
        pc["port"] = ACTIVE_PROXY_PORT
    if str(pc.get("type", "socks5")).lower() == "http":
        return ("http", pc["host"], pc["port"])
    return (socks.SOCKS5, pc["host"], pc["port"])


def auth_marker_path(cfg):
    """扫码成功标记文件：<session>.authorized。

    原理：Telethon 连接时（未扫码）就会生成真实 auth key，无法靠文件内容
    区分登录态；只有扫码成功才写标记，保证 status 不误判。
    """
    return os.path.join(cfg["session"] + ".authorized")


def write_auth_marker(cfg, username=""):
    try:
        with open(auth_marker_path(cfg), "w", encoding="utf-8") as f:
            json.dump({"authorized": True, "username": username,
                       "ts": datetime.now().isoformat(timespec="seconds")}, f)
    except Exception:
        pass


def remove_auth_marker(cfg):
    try:
        os.remove(auth_marker_path(cfg))
    except OSError:
        pass


def check_session(cfg):
    """--check-session 模式：判断 session 是否已扫码登录（不连网，毫秒级）。

    判定依据是扫码成功标记文件（.authorized）——未扫码即使连过 TG 也无标记。
    输出: {"status": "logged_in"} / {"status": "not_logged_in"}
    """
    ok = os.path.exists(auth_marker_path(cfg))
    print(json.dumps({"status": "logged_in" if ok else "not_logged_in"}), flush=True)


async def list_channels_only(cfg):
    """--list-channels 模式：列出账号已加入的频道（含 ID 和标题），输出 JSON 后退出。

    输出: {"status": "ok", "channels": [{"id": -100..., "title": "龙"}, ...]}
    供 tg_station 设置面板的「频道选择器」加载勾选列表用。
    """
    def out(d):
        print(json.dumps(d, ensure_ascii=False), flush=True)

    client = None
    try:
        client = TelegramClient(cfg["session"], cfg["api_id"], cfg["api_hash"],
                                proxy=build_proxy(cfg))
        await client.connect()
        me = await client.get_me()
        if not me:
            out({"status": "failed", "error": "未登录，请先扫码登录"})
            return
        dialogs = await client.get_dialogs()
        channels = [{"id": d.id, "title": d.title or ""}
                    for d in dialogs if d.is_channel and not d.is_group]
        channels.sort(key=lambda c: c["title"] or "")
        out({"status": "ok", "channels": channels})
    except Exception as e:
        out({"status": "failed", "error": str(e)[:200]})
    finally:
        if client:
            try:
                await client.disconnect()
            except Exception:
                pass


async def qr_login_only(cfg):
    """--qr-only 模式：只执行扫码登录，输出一行 JSON 到 stdout 后退出。

    输出格式：
      已登录     {"status": "already", "username": "xxx"}
      二维码就绪 {"status": "qr", "qr_base64": "..."}
      扫码成功   {"status": "success", "username": "xxx"}
      失败/超时  {"status": "failed", "error": "..."}
    """
    def out(d):
        print(json.dumps(d, ensure_ascii=False), flush=True)

    try:
        import qrcode
    except ImportError:
        out({"status": "failed", "error": "缺少 qrcode 库，请运行: pip install qrcode[pil]"})
        return

    client = None
    try:
        client = TelegramClient(cfg["session"], cfg["api_id"], cfg["api_hash"],
                                proxy=build_proxy(cfg))
        await client.connect()
        me = await client.get_me()
        if me:
            write_auth_marker(cfg, me.username or me.first_name or "")
            out({"status": "already", "username": me.username or me.first_name or str(me.id)})
            return

        # 未登录：作废旧标记，生成二维码（120 秒内有效）
        remove_auth_marker(cfg)
        qr = await client.qr_login()
        import base64, io
        buf = io.BytesIO()
        qrcode.make(qr.url, box_size=8, border=2).save(buf, format="PNG")
        out({"status": "qr", "qr_base64": base64.b64encode(buf.getvalue()).decode()})

        try:
            await qr.wait(timeout=120)
        except TimeoutError:
            out({"status": "failed", "error": "扫码超时（120 秒），请重试"})
            return
        me = await client.get_me()
        write_auth_marker(cfg, me.username or me.first_name or "")
        out({"status": "success", "username": me.username or me.first_name or str(me.id)})
    except Exception as e:
        out({"status": "failed", "error": str(e)[:200]})
    finally:
        if client:
            try:
                await client.disconnect()
            except Exception:
                pass


async def run(cfg):
    client = TelegramClient(cfg["session"], cfg["api_id"], cfg["api_hash"],
                            proxy=build_proxy(cfg))
    await client.connect()
    await do_login(client, cfg)

    dialogs = await client.get_dialogs()
    all_channels = [d for d in dialogs if d.is_channel and not d.is_group]
    names = cfg.get("channels") or []

    def match(d):
        for item in names:
            s = str(item).strip()
            if not s:
                continue
            if s.lstrip("-").isdigit() and int(s) == d.id:
                return True
            if s in (d.title or "") or (d.title or "") in s:
                return True
        return False

    if names:
        targets = [d for d in all_channels if match(d)]
    else:
        targets = all_channels

    if not targets:
        log("[!] 没有找到任何频道。确认你的账号加入了频道，或填 config.json 的 channels")
        return

    log(f"[OK] 正在监听 {len(targets)} 个频道:")
    for d in targets:
        log(f"     - {d.title} (id={d.id})")
    log(f"[OK] 下载目录: {cfg['download_dir']}")
    log(f"[OK] 仅转发来源非空: {cfg.get('only_forwarded', True)} | 大小上限: {cfg.get('max_file_size_mb', 500)}MB | 并发: {cfg.get('concurrency', 4)}")
    log("[OK] 开始监听（启动锚点：只处理启动之后的新消息），Ctrl+C 退出")

    queue = asyncio.Queue(maxsize=200)
    n = max(1, min(8, int(cfg.get("concurrency", 4))))
    workers = [asyncio.create_task(worker(client, cfg, queue)) for _ in range(n)]
    retry_task = asyncio.create_task(retry_loop(client, cfg, queue))

    # 空闲自动停止：连续 idle_stop_minutes 分钟无新消息 → 自动断开退出（防封号/省资源）
    idle_minutes = int(cfg.get("idle_stop_minutes") or 0)
    last_activity = [time.time()]

    async def idle_watchdog():
        global stop_requested
        while True:
            await asyncio.sleep(60)
            # 仍有文件在下载（含续传）→ 不因空闲停止，避免下载中断
            if _ACTIVE_DOWNLOADS > 0:
                continue
            # 空闲计时起点：最后一条新消息 与 全部下载完成时刻，取较新者
            idle_since = max(last_activity[0], _LAST_DOWNLOAD_DONE)
            if time.time() - idle_since > idle_minutes * 60:
                log(f"[*] 空闲 {idle_minutes} 分钟无新文件，自动停止监听（需要时点「开始监听」恢复）")
                stop_requested = True
                await client.disconnect()
                return

    if idle_minutes > 0:
        idle_task = asyncio.create_task(idle_watchdog())
    else:
        idle_task = None

    @client.on(events.NewMessage(chats=[d.id for d in targets]))
    async def handler(event):
        try:
            last_activity[0] = time.time()
            chat = await event.get_chat()
            await handle_message(client, cfg, queue, chat, event.message)
        except asyncio.CancelledError:
            raise
        except Exception as e:
            log(f"[!] 处理消息出错: {e}")

    tasks = [client.run_until_disconnected()]
    if idle_task:
        tasks.append(idle_task)
    if cfg.get("download_history", False) and not is_paused():
        for d in targets:
            tasks.append(scan_history(client, cfg, d, queue))
    else:
        log("[*] 启动锚点模式：不扫历史，只监听新消息")
        # 恢复上次中断遗留的下载任务（关机/崩溃后重启不丢）
        for cid, mid, chid, mmsg in db_take_pending():
            try:
                msg = await client.get_messages(chid, ids=mmsg)
                if msg and msg.media:
                    try:
                        chat_ent = await client.get_entity(chid)
                    except Exception:
                        chat_ent = getattr(msg, "chat", None)
                    await queue.put((chat_ent, msg, (cid, mid)))
                    log(f"[R] 恢复上次未完成的下载: 原始({cid},{mid})")
            except Exception as e:
                log(f"[!] 恢复下载失败({cid},{mid}): {e}")
        # 当天补扫：今天 0 点起转发的消息，已下载的自动跳过（SQLite 去重）
        if cfg.get("catchup_today", True):
            for d in targets:
                await catchup_today(client, cfg, d, queue)
    try:
        await asyncio.gather(*tasks)
    finally:
        for t in workers:
            t.cancel()
        retry_task.cancel()
        if idle_task:
            idle_task.cancel()


def main():
    # 空闲自动停止标志：idle_watchdog 触发后置 True，main 循环据此退出进程
    global stop_requested
    stop_requested = False
    # --qr-only：只做扫码登录（供 tg_station 网页调用），输出 JSON 后退出
    if "--qr-only" in sys.argv:
        try:
            ensure_resources()
            ensure_proxy()  # 登录也需要走代理
            cfg = load_config()
        except SystemExit:
            # load_config 在 api_id/api_hash 为空时会 sys.exit(1)
            print(json.dumps({"status": "failed",
                              "error": "config.json 缺少 api_id/api_hash"}, ensure_ascii=False),
                  flush=True)
            return
        try:
            asyncio.run(qr_login_only(cfg))
        except Exception as e:
            print(json.dumps({"status": "failed", "error": str(e)[:200]}, ensure_ascii=False),
                  flush=True)
        return

    # --check-session：本地检查 session 是否有效（供 tg_station 登录状态接口）
    if "--check-session" in sys.argv:
        try:
            cfg = load_config()
        except SystemExit:
            print(json.dumps({"status": "not_logged_in"}), flush=True)
            return
        check_session(cfg)
        return

    # --list-channels：列出账号已加入的频道（供 tg_station 频道选择器）
    if "--list-channels" in sys.argv:
        try:
            ensure_resources()
            ensure_proxy()  # 拉取频道列表也需要走代理
            cfg = load_config()
        except SystemExit:
            print(json.dumps({"status": "failed",
                              "error": "config.json 缺少 api_id/api_hash"}, ensure_ascii=False),
                  flush=True)
            return
        try:
            asyncio.run(list_channels_only(cfg))
        except Exception as e:
            print(json.dumps({"status": "failed", "error": str(e)[:200]}, ensure_ascii=False),
                  flush=True)
        return

    while True:
        try:
            ensure_resources()
            ensure_proxy()
            init_pause()
            cfg = load_config()
            spawn_media_browser(cfg)  # 启动时拉起媒体浏览器
            asyncio.run(run(cfg))
            if stop_requested:
                log("[OK] 已停止（空闲自动停止）")
                break
        except KeyboardInterrupt:
            log("[OK] 已退出")
            break
        except Exception as e:
            log(f"[!] 异常: {e}，5 秒后重启...")
            time.sleep(5)


if __name__ == "__main__":
    main()
