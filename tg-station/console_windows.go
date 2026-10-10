//go:build windows

package main

import "syscall"

// hideConsoleWindow 隐藏当前进程的控制台窗口（后台运行）。
// 注意：ShowWindow 属于 user32.dll，kernel32.dll 中没有该导出函数；
// 用错 DLL 会在无控制台环境（开机自启/服务启动）下 panic。
func hideConsoleWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return // 无控制台窗口（如自启环境），无需隐藏
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	showWindow := user32.NewProc("ShowWindow")
	if err := showWindow.Find(); err != nil {
		return // 函数查找失败也不 panic，后台运行不受影响
	}
	showWindow.Call(hwnd, 0) // SW_HIDE = 0
}
