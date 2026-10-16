# AGENT-CHANGELOG — tg-station 改动记录（给 AI agent 读的单一入口）

> 用途：任何 agent（GPT / Claude / Hermes / Codex …）接手本项目时，**先读本文件**，
> 就能知道：项目文件各管什么、接口有哪些、最近改了什么、为什么改、还有什么坑没填。
> 不需要通读全部源码（10k+ 行 Go + 222KB 单文件前端）。
>
> 维护规则（硬性）：**任何代码/配置/前端改动，必须在同一个提交里更新本文件的「三、改动日志」**，
> 一行改动一次记录，写清文件与行号/函数名。README / DESIGN.md / PRODUCT.md 只管设计与产品，
> 不承担变更追踪职责。

最后更新：2026-10-16

---

## 一、文件职责速览

### 后端（Go，package main，无框架，全部 at 仓库根目录）

| 文件 | 职责 | 关键入口 |
|---|---|---|
| `main.go` | 启动、静态目录、媒体路由注册、单实例与配置加载 | `main()`、路由见 main.go:237-243 |
| `api.go` | 控制类 API（状态/监听开关/进度/取消/设置/刷新/扫码登录/频道/日志） | `registerAPIRoutes()` api.go:20 |
| `downloader.go` | TG 下载器进程生命周期：探测、启动、停止、重启、日志环形缓冲 | `Downloader`、`newRingBuffer()`、`status()` |
| `scan.go` | 媒体目录扫描调度器（去抖 + 串行保护），扫描 `download_dir` | `newScanScheduler()`、`scanFiles()` scan.go:161 |
| `index.go` | 内存索引（按日期/目录分组）、索引落盘与恢复 | `rebuildIndex()`、`filesForDate()`、`loadIndex()` |
| `thumbs.go` | 缩略图生成（ffmpeg + 原子提交 + 缓存 key 含 mtime/size） | `renderThumbAtomic()` thumbs.go:219 |
| `probe.go` | ffprobe/ffmpeg 探测视频时长（带缓存） | `probeDuration()` probe.go:225 |
| `pikpak.go` | PikPak 网盘中转后端：配置、登录（token/扫码）、配额、任务、`/api/pikpak/*` 路由 | `registerPikpakRoutes()` pikpak.go:904 |
| `netdisk.go` | 「网盘 · 同一批」筛选（目录 + 文件名前缀 → 同批文件） | `netdiskFilter()`、`netdiskHandler()` netdisk.go:126 |
| `settings.go` | `config.json` 读写（含默认值与向后兼容 key 映射） | `loadConfig()`、`saveConfig()` |
| `pathsafe*.go` | 路径穿越防护（跨平台实现：`pathsafe_windows.go` / `pathsafe_other.go`） | 所有文件/缩略图/删除接口都过它 |
| `proc_windows.go` / `proc_other.go` | 进程枚举与探测的平台封装 | 供 downloader/pikpak 复用 |
| `console_windows.go` / `console_other.go` | Windows 下隐藏控制台窗口 | — |
| `stationlog.go` | 站内日志汇集（环形缓冲 + 文件轮转），转发两个下载器的输出 | `initStationLog()` |
| `logincallbacks` 相关 | 扫码登录子进程管理（见 downloader.go / pikpak.go） | `reapQr` 逻辑 |

### 前端（`static/`，一功能一套独立文件，禁止塞进 index.html）

| 文件 | 职责 |
|---|---|
| `static/index.html` | 主页面（222KB，含大量内联 UI 与主逻辑，第 1080 行起是主 `<script>`） |
| `static/datejump.js` / `.css` | 点吸顶日期条 → 日期快速跳转（单日模式，可返回时间线）；桥 `window.DateJumpBridge` |
| `static/longpress.js` / `.css` | 长按卡片 → 快进/倒退（真倍速递增）；样式类 `.lp-badge` |
| `static/netdisk.js` | 网盘批次面板前端 |
| `static/manifest.json` / `icon*.png` / `icon.svg` | PWA 安装到主屏 |

### 脚本 / 部署

