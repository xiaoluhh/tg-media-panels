package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

// ── 媒体索引与日期索引（第二阶段 · 第三部分）────────────────────────────
//
// 旧实现里 /api/files/{date}、/api/dates、首页注入每次都遍历整个 mediaIndex
// （5000+ 文件 × 每个请求），日期查询是 O(全部文件数)。
//
// 现在维护一个日期索引 map[string][]*MediaFile，与文件列表在**同一把锁**下
// 一起原子切换（rebuildIndex），保证「列表已更新、日期索引还是旧数据」不可能发生。
// 日期键沿用现有格式 time.Format("2006-01-02")，顺序沿用「全局按修改时间倒序」，
// 因此分页结果与旧实现完全一致。

var (
	// dateIndexPool 保存当前生效的日期索引（只读视图，重建时整体替换）。
	dateIndex map[string][]*MediaFile
	// dateKeys 保存降序排列的日期列表，避免 /api/dates 每次重算排序。
	dateKeys []string
)

// rebuildIndexLocked 用新的文件列表重建索引结构。调用方必须已持有 indexMutex 写锁。
func rebuildIndexLocked(files []*MediaFile) {
	idx := make(map[string][]*MediaFile, 64)
	for _, f := range files {
		key := f.Date.Format("2006-01-02")
		idx[key] = append(idx[key], f)
	}
	keys := make([]string, 0, len(idx))
	for k := range idx {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	mediaIndex = files
	dateIndex = idx
	dateKeys = keys
}

// rebuildIndex 加锁重建索引（文件列表 + 日期索引 + 日期键一起切换）。
func rebuildIndex(files []*MediaFile) {
	indexMutex.Lock()
	rebuildIndexLocked(files)
	lastScan = time.Now()
	indexMutex.Unlock()
}

// filesForDate 返回某个日期的文件（返回共享切片，调用方只读）。
func filesForDate(date string) []*MediaFile {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	return dateIndex[date]
}

// indexDates 返回降序日期列表的副本。
func indexDates() []string {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	out := make([]string, len(dateKeys))
	copy(out, dateKeys)
	return out
}

// indexTotal 返回索引中的文件总数。
func indexTotal() int {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	return len(mediaIndex)
}

// indexSnapshotFiles 返回文件列表的浅拷贝（用于批量缩略图等场景）。
func indexSnapshotFiles() []*MediaFile {
	indexMutex.RLock()
	defer indexMutex.RUnlock()
	out := make([]*MediaFile, len(mediaIndex))
	copy(out, mediaIndex)
	return out
}

// loadIndex 从磁盘加载索引文件（内容为 []*MediaFile）。
func loadIndex() {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		log.Printf("Index file not found: %v", err)
		return
	}

	var files []*MediaFile
	if err := json.Unmarshal(data, &files); err != nil {
		log.Printf("Failed to parse index: %v", err)
		return
	}
	// 索引文件里的字段可能缺失（如旧版本没有 Size），做一次规范化
	for _, f := range files {
		if f == nil {
			continue
		}
		if f.Duration < 0 {
			f.Duration = 0
		}
	}

	rebuildIndex(files)
	log.Printf("Loaded index with %d files", len(files))
}

// saveIndex 持久化当前索引（先取快照再序列化，避免持锁做 IO）。
func saveIndex() {
	files := indexSnapshotFiles()
	data, err := json.Marshal(files)
	if err != nil {
		log.Printf("Failed to marshal index: %v", err)
		return
	}
	if err := os.WriteFile(indexPath, data, 0o644); err != nil {
		log.Printf("Failed to save index: %v", err)
	}
}
