//go:build !windows

package main

// hideConsoleWindow 非 Windows 平台无控制台窗口概念，空实现。
func hideConsoleWindow() {}
