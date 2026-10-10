package main

import (
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// sleepCmd 起一个会自己活一小段时间的子进程（Windows 用 ping，其它平台用 sleep）。
// 用来模拟「等待扫码的下载器子进程」：既能自然退出，也能被 kill。
func sleepCmd(seconds int) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("ping", "-n", strconv.Itoa(seconds+1), "127.0.0.1")
	}
	return exec.Command("sleep", strconv.Itoa(seconds))
}

// newTestQrProc 起子进程并按生产同一条路径挂上唯一收尸者。
func newTestQrProc(t *testing.T, cmd *exec.Cmd) *qrProcState {
	t.Helper()
	if err := cmd.Start(); err != nil {
		t.Fatalf("启动测试子进程失败: %v", err)
	}
	st := &qrProcState{cmd: cmd, done: make(chan struct{})}
	go reapQr(st)
	return st
}

// swapQrHooks 换掉测试钩子，返回还原函数。
func swapQrHooks(killer func(int) error, start func() (*exec.Cmd, map[string]interface{}, string, error)) func() {
	oldKill, oldStart := qrKillTree, qrStart
	qrKillTree, qrStart = killer, start
	return func() { qrKillTree, qrStart = oldKill, oldStart }
}

// passThroughKiller 真杀进程（调生产实现），只额外计数。
func passThroughKiller(counter *int32) func(int) error {
	return func(pid int) error {
		atomic.AddInt32(counter, 1)
		return killProcessTree(pid)
	}
}

func setQrCur(st *qrProcState) {
	qrMu.Lock()
	qrCur = st
	qrMu.Unlock()
}

func qrCurSnapshot() *qrProcState {
	qrMu.Lock()
	defer qrMu.Unlock()
	return qrCur
}

func waitDone(t *testing.T, st *qrProcState, d time.Duration, what string) {
	t.Helper()
	select {
	case <-st.done:
	case <-time.After(d):
		t.Fatalf("%s：%s 内没有等到进程被收尸", what, d)
	}
}

// 进程自然退出后，引用必须被收尸者清掉——否则下次扫码会拿旧 PID 去 taskkill，
// 而 PID 可能已经被系统复用给别的进程。
func TestReapQrClearsReferenceAfterExit(t *testing.T) {
	st := newTestQrProc(t, sleepCmd(1))
	setQrCur(st)

	waitDone(t, st, 20*time.Second, "自然退出")

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if qrCurSnapshot() == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("子进程自然退出后 qrCur 仍指向旧进程（旧 PID 可能已被复用）")
}

// 已经退出的进程一律不许再 kill。
func TestKillQrLockedSkipsExitedProcess(t *testing.T) {
	var kills int32
	restore := swapQrHooks(func(pid int) error {
		atomic.AddInt32(&kills, 1)
		return nil
	}, qrStart)
	defer restore()

	st := newTestQrProc(t, sleepCmd(1))
	waitDone(t, st, 20*time.Second, "自然退出")

	// 复现旧实现的隐患场景：进程已退出，引用还挂着
	setQrCur(st)

	qrMu.Lock()
	killQrLocked()
	qrMu.Unlock()

	if n := atomic.LoadInt32(&kills); n != 0 {
		t.Fatalf("对已退出的进程调用了 %d 次 kill（PID 可能已被复用，会误杀无关进程）", n)
	}
	if qrCurSnapshot() != nil {
		t.Fatal("killQrLocked 之后 qrCur 没有清空")
	}
}

// 活着的进程：kill 一次，并且要等它真的退出（done 关闭）。
func TestKillQrLockedKillsLiveProcess(t *testing.T) {
	var kills int32
	restore := swapQrHooks(passThroughKiller(&kills), qrStart)
	defer restore()

	st := newTestQrProc(t, sleepCmd(30))
	setQrCur(st)

	qrMu.Lock()
	killQrLocked()
	qrMu.Unlock()

	if n := atomic.LoadInt32(&kills); n != 1 {
		t.Fatalf("活着的进程应恰好被 kill 1 次，实际 %d 次", n)
	}
	waitDone(t, st, 10*time.Second, "被 kill 后")
	if qrCurSnapshot() != nil {
		t.Fatal("killQrLocked 之后 qrCur 没有清空")
	}
}

// 并发的扫码请求必须被串行化：每个后来的请求都要先杀掉前一个进程，
// 最终只留一个。若「kill 与 set」不在同一把锁里，就统计不到 n-1 次 kill，
// 而且会有多个进程同时存在（共用 session 有封号风险）。
func TestBeginQrLoginSerializesKillAndStart(t *testing.T) {
	var kills, starts int32
	restore := swapQrHooks(passThroughKiller(&kills),
		func() (*exec.Cmd, map[string]interface{}, string, error) {
			atomic.AddInt32(&starts, 1)
			cmd := sleepCmd(30)
			if err := cmd.Start(); err != nil {
				return nil, nil, "", err
			}
			return cmd, map[string]interface{}{"status": "qr"}, "测试后端", nil
		})
	defer restore()

	const n = 4
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, _, err := beginQrLogin(); err != nil {
				t.Errorf("beginQrLogin 失败: %v", err)
			}
		}()
	}
	wg.Wait()

	if s := atomic.LoadInt32(&starts); s != n {
		t.Fatalf("启动次数 %d，期望 %d", s, n)
	}
	if k := atomic.LoadInt32(&kills); k != n-1 {
		t.Fatalf("并发扫码未串行化：期望 kill %d 次（每个新请求杀掉前一个），实际 %d 次", n-1, k)
	}

	// 收尾，别留测试进程
	qrMu.Lock()
	killQrLocked()
	qrMu.Unlock()
	if qrCurSnapshot() != nil {
		t.Fatal("测试收尾后 qrCur 应为 nil")
	}
}
