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
