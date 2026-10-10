package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// registerAPIRoutes 挂载整合后新增的全部 API 路由。
// 必须在 http.ListenAndServe 之前调用。
func registerAPIRoutes() {
	http.HandleFunc("/api/status", statusHandler)                       // 综合状态（监听器 + 媒体索引）
	http.HandleFunc("/api/listen/start", listenStartHandler)           // 启动下载器
	http.HandleFunc("/api/listen/stop", listenStopHandler)             // 停止下载器
	http.HandleFunc("/api/listen/restart", listenRestartHandler)       // 重启下载器
	http.HandleFunc("/api/listen/status", listenStatusHandler)         // 下载器状态
	http.HandleFunc("/api/download/progress", downloadProgressHandler) // 下载进度（progress.json + dl_status.json）
	http.HandleFunc("/api/download/cancel", downloadCancelHandler)     // 取消单个下载任务（写 cancel.json）
	http.HandleFunc("/api/settings", settingsHandler)                  // 读写配置（GET / POST）
	http.HandleFunc("/api/refresh", refreshHandler)                    // 刷新媒体索引（Python 下载完成回调）
	http.HandleFunc("/api/login/qr", loginQrHandler)                   // 扫码登录：生成二维码 / 检查已登录
	http.HandleFunc("/api/login/status", loginStatusHandler)           // 登录状态：检查 session 文件
	http.HandleFunc("/api/namelist", nameListHandler)                  // 名单文件 (mz.txt) 读写：GET 读取 / POST 保存
	http.HandleFunc("/api/channels", channelsHandler)                  // 频道列表：列出账号已加入的频道（设置面板选择器）
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]interface{}{"ok": false, "error": msg})
}

// ---------- /api/status ----------

func statusHandler(w http.ResponseWriter, r *http.Request) {
	ls := downloader.status()

	indexMutex.RLock()
	total := len(mediaIndex)
	last := lastScan
	indexMutex.RUnlock()

	lastScanStr := ""
	if !last.IsZero() {
		lastScanStr = last.Format(time.RFC3339)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"listener": map[string]interface{}{
			"state":     ls.State,
			"running":   ls.Running,
			"pid":       ls.Pid,
			"startedAt": ls.StartedAt,
			"exitCode":  ls.ExitCode,
			"exitMsg":   ls.ExitMsg,
			"autoStart": downloader.cfg != nil && downloader.cfg.Station.AutoStart,
		},
		"media": map[string]interface{}{
			"total":    total,
			"lastScan": lastScanStr,
			"dir":      mediaDir,
		},
		"movie_min_minutes": movieMinMinutes(),
	})
}

// movieMinMinutes 返回影视筛选阈值（config.json movie_min_minutes，默认 5 分钟）
func movieMinMinutes() float64 {
	minutes := 5.0
	if downloader.cfg != nil {
		if v, ok := downloader.cfg.Raw["movie_min_minutes"]; ok {
			if f, ok := v.(float64); ok && f > 0 {
				minutes = f
			}
		}
	}
	return minutes
}

// ---------- /api/listen/* ----------

func listenStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	if err := downloader.Start(); err != nil {
		log.Printf("[station] 启动下载器失败: %v", err)
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ls := downloader.status()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "下载器已启动",
		"state":   ls.State,
		"pid":     ls.Pid,
	})
}

func listenStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	if err := downloader.Stop(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "下载器已停止",
		"state":   downloader.status().State,
	})
}

func listenRestartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	if err := downloader.Restart(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ls := downloader.status()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "下载器已重启",
		"state":   ls.State,
		"pid":     ls.Pid,
	})
}

func listenStatusHandler(w http.ResponseWriter, r *http.Request) {
	ls := downloader.status()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"listener": ls,
	})
}

// ---------- /api/download/progress ----------

