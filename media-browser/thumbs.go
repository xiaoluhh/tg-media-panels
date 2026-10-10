package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── 缩略图缓存模块（第二阶段 · 第四部分）────────────────────────────────
//
// 旧实现把「相对路径里的 / 换成 _」当作缓存文件名，会碰撞：
//     a/b.jpg  → a_b.jpg
//     a_b.jpg  → a_b.jpg     ← 两个不同文件共用一张缩略图
// 另外旧实现直接 os.WriteFile / os.Create 写正式文件，进程中断或 FFmpeg
// 失败会留下半张图，而「文件存在」被当成有效缓存。
//
// 本模块提供：
//   1. thumbCacheKey：SHA-256(相对路径 + 大小 + 修改时间) 的无碰撞缓存键；
//   2. 原子写入：临时文件 → 校验非空/可解码 → 同目录 os.Rename 提交；
//   3. 单飞（single-flight）：同一缓存键并发只跑一次 FFmpeg，等待者复用结果；
//   4. 旧命名只读兼容 + 命中即迁移。

const (
	// thumbMaxParallel 同时进行的缩略图生成任务上限（不同缓存键之间）。
	thumbMaxParallel = 4
	// thumbMinBytes 输出文件的最小合理体积，用于识别空文件/半张图。
	thumbMinBytes = 256
	// thumbTempPrefix 临时文件前缀（后缀固定 .jpg，让 FFmpeg 能按扩展名选格式）。
	thumbTempPrefix = ".tmp-thumb-"
)

// thumbTimeout 单张缩略图的生成上限；超时即杀进程并清理临时文件。
// 声明为变量以便测试注入更短的超时（生产默认 20s）。
var thumbTimeout = 20 * time.Second

// renderThumbFn 是缩略图渲染入口，测试可替换以验证并发/失败语义。
var renderThumbFn = renderThumbAtomic

// thumbCmdRunner 执行一次 FFmpeg 渲染（写文件），测试可替换以模拟超时/失败。
var thumbCmdRunner = runFFmpegThumb

// thumbCacheKey 生成稳定、无碰撞的缓存键。
// 输入使用经过安全校验的媒体根相对路径（斜杠分隔）+ 文件大小 + 修改时间，
// 文件内容变化（大小或 mtime 变化）后缓存自动失效。
func thumbCacheKey(rel string, size int64, modTime time.Time) string {
	h := sha256.New()
	h.Write([]byte(filepath.ToSlash(rel)))
	h.Write([]byte{0})
	fmt.Fprintf(h, "%d", size)
	h.Write([]byte{0})
	fmt.Fprintf(h, "%d", modTime.UnixNano())
	return hex.EncodeToString(h.Sum(nil))
}

// thumbCachePath 返回新格式缓存文件路径（绝对路径，位于 thumbsDir 内）。
func thumbCachePath(key string) string {
	return filepath.Join(thumbsDir, key+".thumb.jpg")
}

// legacyThumbCachePath 旧格式缓存路径（"/" → "_"）。
// 只用于兼容读取与一次性迁移，绝不用于写入新文件。
func legacyThumbCachePath(rel string) string {
	safe := strings.ReplaceAll(filepath.ToSlash(rel), "/", "_")
	return filepath.Join(thumbsDir, safe+".thumb.jpg")
}

// thumbCacheFresh 判断缓存文件是否可用：存在、体积合理，且不早于源文件修改时间。
func thumbCacheFresh(cachePath string, src os.FileInfo) bool {
	st, err := os.Stat(cachePath)
	if err != nil || st == nil || !st.Mode().IsRegular() {
		return false
	}
	if st.Size() < thumbMinBytes {
		return false
	}
	return !src.ModTime().After(st.ModTime())
}

// ── 单飞（single-flight）────────────────────────────────────────────────

type thumbFlight struct {
	done chan struct{}
	err  error
}

