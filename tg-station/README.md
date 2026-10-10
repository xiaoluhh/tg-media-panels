# TG Station - 下载 + 媒体 一体机

Telegram 转发内容自动下载 + 本地媒体浏览。三个进程各司其职：

| 进程 | 来源 | 职责 |
|------|------|------|
| `tg_station.exe` | Go（本仓库根目录） | HTTP 媒体浏览服务（:8080）、媒体索引、缩略图、网页设置面板 |
| `tg_downloader.exe` | Python + Telethon（[tg_auto_download.py](tg_auto_download.py) PyInstaller 打包） | 监听 TG、SQLite 去重、并发下载、断点续传、失败重试 |
| Clash Verge | 用户自行安装常开 | 代理出口 127.0.0.1:7897（**下载器不再内置/自启 mihomo**，2026-10-09 起） |

## 功能特性

- 媒体浏览：按日期分组（倒序）、照片/视频/影视筛选、视频缩略图（需 FFmpeg）
- 视频播放：自定义控制条、±10 秒、1x/1.5x/2x/3x 倍速、全屏、**影视续播进度**
- 下载器管理：网页一键 开始/停止 监听，按进程名检测与复用（**绝不重复启动，防双实例共用 session 被封号**）
- 扫码登录：设置面板一键生成二维码，手机 Telegram 扫码（依赖 `pip install qrcode[pil]`）
- 智能命名：按 [mz.txt](mz.txt) 名单把 caption 匹配成 `作品-角色`，否则回退 caption
- 去重与续传：SQLite 按「原始频道ID + 原始消息ID」去重；断连重试并保留残片续传
- 补扫：启动时补扫今天（或 `catchup_since` 指定时间）起转发的消息，下过的自动跳过
- 空闲自停：连续 30 分钟无新消息且无活跃下载时自动停止监听（防封号/省资源）
- 下载完成自动刷新媒体索引（下载器 POST `/api/refresh` 通知后端）
- F5 刷新页面自动比对文件数，有变化才重载列表并补生成缩略图
- 后台运行：默认隐藏控制台窗口（`-console` 可保留）

## 快速开始

### 部署（Windows）

1. 把 `tg_station.exe`、`tg_downloader.exe`、`config.json`、`mz.txt`、`ffmpeg.exe` 放同一目录（当前生产目录 `E:\zm\tgstation`）
2. 双击 `tg_station.exe` 运行；下载器可独立开机自启（`Downloader.cmd`），两侧互不影响
3. 浏览器访问 http://localhost:8080
4. 页面右上角 ⚙️ 设置 → 填 Telegram API / 下载参数 / 频道 → 保存
5. 点「▶ 开始监听」，或设置中勾选「程序启动时自动开始监听」

### 命令行参数（tg_station.exe）

| 参数 | 说明 | 默认值 |
|------|------|--------|
| -dir | 媒体目录（覆盖配置 download_dir） | 配置值 |
| -port | HTTP 端口 | `media_browser.port` / 8080 |
| -config | 配置文件路径 | 程序目录/config.json |
| -console | 保留控制台窗口（默认隐藏） | 隐藏 |
| -index | 索引文件路径 | 程序目录/media_index.json |
| -thumbs | 缩略图缓存目录 | 程序目录/thumbs |

### 下载器子模式（tg_auto_download.py / tg_downloader.exe）

| 参数 | 说明 |
|------|------|
| （无参数） | 正常监听下载；异常 5 秒后自动重启 |
| `--qr-only` | 只做扫码登录，输出一行 JSON（供网页扫码）；二维码 120 秒有效 |
| `--check-session` | 本地判断是否已登录（读 `<session>.authorized` 标记，不连网） |
| `--list-channels` | 输出账号已加入的频道列表 JSON（设置面板频道选择器） |

## API

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/status | GET | 综合状态（下载器 + 媒体索引 + 影视阈值） |
| /api/listen/start | POST | 启动下载器 |
| /api/listen/stop | POST | 停止下载器（按进程名杀，含进程树） |
| /api/listen/restart | POST | 重启下载器 |
| /api/listen/status | GET | 下载器状态 + 最近 100 行日志 |
| /api/download/progress | GET | 进行中任务（progress.json）+ 最近完成（dl_status.json） |
| /api/download/cancel | POST | 取消单个任务（写 cancel.json，下载器每段检查） |
| /api/settings | GET/POST | 读写配置（POST 会深度合并，运行中自动重启下载器） |
| /api/refresh | POST | 刷新媒体索引（下载完成回调） |
| /api/login/qr | GET | 扫码登录：生成二维码 / 检查已登录 |
| /api/login/status | GET | 登录状态（本地判断，毫秒级） |
| /api/namelist | GET/POST | 名单文件（mz.txt）读写 |
| /api/channels | GET | 账号已加入频道列表（需先停止监听） |
| /api/dates、/api/files/{date}、/api/file/、/api/delete、/thumb/ | GET/POST | 媒体索引、原图/视频、缩略图、批量删除 |

