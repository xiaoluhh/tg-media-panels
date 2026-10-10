package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ── 媒体路径安全校验测试（第二阶段 · 第一部分）────────────────────────────
//
// 覆盖：正常文件 / 子目录文件 / .. 穿越 / 同前缀兄弟目录 / 绝对路径 /
// 盘符与 UNC / 符号链接（根内允许、根外拒绝、删除不跟随）/ 文件不存在。
//
// 符号链接用例在无权限的系统（Windows 非管理员且未开开发者模式）上会自动跳过，
// 跳过原因会打印到测试日志。

func mkTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
}

// trySymlink 尝试创建符号链接；系统不支持时返回 false（调用方跳过用例）。
func trySymlink(t *testing.T, oldname, newname string) bool {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		t.Logf("当前系统无法创建符号链接（%v），跳过该符号链接用例", err)
		return false
	}
	return true
}

func TestSafeMediaPathAllowsLegitPaths(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	mkTestFile(t, filepath.Join(root, "sub", "b.jpg"), "i")
	mkTestFile(t, filepath.Join(root, "sub", "deep", "c.png"), "i")

	cases := []struct {
		name    string
		req     string
		wantRel string
	}{
		{"根目录下的文件", "a.mp4", "a.mp4"},
		{"子目录中的文件", "sub/b.jpg", "sub/b.jpg"},
		{"多层子目录", "sub/deep/c.png", "sub/deep/c.png"},
		{"带 ./ 前缀", "./a.mp4", "a.mp4"},
		{"重复分隔符", "sub//b.jpg", "sub/b.jpg"},
		{"URL 编码的中文路径仍可解析", "sub/%E4%B8%AD%E6%96%87.jpg", ""},
	}
	// 中文字符串用例单独准备文件（避免与上面 table 混淆）
	mkTestFile(t, filepath.Join(root, "sub", "中文.jpg"), "i")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mp, err := safeMediaPath(root, tc.req)
			if err != nil {
				t.Fatalf("safeMediaPath(%q) 意外失败: %v", tc.req, err)
			}
			if !strings.HasSuffix(filepath.ToSlash(mp.Real), tc.wantRel) && tc.wantRel != "" {
				t.Errorf("Real = %q，期望以 %q 结尾", mp.Real, tc.wantRel)
			}
			if mp.Rel == "" {
				t.Errorf("Rel 不能为空")
			}
			if !filepath.IsAbs(mp.Full) || !filepath.IsAbs(mp.Real) {
				t.Errorf("Full/Real 必须是绝对路径: %q / %q", mp.Full, mp.Real)
			}
		})
	}
}

func TestSafeMediaPathRejectsUnsafePaths(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	sibling := filepath.Join(base, "media2")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	mkTestFile(t, filepath.Join(root, "sub", "b.jpg"), "i")
	mkTestFile(t, filepath.Join(sibling, "c.mp4"), "v")
	outside := filepath.Join(base, "secret.txt")
	mkTestFile(t, outside, "s")

	cases := []struct {
		name string
		req  string
	}{
		{"目录穿越", "../secret.txt"},
		{"深层穿越", "sub/../../secret.txt"},
		{"仅 ..", ".."},
		{"结尾 ..", "sub/.."},
		{"反斜杠穿越", `..\secret.txt`},
		{"同前缀兄弟目录（绝对路径）", outside},
		{"同前缀兄弟目录（sibling）", filepath.Join(sibling, "c.mp4")},
		{"绝对路径（正斜杠）", filepath.ToSlash(outside)},
		{"URL 形式", "file:///etc/passwd"},
		{"空路径", ""},
		{"纯空白", "   "},
		{"NUL 字节", "a\x00.mp4"},
		{"反斜杠普通路径", `sub\b.jpg`},
		{"协议式前缀", "http://example.com/a.mp4"},
	}
	if runtime.GOOS == "windows" {
		// Windows 会静默改写这些路径，导致「校验的路径」≠「打开的文件」
		cases = append(cases,
			struct {
				name string
				req  string
			}{"尾随空格（Windows 静默改写）", "a.mp4 "},
			struct {
				name string
				req  string
			}{"尾随点（Windows 静默改写）", "sub/b.jpg."},
			struct {
				name string
				req  string
			}{"NTFS 数据流", "a.mp4:secret"},
			struct {
				name string
				req  string
			}{"保留设备名", "NUL"},
			struct {
				name string
				req  string
			}{"保留设备名带扩展", "CON.mp4"},
		)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := safeMediaPath(root, tc.req)
			if err == nil {
				t.Fatalf("safeMediaPath(%q) 应当被拒绝，实际通过", tc.req)
			}
			if !errors.Is(err, ErrPathOutside) && !errors.Is(err, ErrPathMissing) {
				t.Fatalf("错误类型应为 ErrPathOutside/ErrPathMissing，实际: %v", err)
			}
		})
	}
}

