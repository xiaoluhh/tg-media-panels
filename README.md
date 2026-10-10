# TG 媒体面板源码（media-browser + tg-station）

> **想一次读完所有源码？** 看仓库根目录的 [`SOURCE-ALL.md`](SOURCE-ALL.md) ——
> 里面把 `tg-station/` 的全部源码、测试、前端与构建脚本按顺序拼成了一个文件，
> 不需要逐个打开子目录（GitHub 子目录页面/子 README 有时会 503）。
> 原始文件直链形如 `https://raw.githubusercontent.com/xiaoluhh/tg-media-panels/main/tg-station/main.go`。


本仓库是同一台机器上两个 **Go + 原生前端** 项目的源码快照，用于代码阅读与讨论。

- 语言：Go（仅标准库，无第三方 Web 框架）
- 前端：原生 HTML / CSS / JavaScript，全部写在单个 `static/index.html` 里（无构建步骤）
- 视频缩略图：调用 `ffmpeg.exe`（未入库，需自备）
- TG 下载：Python + Telethon（`tg_auto_download.py`，用 PyInstaller 打包成独立 exe）

---

## 目录结构

```
media-browser/    独立版「媒体浏览器」（早期项目，功能单一）
tg-station/       TG 中转站（当前主力项目，功能完整）
```

### media-browser/ —— 独立媒体浏览器

扫描本地媒体目录 → 生成缩略图 → 网页展示浏览。

| 文件 | 说明 |
|---|---|
| `main.go` | 后端全部逻辑：HTTP 服务、目录扫描、缩略图生成、索引、分页 API |
| `static/index.html` | 前端全部逻辑（HTML + CSS + JS 单文件） |
| `media-browser.service` | Linux systemd 服务文件 |
| `go.mod` / `go.sum` | 依赖（基本是 Go 标准库） |

### tg-station/ —— TG 中转站（功能更完整）

Telegram 频道媒体下载 + 网页浏览 + 影视筛选 + 播放器 + 设置页，手机浏览器访问。

| 文件 | 说明 |
|---|---|
| `main.go` | 程序入口：路由注册、静态资源、页面渲染、面板逻辑 |
| `api.go` | 全部 HTTP API：列表 / 删除 / 停止监听 / 状态统计等 |
| `downloader.go` | 下载器进程管理（启动 / 停止 / 状态检测，按进程名复用防双实例） |
| `settings.go` | 配置读写（映射 `tg_auto_download.py` 的 config 字段） |
| `static/index.html` | 前端全部逻辑（单文件，约 150KB：列表、筛选、灯箱、播放器、设置页） |
| `tg_auto_download.py` | Telethon 自动下载器：扫码登录、断点续传、空闲自动停止、失败重试 |
| `tg_downloader.spec` | PyInstaller 打包配置 |
| `config.json` | 本机配置示例（**已脱敏**） |
| `config.pi.json` | 树莓派（ARM64）配置示例（**已脱敏**） |
| `_start.ps1` / `_deploy.ps1` | Windows 本机启动 / 交叉编译并部署到树莓派的脚本 |
| `console_*.go`、`proc_*.go` | Windows / 非 Windows 的平台差异（隐藏控制台窗口、进程检测） |
| `media-browser.service` | 树莓派 systemd 服务文件 |
| `README.md` | 项目详细说明：API 列表、配置字段含义、排障 |
| `pathsafe*.go`、`index.go`、`probe.go`、`scan.go`、`thumbs.go` | P2 媒体服务硬化：统一路径校验（含 Windows junction）、日期索引、ffprobe 探测、扫描调度、缩略图缓存 |
| `pathsafe_test.go`、`scan_sched_test.go`、`index_test.go`、`thumbs_test.go`、`probe_test.go`、`media_handlers_test.go` | 上述模块的用例（含越界请求、并发调度） |

---

## 媒体服务与前端性能硬化（2026-10）

四类问题的处理，细节见 [tg-station/README.md](tg-station/README.md) 的「媒体服务硬化（P2）」与
「前端网格：性能与显示约定」两节：

- **路径安全**：所有媒体访问统一走 `safeMediaPath`（用 `filepath.Rel` 判目录边界，不用字符串前缀），
  Windows 上自行解析 junction —— 原先「媒体目录里放一个指向别处的 junction」可以绕过校验。
- **刷新不丢**：扫描期间收到的刷新请求合并为下一轮补扫（原先被静默丢弃，新下载的文件进不了索引）。
- **查询变快**：新增日期索引，`/api/files/{date}` 从线性扫描改为直接查表。
- **缩略图**：缓存键改为「相对路径 + 大小 + mtime」的哈希（旧命名会碰撞）、临时文件校验后原子提交、
  统一超时；并把响应改为长缓存（前端 URL 带 `?s=<文件大小>`），滚动时不再每格回源。
- **前端**：万级格子页面改为「块级离屏回收」+ 事件委托，常驻 DOM 从 1.6 万格降到 1~2 千格；
  视频格子右下角只留一个时长角标，缩略图上不铺任何压暗层。

---

## 两个项目的关系

早期用 `media-browser` 单独浏览媒体；后来写 `tg-station` 把「下载 + 浏览 + 影视筛选 + 播放」整合到一起，`media-browser` 的浏览能力被并入 `tg-station` 并作为独立二进制复用于树莓派。

- 想看**最初、最简单的浏览实现** → `media-browser/main.go`
- 想看**当前实际在用的完整系统** → `tg-station/`（先看 `main.go` → `api.go` → `static/index.html`）

## 阅读建议顺序

1. `tg-station/README.md`（API 与配置概览）
2. `tg-station/main.go`（入口、路由）
3. `tg-station/api.go`（接口实现）
4. `tg-station/downloader.go` + `tg-station/tg_auto_download.py`（下载链路）
5. `tg-station/static/index.html`（前端，搜索 `fetch(` 找接口调用的位置）

## 关于本仓库的脱敏说明

`config.json` / `config.pi.json` 里的 `api_id`、`api_hash`、`channels`（频道 ID）、`session` 路径、下载目录已替换为占位符，仅保留结构与字段含义，**不能直接运行**。真实凭据只存在于本机部署目录，不进入任何公开仓库。

## 许可

私人项目源码快照，仅供阅读参考。
