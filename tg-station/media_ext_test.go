package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 媒体目录里混进无关文件时，网页既不能读也不能删（纵深防御）。
// 允许的只有索引里会出现的图片/视频后缀。

func TestIsAllowedMediaExt(t *testing.T) {
	allowed := []string{"a.jpg", "A.JPEG", "pic.png", "x.gif", "y.webp",
		"v.mp4", "v.avi", "v.mov", "v.mkv", "v.webm", "sub/dir/a.jpg"}
	denied := []string{"config.json", "tg_session.session", "notes.txt", "run.exe",
		"archive.zip", "doc.pdf", "noext", "a.jpg.bak", ".env", "a.js", "a.html"}
	for _, n := range allowed {
		if !isAllowedMediaExt(n) {
			t.Errorf("%q 应被允许（图片/视频）", n)
		}
	}
	for _, n := range denied {
		if isAllowedMediaExt(n) {
			t.Errorf("%q 不该被允许（非媒体类型）", n)
		}
	}
}

func TestFileHandlerRejectsNonMediaFiles(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	// 模拟「媒体目录里混进了不该在的东西」
	writeFileT(t, filepath.Join(env.root, "config.json"), `{"api_hash":"LEAKED"}`)
	writeFileT(t, filepath.Join(env.root, "tg_session.session"), "LEAKED-SESSION")
	writeFileT(t, filepath.Join(env.root, "notes.txt"), "LEAKED-NOTES")
	writeFileT(t, filepath.Join(env.root, "payload.exe"), "MZ")
	writeFileT(t, filepath.Join(env.root, "noext"), "LEAKED-NOEXT")
	writeFileT(t, filepath.Join(env.root, "sub", "deep.json"), `{"deep":"LEAKED"}`)

	for _, target := range []string{"/api/file/config.json", "/api/file/tg_session.session",
		"/api/file/notes.txt", "/api/file/payload.exe", "/api/file/noext", "/api/file/sub/deep.json"} {
		rec := doRequest(fileHandler, http.MethodGet, target, nil)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s 应 403，实际 %d", target, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "LEAKED") {
			t.Errorf("%s 泄露了文件内容", target)
		}
	}

	// 正常图片/视频不受影响
	for _, target := range []string{"/api/file/pic.jpg", "/api/file/sub/deep.jpg", "/api/file/clip.mp4"} {
		if rec := doRequest(fileHandler, http.MethodGet, target, nil); rec.Code != http.StatusOK {
			t.Errorf("%s 应 200，实际 %d", target, rec.Code)
		}
	}
}

func TestDeleteHandlerRefusesNonMediaFiles(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	cfg := filepath.Join(env.root, "config.json")
	sess := filepath.Join(env.root, "tg_session.session")
	writeFileT(t, cfg, `{"api_hash":"LEAKED"}`)
	writeFileT(t, sess, "LEAKED-SESSION")

	body, _ := json.Marshal(map[string]interface{}{
		"files": []string{"config.json", "tg_session.session", "pic.jpg"},
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
	if resp.Deleted != 1 || resp.Failed != 2 {
		t.Errorf("期望 deleted=1 failed=2（只删掉图片），实际 deleted=%d failed=%d", resp.Deleted, resp.Failed)
	}
	if _, err := os.Stat(cfg); err != nil {
		t.Errorf("配置文件不得被删除: %v", err)
	}
	if _, err := os.Stat(sess); err != nil {
		t.Errorf("session 文件不得被删除: %v", err)
	}
	if _, err := os.Stat(filepath.Join(env.root, "pic.jpg")); !os.IsNotExist(err) {
		t.Error("正常图片应被删除")
	}
}

func TestThumbHandlerRejectsNonMediaFiles(t *testing.T) {
	env := setupMediaEnv(t)
	defer env.restore()

	writeFileT(t, filepath.Join(env.root, "config.json"), `{"api_hash":"LEAKED"}`)

	if rec := doRequest(thumbHandler, http.MethodGet, "/thumb/config.json", nil); rec.Code != http.StatusForbidden {
		t.Errorf("非媒体文件缩略图应 403，实际 %d", rec.Code)
	}
	if rec := doRequest(thumbHandler, http.MethodGet, "/thumb/pic.jpg", nil); rec.Code != http.StatusOK {
		t.Errorf("图片缩略图应 200，实际 %d", rec.Code)
	}
}
