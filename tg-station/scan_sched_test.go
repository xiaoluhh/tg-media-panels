package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// ── 扫描调度测试（第二阶段 · 第二部分）──────────────────────────────────
//
// 用可控的假扫描/假缩略图函数验证调度语义：
//   单次扫描 / 扫描期间一次请求 / 多次请求合并 / 第二轮期间再次请求 /
//   扫描 panic 后仍可继续刷新 / 缩略图只在补扫清空后执行 / 并发压力下无并发扫描。

type fakeScanner struct {
	mu        sync.Mutex
	gates     chan chan struct{} // 每轮扫描开始时尝试取出一个 gate 并等待其关闭
	started   chan struct{}      // 每轮扫描开始时的通知（带缓冲）
	rounds    int
	conc      int
	maxConc   int
	panics    int
	panicNext bool
	workDelay time.Duration
	thumbErrs []string
	thumbRuns int
}

func newFakeScanner() *fakeScanner {
	return &fakeScanner{
		gates:   make(chan chan struct{}, 8),
		started: make(chan struct{}, 32),
	}
}

func (f *fakeScanner) scan() {
	f.mu.Lock()
	f.conc++
	if f.conc > f.maxConc {
		f.maxConc = f.conc
	}
	shouldPanic := f.panicNext
	f.panicNext = false
	f.mu.Unlock()

	select {
	case f.started <- struct{}{}:
	default:
	}
	select {
	case g := <-f.gates:
		<-g
	default:
	}
	if f.workDelay > 0 {
		time.Sleep(f.workDelay)
	}

	f.mu.Lock()
	f.conc--
	if shouldPanic {
		f.panics++
		f.mu.Unlock()
		panic("模拟扫描失败")
	}
	f.rounds++
	f.mu.Unlock()
}

func (f *fakeScanner) snapshot() (rounds, maxConc, panics int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rounds, f.maxConc, f.panics
}

// scanSchedulerFor 构造一个带假缩略图回调的调度器，并记录缩略图运行时的调度器状态。
func scanSchedulerFor(fs *fakeScanner) (*scanScheduler, *[]string) {
	var sched *scanScheduler
	states := &[]string{}
	sched = newScanScheduler(fs.scan, func() {
		fs.mu.Lock()
		fs.thumbRuns++
		fs.mu.Unlock()
		*states = append(*states, fmt.Sprintf("pending=%v running=%v", sched.Pending(), sched.Running()))
	})
	return sched, states
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("等待条件超时: %s", what)
}

func TestScanSchedulerSingleRun(t *testing.T) {
	fs := newFakeScanner()
	sched, states := scanSchedulerFor(fs)

	if !sched.Request() {
		t.Fatal("首次请求应当启动扫描")
	}
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待扫描结束超时")
	}
	rounds, maxConc, _ := fs.snapshot()
	if rounds != 1 {
		t.Errorf("扫描轮数 = %d，期望 1", rounds)
	}
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1", maxConc)
	}
	if fs.thumbRuns != 1 {
		t.Errorf("缩略图运行次数 = %d，期望 1", fs.thumbRuns)
	}
	if len(*states) != 1 || (*states)[0] != "pending=false running=true" {
		t.Errorf("缩略图运行状态异常: %v", *states)
	}
}

func TestScanSchedulerMergesRequestsDuringScan(t *testing.T) {
	fs := newFakeScanner()
	g1 := make(chan struct{})
	fs.gates <- g1
	sched, _ := scanSchedulerFor(fs)

	if !sched.Request() {
		t.Fatal("首次请求应当启动扫描")
	}
	waitFor(t, "第一轮开始", func() bool { return drain(fs.started) })

	// 扫描进行中的 5 个请求：必须全部被合并，而不是被静默丢弃，也不启动新扫描
	for i := 0; i < 5; i++ {
		if sched.Request() {
			t.Errorf("扫描进行中的请求应返回 false（合并为补扫）")
		}
	}
	if !sched.Pending() {
		t.Fatal("扫描期间的请求应置位 pending")
	}

	close(g1) // 放开第一轮
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待调度结束超时")
	}
	rounds, maxConc, _ := fs.snapshot()
	if rounds != 2 {
		t.Errorf("扫描轮数 = %d，期望 2（首轮 + 一次合并补扫）", rounds)
	}
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1", maxConc)
	}
	if sched.Pending() {
		t.Error("调度结束后不应残留 pending")
	}
}