var thumbFlights = struct {
	mu sync.Mutex
	m  map[string]*thumbFlight
}{m: make(map[string]*thumbFlight)}

// thumbSem 限制同时进行的缩略图生成数量（不同键可并行，但不超过上限）。
var thumbSem = make(chan struct{}, thumbMaxParallel)

// generateThumbOnce 保证同一缓存键同时只有一个生成任务：
//   - 第一个调用者执行 gen；
//   - 其余调用者等待该任务结束并复用其结果（成功或失败）；
//   - 任务结束后立刻清除 in-flight 状态，失败后的后续请求可以重试。
func generateThumbOnce(key string, gen func() error) error {
	thumbFlights.mu.Lock()
	if f, ok := thumbFlights.m[key]; ok {
		thumbFlights.mu.Unlock()
		<-f.done
		return f.err
	}
	f := &thumbFlight{done: make(chan struct{})}
	thumbFlights.m[key] = f
	thumbFlights.mu.Unlock()

	// 保证 in-flight 状态一定被释放（即使 gen panic 也不会永久残留）
	func() {
		defer func() {
			if r := recover(); r != nil {
				f.err = fmt.Errorf("缩略图生成 panic: %v", r)
				log.Printf("[media] 缩略图生成 panic（key=%s）: %v", key[:12], r)
			}
		}()
		f.err = gen()
	}()

	thumbFlights.mu.Lock()
	delete(thumbFlights.m, key)
	thumbFlights.mu.Unlock()
	close(f.done)
	return f.err
}

// thumbInflightCount 返回当前 in-flight 的缓存键数量（测试用）。
func thumbInflightCount() int {
	thumbFlights.mu.Lock()
	defer thumbFlights.mu.Unlock()
	return len(thumbFlights.m)
}

// ── 生成与原子写入 ──────────────────────────────────────────────────────

// ffmpegThumbArgs 构造缩略图参数（视频带 -ss 取样时间点）。
// 输出到明确的 .jpg 文件（配合 -f image2），避免临时文件后缀不匹配导致 FFmpeg 识别不了格式。
func ffmpegThumbArgs(src, out string, isVideo bool, seek string) []string {
	args := []string{"-hide_banner", "-loglevel", "error"}
	if isVideo && seek != "" {
		args = append(args, "-ss", seek)
	}
	args = append(args,
		"-i", src,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale='if(gt(iw,ih),%d,-2)':'if(gt(iw,ih),-2,%d)':flags=lanczos", thumbSize, thumbSize),
		"-q:v", "2",
		"-f", "image2",
		"-y", out,
	)
	return args
}

// runFFmpegThumb 执行一次 FFmpeg 缩略图生成（带超时），输出写入 out。
func runFFmpegThumb(src, out string, isVideo bool, seek string) error {
	ctx, cancel := context.WithTimeout(context.Background(), thumbTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ffmpegPath, ffmpegThumbArgs(src, out, isVideo, seek)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("缩略图生成超时（%s）", thumbTimeout)
	}
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errFfmpegUnavailable
		}
		return fmt.Errorf("ffmpeg 退出异常: %v（%s）", err, clipParam(strings.TrimSpace(stderr.String()), 240))
	}
	return nil
}

var errFfmpegUnavailable = errors.New("ffmpeg 不可用")

