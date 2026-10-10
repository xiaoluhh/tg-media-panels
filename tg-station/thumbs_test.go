package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ── 缩略图测试（第二阶段 · 第四部分）────────────────────────────────────
//
// 覆盖：旧命名碰撞 → 新键不碰撞 / 内容变化失效 / 旧缓存迁移 /
// 同键并发只生成一次 / 不同键可并行（不超过上限）/ 失败后可重试 /
// 原子提交与临时文件清理 / 超时清理 / 源路径穿越被拒。

// makeJPEGBytes 生成一张足够大的 JPEG（> thumbMinBytes）。
func makeJPEGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("生成测试 JPEG 失败: %v", err)
	}
	return buf.Bytes()
}

func writeJPEG(t *testing.T, path string) []byte {
	t.Helper()
	data := makeJPEGBytes(t)
	writeFileT(t, path, string(data))
	return data
}

// TestLegacyThumbKeyCollision 复现旧命名算法（"/" → "_"）的碰撞，
// 并确认新的哈希键不会碰撞。
func TestLegacyThumbKeyCollision(t *testing.T) {
	oldThumbs := thumbsDir
	thumbsDir = t.TempDir()
	defer func() { thumbsDir = oldThumbs }()

	a, b := "a/b.jpg", "a_b.jpg"
	if legacyThumbCachePath(a) != legacyThumbCachePath(b) {
		t.Fatalf("测试前提不成立：旧算法应发生碰撞（%q vs %q）", legacyThumbCachePath(a), legacyThumbCachePath(b))
	}

	mt := time.Unix(1700000000, 0)
	if thumbCacheKey(a, 100, mt) == thumbCacheKey(b, 100, mt) {
		t.Fatalf("新哈希键不应碰撞: %q 与 %q 生成了相同缓存键", a, b)
	}
	// 同大小不同路径、同路径不同大小/时间都必须产生不同键
	if thumbCacheKey(a, 100, mt) == thumbCacheKey(a, 101, mt) {
		t.Error("大小变化应产生不同缓存键（缓存需要失效）")
	}
	if thumbCacheKey(a, 100, mt) == thumbCacheKey(a, 100, mt.Add(time.Second)) {
		t.Error("修改时间变化应产生不同缓存键（缓存需要失效）")
	}
	// 键长稳定（sha256 十六进制）
	if len(thumbCacheKey(a, 1, mt)) != 64 {
		t.Errorf("缓存键应为 64 位十六进制 sha256，实际 %d", len(thumbCacheKey(a, 1, mt)))
	}
}

// TestLegacyThumbMigration 旧命名缓存应被读取并原子迁移到新命名。
func TestLegacyThumbMigration(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	src := filepath.Join(dir, "sub", "pic.jpg")
	data := writeJPEG(t, src)
	_ = data
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// 写一份旧命名缓存（内容为合法 JPEG，时间晚于源文件）
	legacy := legacyThumbCachePath("sub/pic.jpg")
	cacheData := makeJPEGBytes(t)
	if err := os.WriteFile(legacy, cacheData, 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(legacy, future, future); err != nil {
		t.Fatal(err)
	}

	mp, err := safeMediaPath(mediaDir, "sub/pic.jpg")
	if err != nil {
		t.Fatalf("safeMediaPath 失败: %v", err)
	}
	srcInfo, err := os.Stat(mp.Real)
	if err != nil {
		t.Fatal(err)
	}

	// 渲染必须完全不被触发（旧缓存新鲜）
	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var calls int32
	renderThumbFn = func(string, string, bool) error {
		atomic.AddInt32(&calls, 1)
		return fmt.Errorf("不应被调用")
	}

	got, err := ensureThumb(mp, srcInfo)
	if err != nil {
		t.Fatalf("ensureThumb 失败: %v", err)
	}
	if calls != 0 {
		t.Errorf("命中旧缓存时不应重新生成，实际调用 %d 次", calls)
	}
	// 已迁移到新键
	want := thumbCachePath(thumbCacheKey(mp.Rel, srcInfo.Size(), srcInfo.ModTime()))
	if got != want {
		t.Errorf("应迁移到新缓存键 %s，实际 %s", filepath.Base(want), filepath.Base(got))
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("迁移后的缓存文件不存在: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("旧命名缓存应已被移走")
	}
}

// TestThumbSingleFlight 同一缓存键的并发请求只触发一次生成，且等待者复用结果。
func TestThumbSingleFlight(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	writeFileT(t, filepath.Join(dir, "a.jpg"), "source-bytes")

	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var calls int32
	jpg := makeJPEGBytes(t)
	renderThumbFn = func(src, target string, isVideo bool) error {
		atomic.AddInt32(&calls, 1)
		time.Sleep(80 * time.Millisecond) // 拉长窗口，确保并发请求会重叠
		return os.WriteFile(target, jpg, 0o644)
	}

	mp, err := safeMediaPath(mediaDir, "a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	srcInfo, err := os.Stat(mp.Real)
	if err != nil {
		t.Fatal(err)
	}

	const n = 24
	var wg sync.WaitGroup
	errs := make([]error, n)
	paths := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			paths[i], errs[i] = ensureThumb(mp, srcInfo)
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("同一缓存键并发请求应只生成一次，实际 %d 次", got)
	}
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("第 %d 个请求失败: %v", i, errs[i])
		}
		if paths[i] != paths[0] {
			t.Fatalf("等待者应复用同一缓存路径，%s vs %s", paths[0], paths[i])
		}
	}
	if thumbInflightCount() != 0 {
		t.Errorf("生成结束后 in-flight 状态应被释放，实际 %d", thumbInflightCount())
	}
}

