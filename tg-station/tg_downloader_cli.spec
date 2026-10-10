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
