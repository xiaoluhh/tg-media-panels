package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 分页接口回归测试：page 缺省 / 合法 / 非整数 / 0 / 负数 / 超出页数 / 极大页码。
// 旧实现用 strconv.Atoi 后忽略错误且不校验最小值，page=0 或负数会算出负的切片
// 起点 → 越界 panic；这些用例正是为了让那种回归无法再发生。

func seedIndex(t *testing.T, counts map[string]int) {
	t.Helper()
	indexMutex.Lock()
	old := mediaIndex
	indexMutex.Unlock()
	t.Cleanup(func() {
		indexMutex.Lock()
		rebuildIndexLocked(old)
		indexMutex.Unlock()
	})

	var idx []*MediaFile
	for date, n := range counts {
		d, err := time.Parse("2006-01-02", date)
		if err != nil {
			t.Fatalf("测试日期解析失败: %v", err)
		}
		for i := 0; i < n; i++ {
			idx = append(idx, &MediaFile{
				Path:     fmt.Sprintf("%s/file%03d.jpg", date, i),
				Date:     d.Add(time.Duration(i) * time.Minute),
				IsVideo:  i%10 == 0,
				Size:     int64(1024 * (i + 1)),
				Duration: float64(i),
			})
		}
	}
	// 按时间倒序（与 scanFiles 一致）
	for i, j := 0, len(idx)-1; i < j; i, j = i+1, j-1 {
		idx[i], idx[j] = idx[j], idx[i]
	}
	// 通过索引重建入口写入：文件列表与日期索引必须同步切换
	// （/api/files/{date} 现在直接查日期索引，单独改 mediaIndex 已不足以生效）
	rebuildIndex(idx)
}

type filesResp struct {
	Files []LightMediaFile `json:"files"`
	Total int              `json:"total"`
	Pages int              `json:"pages"`
	Page  int              `json:"page"`
	Date  string           `json:"date"`
	Error string           `json:"error"`
}

func callFilesHandler(t *testing.T, target string) (int, filesResp) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	filesHandler(rec, req) // 旧实现在 page=0/-1 时这里会 panic，测试直接失败
	var resp filesResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	return rec.Code, resp
}

func TestFilesHandlerPagination(t *testing.T) {
	seedIndex(t, map[string]int{"2026-05-01": 250, "2026-05-02": 5})

	cases := []struct {
		name      string
		target    string
		wantCode  int
		wantCount int
		wantTotal int
		wantPages int
		wantPage  int
	}{
		{"未传 page → 第 1 页", "/api/files/2026-05-01", http.StatusOK, 100, 250, 3, 1},
		{"page 为空串 → 第 1 页", "/api/files/2026-05-01?page=", http.StatusOK, 100, 250, 3, 1},
		{"第 2 页", "/api/files/2026-05-01?page=2", http.StatusOK, 100, 250, 3, 2},
		{"最后一页不满", "/api/files/2026-05-01?page=3", http.StatusOK, 50, 250, 3, 3},
		{"超出页数 → 空数组", "/api/files/2026-05-01?page=4", http.StatusOK, 0, 250, 3, 4},
		{"极大页码（上限内）→ 空数组", "/api/files/2026-05-01?page=9999999", http.StatusOK, 0, 250, 3, 9999999},
		{"小集合第 1 页", "/api/files/2026-05-02", http.StatusOK, 5, 5, 1, 1},
		{"日期无文件", "/api/files/1999-01-01", http.StatusOK, 0, 0, 0, 1},

		{"非整数 → 400", "/api/files/2026-05-01?page=abc", http.StatusBadRequest, 0, 0, 0, 0},
		{"小数 → 400", "/api/files/2026-05-01?page=1.5", http.StatusBadRequest, 0, 0, 0, 0},
		{"0 → 400", "/api/files/2026-05-01?page=0", http.StatusBadRequest, 0, 0, 0, 0},
		{"负数 → 400", "/api/files/2026-05-01?page=-1", http.StatusBadRequest, 0, 0, 0, 0},
		{"超过上限 → 400", "/api/files/2026-05-01?page=10000001", http.StatusBadRequest, 0, 0, 0, 0},
		{"整数溢出字符串 → 400", "/api/files/2026-05-01?page=99999999999999999999", http.StatusBadRequest, 0, 0, 0, 0},
		{"带空格 → 400", "/api/files/2026-05-01?page=%20", http.StatusBadRequest, 0, 0, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, resp := callFilesHandler(t, tc.target)
			if code != tc.wantCode {
				t.Fatalf("HTTP 状态码 = %d, 期望 %d（body error=%q）", code, tc.wantCode, resp.Error)
			}
			if tc.wantCode == http.StatusBadRequest {
				if resp.Error == "" {
					t.Fatalf("400 响应必须带 error 说明")
				}
				return
			}
			if len(resp.Files) != tc.wantCount {
				t.Errorf("files 数量 = %d, 期望 %d", len(resp.Files), tc.wantCount)
			}
			if resp.Total != tc.wantTotal {
				t.Errorf("total = %d, 期望 %d", resp.Total, tc.wantTotal)
			}
			if resp.Pages != tc.wantPages {
				t.Errorf("pages = %d, 期望 %d", resp.Pages, tc.wantPages)
			}
			if resp.Page != tc.wantPage {
				t.Errorf("page = %d, 期望 %d", resp.Page, tc.wantPage)
			}
			if resp.Date != "2026-05-01" && strings.Contains(tc.target, "2026-05-01") {
				t.Errorf("date = %q, 期望 2026-05-01", resp.Date)
			}
		})
	}
}

// TestFilesHandlerPageDoesNotPanic 显式覆盖曾经会 panic 的输入。
func TestFilesHandlerPageDoesNotPanic(t *testing.T) {
	seedIndex(t, map[string]int{"2026-05-01": 10})
	for _, page := range []string{"0", "-1", "-999999999", "abc", ""} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("page=%q 触发 panic: %v", page, r)
				}
			}()
			_, _ = callFilesHandler(t, "/api/files/2026-05-01?page="+page)
		}()
	}
}

func TestParsePageParam(t *testing.T) {
	ok := []struct {
		raw  string
		want int
	}{
		{"", 1},
		{"1", 1},
		{"42", 42},
		{"10000000", 10000000},
	}
	for _, c := range ok {
		got, err := parsePageParam(c.raw)
		if err != nil || got != c.want {
			t.Errorf("parsePageParam(%q) = (%d, %v), 期望 (%d, nil)", c.raw, got, err, c.want)
		}
	}

	bad := []string{"abc", "1.5", "0", "-1", "-10000000", "10000001", "99999999999999999999", " ", "1e3"}
	for _, raw := range bad {
		if got, err := parsePageParam(raw); err == nil {
			t.Errorf("parsePageParam(%q) 期望报错，实际 = (%d, nil)", raw, got)
		}
	}
}
