# TG Station - 下载 + 媒体 一体机

Telegram 转发内容自动下载 + 本地媒体浏览。三个进程各司其职：

| 进程 | 来源 | 职责 |
|------|------|------|
| `tg_station.exe` | Go（本仓库根目录） | HTTP 媒体浏览服务（:8080）、媒体索引、缩略图、网页设置面板 |
| `tg_downloader.exe` | Python + Telethon（[tg_auto_download.py](tg_auto_download.py) PyInstaller 打包，`console=False`） | 监听 TG、SQLite 去重、并发下载、断点续传、失败重试 |
| `tg_downloader_cli.exe` | 同一脚本 + [tg_downloader_cli.spec](tg_downloader_cli.spec)（`console=True`） | 只跑 `--qr-only` / `--check-session` / `--list-channels`，把 JSON 写到 stdout 供 Go 读取（扫码、登录状态、频道列表） |
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

1. 把 `tg_station.exe`、`tg_downloader.exe`、`tg_downloader_cli.exe`、`config.json`、`mz.txt`、`ffmpeg.exe` 放同一目录（当前生产目录 `E:\zm\tgstation`）
2. 双击 `tg_station.exe` 运行；下载器可独立开机自启（`Downloader.cmd`），两侧互不影响
3. 浏览器访问 http://localhost:8080
4. 页面右上角 ⚙️ 设置 → 填 Telegram API / 下载参数 / 频道 → 保存
5. 点「▶ 开始监听」，或设置中勾选「程序启动时自动开始监听」

> 三个接口（扫码 `/api/login/qr`、登录状态 `/api/login/status`、频道列表 `/api/channels`）
> 需要下载器输出 JSON。查找顺序：`station.downloader_cli` 配置 → 程序目录/`tg_downloader_cli.exe`
> → 程序目录/downloader/ → 当前目录；找不到 CLI 时自动回退到 `station.python` + `tg_auto_download.py`
> （Python 会真实执行 `--version` 验证，Microsoft Store 的 python 应用执行别名会被跳过）。
> 三个后端都不可用时，接口会返回明确的错误说明，而不是静默失败。
> 部署脚本 [_deploy.ps1](_deploy.ps1) 会自动拷贝 CLI exe（缺失时只打印提醒，功能回退到 Python）。

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

> 上面三个子模式走 **CLI 版**（`tg_downloader_cli.exe`，`console=True`）才能保证 stdout 上
> 的 JSON 协议稳定可用；常驻监听用 `tg_downloader.exe`（`console=False`，不弹黑窗）。
> Go 侧统一用 `CREATE_NO_WINDOW` 启动子进程，所以 CLI 版也不会出现黑窗。
> 没有 CLI 可执行文件时回退到 `python tg_auto_download.py <子模式>`。

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
| /api/settings | GET/POST | 读写配置（POST 会深度合并；原本有下载器在跑时先停旧实例再用新配置启动，重启失败会在响应里带 `restartError`/`warning`） |
| /api/refresh | POST | 刷新媒体索引（下载完成回调） |
| /api/login/qr | GET | 扫码登录：生成二维码 / 检查已登录 |
| /api/login/status | GET | 登录状态（本地判断，毫秒级） |
| /api/namelist | GET/POST | 名单文件（mz.txt）读写 |
| /api/channels | GET | 账号已加入频道列表（需先停止监听） |
| /api/dates、/api/files/{date}、/api/file/、/api/delete、/thumb/ | GET/POST | 媒体索引、原图/视频、缩略图、批量删除 |

> `/api/files/{date}` 分页：每页 100 条。`page` 省略 = 第 1 页；**非整数 / 0 / 负数 / 超过 1e7
> 一律返回 HTTP 400 + JSON `{"ok":false,"error":...}`**（旧实现会因负的切片起点 panic）；
> 页码超出实际页数时返回空数组（200），不会越界。

> 代理 / 订阅相关接口（`/api/proxy`、`/api/subscription`、`/api/mihomo/status`）已于 2026-10-10 随
> 「代理改由 Clash Verge 独立管理」一并移除；换节点/更新订阅请在 Clash Verge 客户端操作。

## 配置说明（config.json）

Go 与 Python **共用一份配置**：Go 启动下载器时注入环境变量 `TG_STATION_CONFIG` 指向同一文件。
Go 端只解析自己需要的字段，其余（api_id、channels、proxy 等）原样透传、不改动。

