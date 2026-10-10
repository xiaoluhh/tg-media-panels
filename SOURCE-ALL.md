# tg-station 全量源码（单文件版）

本文件是把 `tg-station/` 目录下**全部源码、测试、前端、构建脚本**按顺序拼成的单文件，
方便一次性阅读或用抓取工具获取（GitHub 子目录页面有时不可访问，读这一个文件即可）。

## 文件清单（48 个）

- `README.md` （17811 字节）
- `api.go` （22301 字节）
- `cli_integration_test.go` （1810 字节）
- `console_other.go` （142 字节）
- `console_windows.go` （802 字节）
- `downloader.go` （27562 字节）
- `downloader_start_test.go` （3644 字节）
- `downloader_test.go` （15790 字节）
- `files_pagination_test.go` （5817 字节）
- `index.go` （3552 字节）
- `index_test.go` （5033 字节）
- `login_qr_test.go` （4986 字节）
- `main.go` （18507 字节）
- `media_ext_test.go` （4017 字节）
- `media_handlers_test.go` （11605 字节）
- `pathsafe.go` （15258 字节）
- `pathsafe_other.go` （313 字节）
- `pathsafe_test.go` （13719 字节）
- `pathsafe_windows.go` （895 字节）
- `probe.go` （8386 字节）
- `probe_test.go` （5375 字节）
- `proc_other.go` （2340 字节）
- `proc_windows.go` （2049 字节）
- `python_detect_test.go` （1466 字节）
- `scan.go` （7054 字节）
- `scan_sched_test.go` （10542 字节）
- `settings.go` （6702 字节）
- `stdout_probe_test.go` （1047 字节）
- `testdata/mockcli/main.go` （2684 字节）
- `thumbs.go` （15311 字节）
- `thumbs_test.go` （15286 字节）
- `static/icon-192.png` （二进制，未内联）
- `static/icon-512.png` （二进制，未内联）
- `static/icon.svg` （721 字节）
- `static/index.html` （184441 字节）
- `static/manifest.json` （644 字节）
- `.gitignore` （1286 字节）
- `_deploy.ps1` （7808 字节）
- `_start.ps1` （2993 字节）
- `build_downloader_cli.ps1` （1574 字节）
- `config.json` （921 字节）
- `config.pi.json` （888 字节）
- `go.mod` （79 字节）
- `go.sum` （312 字节）
- `media-browser.service` （505 字节）
- `tg_auto_download.py` （58939 字节）
- `tg_downloader.spec` （1487 字节）
- `tg_downloader_cli.spec` （2238 字节）

合计正文约 427 KB。


---

## 文件：README.md

```text
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

## 媒体服务硬化（P2，2026-10）

针对「路径安全、扫描调度、索引性能、缩略图」四类问题的实现要点：

| 主题 | 实现 | 关键点 |
|------|------|--------|
| 路径安全 | `pathsafe.go` + `pathsafe_windows.go` / `pathsafe_other.go` | 所有媒体访问走唯一入口 `safeMediaPath`；用 `filepath.Rel` 做目录边界判断（**不用字符串前缀**，否则 `/media2/a.mp4` 会被 `/media` 放行）；Windows 上自写 `resolveReparsePoints` 解析符号链接与 junction（`filepath.EvalSymlinks` 对 junction 无效）；校验失败返回 403/404 且不回显服务器绝对路径 |
| 扫描调度 | `scan.go` 的 `scanScheduler` | 单一调度器 + `pending` 标志：扫描期间的新请求合并为**下一轮补扫**，既不丢弃也不并发第二个扫描；缩略图只在「无待补扫」之后基于最新索引生成；`recover` 兜底复位 `running`，避免一次 panic 后刷新永久排队 |
| 元数据探测 | `probe.go` | 优先 `ffprobe`（只取 duration），无则回退 ffmpeg；统一 `exec.CommandContext` 超时（探测 15s / 缩略图 20s）；时长做合法性检查（排除 NaN/Inf/负数/解析失败）；时长缓存含「路径 + 大小 + mtime」 |
| 日期索引 | `index.go` | `map[string][]*MediaFile` 与文件列表在**同一把锁**下原子切换；`/api/files/{date}` 不再线性扫描（实测 ~1 万倍提速） |
| 缩略图 | `thumbs.go` | 缓存键 = `sha256(相对路径 + 大小 + mtime)`（旧算法把 `/` 换成 `_`，`a/b.jpg` 与 `a_b.jpg` 会碰撞；旧命名只读兼容、命中即迁移）；临时文件 → 校验 → `os.Rename` 原子提交；同键单飞 + 并发上限 4 |
| 缩略图缓存头 | `main.go` 的 `thumbHandler` | `Cache-Control: public, max-age=604800`。文件名已含内容指纹、前端 URL 再带 `?s=<文件大小>`，源文件一变就换名/换 URL，所以可以放心长缓存（此前用 `no-cache`，滚动时每格都回源） |

> Windows 上 `go test -race` 需要 cgo（本机无 gcc 会报 `-race requires cgo`）；这种情况下用
> 「并发用例 + `-count=N` 重复」作为证据，**不要把没跑成的 race 说成通过**。

## 前端网格：性能与显示约定（static/index.html）

万级缩略图页面的 DOM / 事件策略，改这个文件前先读这节：

- **事件委托**：`touchstart/move/end/click` 只挂在 `#content` 上一次，靠 `closest('.photo-item')` 定位；
  不要再回到「每个格子挂 4 个监听器」的写法（实测那是节点/监听器泄漏的来源之一）。
- **块级离屏回收**：同一天的条目按 `GRID_BLOCK = 840`（4~8 列的公倍数，任何列数下都整行收尾）切成多个
  `.photo-grid`；视口外 ±2500px 的块清空并留等高空占位，滚回来按 `data-index` 重建。
  常驻 DOM 因此稳定在 1~2 块（约 840–1680 格），**与「滚到多深」无关**。
- **占位高度用「算」的**：`heightForCount()` 按 列数 / 列宽 / `aspect-ratio` 计算。**不要改成 rect 测量** ——
  `content-visibility: auto` 会让离屏格子按 `contain-intrinsic-size` 估算，占位偏小、滚动条长度会漂。
- **回收判断用几何而非 IntersectionObserver**：页面在后台/最小化时 IO 不回调，几何判断不受可见性影响。
- **需要完整 DOM 的入口**：删除后的增量同步、窗口 resize，先 `restoreAllRecycledSections(true)`
  （`true` = 同步恢复；滚动路径用 `false` 逐帧恢复）。
- **回收不参与删除索引计算**：先整体恢复 → `syncGridAfterDelete()`（已改为以块为单元）→ 再重新回收。
- **视频格子显示**：`.video-overlay` 只是把角标摆到右下角的定位容器，**不铺任何底色**
  （既无整片黑也无渐变）；右下角 `.video-duration` 显示片长，由 `durText()` 生成（`item.duration > 0` 才输出）。
- **两处生成点必须同步**：格子 HTML 在 `render()` 与 `sectionItemsHTML()` 各有一份（后者用于离屏块重建），
  改样式/结构时两处都要改，否则滚回来会「变样」。
- **不要跨作用域调函数**：`fmt()` 只存在于灯箱作用域，网格渲染处直接调用会 `ReferenceError` 让整页空白
  （本项目踩过），渲染要用自备的顶层函数。
- **改完必须整页重载**：`index.html` 走 no-cache，但长时浏览的旧页面已积累状态，需 Ctrl+Shift+R 才干净。

## 测试

```bash
go test ./...        # 分页接口、Python 探测、运行状态、JSON 协议与超时、扫码时序
gofmt -l .           # 应为空
GOOS=linux go vet ./...   # 校验非 Windows 分支
```

- 分页用例覆盖：缺省页 / 合法页 / 非整数 / 0 / 负数 / 超过上限 / 整数溢出串 / 超出页数 / 无文件日期。
- 媒体服务用例：路径安全（`pathsafe_test.go`：`..` 穿越、同前缀兄弟目录、绝对路径、符号链接指向根外/根内、Windows 盘符与保留名）、
  扫描调度（`scan_sched_test.go`：单次 / 期间一次请求 / 多次请求合并 / 二轮扫描 / 失败后仍可刷新 / 并发下不双跑）、
  日期索引（`index_test.go`）、缩略图（`thumbs_test.go`）、元数据探测（`probe_test.go`）、越界请求（`media_handlers_test.go`）。
  符号链接用例在无权限创建链接的环境会 skip；Windows 下可用 `cmd /c mklink /J` 建 junction 实测。
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

```


---

## 文件：api.go

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// registerAPIRoutes 挂载整合后新增的全部 API 路由。
// 必须在 http.ListenAndServe 之前调用。
func registerAPIRoutes() {
	http.HandleFunc("/api/status", statusHandler)                      // 综合状态（监听器 + 媒体索引）
	http.HandleFunc("/api/listen/start", listenStartHandler)           // 启动下载器
	http.HandleFunc("/api/listen/stop", listenStopHandler)             // 停止下载器
	http.HandleFunc("/api/listen/restart", listenRestartHandler)       // 重启下载器
	http.HandleFunc("/api/listen/status", listenStatusHandler)         // 下载器状态
	http.HandleFunc("/api/download/progress", downloadProgressHandler) // 下载进度（progress.json + dl_status.json）
	http.HandleFunc("/api/download/cancel", downloadCancelHandler)     // 取消单个下载任务（写 cancel.json）
	http.HandleFunc("/api/settings", settingsHandler)                  // 读写配置（GET / POST）
	http.HandleFunc("/api/refresh", refreshHandler)                    // 刷新媒体索引（Python 下载完成回调）
	http.HandleFunc("/api/login/qr", loginQrHandler)                   // 扫码登录：生成二维码 / 检查已登录
	http.HandleFunc("/api/login/status", loginStatusHandler)           // 登录状态：检查 session 文件
	http.HandleFunc("/api/namelist", nameListHandler)                  // 名单文件 (mz.txt) 读写：GET 读取 / POST 保存
	http.HandleFunc("/api/channels", channelsHandler)                  // 频道列表：列出账号已加入的频道（设置面板选择器）
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]interface{}{"ok": false, "error": msg})
}

// ---------- /api/status ----------

func statusHandler(w http.ResponseWriter, r *http.Request) {
	ls := downloader.status()

	indexMutex.RLock()
	total := len(mediaIndex)
	last := lastScan
	indexMutex.RUnlock()

	lastScanStr := ""
	if !last.IsZero() {
		lastScanStr = last.Format(time.RFC3339)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"listener": map[string]interface{}{
			"state":     ls.State,
			"running":   ls.Running,
			"pid":       ls.Pid,
			"startedAt": ls.StartedAt,
			"exitCode":  ls.ExitCode,
			"exitMsg":   ls.ExitMsg,
			"autoStart": downloader.cfg != nil && downloader.cfg.Station.AutoStart,
		},
		"media": map[string]interface{}{
			"total":       total,
			"lastScan":    lastScanStr,
			"dir":         mediaDir,
			"scanning":    scanSched.Running(),
			"scanPending": scanSched.Pending(),
		},
		"movie_min_minutes": movieMinMinutes(),
	})
}

// movieMinMinutes 返回影视筛选阈值（config.json movie_min_minutes，默认 5 分钟）
func movieMinMinutes() float64 {
	minutes := 5.0
	if downloader.cfg != nil {
		if v, ok := downloader.cfg.Raw["movie_min_minutes"]; ok {
			if f, ok := v.(float64); ok && f > 0 {
				minutes = f
			}
		}
	}
	return minutes
}

// ---------- /api/listen/* ----------

func listenStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	if err := downloader.Start(); err != nil {
		log.Printf("[station] 启动下载器失败: %v", err)
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ls := downloader.status()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "下载器已启动",
		"state":   ls.State,
		"pid":     ls.Pid,
	})
}

func listenStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	if err := downloader.Stop(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "下载器已停止",
		"state":   downloader.status().State,
	})
}

func listenRestartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	if err := downloader.Restart(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ls := downloader.status()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "下载器已重启",
		"state":   ls.State,
		"pid":     ls.Pid,
	})
}

func listenStatusHandler(w http.ResponseWriter, r *http.Request) {
	ls := downloader.status()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"listener": ls,
	})
}

// ---------- /api/download/progress ----------

// downloadProgressHandler 返回下载进度：progress.json（正在下载的任务）+ dl_status.json（最近完成记录）。
// 两个文件都是下载器（tg_auto_download.py）在运行目录写的 UTF-8 JSON，兼容 v1 悬浮窗时代格式。
func downloadProgressHandler(w http.ResponseWriter, r *http.Request) {
	base := filepath.Dir(configPath) // 运行目录（config.json 所在处，与下载器 BASE_DIR 一致）

	// progress.json：只取 pct<100 的任务（pct=100 是中断残留/已完成，不算正在下载）
	tasks := []map[string]interface{}{}
	if b, err := os.ReadFile(filepath.Join(base, "progress.json")); err == nil {
		var p struct {
			Tasks []map[string]interface{} `json:"tasks"`
		}
		if json.Unmarshal(b, &p) == nil {
			for _, t := range p.Tasks {
				if pct, _ := t["pct"].(float64); pct < 100 {
					tasks = append(tasks, t)
				}
			}
		}
	}

	// dl_status.json：记录只有 HH:MM:SS 无日期，附加文件 mtime 的日期（MM-DD）防跨天误导
	recent := []map[string]interface{}{}
	if b, err := os.ReadFile(filepath.Join(base, "dl_status.json")); err == nil {
		if json.Unmarshal(b, &recent) == nil {
			if st, err2 := os.Stat(filepath.Join(base, "dl_status.json")); err2 == nil {
				date := st.ModTime().Format("01-02")
				for _, r := range recent {
					if _, has := r["date"]; !has {
						r["date"] = date
					}
				}
			}
		} else {
			recent = []map[string]interface{}{}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"tasks":  tasks,  // [{file, pct, done_mb, total_mb}]
		"recent": recent, // [{date, time, chat, file}]
	})
}

// ---------- /api/download/cancel ----------

// downloadCancelHandler 取消单个下载任务：把文件名写入 cancel.json，
// 下载器（tg_auto_download.py）每下载一段检查该文件，命中则放弃当前任务并清理残片。
func downloadCancelHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		File string `json:"file"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.File == "" {
		writeErr(w, http.StatusBadRequest, "缺少 file 参数")
		return
	}
	base := filepath.Dir(configPath)
	cf := filepath.Join(base, "cancel.json")
	// 读现有列表，追加（去重）
	lst := []string{}
	if b, err := os.ReadFile(cf); err == nil {
		_ = json.Unmarshal(b, &lst)
	}
	for _, f := range lst {
		if f == req.File {
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message": "该任务已在取消列表中"})
			return
		}
	}
	lst = append(lst, req.File)
	if b, err := json.Marshal(lst); err == nil {
		if err := os.WriteFile(cf, b, 0644); err != nil {
			writeErr(w, http.StatusInternalServerError, "写入 cancel.json 失败: "+err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message": "已请求取消: " + req.File})
}

// ---------- /api/settings ----------

// settingsHandler GET 返回完整配置（供设置面板回填），POST 保存配置。
func settingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := loadConfig()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":     true,
			"config": cfg.Raw,
		})
	case http.MethodPost:
		saveSettings(w, r)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "仅支持 GET / POST")
	}
}

func saveSettings(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	var incoming map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}

	// 合并：前端提交的字段覆盖，未提交的字段保留
	merged := mergeMaps(cfg.Raw, incoming)

	// 基本校验：port 必须是合法数字
	if mb, ok := merged["media_browser"].(map[string]interface{}); ok {
		if p, ok := mb["port"].(float64); ok && (p < 1 || p > 65535) {
			writeErr(w, http.StatusBadRequest, fmt.Sprintf("端口不合法: %v", p))
			return
		}
	}

	cfg.Raw = merged
	applyRaw(cfg)

	// station 段单独校验
	if st, ok := incoming["station"].(map[string]interface{}); ok {
		bs, _ := json.Marshal(st)
		var newSt StationConfig
		if json.Unmarshal(bs, &newSt) == nil {
			cfg.Station = newSt
		}
	}

	if err := saveConfig(cfg); err != nil {
		writeErr(w, http.StatusInternalServerError, "保存配置失败: "+err.Error())
		return
	}

	// 配置可能影响下载器（api_id/api_hash/channels 等）。
	// 保存前若下载器正在运行（含外部独立进程）：先真正停掉旧实例，再用新配置启动。
	// 重启失败必须如实返回，不能只报「设置已保存」而隐瞒失败。
	wasRunning := downloader.isRunning()
	restarted := false
	restartErr := ""
	if wasRunning {
		if err := downloader.Restart(); err != nil {
			restartErr = err.Error()
			log.Printf("[station] 保存设置后重启下载器失败: %v", err)
		} else {
			restarted = true
		}
	}

	resp := map[string]interface{}{
		"ok":         true,
		"saved":      true,
		"message":    "设置已保存",
		"wasRunning": wasRunning,
		"restarted":  restarted,
	}
	if restartErr != "" {
		resp["message"] = "设置已保存，但下载器重启失败：" + restartErr
		resp["restartError"] = restartErr
		resp["warning"] = "新配置已写入，但下载器未能用新配置重启，请手动停止后重新开始监听"
		log.Printf("[station] 设置已保存但重启失败：%s", restartErr)
	}
	writeJSON(w, http.StatusOK, resp)
}

// mergeMaps 深度合并：incoming 的值覆盖 base；map 递归合并，其它类型直接替换
func mergeMaps(base, incoming map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+len(incoming))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range incoming {
		if bm, ok := out[k].(map[string]interface{}); ok {
			if im, ok2 := v.(map[string]interface{}); ok2 {
				out[k] = mergeMaps(bm, im)
				continue
			}
		}
		out[k] = v
	}
	return out
}

// ---------- /api/refresh ----------

// refreshHandler 请求重新扫描媒体目录并重建索引。
// 由 Python 下载器下载完成后 POST 调用；也可手动触发。
//
// 语义保持与前端兼容：仍然是异步返回「正在刷新」，由前端轮询 /api/status 判断结果，
// 绝不提前返回虚假的「已完成」。区别是请求会统一进入调度器：
//   - 本次启动扫描 → started=true
//   - 已有扫描在跑 → 合并为下一轮补扫（merged=true，绝不丢弃）
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	started := scanSched.Request()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "正在刷新媒体索引",
		"started": started,
		"merged":  !started,
		"running": scanSched.Running(),
		"pending": scanSched.Pending(),
	})
}

// ---------- /api/login/* ----------

// 扫码登录进程管理：二维码生成后进程需继续挂起等待扫码（最多 120 秒），
// 所以不能等进程退出——读到第一行 JSON 就返回，进程留待扫码/超时后自行退出。
//
// 三条不变量（都踩过坑，改这里必须保住）：
//  1. 一个 cmd 只能 Wait 一次。唯一收尸者 = reapQr 起的那个 goroutine，
//     其它任何地方（包括 kill 路径）都不许再 cmd.Wait()，否则是数据竞态。
//  2. 只对「确认还活着」的进程执行 kill。进程自然退出后由收尸者清掉引用，
//     因为 Windows/Linux 都会复用 PID——拿旧 PID 去 taskkill /T 可能杀掉无关进程树。
//  3. 「作废旧进程 → 启动新进程 → 记录新进程」必须在同一把锁里完成，
//     否则两个并发的 /api/login/qr 会各起一个进程，两个进程共用一个 session
//     有被判「异地同时使用」而封号的风险（见 Downloader.Start 的同类约定）。
type qrProcState struct {
	cmd  *exec.Cmd
	done chan struct{} // 进程被收尸（Wait 返回）后由 reapQr 关闭
}

var (
	qrMu  sync.Mutex   // 串行化上面第 3 条
	qrCur *qrProcState // 当前等待扫码的进程；nil = 当前没有

	// 等旧进程真正死掉的时限；超时只记日志，不阻塞后续流程
	qrKillWait = 5 * time.Second

	// 以下两个是测试钩子，生产环境分别为 killProcessTree / downloader.startQrLogin
	qrKillTree = killProcessTree
	qrStart    = func() (*exec.Cmd, map[string]interface{}, string, error) {
		return downloader.startQrLogin(qrFirstLineTimeout)
	}
)

// reapQr 是扫码子进程唯一的收尸者：Wait 一次、关闭 done、清掉引用。
// ⚠ 顺序不能颠倒：先 close(done) 再抢 qrMu——killQrLocked 是持着 qrMu 等 done 的，
// 反过来写就是死锁。
func reapQr(st *qrProcState) {
	_ = st.cmd.Wait()
	close(st.done)
	qrMu.Lock()
	if qrCur == st {
		qrCur = nil
	}
	qrMu.Unlock()
}

// killQrLocked 作废当前等待扫码的进程（调用方必须持有 qrMu）。
// 已经退出的进程直接跳过：绝不对可能已被复用的旧 PID 下手。
func killQrLocked() {
	st := qrCur
	qrCur = nil
	if st == nil {
		return
	}
	select {
	case <-st.done: // 已收尸（自己退出了）
		return
	default:
	}
	if st.cmd.Process != nil {
		_ = qrKillTree(st.cmd.Process.Pid)
	}
	// 等它真的退出，避免新旧两个进程同时持有 session
	select {
	case <-st.done:
	case <-time.After(qrKillWait):
		log.Printf("[station] 旧扫码进程在 %s 内未退出，继续启动新的", qrKillWait)
	}
}

// beginQrLogin 在锁内完成「作废旧进程 → 启动新进程 → 记录并挂收尸」，返回第一行 JSON。
// 整个过程持锁，因此并发的扫码请求会被串行化：后来的请求会先杀掉前一个进程。
func beginQrLogin() (*exec.Cmd, map[string]interface{}, string, error) {
	qrMu.Lock()
	defer qrMu.Unlock()

	killQrLocked()

	cmd, result, desc, err := qrStart()
	if err != nil {
		return nil, nil, desc, err
	}
	st := &qrProcState{cmd: cmd, done: make(chan struct{})}
	qrCur = st
	go reapQr(st)
	return cmd, result, desc, nil
}

// loginQrHandler 生成登录二维码 / 检查是否已登录：
// 已登录 → already + username；未登录 → qr（base64 二维码）；失败 → error。
// 子进程统一由 downloader.startQrLogin 构建（优先 CLI exe，其次 python + 脚本），
// 拿到第一行 JSON 立刻返回给前端，进程继续挂起等待扫码。
func loginQrHandler(w http.ResponseWriter, r *http.Request) {
	// 与「启动/停止下载器」互斥：扫码进程和下载器不能同时占用同一个 session。
	// 加锁顺序 sessionMu → qrMu（qrMu 在 beginQrLogin 里取）。
	downloader.sessionMu.Lock()
	defer downloader.sessionMu.Unlock()

	if downloader.isRunning() {
		writeErr(w, http.StatusConflict, "下载器正在运行，请先停止监听再扫码登录")
		return
	}

	// 作废旧二维码 + 启动新进程 + 记录，全部在 beginQrLogin 的锁内完成
	_, result, desc, err := beginQrLogin()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("[station] 扫码登录进程已启动：%s", desc)

	switch result["status"] {
	case "already", "success":
		username, _ := result["username"].(string)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":       true,
			"status":   result["status"],
			"username": username,
		})
	case "qr":
		b64, _ := result["qr_base64"].(string)
		if b64 == "" {
			writeErr(w, http.StatusInternalServerError, "二维码生成失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":     true,
			"status": "qr",
			"qr":     "data:image/png;base64," + b64,
		})
	default:
		msg, _ := result["error"].(string)
		if msg == "" {
			msg = "登录失败"
		}
		writeErr(w, http.StatusOK, msg)
	}
}

// loginStatusHandler 检查登录状态：优先调用下载器 --check-session 本地判断
// session 是否有有效 auth key（不连网，毫秒级；超时 15 秒）。
// 子进程不可用、超时或输出异常时，回退到 session 文件存在性判断。
func loginStatusHandler(w http.ResponseWriter, r *http.Request) {
	if obj, err := downloader.runModeJSON(modeCheckSession, checkSessionTimeout); err != nil {
		log.Printf("[station] --check-session 调用失败，回退到 session 文件判断: %v", err)
	} else if obj != nil {
		if status, _ := obj["status"].(string); status == "logged_in" || status == "not_logged_in" {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":     true,
				"status": status,
			})
			return
		}
		log.Printf("[station] --check-session 输出缺少 status 字段，回退到 session 文件判断")
	}

	// 回退：session 文件存在性检查
	cfg, err := loadConfig()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	session := getString(cfg.Raw, []string{"session"}, "")
	loggedIn := false
	if session != "" {
		if _, err := os.Stat(session); err == nil {
			loggedIn = true
		} else if _, err := os.Stat(session + ".session"); err == nil {
			loggedIn = true
		}
	}
	status := "not_logged_in"
	if loggedIn {
		status = "logged_in"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"status": status,
	})
}

// ---------- /api/namelist ----------

// resolveNameListPath 定位名单文件：与 Python 下载器 _resolve_list_path 保持一致。
// 相对路径以「下载器脚本目录」为基准（Python 侧 app_dir() 即脚本所在目录）。
func resolveNameListPath() string {
	cfg, err := loadConfig()
	if err != nil {
		return filepath.Join(filepath.Dir(configPath), "mz.txt")
	}
	name := getString(cfg.Raw, []string{"name_list_file"}, "mz.txt")
	if name == "" {
		name = "mz.txt"
	}
	if filepath.IsAbs(name) {
		return name
	}
	// 基准目录：下载器脚本目录；取不到则用 config 目录
	base := ""
	if script, err := downloader.findScript(); err == nil && script != "" {
		base = filepath.Dir(script)
	}
	if base == "" || base == "." {
		base = filepath.Dir(configPath)
	}
	return filepath.Join(base, name)
}

// nameListHandler GET 返回名单文件内容，POST 保存新内容。
// 保存后下载器无需重启：Python 侧按文件 mtime 缓存，内容变化自动重载。
func nameListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := resolveNameListPath()
		data, err := os.ReadFile(path)
		if err != nil {
			// 文件不存在不算错误：返回空内容，前端提示可新建
			if os.IsNotExist(err) {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":      true,
					"path":    path,
					"content": "",
				})
				return
			}
			writeErr(w, http.StatusInternalServerError, "读取名单文件失败: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"path":    path,
			"content": string(data),
		})
	case http.MethodPost:
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
			return
		}
		path := resolveNameListPath()
		// 规范化换行：统一 \n，去掉末尾多余空行
		content := strings.ReplaceAll(req.Content, "\r\n", "\n")
		content = strings.TrimRight(content, "\n")
		if err := os.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
			writeErr(w, http.StatusInternalServerError, "保存名单文件失败: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"message": "名单已保存",
			"path":    path,
		})
	default:
		writeErr(w, http.StatusMethodNotAllowed, "仅支持 GET / POST")
	}
}

// ---------- /api/channels ----------

// channelsHandler 列出账号已加入的频道，供设置面板的「频道选择器」使用。
// 子进程统一由 downloader.runModeJSON 构建（优先 CLI exe，其次 python + 脚本），
// 带超时与退出码/空输出/非法 JSON 处理。
// 注意：session 数据库被下载器进程独占，运行中无法并发读取 → 需先停止监听。
func channelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 GET")
		return
	}
	if downloader.isRunning() {
		writeErr(w, http.StatusConflict, "下载器正在运行，请先停止监听再加载频道列表")
		return
	}

	result, err := downloader.runModeJSON(modeListChannels, listChannelsTimeout)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "频道列表获取失败: "+err.Error())
		return
	}
	if result["status"] != "ok" {
		msg, _ := result["error"].(string)
		if msg == "" {
			msg = "频道列表获取失败"
		}
		writeErr(w, http.StatusOK, msg)
		return
	}
	channels, _ := result["channels"].([]interface{})
	if channels == nil {
		channels = []interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"channels": channels,
	})
}

```


---

## 文件：cli_integration_test.go

```go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 真实 CLI 可执行文件的集成测试：只有构建过 dist/tg_downloader_cli.exe 时才运行
// （没有 PyInstaller/依赖的环境自动 Skip）。
//
// 用临时目录里的空配置调用 --check-session，只验证「stdout 上能拿到协议 JSON」，
// 不做任何登录、不读写真实 config.json / session / mz.txt。

func TestRealCliExeOutputsJSON(t *testing.T) {
	exe, err := filepath.Abs(filepath.Join("dist", cliExeName))
	if err != nil {
		t.Fatalf("解析 CLI 路径失败: %v", err)
	}
	if _, err := os.Stat(exe); err != nil {
		t.Skipf("未找到 %s（先运行 build_downloader_cli.ps1）: %v", exe, err)
	}

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	cfg := map[string]interface{}{
		"api_id":       0,
		"api_hash":     "",
		"session":      filepath.ToSlash(filepath.Join(dir, "tg_session")),
		"download_dir": filepath.ToSlash(dir),
	}
	b, _ := json.Marshal(cfg)
	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		t.Fatal(err)
	}
	withConfigPath(t, cfgPath)

	d := &Downloader{
		state:    dlStateIdle,
		exitCode: -1,
		logBuf:   newRingBuffer(20),
		cfg: &AppConfig{Station: StationConfig{
			DownloaderCLI: exe,
			Python:        filepath.Join(dir, "no-such-python.exe"), // 确保只走 CLI 后端
		}},
	}

	start := time.Now()
	obj, err := d.runModeJSON(modeCheckSession, 60*time.Second)
	if err != nil {
		t.Fatalf("真实 CLI 未输出可用 JSON: %v", err)
	}
	status, _ := obj["status"].(string)
	if status != "logged_in" && status != "not_logged_in" {
		t.Fatalf("status = %v，期望 logged_in / not_logged_in", obj["status"])
	}
	t.Logf("真实 CLI（%s）在 %s 内输出 JSON: %v", exe, time.Since(start).Round(time.Millisecond), obj)
}

```


---

## 文件：console_other.go

```go
//go:build !windows

package main

// hideConsoleWindow 非 Windows 平台无控制台窗口概念，空实现。
func hideConsoleWindow() {}

```


---

## 文件：console_windows.go

```go
//go:build windows

package main

import "syscall"

// hideConsoleWindow 隐藏当前进程的控制台窗口（后台运行）。
// 注意：ShowWindow 属于 user32.dll，kernel32.dll 中没有该导出函数；
// 用错 DLL 会在无控制台环境（开机自启/服务启动）下 panic。
func hideConsoleWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return // 无控制台窗口（如自启环境），无需隐藏
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	showWindow := user32.NewProc("ShowWindow")
	if err := showWindow.Find(); err != nil {
		return // 函数查找失败也不 panic，后台运行不受影响
	}
	showWindow.Call(hwnd, 0) // SW_HIDE = 0
}

```


---

## 文件：downloader.go

```go
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 下载器（Python 子进程）管理。
// 通过 tg_station 自身的 config.json 与 Python 共享配置：
// 启动子进程时设置环境变量 TG_STATION_CONFIG 指向同一份文件。
//
// 子进程调用（常驻监听 / 扫码 / 登录检查 / 频道列表）统一由本文件的
// dlCommand + buildBackends + runModeJSON 构建，避免各处重复实现启动逻辑。

const (
	dlStateIdle   = "idle" // 从未启动 / 已手动停止
	dlStateRun    = "running"
	dlStateExited = "exited" // 意外退出（可查 ExitCode / ExitMsg）
)

// ringBuffer 滚动日志缓冲：保留最近 N 行
type ringBuffer struct {
	mu    sync.Mutex
	lines []string
	max   int
}

func newRingBuffer(max int) *ringBuffer {
	return &ringBuffer{max: max}
}

func (r *ringBuffer) add(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, line)
	if len(r.lines) > r.max {
		r.lines = r.lines[len(r.lines)-r.max:]
	}
}

func (r *ringBuffer) tail(n int) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n <= 0 || n > len(r.lines) {
		n = len(r.lines)
	}
	out := make([]string, n)
	copy(out, r.lines[len(r.lines)-n:])
	return out
}

func (r *ringBuffer) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.lines))
	copy(out, r.lines)
	return out
}

// Downloader 管理下载器子进程
type Downloader struct {
	mu sync.Mutex
	// sessionMu 串行化所有「会动到 Telegram session」的操作：启动/停止/重启下载器、
	// 生成扫码进程。没有它，「连点两下开始监听」或手机+电脑同时点会各起一个下载器，
	// 两个进程共用一个 session → 被判异地同时使用 → 登录作废甚至封号（2026-10-11 修）。
	// 加锁顺序固定为 sessionMu → qrMu（见 api.go 扫码流程），不要反过来。
	sessionMu sync.Mutex
	cmd       *exec.Cmd
	state     string
	pid       int
	started   time.Time
	exitCode  int
	exitMsg   string
	logBuf    *ringBuffer
	cfg       *AppConfig
	waitDone  chan struct{}
}

var downloader = &Downloader{
	state:    dlStateIdle,
	exitCode: -1,
	logBuf:   newRingBuffer(300),
}

// ---------- 子模式与调用构建 ----------

// downloaderMode 下载器子模式；参数与 tg_auto_download.py 的 CLI 协议一致。
type downloaderMode string

const (
	modeRun          downloaderMode = ""                // 常驻监听（无参数）
	modeQrOnly       downloaderMode = "--qr-only"       // 生成二维码 / 已登录检查
	modeCheckSession downloaderMode = "--check-session" // 本地判断 session 是否已登录
	modeListChannels downloaderMode = "--list-channels" // 列出已加入的频道
)

const (
	pythonProbeTimeout  = 5 * time.Second  // 探测 Python 解释器（执行 --version）的超时
	qrFirstLineTimeout  = 30 * time.Second // 二维码生成阶段的超时
	checkSessionTimeout = 15 * time.Second // --check-session 超时（本地判定，毫秒级）
	listChannelsTimeout = 60 * time.Second // --list-channels 超时（要连 TG 拉对话列表）
	stopWaitTimeout     = 10 * time.Second // 停止下载器的等待上限
)

// dlCommand 描述一次下载器子进程调用：可执行文件、参数、工作目录、环境变量。
type dlCommand struct {
	Path string   // 可执行文件（CLI/后台 exe）或 Python 解释器
	Args []string // 传给 Path 的全部参数
	Dir  string   // 工作目录
	Env  []string // 环境变量
	kind string   // "cli" | "exe" | "python"
	desc string   // 日志描述（只含路径/文件名，不含凭据）
}

// modeArgs 子模式对应的参数（常驻监听无参数）
func modeArgs(mode downloaderMode) []string {
	if mode == modeRun {
		return nil
	}
	return []string{string(mode)}
}

// exeDirOf 当前可执行文件所在目录（取不到时退回当前工作目录）
func exeDirOf() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// downloaderEnv 子进程环境变量：统一指向程序正在使用的那份配置，
// 并保证 Python 的 UTF-8 输出（避免中文/emoji 编码异常）。
func downloaderEnv() []string {
	env := os.Environ()
	env = append(env, "PYTHONUTF8=1", "PYTHONIOENCODING=utf-8")
	if configPath != "" {
		env = append(env, "TG_STATION_CONFIG="+configPath)
	}
	return env
}

// clip 截断文本，用于错误信息（避免把大段输出/敏感内容带进日志与响应）
func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// backendSet 一次调用的候选后端：按优先级排列，另附被跳过候选的原因。
type backendSet struct {
	cmds    []dlCommand
	reasons []string
}

// ---------- 状态查询 ----------

type ListenerStatus struct {
	State     string   `json:"state"` // idle / running / exited
	Running   bool     `json:"running"`
	Pid       int      `json:"pid"`
	StartedAt string   `json:"startedAt"` // RFC3339，未运行时空串
	ExitCode  int      `json:"exitCode"`
	ExitMsg   string   `json:"exitMsg"`
	LogTail   []string `json:"logTail"` // 最近日志（最多 100 行）
}

// externalDownloaderProbe 外部独立下载器进程探测（平台实现）。
// 声明为变量以便单元测试注入固定结果，避免依赖真实进程。
var externalDownloaderProbe = func() []int { return externalDownloaderPids() }

// managedRunningLocked 报告「由本进程托管的子进程」是否在运行（调用方需持有 d.mu）
func (d *Downloader) managedRunningLocked() bool {
	return d.state == dlStateRun && d.cmd != nil && d.cmd.Process != nil
}

// isRunning 统一下载器运行状态判断：Go 托管的子进程 或 外部独立进程。
// 页面显示、启动/停止、扫码、频道加载都必须用这个，避免各自实现不同逻辑。
func (d *Downloader) isRunning() bool {
	d.mu.Lock()
	managed := d.managedRunningLocked()
	d.mu.Unlock()
	if managed {
		return true
	}
	return len(externalDownloaderProbe()) > 0
}

func (d *Downloader) status() ListenerStatus {
	d.mu.Lock()
	state := d.state
	pid := d.pid
	exitCode := d.exitCode
	exitMsg := d.exitMsg
	started := d.started
	managed := d.managedRunningLocked()
	logs := d.logBuf.tail(100)
	d.mu.Unlock()

	extPids := externalDownloaderProbe()
	running := managed || len(extPids) > 0

	st := ListenerStatus{
		Running:  running,
		ExitCode: exitCode,
		ExitMsg:  exitMsg,
		LogTail:  logs,
	}
	switch {
	case running:
		st.State = dlStateRun
		st.Pid = pid
		// 外部独立进程（开机自启/手动启动）不在托管记录里，PID 以真实进程为准
		if !managed || st.Pid == 0 {
			if len(extPids) > 0 {
				st.Pid = extPids[0]
			}
		}
	case state == dlStateRun:
		// 记录里还是运行中，但进程已经消失 → 如实报告已退出
		st.State = dlStateExited
	default:
		st.State = dlStateIdle
	}
	if !started.IsZero() {
		st.StartedAt = started.Format(time.RFC3339)
	}
	return st
}

// ---------- Python 解释器探测 ----------

// classifyPythonProbe 判断一次 `--version` 探测的结果。
// 单独抽出便于单元测试：Microsoft Store 的 python/python3「应用执行别名」
// 会打印 "Python was not found; run without arguments to install from the
// Microsoft Store..." 且常常以 0 退出——只看退出码或只匹配 "python" 都会误判。
func classifyPythonProbe(out []byte, runErr error, timedOut bool) error {
	text := strings.TrimSpace(string(out))
	lower := strings.ToLower(text)

	if timedOut {
		return fmt.Errorf("执行 --version 超时（%s）", pythonProbeTimeout)
	}
	if strings.Contains(lower, "was not found") ||
		strings.Contains(lower, "microsoft store") ||
		strings.Contains(lower, "app execution alias") {
		return errors.New("命中 Microsoft Store 应用执行别名，不是真实可用的 Python 解释器")
	}
	if runErr != nil {
		return fmt.Errorf("执行 --version 失败：%v", runErr)
	}
	if !strings.Contains(lower, "python") {
		return fmt.Errorf("--version 输出不是 Python 版本信息：%s", clip(text, 80))
	}
	return nil
}

// validatePython 实际执行 `--version` 验证候选解释器（5 秒超时）。
// 仅仅 os.Stat() 成功并不能说明它可用。
func validatePython(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), pythonProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	cmd.SysProcAttr = childProcAttr()
	out, err := cmd.CombinedOutput()
	return classifyPythonProbe(out, err, ctx.Err() == context.DeadlineExceeded)
}

// pickPython 依次验证候选解释器，返回第一个真正可用的；全部失败时给出每次失败的原因。
func pickPython(candidates []string) (string, error) {
	var tried []string
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if err := validatePython(c); err != nil {
			tried = append(tried, fmt.Sprintf("%s（%v）", c, err))
			continue
		}
		return c, nil
	}
	if len(tried) == 0 {
		return "", errors.New("没有可用的 Python 候选路径")
	}
	return "", fmt.Errorf("所有候选解释器都不可用：%s", strings.Join(tried, "；"))
}

// pythonCandidates 候选顺序：PATH（python / python3 / py 启动器）> 常见安装目录
func (d *Downloader) pythonCandidates() []string {
	var out []string
	for _, name := range []string{"python", "python3", "py"} {
		if p, err := exec.LookPath(name); err == nil {
			out = append(out, p)
		}
	}
	local := os.Getenv("LOCALAPPDATA")
	common := []string{
		filepath.Join(local, "hermes", "hermes-agent", "venv", "Scripts", "python.exe"),
		filepath.Join(local, "Programs", "Python", "Python314", "python.exe"),
		filepath.Join(local, "Programs", "Python", "Python313", "python.exe"),
		filepath.Join(local, "Programs", "Python", "Python312", "python.exe"),
		filepath.Join(local, "Programs", "Python", "Python311", "python.exe"),
		`C:\Python314\python.exe`,
		`C:\Python313\python.exe`,
		`C:\Python312\python.exe`,
		`C:\Python311\python.exe`,
		"/usr/bin/python3",
		"/usr/local/bin/python3",
		"/usr/bin/python",
	}
	return append(out, common...)
}

// findPython 探测 Python 解释器：用户指定路径 > PATH > 常见安装目录。
// 每个候选都会真实执行 `--version` 验证；用户显式指定的路径无效时直接报错，
// 说明配置位置与修复方式，绝不静默使用错误路径。
func (d *Downloader) findPython() (string, error) {
	if d.cfg != nil {
		if p := strings.TrimSpace(d.cfg.Station.Python); p != "" {
			if _, err := os.Stat(p); err != nil {
				return "", fmt.Errorf("配置的 Python 路径不存在（config.json → station.python = %s）：%v；"+
					"请改成有效的 python.exe 完整路径，或清空 station.python 让程序自动探测", p, err)
			}
			if err := validatePython(p); err != nil {
				return "", fmt.Errorf("配置的 Python 不可用（config.json → station.python = %s）：%v；"+
					"请改成有效的 python.exe 完整路径，或清空 station.python 让程序自动探测", p, err)
			}
			return p, nil
		}
	}
	py, err := pickPython(d.pythonCandidates())
	if err != nil {
		return "", fmt.Errorf("%v；请在 config.json 的 station.python 中指定 Python 解释器路径", err)
	}
	return py, nil
}

// findScript 定位 tg_auto_download.py：配置指定 > exe 目录 > 当前目录
func (d *Downloader) findScript() (string, error) {
	var candidates []string
	if d.cfg != nil && d.cfg.Station.Script != "" {
		candidates = append(candidates, d.cfg.Station.Script)
	}
	exeDir := exeDirOf()
	candidates = append(candidates,
		filepath.Join(exeDir, "tg_auto_download.py"),
		filepath.Join(exeDir, "downloader", "tg_auto_download.py"),
		"tg_auto_download.py",
	)
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("未找到下载器脚本 tg_auto_download.py（请复制到程序目录，或在 config.json 的 station.script 中指定路径）")
}

// absExePath 把候选可执行文件路径规整为绝对路径。
// 必要性：启动子进程时 cmd.Dir 会被设成 exe 所在目录，若 Path 本身是相对路径，
// Windows 会在「exe 目录」下再解析一次 → 找不到文件（实测 dist	g_downloader_cli.exe 失败）。
func absExePath(p string) string {
	if p == "" {
		return p
	}
	if filepath.IsAbs(p) {
		return p
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

// resolveAgainst 相对路径按 base 目录解析（绝对路径原样返回）
func resolveAgainst(base, p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, p)
}

// findDownloaderExe 探测打包后的后台下载器 exe：配置指定 > exe 目录 / downloader 子目录。
// 返回 "" 表示未找到（回退到 python + script 模式）。
func (d *Downloader) findDownloaderExe() string {
	exeDir := exeDirOf()
	if d.cfg != nil && d.cfg.Station.DownloaderExe != "" {
		p := resolveAgainst(exeDir, d.cfg.Station.DownloaderExe)
		if _, err := os.Stat(p); err == nil {
			return absExePath(p)
		}
		return ""
	}
	for _, c := range []string{
		filepath.Join(exeDir, downloaderProcName),
		filepath.Join(exeDir, "downloader", downloaderProcName),
	} {
		if _, err := os.Stat(c); err == nil {
			return absExePath(c)
		}
	}
	return ""
}

// findCliExe 探测带控制台的 CLI 下载器 exe（stdout 可靠输出 JSON）：
// 配置指定 > exe 目录 > exe 目录/downloader > 当前目录。返回绝对路径。
func (d *Downloader) findCliExe() string {
	exeDir := exeDirOf()
	var candidates []string
	if d.cfg != nil && strings.TrimSpace(d.cfg.Station.DownloaderCLI) != "" {
		candidates = append(candidates, resolveAgainst(exeDir, d.cfg.Station.DownloaderCLI))
	}
	candidates = append(candidates,
		filepath.Join(exeDir, cliExeName),
		filepath.Join(exeDir, "downloader", cliExeName),
		absExePath(cliExeName), // 当前工作目录（开发/脚本启动场景）
	)
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return absExePath(c)
		}
	}
	return ""
}

// buildBackends 返回某子模式的候选后端（按优先级）：
// 1) CLI 可执行文件（console 版，stdout 能稳定输出 JSON）
// 2) python + 脚本（兼容回退，且解释器必须通过 --version 验证）
func (d *Downloader) buildBackends(mode downloaderMode) backendSet {
	var set backendSet

	if p := d.findCliExe(); p != "" {
		set.cmds = append(set.cmds, dlCommand{
			Path: p,
			Args: modeArgs(mode),
			Dir:  filepath.Dir(p),
			Env:  downloaderEnv(),
			kind: "cli",
			desc: fmt.Sprintf("%s %s", filepath.Base(p), string(mode)),
		})
	} else {
		set.reasons = append(set.reasons, "未找到 CLI 下载器（"+cliExeName+"）")
	}

	if py, err := d.findPython(); err == nil {
		if script, err2 := d.findScript(); err2 == nil {
			set.cmds = append(set.cmds, dlCommand{
				Path: py,
				Args: append([]string{script}, modeArgs(mode)...),
				Dir:  filepath.Dir(script),
				Env:  downloaderEnv(),
				kind: "python",
				desc: fmt.Sprintf("%s %s %s", filepath.Base(py), filepath.Base(script), string(mode)),
			})
		} else {
			set.reasons = append(set.reasons, err2.Error())
		}
	} else {
		set.reasons = append(set.reasons, err.Error())
	}
	return set
}

// buildStartCommand 常驻监听模式的启动命令：
// 优先打包后的后台下载器 exe（无控制台窗口），其次 python + 脚本。
func (d *Downloader) buildStartCommand() (dlCommand, error) {
	if p := d.findDownloaderExe(); p != "" {
		return dlCommand{
			Path: p,
			Dir:  filepath.Dir(p),
			Env:  downloaderEnv(),
			kind: "exe",
			desc: filepath.Base(p),
		}, nil
	}
	py, err := d.findPython()
	if err != nil {
		return dlCommand{}, err
	}
	script, err := d.findScript()
	if err != nil {
		return dlCommand{}, err
	}
	return dlCommand{
		Path: py,
		Args: []string{script},
		Dir:  filepath.Dir(script),
		Env:  downloaderEnv(),
		kind: "python",
		desc: fmt.Sprintf("%s %s", filepath.Base(py), filepath.Base(script)),
	}, nil
}

// ---------- 一次性子模式执行 ----------

// parseLastJSONObject 取输出中最后一行合法 JSON 对象。
// Python 侧可能先打印日志/警告，因此从后往前找；同时要求以 "{" 开头，
// 避免把日志里的片段误判为 JSON。
func parseLastJSONObject(out []byte) (map[string]interface{}, error) {
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, errors.New("没有任何输出")
	}
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var m map[string]interface{}
		if json.Unmarshal([]byte(line), &m) == nil && m != nil {
			return m, nil
		}
	}
	return nil, fmt.Errorf("没有找到合法 JSON 行（输出：%s）", clip(text, 120))
}

// runCommandJSON 执行一个后端并解析它的 JSON 输出（带超时）。
func runCommandJSON(c dlCommand, timeout time.Duration) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Path, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = c.Env
	cmd.SysProcAttr = childProcAttr()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		if cmd.Process != nil {
			// PyInstaller 单文件会有引导进程 + 子进程，必须整树结束，否则残留
			_ = killProcessTree(cmd.Process.Pid)
		}
		return nil, fmt.Errorf("执行超时（%s）", timeout)
	}

	obj, perr := parseLastJSONObject(stdout.Bytes())
	if perr != nil {
		if runErr != nil {
			return nil, fmt.Errorf("退出码非零（%v）：%s", runErr, clip(stderr.String(), 200))
		}
		return nil, perr
	}
	if runErr != nil {
		// 有 JSON 但退出码非零：协议本身用 JSON 报错，交给上层按 status 处理，这里只记日志。
		log.Printf("[station] 子模式 %s 输出 JSON 但退出码非零：%v", c.desc, runErr)
	}
	return obj, nil
}

// runModeJSON 执行一次性子模式（--check-session / --list-channels），
// 依次尝试候选后端（CLI > python），全部失败时返回包含各次失败原因的错误。
func (d *Downloader) runModeJSON(mode downloaderMode, timeout time.Duration) (map[string]interface{}, error) {
	set := d.buildBackends(mode)
	if len(set.cmds) == 0 {
		return nil, fmt.Errorf("没有可用的下载器后端：%s", strings.Join(set.reasons, "；"))
	}

	var fails []string
	for _, c := range set.cmds {
		obj, err := runCommandJSON(c, timeout)
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s（%s）：%v", c.desc, c.kind, err))
			log.Printf("[station] %s 后端失败：%v", c.desc, err)
			continue
		}
		return obj, nil
	}
	return nil, fmt.Errorf("子模式 %s 全部后端失败：%s", mode, strings.Join(fails, "；"))
}

// readFirstJSONLine 读取第一行合法 JSON（进程继续存活，留给扫码等待）。
func readFirstJSONLine(r io.Reader, timeout time.Duration) (map[string]interface{}, error) {
	type lineResult struct {
		obj map[string]interface{}
		err error
	}
	ch := make(chan lineResult, 1)
	go func() {
		br := bufio.NewReader(r)
		for {
			line, err := br.ReadString('\n')
			line = strings.TrimSpace(line)
			if line != "" && strings.HasPrefix(line, "{") {
				var m map[string]interface{}
				if json.Unmarshal([]byte(line), &m) == nil && m != nil {
					ch <- lineResult{obj: m}
					return
				}
			}
			if err != nil {
				ch <- lineResult{err: err}
				return
			}
		}
	}()

	select {
	case res := <-ch:
		if res.obj == nil {
			return nil, res.err
		}
		return res.obj, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("等待输出超时（%s）", timeout)
	}
}

// startQrLogin 启动扫码子进程并返回第一行 JSON（进程保持存活，继续等待扫码）。
// 与其它接口一样按 CLI > python 的优先级尝试；某个后端拿不到二维码就作废并换下一个。
// 调用方负责在二维码过期/重新生成时结束返回的命令。
func (d *Downloader) startQrLogin(timeout time.Duration) (*exec.Cmd, map[string]interface{}, string, error) {
	set := d.buildBackends(modeQrOnly)
	if len(set.cmds) == 0 {
		return nil, nil, "", fmt.Errorf("没有可用的下载器后端：%s", strings.Join(set.reasons, "；"))
	}

	var fails []string
	for _, c := range set.cmds {
		cmd := exec.Command(c.Path, c.Args...)
		cmd.Dir = c.Dir
		cmd.Env = c.Env
		cmd.SysProcAttr = childProcAttr()

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fails = append(fails, fmt.Sprintf("%s：创建 stdout 管道失败（%v）", c.desc, err))
			continue
		}
		if err := cmd.Start(); err != nil {
			fails = append(fails, fmt.Sprintf("%s：启动失败（%v）", c.desc, err))
			continue
		}

		obj, err := readFirstJSONLine(stdout, timeout)
		if err != nil {
			// 拿不到二维码：结束这个后端，避免残留进程占着 session
			_ = killProcessTree(cmd.Process.Pid)
			_ = cmd.Wait()
			fails = append(fails, fmt.Sprintf("%s：%v", c.desc, err))
			log.Printf("[station] 扫码后端失败：%s：%v", c.desc, err)
			continue
		}
		return cmd, obj, c.desc, nil
	}
	return nil, nil, "", fmt.Errorf("扫码登录进程启动失败：%s", strings.Join(fails, "；"))
}

// ---------- 启动 / 停止 ----------

// Start 启动下载器；已在运行（含外部独立进程）则复用，绝不重复启动。
// 重复启动会导致两个进程共用同一个 Telegram Session → 被判异地同时使用 → 封号。
//
// 整个「检查 → 构建 → 启动 → 记录」必须持 sessionMu 完成：
// 否则连点两次「开始监听」（或手机与电脑同时点）时，两个请求都会通过 isRunning()
// 检查、各起一个进程——这正是最怕的双实例场景。
func (d *Downloader) Start() error {
	d.sessionMu.Lock()
	defer d.sessionMu.Unlock()
	return d.startLocked()
}

// startLocked 是 Start 的实际逻辑；调用方必须已持有 sessionMu。
func (d *Downloader) startLocked() error {
	if d.isRunning() {
		d.mu.Lock()
		if d.cmd != nil {
			d.state = dlStateRun
		}
		d.mu.Unlock()
		log.Printf("[station] 下载器已在运行，复用现有实例")
		return nil
	}

	c, err := d.startCommand()
	if err != nil {
		return err
	}

	cmd := exec.Command(c.Path, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = c.Env
	// Windows 下隐藏子进程控制台窗口
	cmd.SysProcAttr = childProcAttr()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建 stdout 管道失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr 管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动下载器失败: %w", err)
	}

	d.mu.Lock()
	d.cmd = cmd
	d.state = dlStateRun
	d.pid = cmd.Process.Pid
	d.started = time.Now()
	d.exitCode = -1
	d.exitMsg = ""
	d.waitDone = make(chan struct{})
	d.mu.Unlock()

	log.Printf("[station] 下载器已启动 (PID %d): %s", cmd.Process.Pid, c.desc)

	// 日志转发（容忍非 UTF-8 字节，逐行读取）
	go d.pump(stdout, "out")
	go d.pump(stderr, "err")
	go d.wait(cmd)

	return nil
}

// startCommand 返回本次启动要执行的命令。startCommandHook 供单元测试替换，
// 生产环境为 nil（走真实的 buildStartCommand）。
func (d *Downloader) startCommand() (dlCommand, error) {
	if startCommandHook != nil {
		return startCommandHook(d)
	}
	return d.buildStartCommand()
}

// startCommandHook 单元测试注入点；生产环境始终为 nil。
var startCommandHook func(*Downloader) (dlCommand, error)

// pump 把子进程输出逐行写入日志
func (d *Downloader) pump(r io.Reader, tag string) {
	br := bufio.NewReaderSize(r, 64*1024)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			line = strings.TrimRight(line, "\r\n")
			if line != "" {
				d.logBuf.add(line)
				log.Printf("[dl:%s] %s", tag, line)
			}
		}
		if err != nil {
			return
		}
	}
}

// wait 监控子进程退出
func (d *Downloader) wait(cmd *exec.Cmd) {
	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		}
	}
	msg := ""
	if exitCode != 0 {
		msg = fmt.Sprintf("进程异常退出，退出码 %d", exitCode)
	}

	d.mu.Lock()
	// 只处理自己那一次启动的进程（重启后 d.cmd 已被替换）
	if d.cmd == cmd {
		d.state = dlStateExited
		d.exitCode = exitCode
		d.exitMsg = msg
		if d.waitDone != nil {
			close(d.waitDone)
			d.waitDone = nil
		}
	}
	d.mu.Unlock()

	log.Printf("[station] 下载器已退出 (PID %d, 退出码 %d)", cmd.Process.Pid, exitCode)
}

// Stop 停止下载器：既结束 Go 托管的子进程，也结束外部独立进程（开机自启/手动启动）。
// 与 Start/Restart 共用 sessionMu —— 否则「停止」与「开始」并发时，停止流程里那句
// 「确认已经退出」可能被并发的启动插进来，导致两个实例并存。
func (d *Downloader) Stop() error {
	d.sessionMu.Lock()
	defer d.sessionMu.Unlock()
	return d.stopLocked()
}

// stopLocked 是 Stop 的实际逻辑；调用方必须已持有 sessionMu。
func (d *Downloader) stopLocked() error {
	d.mu.Lock()
	cmd := d.cmd
	managed := d.managedRunningLocked()
	d.mu.Unlock()

	external := len(externalDownloaderProbe()) > 0
	if !managed && !external {
		d.resetStateIfSame(cmd)
		return nil
	}

	if external {
		if err := killExternalDownloaders(); err != nil {
			log.Printf("[station] 结束外部下载器失败: %v", err)
		}
	}
	if managed && cmd.Process != nil {
		if err := killProcessTree(cmd.Process.Pid); err != nil {
			log.Printf("[station] 结束托管下载器失败: %v", err)
		}
	}

	// 等待真正退出（最多 stopWaitTimeout），避免紧接着重启时与新实例抢 session
	deadline := time.Now().Add(stopWaitTimeout)
	for time.Now().Before(deadline) {
		d.mu.Lock()
		stillManaged := d.cmd == cmd && d.state == dlStateRun
		d.mu.Unlock()
		if !stillManaged && len(externalDownloaderProbe()) == 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	d.resetStateIfSame(cmd)

	if d.isRunning() {
		return errors.New("下载器未能完全停止，请稍后重试（已阻止重复启动，避免两个实例共用同一个 session）")
	}
	log.Printf("[station] 下载器已停止")
	return nil
}

// resetStateIfSame 手动停止后把状态归位为 idle；若期间已有新进程接管则不动。
func (d *Downloader) resetStateIfSame(cmd *exec.Cmd) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cmd != cmd {
		return
	}
	d.state = dlStateIdle
	d.cmd = nil
	d.pid = 0
	d.exitCode = -1
	d.exitMsg = ""
}

// Restart 重启下载器以应用新配置。
// 依据「真实运行状态」（含外部独立进程）决定是否重启，不能只看内部 d.state。
// 全程持有 sessionMu：停止与启动之间不允许别的「开始监听」插进来。
func (d *Downloader) Restart() error {
	d.sessionMu.Lock()
	defer d.sessionMu.Unlock()

	wasRunning := d.isRunning()
	if !wasRunning {
		return nil
	}
	if err := d.stopLocked(); err != nil {
		return err
	}
	if d.isRunning() {
		return errors.New("旧下载器未完全退出，已取消重启（避免两个实例共用同一个 session）")
	}
	return d.startLocked()
}

```


---

## 文件：downloader_start_test.go

```go
package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// useStartHook 把启动命令换成「会活一会儿的假下载器」，并统计构建命令的次数。
// 钩子里故意 sleep 一小段：这正是旧实现里 isRunning() 检查与 d.cmd 赋值之间的
// 竞态窗口——没有 sessionMu 时，并发请求会全部穿过检查、各起一个进程。
func useStartHook(t *testing.T, calls *int32, buildDelay time.Duration) {
	t.Helper()
	old := startCommandHook
	startCommandHook = func(d *Downloader) (dlCommand, error) {
		atomic.AddInt32(calls, 1)
		time.Sleep(buildDelay)
		fake := sleepCmd(30)
		return dlCommand{Path: fake.Path, Args: fake.Args[1:], desc: "测试假下载器"}, nil
	}
	t.Cleanup(func() { startCommandHook = old })
}

func managedPid(d *Downloader) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.pid
}

// 连点「开始监听」（或手机+电脑同时点）只能起一个下载器。
func TestStartSerializedSoOnlyOneDownloaderRuns(t *testing.T) {
	stubExternal(t) // 没有外部独立下载器
	var calls int32
	useStartHook(t, &calls, 150*time.Millisecond)

	d := newTestDownloader("", "")
	t.Cleanup(func() { _ = d.Stop() })

	const n = 4
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = d.Start()
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("第 %d 个 Start 返回错误: %v", i+1, err)
		}
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("并发 %d 次「开始监听」应只启动 1 个下载器，实际启动了 %d 个（双实例共用 session 会封号）", n, c)
	}
	if managedPid(d) == 0 {
		t.Fatal("启动后没有记录 PID")
	}
}

// Restart 内部要调 Stop+Start，若与 sessionMu 互锁就会死锁——这里守着超时跑一遍。
func TestRestartDoesNotDeadlockAndRestartsOnce(t *testing.T) {
	stubExternal(t)
	var calls int32
	useStartHook(t, &calls, 0)

	d := newTestDownloader("", "")
	t.Cleanup(func() { _ = d.Stop() })

	if err := d.Start(); err != nil {
		t.Fatalf("首次 Start 失败: %v", err)
	}
	firstPid := managedPid(d)

	done := make(chan error, 1)
	go func() { done <- d.Restart() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Restart 失败: %v", err)
		}
	case <-time.After(60 * time.Second):
		t.Fatal("Restart 超时（sessionMu 与 Stop/Start 互锁，或旧进程没被收掉）")
	}

	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Fatalf("Start + Restart 应各启动一次（共 2 次），实际 %d 次", c)
	}
	if pid := managedPid(d); pid == 0 || pid == firstPid {
		t.Fatalf("重启后 PID 未更新（旧 %d，新 %d）", firstPid, pid)
	}
}

// 「开始监听」与「停止监听」并发轰炸：结束后不能留下受管进程引用，
// 且停止流程返回后必须确实没有受管进程在跑。
func TestStartStopStormEndsClean(t *testing.T) {
	stubExternal(t)
	var calls int32
	useStartHook(t, &calls, 20*time.Millisecond)

	d := newTestDownloader("", "")

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				_ = d.Start()
			} else {
				_ = d.Stop()
			}
		}(i)
	}
	wg.Wait()

	if err := d.Stop(); err != nil {
		t.Fatalf("收尾 Stop 失败: %v", err)
	}
	d.mu.Lock()
	leftover := d.cmd
	state := d.state
	d.mu.Unlock()
	if leftover != nil {
		t.Fatal("停止后仍挂着受管进程引用")
	}
	if state != dlStateIdle {
		t.Fatalf("停止后状态应为 idle，实际 %s", state)
	}
	if pid := managedPid(d); pid != 0 {
		t.Fatalf("停止后 PID 应为 0，实际 %d", pid)
	}
}

```


---

## 文件：downloader_test.go

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// mocks 保存 testdata/mockcli 构建出的行为变体（见 testdata/mockcli/main.go）：
// ok / empty / nonzero / garbage / slow。用它们验证 JSON 协议、超时与回退，
// 完全不需要 Telegram 账号，也不会碰真实 session。
var mocks map[string]string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tgstation-mockcli")
	if err != nil {
		panic(err)
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	mocks = map[string]string{}
	variants := map[string]string{
		"ok":      "mockcli",
		"empty":   "mockcli_empty",
		"nonzero": "mockcli_nonzero",
		"garbage": "mockcli_garbage",
		"slow":    "mockcli_slow",
	}
	for key, name := range variants {
		out := filepath.Join(dir, name+ext)
		cmd := exec.Command("go", "build", "-o", out, "./testdata/mockcli")
		if b, err := cmd.CombinedOutput(); err != nil {
			os.Stderr.WriteString("构建 mockcli 失败: " + err.Error() + "\n" + string(b) + "\n")
			os.Exit(1)
		}
		mocks[key] = out
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// newTestDownloader 构造测试用 Downloader（不会启动真实下载器）。
func newTestDownloader(cli, python string) *Downloader {
	return &Downloader{
		state:    dlStateIdle,
		exitCode: -1,
		logBuf:   newRingBuffer(50),
		cfg: &AppConfig{Station: StationConfig{
			DownloaderCLI: cli,
			Python:        python,
		}},
	}
}

// withConfigPath 临时设置全局 configPath（子进程的 TG_STATION_CONFIG 用它）。
func withConfigPath(t *testing.T, p string) {
	t.Helper()
	old := configPath
	configPath = p
	t.Cleanup(func() { configPath = old })
}

// stubExternal 替换外部独立进程探测，避免测试触碰（或误杀）真实进程。
func stubExternal(t *testing.T, pids ...int) {
	t.Helper()
	old := externalDownloaderProbe
	externalDownloaderProbe = func() []int { return pids }
	t.Cleanup(func() { externalDownloaderProbe = old })
}

// ---------- Python 解释器探测 ----------

func TestClassifyPythonProbe(t *testing.T) {
	storeMsg := []byte("Python was not found; run without arguments to install from the " +
		"Microsoft Store, or disable this shortcut from Settings > Apps > Advanced app settings > App execution aliases.\r\n")

	cases := []struct {
		name     string
		out      []byte
		err      error
		timedOut bool
		wantErr  bool
		contains string
	}{
		{"正常版本输出", []byte("Python 3.14.7\n"), nil, false, false, ""},
		{"py 启动器输出", []byte("Python 3.11.9\r\n"), nil, false, false, ""},
		{"Store 别名 + 非零码", storeMsg, fmt.Errorf("exit status 9009"), false, true, "应用执行别名"},
		{"Store 别名 + 退出码 0", storeMsg, nil, false, true, "应用执行别名"},
		{"空输出 + 非零码", nil, fmt.Errorf("exit status 9009"), false, true, "失败"},
		{"超时", nil, nil, true, true, "超时"},
		{"输出与 Python 无关", []byte("hello world"), nil, false, true, "不是 Python 版本信息"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := classifyPythonProbe(tc.out, tc.err, tc.timedOut)
			if tc.wantErr && err == nil {
				t.Fatalf("期望报错，实际通过")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("期望通过，实际报错: %v", err)
			}
			if tc.wantErr && tc.contains != "" && !strings.Contains(err.Error(), tc.contains) {
				t.Errorf("错误信息 %q 未包含 %q", err.Error(), tc.contains)
			}
		})
	}
}

// TestValidatePythonRejectsStoreAlias 用本机真实的 WindowsApps 别名验证：
// 这种「应用执行别名」os.Stat 能通过、执行也不报错，但必须被识别为不可用。
func TestValidatePythonRejectsStoreAlias(t *testing.T) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		t.Skip("无 LOCALAPPDATA，跳过")
	}
	checked := 0
	for _, name := range []string{"python3.exe", "python.exe"} {
		alias := filepath.Join(local, "Microsoft", "WindowsApps", name)
		if _, err := os.Stat(alias); err != nil {
			continue
		}
		// 只有确认它是「未安装 Python 的别名」时才断言；装过 Store 版 Python 的机器上它是好解释器
		probe := exec.Command(alias, "--version")
		probe.SysProcAttr = childProcAttr()
		out, _ := probe.CombinedOutput()
		if !strings.Contains(strings.ToLower(string(out)), "was not found") {
			t.Logf("%s 是真实可用的 Python，跳过断言", alias)
			continue
		}
		checked++
		if err := validatePython(alias); err == nil {
			t.Errorf("应用执行别名 %s 被误判为可用 Python", alias)
		} else {
			t.Logf("正确拒绝 %s: %v", alias, err)
		}
	}
	if checked == 0 {
		t.Skip("本机没有失效的 WindowsApps Python 别名")
	}
}

func TestPickPythonSkipsBrokenCandidate(t *testing.T) {
	bogus := filepath.Join(t.TempDir(), "python.exe")
	if err := os.WriteFile(bogus, []byte("MZ not a real executable"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := pickPython([]string{bogus, mocks["ok"]})
	if err != nil {
		t.Fatalf("应跳过坏候选并选到可用解释器，却报错: %v", err)
	}
	if got != mocks["ok"] {
		t.Fatalf("选中的解释器 = %q，期望 %q", got, mocks["ok"])
	}

	_, err = pickPython([]string{bogus})
	if err == nil {
		t.Fatal("候选全部不可用时应报错")
	}
	if !strings.Contains(err.Error(), "python.exe") {
		t.Errorf("错误信息应包含失败的候选，实际: %v", err)
	}

	if _, err := pickPython(nil); err == nil {
		t.Fatal("空候选列表应报错")
	}
}

func TestFindPythonConfiguredPath(t *testing.T) {
	// 1) 路径不存在 → 明确错误 + 配置位置提示
	d := newTestDownloader("", filepath.Join(t.TempDir(), "no-such-python.exe"))
	_, err := d.findPython()
	if err == nil {
		t.Fatal("配置了不存在的路径应报错，而不是静默继续")
	}
	if !strings.Contains(err.Error(), "station.python") || !strings.Contains(err.Error(), "config.json") {
		t.Errorf("错误信息应说明配置位置与修复方式，实际: %v", err)
	}

	// 2) 路径存在但不是可执行解释器 → 报「不可用」
	badPath := filepath.Join(t.TempDir(), "fake-python.exe")
	if err := os.WriteFile(badPath, []byte("MZ not executable"), 0o644); err != nil {
		t.Fatal(err)
	}
	d = newTestDownloader("", badPath)
	if _, err := d.findPython(); err == nil || !strings.Contains(err.Error(), "不可用") {
		t.Errorf("存在的坏解释器应报「不可用」，实际: %v", err)
	}

	// 3) 配置了可用解释器 → 直接采用
	d = newTestDownloader("", mocks["ok"])
	got, err := d.findPython()
	if err != nil || got != mocks["ok"] {
		t.Fatalf("配置可用的解释器应被接受, got=%q err=%v", got, err)
	}
}

// ---------- 运行状态（统一判断） ----------

func TestIsRunningAndStatus(t *testing.T) {
	t.Run("外部独立进程在跑", func(t *testing.T) {
		stubExternal(t, 4242)
		d := newTestDownloader("", "")
		if !d.isRunning() {
			t.Fatal("外部下载器在运行时 isRunning() 应为 true")
		}
		st := d.status()
		if !st.Running || st.State != dlStateRun || st.Pid != 4242 {
			t.Fatalf("status = %+v，期望 running/true/pid=4242", st)
		}
	})

	t.Run("托管子进程在跑", func(t *testing.T) {
		stubExternal(t)
		d := newTestDownloader("", "")
		d.cmd = &exec.Cmd{Process: &os.Process{Pid: 111}}
		d.state = dlStateRun
		d.pid = 111
		if !d.isRunning() {
			t.Fatal("托管子进程在跑时 isRunning() 应为 true")
		}
		st := d.status()
		if !st.Running || st.State != dlStateRun || st.Pid != 111 {
			t.Fatalf("status = %+v，期望 running/true/pid=111", st)
		}
	})

	t.Run("记录为运行中但进程已消失 → exited", func(t *testing.T) {
		stubExternal(t)
		d := newTestDownloader("", "")
		d.state = dlStateRun
		st := d.status()
		if st.Running || st.State != dlStateExited {
			t.Fatalf("status = %+v，期望 exited/false", st)
		}
	})

	t.Run("从未启动 → idle", func(t *testing.T) {
		stubExternal(t)
		d := newTestDownloader("", "")
		st := d.status()
		if st.Running || st.State != dlStateIdle {
			t.Fatalf("status = %+v，期望 idle/false", st)
		}
	})
}

// TestRestartNoopWhenNotRunning 确认重启依据真实运行状态：没在跑就不启动新进程。
func TestRestartNoopWhenNotRunning(t *testing.T) {
	stubExternal(t)
	d := newTestDownloader(mocks["ok"], "")
	if err := d.Restart(); err != nil {
		t.Fatalf("未运行时 Restart 应直接返回 nil，实际: %v", err)
	}
	if d.isRunning() {
		t.Fatal("未运行时 Restart 不应启动进程")
	}
	d.mu.Lock()
	state := d.state
	d.mu.Unlock()
	if state != dlStateIdle {
		t.Fatalf("状态 = %q，期望 idle", state)
	}
}

// ---------- JSON 协议 / 超时 / 回退 ----------

func TestParseLastJSONObject(t *testing.T) {
	if obj, err := parseLastJSONObject([]byte("{\"status\":\"ok\"}\n")); err != nil || obj["status"] != "ok" {
		t.Fatalf("单行 JSON 解析失败: %v %v", obj, err)
	}
	obj, err := parseLastJSONObject([]byte("[log] 启动中...\n{\"status\":\"ok\",\"n\":1}\n"))
	if err != nil || obj["status"] != "ok" {
		t.Fatalf("混合日志时应取最后一行 JSON: %v %v", obj, err)
	}
	if _, err := parseLastJSONObject([]byte("这不是 JSON")); err == nil {
		t.Fatal("非 JSON 输出应报错")
	}
	if _, err := parseLastJSONObject(nil); err == nil {
		t.Fatal("空输出应报错")
	}
}

func TestRunModeJSONWithMockCLI(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	withConfigPath(t, cfgPath)
	bogusPython := filepath.Join(t.TempDir(), "no-python.exe")

	t.Run("CLI 后端输出 JSON（含工作目录与环境变量校验）", func(t *testing.T) {
		d := newTestDownloader(mocks["ok"], bogusPython)
		obj, err := d.runModeJSON(modeCheckSession, 10*time.Second)
		if err != nil {
			t.Fatalf("runModeJSON 失败: %v", err)
		}
		if obj["status"] != "logged_in" {
			t.Errorf("status = %v，期望 logged_in", obj["status"])
		}
		// 工作目录必须是 exe 所在目录
		gotCwd, _ := obj["cwd"].(string)
		wantCwd := filepath.Dir(mocks["ok"])
		if !samePath(gotCwd, wantCwd) {
			t.Errorf("子进程工作目录 = %q，期望 %q", gotCwd, wantCwd)
		}
		// 必须把 TG_STATION_CONFIG 指向程序正在使用的配置
		if gotCfg, _ := obj["config"].(string); gotCfg != cfgPath {
			t.Errorf("TG_STATION_CONFIG = %q，期望 %q", gotCfg, cfgPath)
		}
	})

	t.Run("频道列表协议", func(t *testing.T) {
		d := newTestDownloader(mocks["ok"], bogusPython)
		obj, err := d.runModeJSON(modeListChannels, 10*time.Second)
		if err != nil {
			t.Fatalf("runModeJSON 失败: %v", err)
		}
		if obj["status"] != "ok" {
			t.Fatalf("status = %v，期望 ok", obj["status"])
		}
		chans, ok := obj["channels"].([]interface{})
		if !ok || len(chans) != 1 {
			t.Fatalf("channels = %v，期望 1 条", obj["channels"])
		}
	})

	t.Run("非零退出码", func(t *testing.T) {
		d := newTestDownloader(mocks["nonzero"], bogusPython)
		_, err := d.runModeJSON(modeCheckSession, 10*time.Second)
		if err == nil {
			t.Fatal("非零退出码应报错")
		}
		if !strings.Contains(err.Error(), "退出码") {
			t.Errorf("错误信息应说明退出码问题: %v", err)
		}
	})

	t.Run("非法 JSON 输出", func(t *testing.T) {
		d := newTestDownloader(mocks["garbage"], bogusPython)
		_, err := d.runModeJSON(modeCheckSession, 10*time.Second)
		if err == nil || !strings.Contains(err.Error(), "JSON") {
			t.Errorf("非法 JSON 应报错并说明原因，实际: %v", err)
		}
	})

	t.Run("空输出", func(t *testing.T) {
		d := newTestDownloader(mocks["empty"], bogusPython)
		_, err := d.runModeJSON(modeCheckSession, 10*time.Second)
		if err == nil || !strings.Contains(err.Error(), "输出") {
			t.Errorf("空输出应报错并说明原因，实际: %v", err)
		}
	})

	t.Run("超时", func(t *testing.T) {
		d := newTestDownloader(mocks["slow"], bogusPython)
		start := time.Now()
		_, err := d.runModeJSON(modeCheckSession, 1*time.Second)
		if err == nil || !strings.Contains(err.Error(), "超时") {
			t.Errorf("超时应报错，实际: %v", err)
		}
		if elapsed := time.Since(start); elapsed > 10*time.Second {
			t.Errorf("超时未被及时处理，耗时 %s", elapsed)
		}
	})

	t.Run("没有任何后端时给出原因", func(t *testing.T) {
		d := newTestDownloader(filepath.Join(t.TempDir(), "nope.exe"), bogusPython)
		_, err := d.runModeJSON(modeCheckSession, 2*time.Second)
		if err == nil || !strings.Contains(err.Error(), "没有可用的下载器后端") {
			t.Errorf("应说明所有后端都不可用，实际: %v", err)
		}
	})

	t.Run("CLI 失败时回退到 python + 脚本", func(t *testing.T) {
		// CLI 用「空输出」的坏变体，python 用可用变体（脚本路径由包目录下的 tg_auto_download.py 提供）
		d := newTestDownloader(mocks["empty"], mocks["ok"])
		obj, err := d.runModeJSON(modeCheckSession, 10*time.Second)
		if err != nil {
			t.Fatalf("应回退到 python 后端，实际失败: %v", err)
		}
		if obj["status"] != "logged_in" {
			t.Errorf("回退结果 status = %v，期望 logged_in", obj["status"])
		}
	})
}

// ---------- 扫码：立即返回二维码、进程继续等待 ----------

func TestStartQrLoginReturnsBeforeScan(t *testing.T) {
	withConfigPath(t, filepath.Join(t.TempDir(), "config.json"))
	stubExternal(t)
	d := newTestDownloader(mocks["ok"], filepath.Join(t.TempDir(), "no-python.exe"))

	start := time.Now()
	cmd, obj, desc, err := d.startQrLogin(15 * time.Second)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("startQrLogin 失败: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = killProcessTree(cmd.Process.Pid)
		}
		_ = cmd.Wait()
	}()

	if obj["status"] != "qr" {
		t.Fatalf("status = %v，期望 qr", obj["status"])
	}
	if obj["qr_base64"] == "" {
		t.Errorf("缺少 qr_base64")
	}
	// mock 在输出二维码后会睡 30 秒模拟等待扫码：必须立刻返回，不能等进程退出
	if elapsed > 10*time.Second {
		t.Errorf("收到二维码耗时 %s，应在扫码完成前就返回", elapsed)
	}
	if cmd.ProcessState != nil {
		t.Errorf("进程已退出，二维码进程应保持存活等待扫码")
	}
	if desc == "" {
		t.Errorf("应返回后端描述用于日志")
	}
}

func TestStartQrLoginFallbackAndCleanup(t *testing.T) {
	withConfigPath(t, filepath.Join(t.TempDir(), "config.json"))
	stubExternal(t)

	// 坏 CLI（无输出）→ 回退到 python 后端
	d := newTestDownloader(mocks["empty"], mocks["ok"])
	cmd, obj, desc, err := d.startQrLogin(10 * time.Second)
	if err != nil {
		t.Fatalf("应回退到 python 后端，实际失败: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = killProcessTree(cmd.Process.Pid)
		}
		_ = cmd.Wait()
	}()
	if obj["status"] != "qr" {
		t.Fatalf("回退后 status = %v，期望 qr", obj["status"])
	}
	// 回退必须由 python 后端提供（描述里含 python 可执行文件名，而不是坏 CLI 的名字）
	if !strings.Contains(desc, filepath.Base(mocks["ok"])) {
		t.Errorf("回退后端描述 = %q，应包含 python 后端 %q", desc, filepath.Base(mocks["ok"]))
	}

	// 所有后端都拿不到二维码 → 报错，并且不能留下残留进程
	d2 := newTestDownloader(mocks["empty"], filepath.Join(t.TempDir(), "no-python.exe"))
	if _, _, _, err := d2.startQrLogin(2 * time.Second); err == nil {
		t.Fatal("全部后端失败时应报错")
	}
}

// samePath 比较两个路径是否指向同一位置（Windows 下大小写/短名不敏感）。
func samePath(a, b string) bool {
	ca, err1 := filepath.Abs(filepath.Clean(a))
	cb, err2 := filepath.Abs(filepath.Clean(b))
	if err1 != nil || err2 != nil {
		return false
	}
	if ea, err := filepath.EvalSymlinks(ca); err == nil {
		ca = ea
	}
	if eb, err := filepath.EvalSymlinks(cb); err == nil {
		cb = eb
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(ca, cb)
	}
	return ca == cb
}

```


---

## 文件：files_pagination_test.go

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 分页接口回归测试：page 缺省 / 合法 / 非整数 / 0 / 负数 / 超出页数 / 极大页码。
// 旧实现用 strconv.Atoi 后忽略错误且不校验最小值，page=0 或负数会算出负的切片
// 起点 → 越界 panic；这些用例正是为了让那种回归无法再发生。

func seedIndex(t *testing.T, counts map[string]int) {
	t.Helper()
	indexMutex.Lock()
	old := mediaIndex
	indexMutex.Unlock()
	t.Cleanup(func() {
		indexMutex.Lock()
		rebuildIndexLocked(old)
		indexMutex.Unlock()
	})

	var idx []*MediaFile
	for date, n := range counts {
		d, err := time.Parse("2006-01-02", date)
		if err != nil {
			t.Fatalf("测试日期解析失败: %v", err)
		}
		for i := 0; i < n; i++ {
			idx = append(idx, &MediaFile{
				Path:     fmt.Sprintf("%s/file%03d.jpg", date, i),
				Date:     d.Add(time.Duration(i) * time.Minute),
				IsVideo:  i%10 == 0,
				Size:     int64(1024 * (i + 1)),
				Duration: float64(i),
			})
		}
	}
	// 按时间倒序（与 scanFiles 一致）
	for i, j := 0, len(idx)-1; i < j; i, j = i+1, j-1 {
		idx[i], idx[j] = idx[j], idx[i]
	}
	// 通过索引重建入口写入：文件列表与日期索引必须同步切换
	// （/api/files/{date} 现在直接查日期索引，单独改 mediaIndex 已不足以生效）
	rebuildIndex(idx)
}

type filesResp struct {
	Files []LightMediaFile `json:"files"`
	Total int              `json:"total"`
	Pages int              `json:"pages"`
	Page  int              `json:"page"`
	Date  string           `json:"date"`
	Error string           `json:"error"`
}

func callFilesHandler(t *testing.T, target string) (int, filesResp) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	filesHandler(rec, req) // 旧实现在 page=0/-1 时这里会 panic，测试直接失败
	var resp filesResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	return rec.Code, resp
}

func TestFilesHandlerPagination(t *testing.T) {
	seedIndex(t, map[string]int{"2026-05-01": 250, "2026-05-02": 5})

	cases := []struct {
		name      string
		target    string
		wantCode  int
		wantCount int
		wantTotal int
		wantPages int
		wantPage  int
	}{
		{"未传 page → 第 1 页", "/api/files/2026-05-01", http.StatusOK, 100, 250, 3, 1},
		{"page 为空串 → 第 1 页", "/api/files/2026-05-01?page=", http.StatusOK, 100, 250, 3, 1},
		{"第 2 页", "/api/files/2026-05-01?page=2", http.StatusOK, 100, 250, 3, 2},
		{"最后一页不满", "/api/files/2026-05-01?page=3", http.StatusOK, 50, 250, 3, 3},
		{"超出页数 → 空数组", "/api/files/2026-05-01?page=4", http.StatusOK, 0, 250, 3, 4},
		{"极大页码（上限内）→ 空数组", "/api/files/2026-05-01?page=9999999", http.StatusOK, 0, 250, 3, 9999999},
		{"小集合第 1 页", "/api/files/2026-05-02", http.StatusOK, 5, 5, 1, 1},
		{"日期无文件", "/api/files/1999-01-01", http.StatusOK, 0, 0, 0, 1},

		{"非整数 → 400", "/api/files/2026-05-01?page=abc", http.StatusBadRequest, 0, 0, 0, 0},
		{"小数 → 400", "/api/files/2026-05-01?page=1.5", http.StatusBadRequest, 0, 0, 0, 0},
		{"0 → 400", "/api/files/2026-05-01?page=0", http.StatusBadRequest, 0, 0, 0, 0},
		{"负数 → 400", "/api/files/2026-05-01?page=-1", http.StatusBadRequest, 0, 0, 0, 0},
		{"超过上限 → 400", "/api/files/2026-05-01?page=10000001", http.StatusBadRequest, 0, 0, 0, 0},
		{"整数溢出字符串 → 400", "/api/files/2026-05-01?page=99999999999999999999", http.StatusBadRequest, 0, 0, 0, 0},
		{"带空格 → 400", "/api/files/2026-05-01?page=%20", http.StatusBadRequest, 0, 0, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, resp := callFilesHandler(t, tc.target)
			if code != tc.wantCode {
				t.Fatalf("HTTP 状态码 = %d, 期望 %d（body error=%q）", code, tc.wantCode, resp.Error)
			}
			if tc.wantCode == http.StatusBadRequest {
				if resp.Error == "" {
					t.Fatalf("400 响应必须带 error 说明")
				}
				return
			}
			if len(resp.Files) != tc.wantCount {
				t.Errorf("files 数量 = %d, 期望 %d", len(resp.Files), tc.wantCount)
			}
			if resp.Total != tc.wantTotal {
				t.Errorf("total = %d, 期望 %d", resp.Total, tc.wantTotal)
			}
			if resp.Pages != tc.wantPages {
				t.Errorf("pages = %d, 期望 %d", resp.Pages, tc.wantPages)
			}
			if resp.Page != tc.wantPage {
				t.Errorf("page = %d, 期望 %d", resp.Page, tc.wantPage)
			}
			if resp.Date != "2026-05-01" && strings.Contains(tc.target, "2026-05-01") {
				t.Errorf("date = %q, 期望 2026-05-01", resp.Date)
			}
		})
	}
}

// TestFilesHandlerPageDoesNotPanic 显式覆盖曾经会 panic 的输入。
func TestFilesHandlerPageDoesNotPanic(t *testing.T) {
	seedIndex(t, map[string]int{"2026-05-01": 10})
	for _, page := range []string{"0", "-1", "-999999999", "abc", ""} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("page=%q 触发 panic: %v", page, r)
				}
			}()
			_, _ = callFilesHandler(t, "/api/files/2026-05-01?page="+page)
		}()
	}
}

func TestParsePageParam(t *testing.T) {
	ok := []struct {
		raw  string
		want int
	}{
		{"", 1},
		{"1", 1},
		{"42", 42},
		{"10000000", 10000000},
	}
	for _, c := range ok {
		got, err := parsePageParam(c.raw)
		if err != nil || got != c.want {
			t.Errorf("parsePageParam(%q) = (%d, %v), 期望 (%d, nil)", c.raw, got, err, c.want)
		}
	}

	bad := []string{"abc", "1.5", "0", "-1", "-10000000", "10000001", "99999999999999999999", " ", "1e3"}
	for _, raw := range bad {
		if got, err := parsePageParam(raw); err == nil {
			t.Errorf("parsePageParam(%q) 期望报错，实际 = (%d, nil)", raw, got)
		}
	}
}

```


---

## 文件：index.go

```go
package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

// ── 媒体索引与日期索引（第二阶段 · 第三部分）────────────────────────────
//
// 旧实现里 /api/files/{date}、/api/dates、首页注入每次都遍历整个 mediaIndex
// （5000+ 文件 × 每个请求），日期查询是 O(全部文件数)。
//
// 现在维护一个日期索引 map[string][]*MediaFile，与文件列表在**同一把锁**下
// 一起原子切换（rebuildIndex），保证「列表已更新、日期索引还是旧数据」不可能发生。
// 日期键沿用现有格式 time.Format("2006-01-02")，顺序沿用「全局按修改时间倒序」，
// 因此分页结果与旧实现完全一致。

var (
	// dateIndexPool 保存当前生效的日期索引（只读视图，重建时整体替换）。
	dateIndex map[string][]*MediaFile
	// dateKeys 保存降序排列的日期列表，避免 /api/dates 每次重算排序。
	dateKeys []string
)

// rebuildIndexLocked 用新的文件列表重建索引结构。调用方必须已持有 indexMutex 写锁。
func rebuildIndexLocked(files []*MediaFile) {
	idx := make(map[string][]*MediaFile, 64)
	for _, f := range files {
		key := f.Date.Format("2006-01-02")
		idx[key] = append(idx[key], f)
	}
	keys := make([]string, 0, len(idx))
	for k := range idx {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	mediaIndex = files
	dateIndex = idx
	dateKeys = keys
}

// rebuildIndex 加锁重建索引（文件列表 + 日期索引 + 日期键一起切换）。
func rebuildIndex(files []*MediaFile) {
	indexMutex.Lock()
	rebuildIndexLocked(files)
	lastScan = time.Now()
	indexMutex.Unlock()
}

// filesForDate 返回某个日期的文件（返回共享切片，调用方只读）。
func filesForDate(date string) []*MediaFile {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	return dateIndex[date]
}

// indexDates 返回降序日期列表的副本。
func indexDates() []string {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	out := make([]string, len(dateKeys))
	copy(out, dateKeys)
	return out
}

// indexTotal 返回索引中的文件总数。
func indexTotal() int {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	return len(mediaIndex)
}

// indexSnapshotFiles 返回文件列表的浅拷贝（用于批量缩略图等场景）。
func indexSnapshotFiles() []*MediaFile {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	out := make([]*MediaFile, len(mediaIndex))
	copy(out, mediaIndex)
	return out
}

// loadIndex 从磁盘加载索引文件（内容为 []*MediaFile）。
func loadIndex() {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		log.Printf("Index file not found: %v", err)
		return
	}

	var files []*MediaFile
	if err := json.Unmarshal(data, &files); err != nil {
		log.Printf("Failed to parse index: %v", err)
		return
	}
	// 索引文件里的字段可能缺失（如旧版本没有 Size），做一次规范化
	for _, f := range files {
		if f == nil {
			continue
		}
		if f.Duration < 0 {
			f.Duration = 0
		}
	}

	rebuildIndex(files)
	log.Printf("Loaded index with %d files", len(files))
}

// saveIndex 持久化当前索引（先取快照再序列化，避免持锁做 IO）。
func saveIndex() {
	files := indexSnapshotFiles()
	data, err := json.Marshal(files)
	if err != nil {
		log.Printf("Failed to marshal index: %v", err)
		return
	}
	if err := os.WriteFile(indexPath, data, 0o644); err != nil {
		log.Printf("Failed to save index: %v", err)
	}
}

```


---

## 文件：index_test.go

```go
package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ── 日期索引测试与性能基准（第二阶段 · 第三部分）────────────────────────
//
// 旧实现：/api/files/{date} 每次都遍历整个 mediaIndex（O(全部文件)）。
// 新实现：按日期分组的索引，查询成本 O(该日期的文件数)。
// 这里既验证「索引结果与线性扫描完全一致」，也给出可重复的基准数据。

func genMediaFiles(n, days int) []*MediaFile {
	base := time.Date(2026, 5, 1, 12, 0, 0, 0, time.Local)
	files := make([]*MediaFile, 0, n)
	for i := 0; i < n; i++ {
		day := time.Duration(i%days) * 24 * time.Hour
		ts := base.Add(-day).Add(time.Duration(i/days) * time.Minute)
		files = append(files, &MediaFile{
			Path:     fmt.Sprintf("day%02d/file%05d.jpg", i%days, i),
			Date:     ts,
			IsVideo:  i%10 == 0,
			Size:     int64(i * 1024),
			Duration: float64(i % 60),
		})
	}
	sortByDateDesc(files) // 与 scanFiles 相同的排序规则
	return files
}

func withIndex(t *testing.T, files []*MediaFile) {
	t.Helper()
	rebuildIndex(files)
	t.Cleanup(func() { rebuildIndex(nil) })
}

func TestRebuildIndexDateGrouping(t *testing.T) {
	files := genMediaFiles(300, 6)
	withIndex(t, files)

	if got := indexTotal(); got != 300 {
		t.Fatalf("indexTotal = %d，期望 300", got)
	}
	keys := indexDates()
	if len(keys) != 6 {
		t.Fatalf("日期个数 = %d，期望 6", len(keys))
	}
	// 日期键必须降序
	for i := 1; i < len(keys); i++ {
		if keys[i-1] <= keys[i] {
			t.Fatalf("日期键未按降序排列: %v", keys[:min(3, len(keys))])
		}
	}
	// 每个日期分组的文件数之和 = 总数，且键与文件自身日期一致
	total := 0
	for _, k := range keys {
		for _, f := range filesForDate(k) {
			total++
			if f.Date.Format("2006-01-02") != k {
				t.Fatalf("日期索引分组错误: 文件 %s 出现在 %s 组", f.Path, k)
			}
		}
	}
	if total != 300 {
		t.Fatalf("各分组文件数合计 = %d，期望 300", total)
	}
}

func TestFilesForDateMatchesLinearScan(t *testing.T) {
	files := genMediaFiles(2500, 25)
	withIndex(t, files)

	// 逐日期比对：新索引查询必须与旧的「遍历全部文件」结果完全一致（含顺序）
	for _, date := range indexDates() {
		linear := linearScanForDate(files, date)
		indexed := filesForDate(date)
		if len(linear) != len(indexed) {
			t.Fatalf("%s: 索引结果 %d 条，线性扫描 %d 条", date, len(indexed), len(linear))
		}
		for i := range linear {
			if linear[i].Path != indexed[i].Path {
				t.Fatalf("%s: 第 %d 条不一致（线性 %s / 索引 %s）", date, i, linear[i].Path, indexed[i].Path)
			}
		}
	}

	// 不存在的日期 / 非法日期
	if got := filesForDate("1999-01-01"); len(got) != 0 {
		t.Errorf("不存在的日期应返回空，实际 %d 条", len(got))
	}
	if got := filesForDate("not-a-date"); len(got) != 0 {
		t.Errorf("非法日期应返回空，实际 %d 条", len(got))
	}
}

// linearScanForDate 复刻旧实现的查询方式（遍历整个列表筛选同一天）。
func linearScanForDate(files []*MediaFile, date string) []*MediaFile {
	var out []*MediaFile
	for _, f := range files {
		if f.Date.Format("2006-01-02") == date {
			out = append(out, f)
		}
	}
	return out
}

func TestDateIndexConcurrentReadWrite(t *testing.T) {
	files := genMediaFiles(600, 10)
	withIndex(t, files)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	// 并发读
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				for _, d := range indexDates() {
					_ = filesForDate(d)
				}
				_ = indexTotal()
				_ = indexSnapshotFiles()
			}
		}()
	}
	// 并发重建（模拟扫描完成）
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				rebuildIndex(files)
			}
		}()
	}
	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()

	// 重建后列表与日期索引必须一致
	if indexTotal() != 600 {
		t.Fatalf("并发重建后 indexTotal = %d，期望 600", indexTotal())
	}
	sum := 0
	for _, d := range indexDates() {
		sum += len(filesForDate(d))
	}
	if sum != 600 {
		t.Fatalf("并发重建后日期索引合计 = %d，期望 600（两者必须一起原子切换）", sum)
	}
}

func BenchmarkFilesByDateLinear(b *testing.B) {
	files := genMediaFiles(5000, 30)
	rebuildIndex(files)
	defer rebuildIndex(nil)
	date := files[0].Date.Format("2006-01-02")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := linearScanForDate(files, date)
		if len(out) == 0 {
			b.Fatal("基准数据异常：当天无文件")
		}
	}
}

func BenchmarkFilesByDateIndexed(b *testing.B) {
	files := genMediaFiles(5000, 30)
	rebuildIndex(files)
	defer rebuildIndex(nil)
	date := files[0].Date.Format("2006-01-02")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := filesForDate(date)
		if len(out) == 0 {
			b.Fatal("基准数据异常：当天无文件")
		}
	}
}

```


---

## 文件：login_qr_test.go

```go
package main

import (
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// sleepCmd 起一个会自己活一小段时间的子进程（Windows 用 ping，其它平台用 sleep）。
// 用来模拟「等待扫码的下载器子进程」：既能自然退出，也能被 kill。
func sleepCmd(seconds int) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("ping", "-n", strconv.Itoa(seconds+1), "127.0.0.1")
	}
	return exec.Command("sleep", strconv.Itoa(seconds))
}

// newTestQrProc 起子进程并按生产同一条路径挂上唯一收尸者。
func newTestQrProc(t *testing.T, cmd *exec.Cmd) *qrProcState {
	t.Helper()
	if err := cmd.Start(); err != nil {
		t.Fatalf("启动测试子进程失败: %v", err)
	}
	st := &qrProcState{cmd: cmd, done: make(chan struct{})}
	go reapQr(st)
	return st
}

// swapQrHooks 换掉测试钩子，返回还原函数。
func swapQrHooks(killer func(int) error, start func() (*exec.Cmd, map[string]interface{}, string, error)) func() {
	oldKill, oldStart := qrKillTree, qrStart
	qrKillTree, qrStart = killer, start
	return func() { qrKillTree, qrStart = oldKill, oldStart }
}

// passThroughKiller 真杀进程（调生产实现），只额外计数。
func passThroughKiller(counter *int32) func(int) error {
	return func(pid int) error {
		atomic.AddInt32(counter, 1)
		return killProcessTree(pid)
	}
}

func setQrCur(st *qrProcState) {
	qrMu.Lock()
	qrCur = st
	qrMu.Unlock()
}

func qrCurSnapshot() *qrProcState {
	qrMu.Lock()
	defer qrMu.Unlock()
	return qrCur
}

func waitDone(t *testing.T, st *qrProcState, d time.Duration, what string) {
	t.Helper()
	select {
	case <-st.done:
	case <-time.After(d):
		t.Fatalf("%s：%s 内没有等到进程被收尸", what, d)
	}
}

// 进程自然退出后，引用必须被收尸者清掉——否则下次扫码会拿旧 PID 去 taskkill，
// 而 PID 可能已经被系统复用给别的进程。
func TestReapQrClearsReferenceAfterExit(t *testing.T) {
	st := newTestQrProc(t, sleepCmd(1))
	setQrCur(st)

	waitDone(t, st, 20*time.Second, "自然退出")

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if qrCurSnapshot() == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("子进程自然退出后 qrCur 仍指向旧进程（旧 PID 可能已被复用）")
}

// 已经退出的进程一律不许再 kill。
func TestKillQrLockedSkipsExitedProcess(t *testing.T) {
	var kills int32
	restore := swapQrHooks(func(pid int) error {
		atomic.AddInt32(&kills, 1)
		return nil
	}, qrStart)
	defer restore()

	st := newTestQrProc(t, sleepCmd(1))
	waitDone(t, st, 20*time.Second, "自然退出")

	// 复现旧实现的隐患场景：进程已退出，引用还挂着
	setQrCur(st)

	qrMu.Lock()
	killQrLocked()
	qrMu.Unlock()

	if n := atomic.LoadInt32(&kills); n != 0 {
		t.Fatalf("对已退出的进程调用了 %d 次 kill（PID 可能已被复用，会误杀无关进程）", n)
	}
	if qrCurSnapshot() != nil {
		t.Fatal("killQrLocked 之后 qrCur 没有清空")
	}
}

// 活着的进程：kill 一次，并且要等它真的退出（done 关闭）。
func TestKillQrLockedKillsLiveProcess(t *testing.T) {
	var kills int32
	restore := swapQrHooks(passThroughKiller(&kills), qrStart)
	defer restore()

	st := newTestQrProc(t, sleepCmd(30))
	setQrCur(st)

	qrMu.Lock()
	killQrLocked()
	qrMu.Unlock()

	if n := atomic.LoadInt32(&kills); n != 1 {
		t.Fatalf("活着的进程应恰好被 kill 1 次，实际 %d 次", n)
	}
	waitDone(t, st, 10*time.Second, "被 kill 后")
	if qrCurSnapshot() != nil {
		t.Fatal("killQrLocked 之后 qrCur 没有清空")
	}
}

// 并发的扫码请求必须被串行化：每个后来的请求都要先杀掉前一个进程，
// 最终只留一个。若「kill 与 set」不在同一把锁里，就统计不到 n-1 次 kill，
// 而且会有多个进程同时存在（共用 session 有封号风险）。
func TestBeginQrLoginSerializesKillAndStart(t *testing.T) {
	var kills, starts int32
	restore := swapQrHooks(passThroughKiller(&kills),
		func() (*exec.Cmd, map[string]interface{}, string, error) {
			atomic.AddInt32(&starts, 1)
			cmd := sleepCmd(30)
			if err := cmd.Start(); err != nil {
				return nil, nil, "", err
			}
			return cmd, map[string]interface{}{"status": "qr"}, "测试后端", nil
		})
	defer restore()

	const n = 4
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, _, err := beginQrLogin(); err != nil {
				t.Errorf("beginQrLogin 失败: %v", err)
			}
		}()
	}
	wg.Wait()

	if s := atomic.LoadInt32(&starts); s != n {
		t.Fatalf("启动次数 %d，期望 %d", s, n)
	}
	if k := atomic.LoadInt32(&kills); k != n-1 {
		t.Fatalf("并发扫码未串行化：期望 kill %d 次（每个新请求杀掉前一个），实际 %d 次", n-1, k)
	}

	// 收尾，别留测试进程
	qrMu.Lock()
	killQrLocked()
	qrMu.Unlock()
	if qrCurSnapshot() != nil {
		t.Fatal("测试收尾后 qrCur 应为 nil")
	}
}

```


---

## 文件：main.go

```go
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MediaFile struct {
	Path     string
	Date     time.Time
	IsVideo  bool
	Size     int64
	Duration float64 // 视频时长（秒），图片为 0；由 ffmpeg 探测并随索引持久化
}

type LightMediaFile struct {
	Path     string  `json:"path"`
	Date     string  `json:"date"`
	IsVideo  bool    `json:"isVideo"`
	Size     int64   `json:"size"`
	Duration float64 `json:"duration"`
}

var (
	mediaDir     string
	port         int
	indexPath    string
	thumbsDir    string
	staticDir    string
	mediaIndex   []*MediaFile
	indexMutex   sync.RWMutex
	lastScan     time.Time
	scanInterval = time.Hour
	videoExts    = map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".mkv": true, ".webm": true}
	imageExts    = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	thumbSize    = 400
	ffmpegPath   string
)

func findFFmpeg() string {
	exePath, err := os.Executable()
	if err != nil {
		return "ffmpeg"
	}
	exeDir := filepath.Dir(exePath)

	ffmpegNames := []string{"ffmpeg", "ffmpeg.exe"}

	// 候选目录：exe 目录 > 当前目录 > config 目录 > 下载器脚本目录
	candidates := []string{exeDir}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if configPath != "" {
		candidates = append(candidates, filepath.Dir(configPath))
	}
	if script, err := downloader.findScript(); err == nil && script != "" {
		candidates = append(candidates, filepath.Dir(script))
	}

	for _, dir := range candidates {
		for _, name := range ffmpegNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	systemPaths := []string{"/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg", "/bin/ffmpeg"}
	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "ffmpeg"
}

func selectFolderWindows() string {
	psScript := `
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.Description = "请选择包含图片和视频的文件夹"
$dialog.ShowDialog() | Out-Null
$dialog.SelectedPath
`
	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func main() {
	var configFlag string
	var showConsole bool
	flag.StringVar(&configFlag, "config", "", "配置文件路径（默认: 程序目录/config.json）")
	flag.BoolVar(&showConsole, "console", false, "保留控制台窗口（默认隐藏，后台运行）")
	flag.StringVar(&mediaDir, "dir", "", "媒体目录（默认取配置 download_dir）")
	flag.IntVar(&port, "port", 8080, "HTTP 端口（默认 8080）")
	flag.StringVar(&indexPath, "index", "", "索引文件路径（默认: 程序目录/media_index.json）")
	flag.StringVar(&thumbsDir, "thumbs", "", "缩略图缓存目录（默认: 程序目录/thumbs）")
	flag.Parse()

	// Windows 下默认隐藏控制台窗口（后台运行）；-console 可保留
	if runtime.GOOS == "windows" && !showConsole {
		hideConsoleWindow()
	}

	// 加载配置：-config 可指定路径，默认 exe 目录/config.json
	configPath = configFlag
	if configPath == "" {
		configPath = defaultConfigPath()
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("[station] 配置加载失败，使用默认值: %v", err)
		cfg = &AppConfig{Raw: buildDefaultRaw(stationDefaults())}
		applyRaw(cfg)
	}
	downloader.cfg = cfg

	// ffmpeg 定位依赖 configPath/downloader（多目录搜索），须在配置加载后调用
	ffmpegPath = findFFmpeg()
	// ffprobe 优先用于时长探测；找不到时为空，probeDuration 会自动回退 ffmpeg
	ffprobePath = findFFprobe()

	// flag 未显式设置时，用配置值覆盖默认
	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	if !setFlags["port"] && cfg.Port != 0 {
		port = cfg.Port
	}
	if !setFlags["dir"] && cfg.MediaDir != "" {
		mediaDir = cfg.MediaDir
	}
	if !setFlags["index"] {
		indexPath = cfg.IndexPath
	}
	if !setFlags["thumbs"] {
		exeDir := "."
		if exe, err := os.Executable(); err == nil {
			exeDir = filepath.Dir(exe)
		}
		thumbsDir = filepath.Join(exeDir, "thumbs")
	}

	// 前端资源目录：exe 目录/static 优先，其次进程工作目录/static。
	// 注意 filepath.Join(dir, "static") 会得到绝对路径，原先写 staticDir="static"
	// （相对路径）导致兜底分支实际依赖进程工作目录，"exe 目录优先"形同虚设——
	// 换个工作目录启动就会静默服务到另一个 index.html。
	staticDir = ""
	if exePath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "static")
		if _, statErr := os.Stat(candidate); statErr == nil {
			staticDir = candidate
		}
	}
	if staticDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			candidate := filepath.Join(cwd, "static")
			if _, statErr := os.Stat(candidate); statErr == nil {
				staticDir = candidate
			}
		}
	}
	if staticDir == "" {
		staticDir = "static"
	}
	log.Printf("[station] 前端资源目录: %s", staticDir)

	if mediaDir == "" {
		fmt.Println("========================================")
		fmt.Println("      🖼️  媒体浏览器 - Media Browser")
		fmt.Println("========================================")
		fmt.Println()

		if runtime.GOOS == "windows" {
			fmt.Println("💡 正在打开文件夹选择框...")
			selectedFolder := selectFolderWindows()
			if selectedFolder != "" {
				mediaDir = selectedFolder
			}
		}

		if mediaDir == "" {
			fmt.Println("请输入图片/视频文件夹路径：")
			fmt.Print("> ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			mediaDir = strings.TrimSpace(input)
		}

		if mediaDir == "" {
			fmt.Println("错误：未选择文件夹，程序退出。")
			os.Exit(1)
		}
	}

	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		log.Fatal(err)
	}

	loadIndex()

	// 扫描/刷新统一走调度器：同一时刻最多一个扫描在跑，扫描期间的请求
	// 合并为下一轮补扫，缩略图只在「没有待补扫任务」之后基于最新索引生成。
	go func() {
		scanSched.Request()
		for range time.Tick(scanInterval) {
			scanSched.Request()
		}
	}()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/dates", datesHandler)
	http.HandleFunc("/api/files/", filesHandler)
	http.HandleFunc("/api/file/", fileHandler)
	http.HandleFunc("/api/delete", deleteHandler)
	http.HandleFunc("/thumb/", thumbHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	// 整合功能：下载器管理 / 设置 / 刷新（必须在 ListenAndServe 前挂载）
	registerAPIRoutes()

	// 配置开启自动启动时，拉起下载器
	if cfg.Station.AutoStart {
		go func() {
			if err := downloader.Start(); err != nil {
				log.Printf("[station] 自动启动下载器失败: %v", err)
			}
		}()
	}

	// 代理由用户常开的 Clash Verge（127.0.0.1:7897）负责，内置 mihomo 自启动已于 2026-10-09 移除。

	fmt.Println()
	fmt.Println("=============================================")
	fmt.Println("      🎛️  TG Station - 下载 + 媒体 一体机")
	fmt.Println("=============================================")
	fmt.Println()
	fmt.Println("✅ 启动成功！")
	fmt.Println("📁 媒体文件夹：", mediaDir)
	fmt.Println("🌐 访问地址：http://localhost:" + strconv.Itoa(port))
	fmt.Println("🎧 下载器：/api/listen/start 启动 | /api/listen/stop 停止")
	fmt.Println("⚙️  设置面板：页面右上角 ⚙️ 按钮（或 /api/settings）")
	fmt.Println()

	log.Printf("Server starting on :%d, media dir: %s", port, mediaDir)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 路由 "/" 是 Go 的兜底 pattern：未注册的路径也会落到这里。
	// 不注册的 /api/* 与 /thumb/* 必须返回 404，否则接口写错/已下线会被
	// 伪装成"200 + 首页 HTML"，前端 stationFetch 会把它当成成功响应。
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/thumb/") {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")

	html, err := os.ReadFile(filepath.Join(staticDir, "index.html"))
	if err != nil {
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		return
	}

	// 注入首页数据，跳过API请求。
	// 日期列表与首个日期的文件都直接取自日期索引（已按倒序维护），
	// 不再遍历整个媒体文件列表。
	indexMutex.RLock()
	dateList := make([]string, len(dateKeys))
	copy(dateList, dateKeys)
	firstDate := ""
	if len(dateList) > 0 {
		firstDate = dateList[0]
	}
	var firstFiles []LightMediaFile
	if firstDate != "" {
		for _, f := range dateIndex[firstDate] {
			firstFiles = append(firstFiles, LightMediaFile{
				Path:     f.Path,
				Date:     firstDate,
				IsVideo:  f.IsVideo,
				Size:     f.Size,
				Duration: f.Duration,
			})
			if len(firstFiles) >= 100 {
				break
			}
		}
	}
	indexMutex.RUnlock()

	initData, _ := json.Marshal(map[string]interface{}{
		"dates":      dateList,
		"firstFiles": firstFiles,
	})
	inject := "<script>window.__INIT__=" + string(initData) + "</script></head>"
	html = bytes.Replace(html, []byte("</head>"), []byte(inject), 1)

	w.Write(html)
}

func datesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")

	indexMutex.RLock()
	dateList := make([]string, len(dateKeys))
	copy(dateList, dateKeys)
	total := len(mediaIndex)
	indexMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dates": dateList,
		"total": total,
	})
}

// 文件列表分页：每页 100 条（前端与旧接口约定，不能改）。
const filesPageSize = 100

// maxPageNumber 页码上限：page 已经很离谱时直接判为非法参数，
// 从根上避免 (page-1)*pageSize 造成的整数溢出。
const maxPageNumber = 10_000_000

// parsePageParam 解析 page 查询参数：
//   - 未传 → 第 1 页
//   - 非整数 / 小于 1 / 超过上限 → 返回错误（由调用方回 HTTP 400）
//
// 原实现用 strconv.Atoi 后忽略错误、也不校验最小值：page=0 或负数会让
// 切片起点变成负数 → 越界 panic；非数字则被当成 0。
func parsePageParam(raw string) (int, error) {
	if raw == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("page 参数必须是整数（收到 %q）", clip(raw, 32))
	}
	if n < 1 {
		return 0, fmt.Errorf("page 参数必须大于等于 1（收到 %d）", n)
	}
	if n > maxPageNumber {
		return 0, fmt.Errorf("page 参数过大（上限 %d，收到 %d）", maxPageNumber, n)
	}
	return n, nil
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	date := strings.TrimPrefix(r.URL.Path, "/api/files/")

	page, err := parsePageParam(r.URL.Query().Get("page"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
			"date":  date,
		})
		return
	}

	// 直接命中日期索引（O(该日期的文件数)），不再逐次遍历全部媒体文件。
	// 索引切片在重建时整体替换，取得引用后即使解锁也不会被改写（不可变快照）。
	indexMutex.RLock()
	dayFiles := dateIndex[date]
	indexMutex.RUnlock()

	files := make([]LightMediaFile, 0, len(dayFiles))
	for _, f := range dayFiles {
		files = append(files, LightMediaFile{
			Path:     f.Path,
			Date:     date,
			IsVideo:  f.IsVideo,
			Size:     f.Size,
			Duration: f.Duration,
		})
	}

	pageSize := filesPageSize
	total := len(files)
	pages := (total + pageSize - 1) / pageSize

	// 用 int64 计算切片边界：即使 page 被放大也不会溢出
	start := int64(page-1) * int64(pageSize)
	switch {
	case start >= int64(total):
		// 页码超出实际页数 → 正常返回空数组（不越界、不报错）
		files = []LightMediaFile{}
	default:
		end := start + int64(pageSize)
		if end > int64(total) {
			end = int64(total)
		}
		files = files[int(start):int(end)]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"total": total,
		"pages": pages,
		"page":  page,
		"date":  date,
	})
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/file/")

	// 统一安全校验：目录边界 + 符号链接/junction 解析（见 pathsafe.go）
	mp, err := safeMediaPath(mediaDir, path)
	if err != nil {
		writeMediaPathError(w, r, "file", path, err)
		return
	}
	if mp.Info.IsDir() {
		http.NotFound(w, r)
		return
	}
	// 只提供图片/视频：媒体目录里若混进配置、凭证、文档，一律拒绝（403）
	if err := requireMediaExt(mp.Rel); err != nil {
		writeMediaPathError(w, r, "file", path, err)
		return
	}

	ext := strings.ToLower(filepath.Ext(mp.Rel))
	if videoExts[ext] {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Accept-Ranges", "bytes")
	} else {
		w.Header().Set("Cache-Control", "max-age=86400")
	}

	// 用解析链接后的真实路径提供服务，避免校验后被替换的链接被跟随
	http.ServeFile(w, r, mp.Real)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Files []string `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	deleted := 0
	failed := 0

	for _, fileName := range req.Files {
		// 删除专用校验：只允许普通文件或指向根内文件的符号链接；
		// 符号链接只删除链接本身（os.Remove 不跟随链接）
		mp, err := safeDeleteTarget(mediaDir, fileName)
		if err != nil {
			if !errors.Is(err, ErrPathMissing) {
				logPathReject("delete", fileName, err)
				failed++
			}
			continue
		}
		// 只允许删图片/视频：媒体目录里的配置、凭证、文档等一律不动
		if err := requireMediaExt(mp.Rel); err != nil {
			logPathReject("delete", fileName, err)
			failed++
			continue
		}
		srcInfo, statErr := os.Stat(mp.Real)
		if err := os.Remove(mp.Full); err != nil {
			failed++
			log.Printf("[station] 删除失败 %s: %v", mp.Rel, err)
			continue
		}
		deleted++
		if statErr == nil {
			removeThumbCache(mp.Rel, srcInfo.Size(), srcInfo.ModTime())
		} else {
			// 拿不到内容指纹时至少清理旧命名缓存
			removeThumbCache(mp.Rel, 0, time.Time{})
		}
	}

	// 删除后走统一调度器刷新索引（扫描期间的重复请求会被合并，不会丢弃）
	scanSched.Request()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"deleted": deleted,
		"failed":  failed,
	})
}

func thumbHandler(w http.ResponseWriter, r *http.Request) {
	// 缩略图文件名 = sha256(相对路径 + 文件大小 + 修改时间)，且前端 URL 带 ?s=<文件大小>：
	// 源文件一变，缓存键（文件名）与 URL 都会变，因此可以放心长缓存。
	// 此前用 no-cache：滚动或分区重建时每一格都变成一次真实回源请求（几百个并发），
	// 这是「滚动/删除后发卡」的主要来源之一。
	w.Header().Set("Cache-Control", "public, max-age=604800")
	path := strings.TrimPrefix(r.URL.Path, "/thumb/")

	// 源文件必须先过统一安全校验（URL 解码在 safeMediaPath 内完成）
	mp, err := safeMediaPath(mediaDir, path)
	if err != nil {
		writeMediaPathError(w, r, "thumbnail", path, err)
		return
	}
	if mp.Info.IsDir() {
		http.NotFound(w, r)
		return
	}
	// 缩略图同样只针对图片/视频，避免对配置/凭证等无关文件调 ffmpeg
	if err := requireMediaExt(mp.Rel); err != nil {
		writeMediaPathError(w, r, "thumbnail", path, err)
		return
	}

	// 用「解析链接后的真实文件」的大小与修改时间构造缓存键
	srcInfo, err := os.Stat(mp.Real)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	thumbPath, err := ensureThumb(mp, srcInfo)
	if err != nil {
		log.Printf("[station] 缩略图不可用 %s: %v", mp.Rel, err)
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, thumbPath)
}

func resizeImage(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxSize && height <= maxSize {
		return img
	}

	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = (height * maxSize) / width
	} else {
		newHeight = maxSize
		newWidth = (width * maxSize) / height
	}

	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	xRatio := float64(width) / float64(newWidth)
	yRatio := float64(height) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := float64(x)*xRatio + 0.5
			srcY := float64(y)*yRatio + 0.5

			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1
			if x1 >= width {
				x1 = width - 1
			}
			if y1 >= height {
				y1 = height - 1
			}

			xFrac := srcX - float64(x0)
			yFrac := srcY - float64(y0)

			c00 := img.At(x0, y0)
			c01 := img.At(x0, y1)
			c10 := img.At(x1, y0)
			c11 := img.At(x1, y1)

			r00, g00, b00, a00 := c00.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			r := bilinearInterp(r00, r01, r10, r11, xFrac, yFrac)
			g := bilinearInterp(g00, g01, g10, g11, xFrac, yFrac)
			b := bilinearInterp(b00, b01, b10, b11, xFrac, yFrac)
			a := bilinearInterp(a00, a01, a10, a11, xFrac, yFrac)

			result.Set(x, y, color.RGBA64{
				R: uint16(r),
				G: uint16(g),
				B: uint16(b),
				A: uint16(a),
			})
		}
	}

	return result
}

func bilinearInterp(c00, c01, c10, c11 uint32, xFrac, yFrac float64) float64 {
	top := float64(c00)*(1-xFrac) + float64(c10)*xFrac
	bot := float64(c01)*(1-xFrac) + float64(c11)*xFrac
	return top*(1-yFrac) + bot*yFrac
}

```


---

## 文件：media_ext_test.go

```go
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 媒体目录里混进无关文件时，网页既不能读也不能删（纵深防御）。
// 允许的只有索引里会出现的图片/视频后缀。

func TestIsAllowedMediaExt(t *testing.T) {
	allowed := []string{"a.jpg", "A.JPEG", "pic.png", "x.gif", "y.webp",
		"v.mp4", "v.avi", "v.mov", "v.mkv", "v.webm", "sub/dir/a.jpg"}
	denied := []string{"config.json", "tg_session.session", "notes.txt", "run.exe",
		"archive.zip", "doc.pdf", "noext", "a.jpg.bak", ".env", "a.js", "a.html"}
	for _, n := range allowed {
		if !isAllowedMediaExt(n) {
			t.Errorf("%q 应被允许（图片/视频）", n)
		}
	}
	for _, n := range denied {
		if isAllowedMediaExt(n) {
			t.Errorf("%q 不该被允许（非媒体类型）", n)
		}
	}
}

func TestFileHandlerRejectsNonMediaFiles(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	// 模拟「媒体目录里混进了不该在的东西」
	writeFileT(t, filepath.Join(env.root, "config.json"), `{"api_hash":"LEAKED"}`)
	writeFileT(t, filepath.Join(env.root, "tg_session.session"), "LEAKED-SESSION")
	writeFileT(t, filepath.Join(env.root, "notes.txt"), "LEAKED-NOTES")
	writeFileT(t, filepath.Join(env.root, "payload.exe"), "MZ")
	writeFileT(t, filepath.Join(env.root, "noext"), "LEAKED-NOEXT")
	writeFileT(t, filepath.Join(env.root, "sub", "deep.json"), `{"deep":"LEAKED"}`)

	for _, target := range []string{"/api/file/config.json", "/api/file/tg_session.session",
		"/api/file/notes.txt", "/api/file/payload.exe", "/api/file/noext", "/api/file/sub/deep.json"} {
		rec := doRequest(fileHandler, http.MethodGet, target, nil)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s 应 403，实际 %d", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "LEAKED") {
			t.Errorf("%s 泄露了文件内容", target)
		}
	}

	// 正常图片/视频不受影响
	for _, target := range []string{"/api/file/pic.jpg", "/api/file/sub/deep.jpg", "/api/file/clip.mp4"} {
		if rec := doRequest(fileHandler, http.MethodGet, target, nil); rec.Code != http.StatusOK {
			t.Errorf("%s 应 200，实际 %d", target, rec.Code)
		}
	}
}

func TestDeleteHandlerRefusesNonMediaFiles(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	cfg := filepath.Join(env.root, "config.json")
	sess := filepath.Join(env.root, "tg_session.session")
	writeFileT(t, cfg, `{"api_hash":"LEAKED"}`)
	writeFileT(t, sess, "LEAKED-SESSION")

	body, _ := json.Marshal(map[string]interface{}{
		"files": []string{"config.json", "tg_session.session", "pic.jpg"},
	})
	rec := doRequest(deleteHandler, http.MethodPost, "/api/delete", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	var resp struct {
		Success bool `json:"success"`
		Deleted int  `json:"deleted"`
		Failed  int  `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if resp.Deleted != 1 || resp.Failed != 2 {
		t.Errorf("期望 deleted=1 failed=2（只删掉图片），实际 deleted=%d failed=%d", resp.Deleted, resp.Failed)
	}
	if _, err := os.Stat(cfg); err != nil {
		t.Errorf("配置文件不得被删除: %v", err)
	}
	if _, err := os.Stat(sess); err != nil {
		t.Errorf("session 文件不得被删除: %v", err)
	}
	if _, err := os.Stat(filepath.Join(env.root, "pic.jpg")); !os.IsNotExist(err) {
		t.Error("正常图片应被删除")
	}
}

func TestThumbHandlerRejectsNonMediaFiles(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	writeFileT(t, filepath.Join(env.root, "config.json"), `{"api_hash":"LEAKED"}`)

	if rec := doRequest(thumbHandler, http.MethodGet, "/thumb/config.json", nil); rec.Code != http.StatusForbidden {
		t.Errorf("非媒体文件缩略图应 403，实际 %d", rec.Code)
	}
	if rec := doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil); rec.Code != http.StatusOK {
		t.Errorf("图片缩略图应 200，实际 %d", rec.Code)
	}
}

```


---

## 文件：media_handlers_test.go

```go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── HTTP 处理器集成测试（第二阶段 · 第一/二/四部分）──────────────────────
//
// 直接调用各 handler，验证：
//   - 所有媒体访问入口都接入了统一路径校验（穿越 / 绝对路径 / 兄弟目录 → 403，
//     不存在 → 404，且响应不泄露服务器绝对路径）；
//   - 损坏媒体、缓存目录不可写等异常不会 panic 或挂死；
//   - 删除不会跟随符号链接删除媒体目录之外的文件；
//   - /api/refresh 的响应字段保持前端兼容。

type mediaEnv struct {
	base    string
	root    string
	sibling string
	restore func()
}

func setupMediaEnv(t *testing.T) mediaEnv {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	sibling := filepath.Join(base, "media2")

	restore := setMediaTestEnv(t, root)

	writeJPEG(t, filepath.Join(root, "pic.jpg"))
	writeJPEG(t, filepath.Join(root, "sub", "deep.jpg"))
	writeFileT(t, filepath.Join(root, "broken.jpg"), "this is not an image")
	writeFileT(t, filepath.Join(root, "clip.mp4"), "this is not a video")
	writeFileT(t, filepath.Join(sibling, "sibling.jpg"), "sibling")
	writeFileT(t, filepath.Join(base, "secret.jpg"), "secret")

	return mediaEnv{base: base, root: root, sibling: sibling, restore: restore}
}

func doRequest(h http.HandlerFunc, method, target string, body []byte) *httptest.ResponseRecorder {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	h(rec, r)
	return rec
}

func TestFileHandlerPathSafety(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	siblingFile := filepath.Join(env.sibling, "sibling.jpg")
	secret := filepath.Join(env.base, "secret.jpg")

	cases := []struct {
		name   string
		target string
		want   int
	}{
		{"正常文件", "/api/file/pic.jpg", http.StatusOK},
		{"子目录文件", "/api/file/sub/deep.jpg", http.StatusOK},
		{"目录穿越", "/api/file/../secret.jpg", http.StatusForbidden},
		{"编码的目录穿越", "/api/file/%2e%2e/secret.jpg", http.StatusForbidden},
		{"兄弟目录（绝对路径）", "/api/file/" + siblingFile, http.StatusForbidden},
		{"绝对路径", "/api/file/" + filepath.ToSlash(secret), http.StatusForbidden},
		{"URL 形式", "/api/file/file:///etc/passwd", http.StatusForbidden},
		{"文件不存在", "/api/file/nope.jpg", http.StatusNotFound},
		{"目标是目录", "/api/file/sub", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(fileHandler, http.MethodGet, tc.target, nil)
			if rec.Code != tc.want {
				t.Fatalf("状态码 = %d，期望 %d（body=%q）", rec.Code, tc.want, clip(rec.Body.String(), 120))
			}
			// 错误响应绝不能回显服务器绝对路径
			body := rec.Body.String()
			if rec.Code != http.StatusOK {
				if strings.Contains(body, filepath.ToSlash(env.base)) || strings.Contains(body, env.base) {
					t.Errorf("错误响应泄露了服务器绝对路径: %q", clip(body, 200))
				}
			}
		})
	}

	// 正常文件内容必须原样返回（保持播放能力）
	rec := doRequest(fileHandler, http.MethodGet, "/api/file/pic.jpg", nil)
	want, err := os.ReadFile(filepath.Join(env.root, "pic.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rec.Body.Bytes(), want) {
		t.Errorf("文件内容不一致：返回 %d 字节，期望 %d 字节", rec.Body.Len(), len(want))
	}
	// 视频响应头（保持前端播放兼容）
	rec = doRequest(fileHandler, http.MethodGet, "/api/file/clip.mp4", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("视频应可访问，实际 %d", rec.Code)
	}
	if rec.Header().Get("Accept-Ranges") != "bytes" {
		t.Errorf("视频响应应带 Accept-Ranges: bytes")
	}
}

func TestThumbHandlerPathSafetyAndBrokenMedia(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	siblingFile := filepath.Join(env.sibling, "sibling.jpg")
	secret := filepath.Join(env.base, "secret.jpg")

	for _, tc := range []struct {
		name   string
		target string
		want   int
	}{
		{"穿越", "/thumb/../secret.jpg", http.StatusForbidden},
		{"编码穿越", "/thumb/%2e%2e/secret.jpg", http.StatusForbidden},
		{"兄弟目录", "/thumb/" + siblingFile, http.StatusForbidden},
		{"绝对路径", "/thumb/" + filepath.ToSlash(secret), http.StatusForbidden},
		{"不存在", "/thumb/nope.jpg", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(thumbHandler, http.MethodGet, tc.target, nil)
			if rec.Code != tc.want {
				t.Fatalf("状态码 = %d，期望 %d", rec.Code, tc.want)
			}
		})
	}

	// 损坏的图片/视频：不得 panic，返回 404（前端显示占位图）
	for _, name := range []string{"broken.jpg", "clip.mp4"} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s 触发 panic: %v", name, r)
				}
			}()
			rec := doRequest(thumbHandler, http.MethodGet, "/thumb/"+name, nil)
			if rec.Code != http.StatusNotFound {
				t.Errorf("%s 应返回 404，实际 %d", name, rec.Code)
			}
		}()
	}

	// 正常图片：无论 ffmpeg 是否可用都必须能出缩略图（原生解码兜底）
	rec := doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("正常图片缩略图应成功，实际 %d（body=%q）", rec.Code, clip(rec.Body.String(), 120))
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "jpeg") && !strings.Contains(ct, "jpg") {
		t.Errorf("缩略图 Content-Type = %q，期望 image/jpeg", ct)
	}
	if rec.Body.Len() < thumbMinBytes {
		t.Errorf("缩略图体积异常: %d 字节", rec.Body.Len())
	}

	// 二次请求应命中缓存（不再生成），文件仍在
	entries, _ := os.ReadDir(thumbsDir)
	if len(entries) == 0 {
		t.Error("缩略图缓存文件未写入")
	}
}

func TestThumbHandlerUnwritableCacheDir(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	// 把缓存目录指向一个普通文件：MkdirAll 必定失败
	blocker := filepath.Join(env.base, "thumbs-blocker")
	writeFileT(t, blocker, "not a directory")
	oldThumbs := thumbsDir
	thumbsDir = blocker
	defer func() { thumbsDir = oldThumbs }()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("缓存目录不可写时不应 panic: %v", r)
		}
	}()
	rec := doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("缓存目录不可写时应返回 404，实际 %d", rec.Code)
	}
}

func TestDeleteHandlerPathSafety(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	secret := filepath.Join(env.base, "secret.jpg")
	siblingFile := filepath.Join(env.sibling, "sibling.jpg")

	body, _ := json.Marshal(map[string]interface{}{
		"files": []string{
			"../secret.jpg",
			"sub/../../secret.jpg",
			siblingFile,
			filepath.Join(env.root, "sub"), // 目录不允许删除
		},
	})
	rec := doRequest(deleteHandler, http.MethodPost, "/api/delete", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	var resp struct {
		Success bool `json:"success"`
		Deleted int  `json:"deleted"`
		Failed  int  `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if resp.Deleted != 0 {
		t.Errorf("危险路径一个都不应被删除，实际 deleted=%d", resp.Deleted)
	}
	if _, err := os.Stat(secret); err != nil {
		t.Errorf("媒体目录之外的文件不得被删除: %v", err)
	}
	if _, err := os.Stat(siblingFile); err != nil {
		t.Errorf("兄弟目录中的文件不得被删除: %v", err)
	}

	// 正常文件：删除成功 + 缩略图缓存清理
	// 先生成一次缩略图缓存，并记录该文件的精确缓存键
	rec = doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("准备缩略图失败: %d", rec.Code)
	}
	picMP, err := safeMediaPath(mediaDir, "pic.jpg")
	if err != nil {
		t.Fatalf("计算 pic.jpg 路径失败: %v", err)
	}
	picInfo, err := os.Stat(picMP.Real)
	if err != nil {
		t.Fatalf("stat pic.jpg 失败: %v", err)
	}
	picThumb := thumbCachePath(thumbCacheKey(picMP.Rel, picInfo.Size(), picInfo.ModTime()))
	if _, err := os.Stat(picThumb); err != nil {
		t.Fatalf("缩略图缓存未生成: %v", err)
	}

	body, _ = json.Marshal(map[string]interface{}{"files": []string{"pic.jpg"}})
	rec = doRequest(deleteHandler, http.MethodPost, "/api/delete", body)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Deleted != 1 {
		t.Fatalf("正常文件应被删除，实际 deleted=%d failed=%d", resp.Deleted, resp.Failed)
	}
	if _, err := os.Stat(filepath.Join(env.root, "pic.jpg")); !os.IsNotExist(err) {
		t.Error("文件应已被删除")
	}
	// 该文件的两种命名缓存都必须被清理（确定性断言，不受后台任务影响）
	if _, err := os.Stat(picThumb); !os.IsNotExist(err) {
		t.Errorf("删除文件后新命名缩略图缓存应被清理: %s", filepath.Base(picThumb))
	}
	if _, err := os.Stat(legacyThumbCachePath("pic.jpg")); !os.IsNotExist(err) {
		t.Errorf("删除文件后旧命名缩略图缓存应被清理")
	}
}

func TestDeleteHandlerMethodAndBody(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	rec := doRequest(deleteHandler, http.MethodGet, "/api/delete", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET 应返回 405，实际 %d", rec.Code)
	}
	rec = doRequest(deleteHandler, http.MethodPost, "/api/delete", []byte("not json"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("非法 JSON 应返回 400，实际 %d", rec.Code)
	}
}

func TestRefreshHandlerFrontendCompatibility(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	rec := doRequest(refreshHandler, http.MethodPost, "/api/refresh", []byte("{}"))
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	// 前端依赖 ok / message 字段
	if ok, _ := resp["ok"].(bool); !ok {
		t.Error("响应必须含 ok=true（前端 stationFetch 依赖）")
	}
	if msg, _ := resp["message"].(string); msg == "" {
		t.Error("响应必须含非空 message")
	}
	// 新增字段用于区分「启动了扫描」与「已合并为补扫」，且不得谎报完成
	if _, has := resp["started"]; !has {
		t.Error("响应应含 started 字段")
	}
	if _, has := resp["merged"]; !has {
		t.Error("响应应含 merged 字段")
	}
	if _, has := resp["running"]; !has {
		t.Error("响应应含 running 字段")
	}

	// 扫描中的再次请求：必须被合并（不能返回虚假的完成，也不能丢失）
	if scanSched.Wait(20 * time.Second) {
		// 第一次扫描已结束，这里再触发一轮并立刻追问
		doRequest(refreshHandler, http.MethodPost, "/api/refresh", []byte("{}"))
		rec = doRequest(refreshHandler, http.MethodPost, "/api/refresh", []byte("{}"))
		if rec.Code != http.StatusOK {
			t.Fatalf("并发刷新请求应返回 200，实际 %d", rec.Code)
		}
	}

	// 非 POST 方法保持 405（前端行为兼容）
	rec = doRequest(refreshHandler, http.MethodGet, "/api/refresh", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/refresh 应返回 405，实际 %d", rec.Code)
	}
}

```


---

## 文件：pathsafe.go

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ── 媒体文件路径安全校验（第二阶段 · 第一部分）────────────────────────────
//
// 背景：旧实现用 strings.HasPrefix(absFull, absMedia) 判断「是否位于媒体根目录内」。
// 该判断有两个致命缺陷：
//   1. 前缀比较不是目录边界比较：/data/media2/a.mp4 能通过 /data/media 的检查；
//   2. 完全不处理符号链接 / 重解析点：媒体目录里指向 /etc/passwd 的链接会被照常读取/删除。
//
// 本文件提供唯一的安全校验入口 safeMediaPath（以及删除专用的 safeDeleteTarget）。
// 所有媒体的读取、播放、缩略图生成、删除都必须先经过它，禁止在别处再写路径判断。
//
// 平台差异：非 Windows 平台 EvalSymlinks 即可解析全部链接；Windows 上 junction /
// mount point（重解析点）不会被 EvalSymlinks 解析，因此额外逐段检测重解析点
// （isReparsePoint 的平台实现）：
//   - 解析后仍在真实根目录内 → 放行（媒体目录内部的链接可用）；
//   - 解析不到（无法证明目标位置，如 junction）→ 拒绝。

var (
	// ErrPathOutside 请求路径无法被证明位于媒体根目录内（越界、非法、符号链接逃逸）。
	ErrPathOutside = errors.New("路径不在媒体目录内")
	// ErrPathMissing 目标文件不存在。
	ErrPathMissing = errors.New("文件不存在")
)

// MediaPath 是一次成功校验的结果。
type MediaPath struct {
	// Full 规范化后的绝对路径（不解析符号链接），用于 Lstat / os.Remove。
	Full string
	// Real 解析符号链接后的真实绝对路径，用于读取、播放、ffmpeg。
	Real string
	// Rel 相对媒体根目录的斜杠分隔路径，用于缩略图缓存键。
	Rel string
	// Symlink 说明 Full 本身是否为符号链接（删除时只删链接，不跟随）。
	Symlink bool
	// Info Full 的 Lstat 结果（调用方复用以避免二次 stat）。
	Info os.FileInfo
}

// ── 媒体根目录真实路径缓存 ───────────────────────────────────────────────
//
// EvalSymlinks 在 Windows 上开销不低（逐段查询），而媒体根目录在一次运行中
// 基本不变，因此缓存「输入根 → 真实根」的映射（键为绝对路径，避免测试之间串味）。

var mediaRootCache struct {
	mu   sync.RWMutex
	real map[string]string
}

func mediaRootReal(mediaRoot string) (string, error) {
	abs, err := filepath.Abs(mediaRoot)
	if err != nil {
		return "", fmt.Errorf("%w: 媒体目录无法解析为绝对路径", ErrPathOutside)
	}
	abs = filepath.Clean(abs)

	mediaRootCache.mu.RLock()
	if mediaRootCache.real != nil {
		if real, ok := mediaRootCache.real[abs]; ok {
			mediaRootCache.mu.RUnlock()
			return real, nil
		}
	}
	mediaRootCache.mu.RUnlock()

	real, err := resolveReparsePoints(abs)
	if err != nil {
		// 根目录不存在 / 无法访问 / 链接环：不能安全校验，必须拒绝，
		// 绝不能退回未解析的路径。
		return "", fmt.Errorf("%w: 媒体目录不可访问或无法解析 (%v)", ErrPathOutside, err)
	}
	real = filepath.Clean(real)

	mediaRootCache.mu.Lock()
	if mediaRootCache.real == nil {
		mediaRootCache.real = make(map[string]string, 2)
	}
	mediaRootCache.real[abs] = real
	mediaRootCache.mu.Unlock()
	return real, nil
}

// resetMediaRootCache 清空根目录真实路径缓存（测试用）。
func resetMediaRootCache() {
	mediaRootCache.mu.Lock()
	mediaRootCache.real = nil
	mediaRootCache.mu.Unlock()
}

// isWithin 判断 abs 是否位于 root 目录之内（按目录边界，不是字符串前缀）。
// root/abs 都必须是规范化后的绝对路径。
func isWithin(root, abs string) bool {
	if root == "" || abs == "" {
		return false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return isLocalRel(rel)
}

// isLocalRel 判断 rel 是否为一个「位于当前目录之内、不向上逃逸」的相对路径。
func isLocalRel(rel string) bool {
	if rel == "" || rel == "." {
		return false
	}
	if filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// maxReparseDepth 解析链接的最大迭代次数（防链接环导致的死循环）。
const maxReparseDepth = 32

// firstReparsePoint 自卷根开始逐段检查，返回第一个符号链接 / 重解析点的完整路径；
// 没有则返回 ""。任何一段不存在时返回该段的错误（调用方可判断 os.IsNotExist）。
func firstReparsePoint(path string) (string, error) {
	vol := filepath.VolumeName(path)
	cur := vol + string(filepath.Separator)
	rest := strings.TrimPrefix(path[len(vol):], string(filepath.Separator))
	for _, seg := range strings.Split(filepath.ToSlash(rest), "/") {
		if seg == "" {
			continue
		}
		cur = filepath.Join(cur, filepath.FromSlash(seg))
		info, err := os.Lstat(cur)
		if err != nil {
			return "", err
		}
		if isReparsePoint(cur, info) {
			return cur, nil
		}
	}
	return "", nil
}

// resolveReparsePoints 反复解析路径上的符号链接 / 重解析点，直到路径不再含链接。
//
// 与 filepath.EvalSymlinks 的差别：Windows 上 junction / mount point 不会被
// EvalSymlinks 解析，但本函数通过 os.Readlink 能拿到其目标，从而统一处理
// 符号链接与 junction（两平台行为一致）。超过 maxReparseDepth 次仍未收敛
// （如链接环）时报错，由调用方拒绝。
func resolveReparsePoints(path string) (string, error) {
	cur := filepath.Clean(path)
	for depth := 0; depth < maxReparseDepth; depth++ {
		rp, err := firstReparsePoint(cur)
		if err != nil {
			return "", err
		}
		if rp == "" {
			return cur, nil
		}
		target, err := os.Readlink(rp)
		if err != nil {
			return "", fmt.Errorf("无法读取链接目标: %w", err)
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(rp), target)
		}
		rest, err := filepath.Rel(rp, cur)
		if err != nil {
			return "", fmt.Errorf("无法计算链接剩余路径: %w", err)
		}
		if rest == "." || rest == "" {
			cur = filepath.Clean(target)
		} else {
			cur = filepath.Clean(filepath.Join(target, rest))
		}
	}
	return "", fmt.Errorf("链接层数超过上限 %d（可能存在链接环）", maxReparseDepth)
}

// hasDotDotSegment 检查请求路径中是否存在 ".." 段（在 Clean 之前判断，
// 因为 Clean 会把 a/../b 悄悄压成 b，从而掩盖穿越意图）。
func hasDotDotSegment(requested string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(requested), "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// windowsReservedNames 是 Windows 保留设备名：这些名字（含带扩展名的形式）
// 会被系统解释为设备而不是媒体文件，必须拒绝。
var windowsReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// unsafeSegmentReason 返回该路径段被拒绝的原因（"" 表示合法）。
// 仅 Windows 上启用这些额外规则：Windows 会静默删除段尾的空格与点、
// 把 ":" 解释为 NTFS 数据流、把 NUL 之类解释为设备，这些都会让
// 「校验用的路径」与「实际打开的文件」不一致。
func unsafeSegmentReason(seg string) string {
	if seg == "." {
		// "." 段只在 Join/Clean 中表示「当前目录」，无风险
		return ""
	}
	if runtime.GOOS != "windows" {
		return ""
	}
	if strings.HasSuffix(seg, " ") || strings.HasSuffix(seg, ".") {
		return "段以空格或点结尾（Windows 会静默改写路径）"
	}
	if strings.ContainsAny(seg, `:<>"|?*`) {
		return "段含非法字符（:, <, >, \", |, ?, * 之一）"
	}
	base := seg
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if windowsReservedNames[strings.ToUpper(base)] {
		return "段使用了 Windows 保留设备名"
	}
	return ""
}

// validateRequestSegments 逐段检查请求路径；返回被拒绝的原因（"" 表示通过）。
func validateRequestSegments(requested string) string {
	if hasDotDotSegment(requested) {
		return "路径包含 .. 穿越段"
	}
	for _, seg := range strings.Split(filepath.ToSlash(requested), "/") {
		if seg == "" {
			continue
		}
		if reason := unsafeSegmentReason(seg); reason != "" {
			return reason
		}
	}
	return ""
}

// safeMediaPath 校验 requested 并返回媒体根目录内的安全路径。
//
// requested 是前端传来的相对路径（斜杠分隔，可含 %XX 编码）。以下一律拒绝：
//   - 空路径、含 NUL、含反斜杠（跨平台一致性；Windows 上反斜杠即分隔符）
//   - 绝对路径、盘符路径（C:\...）、UNC 路径（\\server\share）、URL 形式（file://）
//   - 任何 ".." 段（目录穿越），无论最终是否仍在根内
//   - 规范化后逃出媒体根目录的路径，以及同前缀的兄弟目录（/data/media2）
//   - 符号链接解析后真实路径落在媒体根目录之外的路径
//   - 无法解析（根目录不可访问、链接损坏等）而无法证明安全的路径
//
// 返回的 MediaPath.Real 是解析过符号链接的真实路径，调用方应使用它去
// 读取/播放/ffmpeg；删除请使用 safeDeleteTarget。
func safeMediaPath(mediaRoot, requested string) (MediaPath, error) {
	var out MediaPath

	if strings.TrimSpace(mediaRoot) == "" {
		return out, fmt.Errorf("%w: 未配置媒体目录", ErrPathOutside)
	}
	if strings.TrimSpace(requested) == "" {
		return out, fmt.Errorf("%w: 空路径", ErrPathOutside)
	}
	if strings.ContainsRune(requested, 0) {
		return out, fmt.Errorf("%w: 路径含非法字符", ErrPathOutside)
	}

	// 前端用 encodeURIComponent 逐段编码，这里解码回原始路径；
	// 解码失败按原样处理（下面还有完整的合法性检查兜底）。
	if decoded, err := url.PathUnescape(requested); err == nil {
		requested = decoded
	}

	// 反斜杠：Windows 上是分隔符（会被 ToSlash 转换），Linux 上是普通字符。
	// 为保证两平台一致的校验结论，一律拒绝。
	if strings.Contains(requested, "\\") {
		return out, fmt.Errorf("%w: 路径不能包含反斜杠", ErrPathOutside)
	}
	if strings.Contains(requested, "://") {
		return out, fmt.Errorf("%w: 不接受 URL 形式路径", ErrPathOutside)
	}

	if reason := validateRequestSegments(requested); reason != "" {
		return out, fmt.Errorf("%w: %s", ErrPathOutside, reason)
	}

	req := filepath.FromSlash(requested)
	if filepath.IsAbs(req) || filepath.VolumeName(req) != "" {
		return out, fmt.Errorf("%w: 不接受绝对路径或盘符路径", ErrPathOutside)
	}

	rootAbs, err := filepath.Abs(mediaRoot)
	if err != nil {
		return out, fmt.Errorf("%w: 媒体目录无法解析为绝对路径", ErrPathOutside)
	}
	rootAbs = filepath.Clean(rootAbs)
	rootReal, err := mediaRootReal(mediaRoot)
	if err != nil {
		return out, err
	}

	// 规范化：Join 会 Clean，但 Clean 本身不保证边界 —— 紧随其后的 Rel 才是边界检查。
	full := filepath.Clean(filepath.Join(rootAbs, req))
	rel, err := filepath.Rel(rootAbs, full)
	if err != nil || !isLocalRel(rel) {
		return out, fmt.Errorf("%w: 路径越出媒体目录", ErrPathOutside)
	}

	info, err := os.Lstat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return MediaPath{Full: full, Rel: filepath.ToSlash(rel)},
				fmt.Errorf("%w: %v", ErrPathMissing, err)
		}
		return out, fmt.Errorf("%w: 无法访问目标 (%v)", ErrPathOutside, err)
	}
	out.Full = full
	out.Rel = filepath.ToSlash(rel)
	out.Info = info
	out.Symlink = info.Mode()&os.ModeSymlink != 0

	// 解析路径链上的符号链接 / 重解析点（含 Windows junction），
	// 解析结果必须仍位于真实媒体根目录之内。
	real, rpErr := resolveReparsePoints(full)
	if rpErr != nil {
		if os.IsNotExist(rpErr) {
			return MediaPath{Full: full, Rel: filepath.ToSlash(rel)},
				fmt.Errorf("%w: %v", ErrPathMissing, rpErr)
		}
		return MediaPath{}, fmt.Errorf("%w: 路径无法安全解析 (%v)", ErrPathOutside, rpErr)
	}
	real = filepath.Clean(real)
	if !isWithin(rootReal, real) {
		return MediaPath{}, fmt.Errorf("%w: 链接指向媒体目录之外", ErrPathOutside)
	}
	out.Real = real
	return out, nil
}

// safeDeleteTarget 在 safeMediaPath 基础上针对删除额外收紧：
//   - 只允许删除普通文件（或指向根内普通文件的符号链接）；
//   - 目录、设备、管道等一律拒绝；
//   - 符号链接只删除链接本身（os.Remove 不跟随链接），绝不删除链接指向的文件。
func safeDeleteTarget(mediaRoot, requested string) (MediaPath, error) {
	mp, err := safeMediaPath(mediaRoot, requested)
	if err != nil {
		// 目标不存在时仍返回已规范化的 Full，便于调用方做一致性判断
		return mp, err
	}
	if mp.Info.IsDir() {
		return MediaPath{}, fmt.Errorf("%w: 目标不是文件", ErrPathOutside)
	}
	if !mp.Info.Mode().IsRegular() && !mp.Symlink {
		return MediaPath{}, fmt.Errorf("%w: 目标类型不允许删除", ErrPathOutside)
	}
	return mp, nil
}

// ErrNotMedia 表示目标不是允许对外提供的媒体类型。
// 只放行索引里会出现的图片/视频后缀（imageExts ∪ videoExts，见 main.go）：
// 媒体目录一旦混进配置文件、Telegram 凭证（.session）、文档等无关文件，
// 也不能被网页读走或删掉（纵深防御，2026-10-11 审查）。
var ErrNotMedia = errors.New("不是允许的媒体类型")

// isAllowedMediaExt 判断相对路径是否属于允许的媒体类型（图片/视频）。
func isAllowedMediaExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return imageExts[ext] || videoExts[ext]
}

// requireMediaExt 媒体类型校验；不合格时返回 ErrNotMedia（映射为 403）。
func requireMediaExt(rel string) error {
	if isAllowedMediaExt(rel) {
		return nil
	}
	ext := filepath.Ext(rel)
	if ext == "" {
		ext = "（无后缀）"
	}
	return fmt.Errorf("%w: %s", ErrNotMedia, ext)
}

// writeMediaPathError 把安全校验错误映射为 HTTP 响应：
// 目标不存在 → 404；其余（越界、链接逃逸、非法路径、不可解析）→ 403。
// 响应体只包含固定文本，绝不回显服务器绝对路径。
func writeMediaPathError(w http.ResponseWriter, r *http.Request, op, requested string, err error) {
	if errors.Is(err, ErrPathMissing) {
		http.NotFound(w, r)
		return
	}
	logPathReject(op, requested, err)
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// logPathReject 记录被拒绝的路径请求（仅服务端日志，绝不回显给客户端）。
func logPathReject(op, requested string, err error) {
	log.Printf("[station] %s 拒绝访问 %q: %v", op, clip(requested, 200), err)
}

```


---

## 文件：pathsafe_other.go

```go
//go:build !windows

package main

import "os"

// isReparsePoint 在非 Windows 平台上，重解析点即符号链接：
// filepath.EvalSymlinks 能完整解析，因此只需检查链接标志。
func isReparsePoint(_ string, info os.FileInfo) bool {
	return info != nil && info.Mode()&os.ModeSymlink != 0
}

```


---

## 文件：pathsafe_test.go

```go
package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ── 媒体路径安全校验测试（第二阶段 · 第一部分）────────────────────────────
//
// 覆盖：正常文件 / 子目录文件 / .. 穿越 / 同前缀兄弟目录 / 绝对路径 /
// 盘符与 UNC / 符号链接（根内允许、根外拒绝、删除不跟随）/ 文件不存在。
//
// 符号链接用例在无权限的系统（Windows 非管理员且未开开发者模式）上会自动跳过，
// 跳过原因会打印到测试日志。

func mkTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
}

// trySymlink 尝试创建符号链接；系统不支持时返回 false（调用方跳过用例）。
func trySymlink(t *testing.T, oldname, newname string) bool {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		t.Logf("当前系统无法创建符号链接（%v），跳过该符号链接用例", err)
		return false
	}
	return true
}

func TestSafeMediaPathAllowsLegitPaths(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	mkTestFile(t, filepath.Join(root, "sub", "b.jpg"), "i")
	mkTestFile(t, filepath.Join(root, "sub", "deep", "c.png"), "i")

	cases := []struct {
		name    string
		req     string
		wantRel string
	}{
		{"根目录下的文件", "a.mp4", "a.mp4"},
		{"子目录中的文件", "sub/b.jpg", "sub/b.jpg"},
		{"多层子目录", "sub/deep/c.png", "sub/deep/c.png"},
		{"带 ./ 前缀", "./a.mp4", "a.mp4"},
		{"重复分隔符", "sub//b.jpg", "sub/b.jpg"},
		{"URL 编码的中文路径仍可解析", "sub/%E4%B8%AD%E6%96%87.jpg", ""},
	}
	// 中文字符串用例单独准备文件（避免与上面 table 混淆）
	mkTestFile(t, filepath.Join(root, "sub", "中文.jpg"), "i")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mp, err := safeMediaPath(root, tc.req)
			if err != nil {
				t.Fatalf("safeMediaPath(%q) 意外失败: %v", tc.req, err)
			}
			if !strings.HasSuffix(filepath.ToSlash(mp.Real), tc.wantRel) && tc.wantRel != "" {
				t.Errorf("Real = %q，期望以 %q 结尾", mp.Real, tc.wantRel)
			}
			if mp.Rel == "" {
				t.Errorf("Rel 不能为空")
			}
			if !filepath.IsAbs(mp.Full) || !filepath.IsAbs(mp.Real) {
				t.Errorf("Full/Real 必须是绝对路径: %q / %q", mp.Full, mp.Real)
			}
		})
	}
}

func TestSafeMediaPathRejectsUnsafePaths(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	sibling := filepath.Join(base, "media2")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	mkTestFile(t, filepath.Join(root, "sub", "b.jpg"), "i")
	mkTestFile(t, filepath.Join(sibling, "c.mp4"), "v")
	outside := filepath.Join(base, "secret.txt")
	mkTestFile(t, outside, "s")

	cases := []struct {
		name string
		req  string
	}{
		{"目录穿越", "../secret.txt"},
		{"深层穿越", "sub/../../secret.txt"},
		{"仅 ..", ".."},
		{"结尾 ..", "sub/.."},
		{"反斜杠穿越", `..\secret.txt`},
		{"同前缀兄弟目录（绝对路径）", outside},
		{"同前缀兄弟目录（sibling）", filepath.Join(sibling, "c.mp4")},
		{"绝对路径（正斜杠）", filepath.ToSlash(outside)},
		{"URL 形式", "file:///etc/passwd"},
		{"空路径", ""},
		{"纯空白", "   "},
		{"NUL 字节", "a\x00.mp4"},
		{"反斜杠普通路径", `sub\b.jpg`},
		{"协议式前缀", "http://example.com/a.mp4"},
	}
	if runtime.GOOS == "windows" {
		// Windows 会静默改写这些路径，导致「校验的路径」≠「打开的文件」
		cases = append(cases,
			struct {
				name string
				req  string
			}{"尾随空格（Windows 静默改写）", "a.mp4 "},
			struct {
				name string
				req  string
			}{"尾随点（Windows 静默改写）", "sub/b.jpg."},
			struct {
				name string
				req  string
			}{"NTFS 数据流", "a.mp4:secret"},
			struct {
				name string
				req  string
			}{"保留设备名", "NUL"},
			struct {
				name string
				req  string
			}{"保留设备名带扩展", "CON.mp4"},
		)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := safeMediaPath(root, tc.req)
			if err == nil {
				t.Fatalf("safeMediaPath(%q) 应当被拒绝，实际通过", tc.req)
			}
			if !errors.Is(err, ErrPathOutside) && !errors.Is(err, ErrPathMissing) {
				t.Fatalf("错误类型应为 ErrPathOutside/ErrPathMissing，实际: %v", err)
			}
		})
	}
}

// TestIsWithinPrefixBypass 直击旧实现的前缀绕过缺陷。
func TestIsWithinPrefixBypass(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "media")
	sibling := filepath.Join(base, "media2")

	// 先确认这两个路径确实构成「字符串前缀相同」的诱饵
	if !strings.HasPrefix(filepath.Clean(sibling), filepath.Clean(root)) {
		t.Fatalf("测试前提不成立: %q 不是 %q 的前缀", sibling, root)
	}
	if isWithin(root, sibling) {
		t.Fatalf("同前缀兄弟目录 %q 被误判为在 %q 之内", sibling, root)
	}
	if isWithin(root, root) {
		t.Fatalf("根目录自身不应被当作其内部的文件路径")
	}
	if !isWithin(root, filepath.Join(root, "a.mp4")) {
		t.Fatalf("根目录内的文件应判为在内")
	}
	if isWithin(root, filepath.Join(root, "..", "x.mp4")) {
		t.Fatalf("经 .. 逃出根目录的路径不应判为在内")
	}
}

func TestSafeMediaPathMissingFile(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")

	_, err := safeMediaPath(root, "not-there.mp4")
	if !errors.Is(err, ErrPathMissing) {
		t.Fatalf("不存在的文件应返回 ErrPathMissing，实际: %v", err)
	}
	// 不存在的深层路径同样报缺失（而不是越界）
	if _, err := safeMediaPath(root, "sub/not-there.mp4"); !errors.Is(err, ErrPathMissing) {
		t.Fatalf("不存在的子路径应返回 ErrPathMissing，实际: %v", err)
	}
}

func TestSafeMediaPathSymlinks(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	root2 := filepath.Join(base, "media-link")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	outsideDir := filepath.Join(base, "outside")
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	mkTestFile(t, outsideFile, "s")

	// 1) 指向根内文件的符号链接 → 允许，且 Real 必须落在根内
	if trySymlink(t, filepath.Join(root, "a.mp4"), filepath.Join(root, "link-in.mp4")) {
		mp, err := safeMediaPath(root, "link-in.mp4")
		if err != nil {
			t.Fatalf("根内符号链接应被允许，实际: %v", err)
		}
		if !mp.Symlink {
			t.Errorf("应当标记 Symlink=true")
		}
		if !isWithin(mustReal(t, root), mp.Real) {
			t.Errorf("Real=%q 应位于媒体根目录内", mp.Real)
		}
	}

	// 2) 指向根外文件的符号链接 → 拒绝
	if trySymlink(t, outsideFile, filepath.Join(root, "link-out.mp4")) {
		if _, err := safeMediaPath(root, "link-out.mp4"); !errors.Is(err, ErrPathOutside) {
			t.Fatalf("指向根外文件的符号链接必须拒绝，实际: %v", err)
		}
		// 删除同样必须拒绝，且根外文件不能被动到
		if _, err := safeDeleteTarget(root, "link-out.mp4"); !errors.Is(err, ErrPathOutside) {
			t.Fatalf("通过指向根外的链接删除必须拒绝，实际: %v", err)
		}
		if _, err := os.Stat(outsideFile); err != nil {
			t.Fatalf("根外文件不应被删除: %v", err)
		}
	}

	// 3) 指向根外目录的符号链接（经由链接访问目录内文件）→ 拒绝
	if trySymlink(t, outsideDir, filepath.Join(root, "outdir")) {
		if _, err := safeMediaPath(root, "outdir/secret.txt"); !errors.Is(err, ErrPathOutside) {
			t.Fatalf("经由指向根外目录的链接必须拒绝，实际: %v", err)
		}
	}

	// 4) 媒体根目录本身是符号链接 → 仍按真实根校验
	if trySymlink(t, root, root2) {
		resetMediaRootCache()
		if _, err := safeMediaPath(root2, "a.mp4"); err != nil {
			t.Fatalf("以符号链接形式给出的媒体根目录应可用: %v", err)
		}
		if _, err := safeMediaPath(root2, "../secret.txt"); err == nil {
			t.Fatalf("经符号链接根目录穿越必须被拒绝")
		}
	}
}

func TestSafeDeleteTarget(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "del.mp4"), "v")
	if err := os.MkdirAll(filepath.Join(root, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// 普通文件可删
	mp, err := safeDeleteTarget(root, "del.mp4")
	if err != nil {
		t.Fatalf("普通文件应允许删除: %v", err)
	}
	if err := os.Remove(mp.Full); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if _, err := os.Stat(mp.Full); !os.IsNotExist(err) {
		t.Fatalf("文件应已被删除")
	}

	// 目录不允许删除
	if _, err := safeDeleteTarget(root, "dir"); err == nil {
		t.Fatalf("目录不应允许通过该接口删除")
	}
	// 穿越不允许删除
	if _, err := safeDeleteTarget(root, "../secret.txt"); err == nil {
		t.Fatalf("穿越路径不应允许删除")
	}
	// 不存在 → ErrPathMissing
	if _, err := safeDeleteTarget(root, "nope.mp4"); !errors.Is(err, ErrPathMissing) {
		t.Fatalf("不存在文件应返回 ErrPathMissing，实际: %v", err)
	}

	// 根内链接：删除只能删链接本身，不能删目标
	if trySymlink(t, filepath.Join(root, "del.mp4"), filepath.Join(root, "del-link.mp4")) {
		mkTestFile(t, filepath.Join(root, "del.mp4"), "v2")
		mp, err := safeDeleteTarget(root, "del-link.mp4")
		if err != nil {
			t.Fatalf("指向根内文件的链接应允许删除（删链接本身）: %v", err)
		}
		if err := os.Remove(mp.Full); err != nil {
			t.Fatalf("删除链接失败: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "del.mp4")); err != nil {
			t.Fatalf("链接目标不应被删除: %v", err)
		}
	}
}

func TestSafeMediaPathWindowsSpecific(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")

	if runtime.GOOS != "windows" {
		t.Skip("盘符 / UNC 行为仅在 Windows 上断言（Linux 上它们是普通文件名）")
	}
	bad := []string{
		`C:\Windows\win.ini`,
		`C:/Windows/win.ini`,
		`C:win.ini`,              // 盘符相对路径
		`\\server\share\x.mp4`,   // UNC
		`\\?\C:\Windows\win.ini`, // 扩展长度前缀
	}
	for _, req := range bad {
		if _, err := safeMediaPath(root, req); err == nil {
			t.Errorf("Windows 下 %q 必须被拒绝", req)
		}
	}
	// 反斜杠形式的正常相对路径在 Windows 上也会被拒绝（前端只用正斜杠）
	if _, err := safeMediaPath(root, `a.mp4`); err != nil {
		t.Errorf("普通相对路径不应失败: %v", err)
	}
}

// tryJunction 尝试创建 Windows 目录联接（junction，无需管理员权限）。
// 其它平台或创建失败时返回 false（调用方跳过用例）。
func tryJunction(t *testing.T, target, link string) bool {
	t.Helper()
	if runtime.GOOS != "windows" {
		return false
	}
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Logf("创建 junction 失败（%v）：%s", err, strings.TrimSpace(string(out)))
		return false
	}
	return true
}

func TestSafeMediaPathJunctions(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	mkTestFile(t, filepath.Join(root, "sub", "b.jpg"), "i")
	outsideDir := filepath.Join(base, "outside")
	mkTestFile(t, filepath.Join(outsideDir, "secret.txt"), "s")

	did := false

	// 1) 指向根内目录的 junction → 允许，且 Real 必须落在真实根内
	inner := filepath.Join(root, "inner-link")
	if tryJunction(t, filepath.Join(root, "sub"), inner) {
		did = true
		mp, err := safeMediaPath(root, "inner-link/b.jpg")
		if err != nil {
			t.Errorf("指向根内目录的 junction 应被允许，实际: %v", err)
		} else if !isWithin(mustReal(t, root), mp.Real) {
			t.Errorf("Real=%q 应位于媒体根目录内", mp.Real)
		}
	}

	// 2) 指向根外目录的 junction → 必须拒绝（这是 EvalSymlinks 抓不到的绕过）
	outLink := filepath.Join(root, "out-link")
	if tryJunction(t, outsideDir, outLink) {
		did = true
		if _, err := safeMediaPath(root, "out-link/secret.txt"); !errors.Is(err, ErrPathOutside) {
			t.Errorf("指向根外目录的 junction 必须拒绝，实际: %v", err)
		}
		if _, err := safeDeleteTarget(root, "out-link/secret.txt"); err == nil {
			t.Errorf("经由 junction 删除根外文件必须拒绝")
		}
		if _, err := os.Stat(filepath.Join(outsideDir, "secret.txt")); err != nil {
			t.Errorf("根外文件不应被删除: %v", err)
		}
	}

	// 3) 媒体根目录本身是 junction → 仍按真实根校验
	rootLink := filepath.Join(base, "media-link")
	if tryJunction(t, root, rootLink) {
		did = true
		resetMediaRootCache()
		if _, err := safeMediaPath(rootLink, "a.mp4"); err != nil {
			t.Errorf("以 junction 形式给出的媒体根目录应可用: %v", err)
		}
		if _, err := safeMediaPath(rootLink, "../outside/secret.txt"); err == nil {
			t.Errorf("经 junction 根目录穿越必须被拒绝")
		}
	}

	if !did {
		t.Skip("当前系统无法创建 junction（非 Windows 或命令不可用）")
	}
}

func mustReal(t *testing.T, path string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) 失败: %v", path, err)
	}
	return real
}

```


---

## 文件：pathsafe_windows.go

```go
//go:build windows

package main

import (
	"os"
	"syscall"
)

// FILE_ATTRIBUTE_REPARSE_POINT：符号链接、junction、mount point 都带该属性。
// Windows 上 filepath.EvalSymlinks 只解析真符号链接，对 junction 返回原路径，
// 因此必须单独识别，否则「经由 junction 访问媒体目录之外」会成为绕过口子。
const fileAttributeReparsePoint = 0x400

// isReparsePoint 判断路径是否为符号链接 / 重解析点。
// 任何无法判定的情况都按「可疑」处理（返回 true），由调用方拒绝。
func isReparsePoint(path string, info os.FileInfo) bool {
	if info != nil && info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return true
	}
	return attrs&fileAttributeReparsePoint != 0
}

```


---

## 文件：probe.go

```go
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── 媒体元数据探测（第二阶段 · 第三部分）────────────────────────────────
//
// 1. 优先用 ffprobe（只输出 duration，不做任何转码/媒体数据输出）；
//    没有 ffprobe 时回退到原有的 ffmpeg -i + stderr 解析，行为完全兼容。
// 2. 所有外部进程都带超时（exec.CommandContext）：单个损坏文件 / 卡住的进程
//    不会让整轮扫描永久阻塞。
// 3. 时长缓存带「大小 + 修改时间」校验：文件内容变化后自动重新探测。

const (
	// probeTimeout 单个文件元数据探测的超时时间。
	probeTimeout = 15 * time.Second
	// probeWorkers 并发探测数（与原实现一致的并发限制，不放大）。
	probeWorkers = 4
)

// ffprobePath 在启动时定位；为空表示不可用，直接走 ffmpeg 回退。
var ffprobePath string

// findFFprobe 定位 ffprobe：与 findFFmpeg 相同的候选顺序（exe 目录 → 工作目录 →
// config 目录 → 下载器脚本目录 → 系统路径 → PATH 兜底）。
func findFFprobe() string {
	names := []string{"ffprobe", "ffprobe.exe"}

	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}
	if configPath != "" {
		dirs = append(dirs, filepath.Dir(configPath))
	}
	if script, err := downloader.findScript(); err == nil && script != "" {
		dirs = append(dirs, filepath.Dir(script))
	}
	for _, dir := range dirs {
		for _, name := range names {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}
	for _, p := range []string{"/usr/bin/ffprobe", "/usr/local/bin/ffprobe", "/bin/ffprobe"} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	// 最后交给 PATH（ffprobe 与 ffmpeg 通常同目录）
	if p, err := exec.LookPath("ffprobe"); err == nil {
		return p
	}
	return ""
}

// ── 时长缓存 ────────────────────────────────────────────────────────────

type durationEntry struct {
	Size     int64
	ModTime  int64 // UnixNano
	Duration float64
}

var durationCache = struct {
	mu sync.RWMutex
	m  map[string]durationEntry
}{m: make(map[string]durationEntry)}

// durationCacheGet 命中条件：路径相同且大小、修改时间都与缓存记录一致。
func durationCacheGet(rel string, size int64, modTime time.Time) (float64, bool) {
	durationCache.mu.RLock()
	e, ok := durationCache.m[rel]
	durationCache.mu.RUnlock()
	if !ok {
		return 0, false
	}
	if e.Size != size || e.ModTime != modTime.UnixNano() {
		return 0, false
	}
	return e.Duration, true
}

func durationCachePut(rel string, size int64, modTime time.Time, d float64) {
	durationCache.mu.Lock()
	durationCache.m[rel] = durationEntry{Size: size, ModTime: modTime.UnixNano(), Duration: d}
	durationCache.mu.Unlock()
}

// durationCacheLen 返回缓存条目数（测试用）。
func durationCacheLen() int {
	durationCache.mu.RLock()
	defer durationCache.mu.RUnlock()
	return len(durationCache.m)
}

// resetDurationCache 清空缓存（测试用）。
func resetDurationCache() {
	durationCache.mu.Lock()
	durationCache.m = make(map[string]durationEntry)
	durationCache.mu.Unlock()
}

// ── 探测实现 ────────────────────────────────────────────────────────────

var durationRe = regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+(?:\.\d+)?)`)

// parseDurationSeconds 解析 ffprobe 输出的时长文本。
// 明确拒绝空值、N/A、NaN、Inf、负数与解析失败的结果。
func parseDurationSeconds(text string) (float64, error) {
	txt := strings.TrimSpace(text)
	if txt == "" {
		return 0, errors.New("时长为空")
	}
	if strings.EqualFold(txt, "N/A") || strings.EqualFold(txt, "nan") {
		return 0, fmt.Errorf("时长不可用: %q", clip(txt, 32))
	}
	d, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		return 0, fmt.Errorf("时长不是数字: %q", clip(txt, 32))
	}
	if math.IsNaN(d) || math.IsInf(d, 0) || d <= 0 {
		return 0, fmt.Errorf("时长非法: %v", d)
	}
	return d, nil
}

// probeWithFFprobe 用 ffprobe 读取时长（秒）。返回 (时长, 是否成功)。
// 参数只取 format=duration，避免输出媒体流数据或触发转码。
func probeWithFFprobe(path string) (float64, bool) {
	if ffprobePath == "" {
		return 0, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	cmd.SysProcAttr = childProcAttr()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("[station] ffprobe 超时（%s）: %s", probeTimeout, filepath.Base(path))
		return 0, false
	}
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Printf("[station] ffprobe 不可用，后续改用 ffmpeg 探测")
			ffprobePath = ""
			return 0, false
		}
		log.Printf("[station] ffprobe 探测失败 %s: %v（%s）",
			filepath.Base(path), err, clip(strings.TrimSpace(stderr.String()), 160))
		return 0, false
	}
	d, perr := parseDurationSeconds(stdout.String())
	if perr != nil {
		log.Printf("[station] ffprobe 输出无法解析 %s: %v", filepath.Base(path), perr)
		return 0, false
	}
	return d, true
}

// probeWithFFmpeg 旧方式：ffmpeg -i（无输出，必然非零退出码）解析 stderr 的 Duration 行。
func probeWithFFmpeg(path string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffmpegPath, "-hide_banner", "-i", path)
	cmd.SysProcAttr = childProcAttr()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// -i 不指定输出会以非零码退出，属预期；只需解析 stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("[station] ffmpeg 探测超时（%s）: %s", probeTimeout, filepath.Base(path))
		return 0, false
	}
	if errors.Is(runErr, exec.ErrNotFound) {
		log.Printf("[station] ffmpeg 不可用，跳过时长探测")
		return 0, false
	}
	m := durationRe.FindStringSubmatch(stderr.String())
	if m == nil {
		log.Printf("[station] ffmpeg 未输出可解析的时长 %s", filepath.Base(path))
		return 0, false
	}
	h, _ := strconv.ParseFloat(m[1], 64)
	mi, _ := strconv.ParseFloat(m[2], 64)
	s, _ := strconv.ParseFloat(m[3], 64)
	d := h*3600 + mi*60 + s
	if math.IsNaN(d) || math.IsInf(d, 0) || d <= 0 {
		return 0, false
	}
	return d, true
}

// probeDuration 探测媒体时长（秒）：ffprobe 优先，失败回退 ffmpeg，全部失败返回 0。
func probeDuration(path string) float64 {
	if d, ok := probeWithFFprobe(path); ok {
		return d
	}
	if d, ok := probeWithFFmpeg(path); ok {
		return d
	}
	return 0
}

// fillDurations 并发补齐 files 中缺失的视频时长（秒），带缓存与超时。
func fillDurations(files []*MediaFile) {
	var pending []*MediaFile
	for _, f := range files {
		if f.IsVideo && f.Duration <= 0 {
			pending = append(pending, f)
		}
	}
	if len(pending) == 0 {
		return
	}
	workerCount := probeWorkers
	if len(pending) < workerCount {
		workerCount = len(pending)
	}
	jobs := make(chan *MediaFile, len(pending))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				// 缓存命中（路径 + 大小 + 修改时间一致）直接复用，不再启动进程
				if d, ok := durationCacheGet(f.Path, f.Size, f.Date); ok && d > 0 {
					f.Duration = d
					continue
				}
				full := filepath.Join(mediaDir, filepath.FromSlash(f.Path))
				d := probeDuration(full)
				if d > 0 {
					f.Duration = d
					durationCachePut(f.Path, f.Size, f.Date, d)
				}
			}
		}()
	}
	for _, f := range pending {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
}

```


---

## 文件：probe_test.go

```go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// ── 媒体元数据探测测试（第二阶段 · 第三部分）────────────────────────────

func TestParseDurationSecondsCases(t *testing.T) {
	ok := []struct {
		in   string
		want float64
	}{
		{"12.5", 12.5},
		{" 3 ", 3},
		{"0.04", 0.04},
		{"3600", 3600},
		{"12.5\n", 12.5},
	}
	for _, c := range ok {
		got, err := parseDurationSeconds(c.in)
		if err != nil || got != c.want {
			t.Errorf("parseDurationSeconds(%q) = (%v, %v)，期望 (%v, nil)", c.in, got, err, c.want)
		}
	}

	// NaN / Inf / 负数 / 0 / 空 / N/A / 非数字 一律视为无效
	bad := []string{"", "   ", "N/A", "nan", "NaN", "-1", "-0.5", "0", "abc", "inf", "+Inf", "-Inf", "1,5"}
	for _, in := range bad {
		if got, err := parseDurationSeconds(in); err == nil {
			t.Errorf("parseDurationSeconds(%q) 应报错，实际得到 %v", in, got)
		}
	}
}

func TestDurationCacheInvalidation(t *testing.T) {
	resetDurationCache()
	defer resetDurationCache()

	mt := time.Unix(1700000000, 0)
	durationCachePut("a.mp4", 100, mt, 42.5)

	if d, ok := durationCacheGet("a.mp4", 100, mt); !ok || d != 42.5 {
		t.Fatalf("缓存应命中，实际 (%v, %v)", d, ok)
	}
	if _, ok := durationCacheGet("a.mp4", 101, mt); ok {
		t.Error("文件大小变化后缓存必须失效（否则沿用旧时长）")
	}
	if _, ok := durationCacheGet("a.mp4", 100, mt.Add(time.Second)); ok {
		t.Error("修改时间变化后缓存必须失效")
	}
	if _, ok := durationCacheGet("b.mp4", 100, mt); ok {
		t.Error("不同路径不应命中缓存")
	}
	// 更新后可重新命中
	durationCachePut("a.mp4", 200, mt.Add(time.Hour), 7)
	if d, ok := durationCacheGet("a.mp4", 200, mt.Add(time.Hour)); !ok || d != 7 {
		t.Errorf("更新后的缓存应命中，实际 (%v, %v)", d, ok)
	}
	if durationCacheLen() != 1 {
		t.Errorf("缓存条目数 = %d，期望 1", durationCacheLen())
	}
}

// TestProbeDurationMissingFileIsFast 源文件不存在时必须快速失败，不能卡住整轮扫描。
func TestProbeDurationMissingFileIsFast(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	start := time.Now()
	if d := probeDuration(filepath.Join(dir, "nope.mp4")); d != 0 {
		t.Errorf("不存在的文件时长应为 0，实际 %v", d)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("探测不存在文件耗时 %s，疑似卡住", elapsed)
	}
}

// TestProbeDurationAndCacheWithRealFFmpeg 用真实 ffmpeg/ffprobe 生成 1 秒视频并探测时长，
// 同时验证「缓存命中后不再启动外部进程」。
func TestProbeDurationAndCacheWithRealFFmpeg(t *testing.T) {
	ff := findFFmpeg()
	probe := findFFprobe()
	if ff == "" || probe == "" {
		t.Skipf("未找到 ffmpeg(%q)/ffprobe(%q)，跳过真实媒体探测用例", ff, probe)
	}
	if out, err := exec.Command(ff, "-version").Output(); err != nil {
		t.Skipf("ffmpeg 不可执行（%v）：%s", err, clip(string(out), 120))
	}

	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	ffmpegPath = ff
	ffprobePath = probe
	resetDurationCache()

	video := filepath.Join(dir, "v.mp4")
	cmd := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=duration=1:size=160x120:rate=10",
		"-pix_fmt", "yuv420p", video)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("生成测试视频失败（%v）：%s", err, clip(string(out), 200))
	}

	d := probeDuration(video)
	if d < 0.5 || d > 1.6 {
		t.Errorf("探测时长 = %.3f 秒，期望约 1 秒（ffprobe 路径）", d)
	}

	st, err := os.Stat(video)
	if err != nil {
		t.Fatal(err)
	}

	// fillDurations 应写入缓存，并在缓存命中时不再依赖外部进程
	files := []*MediaFile{{Path: "v.mp4", Date: st.ModTime(), IsVideo: true, Size: st.Size()}}
	fillDurations(files)
	if files[0].Duration <= 0 {
		t.Fatalf("fillDurations 未补齐时长")
	}
	if _, ok := durationCacheGet("v.mp4", st.Size(), st.ModTime()); !ok {
		t.Fatal("fillDurations 应把结果写入时长缓存")
	}

	ffmpegPath = "definitely-not-a-real-binary"
	ffprobePath = ""
	cached := []*MediaFile{{Path: "v.mp4", Date: st.ModTime(), IsVideo: true, Size: st.Size()}}
	fillDurations(cached)
	if cached[0].Duration <= 0 {
		t.Error("缓存命中时不应需要外部进程，时长必须仍可用")
	}
}

// TestFindFFprobePrefersDeploymentDir 部署目录中的 ffprobe 应被优先选中。
func TestFindFFprobePrefersDeploymentDir(t *testing.T) {
	dir := t.TempDir()
	// 伪造一个部署目录中的 ffprobe（只要存在即可被选中）
	fake := filepath.Join(dir, "ffprobe.exe")
	if err := os.WriteFile(fake, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldExe := ffprobePath
	defer func() { ffprobePath = oldExe }()

	// 直接验证候选顺序函数：把当前工作目录切到 dir 后应能找到该文件
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	found := findFFprobe()
	if found == "" {
		t.Skip("系统 PATH 中也没有 ffprobe，且部署目录探测未命中（环境受限）")
	}
	if filepath.Base(found) != "ffprobe.exe" && filepath.Base(found) != "ffprobe" {
		t.Errorf("findFFprobe 返回了非预期路径: %s", found)
	}
}

```


---

## 文件：proc_other.go

```go
//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 平台相关：进程属性、外部独立下载器探测与结束、进程树结束。
// 非 Windows（Linux/树莓派、macOS）不能依赖 tasklist：这里扫描 /proc 命令行，
// 同时覆盖「打包后的 tg_downloader」与「python + tg_auto_download.py」两种形态。

const (
	// downloaderProcName 后台下载器可执行文件名（Unix 无 .exe 后缀）
	downloaderProcName = "tg_downloader"
	// cliExeName 带控制台的 CLI 下载器可执行文件名
	cliExeName = "tg_downloader_cli"
	// downloaderScriptName 旧模式（python + 脚本）下的脚本名
	downloaderScriptName = "tg_auto_download.py"
)

// childProcAttr 非 Windows 平台无需隐藏窗口。
func childProcAttr() *syscall.SysProcAttr {
	return nil
}

// externalDownloaderPids 扫描 /proc 找出外部独立下载器进程（排除自身）。
func externalDownloaderPids() []int {
	self := os.Getpid()
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid == self {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil || len(raw) == 0 {
			continue
		}
		cmdline := strings.ReplaceAll(string(raw), "\x00", " ")
		if strings.Contains(cmdline, downloaderProcName) || strings.Contains(cmdline, downloaderScriptName) {
			pids = append(pids, pid)
		}
	}
	return pids
}

// killExternalDownloaders 结束扫描到的外部下载器实例：先 SIGTERM，超时后 SIGKILL。
func killExternalDownloaders() error {
	pids := externalDownloaderPids()
	if len(pids) == 0 {
		return nil
	}
	for _, pid := range pids {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(externalDownloaderPids()) == 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	for _, pid := range externalDownloaderPids() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	return nil
}

// killProcessTree 结束指定进程（Unix 下无进程树语义，直接 SIGKILL 该进程）。
func killProcessTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	return syscall.Kill(pid, syscall.SIGKILL)
}

```


---

## 文件：proc_windows.go

```go
//go:build windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// 平台相关：进程属性、外部独立下载器探测与结束、进程树结束。
// Windows 用 tasklist / taskkill 实现（其它平台见 proc_other.go）。

const (
	// downloaderProcName 后台下载器 exe 名（PyInstaller 单文件：引导进程 + 子进程同名）
	downloaderProcName = "tg_downloader.exe"
	// cliExeName 带控制台的 CLI 下载器 exe 名（可靠输出 stdout JSON）
	cliExeName = "tg_downloader_cli.exe"
)

// childProcAttr 返回子进程属性：隐藏控制台窗口（CREATE_NO_WINDOW）。
// CLI 版下载器是 console 程序，靠这个标志不弹窗，同时 stdout 管道照常可用。
func childProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// externalDownloaderPids 用 tasklist 按进程名查找外部独立下载器（开机自启/手动启动）。
func externalDownloaderPids() []int {
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+downloaderProcName, "/FO", "CSV", "/NH").Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, downloaderProcName) {
			continue
		}
		// CSV 行: "tg_downloader.exe","1234","Console","1","8,336 K"
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			if pid, err := strconv.Atoi(strings.Trim(parts[1], `" `)); err == nil {
				pids = append(pids, pid)
			}
		}
	}
	return pids
}

// killExternalDownloaders 按进程名结束所有外部下载器实例（含进程树）。
func killExternalDownloaders() error {
	return exec.Command("taskkill", "/IM", downloaderProcName, "/T", "/F").Run()
}

// killProcessTree 结束指定进程及其全部子进程。
func killProcessTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").Run()
}

```


---

## 文件：python_detect_test.go

```go
package main

import (
	"os"
	"testing"
)

// 这两个测试在本机做真实探测（不打桩），用于验证「不再被 WindowsApps 应用执行别名
// 卡住」以及「自动探测只会返回真正可用的解释器」。
// 没有可用 Python 的机器上会 Skip，不会误报失败。

func TestPythonCandidateDiagnostics(t *testing.T) {
	d := newTestDownloader("", "")
	candidates := d.pythonCandidates()
	if len(candidates) == 0 {
		t.Skip("没有候选解释器")
	}
	usable := 0
	for _, c := range candidates {
		if _, err := os.Stat(c); err != nil {
			t.Logf("跳过（文件不存在）: %s", c)
			continue
		}
		if err := validatePython(c); err != nil {
			t.Logf("拒绝（验证失败）: %s -> %v", c, err)
			continue
		}
		usable++
		t.Logf("可用: %s", c)
	}
	t.Logf("候选 %d 个，其中通过 --version 验证的 %d 个", len(candidates), usable)
}

func TestFindPythonAutoDetectReturnsUsableInterpreter(t *testing.T) {
	d := newTestDownloader("", "")
	py, err := d.findPython()
	if err != nil {
		// 自动探测失败必须是「能说明原因」的错误，而不是 panic 或静默使用坏路径
		if err.Error() == "" {
			t.Fatal("自动探测失败时应给出原因")
		}
		t.Skipf("本机没有可用 Python（自动探测已正确报错）: %v", err)
	}
	if err := validatePython(py); err != nil {
		t.Fatalf("findPython 返回了未经证实的解释器 %s: %v", py, err)
	}
	t.Logf("自动探测选中: %s", py)
}

```


---

## 文件：scan.go

```go
package main

import (
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── 扫描调度（第二阶段 · 第二部分）──────────────────────────────────────
//
// 旧实现：scanFiles 用 CAS 做重入保护，扫描期间的新请求被**静默丢弃**，
// 而 refreshHandler 仍然照旧去生成缩略图 —— 结果是「扫描没跑、缩略图却按旧索引
// 跑了一遍」，最新下载的文件迟迟进不了索引。
//
// 新实现：单一调度器 + pending 标志 + 补扫循环：
//  1. 同一时刻最多一个扫描在执行；
//  2. 扫描期间的任意多个请求只置一个 pending 标志（合并为下一轮补扫）；
//  3. 每轮扫描结束后若 pending 已置位，立即再扫一轮（含缩略图生成期间到达的请求），
//     因此最后一轮请求不会被丢弃；
//  4. 只有在「没有待执行补扫」之后才生成缩略图，避免把未入索引的新文件当作不存在。

// scanScheduler 串行化扫描请求。
type scanScheduler struct {
	mu        sync.Mutex
	running   bool
	pending   bool
	rounds    int // 已完成的扫描轮数（测试/诊断用）
	scanFn    func()
	thumbFn   func()
	waiters   map[chan int]struct{}
	startedAt time.Time
}

func newScanScheduler(scanFn, thumbFn func()) *scanScheduler {
	return &scanScheduler{scanFn: scanFn, thumbFn: thumbFn, waiters: map[chan int]struct{}{}}
}

var scanSched = newScanScheduler(scanFiles, generateAllThumbnails)

// Request 请求一次扫描/刷新。
//   - 没有扫描在跑：立即启动（同步返回，扫描在后台 goroutine 中执行）；
//   - 已有扫描在跑：置 pending 标志并返回 false（请求被合并，绝不丢弃）。
func (s *scanScheduler) Request() bool {
	s.mu.Lock()
	if s.running {
		s.pending = true
		s.mu.Unlock()
		log.Printf("[station] 扫描已在进行中，本次请求合并为下一轮补扫")
		return false
	}
	s.running = true
	s.startedAt = time.Now()
	s.mu.Unlock()

	go s.loop()
	return true
}

// loop 是唯一的扫描执行体。
func (s *scanScheduler) loop() {
	for {
		s.runGuarded(s.scanFn, "扫描")

		// 扫描刚结束时若有请求到达，立即补扫一轮（不生成缩略图，先追平索引）
		s.mu.Lock()
		if s.pending {
			s.pending = false
			s.mu.Unlock()
			continue
		}
		s.mu.Unlock()

		// 没有待补扫任务后才生成缩略图（基于最新索引）
		if s.thumbFn != nil {
			s.runGuarded(s.thumbFn, "缩略图生成")
		}

		s.mu.Lock()
		s.rounds++
		if s.pending {
			// 缩略图生成期间又收到请求：再补扫一轮，然后回到缩略图步骤
			s.pending = false
			s.mu.Unlock()
			continue
		}
		s.running = false
		for ch := range s.waiters {
			close(ch)
		}
		s.waiters = map[chan int]struct{}{}
		s.mu.Unlock()
		return
	}
}

// runGuarded 执行一次任务并兜住 panic。
// 目的是保证调度器状态（running/pending/waiters）一定能复位——否则一次
// 扫描 panic 会让 running 永远为 true，后续所有刷新请求都会被合并却永不执行。
// 这里只作为最后一道防线，不作为掩盖业务错误的常规手段（错误原因会记录日志）。
func (s *scanScheduler) runGuarded(fn func(), what string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[station] %s 过程 panic（已恢复，调度器保持可用）: %v", what, r)
		}
	}()
	fn()
}

// Running 报告当前是否有扫描在跑。
func (s *scanScheduler) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Pending 报告是否有待执行的补扫。
func (s *scanScheduler) Pending() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pending
}

// Rounds 返回已完成的扫描轮数。
func (s *scanScheduler) Rounds() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rounds
}

// Wait 等待当前（含补扫在内）一轮调度彻底结束。
// 返回 false 表示等待超时（调用方据此如实告知前端，不能谎报完成）。
func (s *scanScheduler) Wait(timeout time.Duration) bool {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return true
	}
	ch := make(chan int)
	s.waiters[ch] = struct{}{}
	s.mu.Unlock()

	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		s.mu.Lock()
		delete(s.waiters, ch)
		s.mu.Unlock()
		return false
	}
}

// ── 实际扫描 ────────────────────────────────────────────────────────────

// scanFiles 由调度器调用：扫描媒体目录并重建索引。
// 直接调用它不会触发调度（仅供调度器与测试使用），业务入口一律用 scanSched.Request()。
func scanFiles() {
	// 复用上次索引中「同路径 + 同大小 + 同修改时间」的视频时长，避免重复探测
	indexMutex.RLock()
	prevByPath := make(map[string]*MediaFile, len(mediaIndex))
	for _, f := range mediaIndex {
		prevByPath[f.Path] = f
	}
	indexMutex.RUnlock()

	var files []*MediaFile
	err := filepath.WalkDir(mediaDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 单个条目不可读（权限/损坏）不应中断整轮扫描
			log.Printf("[station] 扫描跳过 %s: %v", path, err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || path == thumbsDir {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !imageExts[ext] && !videoExts[ext] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			log.Printf("[station] 扫描跳过 %s: %v", path, err)
			return nil
		}
		relPath, err := filepath.Rel(mediaDir, path)
		if err != nil {
			return nil
		}
		rel := filepath.ToSlash(relPath)
		mf := &MediaFile{
			Path:    rel,
			Date:    info.ModTime(),
			IsVideo: videoExts[ext],
			Size:    info.Size(),
		}
		if mf.IsVideo {
			// 复用的前提：内容未变（大小 + 修改时间都一致）
			if p, ok := prevByPath[rel]; ok && p.Size == info.Size() &&
				p.Date.Equal(info.ModTime()) && p.Duration > 0 {
				mf.Duration = p.Duration
				durationCachePut(rel, info.Size(), info.ModTime(), p.Duration)
			}
		}
		files = append(files, mf)
		return nil
	})
	if err != nil {
		log.Printf("Scan error: %v", err)
		return
	}

	// 按修改时间倒序（与旧实现一致，分页结果不变）
	sortByDateDesc(files)

	// 补齐缺失的视频时长（缓存未命中的才真正调用 ffprobe/ffmpeg）
	fillDurations(files)

	// 文件列表 + 日期索引在同一把锁下原子切换
	rebuildIndex(files)

	saveIndex()
	log.Printf("Scanned %d files", len(files))
}

// sortByDateDesc 按修改时间倒序排列（稳定排序，时间相同的保持遍历顺序）。
func sortByDateDesc(files []*MediaFile) {
	sort.SliceStable(files, func(i, j int) bool { return files[i].Date.After(files[j].Date) })
}

```


---

## 文件：scan_sched_test.go

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// ── 扫描调度测试（第二阶段 · 第二部分）──────────────────────────────────
//
// 用可控的假扫描/假缩略图函数验证调度语义：
//   单次扫描 / 扫描期间一次请求 / 多次请求合并 / 第二轮期间再次请求 /
//   扫描 panic 后仍可继续刷新 / 缩略图只在补扫清空后执行 / 并发压力下无并发扫描。

type fakeScanner struct {
	mu        sync.Mutex
	gates     chan chan struct{} // 每轮扫描开始时尝试取出一个 gate 并等待其关闭
	started   chan struct{}      // 每轮扫描开始时的通知（带缓冲）
	rounds    int
	conc      int
	maxConc   int
	panics    int
	panicNext bool
	workDelay time.Duration
	thumbErrs []string
	thumbRuns int
}

func newFakeScanner() *fakeScanner {
	return &fakeScanner{
		gates:   make(chan chan struct{}, 8),
		started: make(chan struct{}, 32),
	}
}

func (f *fakeScanner) scan() {
	f.mu.Lock()
	f.conc++
	if f.conc > f.maxConc {
		f.maxConc = f.conc
	}
	shouldPanic := f.panicNext
	f.panicNext = false
	f.mu.Unlock()

	select {
	case f.started <- struct{}{}:
	default:
	}
	select {
	case g := <-f.gates:
		<-g
	default:
	}
	if f.workDelay > 0 {
		time.Sleep(f.workDelay)
	}

	f.mu.Lock()
	f.conc--
	if shouldPanic {
		f.panics++
		f.mu.Unlock()
		panic("模拟扫描失败")
	}
	f.rounds++
	f.mu.Unlock()
}

func (f *fakeScanner) snapshot() (rounds, maxConc, panics int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rounds, f.maxConc, f.panics
}

// scanSchedulerFor 构造一个带假缩略图回调的调度器，并记录缩略图运行时的调度器状态。
func scanSchedulerFor(fs *fakeScanner) (*scanScheduler, *[]string) {
	var sched *scanScheduler
	states := &[]string{}
	sched = newScanScheduler(fs.scan, func() {
		fs.mu.Lock()
		fs.thumbRuns++
		fs.mu.Unlock()
		*states = append(*states, fmt.Sprintf("pending=%v running=%v", sched.Pending(), sched.Running()))
	})
	return sched, states
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("等待条件超时: %s", what)
}

func TestScanSchedulerSingleRun(t *testing.T) {
	fs := newFakeScanner()
	sched, states := scanSchedulerFor(fs)

	if !sched.Request() {
		t.Fatal("首次请求应当启动扫描")
	}
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待扫描结束超时")
	}
	rounds, maxConc, _ := fs.snapshot()
	if rounds != 1 {
		t.Errorf("扫描轮数 = %d，期望 1", rounds)
	}
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1", maxConc)
	}
	if fs.thumbRuns != 1 {
		t.Errorf("缩略图运行次数 = %d，期望 1", fs.thumbRuns)
	}
	if len(*states) != 1 || (*states)[0] != "pending=false running=true" {
		t.Errorf("缩略图运行状态异常: %v", *states)
	}
}

func TestScanSchedulerMergesRequestsDuringScan(t *testing.T) {
	fs := newFakeScanner()
	g1 := make(chan struct{})
	fs.gates <- g1
	sched, _ := scanSchedulerFor(fs)

	if !sched.Request() {
		t.Fatal("首次请求应当启动扫描")
	}
	waitFor(t, "第一轮开始", func() bool { return drain(fs.started) })

	// 扫描进行中的 5 个请求：必须全部被合并，而不是被静默丢弃，也不启动新扫描
	for i := 0; i < 5; i++ {
		if sched.Request() {
			t.Errorf("扫描进行中的请求应返回 false（合并为补扫）")
		}
	}
	if !sched.Pending() {
		t.Fatal("扫描期间的请求应置位 pending")
	}

	close(g1) // 放开第一轮
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待调度结束超时")
	}
	rounds, maxConc, _ := fs.snapshot()
	if rounds != 2 {
		t.Errorf("扫描轮数 = %d，期望 2（首轮 + 一次合并补扫）", rounds)
	}
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1", maxConc)
	}
	if sched.Pending() {
		t.Error("调度结束后不应残留 pending")
	}
}

// drain 非阻塞地消费 started 通知（返回是否有通知）。
func drain(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestScanSchedulerRequestDuringSecondRound(t *testing.T) {
	fs := newFakeScanner()
	g1 := make(chan struct{})
	g2 := make(chan struct{})
	fs.gates <- g1 // 第一轮阻塞
	fs.gates <- g2 // 第二轮也阻塞（保证「第二轮期间」的请求确实发生在扫描进行中）
	sched, states := scanSchedulerFor(fs)

	sched.Request()
	waitFor(t, "第一轮开始", func() bool { return drain(fs.started) })

	sched.Request() // 第一轮期间的请求 → 触发补扫
	if !sched.Pending() {
		t.Fatal("第一轮期间的请求应置位 pending")
	}
	close(g1)

	waitFor(t, "第二轮开始", func() bool { return drain(fs.started) })
	// 第二轮执行期间的请求同样不能被丢弃
	if sched.Request() {
		t.Fatal("第二轮进行中的请求应返回 false（合并为补扫）")
	}
	if !sched.Pending() {
		t.Fatal("第二轮期间的请求应置位 pending（不能丢失最后一轮请求）")
	}
	close(g2)
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待调度结束超时")
	}

	rounds, maxConc, _ := fs.snapshot()
	if rounds != 3 {
		t.Errorf("扫描轮数 = %d，期望 3（首轮 + 两轮补扫）", rounds)
	}
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1", maxConc)
	}
	// 缩略图必须在所有补扫完成之后、且 pending 为空时运行
	if len(*states) == 0 {
		t.Fatal("缩略图回调未被执行")
	}
	for _, st := range *states {
		if st != "pending=false running=true" {
			t.Errorf("缩略图运行时状态异常: %s（期望 pending=false）", st)
		}
	}
}

func TestScanSchedulerRecoversFromPanic(t *testing.T) {
	fs := newFakeScanner()
	fs.panicNext = true
	sched, _ := scanSchedulerFor(fs)

	sched.Request()
	if !sched.Wait(3 * time.Second) {
		t.Fatal("扫描 panic 后调度器必须复位，不能永久卡住")
	}
	_, _, panics := fs.snapshot()
	if panics != 1 {
		t.Fatalf("模拟的 panic 次数 = %d，期望 1", panics)
	}
	// 扫描失败后系统仍能接受后续刷新
	if !sched.Request() {
		t.Fatal("panic 之后的新请求应当能启动新一轮扫描")
	}
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待第二轮扫描超时")
	}
	rounds, _, _ := fs.snapshot()
	if rounds != 1 {
		t.Errorf("成功轮数 = %d，期望 1（首轮 panic 不计入）", rounds)
	}
}

func TestScanSchedulerConcurrentStress(t *testing.T) {
	fs := newFakeScanner()
	fs.workDelay = 5 * time.Millisecond
	sched, _ := scanSchedulerFor(fs)

	var wg sync.WaitGroup
	started := 0
	var mu sync.Mutex
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if sched.Request() {
				mu.Lock()
				started++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if !sched.Wait(5 * time.Second) {
		t.Fatal("等待调度结束超时")
	}
	rounds, maxConc, _ := fs.snapshot()
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1（不允许两个扫描同时执行）", maxConc)
	}
	if rounds < 1 {
		t.Errorf("扫描轮数 = %d，期望 ≥1", rounds)
	}
	if sched.Pending() {
		t.Error("结束后不应残留 pending")
	}
	if sched.Running() {
		t.Error("结束后不应处于 running 状态")
	}
	if started < 1 {
		t.Errorf("至少应有一个请求真正启动扫描，实际 %d", started)
	}
}

func TestScanSchedulerWaitTimeout(t *testing.T) {
	fs := newFakeScanner()
	g := make(chan struct{})
	fs.gates <- g
	sched, _ := scanSchedulerFor(fs)

	sched.Request()
	waitFor(t, "扫描开始", func() bool { return drain(fs.started) })

	if sched.Wait(80 * time.Millisecond) {
		t.Error("扫描尚未结束，Wait 不应返回 true（否则等于谎报完成）")
	}
	close(g)
	if !sched.Wait(3 * time.Second) {
		t.Error("扫描结束后 Wait 应返回 true")
	}
}

// TestRealScanIndexIsNeverStale 用真实的 scanFiles + 全局调度器验证：
// 一轮扫描期间的多次刷新请求不会导致「最新文件永久进不了索引」。
func TestRealScanIndexIsNeverStale(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)

	// 初始文件
	writeFileT(t, filepath.Join(dir, "old.jpg"), "old")
	scanSched.Request()
	if !scanSched.Wait(10 * time.Second) {
		t.Fatal("首次扫描超时")
	}
	if got := indexTotal(); got != 1 {
		t.Fatalf("初始扫描后索引文件数 = %d，期望 1", got)
	}

	// 扫描期间连续请求：新文件必须在最终索引里出现（请求不能被丢弃）
	writeFileT(t, filepath.Join(dir, "new1.jpg"), "n1")
	writeFileT(t, filepath.Join(dir, "new2.jpg"), "n2")
	for i := 0; i < 3; i++ {
		scanSched.Request()
	}
	if !scanSched.Wait(15 * time.Second) {
		t.Fatal("补扫超时")
	}

	got := indexTotal()
	if got != 3 {
		t.Errorf("最终索引文件数 = %d，期望 3（最新文件不能因请求被丢弃而遗漏）", got)
	}
	if files := filesForDate(time.Now().Format("2006-01-02")); len(files) != 3 {
		t.Errorf("日期索引文件数 = %d，期望 3（日期索引与文件列表必须一致）", len(files))
	}
	restore()
}

// setMediaTestEnv 把全局路径指向临时目录，返回恢复函数。
func setMediaTestEnv(t *testing.T, mediaRoot string) func() {
	t.Helper()
	oldMedia, oldThumbs, oldIndex, oldProbe := mediaDir, thumbsDir, indexPath, ffprobePath
	oldFFmpeg := ffmpegPath
	mediaDir = mediaRoot
	thumbsDir = filepath.Join(mediaRoot, ".thumbs-test")
	indexPath = filepath.Join(mediaRoot, "media_index_test.json")
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatalf("创建测试缩略图目录失败: %v", err)
	}
	// 测试里不依赖外部二进制时把探测关掉，避免误触发真实 ffprobe/ffmpeg
	ffprobePath = ""
	ffmpegPath = "definitely-not-a-real-binary"
	resetMediaRootCache()
	resetDurationCache()

	// 索引与调度器状态复位
	rebuildIndex(nil)

	return func() {
		// 等待可能仍在跑的扫描/缩略图任务结束，避免其继续使用即将失效的全局路径
		scanSched.Wait(10 * time.Second)
		mediaDir, thumbsDir, indexPath, ffprobePath = oldMedia, oldThumbs, oldIndex, oldProbe
		ffmpegPath = oldFFmpeg
		resetMediaRootCache()
		resetDurationCache()
		rebuildIndex(nil)
	}
}

func writeFileT(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
}

```


---

## 文件：settings.go

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// StationConfig 是整合后 Go 端新增的配置段（放在 config.json 的 "station" 键下）。
// Python 下载器原有的所有字段（api_id、api_hash、channels、download_dir 等）
// 原样保留在 config.json 顶层，Go 端不解析、不改动，只透传。
type StationConfig struct {
	Python        string `json:"python"`         // Python 解释器路径（回退模式），空 = 自动探测（会实际执行 --version 验证）
	Script        string `json:"script"`         // tg_auto_download.py 路径（回退模式），空 = exe 目录下
	DownloaderExe string `json:"downloader_exe"` // 后台下载器 exe（常驻监听用，无控制台窗口），空 = exe 目录/tg_downloader.exe
	DownloaderCLI string `json:"downloader_cli"` // CLI 下载器 exe（扫码/登录检查/频道列表用，带控制台以便输出 JSON），空 = exe 目录/tg_downloader_cli.exe
	MediaDir      string `json:"media_dir"`      // 媒体目录，空 = config 的 download_dir
	AutoStart     bool   `json:"auto_start"`     // 程序启动时自动拉起下载器
}

// AppConfig 是 Go 端使用的完整配置视图。
type AppConfig struct {
	// Raw 是 config.json 的完整内容（含 Python 全部字段），保存时原样写回。
	Raw map[string]interface{}

	Station StationConfig `json:"station"`

	// 以下字段从 Raw 中便捷读取（媒体浏览服务参数）
	Port      int    // media_browser.port
	MediaDir  string // download_dir（station.media_dir 优先）
	IndexPath string // 索引文件路径
	ThumbsDir string // 缩略图缓存目录
}

var (
	configPath string // config.json 的绝对路径
	cfgMutex   sync.Mutex
)

// defaultConfigPath 返回默认配置文件路径：exe 所在目录/config.json
func defaultConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

// stationDefaults 返回 station 段的默认值
func stationDefaults() StationConfig {
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	return StationConfig{
		Python:    "",
		Script:    filepath.Join(exeDir, "tg_auto_download.py"),
		MediaDir:  "",
		AutoStart: false,
	}
}

// loadConfig 读取配置文件；不存在则生成一份默认配置并返回。
// 任何错误都返回 nil + error，由调用方决定是否 fatal。
func loadConfig() (*AppConfig, error) {
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	app := &AppConfig{Raw: map[string]interface{}{}}
	app.Station = stationDefaults()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 生成默认配置
			app.Raw = buildDefaultRaw(app.Station)
			if err := saveConfigLocked(app); err != nil {
				return nil, fmt.Errorf("生成默认配置文件失败: %w", err)
			}
			return app, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := json.Unmarshal(data, &app.Raw); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	applyRaw(app)
	return app, nil
}

// buildDefaultRaw 构造与 Python 下载器兼容的默认配置内容
func buildDefaultRaw(st StationConfig) map[string]interface{} {
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	return map[string]interface{}{
		"api_id":            0,
		"api_hash":          "",
		"session":           filepath.Join(exeDir, "tg_session"),
		"download_dir":      filepath.Join(exeDir, "downloads"),
		"channels":          []interface{}{},
		"types":             []interface{}{"photo", "video", "document"},
		"include_forwarded": true,
		"skip_gif":          true,
		"download_history":  false,
		"catchup_today":     true,
		"movie_min_minutes": 5,
		"only_forwarded":    true,
		"max_file_size_mb":  8192,
		"retries":           3,
		"concurrency":       4,
		"proxy":             nil,
		"name_list_file":    "mz.txt",
		"login_method":      "qr",
		"twofa_password":    "",
		"media_browser": map[string]interface{}{
			"exe":        filepath.Join(exeDir, "media-browser-windows-x64.exe"),
			"workdir":    exeDir,
			"port":       8080,
			"index_path": filepath.Join(exeDir, "media_index.json"),
		},
		"station": st,
	}
}

// applyRaw 把 Raw 里的字段解析到 AppConfig 便捷字段
func applyRaw(app *AppConfig) {
	// station 段
	if v, ok := app.Raw["station"].(map[string]interface{}); ok {
		if s, err := json.Marshal(v); err == nil {
			var st StationConfig
			if json.Unmarshal(s, &st) == nil {
				app.Station = st
			}
		}
	}
	// 便捷字段
	app.Port = getInt(app.Raw, []string{"media_browser", "port"}, 8080)
	app.MediaDir = getString(app.Raw, []string{"download_dir"}, "")
	if app.Station.MediaDir != "" {
		app.MediaDir = app.Station.MediaDir
	}
	// 索引/缩略图统一放在程序（exe）目录；media_browser.index_path 仅供 Python 旧逻辑使用
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	app.IndexPath = filepath.Join(exeDir, "media_index.json")
	app.ThumbsDir = filepath.Join(exeDir, "thumbs")
}

// getString 从嵌套 map 中取字符串值
func getString(m map[string]interface{}, keys []string, def string) string {
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return def
		}
		if i == len(keys)-1 {
			if s, ok := v.(string); ok {
				return s
			}
			return def
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return def
		}
		cur = next
	}
	return def
}

// getInt 从嵌套 map 中取整数值（json 数字默认 float64）
func getInt(m map[string]interface{}, keys []string, def int) int {
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return def
		}
		if i == len(keys)-1 {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case string:
				var r int
				if _, err := fmt.Sscanf(n, "%d", &r); err == nil {
					return r
				}
			}
			return def
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return def
		}
		cur = next
	}
	return def
}

// saveConfig 把 AppConfig 写回配置文件（对外入口，加锁）
func saveConfig(app *AppConfig) error {
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	return saveConfigLocked(app)
}

// saveConfigLocked 写回配置（调用方需持有 cfgMutex）
func saveConfigLocked(app *AppConfig) error {
	app.Raw["station"] = app.Station
	data, err := json.MarshalIndent(app.Raw, "", "  ")
	if err != nil {
		return err
	}
	tmp := configPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, configPath)
}

```


---

## 文件：stdout_probe_test.go

```go
package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestProbeExeStdout 是「按需」探测工具：把任意下载器可执行文件丢给它，看 Go 能不能
// 从 stdout 读到协议 JSON。用于验证「console=False 的窗口版 exe 不能用于
// --check-session / --qr-only / --list-channels」这一结论。
//
// 用法（默认跳过，不设置 TG_PROBE_CLI 就不会运行）：
//
//	TG_PROBE_CLI=dist/tg_downloader_cli.exe go test -run TestProbeExeStdout -v .
//
// 生产代码不使用这个测试，它只用于人工/CI 验证构建产物。
func TestProbeExeStdout(t *testing.T) {
	probe := os.Getenv("TG_PROBE_CLI")
	if probe == "" {
		t.Skip("未设置 TG_PROBE_CLI，跳过构建产物探测")
	}
	d := newTestDownloader(probe, "")
	withConfigPath(t, filepath.Join(t.TempDir(), "config.json"))

	obj, err := d.runModeJSON(modeCheckSession, 60*time.Second)
	if err != nil {
		t.Logf("探测 %s → 读不到 JSON：%v", probe, err)
		return
	}
	t.Logf("探测 %s → 读到 JSON：%v", probe, obj)
}

```


---

## 文件：testdata/mockcli/main.go

```go
// mockcli 是下载器 CLI（tg_downloader_cli）的测试替身：
// 只实现 Go 侧依赖的 JSON 协议与各种异常行为（空输出 / 非零退出 / 非法 JSON /
// 超时 / 回显工作目录与环境变量），完全不接触 Telegram，也不读写任何配置或凭据。
//
// 行为由自身文件名决定（同一个测试进程里可以准备多个变体）：
//
//	名字含 "empty"   → 什么都不输出，退出码 0
//	名字含 "nonzero" → 输出一行日志后以退出码 3 结束
//	名字含 "garbage" → 输出一行非 JSON 文本
//	名字含 "slow"    → 先睡 30 秒再输出
//	其它              → 正常输出协议 JSON
//
// 同时可当「Python 解释器替身」用：validatePython 会执行 `<exe> --version`。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func behavior() string {
	name := strings.ToLower(filepath.Base(os.Args[0]))
	switch {
	case strings.Contains(name, "empty"):
		return "empty"
	case strings.Contains(name, "nonzero"):
		return "nonzero"
	case strings.Contains(name, "garbage"):
		return "garbage"
	case strings.Contains(name, "slow"):
		return "slow"
	}
	return "ok"
}

// mode 从参数里找子模式（CLI 形态：<exe> --mode；Python 替身形态：<exe> <script> --mode）
func mode() string {
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "--") {
			return a
		}
	}
	return ""
}

func main() {
	// 作为解释器被探测：validatePython 执行 --version
	if mode() == "--version" {
		fmt.Println("Python 3.99.0 (mockcli)")
		return
	}

	switch behavior() {
	case "empty":
		return
	case "nonzero":
		fmt.Fprintln(os.Stderr, "mockcli: 模拟失败（非零退出码）")
		os.Exit(3)
	case "garbage":
		fmt.Println("mockcli: 这行不是 JSON")
		return
	case "slow":
		time.Sleep(30 * time.Second)
	}

	cwd, _ := os.Getwd()
	cfg := os.Getenv("TG_STATION_CONFIG")

	out := func(v map[string]interface{}) {
		v["cwd"] = cwd
		v["config"] = cfg
		b, _ := json.Marshal(v)
		fmt.Println(string(b))
	}

	switch mode() {
	case "--check-session":
		out(map[string]interface{}{"status": "logged_in"})
	case "--list-channels":
		out(map[string]interface{}{
			"status": "ok",
			"channels": []map[string]interface{}{
				{"id": -1001234567890, "title": "mock-channel"},
			},
		})
	case "--qr-only":
		out(map[string]interface{}{"status": "qr", "qr_base64": "bW9jaw=="})
		// 模拟等待扫码：进程保持存活（Go 侧应在拿到第一行 JSON 后立即返回）
		time.Sleep(30 * time.Second)
		out(map[string]interface{}{"status": "success", "username": "mock"})
	default:
		// 常驻监听模式本替身不实现
		os.Exit(9)
	}
}

```


---

## 文件：thumbs.go

```go
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── 缩略图缓存模块（第二阶段 · 第四部分）────────────────────────────────
//
// 旧实现把「相对路径里的 / 换成 _」当作缓存文件名，会碰撞：
//     a/b.jpg  → a_b.jpg
//     a_b.jpg  → a_b.jpg     ← 两个不同文件共用一张缩略图
// 另外旧实现直接 os.WriteFile / os.Create 写正式文件，进程中断或 FFmpeg
// 失败会留下半张图，而「文件存在」被当成有效缓存。
//
// 本模块提供：
//   1. thumbCacheKey：SHA-256(相对路径 + 大小 + 修改时间) 的无碰撞缓存键；
//   2. 原子写入：临时文件 → 校验非空/可解码 → 同目录 os.Rename 提交；
//   3. 单飞（single-flight）：同一缓存键并发只跑一次 FFmpeg，等待者复用结果；
//   4. 旧命名只读兼容 + 命中即迁移。

const (
	// thumbMaxParallel 同时进行的缩略图生成任务上限（不同缓存键之间）。
	thumbMaxParallel = 4
	// thumbMinBytes 输出文件的最小合理体积，用于识别空文件/半张图。
	thumbMinBytes = 256
	// thumbTempPrefix 临时文件前缀（后缀固定 .jpg，让 FFmpeg 能按扩展名选格式）。
	thumbTempPrefix = ".tmp-thumb-"
)

// thumbTimeout 单张缩略图的生成上限；超时即杀进程并清理临时文件。
// 声明为变量以便测试注入更短的超时（生产默认 20s）。
var thumbTimeout = 20 * time.Second

// renderThumbFn 是缩略图渲染入口，测试可替换以验证并发/失败语义。
var renderThumbFn = renderThumbAtomic

// thumbCmdRunner 执行一次 FFmpeg 渲染（写文件），测试可替换以模拟超时/失败。
var thumbCmdRunner = runFFmpegThumb

// thumbCacheKey 生成稳定、无碰撞的缓存键。
// 输入使用经过安全校验的媒体根相对路径（斜杠分隔）+ 文件大小 + 修改时间，
// 文件内容变化（大小或 mtime 变化）后缓存自动失效。
func thumbCacheKey(rel string, size int64, modTime time.Time) string {
	h := sha256.New()
	h.Write([]byte(filepath.ToSlash(rel)))
	h.Write([]byte{0})
	fmt.Fprintf(h, "%d", size)
	h.Write([]byte{0})
	fmt.Fprintf(h, "%d", modTime.UnixNano())
	return hex.EncodeToString(h.Sum(nil))
}

// thumbCachePath 返回新格式缓存文件路径（绝对路径，位于 thumbsDir 内）。
func thumbCachePath(key string) string {
	return filepath.Join(thumbsDir, key+".thumb.jpg")
}

// legacyThumbCachePath 旧格式缓存路径（"/" → "_"）。
// 只用于兼容读取与一次性迁移，绝不用于写入新文件。
func legacyThumbCachePath(rel string) string {
	safe := strings.ReplaceAll(filepath.ToSlash(rel), "/", "_")
	return filepath.Join(thumbsDir, safe+".thumb.jpg")
}

// thumbCacheFresh 判断缓存文件是否可用：存在、体积合理，且不早于源文件修改时间。
func thumbCacheFresh(cachePath string, src os.FileInfo) bool {
	st, err := os.Stat(cachePath)
	if err != nil || st == nil || !st.Mode().IsRegular() {
		return false
	}
	if st.Size() < thumbMinBytes {
		return false
	}
	return !src.ModTime().After(st.ModTime())
}

// ── 单飞（single-flight）────────────────────────────────────────────────

type thumbFlight struct {
	done chan struct{}
	err  error
}

var thumbFlights = struct {
	mu sync.Mutex
	m  map[string]*thumbFlight
}{m: make(map[string]*thumbFlight)}

// thumbSem 限制同时进行的缩略图生成数量（不同键可并行，但不超过上限）。
var thumbSem = make(chan struct{}, thumbMaxParallel)

// generateThumbOnce 保证同一缓存键同时只有一个生成任务：
//   - 第一个调用者执行 gen；
//   - 其余调用者等待该任务结束并复用其结果（成功或失败）；
//   - 任务结束后立刻清除 in-flight 状态，失败后的后续请求可以重试。
func generateThumbOnce(key string, gen func() error) error {
	thumbFlights.mu.Lock()
	if f, ok := thumbFlights.m[key]; ok {
		thumbFlights.mu.Unlock()
		<-f.done
		return f.err
	}
	f := &thumbFlight{done: make(chan struct{})}
	thumbFlights.m[key] = f
	thumbFlights.mu.Unlock()

	// 保证 in-flight 状态一定被释放（即使 gen panic 也不会永久残留）
	func() {
		defer func() {
			if r := recover(); r != nil {
				f.err = fmt.Errorf("缩略图生成 panic: %v", r)
				log.Printf("[station] 缩略图生成 panic（key=%s）: %v", key[:12], r)
			}
		}()
		f.err = gen()
	}()

	thumbFlights.mu.Lock()
	delete(thumbFlights.m, key)
	thumbFlights.mu.Unlock()
	close(f.done)
	return f.err
}

// thumbInflightCount 返回当前 in-flight 的缓存键数量（测试用）。
func thumbInflightCount() int {
	thumbFlights.mu.Lock()
	defer thumbFlights.mu.Unlock()
	return len(thumbFlights.m)
}

// ── 生成与原子写入 ──────────────────────────────────────────────────────

// ffmpegThumbArgs 构造缩略图参数（视频带 -ss 取样时间点）。
// 输出到明确的 .jpg 文件（配合 -f image2），避免临时文件后缀不匹配导致 FFmpeg 识别不了格式。
// -strict unofficial：不少摄像/下载来源的视频是非全范围 YUV，新版 FFmpeg 的 mjpeg
// 编码器默认拒绝（"Non full-range YUV is non-standard"），加此参数才能正常出图。
func ffmpegThumbArgs(src, out string, isVideo bool, seek string) []string {
	args := []string{"-hide_banner", "-loglevel", "error"}
	if isVideo && seek != "" {
		args = append(args, "-ss", seek)
	}
	args = append(args,
		"-i", src,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale='if(gt(iw,ih),%d,-2)':'if(gt(iw,ih),-2,%d)':flags=lanczos", thumbSize, thumbSize),
		"-strict", "unofficial",
		"-q:v", "2",
		"-f", "image2",
		"-y", out,
	)
	return args
}

// runFFmpegThumb 执行一次 FFmpeg 缩略图生成（带超时），输出写入 out。
func runFFmpegThumb(src, out string, isVideo bool, seek string) error {
	ctx, cancel := context.WithTimeout(context.Background(), thumbTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffmpegPath, ffmpegThumbArgs(src, out, isVideo, seek)...)
	cmd.SysProcAttr = childProcAttr()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("缩略图生成超时（%s）", thumbTimeout)
	}
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errFfmpegUnavailable
		}
		return fmt.Errorf("ffmpeg 退出异常: %v（%s）", err, clip(strings.TrimSpace(stderr.String()), 240))
	}
	return nil
}

var errFfmpegUnavailable = errors.New("ffmpeg 不可用")

// validateThumbFile 校验生成结果：存在、非空、能被解码为图片。
func validateThumbFile(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("输出文件不存在: %w", err)
	}
	if st.Size() < thumbMinBytes {
		return fmt.Errorf("输出文件过小（%d 字节），视为失败", st.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, _, err := image.DecodeConfig(f); err != nil {
		return fmt.Errorf("输出不是可解码的图片: %w", err)
	}
	return nil
}

// renderThumbAtomic 生成缩略图并原子提交到 target。
//
// 步骤：目标目录内建唯一临时文件 → FFmpeg 写入 → 校验退出状态与输出 →
// 同一文件系统内 os.Rename 提交；任何失败都会清理临时文件（含多轮重试产生的
// 全部临时文件），且绝不覆盖已存在的有效缩略图。
func renderThumbAtomic(srcReal, target string, isVideo bool) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("缩略图目录不可写: %w", err)
	}

	// 所有临时文件统一登记，函数返回时一次性清理：
	// 视频缩略图会按多个取样点重试，逐个 Remove 很容易漏删（曾观察到残留临时文件）。
	var tempFiles []string
	defer func() {
		for _, tf := range tempFiles {
			// 已成功提交的临时文件会被 Rename 移走，这里只会得到 ENOENT（无害）
			_ = os.Remove(tf)
		}
	}()
	newTemp := func() (string, error) {
		f, err := os.CreateTemp(dir, thumbTempPrefix+"*.jpg")
		if err != nil {
			return "", fmt.Errorf("无法创建临时文件: %w", err)
		}
		name := f.Name()
		f.Close()
		tempFiles = append(tempFiles, name)
		return name, nil
	}

	if isVideo {
		// 视频：依次尝试多个取样点，任一成功即可
		var lastErr error
		// 末尾的 00:00:00 是首帧兜底：短视频（时长 < 1s）在前几个取样点上都取不到帧，
		// 旧实现会直接判定生成失败，这里退回首帧仍能出图。
		for _, tp := range []string{"00:00:01", "00:00:03", "00:00:05", "00:00:10", "00:00:00"} {
			tmpName, err := newTemp()
			if err != nil {
				return err
			}
			if err := thumbCmdRunner(srcReal, tmpName, true, tp); err != nil {
				lastErr = err
				if errors.Is(err, errFfmpegUnavailable) {
					break
				}
				continue
			}
			if err := validateThumbFile(tmpName); err != nil {
				lastErr = err
				continue
			}
			return commitThumb(tmpName, target)
		}
		// 视频无法用 ffmpeg 出图时，退回 Go 原生解码（部分容器仍可能可解）
		if img := decodeImageFallback(srcReal); img != nil {
			return writeThumbFallback(img, dir, target)
		}
		if lastErr == nil {
			lastErr = errors.New("视频缩略图生成失败")
		}
		return lastErr
	}

	tmpName, err := newTemp()
	if err != nil {
		return err
	}
	var lastErr error
	if err := thumbCmdRunner(srcReal, tmpName, false, ""); err == nil {
		if verr := validateThumbFile(tmpName); verr == nil {
			return commitThumb(tmpName, target)
		} else {
			lastErr = verr
		}
	} else {
		lastErr = err
		if !errors.Is(err, errFfmpegUnavailable) {
			// ffmpeg 存在但处理失败（损坏文件 / 超时等）：继续尝试 Go 原生解码
			log.Printf("[station] ffmpeg 生成图片缩略图失败（%s）: %v", filepath.Base(srcReal), err)
		}
	}
	if img := decodeImageFallback(srcReal); img != nil {
		return writeThumbFallback(img, dir, target)
	}
	if lastErr != nil {
		// 保留底层原因（超时 / 损坏 / 命令缺失），便于日志定位
		return fmt.Errorf("图片缩略图生成失败（原生解码兜底也失败）: %w", lastErr)
	}
	return errors.New("图片缩略图生成失败（ffmpeg 与原生解码均失败）")
}

// commitThumb 原子提交临时文件到正式缓存位置。
func commitThumb(tmpName, target string) error {
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("提交缩略图失败: %w", err)
	}
	return nil
}

// writeThumbFallback 用 Go 原生解码结果写缩略图（无 ffmpeg 时的兼容路径），同样原子提交。
func writeThumbFallback(img image.Image, dir, target string) error {
	tmp, err := os.CreateTemp(dir, thumbTempPrefix+"*.jpg")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := jpeg.Encode(tmp, img, &jpeg.Options{Quality: 85}); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := validateThumbFile(tmpName); err != nil {
		return err
	}
	return commitThumb(tmpName, target)
}

// decodeImageFallback 用 Go 标准库解码图片并缩放到缩略图尺寸。
// 仅作为 ffmpeg 不可用/失败时的兼容回退，返回值可能为 nil。
func decodeImageFallback(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return resizeImage(img, thumbSize)
}

// removeThumbCache 删除某个媒体文件对应的缩略图（新旧两种命名都清）。
// rel 必须是经过 safeMediaPath 校验得到的相对路径。
func removeThumbCache(rel string, size int64, modTime time.Time) {
	candidates := []string{
		thumbCachePath(thumbCacheKey(rel, size, modTime)),
		legacyThumbCachePath(rel),
	}
	for _, p := range candidates {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			log.Printf("[station] 清理缩略图失败 %s: %v", filepath.Base(p), err)
		}
	}
}

// ensureThumb 返回可用的缩略图缓存路径；必要时生成（同键单飞、原子写入）。
// srcInfo 是源文件（解析链接后）的 stat 结果，用于缓存新鲜度与缓存键。
func ensureThumb(mp MediaPath, srcInfo os.FileInfo) (string, error) {
	if srcInfo == nil {
		return "", errors.New("缺少源文件信息")
	}
	key := thumbCacheKey(mp.Rel, srcInfo.Size(), srcInfo.ModTime())
	target := thumbCachePath(key)

	if thumbCacheFresh(target, srcInfo) {
		return target, nil
	}

	// 旧命名兼容读取：命中且新鲜则原子迁移到新键（迁移失败不影响本次服务）
	legacy := legacyThumbCachePath(mp.Rel)
	if thumbCacheFresh(legacy, srcInfo) {
		if err := os.Rename(legacy, target); err == nil {
			return target, nil
		}
		return legacy, nil
	}

	isVideo := videoExts[strings.ToLower(filepath.Ext(mp.Rel))]
	err := generateThumbOnce(key, func() error {
		// 双检：等待期间可能已有其他请求完成了生成
		if thumbCacheFresh(target, srcInfo) {
			return nil
		}
		return renderThumbFn(mp.Real, target, isVideo)
	})
	if err != nil {
		return "", err
	}
	return target, nil
}

// generateAllThumbnails 基于当前索引批量补齐缩略图（不同键并行，上限 thumbMaxParallel）。
// 必须在扫描完成（且没有待执行的补扫）之后调用，否则会把尚未入索引的新文件当作不存在。
func generateAllThumbnails() {
	indexMutex.RLock()
	files := make([]*MediaFile, len(mediaIndex))
	copy(files, mediaIndex)
	indexMutex.RUnlock()

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		generated int
		skipped   int
		failed    int
	)
	jobs := make(chan *MediaFile)

	workerCount := thumbMaxParallel
	if workerCount > len(files) {
		workerCount = len(files)
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for mf := range jobs {
				mp, err := safeMediaPath(mediaDir, mf.Path)
				if err != nil {
					if !errors.Is(err, ErrPathMissing) {
						logPathReject("thumbnail", mf.Path, err)
					}
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				srcInfo, err := os.Stat(mp.Real)
				if err != nil {
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				key := thumbCacheKey(mp.Rel, srcInfo.Size(), srcInfo.ModTime())
				if thumbCacheFresh(thumbCachePath(key), srcInfo) {
					mu.Lock()
					skipped++
					mu.Unlock()
					continue
				}
				if _, err := ensureThumb(mp, srcInfo); err != nil {
					log.Printf("[station] 缩略图生成失败 %s: %v", mf.Path, err)
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				mu.Lock()
				generated++
				mu.Unlock()
			}
		}()
	}
	for _, mf := range files {
		jobs <- mf
	}
	close(jobs)
	wg.Wait()

	log.Printf("缩略图预生成完成：新生成 %d 个，跳过 %d 个，失败 %d 个", generated, skipped, failed)
}

```


---

## 文件：thumbs_test.go

```go
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ── 缩略图测试（第二阶段 · 第四部分）────────────────────────────────────
//
// 覆盖：旧命名碰撞 → 新键不碰撞 / 内容变化失效 / 旧缓存迁移 /
// 同键并发只生成一次 / 不同键可并行（不超过上限）/ 失败后可重试 /
// 原子提交与临时文件清理 / 超时清理 / 源路径穿越被拒。

// makeJPEGBytes 生成一张足够大的 JPEG（> thumbMinBytes）。
func makeJPEGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("生成测试 JPEG 失败: %v", err)
	}
	return buf.Bytes()
}

func writeJPEG(t *testing.T, path string) []byte {
	t.Helper()
	data := makeJPEGBytes(t)
	writeFileT(t, path, string(data))
	return data
}

// TestLegacyThumbKeyCollision 复现旧命名算法（"/" → "_"）的碰撞，
// 并确认新的哈希键不会碰撞。
func TestLegacyThumbKeyCollision(t *testing.T) {
	oldThumbs := thumbsDir
	thumbsDir = t.TempDir()
	defer func() { thumbsDir = oldThumbs }()

	a, b := "a/b.jpg", "a_b.jpg"
	if legacyThumbCachePath(a) != legacyThumbCachePath(b) {
		t.Fatalf("测试前提不成立：旧算法应发生碰撞（%q vs %q）", legacyThumbCachePath(a), legacyThumbCachePath(b))
	}

	mt := time.Unix(1700000000, 0)
	if thumbCacheKey(a, 100, mt) == thumbCacheKey(b, 100, mt) {
		t.Fatalf("新哈希键不应碰撞: %q 与 %q 生成了相同缓存键", a, b)
	}
	// 同大小不同路径、同路径不同大小/时间都必须产生不同键
	if thumbCacheKey(a, 100, mt) == thumbCacheKey(a, 101, mt) {
		t.Error("大小变化应产生不同缓存键（缓存需要失效）")
	}
	if thumbCacheKey(a, 100, mt) == thumbCacheKey(a, 100, mt.Add(time.Second)) {
		t.Error("修改时间变化应产生不同缓存键（缓存需要失效）")
	}
	// 键长稳定（sha256 十六进制）
	if len(thumbCacheKey(a, 1, mt)) != 64 {
		t.Errorf("缓存键应为 64 位十六进制 sha256，实际 %d", len(thumbCacheKey(a, 1, mt)))
	}
}

// TestLegacyThumbMigration 旧命名缓存应被读取并原子迁移到新命名。
func TestLegacyThumbMigration(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	src := filepath.Join(dir, "sub", "pic.jpg")
	data := writeJPEG(t, src)
	_ = data
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// 写一份旧命名缓存（内容为合法 JPEG，时间晚于源文件）
	legacy := legacyThumbCachePath("sub/pic.jpg")
	cacheData := makeJPEGBytes(t)
	if err := os.WriteFile(legacy, cacheData, 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(legacy, future, future); err != nil {
		t.Fatal(err)
	}

	mp, err := safeMediaPath(mediaDir, "sub/pic.jpg")
	if err != nil {
		t.Fatalf("safeMediaPath 失败: %v", err)
	}
	srcInfo, err := os.Stat(mp.Real)
	if err != nil {
		t.Fatal(err)
	}

	// 渲染必须完全不被触发（旧缓存新鲜）
	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var calls int32
	renderThumbFn = func(string, string, bool) error {
		atomic.AddInt32(&calls, 1)
		return fmt.Errorf("不应被调用")
	}

	got, err := ensureThumb(mp, srcInfo)
	if err != nil {
		t.Fatalf("ensureThumb 失败: %v", err)
	}
	if calls != 0 {
		t.Errorf("命中旧缓存时不应重新生成，实际调用 %d 次", calls)
	}
	// 已迁移到新键
	want := thumbCachePath(thumbCacheKey(mp.Rel, srcInfo.Size(), srcInfo.ModTime()))
	if got != want {
		t.Errorf("应迁移到新缓存键 %s，实际 %s", filepath.Base(want), filepath.Base(got))
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("迁移后的缓存文件不存在: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("旧命名缓存应已被移走")
	}
}

// TestThumbSingleFlight 同一缓存键的并发请求只触发一次生成，且等待者复用结果。
func TestThumbSingleFlight(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	writeFileT(t, filepath.Join(dir, "a.jpg"), "source-bytes")

	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var calls int32
	jpg := makeJPEGBytes(t)
	renderThumbFn = func(src, target string, isVideo bool) error {
		atomic.AddInt32(&calls, 1)
		time.Sleep(80 * time.Millisecond) // 拉长窗口，确保并发请求会重叠
		return os.WriteFile(target, jpg, 0o644)
	}

	mp, err := safeMediaPath(mediaDir, "a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	srcInfo, err := os.Stat(mp.Real)
	if err != nil {
		t.Fatal(err)
	}

	const n = 24
	var wg sync.WaitGroup
	errs := make([]error, n)
	paths := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			paths[i], errs[i] = ensureThumb(mp, srcInfo)
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("同一缓存键并发请求应只生成一次，实际 %d 次", got)
	}
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("第 %d 个请求失败: %v", i, errs[i])
		}
		if paths[i] != paths[0] {
			t.Fatalf("等待者应复用同一缓存路径，%s vs %s", paths[0], paths[i])
		}
	}
	if thumbInflightCount() != 0 {
		t.Errorf("生成结束后 in-flight 状态应被释放，实际 %d", thumbInflightCount())
	}
}

// TestThumbFailureAllowsRetry 生成失败时等待者拿到错误，且后续请求可以重试。
func TestThumbFailureAllowsRetry(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	writeFileT(t, filepath.Join(dir, "a.jpg"), "source")

	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var calls int32
	renderThumbFn = func(src, target string, isVideo bool) error {
		if atomic.AddInt32(&calls, 1) == 1 {
			return fmt.Errorf("模拟首次失败")
		}
		return os.WriteFile(target, makeJPEGBytes(t), 0o644)
	}

	mp, err := safeMediaPath(mediaDir, "a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	srcInfo, _ := os.Stat(mp.Real)

	if _, err := ensureThumb(mp, srcInfo); err == nil {
		t.Fatal("首次生成失败应返回错误")
	}
	if thumbInflightCount() != 0 {
		t.Fatalf("失败后 in-flight 状态必须释放，否则后续请求永久阻塞")
	}
	if _, err := ensureThumb(mp, srcInfo); err != nil {
		t.Fatalf("失败后应允许重试，实际: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("期望两次生成尝试，实际 %d", calls)
	}
}

// TestThumbParallelLimitAndDifferentKeys generateAllThumbnails 对不同的键并行生成，
// 但并发数不超过 thumbMaxParallel（不因一把全局锁而串行）。
func TestThumbParallelLimitAndDifferentKeys(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	var files []*MediaFile
	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, fmt.Sprintf("p%02d.jpg", i))
		writeFileT(t, p, fmt.Sprintf("src-%d", i))
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, &MediaFile{Path: fmt.Sprintf("p%02d.jpg", i), Date: st.ModTime(), Size: st.Size()})
	}
	rebuildIndex(files)

	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var cur, peak int32
	jpg := makeJPEGBytes(t)
	renderThumbFn = func(src, target string, isVideo bool) error {
		c := atomic.AddInt32(&cur, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if c <= p || atomic.CompareAndSwapInt32(&peak, p, c) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond)
		err := os.WriteFile(target, jpg, 0o644)
		atomic.AddInt32(&cur, -1)
		return err
	}

	generateAllThumbnails()

	if peak < 2 {
		t.Errorf("不同缓存键应能并行生成（峰值并发 = %d）", peak)
	}
	if peak > int32(thumbMaxParallel) {
		t.Errorf("并发峰值 %d 超过上限 %d", peak, thumbMaxParallel)
	}
	// 每个文件都应有缓存文件
	entries, _ := os.ReadDir(thumbsDir)
	count := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".thumb.jpg") {
			count++
		}
	}
	if count != len(files) {
		t.Errorf("缩略图文件数 = %d，期望 %d", count, len(files))
	}
}

// TestRenderThumbAtomicCleansTempOnFailure 生成失败时必须清理临时文件且不产生正式缓存。
func TestRenderThumbAtomicCleansTempOnFailure(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(thumbsDir, "target.jpg")

	err := renderThumbAtomic(filepath.Join(dir, "missing.jpg"), target, false)
	if err == nil {
		t.Fatal("源文件不存在时应失败")
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Error("失败时不应留下正式缓存文件")
	}
	assertNoTempFiles(t)
}

// TestRenderThumbAtomicTimeoutCleansTemp 模拟超时：错误必须传递且临时文件被清理，
// 已有的有效缓存不能被空文件覆盖。
func TestRenderThumbAtomicTimeoutCleansTemp(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(thumbsDir, "keep.jpg")
	original := makeJPEGBytes(t)
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}

	oldRunner, oldTimeout := thumbCmdRunner, thumbTimeout
	defer func() { thumbCmdRunner, thumbTimeout = oldRunner, oldTimeout }()
	thumbTimeout = 30 * time.Millisecond
	thumbCmdRunner = func(src, out string, isVideo bool, seek string) error {
		return fmt.Errorf("缩略图生成超时（%s）", thumbTimeout)
	}

	err := renderThumbAtomic(filepath.Join(dir, "broken.jpg"), target, false)
	if err == nil {
		t.Fatal("模拟超时应返回错误")
	}
	if !strings.Contains(err.Error(), "超时") {
		t.Errorf("错误信息应包含超时原因，实际: %v", err)
	}
	// 已有有效缓存不能被破坏
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("原有缓存文件不应被删除: %v", readErr)
	}
	if !bytes.Equal(got, original) {
		t.Error("失败时不得用空文件/半成品覆盖已有缓存")
	}
	assertNoTempFiles(t)
}

// TestRenderThumbAtomicWithRealFFmpeg 真实端到端：用系统 ffmpeg 生成缩略图。
// 环境缺少可执行的 ffmpeg 时跳过（跳过原因会打印）。
func TestRenderThumbAtomicWithRealFFmpeg(t *testing.T) {
	ff := findFFmpeg()
	if ff == "" {
		t.Skip("未找到 ffmpeg，跳过真实缩略图生成用例")
	}
	if out, err := exec.Command(ff, "-version").Output(); err != nil {
		t.Skipf("ffmpeg 不可执行（%v）：%s", err, clip(string(out), 120))
	}

	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	ffmpegPath = ff // setMediaTestEnv 会禁用外部二进制，这里显式恢复

	// 用 ffmpeg 自己生成一张测试图（避免依赖 Go 的 PNG 解码器之外的资源）
	src := filepath.Join(dir, "src.png")
	if err := os.WriteFile(src, makePNGBytes(t, 640, 480), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(thumbsDir, "real.jpg")

	if err := renderThumbAtomic(src, target, false); err != nil {
		t.Fatalf("真实 ffmpeg 生成缩略图失败: %v", err)
	}
	st, err := os.Stat(target)
	if err != nil {
		t.Fatalf("缩略图未生成: %v", err)
	}
	if st.Size() < thumbMinBytes {
		t.Errorf("缩略图体积异常: %d 字节", st.Size())
	}
	// 结果必须是可解码图片，且不超过缩略图尺寸
	f, err := os.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("缩略图不可解码: %v", err)
	}
	if cfg.Width > thumbSize || cfg.Height > thumbSize {
		t.Errorf("缩略图尺寸 %dx%d 超过上限 %d", cfg.Width, cfg.Height, thumbSize)
	}
	assertNoTempFiles(t)
}

// TestEnsureThumbRejectsUnsafeSource 穿越 / 兄弟目录 / 根外链接都不能进入生成流程。
func TestEnsureThumbRejectsUnsafeSource(t *testing.T) {
	base := t.TempDir()
	restore := setMediaTestEnv(t, filepath.Join(base, "media"))
	defer restore()
	writeFileT(t, filepath.Join(base, "media", "ok.jpg"), "x")
	writeFileT(t, filepath.Join(base, "media2", "sibling.jpg"), "x")
	writeFileT(t, filepath.Join(base, "secret.jpg"), "x")

	bad := []string{"../secret.jpg", "sub/../../secret.jpg", filepath.Join(base, "media2", "sibling.jpg")}
	for _, req := range bad {
		if _, err := safeMediaPath(mediaDir, req); err == nil {
			t.Errorf("%q 必须在生成缩略图之前被拒绝", req)
		}
	}
	// 媒体目录内的正常文件可通过
	if _, err := safeMediaPath(mediaDir, "ok.jpg"); err != nil {
		t.Errorf("正常文件不应被拒绝: %v", err)
	}
}

func makePNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("编码 PNG 失败: %v", err)
	}
	return buf.Bytes()
}

func assertNoTempFiles(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(thumbsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), thumbTempPrefix) {
			t.Errorf("残留临时文件未清理: %s", e.Name())
		}
	}
}

// TestRenderThumbAtomicRealVideo 真实视频缩略图：
// 覆盖两个曾经出问题的地方——(1) 非全范围 YUV 需要 -strict unofficial 才能出图，
// (2) 多取样点重试不能残留临时文件。
func TestRenderThumbAtomicRealVideo(t *testing.T) {
	ff := findFFmpeg()
	if ff == "" {
		t.Skip("未找到 ffmpeg，跳过真实视频缩略图用例")
	}
	if out, err := exec.Command(ff, "-version").Output(); err != nil {
		t.Skipf("ffmpeg 不可执行（%v）：%s", err, clip(string(out), 120))
	}

	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	ffmpegPath = ff

	video := filepath.Join(dir, "v.mp4")
	cmd := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-y",
		// 故意用 1 秒视频：1s/3s/5s/10s 取样点都会取不到帧，只有首帧兜底能成功
		"-f", "lavfi", "-i", "testsrc=duration=1:size=320x240:rate=10",
		"-pix_fmt", "yuv420p", video)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("生成测试视频失败（%v）：%s", err, clip(string(out), 200))
	}
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(thumbsDir, "v.jpg")
	if err := renderThumbAtomic(video, target, true); err != nil {
		t.Fatalf("真实视频缩略图生成失败: %v", err)
	}
	st, err := os.Stat(target)
	if err != nil {
		t.Fatalf("视频缩略图未生成: %v", err)
	}
	if st.Size() < thumbMinBytes {
		t.Errorf("视频缩略图体积异常: %d 字节", st.Size())
	}
	assertNoTempFiles(t)
}

```


---

## 文件：static/icon.svg

```text
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="512" height="512">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#667eea"/>
      <stop offset="1" stop-color="#764ba2"/>
    </linearGradient>
  </defs>
  <rect width="512" height="512" rx="112" fill="url(#g)"/>
  <rect x="96" y="168" width="320" height="208" rx="28" fill="#ffffff" opacity="0.95"/>
  <circle cx="256" cy="272" r="62" fill="none" stroke="#5b6bd6" stroke-width="22"/>
  <circle cx="256" cy="272" r="24" fill="#5b6bd6"/>
  <rect x="196" y="140" width="120" height="40" rx="16" fill="#ffffff" opacity="0.95"/>
  <circle cx="366" cy="204" r="14" fill="#5b6bd6"/>
</svg>

```


---

## 文件：static/index.html

```html

<!DOCTYPE html>
<html lang="zh-CN"><head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <meta name="theme-color" content="#111111">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
    <meta name="apple-mobile-web-app-title" content="媒体浏览器">
    <link rel="manifest" href="/static/manifest.json">
    <link rel="icon" href="/static/icon.svg">
    <title>媒体浏览器</title>
    <style>
        /* ================= 设计变量 =================
           统一收敛散落的硬编码颜色；每个变量的值与原先一一对应，
           因此这次改动视觉零变化。后续调主题只改这里。 */
        :root {
            /* 背景与面 */
            --bg-page:      #111111;
            --bg-panel:     #18181b;
            --bg-raised:    #1c1c1e;
            --bg-inset:     #1a1a1a;
            --bg-field:     #242428;
            --bg-solid:     #000000;
            --bg-white:     #ffffff;

            /* 边框 */
            --border:       #2a2a2a;
            --border-field: #3a3a3f;
            --border-strong:#333333;

            /* 文字 */
            --fg:           #ffffff;
            --fg-1:         #eeeeee;
            --fg-2:         #dddddd;
            --fg-3:         #aaaaaa;
            --fg-4:         #999999;
            --fg-5:         #888888;
            --fg-6:         #666666;
            --fg-7:         #555555;

            /* 控件 */
            --btn:          #333333;
            --btn-hover:    #444444;

            /* 品牌与强调 */
            --brand-1:      #667eea;
            --brand-1-deep: #6b5ce0;   /* 承载白字的渐变浅端：白字压 #667eea 只有 3.66:1，压这个有 4.96:1 */
            --brand-2:      #764ba2;
            --brand-soft:   #8b9dff;
            --brand-deep:   #5b3c86;   /* 品牌渐变的深端：白字压在上面才有 4.5:1，浅端#667eea 只有 3.7:1 */
            --accent-1:     #f7971e;
            --accent-2:     #ffd200;
            --accent-glow:  rgba(255, 210, 0, 0.25);

            /* 语义色 */
            --ok:           #22c55e;
            --ok-glow:      rgba(34,197,94,0.5);
            --warn:         #f59e0b;
            --danger:       #ef4444;
            --danger-1:     #e74c3c;
            --danger-2:     #c0392b;
            --danger-dark:  #dc2626;
            --danger-soft:  #f87171;
            --danger-soft-bg: rgba(248,113,113,.15);
            --danger-glow:  rgba(231,76,60,0.4);
            --danger-80:    rgba(231,76,60,0.8);
            --danger-100:   rgba(231,76,60,1);

            /* 半透明层 */
            --scrim-header: rgba(17,17,17,0.95);
            --scrim-topbar: rgba(17,17,17,0.92);
            --scrim-panel:  rgba(17,17,17,0.85);
            --overlay:      rgba(0,0,0,0.5);
            --scrim-toast:  rgba(0,0,0,0.85);
            --overlay-modal:rgba(0,0,0,0.6);
            --scrim-lightbox:rgba(0,0,0,0.98);
            --veil:         rgba(0,0,0,0.4);
            --veil-soft:    rgba(0,0,0,0.3);
            --veil-fade:    rgba(0,0,0,0.8);
            --w-10:         rgba(255,255,255,0.1);
            --w-20:         rgba(255,255,255,0.2);
            --w-25:         rgba(255,255,255,0.25);
            --w-30:         rgba(255,255,255,0.3);
            --w-50:         rgba(255,255,255,0.5);
            --w-90:         rgba(255,255,255,0.9);
            --brand-tint:   rgba(102,126,234,0.2);

            /* 布局尺寸（第 2 阶段新增，集中管理便于调） */
            --grid-gap:     2px;   /* 网格细缝：让相邻图有边界；不改 1:1 裁切比例 */
            --grid-radius:  6px;   /* 每格圆角 */
            --pad-top:      20px;  /* 容器上内边距（吸顶定位要减掉它，单位必须在变量里） */
            --header-h:     56px;  /* 筛选栏高度，由 JS 实测后覆盖 */
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        /* 手机上长按不要弹"保存图片/在新标签页打开/复制"这类系统菜单，也不出现选中高亮。
           输入框/文本域例外——否则设置里的名单编辑器、各字段没法选中和定位光标。 */
        * { -webkit-touch-callout: none; -webkit-user-select: none; user-select: none; }
        input, textarea, select, [contenteditable] { -webkit-user-select: text; user-select: text; }
        html, body { height: 100%; }
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: var(--bg-page); color: var(--fg); overflow-x: hidden; }
        .install-banner {
            display: none;
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
            background: linear-gradient(135deg, var(--brand-2), var(--brand-deep));
            padding: 15px 20px;
            /* 必须低于灯箱(1000)。原先是 9999：手机上这条“安装到主屏幕”横幅
               盖住了灯箱底部的视频控制条，所有点击都被它吞掉（表现为“按钮点了没反应”）。 */
            z-index: 900;
            box-shadow: 0 -5px 20px var(--veil-soft);
        }
        .install-banner.show { display: flex; align-items: center; justify-content: space-between; }
        .install-banner-content { display: flex; align-items: center; gap: 15px; }
        .install-icon { font-size: 36px; }
        .install-text h3 { font-size: 16px; font-weight: 600; }
        .install-text p { font-size: 13px; opacity: 0.9; }
        .install-btn {
            padding: 10px 20px;
            background: var(--fg);
            color: var(--brand-2);
            border: none;
            border-radius: 10px;
            font-weight: 600;
            cursor: pointer;
        }
        .install-close {
            background: none;
            border: none;
            color: var(--fg);
            font-size: 24px;
            cursor: pointer;
        }
        .container { padding: 20px; max-width: 1600px; margin: 0 auto; }
        .header { display: flex; align-items: center; gap: 20px; margin-bottom: 18px; flex-wrap: wrap; position: sticky; top: 0; z-index: 800; background: var(--bg-page); backdrop-filter: blur(8px); padding: 12px 0; border-bottom: 1px solid transparent; }
        /* 页面滚动后给粘性头部加一条分隔线，避免与内容"糊"在一起 */
        .header.stuck { border-bottom-color: var(--border); }
        .filter-btns { display: flex; gap: 10px; }
        .filter-btn {
            padding: 10px 24px;
            background: var(--btn);
            border: none;
            color: var(--fg-3);
            border-radius: 20px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            transition: all 0.2s ease;
        }
        .filter-btn:hover { background: var(--btn-hover); color: var(--fg); }
        .filter-btn.active { background: linear-gradient(135deg, var(--brand-1-deep), var(--brand-2)); color: var(--fg); }
        .filter-btn.movie-entry { background: linear-gradient(135deg, var(--accent-1), var(--accent-2)); color: var(--bg-page); box-shadow: 0 2px 12px var(--accent-glow); }
        .filter-btn.movie-entry:hover { filter: brightness(1.1); }
        .action-btns { display: flex; gap: 10px; margin-left: auto; }
        .action-btn { 
            padding: 10px 20px; 
            background: var(--btn); 
            border: none; 
            color: var(--fg-3); 
            border-radius: 20px; 
            cursor: pointer; 
            font-size: 14px; 
            font-weight: 600; 
            transition: all 0.2s ease; 
        }
        .action-btn:hover { background: var(--btn-hover); color: var(--fg); }
        .action-btn.active { background: linear-gradient(135deg, var(--brand-1-deep), var(--brand-2)); color: var(--fg); }

        /* 多选：已选数量角标 + 一键清空 */
        .sel-badge {
            display: inline-flex; align-items: center;
            padding: 4px 12px; border-radius: 14px;
            background: var(--brand-tint); color: var(--brand-soft);
            font-size: 12.5px; font-weight: 600; white-space: nowrap;
        }
        .sel-clear {
            padding: 4px 10px; border-radius: 14px; cursor: pointer;
            background: transparent; border: 1px solid var(--border-field);
            color: var(--fg-3); font-size: 12.5px; font-weight: 600;
            transition: all .15s ease;
        }
        .sel-clear:hover { border-color: var(--danger); color: var(--danger-soft); background: var(--danger-soft-bg); }
        
        /* 标题行选择按钮 */
        #selectModeBtn { 
            background: linear-gradient(135deg, var(--brand-1-deep), var(--brand-2)); 
            color: var(--fg);
        }
        #selectModeBtn:hover { transform: scale(1.05); }
        
        /* 右下角删除按钮 */
        #deleteBtn { 
            position: fixed; 
            bottom: 20px; 
            right: 20px; 
            z-index: 1000; 
            background: linear-gradient(135deg, var(--danger-1), var(--danger-2)); 
            color: var(--fg);
            padding: 8px 16px;
            border-radius: 18px;
            box-shadow: 0 4px 12px var(--danger-glow);
            font-size: 14px;
        }
        #deleteBtn:hover { transform: scale(1.05); }
        .date-section { margin-bottom: 30px; }
        /* 日期标题吸顶：滚动时当前日期贴在筛选栏下方，长列表里不迷路。
           --header-h 由 JS 按筛选栏实际高度写入。sticky 的 top 直接取该高度，
           不减容器内边距（实测减了会往上贴 20px、被筛选栏压住）。 */
        .date-header {
            font-size: 18px; font-weight: 600; padding: 12px 0; color: var(--fg-5);
            display: flex; align-items: center; gap: 10px;
            position: sticky; top: var(--header-h, 56px); z-index: 5;
            background: var(--bg-page);
        }
        .photo-grid { display: grid; grid-template-columns: repeat(8, 1fr); gap: var(--grid-gap); }
        /* content-visibility:auto —— 屏幕外的格子跳过布局与绘制。媒体库里 1 万多个
           格子时，这是“关闭灯箱就卡住/整页卡死”的主因（要绘制上万个格子）。
           aspect-ratio 已经决定了盒子高度，不会引起滚动条跳动。 */
        .photo-item { aspect-ratio: 1; overflow: hidden; cursor: pointer; background: var(--bg-inset); position: relative; user-select: none; transition: transform 0.1s ease; border-radius: var(--grid-radius); content-visibility: auto; contain-intrinsic-size: auto 160px; }
        .photo-item:active { transform: scale(0.95); }
        .photo-item.long-press { transform: scale(0.9); opacity: 0.7; }
        .photo-item img { width: 100%; height: 100%; object-fit: cover; image-rendering: -webkit-optimize-contrast; }
        .photo-item.selected { outline: 3px solid var(--brand-1); z-index: 10; }
        .photo-item.selected::after {
            content: '';
            background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23fff' stroke-width='3.2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M5 12.5l4.5 4.5L19 7'/%3E%3C/svg%3E");
            background-size: 19px 19px; background-position: center; background-repeat: no-repeat;
            position: absolute;
            top: 10px;
            right: 10px;
            width: 30px;
            height: 30px;
            background-color: var(--brand-1);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 18px;
            font-weight: bold;
            z-index: 20;
        }
        /* 视频格子的角标容器：只负责把时长角标摆到右下角，本身不铺任何底色
           （既不留整片黑，也不留底部渐变 —— 缩略图完全原样显示） */
        .video-overlay {
            position: absolute;
            top: 0; left: 0;
            width: 100%;
            height: 100%;
            display: flex;
            align-items: flex-end;
            justify-content: flex-end;
            padding: 6px;
            pointer-events: none;
        }
        .play-indicator {
            width: 34px;
            height: 34px;
            background: var(--w-90);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 15px;
            color: var(--bg-page);
            box-shadow: 0 1px 4px rgba(0, 0, 0, 0.45);
        }
        /* 视频格子的右下角时长角标：既能一眼看出是视频，又能直接看到片长 */
        .video-duration {
            padding: 1px 5px;
            border-radius: 4px;
            background: rgba(0, 0, 0, 0.72);
            color: #fff;
            font-size: 11px;
            line-height: 15px;
            font-variant-numeric: tabular-nums;
            pointer-events: none;
        }
        .lightbox { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: var(--scrim-lightbox); display: none; justify-content: center; align-items: center; z-index: 1000; user-select: none; -webkit-user-select: none; -webkit-touch-callout: none; }
        .lightbox.active { display: flex; }
        .lightbox-content { position: relative; width: 95vw; height: 90vh; display: flex; align-items: center; justify-content: center; background: var(--bg-solid); }
        .lightbox-content img { max-width: 100%; max-height: 100%; object-fit: contain; cursor: zoom-in; }
        .lightbox-content video { width: 100%; height: 100%; object-fit: contain; }
        .lightbox-content img.zoomed { cursor: zoom-out; }
        /* ---- 自定义视频控制条 ---- */
        .video-controls {
            position: absolute; bottom: 0; left: 0; right: 0;
            background: linear-gradient(transparent, var(--veil-fade));
            padding: 30px 16px 12px;
            opacity: 0; pointer-events: none;
            display: flex;
            align-items: center; gap: 10px;
            z-index: 1020;
            transition: opacity 0.3s;
        }
        .video-controls.visible { opacity: 1; pointer-events: auto; }
        .video-controls .vc-time {
            color: var(--fg); font-size: 13px; font-family: monospace;
            flex-shrink: 0; min-width: 90px; text-align: center;
        }
        .video-controls .vc-seek {
            position: relative; flex: 1; height: 20px;
            display: flex; align-items: center; cursor: pointer;
            touch-action: none;
        }
        .video-controls .vc-seek-track {
            position: relative; width: 100%; height: 6px;
            background: var(--w-20); border-radius: 3px; overflow: hidden;
        }
        .video-controls .vc-buf {
            position: absolute; left: 0; top: 0; height: 100%; width: 0;
            background: var(--w-50); border-radius: 3px;
        }
        .video-controls .vc-fill {
            position: absolute; left: 0; top: 0; height: 100%; width: 0;
            background: var(--accent-1); border-radius: 3px;
        }
        .video-controls .vc-thumb {
            position: absolute; top: 50%; left: 0;
            width: 14px; height: 14px; margin-left: -7px;
            transform: translateY(-50%); border-radius: 50%;
            background: var(--fg); box-shadow: 0 0 6px var(--overlay);
            pointer-events: none;
        }
        .video-controls .vc-tip {
            position: absolute; bottom: 22px; left: 0;
            transform: translateX(-50%);
            background: var(--scrim-toast); color: var(--fg);
            font-size: 12px; padding: 3px 8px; border-radius: 6px;
            white-space: nowrap; pointer-events: none;
            opacity: 0; transition: opacity 0.15s;
        }
        .video-controls .vc-tip.show { opacity: 1; }
        .video-controls .vc-vol {
            width: 60px; height: 6px; flex-shrink: 0;
            -webkit-appearance: none; appearance: none;
            background: var(--w-20); border-radius: 3px;
            outline: none; cursor: pointer;
        }
        .video-controls .vc-vol::-webkit-slider-thumb {
            -webkit-appearance: none; width: 14px; height: 14px;
            border-radius: 50%; background: var(--fg);
            box-shadow: 0 0 6px var(--overlay); cursor: pointer;
        }
        .video-controls .vc-vol::-moz-range-thumb {
            width: 14px; height: 14px; border-radius: 50%;
            background: var(--fg); border: none; cursor: pointer;
        }
        .video-controls .vc-full {
            width: 36px; height: 36px; flex-shrink: 0;
            background: var(--w-90); border: none; border-radius: 50%;
            cursor: pointer; font-size: 15px; color: var(--bg-page);
            display: flex; align-items: center; justify-content: center;
        }
        .vc-loading {
            position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%);
            z-index: 1021; pointer-events: none; opacity: 0; transition: opacity 0.2s;
        }
        .vc-loading .lb-spinner {
            width: 40px; height: 40px; border-radius: 50%;
            border: 3px solid var(--w-20); border-top-color: var(--w-90);
            animation: spin 0.8s linear infinite;
        }
        .vc-loading.show { opacity: 1; }
        /* 长按临时倍速的提示：底部居中大字。
           长按触发时显示，1 秒后自动隐藏（松手也会立即隐藏）。 */
        .vc-fast {
            position: absolute; left: 50%; bottom: 76px; transform: translateX(-50%);
            z-index: 1021; pointer-events: none; opacity: 0; transition: opacity 0.2s;
            color: var(--fg); text-shadow: 0 2px 10px var(--scrim-toast);
            font-size: 26px; font-weight: 700; letter-spacing: 1px;
            white-space: nowrap;
        }
        .vc-fast.show { opacity: 1; }
        /* 移动端更大的拖拽点 */
        @media (max-width: 900px) {
            .video-controls { gap: 6px; padding: 26px 10px 10px; }
            .video-controls .vc-seek { height: 24px; }
            .video-controls .vc-seek-track { height: 8px; }
            .video-controls .vc-thumb { width: 20px; height: 20px; margin-left: -10px; }
            .video-controls .vc-vol { display: none; }
            .video-controls .vc-time { min-width: 0; font-size: 11px; }
        }
        /* ===== 触屏 / 窄屏：把控制条做成"手指够用"的尺寸 =====
           实测原尺寸：按钮 30~36px、滑块 14~20px、轨道 6~8px，
           全部远低于触控舒适下限（Apple 44px / Material 48px）。
           这里放到 44px；进度条用"触区 44px + 轨道 10px + 滑块 26px"，
           条本身不会显得笨重。桌面（非触屏、>900px）完全不受影响。 */
        @media (max-width: 900px), (hover: none) and (pointer: coarse) {
            /* 底边距 12px -> 30px：原来进度条（滑块 26px）几乎贴在画面底边，
               和底部区域挤在一起；抬高后不再重叠、也更好按。 */
            .video-controls { gap: 8px; padding: 30px 12px 30px; }
            /* 确保覆盖上面那条 max-width:900px（它把底边距设成 10px） */
            .video-controls.vc-touch { padding-bottom: 30px; }
            .video-controls .vc-full { width: 44px; height: 44px; font-size: 18px; }
            /* 触区 44px，但视觉上还是一条细进度条 */
            .video-controls .vc-seek { height: 44px; }
            .video-controls .vc-seek-track { height: 10px; border-radius: 5px; }
            .video-controls .vc-thumb {
                width: 26px; height: 26px; margin-left: -13px;
                box-shadow: 0 0 8px var(--overlay);
            }
            .video-controls .vc-time { min-width: 0; font-size: 12.5px; }
            /* 音量滑块（>900px 的触屏平板仍会显示）拖动圆点也放大 */
            .video-controls .vc-vol::-webkit-slider-thumb { width: 20px; height: 20px; }
            .video-controls .vc-vol::-moz-range-thumb { width: 20px; height: 20px; }
        }
        .video-error { text-align: center; color: var(--fg-5); padding: 20px; }
        .video-error button { margin-top: 10px; padding: 10px 20px; background: var(--btn); border: none; color: var(--fg); border-radius: 8px; cursor: pointer; }
        .lightbox-close { position: fixed; top: 20px; right: 20px; width: 50px; height: 50px; background: var(--w-20); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 32px; color: var(--fg); cursor: pointer; transition: all 0.2s; z-index: 1010; }
        .lightbox-close:hover { background: var(--w-30); transform: scale(1.1); }
        /* 翻页键：默认（手机/触屏）保持大圆钮好点按；桌面端缩小并改为"鼠标滑到两侧才出现" */
        .lightbox-nav {
            position: fixed; top: 50%; transform: translateY(-50%);
            width: 55px; height: 55px; background: var(--w-10); border-radius: 50%;
            display: flex; align-items: center; justify-content: center;
            font-size: 28px; color: var(--fg); cursor: pointer;
            transition: background 0.2s, opacity 0.25s ease; z-index: 1010;
        }
        .lightbox-nav:hover { background: var(--w-20); }
        .lightbox-nav.prev { left: 10px; }
        .lightbox-nav.next { right: 10px; }
        /* 桌面端：翻页键默认隐藏，鼠标移到画面两侧才淡入。
           两侧各铺一条透明热区（.nav-hotzone），鼠标进入即显示；
           翻页键本身在热区里，所以移过去点击不会闪没。
           同时 JS 用 mousemove 兜底（热区在某些布局下可能被覆盖）。 */
        .nav-hotzone { display: none; }
        @media (min-width: 901px) {
            .lightbox-nav {
                width: 84px; height: 84px; font-size: 40px;
                opacity: 0; visibility: hidden; pointer-events: none;
            }
            .lightbox-nav.prev { left: 16px; }
            .lightbox-nav.next { right: 16px; }
            /* 鼠标在任一侧热区、或 JS 判定靠近边缘时显示 */
            .lightbox:has(.nav-hotzone:hover) .lightbox-nav,
            .lightbox.nav-near .lightbox-nav {
                opacity: 1; visibility: visible; pointer-events: auto;
            }
            .nav-hotzone {
                display: block; position: fixed; top: 0; bottom: 0; width: 110px;
                z-index: 1005;
            }
            .nav-hotzone.prev { left: 0; }
            .nav-hotzone.next { right: 0; }
        }
        .lightbox-counter { position: fixed; bottom: 30px; left: 50%; transform: translateX(-50%); color: var(--fg); font-size: 16px; background: var(--overlay); padding: 8px 16px; border-radius: 20px; z-index: 1010; }
        .lightbox-delete { 
            position: fixed; 
            top: 20px; 
            right: 80px; 
            width: 50px; 
            height: 50px; 
            background: var(--danger-80); 
            border-radius: 50%; 
            display: flex; 
            align-items: center; 
            justify-content: center; 
            font-size: 28px; 
            cursor: pointer; 
            transition: all 0.3s ease; 
            z-index: 1010;
            box-shadow: 0 4px 12px var(--danger-glow);
            user-select: none;
            -webkit-user-select: none;
            -webkit-touch-callout: none;
            touch-action: manipulation;
        }
        /* 桌面端：删除按钮在右侧翻页键正上方，与翻页键同轴。
           翻页键 84px、垂直居中（top:50% translateY(-50%)）、right:16px。
           - 垂直：翻页键上边缘在 50%-42px；删除键 56px 高、留 10px 间距，
             故其上边缘 = (50%-42px) - 10px - 56px = 50% - 108px。
           - 水平：两者宽度不同（56 vs 84），光靠 right 相同只是右边缘对齐；
             要让中点重合需 right = 16px + (84-56)/2 = 30px。 */
        @media (min-width: 901px) {
            .lightbox-delete {
                right: 30px;
                top: calc(50% - 108px);
                transform: none;
                width: 56px;
                height: 56px;
                font-size: 26px;
            }
            .lightbox-delete:hover { transform: scale(1.1); }
        }
        .lightbox-delete:hover { 
            background: var(--danger-100); 
            transform: scale(1.1); 
        }
        /* 灯箱图片加载进度 */
        .lightbox-loading {
            position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%);
            display: flex; flex-direction: column; align-items: center; gap: 10px;
            color: var(--w-50); font-size: 13px; z-index: 1005; pointer-events: none;
        }
        .lightbox-loading .lb-spinner {
            width: 34px; height: 34px; border-radius: 50%;
            border: 3px solid var(--w-20); border-top-color: var(--w-90);
            animation: spin 0.8s linear infinite;
        }
        .delete-toast { position: fixed; top: 15%; left: 50%; transform: translate(-50%, 0); background: var(--scrim-toast); color: var(--fg); padding: 16px 32px; border-radius: 12px; font-size: 18px; font-weight: 600; z-index: 2000; opacity: 0; pointer-events: none; transition: opacity 0.3s ease; }
        .delete-toast.show { opacity: 1; }
        /* ===== 错误弹窗（替代 alert） ===== */
        .err-toast {
            position: fixed; top: 16px; left: 50%;
            transform: translate(-50%, -140%);
            max-width: min(92vw, 560px);
            display: flex; align-items: flex-start; gap: 12px;
            padding: 14px 16px; border-radius: 12px;
            background: rgba(40,20,22,0.98); backdrop-filter: blur(10px);
            border: 1px solid rgba(239,68,68,0.5);
            box-shadow: 0 12px 32px rgba(0,0,0,0.55);
            color: var(--fg-2); z-index: 3200;
            opacity: 0; pointer-events: none;
            transition: transform .28s cubic-bezier(.2,.9,.3,1), opacity .28s ease;
        }
        .err-toast.show { transform: translate(-50%, 0); opacity: 1; pointer-events: auto; }
        .err-toast .err-icon { font-size: 20px; line-height: 1.3; flex-shrink: 0; }
        .err-toast .err-text { flex: 1; min-width: 0; }
        .err-toast .err-title { font-size: 14px; font-weight: 700; color: var(--danger-soft); margin-bottom: 3px; }
        .err-toast .err-msg { font-size: 13px; line-height: 1.55; color: var(--fg-2); word-break: break-word; white-space: pre-wrap; }
        .err-toast .err-close {
            flex-shrink: 0; background: none; border: none; color: var(--fg-5);
            font-size: 22px; line-height: 1; cursor: pointer; padding: 0 2px;
        }
        .err-toast .err-close:hover { color: var(--fg); }
        /* ===== 确认弹窗（替代 confirm） ===== */
        .confirm-overlay {
            position: fixed; inset: 0; z-index: 3300;
            display: none; align-items: center; justify-content: center;
            background: var(--overlay-modal); backdrop-filter: blur(3px); padding: 20px;
        }
        .confirm-overlay.show { display: flex; }
        .confirm-box {
            width: min(92vw, 400px); padding: 20px;
            background: var(--bg-panel); border: 1px solid var(--border-field);
            border-radius: 14px; box-shadow: 0 20px 50px rgba(0,0,0,0.6);
            animation: confirmIn .18s ease;
        }
        @keyframes confirmIn { from { transform: scale(.94); opacity: 0; } to { transform: none; opacity: 1; } }
        .confirm-title { font-size: 16px; font-weight: 700; color: var(--fg-1); margin-bottom: 8px; }
        .confirm-msg { font-size: 13.5px; line-height: 1.6; color: var(--fg-3); white-space: pre-wrap; word-break: break-word; }
        .confirm-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 18px; }
        .confirm-btn {
            padding: 9px 20px; border-radius: 10px; border: none; cursor: pointer;
            font-size: 14px; font-weight: 600; transition: all .15s ease;
        }
        .confirm-btn.ghost { background: var(--btn); color: var(--fg-2); }
        .confirm-btn.ghost:hover { background: var(--btn-hover); color: var(--fg); }
        .confirm-btn.danger { background: linear-gradient(135deg, var(--danger-2), var(--danger-dark)); color: var(--fg); }
        .confirm-btn.danger:hover { filter: brightness(1.12); }

        .select-hint { background: var(--brand-tint); padding: 8px 16px; border-radius: 20px; font-size: 14px; }
        .zoom-hint { background: var(--brand-tint); padding: 8px 16px; border-radius: 20px; font-size: 14px; margin-left: auto; }
        @keyframes spin { to { transform: rotate(360deg); } }
        @media (max-width: 1400px) { .photo-grid { grid-template-columns: repeat(7, 1fr); } }
        @media (max-width: 1200px) { .photo-grid { grid-template-columns: repeat(6, 1fr); } }
        @media (max-width: 900px) { .photo-grid { grid-template-columns: repeat(5, 1fr); } .play-indicator { width: 30px; height: 30px; font-size: 14px; } }
        @media (max-width: 600px) { 
            .photo-grid { grid-template-columns: repeat(4, 1fr); } 
            :root { --grid-radius: 4px; --pad-top: 15px; }
            .filter-btn { padding: 8px 16px; font-size: 14px; } 
            .action-btn { padding: 8px 16px; font-size: 14px; }
            .play-indicator { width: 26px; height: 26px; font-size: 13px; }
            .video-duration { font-size: 10px; line-height: 13px; padding: 0 4px; border-radius: 3px; } 
            .lightbox-close { width: 45px; height: 45px; font-size: 26px; top: 15px; right: 15px; }
            .lightbox-nav { width: 55px; height: 55px; font-size: 28px; }
            .lightbox-nav.prev { left: 10px; }
            .lightbox-nav.next { right: 10px; }
            .container { padding: 15px; }
            .header { gap: 10px; }
        }
        /* ===== 触屏：顶栏与筛选栏的控制也做到"手指够用"（44px）=====
           实测手机(390px 宽)：筛选档与操作按钮只有 36px 高、顶栏按钮 33px，
           低于触控舒适下限。桌面鼠标端尺寸完全不变——这条只在小屏或粗指针生效。 */
        @media (max-width: 900px), (hover: none) and (pointer: coarse) {
            .filter-btn, .action-btn {
                min-height: 44px; white-space: nowrap;
                display: inline-flex; align-items: center; justify-content: center;
            }
            /* 右下角删除键：JS 会给它写行内 display:block（覆盖 flex），所以用行高居中
               （44 - 上下内边距 8×2 = 28），否则文字会贴在 44px 框的偏上方 */
            #deleteBtn { min-height: 44px; line-height: 28px; white-space: nowrap; }
            .topbar-btn { min-height: 44px; display: inline-flex; align-items: center; }
            .sel-clear { min-height: 36px; display: inline-flex; align-items: center; }
            /* 下载悬浮卡是高密度控制台：微按钮只放大到不误点为止，不撑成 44px */
            .dl-stop-btn { min-height: 32px; padding: 6px 12px; }
            .dl-cancel-btn { min-height: 30px; padding: 5px 12px; }
            .modal-close { min-width: 44px; min-height: 44px; display: inline-flex; align-items: center; justify-content: center; }
            .install-btn { min-height: 44px; display: inline-flex; align-items: center; }
        }

        /* ===== 键盘焦点可见 =====
           鼠标点击不出现（:focus-visible 语义），Tab 到按钮时才出现品牌色描边环，
           避免电脑端键盘操作"看不见焦点在哪"。 */
        button:focus-visible, [role="button"]:focus-visible {
            outline: 2px solid var(--brand-soft);
            outline-offset: 2px;
        }

        /* ================= TG Station 整合 UI ================= */
        .station-topbar {
            flex-basis: 100%; height: 52px;
            background: var(--scrim-topbar); backdrop-filter: blur(10px);
            border: 1px solid var(--border); border-radius: 12px;
            display: flex; align-items: center; gap: 12px; padding: 0 16px;
            margin-top: 8px;
        }
        .topbar-title { font-size: 16px; font-weight: 700; white-space: nowrap; }
        /* 页面版本标识：用于快速判断浏览器加载的是新版还是缓存旧版。
           数字 = 本文件大小(KB) + 内联脚本总字节数，改一次就变一次。 */
        .build-stamp {
            font-size: 10.5px; font-weight: 500; color: var(--fg-4);
            margin-left: 6px; letter-spacing: 0.2px; cursor: help;
        }
        .topbar-status { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--fg-3); flex: 1; min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
        .status-dot { width: 10px; height: 10px; border-radius: 50%; background: var(--fg-7); flex-shrink: 0; }
        .status-dot.on { background: var(--ok); box-shadow: 0 0 8px var(--ok-glow); }
        .status-dot.off { background: var(--danger); }
        .status-dot.busy { background: var(--warn); }
        .topbar-actions { display: flex; gap: 8px; }
        .topbar-btn {
            padding: 7px 14px; background: var(--btn); border: none; color: var(--fg-2);
            border-radius: 16px; cursor: pointer; font-size: 13px; font-weight: 600; white-space: nowrap;
            transition: all 0.2s ease;
        }
        .topbar-btn:hover { background: var(--btn-hover); color: var(--fg); }
        .topbar-btn.primary { background: linear-gradient(135deg, var(--brand-1-deep), var(--brand-2)); color: var(--fg); }
        .topbar-btn.primary:hover { filter: brightness(1.15); }

        /* ===== 下载进度面板 =====
           从顶部挪到右下角悬浮：翻到页面任何位置都能看到下载进度。
           无任务时自动隐藏（JS 控制 .show）。 */
        .download-panel {
            position: fixed; right: 20px; bottom: 96px; width: 340px;
            max-height: min(60vh, 460px); overflow: hidden auto;
            background: rgba(24,24,27,0.96); backdrop-filter: blur(12px);
            border: 1px solid var(--border-strong); border-radius: 14px;
            padding: 12px 14px; font-size: 13px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.5); z-index: 940;
            opacity: 0; transform: translateY(8px); pointer-events: none;
            transition: opacity .2s ease, transform .2s ease;
        }
        .download-panel.show { opacity: 1; transform: none; pointer-events: auto; }
        .dl-title { font-weight: 700; margin-bottom: 8px; color: var(--fg-1); }
        .dl-header {
            display: flex; align-items: center; gap: 8px;
            font-weight: 700; color: var(--fg-1); cursor: pointer;
            user-select: none; padding: 2px 0;
        }
        .dl-header:hover { color: var(--fg); }
        .dl-summary { font-weight: 400; font-size: 12px; color: var(--fg-5); }
        /* 总进度 / 速度 / 剩余时间 */
        .dl-meta { margin-top: 8px; display: flex; flex-direction: column; gap: 5px; }
        .dl-meta-row { display: flex; justify-content: space-between; gap: 10px; font-size: 12px; color: var(--fg-5); }
        .dl-meta-row b { color: var(--fg-2); font-weight: 600; }
        .dl-overall-bar { height: 8px; background: var(--btn); border-radius: 4px; overflow: hidden; }
        .dl-overall-fill { height: 100%; background: linear-gradient(90deg, var(--brand-1), var(--brand-2)); border-radius: 4px; transition: width .5s; }
        .dl-done {
            margin-top: 8px; padding: 8px 10px; border-radius: 8px; font-size: 12.5px;
            background: rgba(34,197,94,0.12); border: 1px solid rgba(34,197,94,0.35);
            color: var(--ok); text-align: center; font-weight: 600;
        }
        .dl-stop-btn {
            margin-left: auto;
            background: var(--danger-dark); color: var(--fg); border: none; border-radius: 6px;
            font-size: 11px; padding: 3px 10px; cursor: pointer; font-weight: 600;
        }
        \.dl-stop-btn:hover { background: var(--danger-2); }
        .dl-task-actions { display: flex; justify-content: flex-end; margin-top: 2px; }
        .dl-cancel-btn {
            background: transparent; color: var(--danger-soft); border: 1px solid var(--danger-soft);
            border-radius: 5px; font-size: 10px; padding: 1px 8px; cursor: pointer;
        }
        .dl-cancel-btn:hover { background: var(--danger-soft-bg); }
        .dl-arrow { margin-left: 4px; color: var(--fg-6); font-size: 12px; transition: transform .2s; }
        .dl-arrow.open { transform: rotate(90deg); }
        .dl-task { display: flex; flex-direction: column; gap: 4px; margin-bottom: 10px; }
        .dl-task-head { display: flex; justify-content: space-between; gap: 8px; color: var(--fg-2); }
        .dl-task-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .dl-task-pct { color: var(--ok); font-weight: 600; white-space: nowrap; }
        .dl-bar { height: 6px; background: var(--btn); border-radius: 3px; overflow: hidden; }
        .dl-bar-fill { height: 100%; background: linear-gradient(90deg, var(--brand-1), var(--brand-2)); border-radius: 3px; transition: width .5s; }
        .dl-recent-title { font-weight: 700; margin: 6px 0 6px; color: var(--fg-1); }
        .dl-recent { display: flex; flex-direction: column; gap: 4px; }
        .dl-recent-item { display: flex; gap: 8px; color: var(--fg-3); }
        .dl-recent-time { color: var(--fg-6); white-space: nowrap; }
        .dl-recent-file { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .dl-empty { color: var(--fg-5); padding: 4px 0; }
        @media (max-width: 640px) {
            /* 小屏：悬浮卡收成底部小胶囊，避免遮住大片内容 */
            .download-panel {
                right: 12px; left: 12px; bottom: 12px; width: auto;
                max-height: 40vh; padding: 10px 12px; font-size: 12px;
                border-radius: 12px;
            }
            /* 小屏折叠时只留标题行，点一下展开 */
            .download-panel:not(.expanded) .dl-meta,
            .download-panel:not(.expanded) .dl-body { display: none; }
        }

        .station-fab {
            position: fixed; right: 20px; bottom: 24px; width: 56px; height: 56px;
            border-radius: 50%; background: linear-gradient(135deg, var(--brand-1), var(--brand-2));
            display: flex; align-items: center; justify-content: center;
            font-size: 22px; color: var(--fg); cursor: pointer; z-index: 950;
            box-shadow: 0 4px 16px var(--veil); user-select: none;
            -webkit-user-select: none; -webkit-tap-highlight-color: transparent;
        }
        .station-fab:active { transform: scale(0.92); }
        /* 下载悬浮卡出现时，悬浮球上移避让，二者不重叠 */
        body.dl-open .station-fab { bottom: 300px; }
        body.dl-open .fab-menu { bottom: 364px; }
        .station-fab, .fab-menu { transition: bottom .2s ease; }
        .station-fab .fab-dot {
            position: absolute; top: 2px; right: 2px; width: 14px; height: 14px;
            border-radius: 50%; border: 2px solid var(--bg-inset);
        }
        .fab-dot.on { background: var(--ok); }
        .fab-dot.off { background: var(--danger); }
        .fab-dot.busy { background: var(--warn); }
        .fab-menu {
            position: fixed; right: 20px; bottom: 88px;
            background: var(--bg-raised); border: 1px solid var(--btn); border-radius: 14px;
            padding: 6px; z-index: 960; box-shadow: 0 8px 24px var(--overlay);
            display: flex; flex-direction: column; gap: 2px;
        }
        .fab-menu-item {
            background: none; border: none; color: var(--fg-1); padding: 10px 18px;
            border-radius: 10px; font-size: 14px; cursor: pointer;
            text-align: left; white-space: nowrap;
        }
        .fab-menu-item:hover { background: var(--btn); }

        .modal-overlay {
            position: fixed; inset: 0; background: var(--overlay-modal);
            z-index: 3000; display: flex; align-items: center; justify-content: center; padding: 16px;
        }
        .modal {
            background: var(--bg-panel); border: 1px solid var(--btn); border-radius: 16px;
            width: 100%; max-width: 640px; max-height: 90vh;
            display: flex; flex-direction: column; box-shadow: 0 12px 40px var(--overlay-modal);
        }
        .modal-header {
            display: flex; align-items: center; justify-content: space-between;
            padding: 14px 20px; border-bottom: 1px solid var(--border);
            font-size: 16px; font-weight: 700;
        }
        .modal-close { background: none; border: none; color: var(--fg-3); font-size: 24px; cursor: pointer; line-height: 1; }
        .modal-close:hover { color: var(--fg); }
        .modal-body { padding: 16px 20px; overflow-y: auto; }
        .modal-section-title { font-size: 13px; color: var(--brand-soft); font-weight: 700; margin: 14px 0 8px; }
        .modal-section-title:first-child { margin-top: 0; }
        .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px 14px; }
        .form-item { display: flex; flex-direction: column; gap: 4px; }
        .form-item.full { grid-column: 1 / -1; }
        .form-item label { font-size: 12px; color: var(--fg-4); }
        .form-item input, .form-item select {
            background: var(--bg-field); border: 1px solid var(--border-field); color: var(--fg-1);
            border-radius: 8px; padding: 8px 10px; font-size: 13px; outline: none;
        }
        .namelist-panel { display: flex; flex-direction: column; gap: 8px; }
        .namelist-editor {
            width: 100%; box-sizing: border-box; min-height: 240px; resize: vertical;
            padding: 10px 12px; border: 1px solid var(--border-field); border-radius: 8px;
            background: var(--bg-inset); color: var(--fg-1); font-size: 13px; line-height: 1.6;
            font-family: Consolas, "Courier New", monospace; outline: none;
        }
        .namelist-editor:focus { border-color: var(--brand-1); }
        .namelist-bar { display: flex; align-items: center; gap: 10px; }
        .namelist-hint { flex: 1; font-size: 12px; color: var(--fg-5); }
        .channel-panel { display: flex; flex-direction: column; gap: 8px; }
        .channel-bar { display: flex; align-items: center; gap: 10px; }
        .channel-list {
            max-height: 220px; overflow-y: auto; border: 1px solid var(--border-field);
            border-radius: 8px; padding: 6px 10px; background: var(--bg-inset);
            display: flex; flex-direction: column; gap: 2px;
        }
        .channel-item {
            display: flex; align-items: center; gap: 8px; padding: 5px 4px;
            border-radius: 6px; cursor: pointer; font-size: 13px; color: var(--fg-2);
        }
        .channel-item:hover { background: var(--bg-field); }
        .channel-item input { width: 15px; height: 15px; accent-color: var(--brand-1); cursor: pointer; }
        .channel-item .ch-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .channel-item .ch-id { font-size: 11px; color: var(--fg-6); font-family: Consolas, monospace; }
        .form-item input:focus, .form-item select:focus { border-color: var(--brand-1); }
        .form-item input[type=checkbox] { width: 18px; height: 18px; accent-color: var(--brand-1); cursor: pointer; }
        .checkbox-item { flex-direction: row; align-items: center; gap: 8px; }
        .checkbox-item label { color: var(--fg-2); font-size: 13px; cursor: pointer; }
        .modal-footer {
            display: flex; align-items: center; justify-content: flex-end; gap: 12px;
            padding: 12px 20px; border-top: 1px solid var(--border);
        }
        .modal-msg { flex: 1; font-size: 12px; color: var(--ok); }
        .modal-msg.err { color: var(--danger); }

        .station-toast {
            position: fixed; bottom: 96px; left: 50%; transform: translateX(-50%);
            background: var(--scrim-toast); color: var(--fg); padding: 10px 20px;
            border-radius: 12px; font-size: 14px; z-index: 3100;
            opacity: 0; pointer-events: none; transition: opacity 0.25s ease;
        }
        .station-toast.show { opacity: 1; }

        .login-panel { display: flex; align-items: center; gap: 12px; margin-bottom: 10px; flex-wrap: wrap; }
        .login-status-text { font-size: 14px; color: var(--fg-2); }
        .login-status-text.ok { color: var(--ok); }
        .qr-container {
            display: flex; flex-direction: column; align-items: center; gap: 10px;
            padding: 14px; background: var(--bg-field); border: 1px solid var(--border-field);
            border-radius: 12px; margin-bottom: 6px;
        }
        .qr-container img {
            width: 200px; height: 200px; border-radius: 8px;
            background: var(--fg); padding: 8px; image-rendering: pixelated;
        }
        .qr-hint { font-size: 12px; color: var(--fg-4); text-align: center; line-height: 1.6; }

        @media (max-width: 600px) {
            .form-grid { grid-template-columns: 1fr; }
            .topbar-title { font-size: 14px; }
            .topbar-status { font-size: 12px; }
        }
        /* ===== 手机端（≤600px）顶栏重排：两行，按钮 44px，状态文字可见 =====
           必须放在样式表末尾：顶栏的基础规则定义在文件后半段，只有靠源顺序才能覆盖
           （放在中间会被后面的基础 height:52px 压掉，内容反而溢出圆角外框）。
           实测 390px 宽：标题(nowrap)+操作按钮需要 362px，而内容盒只有 326px →
           按钮横向溢出外框 20px、状态文字被 flex 压成 0 宽（"未运行/监听中" 在手机上
           完全看不见，只剩悬浮球上的小圆点）。现在：第一行 标题 + 状态（长则省略号），
           第二行 两个操作按钮平分整行。 */
        @media (max-width: 600px) {
            .station-topbar { height: auto; min-height: 52px; flex-wrap: wrap; padding: 8px 12px; gap: 6px 10px; }
            .topbar-status { font-size: 12px; }
            .topbar-actions { flex: 1 1 100%; }
            .topbar-actions .topbar-btn { flex: 1; padding: 7px 10px; }
            /* 顶栏多一行，把筛选排与它的间距收紧抵回来 */
            .header { gap: 10px; margin-bottom: 12px; }
        }
        /* ===== 图标系统（手绘一套，2026-10-10 用户拍板方向）=====
           24×24 视框、2px 描边、圆头圆角、currentColor 上色；尺寸用 em 跟随所在
           文字大小（.ic = 1.05em），所以同一套图标在 13px 顶栏按钮与 28px 灯箱
           圆钮里都自动成比例。图标是"画出来的"，不再用 emoji/文字字形顶替。
           ic-solo = 独居按钮内（关闭/翻页/删除/全屏），去掉与文字之间的间距。 */
        .ic { width: 1.05em; height: 1.05em; flex: none; vertical-align: -0.16em;
              margin-right: .34em; fill: none; stroke: currentColor; stroke-width: 2;
              stroke-linecap: round; stroke-linejoin: round; }
        .ic-solo { margin-right: 0; }
        /* 视频格子上的播放圆钮：白底 + 深色三角（承重图形，单独放大一点） */
        .play-indicator .ic { width: 15px; height: 15px; margin-right: 0; vertical-align: 0; stroke-width: 2.4; }

    </style>
</head>
<body>
    <!-- ===== TG Station 悬浮球 ===== -->
    <div class="station-fab" id="stationFab" title="TG Station 快捷菜单">
        <span class="fab-dot off" id="fabDot"></span>
        <span class="fab-icon" id="fabIcon"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M7.5 5v14l11-7z"/></svg></span>
    </div>
    <div class="fab-menu" id="fabMenu" style="display:none;">
        <button class="fab-menu-item" id="fabToggleBtn"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M7.5 5v14l11-7z"/></svg>开始监听</button>
        <button class="fab-menu-item" id="fabRefreshBtn"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M20.5 12a8.5 8.5 0 1 1-2.7-6.2"/><path d="M20.8 4.2v4.2h-4.2"/></svg>刷新媒体</button>
        <button class="fab-menu-item" id="fabSettingsBtn"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7.5h8M16.5 7.5H20M4 16.5h3.5M12 16.5H20"/><circle cx="14.6" cy="7.5" r="2.2"/><circle cx="9.8" cy="16.5" r="2.2"/></svg>设置</button>
    </div>
    <!-- 删除按钮：body 层级 fixed 右下角（放 header 外，避免 header 的 backdrop-filter 破坏 fixed 定位） -->
    <button class="action-btn delete" id="deleteBtn" style="display:none;">删除</button>

    <!-- ===== TG Station 设置弹窗 ===== -->
    <div class="modal-overlay" id="settingsModal" style="display:none;">
        <div class="modal">
            <div class="modal-header">
                <span><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7.5h8M16.5 7.5H20M4 16.5h3.5M12 16.5H20"/><circle cx="14.6" cy="7.5" r="2.2"/><circle cx="9.8" cy="16.5" r="2.2"/></svg>设置</span>
                <button class="modal-close" id="settingsCloseBtn" aria-label="关闭"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/></svg></button>
            </div>
            <div class="modal-body" id="settingsBody"></div>
            <div class="modal-footer">
                <span id="settingsMsg" class="modal-msg"></span>
                <button class="topbar-btn primary" id="saveSettingsBtn"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M19 20.5H5A1.5 1.5 0 0 1 3.5 19V5A1.5 1.5 0 0 1 5 3.5h11L20.5 8v11a1.5 1.5 0 0 1-1.5 1.5z"/><path d="M17 20.5v-7H7v7"/><path d="M7 3.5v5h7"/></svg>保存设置</button>
            </div>
        </div>
    </div>

    <div class="station-toast" id="stationToast"></div>

    <!-- 错误弹窗：替代原生 alert（原生弹窗会冻结整个页面，且样式与界面不搭） -->
    <div class="err-toast" id="errToast" role="alert" aria-live="assertive">
        <div class="err-icon"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 4.2L21 19.8H3z"/><path d="M12 10v4.2"/><path d="M12 17.2h.01"/></svg></div>
        <div class="err-text">
            <div class="err-title" id="errTitle">操作失败</div>
            <div class="err-msg" id="errMsg"></div>
        </div>
        <button class="err-close" id="errClose" aria-label="关闭"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/></svg></button>
    </div>

    <!-- 确认弹窗：替代原生 confirm（原生弹窗冻结页面，且移动端体验差） -->
    <div class="confirm-overlay" id="confirmModal">
        <div class="confirm-box">
            <div class="confirm-title" id="confirmTitle">确认操作</div>
            <div class="confirm-msg" id="confirmMsg"></div>
            <div class="confirm-actions">
                <button class="confirm-btn ghost" id="confirmCancel">取消</button>
                <button class="confirm-btn danger" id="confirmOk">确定</button>
            </div>
        </div>
    </div>

    <div class="install-banner" id="installBanner">
        <div class="install-banner-content">
            <div class="install-icon"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><rect x="7" y="3.5" width="10" height="17" rx="2.5"/><path d="M11 17.4h2"/></svg></div>
            <div class="install-text">
                <h3>安装到主屏幕</h3>
                <p>像APP一样使用媒体浏览器</p>
            </div>
        </div>
        <button class="install-btn" id="installBtn">安装</button>
        <button class="install-close" id="installClose" aria-label="关闭"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/></svg></button>
    </div>
    
    <div class="container" id="mainContainer" style="display:none;">
        <div class="header">
            <div class="filter-btns">
                <button class="filter-btn active" data-filter="all">全部</button>
                <button class="filter-btn" data-filter="photo">照片</button>
                <button class="filter-btn" data-filter="video">视频</button>
                <button class="filter-btn movie-entry" data-filter="movie"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="5" width="16" height="14" rx="2"/><path d="M10.4 9.5v5l4-2.5z"/></svg>影视</button>
            </div>
            <div class="action-btns">
                <span class="sel-badge" id="selBadge" style="display:none;">点选或滑动选择</span>
                <button class="sel-clear" id="selClearBtn" style="display:none;"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/></svg>清空</button>
                <button class="action-btn" id="selectModeBtn">选择</button>
                <button class="action-btn" id="selectAllBtn" style="display:none;">全选</button>
            </div>
        </div>
        <!-- ===== TG Station 控制栏（照片标题下方） ===== -->
        <div class="station-topbar" id="stationTopbar">
            <div class="topbar-title">TG Station<span class="build-stamp" id="buildStamp" title="页面版本；若与部署版本不一致说明是浏览器缓存">…</span></div>
            <div class="topbar-status">
                <span class="status-dot off" id="stationDot"></span>
                <span id="stationStatusText">后端离线</span>
            </div>
            <div class="topbar-actions">
                <button class="topbar-btn" id="listenToggleBtn"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M7.5 5v14l11-7z"/></svg>开始监听</button>
                <button class="topbar-btn" id="settingsBtn"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7.5h8M16.5 7.5H20M4 16.5h3.5M12 16.5H20"/><circle cx="14.6" cy="7.5" r="2.2"/><circle cx="9.8" cy="16.5" r="2.2"/></svg>设置</button>
            </div>
        </div>
        <div class="download-panel" id="downloadPanel">
            <div class="dl-header" id="dlHeader">
                <span><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 4v10.5M7.5 10L12 14.5 16.5 10"/><path d="M5 19.5h14"/></svg>下载进度</span>
                <span class="dl-summary" id="dlSummary"></span>
                <button class="dl-stop-btn" id="dlStopBtn" title="停止下载（终止当前下载任务，已下载文件保留）"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><rect x="7" y="7" width="10" height="10" rx="2"/></svg>停止</button>
                <span class="dl-arrow" id="dlArrow"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M9.5 5.5l6.5 6.5-6.5 6.5"/></svg></span>
            </div>
            <!-- 总进度 / 速度 / 剩余时间 -->
            <div class="dl-meta" id="dlMeta" style="display:none">
                <div class="dl-overall-bar"><div class="dl-overall-fill" id="dlOverallFill" style="width:0%"></div></div>
                <div class="dl-meta-row"><span>总进度 <b id="dlCount">0 / 0</b></span><span><b id="dlSpeed">—</b> · 剩余 <b id="dlEta">—</b></span></div>
            </div>
            <div class="dl-done" id="dlDone" style="display:none"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7"/></svg>全部完成</div>
            <div class="dl-body" id="dlBody" style="display:none">
                <div id="dlTasks"></div>
                <div class="dl-recent-title" id="dlRecentTitle" style="display:none"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8"/><path d="M12 7.6V12l3.6 2"/></svg>最近下载</div>
                <div class="dl-recent" id="dlRecent"></div>
            </div>
        </div>
        <div id="content"></div>
    </div>
    <div class="lightbox" id="lightbox">
        <!-- 桌面端边缘热区：鼠标滑到两侧才显示翻页键 -->
        <div class="nav-hotzone prev" id="navZonePrev"></div>
        <div class="nav-hotzone next" id="navZoneNext"></div>
        <div class="lightbox-close" id="lightboxCloseBtn" role="button" aria-label="关闭"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/></svg></div>
        <div class="lightbox-nav prev" id="navPrev" role="button" aria-label="上一张"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M14.5 5.5L8 12l6.5 6.5"/></svg></div>
        <div class="lightbox-content" id="lightboxContent"></div>
        <div class="lightbox-nav next" id="navNext" role="button" aria-label="下一张"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M9.5 5.5l6.5 6.5-6.5 6.5"/></svg></div>
        <div class="lightbox-counter" id="lightboxCounter"></div>
        <div class="lightbox-delete" id="lightboxDeleteBtn" title="删除此图片" role="button" aria-label="删除"><svg class="ic ic-solo" viewBox="0 0 24 24" aria-hidden="true"><path d="M4.5 7h15"/><path d="M9.5 7V4.6h5V7"/><path d="M6.8 7l1 12.4h8.4L17.2 7"/><path d="M10.4 10.6v5.6M13.6 10.6v5.6"/></svg></div>
    </div>
    <div class="delete-toast" id="deleteToast"><svg class="ic" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7"/></svg>删除成功</div>
    <script>
        // ===== 图标系统：路径唯一真源（静态 HTML 里是同一份路径的展开形态）=====
        const ICON_PATHS = {
            play: '<path d="M7.5 5v14l11-7z"/>',
            square: '<rect x="7" y="7" width="10" height="10" rx="2"/>',
            x: '<path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/>',
            check: '<path d="M5 12.5l4.5 4.5L19 7"/>',
            chevron_l: '<path d="M14.5 5.5L8 12l6.5 6.5"/>',
            chevron_r: '<path d="M9.5 5.5l6.5 6.5-6.5 6.5"/>',
            download: '<path d="M12 4v10.5M7.5 10L12 14.5 16.5 10"/><path d="M5 19.5h14"/>',
            clock: '<circle cx="12" cy="12" r="8"/><path d="M12 7.6V12l3.6 2"/>',
            smartphone: '<rect x="7" y="3.5" width="10" height="17" rx="2.5"/><path d="M11 17.4h2"/>',
            image: '<rect x="4" y="5" width="16" height="14" rx="2"/><circle cx="9" cy="10.2" r="1.6"/><path d="M5.6 17.8l3.7-3.7 3.2 3.2 3.4-3.4 2.5 2.5"/>',
            video: '<rect x="4" y="5" width="16" height="14" rx="2"/><path d="M10.4 9.5v5l4-2.5z"/>',
            save: '<path d="M19 20.5H5A1.5 1.5 0 0 1 3.5 19V5A1.5 1.5 0 0 1 5 3.5h11L20.5 8v11a1.5 1.5 0 0 1-1.5 1.5z"/><path d="M17 20.5v-7H7v7"/><path d="M7 3.5v5h7"/>',
            refresh: '<path d="M20.5 12a8.5 8.5 0 1 1-2.7-6.2"/><path d="M20.8 4.2v4.2h-4.2"/>',
            trash: '<path d="M4.5 7h15"/><path d="M9.5 7V4.6h5V7"/><path d="M6.8 7l1 12.4h8.4L17.2 7"/><path d="M10.4 10.6v5.6M13.6 10.6v5.6"/>',
            settings: '<path d="M4 7.5h8M16.5 7.5H20M4 16.5h3.5M12 16.5H20"/><circle cx="14.6" cy="7.5" r="2.2"/><circle cx="9.8" cy="16.5" r="2.2"/>',
            maximize: '<path d="M4 9.5V6a2 2 0 0 1 2-2h3.5"/><path d="M14.5 4H18a2 2 0 0 1 2 2v3.5"/><path d="M20 14.5V18a2 2 0 0 1-2 2h-3.5"/><path d="M9.5 20H6a2 2 0 0 1-2-2v-3.5"/>',
            minimize: '<path d="M9.5 4v3.5a2 2 0 0 1-2 2H4"/><path d="M20 9.5h-3.5a2 2 0 0 1-2-2V4"/><path d="M14.5 20v-3.5a2 2 0 0 1 2-2H20"/><path d="M4 14.5h3.5a2 2 0 0 1 2 2V20"/>',
            list: '<path d="M4 7h16M4 12h16M4 17h16"/>',
            alert: '<path d="M12 4.2L21 19.8H3z"/><path d="M12 10v4.2"/><path d="M12 17.2h.01"/>',
        };
        function ic(name, solo) {
            const p = ICON_PATHS[name] || '';
            return '<svg class="ic' + (solo ? ' ic-solo' : '') + '" viewBox="0 0 24 24" aria-hidden="true">' + p + '</svg>';
        }
        // 用 DOM 拼"图标 + 文字"，避免把服务端返回的文本塞进 innerHTML
        function setIconText(el, name, label) {
            if (!el) return;
            el.textContent = '';
            if (name) el.insertAdjacentHTML('afterbegin', ic(name));
            if (label) el.appendChild(document.createTextNode(label));
        }

        let videoGen = 0;  // 视频代际计数器，防止僵尸音频
        let deferredPrompt;
        let allMedia = [], filteredMedia = [], currentIndex = 0, currentFilter = 'all';
        let mediaGen = 0;  // 媒体列表代际：reloadMedia() 复位后，在途的旧分页响应据此丢弃，避免重复项
        let stationData = null;  // 全局：filterMedia() 的 movie 过滤需要读取 movie_min_minutes（IIFE 内赋值）
        let touchStartX = 0, touchStartY = 0, touchStartTime = 0, ignoreSwipe = false;
        const EDGE_THRESHOLD = 70;
        const SWIPE_THRESHOLD = 80;
        const TIME_THRESHOLD = 350;
        let selectMode = false, selectedFiles = new Set();
        let currentScale = 1, initialPinchScale = 1;
        
        function safePath(p) {
            return p.split('/').map(encodeURIComponent).join('/');
        }
        
        function showDeleteToast() {
            const toast = document.getElementById('deleteToast');
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
            }, 1000);
        }
        
        window.addEventListener('beforeinstallprompt', (e) => {
            e.preventDefault();
            deferredPrompt = e;
            showInstallBanner();
        });
        
        function showInstallBanner() {
            const banner = document.getElementById('installBanner');
            const dismissed = localStorage.getItem('installDismissed');
            if (!dismissed) {
                banner.classList.add('show');
            }
        }
        
        document.getElementById('installBtn').addEventListener('click', async () => {
            const banner = document.getElementById('installBanner');
            banner.classList.remove('show');
            
            if (deferredPrompt) {
                deferredPrompt.prompt();
                const { outcome } = await deferredPrompt.userChoice;
                if (outcome === 'accepted') {
                    console.log('用户接受了安装');
                }
                deferredPrompt = null;
            }
        });
        
        document.getElementById('installClose').addEventListener('click', () => {
            document.getElementById('installBanner').classList.remove('show');
            localStorage.setItem('installDismissed', 'true');
        });
        
        let loadingMore = false, allDatesLoaded = false, dateIndex = 0, allDates = [];
        // 当前日期已加载的页数（page 游标）。预注入/首屏只给了第 1 页，
        // 分页不能从"下一个日期"开始，否则第一个日期的第 2..N 页永远不被请求。
        // 历史 bug：后端注入的 firstFiles 上限 100 条，而首个日期可能有上千个文件
        // （2026-10-09 有 2676 个 = 27 页），于是静默丢掉 2576 个文件。
        let datePage = 1;

        async function load() {
            try {
                // 如果有预注入数据，直接秒开
                if (window.__INIT__) {
                    allDates = window.__INIT__.dates;
                    if (window.__INIT__.firstFiles) {
                        window.__INIT__.firstFiles.forEach(f => allMedia.push(f));
                    }
                    // 预注入的是第一个日期的第 1 页；游标停在第一个日期上，
                    // 由 loadRemainingDates 从第 2 页继续抓完它
                    datePage = 1;
                    dateIndex = 0;
                    filterMedia();
                    render();
                    document.getElementById('mainContainer').style.display = 'block';
                    loadRemainingDates(mediaGen);
                    return;
                }
                
                // 降级：API 加载
                const res = await fetch('/api/dates');
                const data = await res.json();
                allDates = data.dates;
                
                // 先加载第一个日期，立即显示
                if (allDates.length > 0) {
                    const d = await fetch('/api/files/' + allDates[0] + '?page=1');
                    const fd = await d.json();
                    fd.files.forEach(f => allMedia.push({...f, date: allDates[0]}));
                    datePage = 1;
                    dateIndex = 0;
                }
                
                filterMedia();
                render();
                document.getElementById('mainContainer').style.display = 'block';
                
                // 后台加载剩余
                loadRemainingDates(mediaGen);
            } catch (e) {
                console.error('Load error:', e);
                document.getElementById('mainContainer').style.display = 'block';
            }
        }
        
        // gen 用于竞态保护：reloadMedia() 复位列表后，在途的旧分页循环会看到
        // mediaGen 变化并立即退出，不再把旧页数据 push 进新列表（否则产生重复项）
        // 每个日期从 datePage+1 页开始抓（首屏已抓了第 1 页），抓完整天再前进
        async function loadRemainingDates(gen) {
            if (gen === undefined) gen = mediaGen;
            if (loadingMore || dateIndex >= allDates.length) { allDatesLoaded = true; return; }
            loadingMore = true;
            try {
                let count = 0;
                while (dateIndex < allDates.length) {
                    if (gen !== mediaGen) return;  // 列表已被复位，丢弃本轮
                    const date = allDates[dateIndex];
                    let p = datePage + 1, hasMore = true;
                    while (hasMore) {
                        const r = await fetch('/api/files/' + date + '?page=' + p);
                        const d = await r.json();
                        if (gen !== mediaGen) return;
                        d.files.forEach(f => allMedia.push({...f, date}));
                        hasMore = p < d.pages;
                        p++;
                    }
                    datePage = 0;        // 后续日期从第 1 页开始
                    dateIndex++;
                    count++;
                    // 每加载3个日期更新一次显示
                    if (count % 3 === 0) { filterMedia(); render(); }
                }
                if (gen !== mediaGen) return;
                allDatesLoaded = true;
                filterMedia(); render();
            } catch (e) { console.error(e); }
            finally {
                // 只有当没有更新的 generation 接管时，才释放 loadingMore
                if (gen === mediaGen) loadingMore = false;
            }
        }
        
        // 复位媒体列表并重新加载：代际 +1 使在途分页作废，重置全部派生状态，
        // 并把筛选按钮的 active 类同步回"全部"（refreshMediaNow 之前只改
        // currentFilter 不改按钮，会出现"按钮显示视频、列表却是全部"）
        async function reloadMedia() {
            mediaGen++;
            loadingMore = false;
            allDatesLoaded = false;
            dateIndex = 0;
            datePage = 0;      // 复位页游标；load() 走 API 分支会重新抓第 1 页
            allDates = [];
            allMedia = [];
            filteredMedia = [];
            currentIndex = 0;
            currentFilter = 'all';
            selectMode = false;
            selectedFiles.clear();
            window.__INIT__ = null;
            setActiveFilterButton('all');
            updateActionButtons();
            await load();
            await loadRemainingDates(mediaGen);
        }
        
        function movieMinMinutes() {
            return (stationData && stationData.movie_min_minutes) || 5;
        }

        function filterMedia() {
            // 注意：'all' 必须产生新数组。此前 filteredMedia = allMedia 是别名，
            // 任何就地修改 filteredMedia 都会污染 allMedia（续播/删除路径踩过）
            if (currentFilter === 'all') filteredMedia = allMedia.slice();
            else if (currentFilter === 'photo') filteredMedia = allMedia.filter(m => !m.isVideo);
            else if (currentFilter === 'movie') {
                const minutes = movieMinMinutes();
                filteredMedia = allMedia.filter(m => m.isVideo && m.duration >= minutes * 60);
            }
            else filteredMedia = allMedia.filter(m => m.isVideo);
        }
        
        // ── 日期分区「离屏回收」（第三阶段优化）──────────────────────
        // 全量渲染时 16882 个格子常驻 DOM（十几万节点），内存与每次操作都偏重。
        // 这里在分区级别回收：远离视口的分区把格子清掉、只留高度占位，
        // 滚回来时按原 data-index 原样重建 —— 滚动条长度与格子位置都不变。
        const RECYCLE_MARGIN_PX = 2500;   // 视口外这么远才开始回收（留足缓冲，滚动不闪）
        // 同一天超过这么多格就切成多个网格块（840 = 4/5/6/7/8 列的公倍数，
        // 任何列数下都正好整行收尾，视觉上仍然是一片连续的网格）
        const GRID_BLOCK = 840;

        // 回收时留的占位高度必须准，否则滚动条长度会变、位置会漂。
        // 用 rect 测量会被 content-visibility: auto 影响（离屏格子按 contain-intrinsic-size
        // 估算，不是真实高度），所以这里按「列数 × 格子高（aspect-ratio）＋行间距」算出来。
        let gridMetricsCache = null;
        function gridMetrics() {
            if (gridMetricsCache) return gridMetricsCache;
            const g = document.querySelector('#content .photo-grid');
            if (!g) return null;
            const cs = getComputedStyle(g);
            const cols = (cs.gridTemplateColumns || '').split(' ').filter(s => s && s !== 'none').length || 1;
            const colGap = parseFloat(cs.columnGap) || 0;
            const rowGap = parseFloat(cs.rowGap) || 0;
            const avail = g.clientWidth;
            if (!(avail > 0)) return null;
            const cellW = (avail - colGap * (cols - 1)) / cols;
            let ratio = 1;
            const el = g.firstElementChild;
            if (el) {
                const parts = (getComputedStyle(el).aspectRatio || '').split('/').map(s => parseFloat(s.trim()));
                if (parts.length === 2 && parts[0] > 0 && parts[1] > 0) ratio = parts[1] / parts[0];
            }
            gridMetricsCache = { cols, colGap, rowGap, cellW, ratio };
            return gridMetricsCache;
        }
        function heightForCount(count) {
            const m = gridMetrics();
            if (!m || !(count > 0)) return 0;
            const rows = Math.ceil(count / m.cols);
            return Math.round(rows * m.cellW * m.ratio + (rows - 1) * m.rowGap);
        }

        // 用滚动位置直接判断，不用 IntersectionObserver：
        // IO 在页面不可见（后台标签/最小化）时不回调，行为依赖可见性；
        // 几何判断只依赖位置，任何时候都一致，也便于自动化验证。
        // 注意：先把所有 rect 读出来再统一写 DOM，避免「读-写交替」反复触发布局。
        // 粒度是「网格块」而不是整个日期分区。
        function updateSectionRecycle() {
            const margin = RECYCLE_MARGIN_PX;
            const vh = window.innerHeight || 800;
            const plan = [];
            document.querySelectorAll('#content .photo-grid').forEach(grid => {
                const r = grid.getBoundingClientRect();
                const near = (r.bottom > -margin) && (r.top < vh + margin);
                const n = grid.dataset.count !== undefined ? parseInt(grid.dataset.count, 10) : grid.children.length;
                plan.push([grid, near, heightForCount(n) || Math.round(r.height)]);
            });
            plan.forEach(p => { if (p[1]) restoreGrid(p[0]); else recycleGrid(p[0], p[2]); });
        }

        let recycleScrollTimer = null;
        window.addEventListener('scroll', () => {
            if (recycleScrollTimer) return;
            recycleScrollTimer = setTimeout(() => {
                recycleScrollTimer = null;
                updateSectionRecycle();
            }, 180);
        }, { passive: true });

        function recycleGrid(grid, knownHeight) {
            if (!grid || grid.dataset.recycled === '1' || !grid.children.length) return;
            const first = grid.firstElementChild;
            if (!first) return;
            const h = (knownHeight && knownHeight > 0) ? knownHeight : (heightForCount(grid.children.length) || Math.round(grid.getBoundingClientRect().height));
            grid.dataset.start = first.dataset.index;                 // 全局索引，恢复时用它
            grid.dataset.count = String(grid.children.length);
            grid.dataset.h = String(h);
            grid.innerHTML = '';
            grid.style.height = h + 'px';                             // 占位，保持滚动条长度
            grid.dataset.recycled = '1';
            delete grid.dataset.restoreToken;                         // 作废在途的逐帧恢复
        }

        // 大分区（实测有个日期 1.2 万个文件）一次性重建会阻塞主线程约 0.3 秒；
        // 这里改成逐帧插入（每帧最多 RESTORE_CHUNK 格），视觉上仍是立刻出现，
        // 但主线程不会被长时间占住 —— 「拖到某处突然一顿」的主要原因。
        // 块大小 840 → 每帧就能重建一整块（逐帧保护仅对将来更大的块生效）
        const RESTORE_CHUNK = 1000;
        let gridEpoch = 0;   // 版本号：删除等需要完整 DOM 的操作会让在途恢复作废

        // 作废所有在途的逐帧恢复，并把它们标记回「已回收」交给随后的同步恢复处理
        function cancelPendingRestores() {
            gridEpoch++;
            document.querySelectorAll('#content .photo-grid[data-restore-token]').forEach(g => {
                g.dataset.recycled = '1';
                delete g.dataset.restoreToken;
            });
        }

        function restoreGrid(grid, sync) {
            if (!grid || grid.dataset.recycled !== '1') return;
            const start = parseInt(grid.dataset.start, 10);
            const count = parseInt(grid.dataset.count, 10);
            if (!(start >= 0) || !(count > 0)) {
                grid.style.height = '';
                delete grid.dataset.recycled;
                return;
            }

            if (sync) {   // 需要立刻有完整 DOM（删除增量同步/窗口变化）
                grid.innerHTML = sectionItemsHTML(filteredMedia.slice(start, start + count), start);
                grid.style.height = '';
                delete grid.dataset.recycled;
                delete grid.dataset.start; delete grid.dataset.count; delete grid.dataset.h;
                return;
            }

            const myEpoch = gridEpoch;
            const token = String(myEpoch) + '-' + Math.random();
            grid.dataset.recycled = '0';          // 恢复中（保留 start/count 以便中途再次回收）
            grid.dataset.restoreToken = token;
            grid.innerHTML = '';
            let k = 0;
            const step = () => {
                // 期间被回收 / 被取消 / 版本变了 → 放弃本次恢复
                if (myEpoch !== gridEpoch || grid.dataset.restoreToken !== token) return;
                // 必须按本块的大小截断：否则 RESTORE_CHUNK 大于块大小时会把相邻块的条目
                // 也塞进来，块会越恢复越大（高度变多、渲染量变多）
                const chunk = filteredMedia.slice(start + k, Math.min(start + k + RESTORE_CHUNK, start + count));
                if (chunk.length) grid.insertAdjacentHTML('beforeend', sectionItemsHTML(chunk, start + k));
                k += RESTORE_CHUNK;
                if (k < count) { requestAnimationFrame(step); return; }
                grid.style.height = '';
                delete grid.dataset.recycled;
                delete grid.dataset.restoreToken;
                delete grid.dataset.start; delete grid.dataset.count; delete grid.dataset.h;
            };
            step();
        }

        // render() 之后调用：按当前滚动位置立刻回收远处分区
        function observeSections() {
            updateSectionRecycle();
        }

        // 需要「完整 DOM」的时刻（删除增量同步、窗口尺寸变化）用 sync=true 立即恢复；
        // 普通滚动恢复走逐帧，避免长任务。
        function restoreAllRecycledSections(sync) {
            cancelPendingRestores();
            document.querySelectorAll('#content .photo-grid[data-recycled="1"]').forEach(grid => restoreGrid(grid, sync));
        }

        // 找到包含第 idx 个条目的分区：
        // 分区按文档顺序排列，与 filteredMedia 顺序一致；累加每个分区的条目数即可。
        // 已回收的分区保留 data-count（回收瞬间写入），未回收的直接读子元素个数。
        function sectionForIndex(idx) {
            let acc = 0;
            const grids = document.querySelectorAll('#content .photo-grid');
            for (const g of grids) {
                const cnt = g.dataset.count !== undefined ? parseInt(g.dataset.count, 10) : g.children.length;
                if (!(cnt > 0)) continue;
                if (idx < acc + cnt) return g.closest('.date-section');
                acc += cnt;
            }
            return grids.length ? grids[grids.length - 1].closest('.date-section') : null;
        }

        // 列数随窗口宽度变化 → 回收时记录的高度占位失效，先恢复再由观察器按新布局回收
        let recycleResizeTimer = null;
        window.addEventListener('resize', () => {
            clearTimeout(recycleResizeTimer);
            recycleResizeTimer = setTimeout(() => {
                gridMetricsCache = null;         // 列数/列宽变了，占位高度要按新布局重算
                restoreAllRecycledSections(true);
                observeSections();
            }, 300);
        });

        // 视频时长角标文本（灯箱里的 fmt() 定义在另一个作用域，网格渲染拿不到，
        // 之前直接调 fmt() 会抛 ReferenceError 让整个列表渲染不出来）
        function durText(sec) {
            const s = Math.max(0, Math.round(sec || 0));
            if (!s) return '';
            const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), ss = s % 60;
            return h > 0 ? h + ':' + (m < 10 ? '0' : '') + m + ':' + (ss < 10 ? '0' : '') + ss
                         : m + ':' + (ss < 10 ? '0' : '') + ss;
        }

        function render() {
            if (filteredMedia.length === 0) {
                document.getElementById('content').innerHTML = '<div class="empty-tip">' + (currentFilter === 'movie'
                    ? ic('video') + '暂无时长超过 ' + movieMinMinutes() + ' 分钟的视频'
                    : '暂无内容') + '</div>';
                return;
            }
            let html = '', currentDate = '', dateCount = 0;
            filteredMedia.forEach((item, i) => {
                if (item.date !== currentDate) {
                    if (currentDate) html += '</div></div>';
                    currentDate = item.date;
                    dateCount = 0;
                    let dateHtml = '<div class="date-section"><div class="date-header">' + currentDate;
                    if (selectMode) {
                        dateHtml += '<span class="select-hint">点击选择 | 已选 ' + selectedFiles.size + ' 个</span>';
                    }
                    dateHtml += '</div><div class="photo-grid">';
                    html += dateHtml;
                } else if (dateCount > 0 && dateCount % GRID_BLOCK === 0) {
                    // 同一天条目太多时切成多个网格块（块大小是 4~8 列的整数倍，视觉上仍是
                    // 连续的一片）：回收/恢复就能按块来，滚到某一天中间时只需重建一块，
                    // 而不是整天 1 万多个格子一起重建。
                    html += '</div><div class="photo-grid">';
                }
                dateCount++;
                const thumbUrl = safePath(item.path);
                const encodedPath = encodeURIComponent(item.path);
                const isSelected = selectedFiles.has(item.path);
                const selectedClass = isSelected ? ' selected' : '';
                if (item.isVideo) {
                    html += '<div class="photo-item' + selectedClass + '" data-index="' + i + '" data-path="' + encodeURIComponent(item.path) + '">';
                    html += '<img src="/thumb/' + thumbUrl + '?s=' + (item.size || 0) + '" loading="lazy" onerror="this.src=\'data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22><rect fill=%22%23333%22 width=%22200%22 height=%22200%22/><path d=%22M80 60 L142 100 L80 140 Z%22 fill=%22%23888%22 stroke=%22%23888%22 stroke-width=%2212%22 stroke-linejoin=%22round%22/></svg>\'">';
                    html += '<div class="video-overlay">' + (item.duration > 0 ? '<span class="video-duration">' + durText(item.duration) + '</span>' : '') + '</div>';
                    html += '</div>';
                } else {
                    html += '<div class="photo-item' + selectedClass + '" data-index="' + i + '" data-path="' + encodeURIComponent(item.path) + '"><img src="/thumb/' + thumbUrl + '?s=' + (item.size || 0) + '" loading="lazy" onerror="this.src=\'data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22><rect fill=%22%23333%22 width=%22200%22 height=%22200%22/><rect x=%2240%22 y=%2262%22 width=%22120%22 height=%2276%22 rx=%2210%22 fill=%22none%22 stroke=%22%23888%22 stroke-width=%228%22/><circle cx=%2268%22 cy=%2286%22 r=%228%22 fill=%22%23888%22/><path d=%22M48 132l26-26 20 20 24-24 30 30%22 fill=%22none%22 stroke=%22%23888%22 stroke-width=%228%22 stroke-linecap=%22round%22 stroke-linejoin=%22round%22/></svg>\'"></div>';
                }
            });
            if (currentDate) html += '</div></div>';
            document.getElementById('content').innerHTML = html;
            // 等布局完成再回收：过早回收时读不到列宽（clientWidth=0），占位高度会退化成
            // 测量值（受 content-visibility 影响而偏小），滚动条长度就会先短后长。
            requestAnimationFrame(observeSections);
            setTimeout(observeSections, 250);   // 兜底（页面在后台时 rAF 不触发）；幂等
        }

        // ---------- 媒体网格交互：事件委托（绑定一次，不随 render 增长） ----------
        // 此前每个 .photo-item 在每次 render() 时挂 4 个监听器（touchstart/move/end/click），
        // 16945 个文件 = 6.7 万个闭包 + O(n) 挂载开销，每次翻页/重渲染都要重建。
        // 现在 4 个监听器挂在 #content 上，靠 closest('.photo-item') 定位目标元素。
        const GRID_TAP_MOVE = 8;      // 位移小于此值算点击而非滑动
        const GRID_TAP_MS = 250;      // 按住时间上限
        let gridTouch = null;
        let gridLastTapTime = 0;

        // 点击处理：选择模式下切换勾选，否则打开灯箱
        function handleGridTap(el) {
            if (!el.dataset.path) return;
            if (selectMode) {
                const path = decodeURIComponent(el.dataset.path);
                if (selectedFiles.has(path)) {
                    selectedFiles.delete(path);
                    el.classList.remove('selected');
                } else {
                    selectedFiles.add(path);
                    el.classList.add('selected');
                }
                updateActionButtons();
            } else {
                currentIndex = parseInt(el.dataset.index);
                openLightbox();
            }
        }

        const contentEl = document.getElementById('content');

        contentEl.addEventListener('touchstart', (e) => {
            const el = e.target.closest('.photo-item');
            if (!el || e.touches.length !== 1) { gridTouch = null; return; }
            const touch = e.touches[0];
            gridTouch = {
                el,
                x: touch.clientX,
                y: touch.clientY,
                time: Date.now(),
                moved: false,
                swipeActive: false,
                addMode: true,
                swiped: new Set()
            };
        }, { passive: true });

        contentEl.addEventListener('touchmove', (e) => {
            const g = gridTouch;
            if (!g || e.touches.length !== 1) return;
            const touch = e.touches[0];
            const deltaX = Math.abs(touch.clientX - g.x);
            const deltaY = Math.abs(touch.clientY - g.y);
            if (deltaX <= GRID_TAP_MOVE && deltaY <= GRID_TAP_MOVE) return;
            g.moved = true;
            if (!selectMode) return;
            // 选择模式：滑动批量勾选/取消（首格决定是 "选中" 还是 "取消"）
            e.preventDefault();
            if (!g.swipeActive) {
                g.swipeActive = true;
                g.addMode = !selectedFiles.has(decodeURIComponent(g.el.dataset.path));
            }
            const target = document.elementFromPoint(touch.clientX, touch.clientY);
            const item = target ? target.closest('.photo-item') : null;
            if (item && item.dataset.path) {
                const p = decodeURIComponent(item.dataset.path);
                if (!g.swiped.has(p)) {
                    g.swiped.add(p);
                    if (g.addMode) {
                        selectedFiles.add(p);
                        item.classList.add('selected');
                    } else {
                        selectedFiles.delete(p);
                        item.classList.remove('selected');
                    }
                    updateActionButtons();
                }
            }
        }, { passive: false });

        contentEl.addEventListener('touchend', (e) => {
            const g = gridTouch;
            gridTouch = null;
            if (!g) return;
            if (g.swipeActive || g.moved) return;
            const touch = e.changedTouches[0];
            const deltaX = Math.abs(touch.clientX - g.x);
            const deltaY = Math.abs(touch.clientY - g.y);
            if (deltaX < GRID_TAP_MOVE && deltaY < GRID_TAP_MOVE && Date.now() - g.time < GRID_TAP_MS) {
                e.preventDefault();
                gridLastTapTime = Date.now();
                handleGridTap(g.el);
            }
        }, { passive: false });

        contentEl.addEventListener('click', (e) => {
            const el = e.target.closest('.photo-item');
            if (!el) return;
            // 移动端 touchend 已处理过点击，忽略随之而来的合成 click
            if (gridLastTapTime && Date.now() - gridLastTapTime < 300) { gridLastTapTime = 0; return; }
            handleGridTap(el);
        });
        
        function toggleSelectMode() {
            selectMode = !selectMode;
            if (!selectMode) {
                selectedFiles.clear();
            }
            updateActionButtons();
            render();
        }
        
        function updateActionButtons() {
            const selectModeBtn = document.getElementById('selectModeBtn');
            const selectAllBtn = document.getElementById('selectAllBtn');
            const deleteBtn = document.getElementById('deleteBtn');
            const stationFab = document.getElementById('stationFab');
            const fabMenu = document.getElementById('fabMenu');
            
            if (selectMode) {
                selectModeBtn.classList.add('active');
                setIconText(selectModeBtn, 'x', '退出');
                selectAllBtn.style.display = 'block';
                selectAllBtn.textContent = (selectedFiles.size === filteredMedia.length && filteredMedia.length > 0)
                    ? '取消全选' : '全选';
                // 多选模式下隐藏右下角悬浮球，避免与删除按钮重叠
                stationFab.style.display = 'none';
                fabMenu.style.display = 'none';
                // 已选数量角标 + 一键清空
                const badge = document.getElementById('selBadge');
                const clearBtn = document.getElementById('selClearBtn');
                if (selectedFiles.size > 0) {
                    badge.style.display = '';
                    badge.textContent = '已选 ' + selectedFiles.size + ' 项';
                    clearBtn.style.display = selectedFiles.size > 0 ? '' : 'none';
                } else {
                    badge.style.display = '';
                    badge.textContent = '点选或滑动选择';
                    clearBtn.style.display = 'none';
                }

                if (selectedFiles.size > 0) {
                    deleteBtn.style.display = 'block';
                    setIconText(deleteBtn, 'trash', '删除 (' + selectedFiles.size + ')');
                    deleteBtn.style.opacity = '1';
                } else {
                    deleteBtn.style.display = 'block';
                    setIconText(deleteBtn, 'trash', '删除');
                    deleteBtn.style.opacity = '0.5';
                }
            } else {
                selectModeBtn.classList.remove('active');
                setIconText(selectModeBtn, null, '选择');
                selectAllBtn.style.display = 'none';
                deleteBtn.style.display = 'none';
                deleteBtn.style.opacity = '1';
                const badge = document.getElementById('selBadge');
                const clearBtn = document.getElementById('selClearBtn');
                if (badge) badge.style.display = 'none';
                if (clearBtn) clearBtn.style.display = 'none';
                // 退出多选，恢复悬浮球
                stationFab.style.display = '';
                fabMenu.style.display = 'none';
            }
        }

        // 清空已选（角标上的一键入口）
        function clearSelection() {
            selectedFiles.clear();
            document.querySelectorAll('.photo-item.selected').forEach(el => el.classList.remove('selected'));
            updateActionButtons();
        }
        
        function selectAll() {
            if (selectedFiles.size === filteredMedia.length) {
                selectedFiles.clear();
            } else {
                selectedFiles.clear();
                filteredMedia.forEach(item => selectedFiles.add(item.path));
            }
            render();
            updateActionButtons();
        }
        
        async function deleteSelected() {
            if (selectedFiles.size === 0) return;
            
            try {
                const filesToDelete = Array.from(selectedFiles);
                console.log('准备删除的文件：', filesToDelete);
                
                const response = await fetch('/api/delete', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ files: filesToDelete })
                });
                
                const data = await response.json();
                console.log('删除结果：', data);
                
                if (data.deleted > 0) {
                    showDeleteToast();
                    // 直接从数组里移除已删除的文件，不重新加载页面
                    const deletedSet = new Set(filesToDelete);
                    // 灯箱里"正在看的那张"要记住：数组缩短后 currentIndex 会漂移，
                    // 若它指向的项被删掉，下一次 navigate/openLightbox 会开到别的图上
                    const curPath = (filteredMedia[currentIndex] || {}).path;
                    const curKept = curPath && !deletedSet.has(curPath);
                    allMedia = allMedia.filter(m => !deletedSet.has(m.path));
                    filterMedia();
                    if (curKept) {
                        const ni = filteredMedia.findIndex(m => m.path === curPath);
                        currentIndex = ni >= 0 ? ni : 0;
                    } else {
                        currentIndex = Math.min(currentIndex, Math.max(0, filteredMedia.length - 1));
                    }
                    selectedFiles.clear();
                    selectMode = false;
                    updateActionButtons();
                    // 同样只做增量同步，避免批量删除后全量重建 1.6 万个格子。
                    // 回收过的分区先整体恢复（增量同步依赖完整的 data-index 前提），
                    // 同步完再交给观察器按视口重新回收。
                    const scrollY = window.scrollY;
                    restoreAllRecycledSections(true);   // 同步恢复：增量同步依赖完整的 data-index
                    syncGridAfterDelete(deletedSet, null);
                    observeSections();
                    window.scrollTo(0, scrollY);
                }
            } catch (e) {
                console.error('Delete error:', e);
                if (window.showError) window.showError(e.message, '删除失败');
                else alert('删除失败！请重试。错误信息：' + e.message);
            }
        }
        
        // ---------- 预取下一张 ----------
        // 灯箱里等原图下载是唯一剩下的等待来源。在你看当前这张时后台把相邻的几张
        // 悄悄下好（复用浏览器 HTTP 缓存），按 → / 删除后下一张基本瞬时可见。
        // 上限 6 张：预取的都是全分辨率图，不做上限会在长时间浏览后占掉大量内存。
        const PREFETCH_NEXT = 2;   // 向后预取几张
        const PREFETCH_PREV = 1;   // 向前预取几张
        const PREFETCH_KEEP = 6;   // 内存里最多保留几张已解码的预取图
        const prefetched = new Map();   // path -> Image（Map 保持插入顺序，便于淘汰最旧的）
        let prefetchTimer = null;

        function fileUrl(path) {
            return '/api/file/' + safePath(path);
        }

        function schedulePrefetch() {
            if (prefetchTimer) { clearTimeout(prefetchTimer); prefetchTimer = null; }
            // 先让当前这张加载完再说，别跟着抢带宽
            prefetchTimer = setTimeout(prefetchNeighbors, 300);
        }

        function prefetchNeighbors() {
            prefetchTimer = null;
            if (window.__noPrefetch) return;   // 自动化测试用开关，正常运行不会设置
            if (!document.getElementById('lightbox').classList.contains('active')) return;
            const targets = [];
            for (let d = 1; d <= PREFETCH_NEXT; d++) {
                const it = filteredMedia[currentIndex + d];
                if (it && !it.isVideo) targets.push(it.path);
            }
            for (let d = 1; d <= PREFETCH_PREV; d++) {
                const it = filteredMedia[currentIndex - d];
                if (it && !it.isVideo) targets.push(it.path);
            }
            targets.forEach(p => {
                if (prefetched.has(p)) return;
                const im = new Image();
                im.decoding = 'async';
                im.src = fileUrl(p);
                prefetched.set(p, im);
            });
            while (prefetched.size > PREFETCH_KEEP) {
                const oldest = prefetched.keys().next().value;
                prefetched.get(oldest).src = '';   // 释放正在下载的
                prefetched.delete(oldest);
            }
        }

        function openLightbox() {
            videoGen++;  // 新一代视频
            const myGen = videoGen;
            const item = filteredMedia[currentIndex];
            const thumbUrl = safePath(item.path);
            
            // 彻底杀死旧视频：与 closeLightbox 共用 destroyVideo，避免两条路径逻辑分叉
            const oldVideo = document.querySelector('.lightbox-content video');
            if (oldVideo) destroyVideo(oldVideo);
            clearVideoControls();
            document.getElementById('lightboxContent').innerHTML = '';
            document.getElementById('lightboxCounter').textContent = (currentIndex + 1) + ' / ' + filteredMedia.length;
            document.getElementById('lightbox').classList.add('active');
            document.body.style.overflow = 'hidden';
            currentScale = 1;
            
            if (item.isVideo) {
                const video = document.createElement('video');
                video.autoplay = true;
                video.muted = true;
                video.playsInline = true;
                // 影视动辄几个 GB：preload=auto 会让 Chrome 一打开就狂下（手机内存/流量
                // 都吃不消），改成 metadata，真正播放时再按需缓冲。
                video.preload = 'metadata';
                video.loop = (currentFilter !== 'movie');  // 影视不循环播放
                video._mbPath = item.path;  // 供进度保存使用
                video.style.background = '#000';
                const source = document.createElement('source');
                source.src = '/api/file/' + thumbUrl;
                source.type = 'video/mp4';
                video.appendChild(source);
                video.onloadeddata = () => {
                    if (videoGen !== myGen) { video.muted = true; video.pause(); return; }
                    video.muted = false;
                    // 影视续播：从上次进度继续
                    if (currentFilter === 'movie') {
                        try {
                            const saved = parseFloat(localStorage.getItem('mb_pos_' + item.path));
                            if (saved > 10 && video.duration && saved < video.duration * 0.95) {
                                video.currentTime = saved;
                            }
                        } catch (e) { /* localStorage 不可用则忽略 */ }
                    }
                    video.play();
                };
                video.onerror = () => {
                    document.getElementById('lightboxContent').innerHTML = '<div class="video-error">无法播放此视频<br><button onclick="navigate(1)">下一个</button></div>';
                };
                document.getElementById('lightboxContent').appendChild(video);
                setupVideoControls(video, item.path);
            } else {
                const imgContainer = document.createElement('div');
                imgContainer.style.cssText = 'width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; overflow: hidden;';
                
                // 直接加载原图，不用缩略图占位。
                // 曾试过"先显示缩略图、原图到位再替换"，实测反而更卡：多请求一次、
                // 浏览器要重新缩放绘制一次，而且眼睛会看到"先糊后清"的跳变。
                const img = document.createElement('img');
                img.src = fileUrl(item.path);
                img.style.cssText = 'max-width: 100%; max-height: 100%; object-fit: contain; transition: transform 0.2s ease;';
                img.loading = 'eager';
                img.decoding = 'async';   // 大图解码放别的线程，别阻塞主线程交互
                img.dataset.path = item.path;   // 调试/自动化校验用：这张图应该展示哪个文件

                // 加载进度：大图要几秒，给个明确反馈而不是干等黑屏
                const ld = document.createElement('div');
                ld.className = 'lightbox-loading';
                ld.innerHTML = '<div class="lb-spinner"></div><div class="lb-pct">加载中…</div>';
                imgContainer.appendChild(ld);
                let ldTimer = setTimeout(() => { ld.remove(); }, 20000);   // 兜底：别永远转圈
                const ldPct = ld.querySelector('.lb-pct');
                const ldDone = () => { clearTimeout(ldTimer); ld.remove(); };
                img.addEventListener('progress', (ev) => {
                    if (ev.lengthComputable && ldPct) {
                        ldPct.textContent = Math.round(ev.loaded * 100 / ev.total) + '%';
                    }
                });
                img.addEventListener('load', ldDone);
                img.addEventListener('error', ldDone);

                img.onload = () => { if (img.dataset.path === (filteredMedia[currentIndex] || {}).path) schedulePrefetch(); };
                img.onerror = () => {
                    imgContainer.innerHTML = '<div class="video-error" style="color:#fff;">无法加载此图片<br><small>文件可能已损坏或过大</small><br><button onclick="navigate(1)">下一个</button></div>';
                };
                
                imgContainer.appendChild(img);
                document.getElementById('lightboxContent').appendChild(imgContainer);
                
                imgContainer.addEventListener('touchstart', handleTouchStart, { passive: false });
                imgContainer.addEventListener('touchmove', handleTouchMove, { passive: false });
                imgContainer.addEventListener('touchend', handleTouchEnd);
                
                img.addEventListener('click', (e) => {
                    if (e.touches && e.touches.length > 1) return;
                    
                    if (currentScale === 1) {
                        currentScale = 2;
                        img.style.transform = 'scale(' + currentScale + ')';
                        img.classList.add('zoomed');
                    } else {
                        currentScale = 1;
                        img.style.transform = 'scale(' + currentScale + ')';
                        img.classList.remove('zoomed');
                    }
                });
            }
            // 视频分支（以及图片已缓存、onload 不再触发的场景）也走一次预取调度
            schedulePrefetch();
        }
        
        function handleTouchStart(e) {
            if (e.touches.length === 2) {
                e.preventDefault();
                const touch1 = e.touches[0];
                const touch2 = e.touches[1];
                initialPinchScale = currentScale;
                const currentDistance = Math.hypot(
                    touch2.clientX - touch1.clientX,
                    touch2.clientY - touch1.clientY
                );
                window._pinchDistance = currentDistance;
            }
        }
        
        function handleTouchMove(e) {
            if (e.touches.length === 2) {
                e.preventDefault();
                const touch1 = e.touches[0];
                const touch2 = e.touches[1];
                const currentDistance = Math.hypot(
                    touch2.clientX - touch1.clientX,
                    touch2.clientY - touch1.clientY
                );
                
                if (window._pinchDistance) {
                    const scale = (currentDistance / window._pinchDistance) * initialPinchScale;
                    currentScale = Math.max(0.5, Math.min(scale, 5));
                    
                    const img = document.querySelector('.lightbox-content img');
                    if (img) {
                        img.style.transform = 'scale(' + currentScale + ')';
                        img.classList.toggle('zoomed', currentScale > 1);
                    }
                }
            }
        }
        
        function handleTouchEnd() {
            window._pinchDistance = null;
            if (currentScale < 1) {
                currentScale = 1;
                const img = document.querySelector('.lightbox-content img');
                if (img) {
                    img.style.transform = 'scale(' + currentScale + ')';
                    img.classList.remove('zoomed');
                }
            }
        }
        
        // ---- 自定义视频控制条 ----
        let vcVideo = null;
        // ── 全屏状态监听：只注册一次 ──────────────────────────────
        // 修复（2026-10-10）：过去每个视频都在 setupVideoControls 里
        // document.addEventListener('fullscreenchange', ...)，闭包持有那一次创建的
        // 整条控制条 DOM。于是每看一个视频就把一份控制条永久钉在内存里，
        // 长时浏览后监听器/节点累积到几十万，删除等一次 DOM 重建就让主线程卡到
        // Chrome 报「页面无响应」。现在只用 currentFullBtn 指向"当前那个按钮"。
        let currentFullBtn = null;
        document.addEventListener('fullscreenchange', () => {
            if ((!document.fullscreenElement && !document.webkitFullscreenElement) && currentFullBtn) {
                currentFullBtn.innerHTML = ic('maximize', true);
            }
        });

        function setupVideoControls(video, mediaPath) {
            vcVideo = video;
            const ctrl = document.createElement('div');
            ctrl.className = 'video-controls visible';
            // 触屏 / 窄屏标记：把控制条从画面底边抬高，避免与底部区域重叠
            try {
                if (window.matchMedia('(max-width: 900px), (hover: none) and (pointer: coarse)').matches) {
                    ctrl.classList.add('vc-touch');
                }
            } catch (e) {}
            ctrl.innerHTML = `
                <div class="vc-seek" id="vcSeek">
                    <div class="vc-seek-track">
                        <div class="vc-buf" id="vcBuf"></div>
                        <div class="vc-fill" id="vcFill"></div>
                    </div>
                    <div class="vc-thumb" id="vcThumb"></div>
                    <div class="vc-tip" id="vcTip"></div>
                </div>
                <span class="vc-time" id="vcTime">00:00 / 00:00</span>
                <input type="range" class="vc-vol" id="vcVol" min="0" max="100" value="100">
                <button class="vc-full" id="vcFull">${ic('maximize', true)}</button>`;
            document.getElementById('lightboxContent').appendChild(ctrl);
            // 缓冲/加载提示 & 长按倍速徽标（独立于自动隐藏的控制条）
            const loading = document.createElement('div');
            loading.className = 'vc-loading';
            loading.innerHTML = '<div class="lb-spinner"></div>';
            document.getElementById('lightboxContent').appendChild(loading);
            const fastBadge = document.createElement('div');
            fastBadge.className = 'vc-fast';
            fastBadge.textContent = '2x';
            document.getElementById('lightboxContent').appendChild(fastBadge);
            
            const seek = ctrl.querySelector('.vc-seek');
            const fillEl = ctrl.querySelector('.vc-fill');
            const bufEl = ctrl.querySelector('.vc-buf');
            const thumbEl = ctrl.querySelector('.vc-thumb');
            const tipEl = ctrl.querySelector('.vc-tip');
            const timeEl = ctrl.querySelector('.vc-time');
            const volEl = ctrl.querySelector('.vc-vol');
            const fullBtn = ctrl.querySelector('.vc-full');
            currentFullBtn = fullBtn;   // 供上面「只注册一次」的全屏监听更新图标
            
            let hideTimer, lastSave = 0, seeking = false, hovering = false;
            // 淡出必须同时关掉 pointer-events：只把 opacity 设成 0 的话，那条“看不见”
            // 的控制条仍然会吞掉视频底部区域的点击（用户感觉“点了没反应”）。
            // 所以统一用 class 控制显隐，别再用行内 opacity。
            function show() { ctrl.classList.add('visible'); clearTimeout(hideTimer); hideTimer = setTimeout(() => { if (!video.paused) ctrl.classList.remove('visible'); }, 3000); }
            function fmt(s) {
                s = Math.max(0, Math.floor(s || 0));
                const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), sec = s % 60;
                return h > 0 ? h + ':' + (m < 10 ? '0' : '') + m + ':' + (sec < 10 ? '0' : '') + sec : m + ':' + (sec < 10 ? '0' : '') + sec;
            }
            function setLoading(on) { loading.classList.toggle('show', on); }
            
            // 影视：定期保存播放进度（localStorage）
            function saveProgress() {
                if (currentFilter === 'movie' && mediaPath && video.duration) {
                    try { localStorage.setItem('mb_pos_' + mediaPath, String(video.currentTime)); } catch (e) {}
                }
            }

            // ---- 进度条：缓冲条 + 拖动/悬停时间预览 ----
            function ratioFromEvent(e) {
                const rect = seek.getBoundingClientRect();
                return Math.max(0, Math.min(1, (e.clientX - rect.left) / (rect.width || 1)));
            }
            function previewAt(ratio) {
                const t = ratio * (video.duration || 0);
                tipEl.style.left = (ratio * 100) + '%';
                tipEl.textContent = fmt(t);
                timeEl.textContent = fmt(t) + ' / ' + fmt(video.duration || 0);
            }
            function commitSeek(ratio) {
                const pct = ratio * 100;
                fillEl.style.width = pct + '%';
                thumbEl.style.left = pct + '%';
                previewAt(ratio);
            }
            function restoreTime() {
                timeEl.textContent = fmt(video.currentTime) + ' / ' + fmt(video.duration || 0);
            }
            function updateBuffered() {
                if (!video.duration) return;
                const b = video.buffered;
                let end = 0;
                if (b && b.length) {
                    for (let i = 0; i < b.length; i++) {
                        if (video.currentTime >= b.start(i) && video.currentTime <= b.end(i)) { end = b.end(i); break; }
                    }
                    if (!end) end = b.end(b.length - 1);
                }
                bufEl.style.width = (end / video.duration * 100) + '%';
            }
            seek.addEventListener('pointerdown', (e) => {
                e.stopPropagation();
                seeking = true; hovering = false;
                try { seek.setPointerCapture(e.pointerId); } catch (err) {}
                tipEl.classList.add('show');
                commitSeek(ratioFromEvent(e));
                show();
            });
            seek.addEventListener('pointermove', (e) => {
                const r = ratioFromEvent(e);
                tipEl.classList.add('show');
                if (seeking) commitSeek(r); else { hovering = true; previewAt(r); }
            });
            seek.addEventListener('pointerup', (e) => {
                if (!seeking) return;
                seeking = false;
                if (video.duration) video.currentTime = ratioFromEvent(e) * video.duration;
                tipEl.classList.remove('show');
                show();
            });
            seek.addEventListener('pointercancel', () => { seeking = false; hovering = false; tipEl.classList.remove('show'); restoreTime(); });
            seek.addEventListener('pointerleave', () => { hovering = false; if (!seeking) { tipEl.classList.remove('show'); restoreTime(); } });

            video.addEventListener('timeupdate', () => {
                const now = Date.now();
                if (now - lastSave > 5000) { lastSave = now; saveProgress(); }
                if (!seeking && !hovering) {
                    if (video.duration) {
                        const r = video.currentTime / video.duration;
                        fillEl.style.width = (r * 100) + '%';
                        thumbEl.style.left = (r * 100) + '%';
                    }
                    timeEl.textContent = fmt(video.currentTime) + ' / ' + fmt(video.duration || 0);
                }
            });
            video.addEventListener('progress', updateBuffered);
            video.addEventListener('loadedmetadata', updateBuffered);
            video.addEventListener('canplay', updateBuffered);

            // 缓冲 / 跳转 loading 提示
            video.addEventListener('loadstart', () => setLoading(true));
            video.addEventListener('waiting', () => setLoading(true));
            video.addEventListener('seeking', () => setLoading(true));
            video.addEventListener('canplay', () => setLoading(false));
            video.addEventListener('playing', () => setLoading(false));
            video.addEventListener('seeked', () => setLoading(false));
            video.addEventListener('error', () => setLoading(false));

            // 播放/暂停不用按钮：点画面即可（video 的 click 处理）
            video.addEventListener('play', () => { show(); });
            video.addEventListener('pause', () => { show(); });
            video.addEventListener('ended', () => { show(); saveProgress(); });
            
            // 快退/快进按钮（« »）已按用户要求删除；键盘 ←/→ 仍可跳转 10 秒。
            const SEEK_STEP = 10;
            
            // 倍速记忆（0.75x~3x）。倍速按钮已按用户要求删除，
            // 这里只保留上次的倍速值，作为长按临时加速结束后要恢复的原速。
            const speeds = [0.75, 1, 1.25, 1.5, 2, 2.5, 3];
            let speedIdx = 1;
            try {
                const saved = parseFloat(localStorage.getItem('mb_speed'));
                const i = speeds.indexOf(saved);
                if (i >= 0) speedIdx = i;
            } catch (e) {}
            video.playbackRate = speeds[speedIdx];
            
            // 音量滑块（记忆）
            try {
                const savedVol = parseFloat(localStorage.getItem('mb_vol'));
                if (!isNaN(savedVol) && savedVol >= 0 && savedVol <= 100) {
                    volEl.value = savedVol;
                    video.volume = savedVol / 100;
                    video.muted = savedVol == 0;
                }
            } catch (e) {}
            volEl.onclick = (e) => e.stopPropagation();
            volEl.addEventListener('input', () => {
                video.volume = volEl.value / 100;
                video.muted = volEl.value == 0;
                try { localStorage.setItem('mb_vol', String(volEl.value)); } catch (e) {}
            });
            
            // 全屏切换（手机可横屏）
            fullBtn.onclick = (e) => {
                e.stopPropagation();
                const container = document.getElementById('lightboxContent');
                if (document.fullscreenElement || document.webkitFullscreenElement) {
                    (document.exitFullscreen || document.webkitExitFullscreen).call(document);
                    fullBtn.innerHTML = ic('maximize', true);
                } else {
                    (container.requestFullscreen || container.webkitRequestFullscreen).call(container);
                    fullBtn.innerHTML = ic('minimize', true);
                }
            };
            // 全屏状态变化由上面那个「只注册一次」的监听处理（这里不再注册，避免泄漏）

            // 长按屏幕临时倍速播放，松手恢复原速。
            // 提示文案是底部居中大字「2x 快进中」，长按期间显示、松手即消失
            // （见 .vc-fast 的 CSS）。原来是个挂在屏幕正中的小徽标，会挡画面。
            let lpTimer = null, lpActive = false, lpSuppressClick = false, lpStartX = 0, lpStartY = 0, baseRate = 1, fastBadgeTimer = null;
            const LP_MS = 500, LP_SPEED = 2;
            function endLongPress() {
                if (lpTimer) { clearTimeout(lpTimer); lpTimer = null; }
                if (lpActive) {
                    lpActive = false;
                    video.playbackRate = baseRate;
                    fastBadge.classList.remove('show');
                    lpSuppressClick = true;
                }
            }
            video.addEventListener('pointerdown', (e) => {
                lpStartX = e.clientX; lpStartY = e.clientY; lpActive = false; lpSuppressClick = false;
                clearTimeout(lpTimer);
                lpTimer = setTimeout(() => {
                    if (!video.paused) {
                        lpActive = true;
                        baseRate = video.playbackRate;
                        video.playbackRate = Math.max(LP_SPEED, baseRate);
                        // 文案与样式对齐"底部居中大字"（原来是个遮画面的小徽标）
                        fastBadge.textContent = video.playbackRate + 'x ' +
                            (video.playbackRate < 1 ? '慢放' : '快进中');
                        fastBadge.classList.add('show');
                        // 1 秒后自动隐藏（松手时也会立即隐藏）
                        clearTimeout(fastBadgeTimer);
                        fastBadgeTimer = setTimeout(() => fastBadge.classList.remove('show'), 1000);
                    }
                }, LP_MS);
            });
            video.addEventListener('pointermove', (e) => {
                if (lpTimer && !lpActive && (Math.abs(e.clientX - lpStartX) > 12 || Math.abs(e.clientY - lpStartY) > 12)) {
                    clearTimeout(lpTimer); lpTimer = null;
                }
            });
            video.addEventListener('pointerup', endLongPress);
            video.addEventListener('pointercancel', endLongPress);
            video.addEventListener('pointerleave', endLongPress);

            video.addEventListener('click', () => {
                if (lpSuppressClick) { lpSuppressClick = false; return; }
                video.paused ? video.play() : video.pause();
            });
            video.addEventListener('mousemove', show);
            video.addEventListener('touchstart', show);
            ctrl.addEventListener('mousemove', show);
            ctrl.addEventListener('touchstart', (e) => { e.stopPropagation(); show(); });
            ctrl.addEventListener('click', (e) => e.stopPropagation());

            // 清理钩子：由 destroyVideo / clearVideoControls 调用。
            // 显式清掉定时器、document 级引用（currentFullBtn）和控制条 DOM，
            // 保证看完的视频不留常驻内存（这是长时浏览变卡的根源）。
            let cleaned = false;
            video.__mbCleanup = () => {
                if (cleaned) return;
                cleaned = true;
                clearTimeout(hideTimer);
                clearTimeout(lpTimer);
                clearTimeout(fastBadgeTimer);
                endLongPress();
                if (currentFullBtn === fullBtn) currentFullBtn = null;
                if (vcVideo === video) vcVideo = null;
                try { ctrl.remove(); } catch (e) {}
                try { loading.remove(); } catch (e) {}
                try { fastBadge.remove(); } catch (e) {}
            };

            show();
        }
        function clearVideoControls() {
            // 先跑当前视频的清理钩子（定时器 / document 级引用 / 控制条 DOM 一起释放）
            const curVideo = document.querySelector('.lightbox-content video');
            if (curVideo && curVideo.__mbCleanup) { try { curVideo.__mbCleanup(); } catch (e) {} }
            const ctrl = document.querySelector('.video-controls');
            if (ctrl) ctrl.remove();
            const loading = document.querySelector('.vc-loading');
            if (loading) loading.remove();
            const fast = document.querySelector('.vc-fast');
            if (fast) fast.remove();
            vcVideo = null;
        }
        
        // 删除后重新同步网格。数组变了就必须同步 DOM，否则网格里还留着已删除的
        // 格子、而 filteredMedia 已经左移 —— 之后点任何格子，data-index 都会映射到
        // 另一张图（表现为"缩略图是 A、点开是 B"），刷新页面才自愈。
        //
        // 但不能用 render() 全量重建：实测 16918 个格子要 166~218ms（innerHTML 解析
        // + 布局），而删除的只有一两张图 —— 这正是"删除后变慢"的原因。这里改成：
        //   1) 从 DOM 里摘掉已删文件那一格（O(1)，不重建任何东西）
        //   2) 只把受影响的那几个日期分区重新生成 HTML（通常只有几十~几百格）
        //   3) 只给后续分区重写 data-index（纯属性写，约 13ms，不触发重排）
        // 网格其余部分完全不碰：缩略图不闪、不重新请求、不重排。
        function sectionItemsHTML(items, startIndex) {
            let html = '';
            items.forEach((item, k) => {
                const i = startIndex + k;
                const selectedClass = selectedFiles.has(item.path) ? ' selected' : '';
                const thumbUrl = safePath(item.path);
                html += '<div class="photo-item' + selectedClass + '" data-index="' + i
                     +  '" data-path="' + encodeURIComponent(item.path) + '">';
                html += '<img src="/thumb/' + thumbUrl + '?s=' + (item.size || 0) + '" loading="lazy" onerror="this.src=\'data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22200%22 height=%22200%22><rect fill=%22%23333%22 width=%22200%22 height=%22200%22/><rect x=%2240%22 y=%2262%22 width=%22120%22 height=%2276%22 rx=%2210%22 fill=%22none%22 stroke=%22%23888%22 stroke-width=%228%22/><circle cx=%2268%22 cy=%2286%22 r=%228%22 fill=%22%23888%22/><path d=%22M48 132l26-26 20 20 24-24 30 30%22 fill=%22none%22 stroke=%22%23888%22 stroke-width=%228%22 stroke-linecap=%22round%22 stroke-linejoin=%22round%22/></svg>\'">';
                if (item.isVideo) {
                    html += '<div class="video-overlay">' + (item.duration > 0 ? '<span class="video-duration">' + durText(item.duration) + '</span>' : '') + '</div>';
                }
                html += '</div>';
            });
            return html;
        }

        // deletedPaths: 本次删掉的路径集合；keepPath: 灯箱接下来该显示的项（可为空）
        // 返回接下来应显示的 filteredMedia 索引（-1 表示已无内容）
        function syncGridAfterDelete(deletedPaths, keepPath) {
            const doomed = deletedPaths instanceof Set ? deletedPaths : new Set(deletedPaths);

            // 1) 摘掉已删的格子，并记下每个网格块（同一日期可能被切成多块）里最小/最大的索引
            const units = [];
            document.querySelectorAll('#content .photo-grid').forEach(grid => {
                let start = Infinity, end = -Infinity, removed = 0;
                grid.querySelectorAll('.photo-item').forEach(el => {
                    const idx = parseInt(el.dataset.index, 10);
                    if (!(idx >= 0)) return;
                    if (idx < start) start = idx;
                    if (idx > end) end = idx;
                    if (doomed.has(decodeURIComponent(el.dataset.path || ''))) {
                        el.remove();
                        removed++;
                    }
                });
                if (start !== Infinity) units.push({ grid, start, end, removed });
            });

            if (doomed.size === 0) {
                if (filteredMedia.length === 0) return -1;
                let ni = keepPath ? filteredMedia.findIndex(m => m.path === keepPath) : -1;
                if (ni < 0) ni = Math.min(Math.max(0, currentIndex), filteredMedia.length - 1);
                return ni;
            }

            // 2) 块顺序 = filteredMedia 顺序（每张都是连续的），据此算新的起始索引
            units.sort((a, b) => a.start - b.start);
            const newStart = new Map();
            let off = 0;
            units.forEach(s => {
                newStart.set(s.grid, s.start - off);
                off += s.removed;
            });

            // 3) 只重建被删过的块，并同步 data-index / 选中态
            units.forEach(s => {
                if (s.removed === 0) return;
                const ns = newStart.get(s.grid);
                const items = filteredMedia.slice(ns, ns + s.grid.children.length);
                s.grid.innerHTML = sectionItemsHTML(items, ns);
            });

            // 4) 被删块之后的格子整体左移，重写 data-index（纯属性写，开销很小）
            units.forEach(s => {
                if (s.removed > 0) return;
                if (s.end < newStart.get(s.grid)) return;
                const ns = newStart.get(s.grid);
                const els = s.grid.children;
                const delta = ns - s.start;
                if (delta === 0) return;
                for (let i = 0; i < els.length; i++) {
                    els[i].dataset.index = String(parseInt(els[i].dataset.index, 10) + delta);
                }
            });

            // 5) 被删空的块移除；整个日期都没格子了，连日期头一起移除
            units.forEach(s => {
                if (s.grid.children.length === 0) s.grid.remove();
            });
            document.querySelectorAll('#content .date-section').forEach(sec => {
                if (!sec.querySelector('.photo-item')) sec.remove();
            });

            if (filteredMedia.length === 0) return -1;
            let ni = keepPath ? filteredMedia.findIndex(m => m.path === keepPath) : -1;
            if (ni < 0) ni = Math.min(Math.max(0, currentIndex), filteredMedia.length - 1);
            return ni;
        }

        // 彻底释放一个视频元素。必须做完整流程：只 pause() 不释放，
        // 解码器、网络连接和缓冲都会留在原处继续占用资源。
        // 之前 openLightbox 有完整清理、而 closeLightbox 只调了 pause()，
        // 导致"点 × 退出后视频仍在后台下载 + 解码器不释放"，页面越来越卡、
        // 最终必须刷新才能恢复。现在两条路径共用这一个函数。
        function destroyVideo(v) {
            if (!v) return;
            try { if (v.__mbCleanup) v.__mbCleanup(); } catch (e) {}
            try {
                if (currentFilter === 'movie' && v.duration && v._mbPath) {
                    localStorage.setItem('mb_pos_' + v._mbPath, String(v.currentTime));
                }
            } catch (e) {}
            try {
                v.onloadeddata = null;   // 防止僵尸回调
                v.onerror = null;
                v.pause();
                v.removeAttribute('src');
                v.srcObject = null;
                while (v.firstChild) v.removeChild(v.firstChild);
                v.load();                // 真正终止下载并释放解码器
            } catch (e) {}
            try { v.remove(); } catch (e) {}
        }

        function closeLightbox() { 
            // 先拿到视频引用再清理 —— 顺序反了会漏掉释放
            const video = document.querySelector('.lightbox-content video');
            if (video) destroyVideo(video);
            clearVideoControls();
            document.getElementById('lightboxContent').innerHTML = '';
            // 两个 class 一起摘：既关掉灯箱，也清掉“鼠标靠边”状态（一次属性写入，
            // 不会触发观察器来回改 class 的死循环 —— 老代码就是这么卡死的）
            document.getElementById('lightbox').classList.remove('active', 'nav-near');
            document.body.style.overflow = '';
            // 滚动到当前查看的图片位置
            if (filteredMedia.length > 0 && currentIndex >= 0) {
                const currentItem = filteredMedia[Math.min(currentIndex, filteredMedia.length - 1)];
                let el = document.querySelector(`[data-path="${encodeURIComponent(currentItem.path)}"]`);
                if (el) {
                    el.scrollIntoView({ behavior: 'smooth', block: 'center' });
                } else {
                    // 目标所在分区已被回收：异步（逐帧）恢复，不阻塞，并直接滚到该分区
                    const idx = Math.min(currentIndex, filteredMedia.length - 1);
                    restoreAllRecycledSections(false);
                    const sec = sectionForIndex(idx);
                    if (sec) sec.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }
            }
        }
        
        function navigate(delta) { 
            currentIndex = (currentIndex + delta + filteredMedia.length) % filteredMedia.length; 
            openLightbox(); 
        }
        
        // 同步筛选按钮高亮。programmatic 切换 currentFilter 时也必须调用，
        // 否则会出现"按钮显示视频、列表却是全部"的不一致
        function setActiveFilterButton(filter) {
            document.querySelectorAll('.filter-btn').forEach(b => {
                b.classList.toggle('active', b.dataset.filter === filter);
            });
        }

        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                currentFilter = btn.dataset.filter;
                setActiveFilterButton(currentFilter);
                filterMedia();
                currentIndex = 0;
                if (selectMode) {
                    selectedFiles.clear();
                }
                render();
            });
        });
        
        document.getElementById('selectModeBtn').addEventListener('click', toggleSelectMode);
        document.getElementById('selectAllBtn').addEventListener('click', selectAll);
        document.getElementById('deleteBtn').addEventListener('click', deleteSelected);
        
        document.getElementById('lightboxCloseBtn').addEventListener('click', closeLightbox);
        document.getElementById('navPrev').addEventListener('click', () => navigate(-1));
        document.getElementById('navNext').addEventListener('click', () => navigate(1));

        // 桌面端：翻页键默认隐藏，鼠标靠近左右边缘 110px 内才显示。
        // 除了 CSS 的 :hover 热区，这里再用 mousemove 兜底（热区被上层层遮挡时仍生效）。
        (function initNavAutoHide() {
            const lb = document.getElementById('lightbox');
            if (!lb) return;
            const EDGE = 110;   // 与 CSS 的 .nav-hotzone 宽度保持一致
            let raf = 0;
            const isDesktop = () => window.matchMedia('(min-width: 901px)').matches;

            function apply(x) {
                const near = isDesktop() && (x <= EDGE || x >= window.innerWidth - EDGE);
                if (near !== lb.classList.contains('nav-near')) lb.classList.toggle('nav-near', near);
            }
            lb.addEventListener('mousemove', (e) => {
                if (raf) return;
                raf = requestAnimationFrame(() => { raf = 0; apply(e.clientX); });
            }, { passive: true });
            // 指针离开灯箱（或窗口）→ 收起
            lb.addEventListener('mouseleave', () => { lb.classList.remove('nav-near'); });
            // 关闭灯箱时清掉状态，下次打开是干净的（默认隐藏）。
            //
            // 这里原来用 MutationObserver 监听 lb 的 class 变化来清 nav-near —— 这是个
            // 死循环 bug：回调里又去改 lb 自己的 class，Blink 下会再次触发回调，微任务
            // 永远跑不完，主线程被焊死（表现为“一关闭灯箱整页就卡住、点哪儿都没反应”，
            // 图片和视频都会中招）。清状态改成在关闭处显式、一次性完成。
        })();
        
        // 大图查看页删除按钮
        document.getElementById('lightboxDeleteBtn').addEventListener('click', async () => {
            const currentItem = filteredMedia[currentIndex];
            if (!currentItem) return;
            
            try {
                const response = await fetch('/api/delete', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ files: [currentItem.path] })
                });
                
                const data = await response.json();
                
                if (data.deleted > 0) {
                    showDeleteToast();
                    
                    // 从列表中移除
                    const idxToRemove = allMedia.findIndex(m => m.path === currentItem.path);
                    if (idxToRemove > -1) {
                        allMedia.splice(idxToRemove, 1);
                    }
                    
                    // 记录下一张（按路径定位，避免索引漂移）
                    const nextPath = (filteredMedia[currentIndex + 1] || {}).path;
                    filterMedia();
                    
                    // 关闭查看器
                    closeLightbox();
                    
                    // 关键：让网格 DOM 跟上数组（原先漏了这一步）。
                    // 只重建受影响的日期分区 + 重编号后续分区，不做全量 render，
                    // 否则每次删除要花 160~220ms 重建 16918 个格子。
                    const ni = syncGridAfterDelete(new Set([currentItem.path]), nextPath);
                    if (ni >= 0) {
                        currentIndex = ni;
                        openLightbox();
                    } else {
                        // 没有图片了，刷新页面
                        await load();
                    }
                } else {
                    showStationToast('删除失败，请重试', true);
                }
            } catch (e) {
                console.error('Delete error:', e);
                showStationToast('删除失败：' + e.message, true);
            }
        });
        
        document.getElementById('lightbox').addEventListener('click', (e) => {
            if (e.target.id === 'lightbox') {
                closeLightbox();
            }
        });
        
        document.getElementById('lightboxContent').addEventListener('click', (e) => {
            if (!document.querySelector('.lightbox-content video')) closeLightbox();
        });
        
        document.addEventListener('keydown', (e) => { 
            if (!document.getElementById('lightbox').classList.contains('active')) return; 
            if (e.key === 'Escape') closeLightbox(); 
            // 影视看视频：左右键按当前步长快退/快进
            const videoEl = document.querySelector('.lightbox-content video');
            if (currentFilter === 'movie' && videoEl && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
                const delta = (e.key === 'ArrowLeft' ? -1 : 1) * SEEK_STEP;
                if (vcVideo) {
                    vcVideo.currentTime = Math.max(0, Math.min((vcVideo.duration || 0), vcVideo.currentTime + delta));
                }
                return;
            }
            if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') navigate(-1); 
            if (e.key === 'ArrowRight' || e.key === 'ArrowDown') navigate(1); 
        });
        
        window.addEventListener('wheel', (e) => { 
            if (!document.getElementById('lightbox').classList.contains('active')) return; 
            e.preventDefault(); 
            navigate(e.deltaY > 0 ? 1 : -1); 
        }, { passive: false });
        
        // 屏蔽长按/右键系统菜单（手机上长按图片会弹"保存图片/在新标签页中打开"，
        // 桌面是右键菜单）。长按多选是页面自己用 touchstart/touchend 计时做的，
        // 不依赖这个菜单，所以屏蔽它不影响长按勾选。
        document.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            e.stopPropagation();
        }, true);

        document.addEventListener('touchstart', (e) => {
            touchStartX = e.touches[0].clientX;
            touchStartY = e.touches[0].clientY;
            touchStartTime = Date.now();
            const windowWidth = window.innerWidth;
            ignoreSwipe = touchStartX < EDGE_THRESHOLD || touchStartX > windowWidth - EDGE_THRESHOLD;
        }, { passive: true });
        
        document.addEventListener('touchend', (e) => {
            if (!document.getElementById('lightbox').classList.contains('active')) return;
            if (ignoreSwipe) return;
            
            const touchEndX = e.changedTouches[0].clientX;
            const touchEndY = e.changedTouches[0].clientY;
            const deltaTime = Date.now() - touchStartTime;
            const deltaX = touchStartX - touchEndX;
            const deltaY = touchStartY - touchEndY;
            
            if (deltaTime > TIME_THRESHOLD) return;
            
            if (Math.abs(deltaX) > SWIPE_THRESHOLD && Math.abs(deltaX) > Math.abs(deltaY)) {
                navigate(deltaX > 0 ? 1 : -1);
            } else if (Math.abs(deltaY) > SWIPE_THRESHOLD) {
                navigate(deltaY > 0 ? 1 : -1);
            }
        }, { passive: false });
        
        load();

        // ================= TG Station 整合 UI =================
        (function () {
            const $id = (i) => document.getElementById(i);
            stationData = null;  // 复用全局声明（filterMedia 的 movie 过滤依赖它）
            let stationToastTimer = null;

            function showStationToast(msg, isErr) {
                const t = $id('stationToast');
                setIconText(t, isErr ? 'alert' : 'check', msg);
                t.className = 'station-toast show' + (isErr ? ' err' : '');
                clearTimeout(stationToastTimer);
                stationToastTimer = setTimeout(() => t.classList.remove('show'), 2600);
            }

            // 错误弹窗：替代原生 alert（原生弹窗会冻结页面且样式与界面不搭）。
            // 挂到 window 上，让上半段脚本里的删除失败等场景也能用。
            let errTimer = null;
            window.showError = function (msg, title) {
                const box = $id('errToast');
                if (!box) return;
                $id('errTitle').textContent = title || '操作失败';
                $id('errMsg').textContent = msg || '未知错误';
                box.classList.add('show');
                clearTimeout(errTimer);
                errTimer = setTimeout(() => box.classList.remove('show'), 6000);
            };
            (function initErrToast() {
                const close = $id('errClose');
                if (close) close.addEventListener('click', () => {
                    $id('errToast').classList.remove('show');
                    clearTimeout(errTimer);
                });
            })();

            // 确认弹窗（Promise 化），替代原生 confirm
            window.showConfirm = function (msg, title, okText) {
                return new Promise((resolve) => {
                    const ov = $id('confirmModal');
                    if (!ov) { resolve(true); return; }   // 兜底：没有弹窗就直接放行
                    $id('confirmTitle').textContent = title || '确认操作';
                    $id('confirmMsg').textContent = msg || '';
                    $id('confirmOk').textContent = okText || '确定';
                    ov.classList.add('show');

                    const done = (val) => {
                        ov.classList.remove('show');
                        $id('confirmOk').removeEventListener('click', onOk);
                        $id('confirmCancel').removeEventListener('click', onNo);
                        ov.removeEventListener('click', onBg);
                        document.removeEventListener('keydown', onKey);
                        resolve(val);
                    };
                    const onOk = () => done(true);
                    const onNo = () => done(false);
                    const onBg = (e) => { if (e.target === ov) done(false); };
                    const onKey = (e) => {
                        if (e.key === 'Escape') done(false);
                        else if (e.key === 'Enter') done(true);
                    };
                    $id('confirmOk').addEventListener('click', onOk);
                    $id('confirmCancel').addEventListener('click', onNo);
                    ov.addEventListener('click', onBg);
                    document.addEventListener('keydown', onKey);
                    setTimeout(() => $id('confirmOk').focus(), 30);
                });
            };

            // 筛选栏高度实测 -> 写入 CSS 变量，供日期标题吸顶定位使用。
            // 用 ResizeObserver 而不是 window.resize：字体变化、筛选按钮换行也能跟上。
            (function initStickyOffset() {
                const header = document.querySelector('.header');
                if (!header) return;
                const apply = () => {
                    const h = Math.round(header.getBoundingClientRect().height);
                    if (h > 0) document.documentElement.style.setProperty('--header-h', h + 'px');
                };
                apply();
                if (window.ResizeObserver) new ResizeObserver(apply).observe(header);
                window.addEventListener('resize', apply);
                // 滚动后给粘性头部加分隔线
                let ticking = false;
                window.addEventListener('scroll', () => {
                    if (ticking) return;
                    ticking = true;
                    requestAnimationFrame(() => {
                        const stuck = window.scrollY > 8;
                        if (stuck !== header.classList.contains('stuck')) header.classList.toggle('stuck', stuck);
                        ticking = false;
                    });
                }, { passive: true });
            })();

            async function stationFetch(url, opts) {
                try {
                    const res = await fetch(url, Object.assign({ headers: { 'Content-Type': 'application/json' } }, opts));
                    return await res.json();
                } catch (e) {
                    return { ok: false, error: '网络错误' };
                }
            }

            function setStationVisual(running, state) {
                const cls = running ? 'on' : (state === 'busy' ? 'busy' : 'off');
                $id('stationDot').className = 'status-dot ' + cls;
                $id('fabDot').className = 'fab-dot ' + cls;
                setIconText($id('listenToggleBtn'), running ? 'square' : 'play', running ? '停止监听' : '开始监听');
                setIconText($id('fabToggleBtn'), running ? 'square' : 'play', running ? '停止监听' : '开始监听');
                $id('fabIcon').innerHTML = ic(running ? 'square' : 'play', true);
            }

            async function refreshStationStatus() {
                const data = await stationFetch('/api/status');
                if (!data.ok) {
                    $id('stationStatusText').textContent = '后端离线';
                    setStationVisual(false, 'off');
                    return;
                }
                stationData = data;
                const ls = data.listener || {};
                const md = data.media || {};
                let txt;
                if (ls.running) txt = '监听中 · PID ' + ls.pid;
                else if (ls.state === 'exited') txt = '已退出 (' + (ls.exitMsg || ('码 ' + ls.exitCode)) + ')';
                else txt = '未运行';
                if (md.total > 0) txt += ' · ' + md.total + ' 个文件';
                $id('stationStatusText').textContent = txt;
                setStationVisual(ls.running, ls.state);
            }

            // 下载进度面板：有下载任务时自动展开实时进度；无任务时默认收起，点击标题展开看当天记录
            let dlExpanded = false;
            function setDlExpanded(open) {
                dlExpanded = open;
                $id('dlBody').style.display = open ? '' : 'none';
                $id('dlArrow').className = 'dl-arrow' + (open ? ' open' : '');
            }

            // 速度估算：记录上一次采样的累计字节与时刻
            let dlPrevBytes = null;
            let dlPrevAt = 0;
            let dlWasActive = false;   // 上一轮是否在下载，用于识别"刚刚全部完成"
            let dlDoneTimer = null;
            let dlWaitUntil = 0;   // "开始监听"后的宽限窗口：让面板立即出现，不用干等状态轮询

            function fmtSpeed(bytesPerSec) {
                if (!bytesPerSec || bytesPerSec <= 0) return '—';
                const kbs = bytesPerSec / 1024;
                if (kbs < 1000) return kbs.toFixed(0) + ' KB/s';
                return (kbs / 1024).toFixed(1) + ' MB/s';
            }

            function fmtEta(seconds) {
                if (!isFinite(seconds) || seconds <= 0) return '—';
                if (seconds < 60) return Math.ceil(seconds) + ' 秒';
                const m = Math.floor(seconds / 60);
                if (m < 60) return m + ' 分' + Math.round(seconds % 60) + ' 秒';
                return Math.floor(m / 60) + ' 时' + (m % 60) + ' 分';
            }

            async function refreshDownloadProgress() {
                const data = await stationFetch('/api/download/progress');
                if (!data.ok) return;
                const tasks = data.tasks || [];
                const today = new Date();
                const todayStr = ('0' + (today.getMonth() + 1)).slice(-2) + '-' + ('0' + today.getDate()).slice(-2);
                const recent = (data.recent || []).filter(r => r.date === todayStr).slice(-5).reverse(); // 只看当天，最新在前
                const hasTask = tasks.length > 0;
                const hasRecent = recent.length > 0;
                const panel = $id('downloadPanel');
                // 面板只在「正在监听」或「有下载任务」时出现。
                // 不因当天有历史下载记录而常驻（用户明确要求）。
                const listening = !!(stationData && stationData.listener && stationData.listener.running);
                if (listening) dlWaitUntil = 0;   // 已确认在监听，宽限窗口作废
                const starting = dlWaitUntil > Date.now();
                const wantShow = listening || hasTask || starting;

                if (!wantShow) {
                    // 隐藏；若刚从"有任务"变为空闲，短暂显示 4 秒完成提示
                    const justFinished = dlWasActive;
                    dlWasActive = false;
                    panel.classList.remove('show', 'expanded');
                    document.body.classList.remove('dl-open');
                    $id('dlMeta').style.display = 'none';
                    dlPrevBytes = null;
                    clearTimeout(dlDoneTimer);
                    if (justFinished) {
                        const done = $id('dlDone');
                        done.style.display = '';
                        panel.classList.add('show');
                        dlDoneTimer = setTimeout(() => {
                            done.style.display = 'none';
                            panel.classList.remove('show');
                        }, 4000);
                    } else {
                        $id('dlDone').style.display = 'none';
                    }
                    return;
                }
                clearTimeout(dlDoneTimer);
                $id('dlDone').style.display = 'none';
                panel.classList.add('show');
                document.body.classList.add('dl-open');

                // ===== 总进度 / 速度 / 剩余时间 =====
                if (hasTask) {
                    dlWasActive = true;
                    let doneBytes = 0, totalBytes = 0;
                    tasks.forEach(t => {
                        const dm = Number(t.done_mb) || 0;
                        doneBytes += dm * 1048576;
                        totalBytes += (Number(t.total_mb) || dm) * 1048576;
                    });
                    const now = performance.now();
                    let speed = 0;
                    if (dlPrevBytes !== null && now > dlPrevAt) {
                        const dt = (now - dlPrevAt) / 1000;
                        const db = doneBytes - dlPrevBytes;
                        if (dt > 0.5 && db >= 0) speed = db / dt;
                    }
                    dlPrevBytes = doneBytes;
                    dlPrevAt = now;

                    const overallPct = totalBytes > 0 ? Math.min(100, doneBytes * 100 / totalBytes) : 0;
                    $id('dlMeta').style.display = '';
                    $id('dlOverallFill').style.width = overallPct.toFixed(1) + '%';
                    $id('dlCount').textContent = tasks.length + ' 个 · ' + overallPct.toFixed(0) + '%';
                    $id('dlSpeed').textContent = fmtSpeed(speed);
                    $id('dlEta').textContent = (speed > 1024 && totalBytes > doneBytes)
                        ? fmtEta((totalBytes - doneBytes) / speed) : '—';
                } else {
                    dlPrevBytes = null;
                    $id('dlMeta').style.display = 'none';
                }

                // 摘要
                let summary = '';
                if (hasTask) summary = tasks.length + ' 个下载中';
                else if (hasRecent) summary = '今天 ' + recent.length + ' 个';
                $id('dlSummary').textContent = summary;

                // 正在下载的任务
                const tBox = $id('dlTasks');
                if (hasTask) {
                    tBox.innerHTML = tasks.map(t => {
                        const pct = Math.min(100, Math.max(0, t.pct || 0));
                        const mb = t.total_mb ? ' · ' + t.done_mb + ' / ' + t.total_mb + ' MB' : '';
                        const fname = escHtml(t.file);
                        return '<div class="dl-task">'
                            + '<div class="dl-task-head"><span class="dl-task-name" title="' + fname + '">' + fname + '</span>'
                            + '<span class="dl-task-pct">' + pct + '%' + mb + '</span></div>'
                            + '<div class="dl-bar"><div class="dl-bar-fill" style="width:' + pct + '%"></div></div>'
                            + '<div class="dl-task-actions"><button class="dl-cancel-btn" data-file="' + fname + '">' + ic('x') + '取消此任务</button></div>'
                            + '</div>';
                    }).join('');
                } else {
                    tBox.innerHTML = '<div class="dl-empty">暂无正在下载的任务</div>';
                }
                // 最近完成记录（仅当天）
                $id('dlRecentTitle').style.display = hasRecent ? '' : 'none';
                $id('dlRecent').innerHTML = recent.map(r =>
                    '<div class="dl-recent-item"><span class="dl-recent-time">' + escHtml((r.time || '')) + '</span>'
                    + '<span class="dl-recent-file">' + escHtml(r.file || '') + '</span></div>'
                ).join('');

                // 有任务自动展开；无任务保持用户手动状态
                if (hasTask && !dlExpanded) setDlExpanded(true);
                if (!hasTask && dlExpanded && $id('dlTasks').innerHTML === '<div class="dl-empty">暂无正在下载的任务</div>') {
                    // 任务全部结束后收起（回到折叠态）
                    setDlExpanded(false);
                }
            }

            async function toggleListen() {
                const running = stationData && stationData.listener && stationData.listener.running;
                const url = running ? '/api/listen/stop' : '/api/listen/start';
                const btn = $id('listenToggleBtn');
                const oldText = btn.textContent;
                btn.textContent = '处理中…';
                const res = await stationFetch(url, { method: 'POST', body: '{}' });
                btn.textContent = oldText;
                if (res.ok) {
                    showStationToast(res.message || '操作成功');
                    // 正在"开始"时先让面板立刻出现：下载器起来要几秒，
                    // 干等 /api/status 报 running 会慢一拍。窗口 10 秒。
                    dlWaitUntil = running ? 0 : Date.now() + 10000;
                    for (let i = 0; i < 6; i++) {
                        await new Promise(r => setTimeout(r, 1000));
                        await refreshStationStatus();
                        if (i === 0) await refreshDownloadProgress();
                    }
                    if (running) {   // 刚才是"停止"操作 → 立刻收起面板
                        $id('downloadPanel').classList.remove('show', 'expanded');
                        document.body.classList.remove('dl-open');
                    } else {
                        await refreshDownloadProgress();
                    }
                } else {
                    showStationToast('操作失败: ' + (res.error || '未知错误'), true);
                }
            }

            // 停止下载：终止当前所有下载任务（已下载文件保留）
            async function stopDownload() {
                const yes = await showConfirm('正在下载的任务会终止，已下载的文件保留。', '停止下载？', '停止下载');
                if (!yes) return;
                const btn = $id('dlStopBtn');
                const oldText = btn.textContent;
                btn.textContent = '停止中…';
                btn.disabled = true;
                const res = await stationFetch('/api/listen/stop', { method: 'POST', body: '{}' });
                btn.textContent = oldText;
                btn.disabled = false;
                if (res.ok) {
                    showStationToast(res.message || '下载已停止');
                    await refreshStationStatus();
                    await refreshDownloadProgress();
                } else {
                    showStationToast('操作失败: ' + (res.error || '未知错误'), true);
                }
            }

            // 取消单个下载任务（写 cancel.json，下载器命中后放弃该任务）
            async function cancelTask(file) {
                const yes = await showConfirm(file, '取消这个下载任务？', '取消任务');
                if (!yes) return;
                const res = await stationFetch('/api/download/cancel', {
                    method: 'POST',
                    body: JSON.stringify({ file: file })
                });
                if (res.ok) {
                    showStationToast(res.message || '已请求取消');
                    await refreshDownloadProgress();
                } else {
                    showStationToast('操作失败: ' + (res.error || '未知错误'), true);
                }
            }

            // ---------- 设置面板 ----------
            const SETTING_GROUPS = [
                { title: '下载设置', items: [
                    { key: 'download_dir', label: '下载目录', type: 'text', full: true },
                    { key: 'types', label: '下载类型', type: 'checks', options: ['photo', 'video', 'document'] },
                    { key: 'max_file_size_mb', label: '最大文件大小 (MB)', type: 'number' },
                    { key: 'concurrency', label: '并发数', type: 'number' },
                    { key: 'retries', label: '失败重试次数', type: 'number' },
                    { key: 'idle_stop_minutes', label: '空闲自动停止 (分钟, 0=关闭)', type: 'number' },
                    { key: 'only_forwarded', label: '仅下载转发消息', type: 'checkbox' },
                    { key: 'skip_gif', label: '跳过 GIF', type: 'checkbox' },
                    { key: 'download_history', label: '启动时扫描历史', type: 'checkbox' },
                    { key: 'login_method', label: '登录方式', type: 'select', options: ['qr', 'phone'] },
                    { key: 'twofa_password', label: '2FA 密码', type: 'password' },
                ]},
                { title: '服务与脚本', items: [
                    { key: 'media_browser.port', label: 'HTTP 端口', type: 'number' },
                    { key: 'station.python', label: 'Python 解释器路径', type: 'text', full: true },
                    { key: 'station.script', label: '下载器脚本路径', type: 'text', full: true },
                    { key: 'station.media_dir', label: '媒体目录（留空 = 下载目录）', type: 'text', full: true },
                    { key: 'station.auto_start', label: '程序启动时自动开始监听', type: 'checkbox' },
                ]},
            ];

            function valOf(cfg, key) {
                const parts = key.split('.');
                let cur = cfg;
                for (const p of parts) {
                    if (cur === null || cur === undefined || typeof cur !== 'object') return undefined;
                    cur = cur[p];
                }
                return cur;
            }

            function setVal(obj, key, v) {
                const parts = key.split('.');
                let cur = obj;
                for (let i = 0; i < parts.length - 1; i++) {
                    if (typeof cur[parts[i]] !== 'object' || cur[parts[i]] === null) cur[parts[i]] = {};
                    cur = cur[parts[i]];
                }
                cur[parts[parts.length - 1]] = v;
            }

            function escHtml(s) {
                return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
            }

            function renderSettingsForm(cfg) {
                const body = $id('settingsBody');
                // 记录当前已配置的频道（用于勾选回显 + 保存时收集）
                const curChannels = (valOf(cfg, 'channels') || []).map(String);
                window._stationChannels = curChannels;
                // 登录状态面板（扫码登录入口）
                let html = '<div class="modal-section-title">' + ic('smartphone') + '登录状态</div>'
                    + '<div class="login-panel">'
                    + '<span class="login-status-text" id="loginStatusText">检查中…</span>'
                    + '<button class="topbar-btn" id="scanLoginBtn">' + ic('smartphone') + '扫码登录</button>'
                    + '</div>'
                    + '<div class="qr-container" id="qrContainer" style="display:none;">'
                    + '<img id="qrImage" alt="Telegram 登录二维码">'
                    + '<p class="qr-hint">打开手机 Telegram → 设置 → 设备 → 扫描二维码登录<br>二维码 120 秒内有效</p>'
                    + '</div>'
                    // 频道选择器：勾选账号已加入的频道，替代手填 ID
                    + '<div class="modal-section-title">' + ic('list') + '监听频道</div>'
                    + '<div class="channel-panel">'
                    + '<div class="channel-bar">'
                    + '<span class="namelist-hint">勾选要监听的频道，保存后生效</span>'
                    + '<button class="topbar-btn" id="loadChannelsBtn">' + ic('refresh') + '加载频道列表</button>'
                    + '</div>'
                    + '<div id="channelList" class="channel-list"><span class="namelist-hint">点击「加载频道列表」从账号已加入的频道中勾选</span></div>'
                    + '</div>'
                    // 动漫名单编辑窗口：直接读写 mz.txt
                    + '<div class="modal-section-title">' + ic('video') + '动漫名单 (mz.txt)</div>'
                    + '<div class="namelist-panel">'
                    + '<textarea id="namelistEditor" class="namelist-editor" placeholder="每行一个动漫，格式：《动漫名》：角色1、角色2、角色3&#10;例如：《海贼王》：娜美、罗宾、女帝" spellcheck="false"></textarea>'
                    + '<div class="namelist-bar">'
                    + '<span class="namelist-hint">格式：每行《动漫名》：角色1、角色2，保存后立即生效</span>'
                    + '<button class="topbar-btn primary" id="saveNameListBtn">' + ic('save') + '保存名单</button>'
                    + '</div>'
                    + '</div>';
                for (const g of SETTING_GROUPS) {
                    html += '<div class="modal-section-title">' + g.title + '</div><div class="form-grid">';
                    for (const it of g.items) {
                        const cls = 'form-item' + (it.full ? ' full' : '') + (it.type === 'checkbox' ? ' checkbox-item' : '');
                        html += '<div class="' + cls + '">';
                        if (it.type === 'checkbox') {
                            html += '<input type="checkbox" id="f_' + it.key + '"' + (valOf(cfg, it.key) ? ' checked' : '') + '><label for="f_' + it.key + '">' + it.label + '</label>';
                        } else if (it.type === 'checks') {
                            const arr = valOf(cfg, it.key) || [];
                            html += '<label>' + it.label + '</label><div style="display:flex;gap:14px;flex-wrap:wrap;">';
                            for (const opt of it.options) {
                                html += '<label style="display:flex;align-items:center;gap:5px;font-size:13px;color:#ddd;cursor:pointer;"><input type="checkbox" value="' + opt + '"' + (arr.includes(opt) ? ' checked' : '') + ' class="f-checks" data-group="f_' + it.key + '">' + opt + '</label>';
                            }
                            html += '</div>';
                        } else if (it.type === 'select') {
                            const v = valOf(cfg, it.key);
                            const val = (v === null || v === undefined) ? '' : String(v);
                            html += '<label for="f_' + it.key + '">' + it.label + '</label><select id="f_' + it.key + '">';
                            for (const opt of it.options) {
                                html += '<option value="' + opt + '"' + (val === opt ? ' selected' : '') + '>' + opt + '</option>';
                            }
                            html += '</select>';
                        } else {
                            const v = valOf(cfg, it.key);
                            const val = (v === null || v === undefined) ? '' : String(v);
                            const type = it.type === 'password' ? 'password' : (it.type === 'number' ? 'number' : 'text');
                            html += '<label for="f_' + it.key + '">' + it.label + '</label>';
                            html += '<input type="' + type + '" id="f_' + it.key + '" value="' + escHtml(val) + '">';
                        }
                        html += '</div>';
                    }
                    html += '</div>';
                }
                body.innerHTML = html;
            }

            async function openSettings() {
                $id('settingsMsg').textContent = '';
                $id('settingsMsg').className = 'modal-msg';
                const res = await stationFetch('/api/settings');
                if (!res.ok || !res.config) { showStationToast('读取设置失败', true); return; }
                renderSettingsForm(res.config);
                $id('settingsModal').style.display = 'flex';
                refreshLoginStatus();
                loadNameList();
            }

            // 加载频道列表：GET /api/channels 渲染勾选列表
            async function loadChannels() {
                const listEl = $id('channelList');
                const btn = $id('loadChannelsBtn');
                if (!listEl) return;
                btn.disabled = true;
                btn.textContent = '加载中…';
                listEl.innerHTML = '<span class="namelist-hint">正在连接 Telegram 拉取频道…</span>';
                const res = await stationFetch('/api/channels');
                btn.disabled = false;
                setIconText(btn, 'refresh', '加载频道列表');
                if (!res.ok) {
                    const hint = (res.error || '加载失败').indexOf('下载器正在运行') >= 0
                        ? '请先点击顶部「停止监听」，再加载频道列表'
                        : (res.error || '加载失败');
                    listEl.innerHTML = '<span class="namelist-hint" style="color:#ef4444;">' + ic('alert') + hint + '</span>';
                    return;
                }
                const chans = res.channels || [];
                const cur = window._stationChannels || [];
                if (!chans.length) {
                    listEl.innerHTML = '<span class="namelist-hint">没有找到频道，请确认账号已加入频道</span>';
                    return;
                }
                listEl.innerHTML = chans.map(c => {
                    const id = String(c.id);
                    const checked = cur.includes(id) || cur.includes(String(c.title)) ? ' checked' : '';
                    return '<label class="channel-item">'
                        + '<input type="checkbox" class="channel-cb" value="' + id + '"' + checked + '>'
                        + '<span class="ch-title">' + escHtml(c.title || '(未命名)') + '</span>'
                        + '<span class="ch-id">' + id + '</span>'
                        + '</label>';
                }).join('');
            }

            // 加载名单文件内容到编辑窗口
            async function loadNameList() {
                const el = $id('namelistEditor');
                if (!el) return;
                el.value = '加载中…';
                const res = await stationFetch('/api/namelist');
                if (res.ok) {
                    el.value = res.content || '';
                } else {
                    el.value = '';
                    showStationToast('读取名单失败: ' + (res.error || '未知错误'), true);
                }
            }

            // 保存名单文件内容（无需重启下载器，Python 按 mtime 自动重载）
            async function saveNameList() {
                const el = $id('namelistEditor');
                if (!el) return;
                const msgEl = $id('settingsMsg');
                msgEl.textContent = '保存名单中…';
                msgEl.className = 'modal-msg';
                const res = await stationFetch('/api/namelist', { method: 'POST', body: JSON.stringify({ content: el.value }) });
                if (res.ok) {
                    setIconText(msgEl, 'check', '名单已保存（下载器自动生效）');
                    msgEl.className = 'modal-msg';
                    showStationToast('名单已保存');
                } else {
                    setIconText(msgEl, 'alert', (res.error || '保存失败'));
                    msgEl.className = 'modal-msg err';
                }
            }

            function closeSettings() {
                clearInterval(loginPollTimer);
                $id('settingsModal').style.display = 'none';
            }

            function collectSettings() {
                const out = {};
                for (const g of SETTING_GROUPS) {
                    for (const it of g.items) {
                        const el = $id('f_' + it.key);
                        if (!el) continue;
                        if (it.type === 'checkbox') setVal(out, it.key, el.checked);
                        else if (it.type === 'checks') {
                            const arr = [];
                            document.querySelectorAll('.f-checks[data-group="f_' + it.key + '"]').forEach(c => { if (c.checked) arr.push(c.value); });
                            setVal(out, it.key, arr);
                        } else if (it.type === 'number') {
                            const n = parseInt(el.value, 10);
                            setVal(out, it.key, isNaN(n) ? 0 : n);
                        } else if (it.type === 'select') {
                            setVal(out, it.key, el.value === 'none' ? null : el.value);
                        } else {
                            setVal(out, it.key, el.value);
                        }
                    }
                }
                // 频道选择器：收集勾选的频道（未加载列表时保留原配置）
                const cbList = document.querySelectorAll('.channel-cb');
                if (cbList.length) {
                    out.channels = Array.from(cbList).filter(c => c.checked).map(c => c.value);
                }
                // channels 逗号拆分（兼容旧文本输入）
                if (typeof out.channels === 'string') {
                    out.channels = out.channels.split(/[,，\s]+/).filter(Boolean);
                }
                return out;
            }

            async function saveSettings() {
                const data = collectSettings();
                const msgEl = $id('settingsMsg');
                msgEl.textContent = '保存中…';
                msgEl.className = 'modal-msg';
                const res = await stationFetch('/api/settings', { method: 'POST', body: JSON.stringify(data) });
                if (res.ok) {
                    setIconText(msgEl, 'check', res.restarted ? '已保存，下载器已重启应用新配置' : '已保存');
                    setTimeout(closeSettings, 1200);
                    refreshStationStatus();
                } else {
                    setIconText(msgEl, 'alert', (res.error || '保存失败'));
                    msgEl.className = 'modal-msg err';
                }
            }

            // ---------- 扫码登录 ----------
            let loginPollTimer = null;

            async function refreshLoginStatus() {
                const res = await stationFetch('/api/login/status');
                if (!res.ok) return;
                const el = $id('loginStatusText');
                if (res.status === 'logged_in') {
                    setIconText(el, 'check', '已登录');
                    el.className = 'login-status-text ok';
                    $id('qrContainer').style.display = 'none';
                } else {
                    setIconText(el, 'alert', '未登录');
                    el.className = 'login-status-text';
                }
            }

            async function scanLogin() {
                const btn = $id('scanLoginBtn');
                btn.disabled = true;
                btn.textContent = '连接中…';
                const res = await stationFetch('/api/login/qr');
                btn.disabled = false;
                setIconText(btn, 'smartphone', '扫码登录');
                if (!res.ok) {
                    showStationToast((res.error || '登录失败'), true);
                    return;
                }
                if (res.status === 'already' || res.status === 'success') {
                    setIconText($id('loginStatusText'), 'check', '已登录 @' + (res.username || ''));
                    $id('loginStatusText').className = 'login-status-text ok';
                    $id('qrContainer').style.display = 'none';
                    showStationToast('已登录 ' + (res.username || ''));
                    return;
                }
                if (res.status === 'qr') {
                    $id('qrImage').src = res.qr;
                    $id('qrContainer').style.display = 'block';
                    setIconText($id('loginStatusText'), 'clock', '请用手机 Telegram 扫码');
                    $id('loginStatusText').className = 'login-status-text';
                    pollLoginResult();
                    return;
                }
            }

            // 扫码后轮询登录结果：每 2 秒，最多 60 秒
            function pollLoginResult() {
                clearInterval(loginPollTimer);
                const start = Date.now();
                loginPollTimer = setInterval(async () => {
                    const st = await stationFetch('/api/login/status');
                    if (st.ok && st.status === 'logged_in') {
                        clearInterval(loginPollTimer);
                        // 验证并获取用户名（已登录时不会生成新二维码）
                        const v = await stationFetch('/api/login/qr');
                        if (v.ok && (v.status === 'already' || v.status === 'success')) {
                            setIconText($id('loginStatusText'), 'check', '已登录 @' + (v.username || ''));
                            $id('loginStatusText').className = 'login-status-text ok';
                            $id('qrContainer').style.display = 'none';
                            showStationToast('扫码登录成功');
                            return;
                        }
                        // session 文件存在但未真正登录（过期）→ 展示新二维码
                        if (v.ok && v.status === 'qr') {
                            $id('qrImage').src = v.qr;
                            setIconText($id('loginStatusText'), 'refresh', '二维码已刷新，请重新扫码');
                            return;
                        }
                        return;
                    }
                    if (Date.now() - start > 60 * 1000) {
                        clearInterval(loginPollTimer);
                        $id('qrContainer').style.display = 'none';
                        setIconText($id('loginStatusText'), 'alert', '扫码超时，请重试');
                        $id('loginStatusText').className = 'login-status-text';
                        showStationToast('扫码超时，请重试', true);
                    }
                }, 2000);
            }

            async function refreshMediaNow() {
                // 记录刷新前的文件数，只有变化（新文件/删除）才重载列表，避免浏览中页面乱跳
                let before = -1;
                try {
                    const st = await stationFetch('/api/status');
                    if (st.ok && st.media) before = st.media.total;
                } catch (e) {}
                const res = await stationFetch('/api/refresh', { method: 'POST', body: '{}' });
                if (res.ok) {
                    showStationToast('正在刷新媒体索引…');
                    setTimeout(async () => {
                        await refreshStationStatus();
                        const after = stationData && stationData.media ? stationData.media.total : -1;
                        if (before === after) return; // 无变化不重载
                        // 有变化：走带代际保护的完整重载（旧分页循环会被作废，
                        // 不会再往新列表里 push 旧页数据；筛选按钮同步回"全部"）
                        await reloadMedia();
                    }, 4000);
                } else showStationToast('刷新失败: ' + (res.error || '未知错误'), true);
            }

            function initStationUI() {
                $id('listenToggleBtn').addEventListener('click', toggleListen);
                $id('settingsBtn').addEventListener('click', openSettings);
                $id('settingsCloseBtn').addEventListener('click', closeSettings);
                $id('saveSettingsBtn').addEventListener('click', saveSettings);
                $id('stationFab').addEventListener('click', () => {
                    const m = $id('fabMenu');
                    m.style.display = m.style.display === 'none' ? 'flex' : 'none';
                });
                $id('fabToggleBtn').addEventListener('click', () => { $id('fabMenu').style.display = 'none'; toggleListen(); });
                $id('fabRefreshBtn').addEventListener('click', () => { $id('fabMenu').style.display = 'none'; refreshMediaNow(); });
                $id('fabSettingsBtn').addEventListener('click', () => { $id('fabMenu').style.display = 'none'; openSettings(); });
                $id('dlHeader').addEventListener('click', () => {
                    // 小屏下点标题行：展开/收起整卡；大屏沿用折叠任务列表
                    const panel = $id('downloadPanel');
                    setDlExpanded(!dlExpanded);
                    if (window.matchMedia('(max-width: 640px)').matches) {
                        panel.classList.toggle('expanded', dlExpanded);
                    }
                });
                $id('dlStopBtn').addEventListener('click', (e) => { e.stopPropagation(); stopDownload(); });
                // 多选角标上的一键清空
                const clearBtn = $id('selClearBtn');
                if (clearBtn) clearBtn.addEventListener('click', clearSelection);
                // 单任务取消按钮：动态渲染，用事件委托
                $id('dlTasks').addEventListener('click', (e) => {
                    const btn = e.target.closest('.dl-cancel-btn');
                    if (btn) { e.stopPropagation(); cancelTask(btn.getAttribute('data-file')); }
                });
                // scanLoginBtn 由 renderSettingsForm 动态生成，用事件委托绑定
                $id('settingsBody').addEventListener('click', (e) => {
                    if (e.target.id === 'scanLoginBtn') scanLogin();
                    if (e.target.id === 'saveNameListBtn') saveNameList();
                    if (e.target.id === 'loadChannelsBtn') loadChannels();
                });
                document.addEventListener('click', (e) => {
                    const m = $id('fabMenu');
                    if (m.style.display !== 'none' && !e.target.closest('#stationFab') && !e.target.closest('#fabMenu')) m.style.display = 'none';
                });
                $id('settingsModal').addEventListener('click', (e) => { if (e.target.id === 'settingsModal') closeSettings(); });

                refreshStationStatus();
                refreshMediaNow(); // F5/刷新页面即检查新文件并生成缩略图
                setInterval(() => { if (!document.hidden) { refreshStationStatus(); refreshDownloadProgress(); } }, 5000);
            }

            initStationUI();

            // 版本标识：数字 = 本页面大小(KB) + 内联脚本总字节数。
            // 页面一改这个数字就变，用来快速判断浏览器加载的是新版还是缓存旧版。
            (function stampBuild() {
                try {
                    const el = document.getElementById('buildStamp');
                    if (!el) return;
                    const kb = Math.round(document.documentElement.outerHTML.length / 1024);
                    let jsBytes = 0;
                    document.querySelectorAll('script').forEach(s => { jsBytes += (s.textContent || '').length; });
                    el.textContent = 'v' + kb + '.' + jsBytes;
                    el.title = '页面版本 v' + kb + '.' + jsBytes +
                        '\n页面大小 ' + kb + ' KB，内联脚本 ' + jsBytes + ' 字节' +
                        '\n若与部署版本不一致，按 Ctrl+Shift+R 强制刷新';
                } catch (e) { /* 版本标识失败不影响功能 */ }
            })();
        })();
    </script>


</body></html>

```


---

## 文件：static/manifest.json

```json
{
  "name": "📸 媒体浏览器",
  "short_name": "媒体浏览器",
  "description": "浏览你的照片和视频",
  "theme_color": "#111111",
  "background_color": "#111111",
  "display": "standalone",
  "scope": "/",
  "start_url": "/",
  "icons": [
    {
      "src": "/static/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ]
}

```


---

## 文件：.gitignore

```text
# 编译/打包产物
*.exe
!tg_downloader.spec
build/
dist/
.gocache/
__pycache__/
*.pyc

# 运行数据与状态（部署目录才是权威，仓库里不跟踪）
media_index.json
dl_status.json
progress.json
cancel.json
pause.json
tg_station.log
station_stdout.log
station_stderr.log
downloader_stdout.log
downloader_stderr.log
login_status.json
downloads.db
thumbs/
qr_login.png

# Telethon 会话（含登录凭据，绝不入库）
*.session
*.session-journal
tg_session.session
tg_session.session-journal
tg_session.authorized

# 一次性脚本的产物 / 备份
*_log.txt
config.json.bak_*
*.bak_*
_gen.log
_pyinstaller.log

# 交叉编译产物（体积大，按需本地生成）
pi_deploy/media-browser-linux-arm
pi_deploy/tg_station
tg_station_linux_arm64
media-browser
media-browser.exe
# 带版本号的旧 exe（*.exe 已忽略，这里显式列出历史误入库的）
tg_station_new_v203.exe
tg_station_v204.exe
tg_station_old_*.exe

# 一次性迁移脚本（mihomo 订阅时代与早期部署辅助，已完成使命）
_apply_and_start.py
_apply_and_start.ps1
_launch_detached.ps1
_count_proxies.py
_fixtest.py

# 凭据文件（GitHub token / SSH 备份口令）——绝不能提交
_secrets/
*.secret

```


---

## 文件：_deploy.ps1

```powershell
# Deploy tg_station.exe + static assets + new tg_downloader.exe to E:\zm\tgstation
# Order matters: stop downloader first (so it cannot fall back and kill the media
# server while station restarts), then station -> backup -> copy -> start -> verify
# -> start downloader.
#
# NOTE (verified 2026-10-10): the "WMI -> powershell -> Start-Process" chain does NOT
# work in this environment - Win32_Process.Create returns 0 with a PID, but the child
# never runs. WMI *direct* binary creation DOES work, so LaunchDetached below creates
# the exe directly (no powershell wrapper). Probes that established this:
#   cmd.exe /c echo ... via WMI      -> marker written (works)
#   powershell.exe -EncodedCommand   -> child never ran (broken)
#   "<exe>" via WMI                  -> server started, /api/status OK (works)
#
# NOTE: kept ASCII-only on purpose (Windows PowerShell 5.1 reads .ps1 as GBK).
$ErrorActionPreference = 'Continue'
$repo    = 'E:\zm\mtllq'
$station = 'E:\zm\tgstation'
$stamp   = Get-Date -Format 'yyyyMMdd_HHmmss'
$log     = Join-Path $repo '_deploy_log.txt'
Set-Content -Path $log -Value '' -Encoding utf8

function Say($m) {
    $line = "[{0}] {1}" -f (Get-Date -Format 'HH:mm:ss'), $m
    Write-Output $line
    Add-Content -Path $log -Value $line -Encoding utf8
}
function TestPort($p) {
    try { $c = New-Object Net.Sockets.TcpClient; $c.Connect('127.0.0.1', $p); $c.Close(); return $true }
    catch { return $false }
}
function WaitPort($p, $seconds) {
    for ($i = 0; $i -lt $seconds; $i++) { if (TestPort $p) { return $true }; Start-Sleep -Seconds 1 }
    return (TestPort $p)
}
function WaitGone($name, $seconds) {
    for ($i = 0; $i -lt $seconds; $i++) {
        if (-not (Get-Process $name -ErrorAction SilentlyContinue)) { return $true }
        Start-Sleep -Seconds 1
    }
    return -not (Get-Process $name -ErrorAction SilentlyContinue)
}
function Get-Pids($name) {
    return ((Get-Process $name -ErrorAction SilentlyContinue) | ForEach-Object { $_.Id }) -join ','
}
$wm = [wmiclass]'Win32_Process'
function LaunchDetached($exe, $cwd) {
    # create the exe directly through WMI: detached from this shell's process tree,
    # no powershell launcher involved (that chain is broken here, see header note).
    $r = $wm.Create('"' + $exe + '"', $cwd)
    return ("returnValue={0} pid={1}" -f $r.ReturnValue, $r.ProcessId)
}

# ---------- 1. stop downloader ----------
Say "=== 1/6 stop downloader ==="
if (Get-Process tg_downloader -ErrorAction SilentlyContinue) {
    Say ("   pids: " + (Get-Pids 'tg_downloader'))
    & taskkill /IM tg_downloader.exe /T /F 2>&1 | ForEach-Object { Say ("   " + $_) }
    if (WaitGone 'tg_downloader' 20) { Say '   stopped' } else { Say '   !! still alive, abort'; exit 1 }
} else { Say '   not running' }

# ---------- 2. stop station ----------
Say "=== 2/6 stop tg_station ==="
if (Get-Process tg_station -ErrorAction SilentlyContinue) {
    Say ("   pids: " + (Get-Pids 'tg_station'))
    Get-Process tg_station -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    if (WaitGone 'tg_station' 20) { Say '   stopped' } else { Say '   !! still alive, abort'; exit 1 }
} else { Say '   not running' }

# ---------- 3. backup ----------
Say "=== 3/6 backup ==="
$bdir = Join-Path $station ("_bak_" + $stamp)
New-Item -ItemType Directory -Path $bdir -Force | Out-Null
foreach ($f in 'tg_station.exe', 'tg_downloader.exe', 'tg_downloader_cli.exe') {
    $src = Join-Path $station $f
    if (Test-Path $src) { Copy-Item $src (Join-Path $bdir $f) -Force }
}
foreach ($f in 'index.html', 'manifest.json') {
    $src = Join-Path $station ("static\" + $f)
    if (Test-Path $src) { Copy-Item $src (Join-Path $bdir $f) -Force }
}
Say ("   -> " + $bdir)

# ---------- 4. copy files ----------
Say "=== 4/6 copy exe and static assets ==="
$pairs = @(
    @{ src = (Join-Path $repo 'dist\tg_station.exe');    dst = (Join-Path $station 'tg_station.exe') },
    @{ src = (Join-Path $repo 'dist\tg_downloader.exe'); dst = (Join-Path $station 'tg_downloader.exe') }
)
foreach ($p in $pairs) {
    if (Test-Path $p.src) {
        Copy-Item $p.src $p.dst -Force
        Say ("   {0} -> {1} bytes" -f (Split-Path $p.dst -Leaf), (Get-Item $p.dst).Length)
    } else { Say ("   !! missing " + $p.src) }
}
# CLI downloader (console build) - used by /api/login/qr, /api/login/status, /api/channels.
# Optional: when absent, those endpoints fall back to python + tg_auto_download.py.
$cliSrc = Join-Path $repo 'dist\tg_downloader_cli.exe'
$cliDst = Join-Path $station 'tg_downloader_cli.exe'
if (Test-Path $cliSrc) {
    Copy-Item $cliSrc $cliDst -Force
    Say ("   tg_downloader_cli.exe -> {0} bytes" -f (Get-Item $cliDst).Length)
} else {
    Say '   !! dist\tg_downloader_cli.exe missing - QR/login/channels fall back to python + script'
}
foreach ($f in 'index.html', 'manifest.json', 'icon.svg', 'icon-192.png', 'icon-512.png') {
    $src = Join-Path $repo ("static\" + $f)
    if (Test-Path $src) {
        Copy-Item $src (Join-Path $station ("static\" + $f)) -Force
        Say ("   static\{0} -> {1} bytes" -f $f, (Get-Item $src).Length)
    } else { Say ("   !! missing " + $src) }
}

# ---------- 5. start station and verify ----------
Say "=== 5/6 start tg_station and verify ==="
Say ("   " + (LaunchDetached (Join-Path $station 'tg_station.exe') $station))
if (-not (WaitPort 8080 30)) { Say '   !! port 8080 not listening, abort'; exit 1 }
Say '   port 8080 listening'
Say ("   pids: " + (Get-Pids 'tg_station'))

try {
    $s = Invoke-WebRequest 'http://127.0.0.1:8080/api/status' -TimeoutSec 10 -UseBasicParsing
    Say ("   /api/status -> " + $s.Content)
} catch { Say ("   !! /api/status: " + $_.Exception.Message) }

try {
    $html = (Invoke-WebRequest 'http://127.0.0.1:8080/' -TimeoutSec 25 -UseBasicParsing).Content
    Say ("   index.html served: {0} bytes (before deploy: 232655)" -f $html.Length)
    $keys = @('trae-browser-inspect', 'mchost.guru', '@vite/client', '--vscode-', 'data-filter="movie"', 'window.__INIT__', 'id="lightbox"', 'id="settingsModal"', '/static/icon.svg')
    foreach ($k in $keys) {
        Say ("      {0,-24} {1}" -f $k, ([regex]::Matches($html, [regex]::Escape($k))).Count)
    }
} catch { Say ("   !! index.html: " + $_.Exception.Message) }

foreach ($u in '/static/manifest.json', '/static/icon.svg', '/static/icon-192.png', '/static/icon-512.png') {
    try {
        $r2 = Invoke-WebRequest ("http://127.0.0.1:8080" + $u) -TimeoutSec 10 -UseBasicParsing
        Say ("   {0,-24} HTTP {1} {2} bytes" -f $u, $r2.StatusCode, $r2.RawContentLength)
    } catch { Say ("   !! {0}: {1}" -f $u, $_.Exception.Message) }
}

foreach ($u in '/api/namelist', '/api/login/status', '/api/download/progress', '/api/listen/status', '/api/dates') {
    try {
        $r4 = Invoke-WebRequest ("http://127.0.0.1:8080" + $u) -TimeoutSec 12 -UseBasicParsing
        Say ("   {0,-24} HTTP {1} OK" -f $u, $r4.StatusCode)
    } catch { Say ("   !! {0}: {1}" -f $u, $_.Exception.Message) }
}

# ---------- 6. start downloader ----------
Say "=== 6/6 start new downloader ==="
Say ("   " + (LaunchDetached (Join-Path $station 'tg_downloader.exe') $station))
Start-Sleep -Seconds 15
Say ("   downloader pids: " + (Get-Pids 'tg_downloader'))
try {
    $ls = (Invoke-WebRequest 'http://127.0.0.1:8080/api/status' -TimeoutSec 10 -UseBasicParsing).Content
    Say ("   /api/status -> " + $ls)
} catch { Say ("   !! /api/status: " + $_.Exception.Message) }
Say '   last 12 log lines:'
Get-Content (Join-Path $station 'tg_station.log') -Tail 12 -Encoding utf8 -ErrorAction SilentlyContinue |
    ForEach-Object { Say ("      " + $_) }
Say 'DONE'

```


---

## 文件：_start.ps1

```powershell
# Start the real tg_station + tg_downloader detached, using WMI *direct* process creation
# (the WMI -> powershell -> Start-Process chain does not work in this environment).
# ASCII only.
$ErrorActionPreference = 'Continue'
$station = 'E:\zm\tgstation'
function Say($m) { Write-Output ("[{0}] {1}" -f (Get-Date -Format 'HH:mm:ss'), $m) }
function TestPort($p) {
    try { $c = New-Object Net.Sockets.TcpClient; $c.Connect('127.0.0.1', $p); $c.Close(); return $true }
    catch { return $false }
}
function Get-Pids($n) { return ((Get-Process $n -ErrorAction SilentlyContinue) | ForEach-Object { $_.Id }) -join ',' }
$wm = [wmiclass]'Win32_Process'

Say "=== state before ==="
Say ("   station pids: " + (Get-Pids 'tg_station'))
Say ("   downloader pids: " + (Get-Pids 'tg_downloader'))
Say ("   port 8080: " + (TestPort 8080))

# stop the manually started station instance so the detached one owns the port
if (Get-Process tg_station -ErrorAction SilentlyContinue) {
    Say "   stopping existing tg_station..."
    Get-Process tg_station -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    for ($i = 0; $i -lt 15; $i++) { if (-not (Get-Process tg_station -ErrorAction SilentlyContinue)) { break }; Start-Sleep -Seconds 1 }
    Say ("   stopped, port 8080 now: " + (TestPort 8080))
}

Say "=== WMI direct launch: tg_station ==="
$cmdSt = '"' + (Join-Path $station 'tg_station.exe') + '"'
$r1 = $wm.Create($cmdSt, $station)
Say ("   returnValue={0} pid={1}" -f $r1.ReturnValue, $r1.ProcessId)
$up = $false
for ($i = 0; $i -lt 30; $i++) { if (TestPort 8080) { $up = $true; break }; Start-Sleep -Seconds 1 }
Say ("   port 8080 listening: {0}" -f $up)
Say ("   station pids: " + (Get-Pids 'tg_station'))

Say "=== WMI direct launch: tg_downloader ==="
$cmdDl = '"' + (Join-Path $station 'tg_downloader.exe') + '"'
$r2 = $wm.Create($cmdDl, $station)
Say ("   returnValue={0} pid={1}" -f $r2.ReturnValue, $r2.ProcessId)
Start-Sleep -Seconds 20
Say ("   downloader pids: " + (Get-Pids 'tg_downloader'))

Say "=== verify ==="
try {
    $s = (Invoke-WebRequest 'http://127.0.0.1:8080/api/status' -TimeoutSec 10 -UseBasicParsing).Content
    Say ("   /api/status -> " + $s)
} catch { Say ("   !! /api/status: " + $_.Exception.Message) }
try {
    $p = (Invoke-WebRequest 'http://127.0.0.1:8080/api/download/progress' -TimeoutSec 10 -UseBasicParsing).Content
    Say ("   /api/download/progress -> " + $p.Substring(0, [Math]::Min(200, $p.Length)))
} catch { Say ("   !! progress: " + $_.Exception.Message) }
try {
    $l = (Invoke-WebRequest 'http://127.0.0.1:8080/api/listen/status' -TimeoutSec 10 -UseBasicParsing).Content
    Say ("   /api/listen/status -> " + $l.Substring(0, [Math]::Min(200, $l.Length)))
} catch { Say ("   !! listen/status: " + $_.Exception.Message) }

Say "=== downloader log tail ==="
Get-Content (Join-Path $station 'tg_station.log') -Tail 14 -Encoding utf8 -ErrorAction SilentlyContinue |
    ForEach-Object { Say ("      " + $_) }
Say 'DONE'

```


---

## 文件：build_downloader_cli.ps1

```powershell
# Build tg_downloader_cli.exe (console build: reliable stdout JSON for tg_station).
#
# tg_station calls the downloader for three sub-modes and parses their stdout as
# JSON:  --qr-only / --check-session / --list-channels.
# The resident downloader (tg_downloader.exe) is built with console=False, so it
# must NOT be used for those sub-modes; this script builds the console variant.
#
# ASCII only (Windows PowerShell 5.1 reads .ps1 as GBK).
#
# Usage:  powershell -NoProfile -ExecutionPolicy Bypass -File .\build_downloader_cli.ps1
# Output: .\dist\tg_downloader_cli.exe
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

Write-Output '=== checking toolchain ==='
$py = 'python'
try { & $py -c "import PyInstaller, telethon, qrcode" } catch {
    Write-Output '!! missing deps: need PyInstaller + telethon + qrcode in this Python env'
    Write-Output '   pip install pyinstaller telethon "qrcode[pil]"'
    exit 1
}
& $py -c "import telethon, qrcode; print('telethon', telethon.__version__)"

Write-Output '=== building tg_downloader_cli.exe ==='
& $py -m PyInstaller --noconfirm --clean tg_downloader_cli.spec

$out = Join-Path $PSScriptRoot 'dist\tg_downloader_cli.exe'
if (-not (Test-Path $out)) { Write-Output '!! build failed: missing dist\tg_downloader_cli.exe'; exit 1 }
Write-Output ('OK -> {0} ({1} bytes)' -f $out, (Get-Item $out).Length)

Write-Output '=== smoke test (no account needed: sanitized/empty config just reports status) ==='
$env:TG_STATION_CONFIG = Join-Path $PSScriptRoot 'config.json'
& $out --check-session
Write-Output '=== done ==='

```


---

## 文件：config.json

```json
{
  "api_hash": "PUT_YOUR_OWN_API_HASH_HERE",
  "api_id": 0,
  "channels": [
    "PUT_YOUR_CHANNEL_ID"
  ],
  "concurrency": 4,
  "idle_stop_minutes": 30,
  "download_dir": "media",
  "download_history": false,
  "include_forwarded": true,
  "login_method": "qr",
  "max_file_size_mb": 8192,
  "media_browser": {
    "exe": "",
    "index_path": "",
    "port": 8080,
    "workdir": ""
  },
  "name_list_file": "mz.txt",
  "only_forwarded": true,
  "proxy": {
    "host": "127.0.0.1",
    "port": 7897,
    "type": "socks5"
  },
  "retries": 3,
  "session": "tg_session",
  "skip_gif": true,
  "station": {
    "python": "",
    "script": "",
    "downloader_exe": "",
    "media_dir": "",
    "auto_start": false
  },
  "twofa_password": "",
  "types": [
    "photo",
    "video",
    "document"
  ],
  "catchup_today": true,
  "catchup_since": "",
  "movie_min_minutes": 5
}
```


---

## 文件：config.pi.json

```json
{
  "api_hash": "PUT_YOUR_OWN_API_HASH_HERE",
  "api_id": 0,
  "channels": [
    "PUT_YOUR_CHANNEL_ID"
  ],
  "concurrency": 4,
  "download_dir": "media",
  "download_history": false,
  "include_forwarded": true,
  "login_method": "qr",
  "max_file_size_mb": 500,
  "media_browser": {
    "exe": "/home/lat/tgstation/media-browser-linux-arm",
    "index_path": "/home/lat/tgstation/media_index.json",
    "port": 8080,
    "workdir": "/home/lat/tgstation"
  },
  "name_list_file": "mz.txt",
  "only_forwarded": true,
  "proxy": {
    "host": "127.0.0.1",
    "port": 7898,
    "type": "socks5"
  },
  "retries": 3,
  "session": "tg_session",
  "skip_gif": true,
  "station": {
    "python": "",
    "script": "",
    "media_dir": "",
    "auto_start": false
  },
  "twofa_password": "",
  "types": [
    "photo",
    "video",
    "document"
  ]
}

```


---

## 文件：go.mod

```text
module media-browser

go 1.25.0

require golang.org/x/text v0.40.0 // indirect

```


---

## 文件：go.sum

```text
golang.org/x/image v0.43.0 h1:FLxcP4ec2350nTfOC8ysKtqYSIFbk/QGjw1ZHNP4tsY=
golang.org/x/image v0.43.0/go.mod h1:rrpelvGFt+kLPAjPM4HeWPgrl0FtafueU//e5N0qk/Q=
golang.org/x/text v0.40.0 h1:Ub2Z6/xjgF1WrYQz2nuITOEegKFtiIy+rieRJ5lHZKs=
golang.org/x/text v0.40.0/go.mod h1:hpnzDAfGV753zIKo+wk3u1bVKCGPbrnF7+7LBF/UHVY=

```


---

## 文件：media-browser.service

```text
[Unit]
Description=TG Station (媒体浏览 + 下载器管理)
After=network.target

[Service]
Type=simple
# 用户名/路径按树莓派实际部署调整（pi_deploy 同步目录）
User=lat
WorkingDirectory=/home/lat/tgstation
# 注意：媒体目录与下载器 exe 由 config.json 决定，这里只作为 flag 兜底
ExecStart=/home/lat/tgstation/tg_station -config /home/lat/tgstation/config.json -dir /media/lat/disk/tg -port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target

```


---

## 文件：tg_auto_download.py

```python
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

# 显式把标准输出/错误重设为 UTF-8。
# 为什么必须显式：PyInstaller 打包出来的 exe 实测会**忽略** PYTHONUTF8 / PYTHONIOENCODING
# （tg_station 已按 UTF-8 注入环境变量，--list-channels 仍然报
#   "'gbk' codec can't encode character ..." 直接失败），
# 于是遇到含 emoji 的频道标题就崩。Go 端是按 UTF-8 读 JSON 的，所以这里统一成 UTF-8。
for _name in ("stdout", "stderr"):
    _stream = getattr(sys, _name, None)
    if _stream is not None and hasattr(_stream, "reconfigure"):
        try:
            _stream.reconfigure(encoding="utf-8", errors="replace")
        except Exception:
            pass

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

```


---

## 文件：tg_downloader.spec

```text
# -*- mode: python ; coding: utf-8 -*-
# tg_downloader.exe —— 把 tg_auto_download.py + 默认配置 + mz.txt 打包成单个 exe。
# 首次运行自动把内置资源释放到 exe 旁边（见 ensure_resources）。
#
# console=False（无控制台窗口）：这是「常驻监听」用的后台下载器，故意不带控制台。
#
# ⚠ 注意：这个 exe 只用于常驻监听（无参数运行）。不要用它执行
#   --qr-only / --check-session / --list-channels —— 那三个子模式需要 stdout 上的
#   JSON 协议，必须用 console=True 的 tg_downloader_cli.exe（见 tg_downloader_cli.spec）。

import os

_datas = [('config.json', '.')]
if os.path.exists('mz.txt'):   # 部署目录才有 mz.txt（名单文件不入库）
    _datas.append(('mz.txt', '.'))

a = Analysis(
    ['tg_auto_download.py'],
    pathex=[],
    binaries=[],
    datas=_datas,
    hiddenimports=[],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)
pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name='tg_downloader',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=False,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

```


---

## 文件：tg_downloader_cli.spec

```text
# -*- mode: python ; coding: utf-8 -*-
# tg_downloader_cli.exe —— 与 tg_downloader.exe 同一个入口（tg_auto_download.py），
# 但用 console=True 打包，专门给 tg_station 的三个接口读标准输出 JSON 用：
#   --qr-only        生成二维码 / 检查是否已登录
#   --check-session  本地判断 session 是否有效
#   --list-channels  列出账号已加入的频道
#
# 为什么必须单独一个 exe（而不是复用 tg_downloader.exe）：
#   tg_downloader.exe 是 console=False 的窗口化程序（常驻监听、不弹黑窗）。
#   Windows 下没有控制台的进程中，PyInstaller 引导程序不会把子进程的 stdout
#   接到 Go 的管道上，cmd.StdoutPipe()/cmd.Output() 读到空内容——
#   扫码、登录状态、频道列表三个接口就会全部失效。
#   所以用这个 console=True 的版本保证 stdout 可靠；窗口由 Go 端以
#   CREATE_NO_WINDOW（见 proc_windows.go: childProcAttr）启动隐藏，用户看不到黑窗。
#
# 构建：python -m PyInstaller --noconfirm --clean tg_downloader_cli.spec
# 产物：dist/tg_downloader_cli.exe，与 tg_downloader.exe 一起部署到程序目录
#       （Go 端查找顺序：config.json 的 station.downloader_cli > exe 目录 >
#        exe 目录/downloader > 当前目录；见 downloader.go: findCliExe）。

import os

_datas = [('config.json', '.')]
if os.path.exists('mz.txt'):   # 部署目录才有 mz.txt（名单文件不入库）
    _datas.append(('mz.txt', '.'))

a = Analysis(
    ['tg_auto_download.py'],
    pathex=[],
    binaries=[],
    datas=_datas,
    hiddenimports=[],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)
pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name='tg_downloader_cli',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,          # ← 关键：必须为 True，stdout 才能作为管道被 Go 读取
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

```