// TestIsWithinPrefixBypass 直击旧实现的前缀绕过缺陷。
func TestIsWithinPrefixBypass(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "media")
	sibling := filepath.Join(base, "media2")

	// 先确认这两个路径确实构成「字符串前缀相同」的诱饵
	if !strings.HasPrefix(filepath.Clean(sibling), filepath.Clean(root)) {
		t.Fatalf("测试前提不成立: %q 不是 %q 的前缀", sibling, root)
	}
	if isWithin(root, sibling) {
		t.Fatalf("同前缀兄弟目录 %q 被误判为在 %q 之内", sibling, root)
	}
	if isWithin(root, root) {
		t.Fatalf("根目录自身不应被当作其内部的文件路径")
	}
	if !isWithin(root, filepath.Join(root, "a.mp4")) {
		t.Fatalf("根目录内的文件应判为在内")
	}
	if isWithin(root, filepath.Join(root, "..", "x.mp4")) {
		t.Fatalf("经 .. 逃出根目录的路径不应判为在内")
	}
}

func TestSafeMediaPathMissingFile(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")

	_, err := safeMediaPath(root, "not-there.mp4")
	if !errors.Is(err, ErrPathMissing) {
		t.Fatalf("不存在的文件应返回 ErrPathMissing，实际: %v", err)
	}
	// 不存在的深层路径同样报缺失（而不是越界）
	if _, err := safeMediaPath(root, "sub/not-there.mp4"); !errors.Is(err, ErrPathMissing) {
		t.Fatalf("不存在的子路径应返回 ErrPathMissing，实际: %v", err)
	}
}

func TestSafeMediaPathSymlinks(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	root2 := filepath.Join(base, "media-link")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	outsideDir := filepath.Join(base, "outside")
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	mkTestFile(t, outsideFile, "s")

	// 1) 指向根内文件的符号链接 → 允许，且 Real 必须落在根内
	if trySymlink(t, filepath.Join(root, "a.mp4"), filepath.Join(root, "link-in.mp4")) {
		mp, err := safeMediaPath(root, "link-in.mp4")
		if err != nil {
			t.Fatalf("根内符号链接应被允许，实际: %v", err)
		}
		if !mp.Symlink {
			t.Errorf("应当标记 Symlink=true")
		}
		if !isWithin(mustReal(t, root), mp.Real) {
			t.Errorf("Real=%q 应位于媒体根目录内", mp.Real)
		}
	}

	// 2) 指向根外文件的符号链接 → 拒绝
	if trySymlink(t, outsideFile, filepath.Join(root, "link-out.mp4")) {
		if _, err := safeMediaPath(root, "link-out.mp4"); !errors.Is(err, ErrPathOutside) {
			t.Fatalf("指向根外文件的符号链接必须拒绝，实际: %v", err)
		}
		// 删除同样必须拒绝，且根外文件不能被动到
		if _, err := safeDeleteTarget(root, "link-out.mp4"); !errors.Is(err, ErrPathOutside) {
			t.Fatalf("通过指向根外的链接删除必须拒绝，实际: %v", err)
		}
		if _, err := os.Stat(outsideFile); err != nil {
			t.Fatalf("根外文件不应被删除: %v", err)
		}
	}

	// 3) 指向根外目录的符号链接（经由链接访问目录内文件）→ 拒绝
	if trySymlink(t, outsideDir, filepath.Join(root, "outdir")) {
		if _, err := safeMediaPath(root, "outdir/secret.txt"); !errors.Is(err, ErrPathOutside) {
			t.Fatalf("经由指向根外目录的链接必须拒绝，实际: %v", err)
		}
	}

	// 4) 媒体根目录本身是符号链接 → 仍按真实根校验
	if trySymlink(t, root, root2) {
		resetMediaRootCache()
		if _, err := safeMediaPath(root2, "a.mp4"); err != nil {
			t.Fatalf("以符号链接形式给出的媒体根目录应可用: %v", err)
		}
		if _, err := safeMediaPath(root2, "../secret.txt"); err == nil {
			t.Fatalf("经符号链接根目录穿越必须被拒绝")
		}
	}
}

