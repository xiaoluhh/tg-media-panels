package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 下载器（Python 子进程）管理。
// 通过 tg_station 自身的 config.json 与 Python 共享配置：
// 启动子进程时设置环境变量 TG_STATION_CONFIG 指向同一份文件。

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

// Downloader 管理 Python 下载器子进程
type Downloader struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	state    string
	pid      int
	started  time.Time
	exitCode int
	exitMsg  string
	logBuf   *ringBuffer
	cfg      *AppConfig
	waitDone chan struct{}
}

var downloader = &Downloader{
	state:    dlStateIdle,
	exitCode: -1,
	logBuf:   newRingBuffer(300),
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

func (d *Downloader) status() ListenerStatus {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 独立进程模式：以实际进程检测为准（下载器可能由开机自启启动，非本进程拉起）
	alive := downloaderAlive()
	pids := findDownloaderPids()

	st := ListenerStatus{
		State:    d.state,
		Running:  alive,
		Pid:      d.pid,
		ExitCode: d.exitCode,
		ExitMsg:  d.exitMsg,
		LogTail:  d.logBuf.tail(100),
	}
	if alive {
		st.State = dlStateRun
		if len(pids) > 0 {
			st.Pid = pids[0]
		}
	} else if d.state == dlStateRun {
		st.State = dlStateExited
	}
	if !d.started.IsZero() {
		st.StartedAt = d.started.Format(time.RFC3339)
	}
	return st
}

// ---------- 启动 / 停止 ----------

// findPython 探测 Python 解释器：配置指定 > PATH > 常见安装路径
func (d *Downloader) findPython() (string, error) {
	if p := d.cfg.Station.Python; p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		return "", fmt.Errorf("配置的 Python 不存在: %s", p)
	}

	// PATH 搜索
	for _, name := range []string{"python", "python.exe", "py"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	// 常见安装路径（含本机 hermes venv）
	candidates := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "hermes", "hermes-agent", "venv", "Scripts", "python.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Python313", "python.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Python312", "python.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Python", "Python311", "python.exe"),
		"C:\\Python313\\python.exe", "C:\\Python312\\python.exe", "C:\\Python311\\python.exe",
	}
	for _, p := range candidates {
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("未找到 Python 解释器，请在设置中填写 python 路径")
}

// findScript 定位 tg_auto_download.py：配置指定 > exe 目录 > 当前目录
func (d *Downloader) findScript() (string, error) {
	candidates := []string{}
	if d.cfg.Station.Script != "" {
		candidates = append(candidates, d.cfg.Station.Script)
	}
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
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
	return "", fmt.Errorf("未找到下载器脚本 tg_auto_download.py（请复制到程序目录或在设置中指定 script 路径）")
}

// findDownloaderExe 探测打包后的下载器 exe：配置指定 > exe 目录/tg_downloader.exe
// 返回 "" 表示未找到（回退到 python+script 旧模式）
func (d *Downloader) findDownloaderExe() string {
	if p := d.cfg.Station.DownloaderExe; p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		return ""
	}
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	for _, c := range []string{
		filepath.Join(exeDir, "tg_downloader.exe"),
		filepath.Join(exeDir, "downloader", "tg_downloader.exe"),
	} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// Start 启动下载器子进程；已在运行则直接返回成功
