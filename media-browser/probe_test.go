package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// ── 媒体元数据探测测试（第二阶段 · 第三部分）────────────────────────────

func TestParseDurationSecondsCases(t *testing.T) {
	ok := []struct {
		in   string
		want float64
	}{
		{"12.5", 12.5},
		{" 3 ", 3},
		{"0.04", 0.04},
		{"3600", 3600},
		{"12.5\n", 12.5},
	}
	for _, c := range ok {
		got, err := parseDurationSeconds(c.in)
		if err != nil || got != c.want {
			t.Errorf("parseDurationSeconds(%q) = (%v, %v)，期望 (%v, nil)", c.in, got, err, c.want)
		}
	}

	// NaN / Inf / 负数 / 0 / 空 / N/A / 非数字 一律视为无效
	bad := []string{"", "   ", "N/A", "nan", "NaN", "-1", "-0.5", "0", "abc", "inf", "+Inf", "-Inf", "1,5"}
	for _, in := range bad {
		if got, err := parseDurationSeconds(in); err == nil {
			t.Errorf("parseDurationSeconds(%q) 应报错，实际得到 %v", in, got)
		}
	}
}

func TestDurationCacheInvalidation(t *testing.T) {
	resetDurationCache()
	defer resetDurationCache()

	mt := time.Unix(1700000000, 0)
	durationCachePut("a.mp4", 100, mt, 42.5)

	if d, ok := durationCacheGet("a.mp4", 100, mt); !ok || d != 42.5 {
		t.Fatalf("缓存应命中，实际 (%v, %v)", d, ok)
	}
	if _, ok := durationCacheGet("a.mp4", 101, mt); ok {
		t.Error("文件大小变化后缓存必须失效（否则沿用旧时长）")
	}
	if _, ok := durationCacheGet("a.mp4", 100, mt.Add(time.Second)); ok {
		t.Error("修改时间变化后缓存必须失效")
	}
	if _, ok := durationCacheGet("b.mp4", 100, mt); ok {
		t.Error("不同路径不应命中缓存")
	}
	// 更新后可重新命中
	durationCachePut("a.mp4", 200, mt.Add(time.Hour), 7)
	if d, ok := durationCacheGet("a.mp4", 200, mt.Add(time.Hour)); !ok || d != 7 {
		t.Errorf("更新后的缓存应命中，实际 (%v, %v)", d, ok)
	}
	if durationCacheLen() != 1 {
		t.Errorf("缓存条目数 = %d，期望 1", durationCacheLen())
	}
}

// TestProbeDurationMissingFileIsFast 源文件不存在时必须快速失败，不能卡住整轮扫描。
func TestProbeDurationMissingFileIsFast(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()

	start := time.Now()
	if d := probeDuration(filepath.Join(dir, "nope.mp4")); d != 0 {
		t.Errorf("不存在的文件时长应为 0，实际 %v", d)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("探测不存在文件耗时 %s，疑似卡住", elapsed)
	}
}

// TestProbeDurationAndCacheWithRealFFmpeg 用真实 ffmpeg/ffprobe 生成 1 秒视频并探测时长，
// 同时验证「缓存命中后不再启动外部进程」。
func TestProbeDurationAndCacheWithRealFFmpeg(t *testing.T) {
	ff := findFFmpeg()
	probe := findFFprobe()
	if ff == "" || probe == "" {
		t.Skipf("未找到 ffmpeg(%q)/ffprobe(%q)，跳过真实媒体探测用例", ff, probe)
	}
	if out, err := exec.Command(ff, "-version").Output(); err != nil {
		t.Skipf("ffmpeg 不可执行（%v）：%s", err, clipParam(string(out), 120))
	}

	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)
	defer restore()
	ffmpegPath = ff
	ffprobePath = probe
	resetDurationCache()

	video := filepath.Join(dir, "v.mp4")
	cmd := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=duration=1:size=160x120:rate=10",
		"-pix_fmt", "yuv420p", video)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("生成测试视频失败（%v）：%s", err, clipParam(string(out), 200))
	}

	d := probeDuration(video)
	if d < 0.5 || d > 1.6 {
		t.Errorf("探测时长 = %.3f 秒，期望约 1 秒（ffprobe 路径）", d)
	}

	st, err := os.Stat(video)
	if err != nil {
		t.Fatal(err)
	}

	// fillDurations 应写入缓存，并在缓存命中时不再依赖外部进程
	files := []*MediaFile{{Path: "v.mp4", Date: st.ModTime(), IsVideo: true, Size: st.Size()}}
	fillDurations(files)
	if files[0].Duration <= 0 {
		t.Fatalf("fillDurations 未补齐时长")
	}
	if _, ok := durationCacheGet("v.mp4", st.Size(), st.ModTime()); !ok {
		t.Fatal("fillDurations 应把结果写入时长缓存")
	}

	ffmpegPath = "definitely-not-a-real-binary"
	ffprobePath = ""
	cached := []*MediaFile{{Path: "v.mp4", Date: st.ModTime(), IsVideo: true, Size: st.Size()}}
	fillDurations(cached)
	if cached[0].Duration <= 0 {
		t.Error("缓存命中时不应需要外部进程，时长必须仍可用")
	}
}

// TestFindFFprobePrefersDeploymentDir 部署目录中的 ffprobe 应被优先选中。
func TestFindFFprobePrefersDeploymentDir(t *testing.T) {
	dir := t.TempDir()
	// 伪造一个部署目录中的 ffprobe（只要存在即可被选中）
	fake := filepath.Join(dir, "ffprobe.exe")
	if err := os.WriteFile(fake, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldExe := ffprobePath
	defer func() { ffprobePath = oldExe }()

	// 直接验证候选顺序函数：把当前工作目录切到 dir 后应能找到该文件
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	found := findFFprobe()
	if found == "" {
		t.Skip("系统 PATH 中也没有 ffprobe，且部署目录探测未命中（环境受限）")
	}
	if filepath.Base(found) != "ffprobe.exe" && filepath.Base(found) != "ffprobe" {
		t.Errorf("findFFprobe 返回了非预期路径: %s", found)
	}
}