// drain 非阻塞地消费 started 通知（返回是否有通知）。
func drain(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestScanSchedulerRequestDuringSecondRound(t *testing.T) {
	fs := newFakeScanner()
	g1 := make(chan struct{})
	g2 := make(chan struct{})
	fs.gates <- g1 // 第一轮阻塞
	fs.gates <- g2 // 第二轮也阻塞（保证「第二轮期间」的请求确实发生在扫描进行中）
	sched, states := scanSchedulerFor(fs)

	sched.Request()
	waitFor(t, "第一轮开始", func() bool { return drain(fs.started) })

	sched.Request() // 第一轮期间的请求 → 触发补扫
	if !sched.Pending() {
		t.Fatal("第一轮期间的请求应置位 pending")
	}
	close(g1)

	waitFor(t, "第二轮开始", func() bool { return drain(fs.started) })
	// 第二轮执行期间的请求同样不能被丢弃
	if sched.Request() {
		t.Fatal("第二轮进行中的请求应返回 false（合并为补扫）")
	}
	if !sched.Pending() {
		t.Fatal("第二轮期间的请求应置位 pending（不能丢失最后一轮请求）")
	}
	close(g2)
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待调度结束超时")
	}

	rounds, maxConc, _ := fs.snapshot()
	if rounds != 3 {
		t.Errorf("扫描轮数 = %d，期望 3（首轮 + 两轮补扫）", rounds)
	}
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1", maxConc)
	}
	// 缩略图必须在所有补扫完成之后、且 pending 为空时运行
	if len(*states) == 0 {
		t.Fatal("缩略图回调未被执行")
	}
	for _, st := range *states {
		if st != "pending=false running=true" {
			t.Errorf("缩略图运行时状态异常: %s（期望 pending=false）", st)
		}
	}
}

func TestScanSchedulerRecoversFromPanic(t *testing.T) {
	fs := newFakeScanner()
	fs.panicNext = true
	sched, _ := scanSchedulerFor(fs)

	sched.Request()
	if !sched.Wait(3 * time.Second) {
		t.Fatal("扫描 panic 后调度器必须复位，不能永久卡住")
	}
	_, _, panics := fs.snapshot()
	if panics != 1 {
		t.Fatalf("模拟的 panic 次数 = %d，期望 1", panics)
	}
	// 扫描失败后系统仍能接受后续刷新
	if !sched.Request() {
		t.Fatal("panic 之后的新请求应当能启动新一轮扫描")
	}
	if !sched.Wait(3 * time.Second) {
		t.Fatal("等待第二轮扫描超时")
	}
	rounds, _, _ := fs.snapshot()
	if rounds != 1 {
		t.Errorf("成功轮数 = %d，期望 1（首轮 panic 不计入）", rounds)
	}
}

