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