func (d *Downloader) Start() error {
	d.mu.Lock()
	if d.state == dlStateRun && d.cmd != nil && d.cmd.Process != nil {
		d.mu.Unlock()
		return nil
	}
	d.mu.Unlock()

	// 独立进程模式：若已有下载器实例在跑（开机自启或手动启动），直接复用，不重复启动。
	// 防止双实例共用 session → Telegram 判定异地同时使用 → 封 session。
	if downloaderAlive() {
		log.Printf("[station] 下载器已在运行（独立进程），复用")
		d.mu.Lock()
		d.state = dlStateRun
		d.mu.Unlock()
		return nil
	}

	var (
		cmd  *exec.Cmd
		desc string
	)
	if exe := d.findDownloaderExe(); exe != "" {
		// 新模式：直接运行打包后的 tg_downloader.exe，不依赖 Python 环境
		cmd = exec.Command(exe)
		cmd.Dir = filepath.Dir(exe)
		desc = exe
	} else {
		// 旧模式：python + 脚本
		py, err := d.findPython()
		if err != nil {
			return err
		}
		script, err := d.findScript()
		if err != nil {
			return err
		}
		cmd = exec.Command(py, script)
		cmd.Dir = filepath.Dir(script)
		desc = py + " " + script
	}

	// 与 Python 共享同一份 config.json
	cmd.Env = append(os.Environ(),
		"TG_STATION_CONFIG="+configPath,
		"PYTHONUTF8=1",
		"PYTHONIOENCODING=utf-8",
	)

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

	log.Printf("[station] 下载器已启动 (PID %d): %s", d.pid, desc)

	// 日志转发（容忍非 UTF-8 字节，逐行读取）
	go d.pump(stdout, "out")
	go d.pump(stderr, "err")
	go d.wait(cmd)

	return nil
}

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
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
	}
	msg := ""
	if exitCode != 0 {
		msg = fmt.Sprintf("进程异常退出，退出码 %d", exitCode)
	}

	d.mu.Lock()
	d.state = dlStateExited
	d.exitCode = exitCode
	d.exitMsg = msg
	if d.waitDone != nil {
		close(d.waitDone)
		d.waitDone = nil
	}
	d.mu.Unlock()

	log.Printf("[station] 下载器已退出 (PID %d, 退出码 %d)", d.pid, exitCode)
}

// Stop 停止下载器：独立进程模式按进程名杀所有实例（含进程树）
func (d *Downloader) Stop() error {
	d.mu.Lock()
	cmd := d.cmd
	state := d.state
	d.mu.Unlock()

	if !downloaderAlive() && !(cmd != nil && state == dlStateRun) {
		return nil
	}

	if runtime.GOOS == "windows" {
		// 独立模式：按进程名杀（PyInstaller 引导进程 + python 子进程整树）
		killDownloaderByName()
	} else if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}

	// 等待退出（最多 10 秒）
	done := make(chan struct{})
	go func() {
		d.mu.Lock()
		ch := d.waitDone
		d.mu.Unlock()
		if ch != nil {
			<-ch
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}

	// 手动停止：无论 wait goroutine 记录了什么退出状态，都归位为 idle。
	// 若期间已有新进程启动（d.cmd 被替换），则不动状态。
	d.mu.Lock()
	if d.cmd == cmd {
		d.state = dlStateIdle
		d.cmd = nil
		d.pid = 0
		d.exitCode = -1
		d.exitMsg = ""
	}
	d.mu.Unlock()

	log.Printf("[station] 下载器已停止")
	return nil
}

// Restart 重启下载器（设置变更后应用新配置）
func (d *Downloader) Restart() error {
	wasRunning := d.isRunning()
	if err := d.Stop(); err != nil {
		return err
	}
	if !wasRunning {
		return nil
	}
	return d.Start()
}

func (d *Downloader) isRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state == dlStateRun
}

// ---------- 独立进程模式（下载器不再由 tg_station 托管） ----------
//
// 2026-08-09 起：下载器改为开机自启的独立进程（Downloader.cmd），tg_station 重启
// 不再影响下载。tg_station 通过进程名检测下载器是否存活，网页"开始/停止"按钮
// 也只按进程名操作，绝不重复启动（防止双实例共用 session 导致 Telegram 封号）。

const downloaderProcName = "tg_downloader.exe"

// findDownloaderPids 按进程名查找 tg_downloader.exe，返回 PID 列表
func findDownloaderPids() []int {
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

// downloaderAlive 下载器是否已在运行（独立进程模式）
func downloaderAlive() bool {
	return len(findDownloaderPids()) > 0
}

// killDownloaderByName 按进程名杀所有下载器实例（含进程树，PyInstaller 双进程）
func killDownloaderByName() {
	exec.Command("taskkill", "/IM", downloaderProcName, "/T", "/F").Run()
}

