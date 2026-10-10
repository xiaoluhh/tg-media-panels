package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── 媒体元数据探测（第二阶段 · 第三部分）────────────────────────────────
//
// 1. 优先用 ffprobe（只输出 duration，不做任何转码/媒体数据输出）；
//    没有 ffprobe 时回退到原有的 ffmpeg -i + stderr 解析，行为完全兼容。
// 2. 所有外部进程都带超时（exec.CommandContext）：单个损坏文件 / 卡住的进程
//    不会让整轮扫描永久阻塞。
// 3. 时长缓存带「大小 + 修改时间」校验：文件内容变化后自动重新探测。

const (
	// probeTimeout 单个文件元数据探测的超时时间。
	probeTimeout = 15 * time.Second
	// probeWorkers 并发探测数（与原实现一致的并发限制，不放大）。
	probeWorkers = 4
)

// ffprobePath 在启动时定位；为空表示不可用，直接走 ffmpeg 回退。
var ffprobePath string

// findFFprobe 定位 ffprobe：与 findFFmpeg 相同的候选顺序（exe 目录 → 工作目录 →
// 系统路径 → PATH 兜底）。
func findFFprobe() string {
	names := []string{"ffprobe", "ffprobe.exe"}

	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}
	for _, dir := range dirs {
		for _, name := range names {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}
	for _, p := range []string{"/usr/bin/ffprobe", "/usr/local/bin/ffprobe", "/bin/ffprobe"} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	// 最后交给 PATH（ffprobe 与 ffmpeg 通常同目录）
	if p, err := exec.LookPath("ffprobe"); err == nil {
		return p
	}
	return ""
}

// ── 时长缓存 ────────────────────────────────────────────────────────────

type durationEntry struct {
	Size     int64
	ModTime  int64 // UnixNano
	Duration float64
}

var durationCache = struct {
	mu sync.RWMutex
	m  map[string]durationEntry
}{m: make(map[string]durationEntry)}

// durationCacheGet 命中条件：路径相同且大小、修改时间都与缓存记录一致。
func durationCacheGet(rel string, size int64, modTime time.Time) (float64, bool) {
	durationCache.mu.RLock()
	e, ok := durationCache.m[rel]
	durationCache.mu.RUnlock()
	if !ok {
		return 0, false
	}
	if e.Size != size || e.ModTime != modTime.UnixNano() {
		return 0, false
	}
	return e.Duration, true
}

func durationCachePut(rel string, size int64, modTime time.Time, d float64) {
	durationCache.mu.Lock()
	durationCache.m[rel] = durationEntry{Size: size, ModTime: modTime.UnixNano(), Duration: d}
	durationCache.mu.Unlock()
}

// durationCacheLen 返回缓存条目数（测试用）。
func durationCacheLen() int {
	durationCache.mu.RLock()
	defer durationCache.mu.RUnlock()
	return len(durationCache.m)
}

// resetDurationCache 清空缓存（测试用）。
func resetDurationCache() {
	durationCache.mu.Lock()
	durationCache.m = make(map[string]durationEntry)
	durationCache.mu.Unlock()
}

// ── 探测实现 ────────────────────────────────────────────────────────────

var durationRe = regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+(?:\.\d+)?)`)

// parseDurationSeconds 解析 ffprobe 输出的时长文本。
// 明确拒绝空值、N/A、NaN、Inf、负数与解析失败的结果。
func parseDurationSeconds(text string) (float64, error) {
	txt := strings.TrimSpace(text)
	if txt == "" {
		return 0, errors.New("时长为空")
	}
	if strings.EqualFold(txt, "N/A") || strings.EqualFold(txt, "nan") {
		return 0, fmt.Errorf("时长不可用: %q", clipParam(txt, 32))
	}
	d, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		return 0, fmt.Errorf("时长不是数字: %q", clipParam(txt, 32))
	}
	if math.IsNaN(d) || math.IsInf(d, 0) || d <= 0 {
		return 0, fmt.Errorf("时长非法: %v", d)
	}
	return d, nil
}

// probeWithFFprobe 用 ffprobe 读取时长（秒）。返回 (时长, 是否成功)。
// 参数只取 format=duration，避免输出媒体流数据或触发转码。
func probeWithFFprobe(path string) (float64, bool) {
	if ffprobePath == "" {
		return 0, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("[media] ffprobe 超时（%s）: %s", probeTimeout, filepath.Base(path))
		return 0, false
	}
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Printf("[media] ffprobe 不可用，后续改用 ffmpeg 探测")
			ffprobePath = ""
			return 0, false
		}
		log.Printf("[media] ffprobe 探测失败 %s: %v（%s）",
			filepath.Base(path), err, clipParam(strings.TrimSpace(stderr.String()), 160))
		return 0, false
	}
	d, perr := parseDurationSeconds(stdout.String())
	if perr != nil {
		log.Printf("[media] ffprobe 输出无法解析 %s: %v", filepath.Base(path), perr)
		return 0, false
	}
	return d, true
}

// probeWithFFmpeg 旧方式：ffmpeg -i（无输出，必然非零退出码）解析 stderr 的 Duration 行。
func probeWithFFmpeg(path string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffmpegPath, "-hide_banner", "-i", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// -i 不指定输出会以非零码退出，属预期；只需解析 stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("[media] ffmpeg 探测超时（%s）: %s", probeTimeout, filepath.Base(path))
		return 0, false
	}
	if errors.Is(runErr, exec.ErrNotFound) {
		log.Printf("[media] ffmpeg 不可用，跳过时长探测")
		return 0, false
	}
	m := durationRe.FindStringSubmatch(stderr.String())
	if m == nil {
		log.Printf("[media] ffmpeg 未输出可解析的时长 %s", filepath.Base(path))
		return 0, false
	}
	h, _ := strconv.ParseFloat(m[1], 64)
	mi, _ := strconv.ParseFloat(m[2], 64)
	s, _ := strconv.ParseFloat(m[3], 64)
	d := h*3600 + mi*60 + s
	if math.IsNaN(d) || math.IsInf(d, 0) || d <= 0 {
		return 0, false
	}
	return d, true
}

// probeDuration 探测媒体时长（秒）：ffprobe 优先，失败回退 ffmpeg，全部失败返回 0。
func probeDuration(path string) float64 {
	if d, ok := probeWithFFprobe(path); ok {
		return d
	}
	if d, ok := probeWithFFmpeg(path); ok {
		return d
	}
	return 0
}

// fillDurations 并发补齐 files 中缺失的视频时长（秒），带缓存与超时。
func fillDurations(files []*MediaFile) {
	var pending []*MediaFile
	for _, f := range files {
		if f.IsVideo && f.Duration <= 0 {
			pending = append(pending, f)
		}
	}
	if len(pending) == 0 {
		return
	}
	workerCount := probeWorkers
	if len(pending) < workerCount {
		workerCount = len(pending)
	}
	jobs := make(chan *MediaFile, len(pending))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				// 缓存命中（路径 + 大小 + 修改时间一致）直接复用，不再启动进程
				if d, ok := durationCacheGet(f.Path, f.Size, f.Date); ok && d > 0 {
					f.Duration = d
					continue
				}
				full := filepath.Join(mediaDir, filepath.FromSlash(f.Path))
				d := probeDuration(full)
				if d > 0 {
					f.Duration = d
					durationCachePut(f.Path, f.Size, f.Date, d)
				}
			}
		}()
	}
	for _, f := range pending {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
}
