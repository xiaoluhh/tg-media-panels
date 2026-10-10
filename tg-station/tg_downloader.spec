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