| 文件 | 职责 |
|---|---|
| `tg_downloader*` / `tg_auto_download.py` | TG 下载器（PyInstaller 打包，独立于站，开机自启） |
| `pikpak_downloader.py` / `pikpak_downloader_cli.spec` | PikPak 中转下载器 |
| `_deploy.ps1` / `_start.ps1` / `build_*.ps1` | 构建与部署脚本 |
| `pi_deploy/`、`tg_station_linux_arm64` | 树莓派 arm64 静态编译产物（用户说「上传」才部署） |
| `config.json` / `config.pi.json` | 本机 / 树莓派配置（公开快照仓库里必须用脱敏版） |
| `_secrets/` | 凭据，已在 .gitignore，**禁止提交** |

### 远端与真源

- 真源：`E:\zm\mtllq`（本目录）；双远端 `origin=github.com/xiaoluhh/tg-station`、`gitee=git@gitee.com:lovehgt/tg-station`（均私有）。
- 对外分享快照：`xiaoluhh/tg-media-panels`（PUBLIC，含 media-browser/ 与 tg-station/，更新时只覆盖 tg-station/，config 保留脱敏版）。
- 线上运行目录：`E:\zm\tgstation`（改动需热替换部署），日志 `E:\zm\tgstation\tg_station.log`。

---

## 二、HTTP 接口清单

媒体路由（main.go:237-243）：
`GET /`、`GET /api/dates`、`GET /api/files/{date|dir}`、`GET /api/file/{path}`、`POST /api/delete`、`GET /thumb/{path}`、`GET /static/*`

控制路由（api.go:21-34）：
`/api/status`、`/api/listen/{start,stop,restart,status}`、`/api/download/{progress,cancel}`、
`/api/settings`、`/api/refresh`、`/api/login/{qr,status}`、`/api/channels`、`/api/netdisk`、`/api/logs`

PikPak 路由（pikpak.go:904-915）：
`/api/pikpak/{status,start,stop,restart,login,qr,qr/status,logout,quota,files,cancel,scan}`

> ⚠️ **鉴权现状：所有写接口（含 `/api/delete`、`/api/listen/*`、`/api/settings`、`/api/pikpak/*`）无任何鉴权**。
> 服务绑 `0.0.0.0`，一旦经 cpolar 暴露公网，任何拿到地址的人都能删文件、停监听、改配置。**未决（用户未拍板）。**

---

## 三、改动日志（新 → 旧）

每条格式：`日期 · 提交 · 类型 · 文件 · 改了什么 · 为什么 · 影响/遗留`。更早历史见 `git log`。

- 2026-10-16 · `docs` · 新增本文件 `AGENT-CHANGELOG.md`：建立「单一入口」变更记录，要求每次改动实时更新，避免其他 agent 通读源码。 · 用户明确要求。
- 2026-10-15 · `3c6885a` · docs · `DESIGN.md`：把已删掉的旧样式类 `.vc-fast` 引用改为 `.lp-badge`（长按快进提示），保持设计文档与实际类名一致。 · 避免 agent 按文档写出失效类名。
- 2026-10-14 · `596f79b` · feat · `static/longpress.js|css`（新增）、`pikpak.go`、`netdisk.go`、`static/netdisk.js`、`pikpak_downloader*.py`、`config.json`：①长按卡片快进/倒退，倍数真实递增（不是跳变）；②PikPak 网盘中转（云端 5-6GB 当循环中转仓，下载→校验→`batchDelete` 永久删云端，回收站也占配额）；③网盘「同一批」筛选（目录 + 文件名前缀）。 · 用户需求；PikPak 非官方接口、国内直连被 RST，必须走 Clash 7897。 · 遗留：中转方案细节见 `pikpak-unofficial-api.md` / `pikpak-qr-login-research.md`，PikPak 相关「待用户拍板」项见第四节。
- 2026-10-14 · `api.go.bak_20261014_2357` 保留的备份文件：改动前的 api.go 原样快照，供回滚（非必需，可清理）。
- 2026-10-12 · `56ac208` · feat(ui) · `static/datejump.js|css`（新增）、`static/index.html`：点吸顶日期条按日期快速跳转；单日模式下可返回时间线。桥接用 `window.DateJumpBridge`。 · 用户要求快速定位某天。
- 2026-10-12 · `1ef2c70` · fix(media) · `main.go`（file/delete/thumb 三个 handler）：只放行图片与视频后缀，其他类型拒绝。 · 收紧攻击面（配合无鉴权现状）。
- 2026-10-12 · `bce2e08` · fix(listen) · `api.go`（`listenStartHandler`）：加 `sessionMu`，杜绝连点两下启动两个下载器。 · 双实例共用 TG session 会被判定「异地同时使用」→ 作废登录。
- 2026-10-12 · `a9b3c77` · fix(login) · `downloader.go` / `pikpak.go`（`reapQr`）：扫码子进程收尸改成单一 `Wait` + `done` 通道，kill 与启动共用同一把锁。 · 原逻辑多 Wait 会误判/漏杀。
- 2026-10-11 · `45b048b` · chore · `.gitignore`：补 `*.session` / `*.session-journal`。 · 防 Telethon 会话文件（含登录凭据）误入库。
- 2026-10-11 · `414b5b4` · fix(deploy) · `build_*.ps1`、`*.spec`：补齐 P1 的 CLI exe 部署与 spec 整理。 · P1+P2 合流时漏了这两个文件，导致部署回退成 mihomo 旧版。
- 2026-10-10 · `9737c6c` · fix(downloader) · 下载器：显式把标准流设为 UTF-8。 · PyInstaller 打包版忽略 `PYTHONUTF8`，中文日志乱码/崩溃。
- 2026-10-10 · `c0eda80` · chore · `.gitignore`：`_secrets/` 禁止提交。
- 2026-10-10 · `9c45c60` · merge(P1+P2) · 后端硬化 + 前端性能优化合流，与 `tg-media-panels` 快照对齐。
- 更早 · `ced645e` 等 · `DESIGN.md` 图标系统（19 个手绘线性图标 + Drawn-Never-Typed 规则）。