func TestSafeDeleteTarget(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "del.mp4"), "v")
	if err := os.MkdirAll(filepath.Join(root, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// 普通文件可删
	mp, err := safeDeleteTarget(root, "del.mp4")
	if err != nil {
		t.Fatalf("普通文件应允许删除: %v", err)
	}
	if err := os.Remove(mp.Full); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if _, err := os.Stat(mp.Full); !os.IsNotExist(err) {
		t.Fatalf("文件应已被删除")
	}

	// 目录不允许删除
	if _, err := safeDeleteTarget(root, "dir"); err == nil {
		t.Fatalf("目录不应允许通过该接口删除")
	}
	// 穿越不允许删除
	if _, err := safeDeleteTarget(root, "../secret.txt"); err == nil {
		t.Fatalf("穿越路径不应允许删除")
	}
	// 不存在 → ErrPathMissing
	if _, err := safeDeleteTarget(root, "nope.mp4"); !errors.Is(err, ErrPathMissing) {
		t.Fatalf("不存在文件应返回 ErrPathMissing，实际: %v", err)
	}

	// 根内链接：删除只能删链接本身，不能删目标
	if trySymlink(t, filepath.Join(root, "del.mp4"), filepath.Join(root, "del-link.mp4")) {
		mkTestFile(t, filepath.Join(root, "del.mp4"), "v2")
		mp, err := safeDeleteTarget(root, "del-link.mp4")
		if err != nil {
			t.Fatalf("指向根内文件的链接应允许删除（删链接本身）: %v", err)
		}
		if err := os.Remove(mp.Full); err != nil {
			t.Fatalf("删除链接失败: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "del.mp4")); err != nil {
			t.Fatalf("链接目标不应被删除: %v", err)
		}
	}
}

func TestSafeMediaPathWindowsSpecific(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")

	if runtime.GOOS != "windows" {
		t.Skip("盘符 / UNC 行为仅在 Windows 上断言（Linux 上它们是普通文件名）")
	}
	bad := []string{
		`C:\Windows\win.ini`,
		`C:/Windows/win.ini`,
		`C:win.ini`,              // 盘符相对路径
		`\\server\share\x.mp4`,   // UNC
		`\\?\C:\Windows\win.ini`, // 扩展长度前缀
	}
	for _, req := range bad {
		if _, err := safeMediaPath(root, req); err == nil {
			t.Errorf("Windows 下 %q 必须被拒绝", req)
		}
	}
	// 反斜杠形式的正常相对路径在 Windows 上也会被拒绝（前端只用正斜杠）
	if _, err := safeMediaPath(root, `a.mp4`); err != nil {
		t.Errorf("普通相对路径不应失败: %v", err)
	}
}

// tryJunction 尝试创建 Windows 目录联接（junction，无需管理员权限）。
// 其它平台或创建失败时返回 false（调用方跳过用例）。
func tryJunction(t *testing.T, target, link string) bool {
	t.Helper()
	if runtime.GOOS != "windows" {
		return false
	}
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Logf("创建 junction 失败（%v）：%s", err, strings.TrimSpace(string(out)))
		return false
	}
	return true
}

func TestSafeMediaPathJunctions(t *testing.T) {
	resetMediaRootCache()
	base := t.TempDir()
	root := filepath.Join(base, "media")
	mkTestFile(t, filepath.Join(root, "a.mp4"), "v")
	mkTestFile(t, filepath.Join(root, "sub", "b.jpg"), "i")
	outsideDir := filepath.Join(base, "outside")
	mkTestFile(t, filepath.Join(outsideDir, "secret.txt"), "s")

	did := false

	// 1) 指向根内目录的 junction → 允许，且 Real 必须落在真实根内
	inner := filepath.Join(root, "inner-link")
	if tryJunction(t, filepath.Join(root, "sub"), inner) {
		did = true
		mp, err := safeMediaPath(root, "inner-link/b.jpg")
		if err != nil {
			t.Errorf("指向根内目录的 junction 应被允许，实际: %v", err)
		} else if !isWithin(mustReal(t, root), mp.Real) {
			t.Errorf("Real=%q 应位于媒体根目录内", mp.Real)
		}
	}

	// 2) 指向根外目录的 junction → 必须拒绝（这是 EvalSymlinks 抓不到的绕过）
	outLink := filepath.Join(root, "out-link")
	if tryJunction(t, outsideDir, outLink) {
		did = true
		if _, err := safeMediaPath(root, "out-link/secret.txt"); !errors.Is(err, ErrPathOutside) {
			t.Errorf("指向根外目录的 junction 必须拒绝，实际: %v", err)
		}
		if _, err := safeDeleteTarget(root, "out-link/secret.txt"); err == nil {
			t.Errorf("经由 junction 删除根外文件必须拒绝")
		}
		if _, err := os.Stat(filepath.Join(outsideDir, "secret.txt")); err != nil {
			t.Errorf("根外文件不应被删除: %v", err)
		}
	}

	// 3) 媒体根目录本身是 junction → 仍按真实根校验
	rootLink := filepath.Join(base, "media-link")
	if tryJunction(t, root, rootLink) {
		did = true
		resetMediaRootCache()
		if _, err := safeMediaPath(rootLink, "a.mp4"); err != nil {
			t.Errorf("以 junction 形式给出的媒体根目录应可用: %v", err)
		}
		if _, err := safeMediaPath(rootLink, "../outside/secret.txt"); err == nil {
			t.Errorf("经 junction 根目录穿越必须被拒绝")
		}
	}

	if !did {
		t.Skip("当前系统无法创建 junction（非 Windows 或命令不可用）")
	}
}

func mustReal(t *testing.T, path string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) 失败: %v", path, err)
	}
	return real
}