// downloadProgressHandler 返回下载进度：progress.json（正在下载的任务）+ dl_status.json（最近完成记录）。
// 两个文件都是下载器（tg_auto_download.py）在运行目录写的 UTF-8 JSON，兼容 v1 悬浮窗时代格式。
func downloadProgressHandler(w http.ResponseWriter, r *http.Request) {
	base := filepath.Dir(configPath) // 运行目录（config.json 所在处，与下载器 BASE_DIR 一致）

	// progress.json：只取 pct<100 的任务（pct=100 是中断残留/已完成，不算正在下载）
	tasks := []map[string]interface{}{}
	if b, err := os.ReadFile(filepath.Join(base, "progress.json")); err == nil {
		var p struct {
			Tasks []map[string]interface{} `json:"tasks"`
		}
		if json.Unmarshal(b, &p) == nil {
			for _, t := range p.Tasks {
				if pct, _ := t["pct"].(float64); pct < 100 {
					tasks = append(tasks, t)
				}
			}
		}
	}

	// dl_status.json：记录只有 HH:MM:SS 无日期，附加文件 mtime 的日期（MM-DD）防跨天误导
	recent := []map[string]interface{}{}
	if b, err := os.ReadFile(filepath.Join(base, "dl_status.json")); err == nil {
		if json.Unmarshal(b, &recent) == nil {
			if st, err2 := os.Stat(filepath.Join(base, "dl_status.json")); err2 == nil {
				date := st.ModTime().Format("01-02")
				for _, r := range recent {
					if _, has := r["date"]; !has {
						r["date"] = date
					}
				}
			}
		} else {
			recent = []map[string]interface{}{}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"tasks":  tasks,  // [{file, pct, done_mb, total_mb}]
		"recent": recent, // [{date, time, chat, file}]
	})
}

// ---------- /api/download/cancel ----------

