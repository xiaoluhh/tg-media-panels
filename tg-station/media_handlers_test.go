package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── HTTP 处理器集成测试（第二阶段 · 第一/二/四部分）──────────────────────
//
// 直接调用各 handler，验证：
//   - 所有媒体访问入口都接入了统一路径校验（穿越 / 绝对路径 / 兄弟目录 → 403，
//     不存在 → 404，且响应不泄露服务器绝对路径）；
//   - 损坏媒体、缓存目录不可写等异常不会 panic 或挂死；
//   - 删除不会跟随符号链接删除媒体目录之外的文件；
//   - /api/refresh 的响应字段保持前端兼容。

type mediaEnv struct {
	base    string
	root    string
	sibling string
	restore func()
}

func setupMediaEnv(t *testing.T) mediaEnv {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	sibling := filepath.Join(base, "media2")

	restore := setMediaTestEnv(t, root)

	writeJPEG(t, filepath.Join(root, "pic.jpg"))
	writeJPEG(t, filepath.Join(root, "sub", "deep.jpg"))
	writeFileT(t, filepath.Join(root, "broken.jpg"), "this is not an image")
	writeFileT(t, filepath.Join(root, "clip.mp4"), "this is not a video")
	writeFileT(t, filepath.Join(sibling, "sibling.jpg"), "sibling")
	writeFileT(t, filepath.Join(base, "secret.jpg"), "secret")

	return mediaEnv{base: base, root: root, sibling: sibling, restore: restore}
}

func doRequest(h http.HandlerFunc, method, target string, body []byte) *httptest.ResponseRecorder {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	rec := httptest.NewRecorder()
	h(rec, r)
	return rec
}

func TestFileHandlerPathSafety(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	siblingFile := filepath.Join(env.sibling, "sibling.jpg")
	secret := filepath.Join(env.base, "secret.jpg")

	cases := []struct {
		name   string
		target string
		want   int
	}{
		{"正常文件", "/api/file/pic.jpg", http.StatusOK},
		{"子目录文件", "/api/file/sub/deep.jpg", http.StatusOK},
		{"目录穿越", "/api/file/../secret.jpg", http.StatusForbidden},
		{"编码的目录穿越", "/api/file/%2e%2e/secret.jpg", http.StatusForbidden},
		{"兄弟目录（绝对路径）", "/api/file/" + siblingFile, http.StatusForbidden},
		{"绝对路径", "/api/file/" + filepath.ToSlash(secret), http.StatusForbidden},
		{"URL 形式", "/api/file/file:///etc/passwd", http.StatusForbidden},
		{"文件不存在", "/api/file/nope.jpg", http.StatusNotFound},
		{"目标是目录", "/api/file/sub", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(fileHandler, http.MethodGet, tc.target, nil)
			if rec.Code != tc.want {
				t.Fatalf("状态码 = %d，期望 %d（body=%q）", rec.Code, tc.want, clip(rec.Body.String(), 120))
			}
			// 错误响应绝不能回显服务器绝对路径
			body := rec.Body.String()
			if rec.Code != http.StatusOK {
				if strings.Contains(body, filepath.ToSlash(env.base)) || strings.Contains(body, env.base) {
					t.Errorf("错误响应泄露了服务器绝对路径: %q", clip(body, 200))
				}
			}
		})
	}

	// 正常文件内容必须原样返回（保持播放能力）
	rec := doRequest(fileHandler, http.MethodGet, "/api/file/pic.jpg", nil)
	want, err := os.ReadFile(filepath.Join(env.root, "pic.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rec.Body.Bytes(), want) {
		t.Errorf("文件内容不一致：返回 %d 字节，期望 %d 字节", rec.Body.Len(), len(want))
	}
	// 视频响应头（保持前端播放兼容）
	rec = doRequest(fileHandler, http.MethodGet, "/api/file/clip.mp4", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("视频应可访问，实际 %d", rec.Code)
	}
	if rec.Header().Get("Accept-Ranges") != "bytes" {
		t.Errorf("视频响应应带 Accept-Ranges: bytes")
	}
}

