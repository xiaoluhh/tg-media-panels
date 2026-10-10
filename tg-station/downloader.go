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
