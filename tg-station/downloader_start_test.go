package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// useStartHook 把启动命令换成「会活一会儿的假下载器」，并统计构建命令的次数。
// 钩子里故意 sleep 一小段：这正是旧实现里 isRunning() 检查与 d.cmd 赋值之间的
// 竞态窗口——没有 sessionMu 时，并发请求会全部穿过检查、各起一个进程。
func useStartHook(t *testing.T, calls *int32, buildDelay time.Duration) {
	t.Helper()
	old := startCommandHook
	startCommandHook = func(d *Downloader) (dlCommand, error) {
		atomic.AddInt32(calls, 1)
		time.Sleep(buildDelay)
		fake := sleepCmd(30)
		return dlCommand{Path: fake.Path, Args: fake.Args[1:], desc: "测试假下载器"}, nil
	}
	t.Cleanup(func() { startCommandHook = old })
}

func managedPid(d *Downloader) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.pid
}

// 连点「开始监听」（或手机+电脑同时点）只能起一个下载器。
func TestStartSerializedSoOnlyOneDownloaderRuns(t *testing.T) {
	stubExternal(t) // 没有外部独立下载器
	var calls int32
	useStartHook(t, &calls, 150*time.Millisecond)

	d := newTestDownloader("", "")
	t.Cleanup(func() { _ = d.Stop() })

	const n = 4
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = d.Start()
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("第 %d 个 Start 返回错误: %v", i+1, err)
		}
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("并发 %d 次「开始监听」应只启动 1 个下载器，实际启动了 %d 个（双实例共用 session 会封号）", n, c)
	}
	if managedPid(d) == 0 {
		t.Fatal("启动后没有记录 PID")
	}
}

// Restart 内部要调 Stop+Start，若与 sessionMu 互锁就会死锁——这里守着超时跑一遍。
func TestRestartDoesNotDeadlockAndRestartsOnce(t *testing.T) {
	stubExternal(t)
	var calls int32
	useStartHook(t, &calls, 0)

	d := newTestDownloader("", "")
	t.Cleanup(func() { _ = d.Stop() })

	if err := d.Start(); err != nil {
		t.Fatalf("首次 Start 失败: %v", err)
	}
	firstPid := managedPid(d)

	done := make(chan error, 1)
	go func() { done <- d.Restart() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Restart 失败: %v", err)
		}
	case <-time.After(60 * time.Second):
		t.Fatal("Restart 超时（sessionMu 与 Stop/Start 互锁，或旧进程没被收掉）")
	}

	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Fatalf("Start + Restart 应各启动一次（共 2 次），实际 %d 次", c)
	}
	if pid := managedPid(d); pid == 0 || pid == firstPid {
		t.Fatalf("重启后 PID 未更新（旧 %d，新 %d）", firstPid, pid)
	}
}

// 「开始监听」与「停止监听」并发轰炸：结束后不能留下受管进程引用，
// 且停止流程返回后必须确实没有受管进程在跑。
func TestStartStopStormEndsClean(t *testing.T) {
	stubExternal(t)
	var calls int32
	useStartHook(t, &calls, 20*time.Millisecond)

	d := newTestDownloader("", "")

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				_ = d.Start()
			} else {
				_ = d.Stop()
			}
		}(i)
	}
	wg.Wait()

	if err := d.Stop(); err != nil {
		t.Fatalf("收尾 Stop 失败: %v", err)
	}
	d.mu.Lock()
	leftover := d.cmd
	state := d.state
	d.mu.Unlock()
	if leftover != nil {
		t.Fatal("停止后仍挂着受管进程引用")
	}
	if state != dlStateIdle {
		t.Fatalf("停止后状态应为 idle，实际 %s", state)
	}
	if pid := managedPid(d); pid != 0 {
		t.Fatalf("停止后 PID 应为 0，实际 %d", pid)
	}
}
