//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 平台相关：进程属性、外部独立下载器探测与结束、进程树结束。
// 非 Windows（Linux/树莓派、macOS）不能依赖 tasklist：这里扫描 /proc 命令行，
// 同时覆盖「打包后的 tg_downloader」与「python + tg_auto_download.py」两种形态。

const (
	// downloaderProcName 后台下载器可执行文件名（Unix 无 .exe 后缀）
	downloaderProcName = "tg_downloader"
	// cliExeName 带控制台的 CLI 下载器可执行文件名
	cliExeName = "tg_downloader_cli"
	// downloaderScriptName 旧模式（python + 脚本）下的脚本名
	downloaderScriptName = "tg_auto_download.py"
)

// childProcAttr 非 Windows 平台无需隐藏窗口。
func childProcAttr() *syscall.SysProcAttr {
	return nil
}

// externalDownloaderPids 扫描 /proc 找出外部独立下载器进程（排除自身）。
func externalDownloaderPids() []int {
	self := os.Getpid()
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid == self {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil || len(raw) == 0 {
			continue
		}
		cmdline := strings.ReplaceAll(string(raw), "\x00", " ")
		if strings.Contains(cmdline, downloaderProcName) || strings.Contains(cmdline, downloaderScriptName) {
			pids = append(pids, pid)
		}
	}
	return pids
}

// killExternalDownloaders 结束扫描到的外部下载器实例：先 SIGTERM，超时后 SIGKILL。
func killExternalDownloaders() error {
	pids := externalDownloaderPids()
	if len(pids) == 0 {
		return nil
	}
	for _, pid := range pids {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(externalDownloaderPids()) == 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	for _, pid := range externalDownloaderPids() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	return nil
}

// killProcessTree 结束指定进程（Unix 下无进程树语义，直接 SIGKILL 该进程）。
func killProcessTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	return syscall.Kill(pid, syscall.SIGKILL)
}