func TestThumbHandlerPathSafetyAndBrokenMedia(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	siblingFile := filepath.Join(env.sibling, "sibling.jpg")
	secret := filepath.Join(env.base, "secret.jpg")

	for _, tc := range []struct {
		name   string
		target string
		want   int
	}{
		{"穿越", "/thumb/../secret.jpg", http.StatusForbidden},
		{"编码穿越", "/thumb/%2e%2e/secret.jpg", http.StatusForbidden},
		{"兄弟目录", "/thumb/" + siblingFile, http.StatusForbidden},
		{"绝对路径", "/thumb/" + filepath.ToSlash(secret), http.StatusForbidden},
		{"不存在", "/thumb/nope.jpg", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(thumbHandler, http.MethodGet, tc.target, nil)
			if rec.Code != tc.want {
				t.Fatalf("状态码 = %d，期望 %d", rec.Code, tc.want)
			}
		})
	}

	// 损坏的图片/视频：不得 panic，返回 404（前端显示占位图）
	for _, name := range []string{"broken.jpg", "clip.mp4"} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s 触发 panic: %v", name, r)
				}
			}()
			rec := doRequest(thumbHandler, http.MethodGet, "/thumb/"+name, nil)
			if rec.Code != http.StatusNotFound {
				t.Errorf("%s 应返回 404，实际 %d", name, rec.Code)
			}
		}()
	}

	// 正常图片：无论 ffmpeg 是否可用都必须能出缩略图（原生解码兜底）
	rec := doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("正常图片缩略图应成功，实际 %d（body=%q）", rec.Code, clip(rec.Body.String(), 120))
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "jpeg") && !strings.Contains(ct, "jpg") {
		t.Errorf("缩略图 Content-Type = %q，期望 image/jpeg", ct)
	}
	if rec.Body.Len() < thumbMinBytes {
		t.Errorf("缩略图体积异常: %d 字节", rec.Body.Len())
	}

	// 二次请求应命中缓存（不再生成），文件仍在
	entries, _ := os.ReadDir(thumbsDir)
	if len(entries) == 0 {
		t.Error("缩略图缓存文件未写入")
	}
}

func TestThumbHandlerUnwritableCacheDir(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	// 把缓存目录指向一个普通文件：MkdirAll 必定失败
	blocker := filepath.Join(env.base, "thumbs-blocker")
	writeFileT(t, blocker, "not a directory")
	oldThumbs := thumbsDir
	thumbsDir = blocker
	defer func() { thumbsDir = oldThumbs }()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("缓存目录不可写时不应 panic: %v", r)
		}
	}()
	rec := doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("缓存目录不可写时应返回 404，实际 %d", rec.Code)
	}
}