// validateThumbFile 校验生成结果：存在、非空、能被解码为图片。
func validateThumbFile(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("输出文件不存在: %w", err)
	}
	if st.Size() < thumbMinBytes {
		return fmt.Errorf("输出文件过小（%d 字节），视为失败", st.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, _, err := image.DecodeConfig(f); err != nil {
		return fmt.Errorf("输出不是可解码的图片: %w", err)
	}
	return nil
}

// renderThumbAtomic 生成缩略图并原子提交到 target。
//
// 步骤：目标目录内建唯一临时文件 → FFmpeg 写入 → 校验退出状态与输出 →
// 同一文件系统内 os.Rename 提交；任何失败都会清理临时文件（含多轮重试产生的
// 全部临时文件），且绝不覆盖已存在的有效缩略图。
func renderThumbAtomic(srcReal, target string, isVideo bool) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("缩略图目录不可写: %w", err)
	}

	// 所有临时文件统一登记，函数返回时一次性清理：
	// 视频缩略图会按多个取样点重试，逐个 Remove 很容易漏删（曾观察到残留临时文件）。
	var tempFiles []string
	defer func() {
		for _, tf := range tempFiles {
			// 已成功提交的临时文件会被 Rename 移走，这里只会得到 ENOENT（无害）
			_ = os.Remove(tf)
		}
	}()
	newTemp := func() (string, error) {
		f, err := os.CreateTemp(dir, thumbTempPrefix+"*.jpg")
		if err != nil {
			return "", fmt.Errorf("无法创建临时文件: %w", err)
		}
		name := f.Name()
		f.Close()
		tempFiles = append(tempFiles, name)
		return name, nil
	}

	if isVideo {
		// 视频：依次尝试多个取样点，任一成功即可
		var lastErr error
		// 末尾的 00:00:00 是首帧兜底：短视频（时长 < 1s）在前几个取样点上都取不到帧，
		// 旧实现会直接判定生成失败，这里退回首帧仍能出图。
		for _, tp := range []string{"00:00:01", "00:00:03", "00:00:05", "00:00:10", "00:00:00"} {
			tmpName, err := newTemp()
			if err != nil {
				return err
			}
			if err := thumbCmdRunner(srcReal, tmpName, true, tp); err != nil {
				lastErr = err
				if errors.Is(err, errFfmpegUnavailable) {
					break
				}
				continue
			}
			if err := validateThumbFile(tmpName); err != nil {
				lastErr = err
				continue
			}
			return commitThumb(tmpName, target)
		}
		// 视频无法用 ffmpeg 出图时，退回 Go 原生解码（部分容器仍可能可解）
		if img := decodeImageFallback(srcReal); img != nil {
			return writeThumbFallback(img, dir, target)
		}
		if lastErr == nil {
			lastErr = errors.New("视频缩略图生成失败")
		}
		return lastErr
	}

	tmpName, err := newTemp()
	if err != nil {
		return err
	}
	var lastErr error
	if err := thumbCmdRunner(srcReal, tmpName, false, ""); err == nil {
		if verr := validateThumbFile(tmpName); verr == nil {
			return commitThumb(tmpName, target)
		} else {
			lastErr = verr
		}
	} else {
		lastErr = err
		if !errors.Is(err, errFfmpegUnavailable) {
			// ffmpeg 存在但处理失败（损坏文件 / 超时等）：继续尝试 Go 原生解码
			log.Printf("[media] ffmpeg 生成图片缩略图失败（%s）: %v", filepath.Base(srcReal), err)
		}
	}
	if img := decodeImageFallback(srcReal); img != nil {
		return writeThumbFallback(img, dir, target)
	}
	if lastErr != nil {
		// 保留底层原因（超时 / 损坏 / 命令缺失），便于日志定位
		return fmt.Errorf("图片缩略图生成失败（原生解码兜底也失败）: %w", lastErr)
	}
	return errors.New("图片缩略图生成失败（ffmpeg 与原生解码均失败）")
}

// commitThumb 原子提交临时文件到正式缓存位置。
func commitThumb(tmpName, target string) error {
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("提交缩略图失败: %w", err)
	}
	return nil
}

// writeThumbFallback 用 Go 原生解码结果写缩略图（无 ffmpeg 时的兼容路径），同样原子提交。
func writeThumbFallback(img image.Image, dir, target string) error {
	tmp, err := os.CreateTemp(dir, thumbTempPrefix+"*.jpg")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := jpeg.Encode(tmp, img, &jpeg.Options{Quality: 85}); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := validateThumbFile(tmpName); err != nil {
		return err
	}
	return commitThumb(tmpName, target)
}

