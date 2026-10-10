# -*- mode: python ; coding: utf-8 -*-
# tg_downloader.exe —— 把 tg_auto_download.py + 默认配置 + mz.txt + mihomo 代理
# 打包成单个 exe。首次运行自动把内置资源释放到 exe 旁边（见 ensure_resources）。

a = Analysis(
    ['tg_auto_download.py'],
    pathex=[],
    binaries=[],
    datas=[('config.json', '.'), ('mz.txt', '.')],
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
