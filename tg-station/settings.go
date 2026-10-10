package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// StationConfig 是整合后 Go 端新增的配置段（放在 config.json 的 "station" 键下）。
// Python 下载器原有的所有字段（api_id、api_hash、channels、download_dir 等）
// 原样保留在 config.json 顶层，Go 端不解析、不改动，只透传。
type StationConfig struct {
	Python        string `json:"python"`         // Python 解释器路径（旧模式），空 = 自动探测
	Script        string `json:"script"`         // tg_auto_download.py 路径（旧模式），空 = exe 目录下
	DownloaderExe string `json:"downloader_exe"` // 打包后的下载器 exe（新模式），空 = exe 目录/tg_downloader.exe
	MediaDir      string `json:"media_dir"`      // 媒体目录，空 = config 的 download_dir
	AutoStart     bool   `json:"auto_start"`     // 程序启动时自动拉起下载器
}

// AppConfig 是 Go 端使用的完整配置视图。
type AppConfig struct {
	// Raw 是 config.json 的完整内容（含 Python 全部字段），保存时原样写回。
	Raw map[string]interface{}

	Station StationConfig `json:"station"`

	// 以下字段从 Raw 中便捷读取（媒体浏览服务参数）
	Port      int    // media_browser.port
	MediaDir  string // download_dir（station.media_dir 优先）
	IndexPath string // 索引文件路径
	ThumbsDir string // 缩略图缓存目录
}

var (
	configPath string // config.json 的绝对路径
	cfgMutex   sync.Mutex
)

// defaultConfigPath 返回默认配置文件路径：exe 所在目录/config.json
func defaultConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

// stationDefaults 返回 station 段的默认值
func stationDefaults() StationConfig {
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	return StationConfig{
		Python:    "",
		Script:    filepath.Join(exeDir, "tg_auto_download.py"),
		MediaDir:  "",
		AutoStart: false,
	}
}

// loadConfig 读取配置文件；不存在则生成一份默认配置并返回。
// 任何错误都返回 nil + error，由调用方决定是否 fatal。
func loadConfig() (*AppConfig, error) {
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	app := &AppConfig{Raw: map[string]interface{}{}}
	app.Station = stationDefaults()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 生成默认配置
			app.Raw = buildDefaultRaw(app.Station)
			if err := saveConfigLocked(app); err != nil {
				return nil, fmt.Errorf("生成默认配置文件失败: %w", err)
			}
			return app, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := json.Unmarshal(data, &app.Raw); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	applyRaw(app)
	return app, nil
}

// buildDefaultRaw 构造与 Python 下载器兼容的默认配置内容
func buildDefaultRaw(st StationConfig) map[string]interface{} {
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	return map[string]interface{}{
		"api_id":            0,
		"api_hash":          "",
		"session":           filepath.Join(exeDir, "tg_session"),
		"download_dir":      filepath.Join(exeDir, "downloads"),
		"channels":          []interface{}{},
		"types":             []interface{}{"photo", "video", "document"},
		"include_forwarded": true,
		"skip_gif":          true,
		"download_history":  false,
		"catchup_today":     true,
		"movie_min_minutes": 5,
		"only_forwarded":    true,
		"max_file_size_mb":  8192,
		"retries":           3,
		"concurrency":       4,
		"proxy":             nil,
		"name_list_file":    "mz.txt",
		"login_method":      "qr",
		"twofa_password":    "",
		"media_browser": map[string]interface{}{
			"exe":        filepath.Join(exeDir, "media-browser-windows-x64.exe"),
			"workdir":    exeDir,
			"port":       8080,
			"index_path": filepath.Join(exeDir, "media_index.json"),
		},
		"station": st,
	}
}

// applyRaw 把 Raw 里的字段解析到 AppConfig 便捷字段
func applyRaw(app *AppConfig) {
	// station 段
	if v, ok := app.Raw["station"].(map[string]interface{}); ok {
		if s, err := json.Marshal(v); err == nil {
			var st StationConfig
			if json.Unmarshal(s, &st) == nil {
				app.Station = st
			}
		}
	}
	// 便捷字段
	app.Port = getInt(app.Raw, []string{"media_browser", "port"}, 8080)
	app.MediaDir = getString(app.Raw, []string{"download_dir"}, "")
	if app.Station.MediaDir != "" {
		app.MediaDir = app.Station.MediaDir
	}
	// 索引/缩略图统一放在程序（exe）目录；media_browser.index_path 仅供 Python 旧逻辑使用
	exeDir := "."
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	app.IndexPath = filepath.Join(exeDir, "media_index.json")
	app.ThumbsDir = filepath.Join(exeDir, "thumbs")
}

// getString 从嵌套 map 中取字符串值
func getString(m map[string]interface{}, keys []string, def string) string {
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return def
		}
		if i == len(keys)-1 {
			if s, ok := v.(string); ok {
				return s
			}
			return def
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return def
		}
		cur = next
	}
	return def
}

// getInt 从嵌套 map 中取整数值（json 数字默认 float64）
func getInt(m map[string]interface{}, keys []string, def int) int {
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return def
		}
		if i == len(keys)-1 {
			switch n := v.(type) {
			case float64:
				return int(n)
			case int:
				return n
			case string:
				var r int
				if _, err := fmt.Sscanf(n, "%d", &r); err == nil {
					return r
				}
			}
			return def
		}
		next, ok := v.(map[string]interface{})
		if !ok {
			return def
		}
		cur = next
	}
	return def
}

// saveConfig 把 AppConfig 写回配置文件（对外入口，加锁）
func saveConfig(app *AppConfig) error {
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	return saveConfigLocked(app)
}

// saveConfigLocked 写回配置（调用方需持有 cfgMutex）
func saveConfigLocked(app *AppConfig) error {
	app.Raw["station"] = app.Station
	data, err := json.MarshalIndent(app.Raw, "", "  ")
	if err != nil {
		return err
	}
	tmp := configPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, configPath)
}