// decodeImageFallback 用 Go 标准库解码图片并缩放到缩略图尺寸。
// 仅作为 ffmpeg 不可用/失败时的兼容回退，返回值可能为 nil。
func decodeImageFallback(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return resizeImage(img, thumbSize)
}

// removeThumbCache 删除某个媒体文件对应的缩略图（新旧两种命名都清）。
// rel 必须是经过 safeMediaPath 校验得到的相对路径。
func removeThumbCache(rel string, size int64, modTime time.Time) {
	candidates := []string{
		thumbCachePath(thumbCacheKey(rel, size, modTime)),
		legacyThumbCachePath(rel),
	}
	for _, p := range candidates {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			log.Printf("[media] 清理缩略图失败 %s: %v", filepath.Base(p), err)
		}
	}
}

// ensureThumb 返回可用的缩略图缓存路径；必要时生成（同键单飞、原子写入）。
// srcInfo 是源文件（解析链接后）的 stat 结果，用于缓存新鲜度与缓存键。
func ensureThumb(mp MediaPath, srcInfo os.FileInfo) (string, error) {
	if srcInfo == nil {
		return "", errors.New("缺少源文件信息")
	}
	key := thumbCacheKey(mp.Rel, srcInfo.Size(), srcInfo.ModTime())
	target := thumbCachePath(key)

	if thumbCacheFresh(target, srcInfo) {
		return target, nil
	}

	// 旧命名兼容读取：命中且新鲜则原子迁移到新键（迁移失败不影响本次服务）
	legacy := legacyThumbCachePath(mp.Rel)
	if thumbCacheFresh(legacy, srcInfo) {
		if err := os.Rename(legacy, target); err == nil {
			return target, nil
		}
		return legacy, nil
	}

	isVideo := videoExts[strings.ToLower(filepath.Ext(mp.Rel))]
	err := generateThumbOnce(key, func() error {
		// 双检：等待期间可能已有其他请求完成了生成
		if thumbCacheFresh(target, srcInfo) {
			return nil
		}
		return renderThumbFn(mp.Real, target, isVideo)
	})
	if err != nil {
		return "", err
	}
	return target, nil
}

// generateAllThumbnails 基于当前索引批量补齐缩略图（不同键并行，上限 thumbMaxParallel）。
// 必须在扫描完成（且没有待执行的补扫）之后调用，否则会把尚未入索引的新文件当作不存在。
func generateAllThumbnails() {
	indexMutex.RLock()
	files := make([]*MediaFile, len(mediaIndex))
	copy(files, mediaIndex)
	indexMutex.RUnlock()

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		generated int
		skipped   int
		failed    int
	)
	jobs := make(chan *MediaFile)

	workerCount := thumbMaxParallel
	if workerCount > len(files) {
		workerCount = len(files)
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for mf := range jobs {
				mp, err := safeMediaPath(mediaDir, mf.Path)
				if err != nil {
					if !errors.Is(err, ErrPathMissing) {
						logPathReject("thumbnail", mf.Path, err)
					}
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				srcInfo, err := os.Stat(mp.Real)
				if err != nil {
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				key := thumbCacheKey(mp.Rel, srcInfo.Size(), srcInfo.ModTime())
				if thumbCacheFresh(thumbCachePath(key), srcInfo) {
					mu.Lock()
					skipped++
					mu.Unlock()
					continue
				}
				if _, err := ensureThumb(mp, srcInfo); err != nil {
					log.Printf("[media] 缩略图生成失败 %s: %v", mf.Path, err)
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				mu.Lock()
				generated++
				mu.Unlock()
			}
		}()
	}
	for _, mf := range files {
		jobs <- mf
	}
	close(jobs)
	wg.Wait()

	log.Printf("缩略图预生成完成：新生成 %d 个，跳过 %d 个，失败 %d 个", generated, skipped, failed)
}
