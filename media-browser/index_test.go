package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ── 日期索引测试与性能基准（第二阶段 · 第三部分）────────────────────────
//
// 旧实现：/api/files/{date} 每次都遍历整个 mediaIndex（O(全部文件)）。
// 新实现：按日期分组的索引，查询成本 O(该日期的文件数)。
// 这里既验证「索引结果与线性扫描完全一致」，也给出可重复的基准数据。

func genMediaFiles(n, days int) []*MediaFile {
	base := time.Date(2026, 5, 1, 12, 0, 0, 0, time.Local)
	files := make([]*MediaFile, 0, n)
	for i := 0; i < n; i++ {
		day := time.Duration(i%days) * 24 * time.Hour
		ts := base.Add(-day).Add(time.Duration(i/days) * time.Minute)
		files = append(files, &MediaFile{
			Path:     fmt.Sprintf("day%02d/file%05d.jpg", i%days, i),
			Date:     ts,
			IsVideo:  i%10 == 0,
			Size:     int64(i * 1024),
			Duration: float64(i % 60),
		})
	}
	sortByDateDesc(files) // 与 scanFiles 相同的排序规则
	return files
}

func withIndex(t *testing.T, files []*MediaFile) {
	t.Helper()
	rebuildIndex(files)
	t.Cleanup(func() { rebuildIndex(nil) })
}

func TestRebuildIndexDateGrouping(t *testing.T) {
	files := genMediaFiles(300, 6)
	withIndex(t, files)

	if got := indexTotal(); got != 300 {
		t.Fatalf("indexTotal = %d，期望 300", got)
	}
	keys := indexDates()
	if len(keys) != 6 {
		t.Fatalf("日期个数 = %d，期望 6", len(keys))
	}
	// 日期键必须降序
	for i := 1; i < len(keys); i++ {
		if keys[i-1] <= keys[i] {
			t.Fatalf("日期键未按降序排列: %v", keys[:min(3, len(keys))])
		}
	}
	// 每个日期分组的文件数之和 = 总数，且键与文件自身日期一致
	total := 0
	for _, k := range keys {
		for _, f := range filesForDate(k) {
			total++
			if f.Date.Format("2006-01-02") != k {
				t.Fatalf("日期索引分组错误: 文件 %s 出现在 %s 组", f.Path, k)
			}
		}
	}
	if total != 300 {
		t.Fatalf("各分组文件数合计 = %d，期望 300", total)
	}
}

func TestFilesForDateMatchesLinearScan(t *testing.T) {
	files := genMediaFiles(2500, 25)
	withIndex(t, files)

	// 逐日期比对：新索引查询必须与旧的「遍历全部文件」结果完全一致（含顺序）
	for _, date := range indexDates() {
		linear := linearScanForDate(files, date)
		indexed := filesForDate(date)
		if len(linear) != len(indexed) {
			t.Fatalf("%s: 索引结果 %d 条，线性扫描 %d 条", date, len(indexed), len(linear))
		}
		for i := range linear {
			if linear[i].Path != indexed[i].Path {
				t.Fatalf("%s: 第 %d 条不一致（线性 %s / 索引 %s）", date, i, linear[i].Path, indexed[i].Path)
			}
		}
	}

	// 不存在的日期 / 非法日期
	if got := filesForDate("1999-01-01"); len(got) != 0 {
		t.Errorf("不存在的日期应返回空，实际 %d 条", len(got))
	}
	if got := filesForDate("not-a-date"); len(got) != 0 {
		t.Errorf("非法日期应返回空，实际 %d 条", len(got))
	}
}

// linearScanForDate 复刻旧实现的查询方式（遍历整个列表筛选同一天）。
func linearScanForDate(files []*MediaFile, date string) []*MediaFile {
	var out []*MediaFile
	for _, f := range files {
		if f.Date.Format("2006-01-02") == date {
			out = append(out, f)
		}
	}
	return out
}

func TestDateIndexConcurrentReadWrite(t *testing.T) {
	files := genMediaFiles(600, 10)
	withIndex(t, files)

	stop := make(chan struct{})
	var wg sync.WaitGroup
	// 并发读
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				for _, d := range indexDates() {
					_ = filesForDate(d)
				}
				_ = indexTotal()
				_ = indexSnapshotFiles()
			}
		}()
	}
	// 并发重建（模拟扫描完成）
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				rebuildIndex(files)
			}
		}()
	}
	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()

	// 重建后列表与日期索引必须一致
	if indexTotal() != 600 {
		t.Fatalf("并发重建后 indexTotal = %d，期望 600", indexTotal())
	}
	sum := 0
	for _, d := range indexDates() {
		sum += len(filesForDate(d))
	}
	if sum != 600 {
		t.Fatalf("并发重建后日期索引合计 = %d，期望 600（两者必须一起原子切换）", sum)
	}
}

func BenchmarkFilesByDateLinear(b *testing.B) {
	files := genMediaFiles(5000, 30)
	rebuildIndex(files)
	defer rebuildIndex(nil)
	date := files[0].Date.Format("2006-01-02")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := linearScanForDate(files, date)
		if len(out) == 0 {
			b.Fatal("基准数据异常：当天无文件")
		}
	}
}

func BenchmarkFilesByDateIndexed(b *testing.B) {
	files := genMediaFiles(5000, 30)
	rebuildIndex(files)
	defer rebuildIndex(nil)
	date := files[0].Date.Format("2006-01-02")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := filesForDate(date)
		if len(out) == 0 {
			b.Fatal("基准数据异常：当天无文件")
		}
	}
}
