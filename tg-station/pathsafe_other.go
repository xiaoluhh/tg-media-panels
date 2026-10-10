//go:build !windows

package main

import "os"

// isReparsePoint 在非 Windows 平台上，重解析点即符号链接：
// filepath.EvalSymlinks 能完整解析，因此只需检查链接标志。
func isReparsePoint(_ string, info os.FileInfo) bool {
	return info != nil && info.Mode()&os.ModeSymlink != 0
}