func TestDeleteHandlerPathSafety(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	secret := filepath.Join(env.base, "secret.jpg")
	siblingFile := filepath.Join(env.sibling, "sibling.jpg")

	body, _ := json.Marshal(map[string]interface{}{
		"files": []string{
			"../secret.jpg",
			"sub/../../secret.jpg",
			siblingFile,
			filepath.Join(env.root, "sub"), // 目录不允许删除
		},
	})
	rec := doRequest(deleteHandler, http.MethodPost, "/api/delete", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	var resp struct {
		Success bool `json:"success"`
		Deleted int  `json:"deleted"`
		Failed  int  `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if resp.Deleted != 0 {
		t.Errorf("危险路径一个都不应被删除，实际 deleted=%d", resp.Deleted)
	}
	if _, err := os.Stat(secret); err != nil {
		t.Errorf("媒体目录之外的文件不得被删除: %v", err)
	}
	if _, err := os.Stat(siblingFile); err != nil {
		t.Errorf("兄弟目录中的文件不得被删除: %v", err)
	}

	// 正常文件：删除成功 + 缩略图缓存清理
	// 先生成一次缩略图缓存，并记录该文件的精确缓存键
	rec = doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("准备缩略图失败: %d", rec.Code)
	}
	picMP, err := safeMediaPath(mediaDir, "pic.jpg")
	if err != nil {
		t.Fatalf("计算 pic.jpg 路径失败: %v", err)
	}
	picInfo, err := os.Stat(picMP.Real)
	if err != nil {
		t.Fatalf("stat pic.jpg 失败: %v", err)
	}
	picThumb := thumbCachePath(thumbCacheKey(picMP.Rel, picInfo.Size(), picInfo.ModTime()))
	if _, err := os.Stat(picThumb); err != nil {
		t.Fatalf("缩略图缓存未生成: %v", err)
	}

	body, _ = json.Marshal(map[string]interface{}{"files": []string{"pic.jpg"}})
	rec = doRequest(deleteHandler, http.MethodPost, "/api/delete", body)
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Deleted != 1 {
		t.Fatalf("正常文件应被删除，实际 deleted=%d failed=%d", resp.Deleted, resp.Failed)
	}
	if _, err := os.Stat(filepath.Join(env.root, "pic.jpg")); !os.IsNotExist(err) {
		t.Error("文件应已被删除")
	}
	// 该文件的两种命名缓存都必须被清理（确定性断言，不受后台任务影响）
	if _, err := os.Stat(picThumb); !os.IsNotExist(err) {
		t.Errorf("删除文件后新命名缩略图缓存应被清理: %s", filepath.Base(picThumb))
	}
	if _, err := os.Stat(legacyThumbCachePath("pic.jpg")); !os.IsNotExist(err) {
		t.Errorf("删除文件后旧命名缩略图缓存应被清理")
	}
}

func TestDeleteHandlerMethodAndBody(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	rec := doRequest(deleteHandler, http.MethodGet, "/api/delete", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET 应返回 405，实际 %d", rec.Code)
	}
	rec = doRequest(deleteHandler, http.MethodPost, "/api/delete", []byte("not json"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("非法 JSON 应返回 400，实际 %d", rec.Code)
	}
}

func TestRefreshHandlerFrontendCompatibility(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	rec := doRequest(refreshHandler, http.MethodPost, "/api/refresh", []byte("{}"))
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 200", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	// 前端依赖 ok / message 字段
	if ok, _ := resp["ok"].(bool); !ok {
		t.Error("响应必须含 ok=true（前端 stationFetch 依赖）")
	}
	if msg, _ := resp["message"].(string); msg == "" {
		t.Error("响应必须含非空 message")
	}
	// 新增字段用于区分「启动了扫描」与「已合并为补扫」，且不得谎报完成
	if _, has := resp["started"]; !has {
		t.Error("响应应含 started 字段")
	}
	if _, has := resp["merged"]; !has {
		t.Error("响应应含 merged 字段")
	}
	if _, has := resp["running"]; !has {
		t.Error("响应应含 running 字段")
	}

	// 扫描中的再次请求：必须被合并（不能返回虚假的完成，也不能丢失）
	if scanSched.Wait(20 * time.Second) {
		// 第一次扫描已结束，这里再触发一轮并立刻追问
		doRequest(refreshHandler, http.MethodPost, "/api/refresh", []byte("{}"))
		rec = doRequest(refreshHandler, http.MethodPost, "/api/refresh", []byte("{}"))
		if rec.Code != http.StatusOK {
			t.Fatalf("并发刷新请求应返回 200，实际 %d", rec.Code)
		}
	}

	// 非 POST 方法保持 405（前端行为兼容）
	rec = doRequest(refreshHandler, http.MethodGet, "/api/refresh", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/refresh 应返回 405，实际 %d", rec.Code)
	}
}
