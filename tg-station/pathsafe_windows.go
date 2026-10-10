//go:build windows

package main

import (
	"os"
	"syscall"
)

// FILE_ATTRIBUTE_REPARSE_POINT：符号链接、junction、mount point 都带该属性。
// Windows 上 filepath.EvalSymlinks 只解析真符号链接，对 junction 返回原路径，
// 因此必须单独识别，否则「经由 junction 访问媒体目录之外」会成为绕过口子。
const fileAttributeReparsePoint = 0x400

// isReparsePoint 判断路径是否为符号链接 / 重解析点。
// 任何无法判定的情况都按「可疑」处理（返回 true），由调用方拒绝。
func isReparsePoint(path string, info os.FileInfo) bool {
	if info != nil && info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return true
	}
	return attrs&fileAttributeReparsePoint != 0
}
