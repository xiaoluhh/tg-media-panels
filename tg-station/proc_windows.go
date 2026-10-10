//go:build windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// 平台相关：进程属性、外部独立下载器探测与结束、进程树结束。
// Windows 用 tasklist / taskkill 实现（其它平台见 proc_other.go）。

const (
	// downloaderProcName 后台下载器 exe 名（PyInstaller 单文件：引导进程 + 子进程同名）
	downloaderProcName = "tg_downloader.exe"
	// cliExeName 带控制台的 CLI 下载器 exe 名（可靠输出 stdout JSON）
	cliExeName = "tg_downloader_cli.exe"
)

// childProcAttr 返回子进程属性：隐藏控制台窗口（CREATE_NO_WINDOW）。
// CLI 版下载器是 console 程序，靠这个标志不弹窗，同时 stdout 管道照常可用。
func childProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// externalDownloaderPids 用 tasklist 按进程名查找外部独立下载器（开机自启/手动启动）。
func externalDownloaderPids() []int {
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

// killExternalDownloaders 按进程名结束所有外部下载器实例（含进程树）。
func killExternalDownloaders() error {
	return exec.Command("taskkill", "/IM", downloaderProcName, "/T", "/F").Run()
}

// killProcessTree 结束指定进程及其全部子进程。
func killProcessTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").Run()
}