```json
"station": {
  "python": "",         // Python 解释器路径（回退模式），空=自动探测（候选会实际执行 --version 验证）
  "script": "",         // 下载器脚本路径，空=程序目录/tg_auto_download.py
  "downloader_exe": "", // 常驻监听用的后台下载器 exe，空=程序目录/tg_downloader.exe
  "downloader_cli": "", // 扫码/登录检查/频道列表用的 CLI 下载器 exe，空=程序目录/tg_downloader_cli.exe
  "media_dir": "",      // 媒体目录，空=download_dir
  "auto_start": false   // tg_station 启动时自动拉起下载器
}
```

> `python` 只在没有 CLI 可执行文件时使用。自动探测顺序：PATH（python / python3 / py 启动器）
> → 常见安装目录；每个候选都执行 `--version` 验证，失败/超时/命中 Microsoft Store
> 应用执行别名就继续下一个；若显式配置的路径无效，接口会报出「配置位置 + 修复方式」而不是静默使用。

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

- **进程解耦**：下载器是独立进程，`tg_station` 用平台实现探测/杀进程（Windows：`tasklist` /
  `taskkill /T /F` 按进程名 `tg_downloader.exe`；Linux：扫 `/proc/*/cmdline`，命中
  `tg_downloader` 或 `tg_auto_download.py`，先 SIGTERM 再 SIGKILL），**绝不重复拉起**。
- **统一运行状态**：`downloader.isRunning()` 同时看「Go 托管的子进程」与「外部独立进程」，
  `/api/status`、启动/停止/重启、扫码、频道加载全用它；记录的 `running` 但进程已消失
  会如实报 `exited`，`Restart()` 也按真实状态决定是否重启。
- **子进程调用统一构建**：`dlCommand` + `buildBackends`（CLI exe 优先，python 回退）统一负责
  可执行文件定位、参数、工作目录与 `TG_STATION_CONFIG` / `PYTHONUTF8` / `PYTHONIOENCODING`
  环境变量；一次性子模式带超时（登录检查 15s、频道列表 60s）并处理非零退出码、空输出、非法 JSON。
- **扫码不阻塞**：读到第一行 JSON（二维码 base64）立刻返回前端，进程继续挂起等待扫码；
  重新生成二维码时先结束上一个等待中的进程（按进程树杀，兼容 PyInstaller 双进程）。
- **session 独占**：扫码/频道列表前一律检查统一下载器状态，运行中直接返回 409，
  避免第二个进程同时读同一个 Telethon session。
- **断点续传**：Telethon `download_media` 不支持 offset，改用 `iter_download(offset=已有大小)`
  从残片继续；失败清理按 `msg.id_` 前缀删同目录残片。
- **补扩展名**：图包/长 caption 场景 Telethon 可能补不出扩展名，按文件头 `sniff_media_ext()`
  补 `.jpg/.png/.mp4/.mkv` 等，否则「下好了但页面看不见」。
- **原子写**：配置与索引均先写 `.tmp` 再 `os.Rename`/`os.replace`，避免半截文件。
- **平台代码分离**：控制台隐藏、子进程 `CREATE_NO_WINDOW`、进程探测/结束都按
  `//go:build windows` / `!windows` 分文件（`proc_windows.go` / `proc_other.go`），
  不把 Windows 命令当通用实现；`ShowWindow` 必须从 user32.dll 取（用 kernel32 会在无控制台环境 panic）。

## 测试

```bash
go test ./...        # 分页接口、Python 探测、运行状态、JSON 协议与超时、扫码时序
gofmt -l .           # 应为空
GOOS=linux go vet ./...   # 校验非 Windows 分支
```

- 分页用例覆盖：缺省页 / 合法页 / 非整数 / 0 / 负数 / 超过上限 / 整数溢出串 / 超出页数 / 无文件日期。
- 下载器协议用例用 `testdata/mockcli`（一个 Go 写的假 CLI）模拟正常 JSON、空输出、非零退出码、
  非法 JSON、超时与回退，**不需要 Telegram 账号，也不触碰真实 session**。
- 构建出 `dist/tg_downloader_cli.exe` 后，`TestRealCliExeOutputsJSON` 会真的驱动它并解析 stdout；
  另可用 `TG_PROBE_CLI=<exe 路径> go test -run TestProbeExeStdout -v .` 探测任意构建产物。

## 依赖与编译

```bash
# 依赖：视频缩略图需 ffmpeg.exe；下载器需 Python 3 + telethon + qrcode[pil]
go build -o tg_station.exe .                        # Go 服务
pyinstaller --noconfirm tg_downloader.spec          # 常驻下载器 → dist/tg_downloader.exe
powershell -ExecutionPolicy Bypass -File .\build_downloader_cli.ps1   # CLI 版 → dist/tg_downloader_cli.exe
```

树莓派（ARM64 Linux）见 [pi_deploy/](pi_deploy)、[media-browser.service](media-browser.service)、
[config.pi.json](config.pi.json)。