// downloadCancelHandler 取消单个下载任务：把文件名写入 cancel.json，
// 下载器（tg_auto_download.py）每下载一段检查该文件，命中则放弃当前任务并清理残片。
func downloadCancelHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		File string `json:"file"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.File == "" {
		writeErr(w, http.StatusBadRequest, "缺少 file 参数")
		return
	}
	base := filepath.Dir(configPath)
	cf := filepath.Join(base, "cancel.json")
	// 读现有列表，追加（去重）
	lst := []string{}
	if b, err := os.ReadFile(cf); err == nil {
		_ = json.Unmarshal(b, &lst)
	}
	for _, f := range lst {
		if f == req.File {
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message": "该任务已在取消列表中"})
			return
		}
	}
	lst = append(lst, req.File)
	if b, err := json.Marshal(lst); err == nil {
		if err := os.WriteFile(cf, b, 0644); err != nil {
			writeErr(w, http.StatusInternalServerError, "写入 cancel.json 失败: "+err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message": "已请求取消: " + req.File})
}

// ---------- /api/settings ----------

// settingsHandler GET 返回完整配置（供设置面板回填），POST 保存配置。
func settingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := loadConfig()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":     true,
			"config": cfg.Raw,
		})
	case http.MethodPost:
		saveSettings(w, r)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "仅支持 GET / POST")
	}
}

func saveSettings(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	var incoming map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}

	// 合并：前端提交的字段覆盖，未提交的字段保留
	merged := mergeMaps(cfg.Raw, incoming)

	// 基本校验：port 必须是合法数字
	if mb, ok := merged["media_browser"].(map[string]interface{}); ok {
		if p, ok := mb["port"].(float64); ok && (p < 1 || p > 65535) {
			writeErr(w, http.StatusBadRequest, fmt.Sprintf("端口不合法: %v", p))
			return
		}
	}

	cfg.Raw = merged
	applyRaw(cfg)

	// station 段单独校验
	if st, ok := incoming["station"].(map[string]interface{}); ok {
		bs, _ := json.Marshal(st)
		var newSt StationConfig
		if json.Unmarshal(bs, &newSt) == nil {
			cfg.Station = newSt
		}
	}

	if err := saveConfig(cfg); err != nil {
		writeErr(w, http.StatusInternalServerError, "保存配置失败: "+err.Error())
		return
	}

	// 配置可能影响下载器（api_id/api_hash/channels 等），运行中则自动重启以应用
	restarted := false
	if downloader.isRunning() {
		if err := downloader.Restart(); err == nil {
			restarted = true
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"message":   "设置已保存",
		"restarted": restarted,
	})
}

// mergeMaps 深度合并：incoming 的值覆盖 base；map 递归合并，其它类型直接替换
func mergeMaps(base, incoming map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+len(incoming))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range incoming {
		if bm, ok := out[k].(map[string]interface{}); ok {
			if im, ok2 := v.(map[string]interface{}); ok2 {
				out[k] = mergeMaps(bm, im)
				continue
			}
		}
		out[k] = v
	}
	return out
}

// ---------- /api/refresh ----------

// refreshHandler 重新扫描媒体目录并重建索引（异步），返回后前端轮询 /api/status。
// 由 Python 下载器下载完成后 POST 调用；也可手动触发。
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 POST")
		return
	}
	go func() {
		scanFiles()
		go generateAllThumbnails()
	}()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "正在刷新媒体索引",
	})
}

// ---------- /api/login/* ----------

// 扫码登录进程管理：二维码生成后进程需继续挂起等待扫码（最多 120 秒），
// 所以不能等进程退出——读到第一行 JSON 就返回，进程留待扫码/超时后自行退出。
var (
	qrProcMu sync.Mutex
	qrProc   *exec.Cmd // 正在等待扫码的进程（下次生成前杀掉，避免堆积）
)

// loginQrHandler 调用 Python 脚本的 --qr-only 模式：
// 已登录 → already + username；未登录 → qr（base64 二维码）；失败 → error。
// 二维码生成阶段超时 30 秒。
func loginQrHandler(w http.ResponseWriter, r *http.Request) {
	if downloader.isRunning() {
		writeErr(w, http.StatusConflict, "下载器正在运行，请先停止监听再扫码登录")
		return
	}
	py, err := downloader.findPython()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	script, err := downloader.findScript()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 作废上一张二维码：杀掉上一个等待扫码的进程
	qrProcMu.Lock()
	if qrProc != nil && qrProc.Process != nil {
		qrProc.Process.Kill()
		qrProc.Wait()
	}
	qrProc = nil
	qrProcMu.Unlock()

	cmd := exec.Command(py, script, "--qr-only")
	cmd.Dir = filepath.Dir(script)
	cmd.Env = append(os.Environ(),
		"TG_STATION_CONFIG="+configPath,
		"PYTHONUTF8=1",
		"PYTHONIOENCODING=utf-8",
	)
	cmd.SysProcAttr = childProcAttr()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "登录进程管道创建失败")
		return
	}
	if err := cmd.Start(); err != nil {
		writeErr(w, http.StatusInternalServerError, "登录进程启动失败: "+err.Error())
		return
	}

	// 读取第一行合法 JSON（二维码就绪 / 已登录 / 失败），进程继续挂起等扫码
	type lineResult struct {
		line []byte
		err  error
	}
	lineCh := make(chan lineResult, 1)
	go func() {
		br := bufio.NewReader(stdout)
		for {
			line, err := br.ReadString('\n')
			line = strings.TrimSpace(line)
			if line != "" && json.Valid([]byte(line)) {
				lineCh <- lineResult{line: []byte(line)}
				return
			}
			if err != nil {
				lineCh <- lineResult{err: err}
				return
			}
		}
	}()

	var raw []byte
	select {
	case res := <-lineCh:
		if res.err != nil && len(res.line) == 0 {
			cmd.Process.Kill()
			writeErr(w, http.StatusInternalServerError, "登录进程无输出")
			return
		}
		raw = res.line
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		writeErr(w, http.StatusGatewayTimeout, "登录请求超时（30 秒），请重试")
		return
	}

	// 收尸：扫码成功/超时后进程自行退出
	go func() { cmd.Wait() }()

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		writeErr(w, http.StatusInternalServerError, "登录进程输出异常")
		return
	}

	switch result["status"] {
	case "already", "success":
		username, _ := result["username"].(string)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":       true,
			"status":   result["status"],
			"username": username,
		})
	case "qr":
		b64, _ := result["qr_base64"].(string)
		if b64 == "" {
			writeErr(w, http.StatusInternalServerError, "二维码生成失败")
			return
		}
		// 记录等待中的进程，下次生成前作废
		qrProcMu.Lock()
		qrProc = cmd
		qrProcMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":     true,
			"status": "qr",
			"qr":     "data:image/png;base64," + b64,
		})
	default:
		msg, _ := result["error"].(string)
		if msg == "" {
			msg = "登录失败"
		}
		writeErr(w, http.StatusOK, msg)
	}
}

// loginStatusHandler 检查登录状态：调用 Python --check-session 本地判断
// session 是否有有效 auth key（不连网，毫秒级）。
// 进程调用失败时回退到文件存在性判断。
func loginStatusHandler(w http.ResponseWriter, r *http.Request) {
	py, err := downloader.findPython()
	script, err2 := downloader.findScript()

	if err == nil && err2 == nil {
		cmd := exec.Command(py, script, "--check-session")
		cmd.Dir = filepath.Dir(script)
		cmd.Env = append(os.Environ(),
			"TG_STATION_CONFIG="+configPath,
			"PYTHONUTF8=1",
			"PYTHONIOENCODING=utf-8",
		)
		cmd.SysProcAttr = childProcAttr()
		if out, err := cmd.Output(); err == nil {
			var result map[string]interface{}
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				line := strings.TrimSpace(lines[i])
				if line == "" {
					continue
				}
				if json.Unmarshal([]byte(line), &result) == nil {
					break
				}
			}
			if result != nil {
				switch result["status"] {
				case "logged_in", "not_logged_in":
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"ok":     true,
						"status": result["status"],
					})
					return
				}
			}
		}
	}

	// 回退：session 文件存在性检查
	cfg, err := loadConfig()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	session := getString(cfg.Raw, []string{"session"}, "")
	loggedIn := false
	if session != "" {
		if _, err := os.Stat(session); err == nil {
			loggedIn = true
		} else if _, err := os.Stat(session + ".session"); err == nil {
			loggedIn = true
		}
	}
	status := "not_logged_in"
	if loggedIn {
		status = "logged_in"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"status": status,
	})
}

// ---------- /api/namelist ----------

// resolveNameListPath 定位名单文件：与 Python 下载器 _resolve_list_path 保持一致。
// 相对路径以「下载器脚本目录」为基准（Python 侧 app_dir() 即脚本所在目录）。
func resolveNameListPath() string {
	cfg, err := loadConfig()
	if err != nil {
		return filepath.Join(filepath.Dir(configPath), "mz.txt")
	}
	name := getString(cfg.Raw, []string{"name_list_file"}, "mz.txt")
	if name == "" {
		name = "mz.txt"
	}
	if filepath.IsAbs(name) {
		return name
	}
	// 基准目录：下载器脚本目录；取不到则用 config 目录
	base := ""
	if script, err := downloader.findScript(); err == nil && script != "" {
		base = filepath.Dir(script)
	}
	if base == "" || base == "." {
		base = filepath.Dir(configPath)
	}
	return filepath.Join(base, name)
}

// nameListHandler GET 返回名单文件内容，POST 保存新内容。
// 保存后下载器无需重启：Python 侧按文件 mtime 缓存，内容变化自动重载。
func nameListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := resolveNameListPath()
		data, err := os.ReadFile(path)
		if err != nil {
			// 文件不存在不算错误：返回空内容，前端提示可新建
			if os.IsNotExist(err) {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":      true,
					"path":    path,
					"content": "",
				})
				return
			}
			writeErr(w, http.StatusInternalServerError, "读取名单文件失败: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"path":    path,
			"content": string(data),
		})
	case http.MethodPost:
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
			return
		}
		path := resolveNameListPath()
		// 规范化换行：统一 \n，去掉末尾多余空行
		content := strings.ReplaceAll(req.Content, "\r\n", "\n")
		content = strings.TrimRight(content, "\n")
		if err := os.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
			writeErr(w, http.StatusInternalServerError, "保存名单文件失败: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"message": "名单已保存",
			"path":    path,
		})
	default:
		writeErr(w, http.StatusMethodNotAllowed, "仅支持 GET / POST")
	}
}

// ---------- /api/channels ----------

// channelsHandler 列出账号已加入的频道，供设置面板的「频道选择器」使用。
// 调用 Python 脚本 --list-channels 子模式（连接 TG 拉取对话列表）。
// 注意：session 数据库被下载器进程独占，运行中无法并发读取 → 需先停止监听。
func channelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "请使用 GET")
		return
	}
	if downloader.isRunning() {
		writeErr(w, http.StatusConflict, "下载器正在运行，请先停止监听再加载频道列表")
		return
	}
	py, err := downloader.findPython()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	script, err := downloader.findScript()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	cmd := exec.Command(py, script, "--list-channels")
	cmd.Dir = filepath.Dir(script)
	cmd.Env = append(os.Environ(),
		"TG_STATION_CONFIG="+configPath,
		"PYTHONUTF8=1",
		"PYTHONIOENCODING=utf-8",
	)
	cmd.SysProcAttr = childProcAttr()

	out, err := cmd.Output()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "频道列表获取失败: "+err.Error())
		return
	}

	// 取最后一行合法 JSON（Python 可能输出日志/警告）
	var result map[string]interface{}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if json.Unmarshal([]byte(line), &result) == nil {
			break
		}
	}
	if result == nil {
		writeErr(w, http.StatusInternalServerError, "频道列表输出异常")
		return
	}
	if result["status"] != "ok" {
		msg, _ := result["error"].(string)
		if msg == "" {
			msg = "频道列表获取失败"
		}
		writeErr(w, http.StatusOK, msg)
		return
	}
	channels, _ := result["channels"].([]interface{})
	if channels == nil {
		channels = []interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"channels": channels,
	})
}