// TestThumbFailureAllowsRetry 生成失败时等待者拿到错误，且后续请求可以重试。
func TestThumbFailureAllowsRetry(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	writeFileT(t, filepath.Join(dir, "a.jpg"), "source")

	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var calls int32
	renderThumbFn = func(src, target string, isVideo bool) error {
		if atomic.AddInt32(&calls, 1) == 1 {
			return fmt.Errorf("模拟首次失败")
		}
		return os.WriteFile(target, makeJPEGBytes(t), 0o644)
	}

	mp, err := safeMediaPath(mediaDir, "a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	srcInfo, _ := os.Stat(mp.Real)

	if _, err := ensureThumb(mp, srcInfo); err == nil {
		t.Fatal("首次生成失败应返回错误")
	}
	if thumbInflightCount() != 0 {
		t.Fatalf("失败后 in-flight 状态必须释放，否则后续请求永久阻塞")
	}
	if _, err := ensureThumb(mp, srcInfo); err != nil {
		t.Fatalf("失败后应允许重试，实际: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("期望两次生成尝试，实际 %d", calls)
	}
}

// TestThumbParallelLimitAndDifferentKeys generateAllThumbnails 对不同的键并行生成，
// 但并发数不超过 thumbMaxParallel（不因一把全局锁而串行）。
func TestThumbParallelLimitAndDifferentKeys(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	var files []*MediaFile
	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, fmt.Sprintf("p%02d.jpg", i))
		writeFileT(t, p, fmt.Sprintf("src-%d", i))
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, &MediaFile{Path: fmt.Sprintf("p%02d.jpg", i), Date: st.ModTime(), Size: st.Size()})
	}
	rebuildIndex(files)

	oldRender := renderThumbFn
	defer func() { renderThumbFn = oldRender }()
	var cur, peak int32
	jpg := makeJPEGBytes(t)
	renderThumbFn = func(src, target string, isVideo bool) error {
		c := atomic.AddInt32(&cur, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if c <= p || atomic.CompareAndSwapInt32(&peak, p, c) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond)
		err := os.WriteFile(target, jpg, 0o644)
		atomic.AddInt32(&cur, -1)
		return err
	}

	generateAllThumbnails()

	if peak < 2 {
		t.Errorf("不同缓存键应能并行生成（峰值并发 = %d）", peak)
	}
	if peak > int32(thumbMaxParallel) {
		t.Errorf("并发峰值 %d 超过上限 %d", peak, thumbMaxParallel)
	}
	// 每个文件都应有缓存文件
	entries, _ := os.ReadDir(thumbsDir)
	count := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".thumb.jpg") {
			count++
		}
	}
	if count != len(files) {
		t.Errorf("缩略图文件数 = %d，期望 %d", count, len(files))
	}
}

// TestRenderThumbAtomicCleansTempOnFailure 生成失败时必须清理临时文件且不产生正式缓存。
func TestRenderThumbAtomicCleansTempOnFailure(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(thumbsDir, "target.jpg")

	err := renderThumbAtomic(filepath.Join(dir, "missing.jpg"), target, false)
	if err == nil {
		t.Fatal("源文件不存在时应失败")
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Error("失败时不应留下正式缓存文件")
	}
	assertNoTempFiles(t)
}

// TestRenderThumbAtomicTimeoutCleansTemp 模拟超时：错误必须传递且临时文件被清理，
// 已有的有效缓存不能被空文件覆盖。
func TestRenderThumbAtomicTimeoutCleansTemp(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(thumbsDir, "keep.jpg")
	original := makeJPEGBytes(t)
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}

	oldRunner, oldTimeout := thumbCmdRunner, thumbTimeout
	defer func() { thumbCmdRunner, thumbTimeout = oldRunner, oldTimeout }()
	thumbTimeout = 30 * time.Millisecond
	thumbCmdRunner = func(src, out string, isVideo bool, seek string) error {
		return fmt.Errorf("缩略图生成超时（%s）", thumbTimeout)
	}

	err := renderThumbAtomic(filepath.Join(dir, "broken.jpg"), target, false)
	if err == nil {
		t.Fatal("模拟超时应返回错误")
	}
	if !strings.Contains(err.Error(), "超时") {
		t.Errorf("错误信息应包含超时原因，实际: %v", err)
	}
	// 已有有效缓存不能被破坏
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("原有缓存文件不应被删除: %v", readErr)
	}
	if !bytes.Equal(got, original) {
		t.Error("失败时不得用空文件/半成品覆盖已有缓存")
	}
	assertNoTempFiles(t)
}

