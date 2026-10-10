package main

import (
	"os"
	"testing"
)

// 这两个测试在本机做真实探测（不打桩），用于验证「不再被 WindowsApps 应用执行别名
// 卡住」以及「自动探测只会返回真正可用的解释器」。
// 没有可用 Python 的机器上会 Skip，不会误报失败。

func TestPythonCandidateDiagnostics(t *testing.T) {
	d := newTestDownloader("", "")
	candidates := d.pythonCandidates()
	if len(candidates) == 0 {
		t.Skip("没有候选解释器")
	}
	usable := 0
	for _, c := range candidates {
		if _, err := os.Stat(c); err != nil {
			t.Logf("跳过（文件不存在）: %s", c)
			continue
		}
		if err := validatePython(c); err != nil {
			t.Logf("拒绝（验证失败）: %s -> %v", c, err)
			continue
		}
		usable++
		t.Logf("可用: %s", c)
	}
	t.Logf("候选 %d 个，其中通过 --version 验证的 %d 个", len(candidates), usable)
}

func TestFindPythonAutoDetectReturnsUsableInterpreter(t *testing.T) {
	d := newTestDownloader("", "")
	py, err := d.findPython()
	if err != nil {
		// 自动探测失败必须是「能说明原因」的错误，而不是 panic 或静默使用坏路径
		if err.Error() == "" {
			t.Fatal("自动探测失败时应给出原因")
		}
		t.Skipf("本机没有可用 Python（自动探测已正确报错）: %v", err)
	}
	if err := validatePython(py); err != nil {
		t.Fatalf("findPython 返回了未经证实的解释器 %s: %v", py, err)
	}
	t.Logf("自动探测选中: %s", py)
}
