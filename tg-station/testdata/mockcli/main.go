// mockcli 是下载器 CLI（tg_downloader_cli）的测试替身：
// 只实现 Go 侧依赖的 JSON 协议与各种异常行为（空输出 / 非零退出 / 非法 JSON /
// 超时 / 回显工作目录与环境变量），完全不接触 Telegram，也不读写任何配置或凭据。
//
// 行为由自身文件名决定（同一个测试进程里可以准备多个变体）：
//
//	名字含 "empty"   → 什么都不输出，退出码 0
//	名字含 "nonzero" → 输出一行日志后以退出码 3 结束
//	名字含 "garbage" → 输出一行非 JSON 文本
//	名字含 "slow"    → 先睡 30 秒再输出
//	其它              → 正常输出协议 JSON
//
// 同时可当「Python 解释器替身」用：validatePython 会执行 `<exe> --version`。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func behavior() string {
	name := strings.ToLower(filepath.Base(os.Args[0]))
	switch {
	case strings.Contains(name, "empty"):
		return "empty"
	case strings.Contains(name, "nonzero"):
		return "nonzero"
	case strings.Contains(name, "garbage"):
		return "garbage"
	case strings.Contains(name, "slow"):
		return "slow"
	}
	return "ok"
}

// mode 从参数里找子模式（CLI 形态：<exe> --mode；Python 替身形态：<exe> <script> --mode）
func mode() string {
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "--") {
			return a
		}
	}
	return ""
}

func main() {
	// 作为解释器被探测：validatePython 执行 --version
	if mode() == "--version" {
		fmt.Println("Python 3.99.0 (mockcli)")
		return
	}

	switch behavior() {
	case "empty":
		return
	case "nonzero":
		fmt.Fprintln(os.Stderr, "mockcli: 模拟失败（非零退出码）")
		os.Exit(3)
	case "garbage":
		fmt.Println("mockcli: 这行不是 JSON")
		return
	case "slow":
		time.Sleep(30 * time.Second)
	}

	cwd, _ := os.Getwd()
	cfg := os.Getenv("TG_STATION_CONFIG")

	out := func(v map[string]interface{}) {
		v["cwd"] = cwd
		v["config"] = cfg
		b, _ := json.Marshal(v)
		fmt.Println(string(b))
	}

	switch mode() {
	case "--check-session":
		out(map[string]interface{}{"status": "logged_in"})
	case "--list-channels":
		out(map[string]interface{}{
			"status": "ok",
			"channels": []map[string]interface{}{
				{"id": -1001234567890, "title": "mock-channel"},
			},
		})
	case "--qr-only":
		out(map[string]interface{}{"status": "qr", "qr_base64": "bW9jaw=="})
		// 模拟等待扫码：进程保持存活（Go 侧应在拿到第一行 JSON 后立即返回）
		time.Sleep(30 * time.Second)
		out(map[string]interface{}{"status": "success", "username": "mock"})
	default:
		// 常驻监听模式本替身不实现
		os.Exit(9)
	}
}
