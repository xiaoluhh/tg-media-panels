package main

import (
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── 扫描调度（第二阶段 · 第二部分）──────────────────────────────────────
//
// 旧实现：scanFiles 用 CAS 做重入保护，扫描期间的新请求被**静默丢弃**，
// 而 refreshHandler 仍然照旧去生成缩略图 —— 结果是「扫描没跑、缩略图却按旧索引
// 跑了一遍」，最新下载的文件迟迟进不了索引。
//
// 新实现：单一调度器 + pending 标志 + 补扫循环：
//  1. 同一时刻最多一个扫描在执行；
//  2. 扫描期间的任意多个请求只置一个 pending 标志（合并为下一轮补扫）；
//  3. 每轮扫描结束后若 pending 已置位，立即再扫一轮（含缩略图生成期间到达的请求），
//     因此最后一轮请求不会被丢弃；
//  4. 只有在「没有待执行补扫」之后才生成缩略图，避免把未入索引的新文件当作不存在。

// scanScheduler 串行化扫描请求。
type scanScheduler struct {
	mu        sync.Mutex
	running   bool
	pending   bool
	rounds    int // 已完成的扫描轮数（测试/诊断用）
	scanFn    func()
	thumbFn   func()
	waiters   map[chan int]struct{}
	startedAt time.Time
}

func newScanScheduler(scanFn, thumbFn func()) *scanScheduler {
	return &scanScheduler{scanFn: scanFn, thumbFn: thumbFn, waiters: map[chan int]struct{}{}}
}

var scanSched = newScanScheduler(scanFiles, generateAllThumbnails)

// Request 请求一次扫描/刷新。
//   - 没有扫描在跑：立即启动（同步返回，扫描在后台 goroutine 中执行）；
//   - 已有扫描在跑：置 pending 标志并返回 false（请求被合并，绝不丢弃）。
func (s *scanScheduler) Request() bool {
	s.mu.Lock()
	if s.running {
		s.pending = true
		s.mu.Unlock()
		log.Printf("[media] 扫描已在进行中，本次请求合并为下一轮补扫")
		return false
	}
	s.running = true
	s.startedAt = time.Now()
	s.mu.Unlock()

	go s.loop()
	return true
}

// loop 是唯一的扫描执行体。
func (s *scanScheduler) loop() {
	for {
		s.runGuarded(s.scanFn, "扫描")

		// 扫描刚结束时若有请求到达，立即补扫一轮（不生成缩略图，先追平索引）
		s.mu.Lock()
		if s.pending {
			s.pending = false
			s.mu.Unlock()
			continue
		}
		s.mu.Unlock()

		// 没有待补扫任务后才生成缩略图（基于最新索引）
		if s.thumbFn != nil {
			s.runGuarded(s.thumbFn, "缩略图生成")
		}

		s.mu.Lock()
		s.rounds++
		if s.pending {
			// 缩略图生成期间又收到请求：再补扫一轮，然后回到缩略图步骤
			s.pending = false
			s.mu.Unlock()
			continue
		}
		s.running = false
		for ch := range s.waiters {
			close(ch)
		}
		s.waiters = map[chan int]struct{}{}
		s.mu.Unlock()
		return
	}
}

// runGuarded 执行一次任务并兜住 panic。
// 目的是保证调度器状态（running/pending/waiters）一定能复位——否则一次
// 扫描 panic 会让 running 永远为 true，后续所有刷新请求都会被合并却永不执行。
// 这里只作为最后一道防线，不作为掩盖业务错误的常规手段（错误原因会记录日志）。
func (s *scanScheduler) runGuarded(fn func(), what string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[media] %s 过程 panic（已恢复，调度器保持可用）: %v", what, r)
		}
	}()
	fn()
}

// Running 报告当前是否有扫描在跑。
func (s *scanScheduler) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Pending 报告是否有待执行的补扫。
func (s *scanScheduler) Pending() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pending
}

// Rounds 返回已完成的扫描轮数。
func (s *scanScheduler) Rounds() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rounds
}

// Wait 等待当前（含补扫在内）一轮调度彻底结束。
// 返回 false 表示等待超时（调用方据此如实告知前端，不能谎报完成）。
func (s *scanScheduler) Wait(timeout time.Duration) bool {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return true
	}
	ch := make(chan int)
	s.waiters[ch] = struct{}{}
	s.mu.Unlock()

	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		s.mu.Lock()
		delete(s.waiters, ch)
		s.mu.Unlock()
		return false
	}
}

// ── 实际扫描 ────────────────────────────────────────────────────────────

// scanFiles 由调度器调用：扫描媒体目录并重建索引。
// 直接调用它不会触发调度（仅供调度器与测试使用），业务入口一律用 scanSched.Request()。
func scanFiles() {
	// 复用上次索引中「同路径 + 同大小 + 同修改时间」的视频时长，避免重复探测
	indexMutex.RLock()
	prevByPath := make(map[string]*MediaFile, len(mediaIndex))
	for _, f := range mediaIndex {
		prevByPath[f.Path] = f
	}
	indexMutex.RUnlock()

	var files []*MediaFile
	err := filepath.WalkDir(mediaDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 单个条目不可读（权限/损坏）不应中断整轮扫描
			log.Printf("[media] 扫描跳过 %s: %v", path, err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || path == thumbsDir {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !imageExts[ext] && !videoExts[ext] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			log.Printf("[media] 扫描跳过 %s: %v", path, err)
			return nil
		}
		relPath, err := filepath.Rel(mediaDir, path)
		if err != nil {
			return nil
		}
		rel := filepath.ToSlash(relPath)
		mf := &MediaFile{
			Path:    rel,
			Date:    info.ModTime(),
			IsVideo: videoExts[ext],
			Size:    info.Size(),
		}
		if mf.IsVideo {
			// 复用的前提：内容未变（大小 + 修改时间都一致）
			if p, ok := prevByPath[rel]; ok && p.Size == info.Size() &&
				p.Date.Equal(info.ModTime()) && p.Duration > 0 {
				mf.Duration = p.Duration
				durationCachePut(rel, info.Size(), info.ModTime(), p.Duration)
			}
		}
		files = append(files, mf)
		return nil
	})
	if err != nil {
		log.Printf("Scan error: %v", err)
		return
	}

	// 按修改时间倒序（与旧实现一致，分页结果不变）
	sortByDateDesc(files)

	// 补齐缺失的视频时长（缓存未命中的才真正调用 ffprobe/ffmpeg）
	fillDurations(files)

	// 文件列表 + 日期索引在同一把锁下原子切换
	rebuildIndex(files)

	saveIndex()
	log.Printf("Scanned %d files", len(files))
}

// sortByDateDesc 按修改时间倒序排列（稳定排序，时间相同的保持遍历顺序）。
func sortByDateDesc(files []*MediaFile) {
	sort.SliceStable(files, func(i, j int) bool { return files[i].Date.After(files[j].Date) })
}
