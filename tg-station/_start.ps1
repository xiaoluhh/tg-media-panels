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