// TestRenderThumbAtomicWithRealFFmpeg 真实端到端：用系统 ffmpeg 生成缩略图。
// 环境缺少可执行的 ffmpeg 时跳过（跳过原因会打印）。
func TestRenderThumbAtomicWithRealFFmpeg(t *testing.T) {
	ff := findFFmpeg()
	if ff == "" {
		t.Skip("未找到 ffmpeg，跳过真实缩略图生成用例")
	}
	if out, err := exec.Command(ff, "-version").Output(); err != nil {
		t.Skipf("ffmpeg 不可执行（%v）：%s", err, clip(string(out), 120))
	}

	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	ffmpegPath = ff // setMediaTestEnv 会禁用外部二进制，这里显式恢复

	// 用 ffmpeg 自己生成一张测试图（避免依赖 Go 的 PNG 解码器之外的资源）
	src := filepath.Join(dir, "src.png")
	if err := os.WriteFile(src, makePNGBytes(t, 640, 480), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(thumbsDir, "real.jpg")

	if err := renderThumbAtomic(src, target, false); err != nil {
		t.Fatalf("真实 ffmpeg 生成缩略图失败: %v", err)
	}
	st, err := os.Stat(target)
	if err != nil {
		t.Fatalf("缩略图未生成: %v", err)
	}
	if st.Size() < thumbMinBytes {
		t.Errorf("缩略图体积异常: %d 字节", st.Size())
	}
	// 结果必须是可解码图片，且不超过缩略图尺寸
	f, err := os.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("缩略图不可解码: %v", err)
	}
	if cfg.Width > thumbSize || cfg.Height > thumbSize {
		t.Errorf("缩略图尺寸 %dx%d 超过上限 %d", cfg.Width, cfg.Height, thumbSize)
	}
	assertNoTempFiles(t)
}

// TestEnsureThumbRejectsUnsafeSource 穿越 / 兄弟目录 / 根外链接都不能进入生成流程。
func TestEnsureThumbRejectsUnsafeSource(t *testing.T) {
	base := t.TempDir()
	restore := setMediaTestEnv(t, filepath.Join(base, "media"))
	defer restore()
	writeFileT(t, filepath.Join(base, "media", "ok.jpg"), "x")
	writeFileT(t, filepath.Join(base, "media2", "sibling.jpg"), "x")
	writeFileT(t, filepath.Join(base, "secret.jpg"), "x")

	bad := []string{"../secret.jpg", "sub/../../secret.jpg", filepath.Join(base, "media2", "sibling.jpg")}
	for _, req := range bad {
		if _, err := safeMediaPath(mediaDir, req); err == nil {
			t.Errorf("%q 必须在生成缩略图之前被拒绝", req)
		}
	}
	// 媒体目录内的正常文件可通过
	if _, err := safeMediaPath(mediaDir, "ok.jpg"); err != nil {
		t.Errorf("正常文件不应被拒绝: %v", err)
	}
}

func makePNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("编码 PNG 失败: %v", err)
	}
	return buf.Bytes()
}

func assertNoTempFiles(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(thumbsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), thumbTempPrefix) {
			t.Errorf("残留临时文件未清理: %s", e.Name())
		}
	}
}

// TestRenderThumbAtomicRealVideo 真实视频缩略图：
// 覆盖两个曾经出问题的地方——(1) 非全范围 YUV 需要 -strict unofficial 才能出图，
// (2) 多取样点重试不能残留临时文件。
func TestRenderThumbAtomicRealVideo(t *testing.T) {
	ff := findFFmpeg()
	if ff == "" {
		t.Skip("未找到 ffmpeg，跳过真实视频缩略图用例")
	}
	if out, err := exec.Command(ff, "-version").Output(); err != nil {
		t.Skipf("ffmpeg 不可执行（%v）：%s", err, clip(string(out), 120))
	}

	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	ffmpegPath = ff

	video := filepath.Join(dir, "v.mp4")
	cmd := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-y",
		// 故意用 1 秒视频：1s/3s/5s/10s 取样点都会取不到帧，只有首帧兜底能成功
		"-f", "lavfi", "-i", "testsrc=duration=1:size=320x240:rate=10",
		"-pix_fmt", "yuv420p", video)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("生成测试视频失败（%v）：%s", err, clip(string(out), 200))
	}
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(thumbsDir, "v.jpg")
	if err := renderThumbAtomic(video, target, true); err != nil {
		t.Fatalf("真实视频缩略图生成失败: %v", err)
	}
	st, err := os.Stat(target)
	if err != nil {
		t.Fatalf("视频缩略图未生成: %v", err)
	}
	if st.Size() < thumbMinBytes {
		t.Errorf("视频缩略图体积异常: %d 字节", st.Size())
	}
	assertNoTempFiles(t)
}
