//go:build !windows

package main

import "syscall"

// childProcAttr 非 Windows 平台无需隐藏窗口。
func childProcAttr() *syscall.SysProcAttr {
	return nil
}
