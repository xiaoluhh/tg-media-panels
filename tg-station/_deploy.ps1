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
foreach ($f in 'tg_station.exe', 'tg_downloader.exe') {
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
