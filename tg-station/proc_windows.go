//go:build windows

package main

import "syscall"

// childProcAttr 返回子进程属性：隐藏控制台窗口（CREATE_NO_WINDOW）。
func childProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