func TestScanSchedulerConcurrentStress(t *testing.T) {
	fs := newFakeScanner()
	fs.workDelay = 5 * time.Millisecond
	sched, _ := scanSchedulerFor(fs)

	var wg sync.WaitGroup
	started := 0
	var mu sync.Mutex
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if sched.Request() {
				mu.Lock()
				started++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if !sched.Wait(5 * time.Second) {
		t.Fatal("等待调度结束超时")
	}
	rounds, maxConc, _ := fs.snapshot()
	if maxConc != 1 {
		t.Errorf("并发扫描峰值 = %d，期望 1（不允许两个扫描同时执行）", maxConc)
	}
	if rounds < 1 {
		t.Errorf("扫描轮数 = %d，期望 ≥1", rounds)
	}
	if sched.Pending() {
		t.Error("结束后不应残留 pending")
	}
	if sched.Running() {
		t.Error("结束后不应处于 running 状态")
	}
	if started < 1 {
		t.Errorf("至少应有一个请求真正启动扫描，实际 %d", started)
	}
}

func TestScanSchedulerWaitTimeout(t *testing.T) {
	fs := newFakeScanner()
	g := make(chan struct{})
	fs.gates <- g
	sched, _ := scanSchedulerFor(fs)

	sched.Request()
	waitFor(t, "扫描开始", func() bool { return drain(fs.started) })

	if sched.Wait(80 * time.Millisecond) {
		t.Error("扫描尚未结束，Wait 不应返回 true（否则等于谎报完成）")
	}
	close(g)
	if !sched.Wait(3 * time.Second) {
		t.Error("扫描结束后 Wait 应返回 true")
	}
}

// TestRealScanIndexIsNeverStale 用真实的 scanFiles + 全局调度器验证：
// 一轮扫描期间的多次刷新请求不会导致「最新文件永久进不了索引」。
func TestRealScanIndexIsNeverStale(t *testing.T) {
	dir := t.TempDir()
	restore := setMediaTestEnv(t, dir)

	// 初始文件
	writeFileT(t, filepath.Join(dir, "old.jpg"), "old")
	scanSched.Request()
	if !scanSched.Wait(10 * time.Second) {
		t.Fatal("首次扫描超时")
	}
	if got := indexTotal(); got != 1 {
		t.Fatalf("初始扫描后索引文件数 = %d，期望 1", got)
	}

	// 扫描期间连续请求：新文件必须在最终索引里出现（请求不能被丢弃）
	writeFileT(t, filepath.Join(dir, "new1.jpg"), "n1")
	writeFileT(t, filepath.Join(dir, "new2.jpg"), "n2")
	for i := 0; i < 3; i++ {
		scanSched.Request()
	}
	if !scanSched.Wait(15 * time.Second) {
		t.Fatal("补扫超时")
	}

	got := indexTotal()
	if got != 3 {
		t.Errorf("最终索引文件数 = %d，期望 3（最新文件不能因请求被丢弃而遗漏）", got)
	}
	if files := filesForDate(time.Now().Format("2006-01-02")); len(files) != 3 {
		t.Errorf("日期索引文件数 = %d，期望 3（日期索引与文件列表必须一致）", len(files))
	}
	restore()
}

// setMediaTestEnv 把全局路径指向临时目录，返回恢复函数。
func setMediaTestEnv(t *testing.T, mediaRoot string) func() {
	t.Helper()
	oldMedia, oldThumbs, oldIndex, oldProbe := mediaDir, thumbsDir, indexPath, ffprobePath
	oldFFmpeg := ffmpegPath
	mediaDir = mediaRoot
	thumbsDir = filepath.Join(mediaRoot, ".thumbs-test")
	indexPath = filepath.Join(mediaRoot, "media_index_test.json")
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		t.Fatalf("创建测试缩略图目录失败: %v", err)
	}
	// 测试里不依赖外部二进制时把探测关掉，避免误触发真实 ffprobe/ffmpeg
	ffprobePath = ""
	ffmpegPath = "definitely-not-a-real-binary"
	resetMediaRootCache()
	resetDurationCache()

	// 索引与调度器状态复位
	rebuildIndex(nil)

	return func() {
		// 等待可能仍在跑的扫描/缩略图任务结束，避免其继续使用即将失效的全局路径
		scanSched.Wait(10 * time.Second)
		mediaDir, thumbsDir, indexPath, ffprobePath = oldMedia, oldThumbs, oldIndex, oldProbe
		ffmpegPath = oldFFmpeg
		resetMediaRootCache()
		resetDurationCache()
		rebuildIndex(nil)
	}
}

func writeFileT(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
}
