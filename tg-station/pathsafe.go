package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ── 媒体文件路径安全校验（第二阶段 · 第一部分）────────────────────────────
//
// 背景：旧实现用 strings.HasPrefix(absFull, absMedia) 判断「是否位于媒体根目录内」。
// 该判断有两个致命缺陷：
//   1. 前缀比较不是目录边界比较：/data/media2/a.mp4 能通过 /data/media 的检查；
//   2. 完全不处理符号链接 / 重解析点：媒体目录里指向 /etc/passwd 的链接会被照常读取/删除。
//
// 本文件提供唯一的安全校验入口 safeMediaPath（以及删除专用的 safeDeleteTarget）。
// 所有媒体的读取、播放、缩略图生成、删除都必须先经过它，禁止在别处再写路径判断。
//
// 平台差异：非 Windows 平台 EvalSymlinks 即可解析全部链接；Windows 上 junction /
// mount point（重解析点）不会被 EvalSymlinks 解析，因此额外逐段检测重解析点
// （isReparsePoint 的平台实现）：
//   - 解析后仍在真实根目录内 → 放行（媒体目录内部的链接可用）；
//   - 解析不到（无法证明目标位置，如 junction）→ 拒绝。

var (
	// ErrPathOutside 请求路径无法被证明位于媒体根目录内（越界、非法、符号链接逃逸）。
	ErrPathOutside = errors.New("路径不在媒体目录内")
	// ErrPathMissing 目标文件不存在。
	ErrPathMissing = errors.New("文件不存在")
)

// MediaPath 是一次成功校验的结果。
type MediaPath struct {
	// Full 规范化后的绝对路径（不解析符号链接），用于 Lstat / os.Remove。
	Full string
	// Real 解析符号链接后的真实绝对路径，用于读取、播放、ffmpeg。
	Real string
	// Rel 相对媒体根目录的斜杠分隔路径，用于缩略图缓存键。
	Rel string
	// Symlink 说明 Full 本身是否为符号链接（删除时只删链接，不跟随）。
	Symlink bool
	// Info Full 的 Lstat 结果（调用方复用以避免二次 stat）。
	Info os.FileInfo
}

// ── 媒体根目录真实路径缓存 ───────────────────────────────────────────────
//
// EvalSymlinks 在 Windows 上开销不低（逐段查询），而媒体根目录在一次运行中
// 基本不变，因此缓存「输入根 → 真实根」的映射（键为绝对路径，避免测试之间串味）。

var mediaRootCache struct {
	mu   sync.RWMutex
	real map[string]string
}

func mediaRootReal(mediaRoot string) (string, error) {
	abs, err := filepath.Abs(mediaRoot)
	if err != nil {
		return "", fmt.Errorf("%w: 媒体目录无法解析为绝对路径", ErrPathOutside)
	}
	abs = filepath.Clean(abs)

	mediaRootCache.mu.RLock()
	if mediaRootCache.real != nil {
		if real, ok := mediaRootCache.real[abs]; ok {
			mediaRootCache.mu.RUnlock()
			return real, nil
		}
	}
	mediaRootCache.mu.RUnlock()

	real, err := resolveReparsePoints(abs)
	if err != nil {
		// 根目录不存在 / 无法访问 / 链接环：不能安全校验，必须拒绝，
		// 绝不能退回未解析的路径。
		return "", fmt.Errorf("%w: 媒体目录不可访问或无法解析 (%v)", ErrPathOutside, err)
	}
	real = filepath.Clean(real)

	mediaRootCache.mu.Lock()
	if mediaRootCache.real == nil {
		mediaRootCache.real = make(map[string]string, 2)
	}
	mediaRootCache.real[abs] = real
	mediaRootCache.mu.Unlock()
	return real, nil
}

// resetMediaRootCache 清空根目录真实路径缓存（测试用）。
func resetMediaRootCache() {
	mediaRootCache.mu.Lock()
	mediaRootCache.real = nil
	mediaRootCache.mu.Unlock()
}