> 代理 / 订阅相关接口（`/api/proxy`、`/api/subscription`、`/api/mihomo/status`）已于 2026-10-10 随
> 「代理改由 Clash Verge 独立管理」一并移除；换节点/更新订阅请在 Clash Verge 客户端操作。

## 配置说明（config.json）

Go 与 Python **共用一份配置**：Go 启动下载器时注入环境变量 `TG_STATION_CONFIG` 指向同一文件。
Go 端只解析自己需要的字段，其余（api_id、channels、proxy 等）原样透传、不改动。

```json
"station": {
  "python": "",         // Python 解释器路径，空=自动探测
  "script": "",         // 下载器脚本路径，空=程序目录/tg_auto_download.py
  "downloader_exe": "", // 打包后的下载器 exe，空=程序目录/tg_downloader.exe
  "media_dir": "",      // 媒体目录，空=download_dir
  "auto_start": false   // tg_station 启动时自动拉起下载器
}
```

下载器侧常用字段：

| 字段 | 说明 |
|------|------|
| download_dir | 媒体落盘根目录（实际部署为 `H:\tg`） |
| channels | 监听频道（ID 或标题，支持子串匹配） |
| proxy | `{"type":"socks5","host":"127.0.0.1","port":7897}`（Clash Verge） |
| only_forwarded | 仅下载转发来源非空的消息 |
| max_file_size_mb | 单文件上限，超出跳过（0=不限） |
| concurrency | 并发下载数（上限 8） |
| idle_stop_minutes | 空闲自动停止分钟数（0=不启用） |
| catchup_today / catchup_since | 启动补扫开关 / 起始时间（空=今天 0 点） |
| movie_min_minutes | 网页「🎬影视」筛选阈值：视频时长下限（分钟，默认 5） |
| types | 收录类型：photo / video / document |

## 运行数据文件（与 exe 同目录）

| 文件 | 说明 |
|------|------|
| downloads.db | SQLite：任务去重与状态（pending/downloading/done/failed/skipped/cancelled） |
| media_index.json | Go 媒体索引（扫描生成） |
| thumbs/ | 缩略图缓存（<文件名去斜杠>.thumb.jpg） |
| progress.json | 进行中下载进度（前端轮询） |
| dl_status.json | 最近 10 条下载完成记录 |
| cancel.json | 取消任务名单（Go 写、下载器读） |
| pause.json | 暂停标志（保留语义） |
| tg_station.log | 下载器日志（Go 侧日志走控制台/重定向） |
| tg_session* | Telethon session；`<session>.authorized` 为扫码成功标记 |

## 涉及的设计要点

- **进程解耦**：下载器是独立进程，`tg_station` 只用 `tasklist` 按进程名 `tg_downloader.exe`
  探测/杀进程，绝不重复拉起；`Stop()` 走 `taskkill /T /F` 杀净进程树（PyInstaller 双进程）。
- **断点续传**：Telethon `download_media` 不支持 offset，改用 `iter_download(offset=已有大小)`
  从残片继续；失败清理按 `msg.id_` 前缀删同目录残片。
- **补扩展名**：图包/长 caption 场景 Telethon 可能补不出扩展名，按文件头 `sniff_media_ext()`
  补 `.jpg/.png/.mp4/.mkv` 等，否则「下好了但页面看不见」。
- **原子写**：配置与索引均先写 `.tmp` 再 `os.Rename`/`os.replace`，避免半截文件。
- **平台代码分离**：控制台隐藏与子进程 `CREATE_NO_WINDOW` 用 `//go:build windows` 分文件，
  `ShowWindow` 必须从 user32.dll 取（用 kernel32 会在无控制台环境 panic）。

## 依赖与编译

```bash
# 依赖：视频缩略图需 ffmpeg.exe；下载器需 Python 3 + telethon + qrcode[pil]
go build -o tg_station.exe .                        # Go 服务
pyinstaller --noconfirm tg_downloader.spec          # 下载器（产出 dist/tg_downloader.exe）
```

树莓派（ARM64 Linux）见 [pi_deploy/](pi_deploy)、[media-browser.service](media-browser.service)、
[config.pi.json](config.pi.json)。
