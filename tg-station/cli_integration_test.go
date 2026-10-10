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