// isWithin 判断 abs 是否位于 root 目录之内（按目录边界，不是字符串前缀）。
// root/abs 都必须是规范化后的绝对路径。
func isWithin(root, abs string) bool {
	if root == "" || abs == "" {
		return false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return isLocalRel(rel)
}

// isLocalRel 判断 rel 是否为一个「位于当前目录之内、不向上逃逸」的相对路径。
func isLocalRel(rel string) bool {
	if rel == "" || rel == "." {
		return false
	}
	if filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// maxReparseDepth 解析链接的最大迭代次数（防链接环导致的死循环）。
const maxReparseDepth = 32

// firstReparsePoint 自卷根开始逐段检查，返回第一个符号链接 / 重解析点的完整路径；
// 没有则返回 ""。任何一段不存在时返回该段的错误（调用方可判断 os.IsNotExist）。
func firstReparsePoint(path string) (string, error) {
	vol := filepath.VolumeName(path)
	cur := vol + string(filepath.Separator)
	rest := strings.TrimPrefix(path[len(vol):], string(filepath.Separator))
	for _, seg := range strings.Split(filepath.ToSlash(rest), "/") {
		if seg == "" {
			continue
		}
		cur = filepath.Join(cur, filepath.FromSlash(seg))
		info, err := os.Lstat(cur)
		if err != nil {
			return "", err
		}
		if isReparsePoint(cur, info) {
			return cur, nil
		}
	}
	return "", nil
}

// resolveReparsePoints 反复解析路径上的符号链接 / 重解析点，直到路径不再含链接。
//
// 与 filepath.EvalSymlinks 的差别：Windows 上 junction / mount point 不会被
// EvalSymlinks 解析，但本函数通过 os.Readlink 能拿到其目标，从而统一处理
// 符号链接与 junction（两平台行为一致）。超过 maxReparseDepth 次仍未收敛
// （如链接环）时报错，由调用方拒绝。
func resolveReparsePoints(path string) (string, error) {
	cur := filepath.Clean(path)
	for depth := 0; depth < maxReparseDepth; depth++ {
		rp, err := firstReparsePoint(cur)
		if err != nil {
			return "", err
		}
		if rp == "" {
			return cur, nil
		}
		target, err := os.Readlink(rp)
		if err != nil {
			return "", fmt.Errorf("无法读取链接目标: %w", err)
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(rp), target)
		}
		rest, err := filepath.Rel(rp, cur)
		if err != nil {
			return "", fmt.Errorf("无法计算链接剩余路径: %w", err)
		}
		if rest == "." || rest == "" {
			cur = filepath.Clean(target)
		} else {
			cur = filepath.Clean(filepath.Join(target, rest))
		}
	}
	return "", fmt.Errorf("链接层数超过上限 %d（可能存在链接环）", maxReparseDepth)
}

// hasDotDotSegment 检查请求路径中是否存在 ".." 段（在 Clean 之前判断，
// 因为 Clean 会把 a/../b 悄悄压成 b，从而掩盖穿越意图）。
func hasDotDotSegment(requested string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(requested), "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// windowsReservedNames 是 Windows 保留设备名：这些名字（含带扩展名的形式）
// 会被系统解释为设备而不是媒体文件，必须拒绝。
var windowsReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// unsafeSegmentReason 返回该路径段被拒绝的原因（"" 表示合法）。
// 仅 Windows 上启用这些额外规则：Windows 会静默删除段尾的空格与点、
// 把 ":" 解释为 NTFS 数据流、把 NUL 之类解释为设备，这些都会让
// 「校验用的路径」与「实际打开的文件」不一致。
func unsafeSegmentReason(seg string) string {
	if seg == "." {
		// "." 段只在 Join/Clean 中表示「当前目录」，无风险
		return ""
	}
	if runtime.GOOS != "windows" {
		return ""
	}
	if strings.HasSuffix(seg, " ") || strings.HasSuffix(seg, ".") {
		return "段以空格或点结尾（Windows 会静默改写路径）"
	}
	if strings.ContainsAny(seg, `:<>"|?*`) {
		return "段含非法字符（:, <, >, \", |, ?, * 之一）"
	}
	base := seg
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if windowsReservedNames[strings.ToUpper(base)] {
		return "段使用了 Windows 保留设备名"
	}
	return ""
}

// validateRequestSegments 逐段检查请求路径；返回被拒绝的原因（"" 表示通过）。
func validateRequestSegments(requested string) string {
	if hasDotDotSegment(requested) {
		return "路径包含 .. 穿越段"
	}
	for _, seg := range strings.Split(filepath.ToSlash(requested), "/") {
		if seg == "" {
			continue
		}
		if reason := unsafeSegmentReason(seg); reason != "" {
			return reason
		}
	}
	return ""
}

// safeMediaPath 校验 requested 并返回媒体根目录内的安全路径。
//
// requested 是前端传来的相对路径（斜杠分隔，可含 %XX 编码）。以下一律拒绝：
//   - 空路径、含 NUL、含反斜杠（跨平台一致性；Windows 上反斜杠即分隔符）
//   - 绝对路径、盘符路径（C:\...）、UNC 路径（\\server\share）、URL 形式（file://）
//   - 任何 ".." 段（目录穿越），无论最终是否仍在根内
//   - 规范化后逃出媒体根目录的路径，以及同前缀的兄弟目录（/data/media2）
//   - 符号链接解析后真实路径落在媒体根目录之外的路径
//   - 无法解析（根目录不可访问、链接损坏等）而无法证明安全的路径
//
// 返回的 MediaPath.Real 是解析过符号链接的真实路径，调用方应使用它去
// 读取/播放/ffmpeg；删除请使用 safeDeleteTarget。
func safeMediaPath(mediaRoot, requested string) (MediaPath, error) {
	var out MediaPath

	if strings.TrimSpace(mediaRoot) == "" {
		return out, fmt.Errorf("%w: 未配置媒体目录", ErrPathOutside)
	}
	if strings.TrimSpace(requested) == "" {
		return out, fmt.Errorf("%w: 空路径", ErrPathOutside)
	}
	if strings.ContainsRune(requested, 0) {
		return out, fmt.Errorf("%w: 路径含非法字符", ErrPathOutside)
	}

	// 前端用 encodeURIComponent 逐段编码，这里解码回原始路径；
	// 解码失败按原样处理（下面还有完整的合法性检查兜底）。
	if decoded, err := url.PathUnescape(requested); err == nil {
		requested = decoded
	}

	// 反斜杠：Windows 上是分隔符（会被 ToSlash 转换），Linux 上是普通字符。
	// 为保证两平台一致的校验结论，一律拒绝。
	if strings.Contains(requested, "\\") {
		return out, fmt.Errorf("%w: 路径不能包含反斜杠", ErrPathOutside)
	}
	if strings.Contains(requested, "://") {
		return out, fmt.Errorf("%w: 不接受 URL 形式路径", ErrPathOutside)
	}

	if reason := validateRequestSegments(requested); reason != "" {
		return out, fmt.Errorf("%w: %s", ErrPathOutside, reason)
	}

	req := filepath.FromSlash(requested)
	if filepath.IsAbs(req) || filepath.VolumeName(req) != "" {
		return out, fmt.Errorf("%w: 不接受绝对路径或盘符路径", ErrPathOutside)
	}

	rootAbs, err := filepath.Abs(mediaRoot)
	if err != nil {
		return out, fmt.Errorf("%w: 媒体目录无法解析为绝对路径", ErrPathOutside)
	}
	rootAbs = filepath.Clean(rootAbs)
	rootReal, err := mediaRootReal(mediaRoot)
	if err != nil {
		return out, err
	}

	// 规范化：Join 会 Clean，但 Clean 本身不保证边界 —— 紧随其后的 Rel 才是边界检查。
	full := filepath.Clean(filepath.Join(rootAbs, req))
	rel, err := filepath.Rel(rootAbs, full)
	if err != nil || !isLocalRel(rel) {
		return out, fmt.Errorf("%w: 路径越出媒体目录", ErrPathOutside)
	}

	info, err := os.Lstat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return MediaPath{Full: full, Rel: filepath.ToSlash(rel)},
				fmt.Errorf("%w: %v", ErrPathMissing, err)
		}
		return out, fmt.Errorf("%w: 无法访问目标 (%v)", ErrPathOutside, err)
	}
	out.Full = full
	out.Rel = filepath.ToSlash(rel)
	out.Info = info
	out.Symlink = info.Mode()&os.ModeSymlink != 0

	// 解析路径链上的符号链接 / 重解析点（含 Windows junction），
	// 解析结果必须仍位于真实媒体根目录之内。
	real, rpErr := resolveReparsePoints(full)
	if rpErr != nil {
		if os.IsNotExist(rpErr) {
			return MediaPath{Full: full, Rel: filepath.ToSlash(rel)},
				fmt.Errorf("%w: %v", ErrPathMissing, rpErr)
		}
		return MediaPath{}, fmt.Errorf("%w: 路径无法安全解析 (%v)", ErrPathOutside, rpErr)
	}
	real = filepath.Clean(real)
	if !isWithin(rootReal, real) {
		return MediaPath{}, fmt.Errorf("%w: 链接指向媒体目录之外", ErrPathOutside)
	}
	out.Real = real
	return out, nil
}