### 项目级长期约束（改前端/功能前必读，来自用户明确要求）

1. 一个功能一套独立文件（后端独立 `.go` + 前端独立 `.js`/`.css`），**禁止塞进 `index.html` 或 `api.go`**。
2. 照片/视频筛选按钮不能删不能替换；影视 = 当前页筛选按钮（`data-filter=movie`），不做独立页面（v2.1.0 独立 `/movie` 页已被否）。
3. 视频播放方式**改动前必须先问清用户预期**——历史上「<100MB 卡片内嵌播放」和「新标签直链播放」两次都被判错并 revert（现回退到 9b711b8 行为）。
4. 影视阈值 `movie_min_size_gb=0.5`（500MB），改配置需重启 tg_station。
5. 下载中断必须断点续传，不能整文件重下。
6. 下载器与站**彻底解耦**：按进程名检测/控制，禁止起孤儿进程（双实例 = session 作废）。
7. 改前端文件不需要重启 tg_station；改后端/配置需要。
8. 调试网页时**必须先静音所有 media**（用户会立刻纠正「调试别播放声音」）。

---

## 四、当前未决 / 待办

| 项 | 状态 |
|---|---|
| 所有写接口无鉴权（绑 0.0.0.0 + cpolar 公网可被任意删除/停服务） | 用户未拍板 |
| PikPak 云端中转方案（pikpak_transfer.py 断点续传 + 永久删云端循环腾空间） | 方案已出，待用户拍板开工 |
| 树莓派 arm64 部署包 `pi_deploy/` | 已交叉编译，等用户说「上传」才部署 |
| 分页 OCR 工作区 `E:\zm\book_ocr_work`（刑法典，非本项目） | 未跑完 |

---

## 五、维护规则（每个 agent 都必须遵守）

1. **改动即记录**：动这个仓库里任何代码/配置/前端文件，**同一提交内**在第三节顶部追加一条（倒序，最新在最上面）。
2. 每条记录必须含：日期、提交号（提交前可写 `pending`，提交后补上）、类型（feat/fix/docs/chore/refactor/perf）、涉及文件（含函数名或行号）、改了什么、为什么、影响面与遗留问题。
3. 新增文件 / 新接口 / 新的用户约束 → 同时更新第一、二节或「项目级长期约束」。
4. 未验证的改动必须显式标注「未验证」或「未执行」，禁止把没跑过的步骤写成已完成。
5. 记录用中文，保留原有专有名词（接口路径、函数名、配置 key）不翻译。
