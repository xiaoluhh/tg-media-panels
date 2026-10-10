package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestProbeExeStdout 是「按需」探测工具：把任意下载器可执行文件丢给它，看 Go 能不能
// 从 stdout 读到协议 JSON。用于验证「console=False 的窗口版 exe 不能用于
// --check-session / --qr-only / --list-channels」这一结论。
//
// 用法（默认跳过，不设置 TG_PROBE_CLI 就不会运行）：
//
//	TG_PROBE_CLI=dist/tg_downloader_cli.exe go test -run TestProbeExeStdout -v .
//
// 生产代码不使用这个测试，它只用于人工/CI 验证构建产物。
func TestProbeExeStdout(t *testing.T) {
	probe := os.Getenv("TG_PROBE_CLI")
	if probe == "" {
		t.Skip("未设置 TG_PROBE_CLI，跳过构建产物探测")
	}
	d := newTestDownloader(probe, "")
	withConfigPath(t, filepath.Join(t.TempDir(), "config.json"))

	obj, err := d.runModeJSON(modeCheckSession, 60*time.Second)
	if err != nil {
		t.Logf("探测 %s → 读不到 JSON：%v", probe, err)
		return
	}
	t.Logf("探测 %s → 读到 JSON：%v", probe, obj)
}