// safeDeleteTarget 在 safeMediaPath 基础上针对删除额外收紧：
//   - 只允许删除普通文件（或指向根内普通文件的符号链接）；
//   - 目录、设备、管道等一律拒绝；
//   - 符号链接只删除链接本身（os.Remove 不跟随链接），绝不删除链接指向的文件。
func safeDeleteTarget(mediaRoot, requested string) (MediaPath, error) {
	mp, err := safeMediaPath(mediaRoot, requested)
	if err != nil {
		// 目标不存在时仍返回已规范化的 Full，便于调用方做一致性判断
		return mp, err
	}
	if mp.Info.IsDir() {
		return MediaPath{}, fmt.Errorf("%w: 目标不是文件", ErrPathOutside)
	}
	if !mp.Info.Mode().IsRegular() && !mp.Symlink {
		return MediaPath{}, fmt.Errorf("%w: 目标类型不允许删除", ErrPathOutside)
	}
	return mp, nil
}

// writeMediaPathError 把安全校验错误映射为 HTTP 响应：
// 目标不存在 → 404；其余（越界、链接逃逸、非法路径、不可解析）→ 403。
// 响应体只包含固定文本，绝不回显服务器绝对路径。
func writeMediaPathError(w http.ResponseWriter, r *http.Request, op, requested string, err error) {
	if errors.Is(err, ErrPathMissing) {
		http.NotFound(w, r)
		return
	}
	logPathReject(op, requested, err)
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// logPathReject 记录被拒绝的路径请求（仅服务端日志，绝不回显给客户端）。
func logPathReject(op, requested string, err error) {
	log.Printf("[station] %s 拒绝访问 %q: %v", op, clip(requested, 200), err)
}
