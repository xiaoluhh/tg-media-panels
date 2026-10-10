package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MediaFile struct {
	Path     string
	Date     time.Time
	IsVideo  bool
	Size     int64
	Duration float64 // 视频时长（秒），图片为 0；由 ffprobe/ffmpeg 探测并随索引持久化
}

type LightMediaFile struct {
	Path     string  `json:"path"`
	Date     string  `json:"date"`
	IsVideo  bool    `json:"isVideo"`
	Size     int64   `json:"size"`
	Duration float64 `json:"duration"`
}

var (
	mediaDir     string
	port         int
	indexPath    string
	thumbsDir    string
	mediaIndex   []*MediaFile
	indexMutex   sync.RWMutex
	lastScan     time.Time
	scanInterval = time.Hour
	videoExts    = map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".mkv": true, ".webm": true}
	imageExts    = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	thumbSize    = 400
	ffmpegPath   string
)

func findFFmpeg() string {
	exePath, err := os.Executable()
	if err != nil {
		return "ffmpeg"
	}
	exeDir := filepath.Dir(exePath)

	ffmpegNames := []string{"ffmpeg", "ffmpeg.exe"}

	for _, name := range ffmpegNames {
		path := filepath.Join(exeDir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	systemPaths := []string{"/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg", "/bin/ffmpeg"}
	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "ffmpeg"
}

func selectFolderWindows() string {
	psScript := `
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.Description = "请选择包含图片和视频的文件夹"
$dialog.ShowDialog() | Out-Null
$dialog.SelectedPath
`
	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func main() {
	ffmpegPath = findFFmpeg()
	// ffprobe 优先用于时长探测；找不到时为空，probeDuration 会自动回退 ffmpeg
	ffprobePath = findFFprobe()

	flag.StringVar(&mediaDir, "dir", "", "Media directory (required)")
	flag.IntVar(&port, "port", 8080, "HTTP port (default: 8080)")
	flag.StringVar(&indexPath, "index", "media_index.json", "Index file path")
	flag.StringVar(&thumbsDir, "thumbs", "thumbs", "Thumbnail directory")
	flag.Parse()

	if mediaDir == "" {
		fmt.Println("========================================")
		fmt.Println("      🖼️  媒体浏览器 - Media Browser")
		fmt.Println("========================================")
		fmt.Println()

		if runtime.GOOS == "windows" {
			fmt.Println("💡 正在打开文件夹选择框...")
			selectedFolder := selectFolderWindows()
			if selectedFolder != "" {
				mediaDir = selectedFolder
			}
		}

		if mediaDir == "" {
			fmt.Println("请输入图片/视频文件夹路径：")
			fmt.Print("> ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			mediaDir = strings.TrimSpace(input)
		}

		if mediaDir == "" {
			fmt.Println("错误：未选择文件夹，程序退出。")
			os.Exit(1)
		}
	}

	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		log.Fatal(err)
	}

	loadIndex()

	// 扫描/刷新统一走调度器：同一时刻最多一个扫描在跑，扫描期间的请求
	// 合并为下一轮补扫，缩略图只在「没有待补扫任务」之后基于最新索引生成。
	go func() {
		scanSched.Request()
		for range time.Tick(scanInterval) {
			scanSched.Request()
		}
	}()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/dates", datesHandler)
	http.HandleFunc("/api/files/", filesHandler)
	http.HandleFunc("/api/file/", fileHandler)
	http.HandleFunc("/api/delete", deleteHandler)
	http.HandleFunc("/api/refresh", refreshHandler)
	http.HandleFunc("/thumb/", thumbHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println()
	fmt.Println("✅ 启动成功！")
	fmt.Println("📁 媒体文件夹：", mediaDir)
	fmt.Println("🌐 访问地址：http://localhost:" + strconv.Itoa(port))
	fmt.Println()
	fmt.Println("💡 提示：在浏览器中打开上面的地址即可访问")
	fmt.Println("💡 按 Ctrl+C 可停止程序")
	fmt.Println()

	log.Printf("Server starting on :%d, media dir: %s", port, mediaDir)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")

	html, err := os.ReadFile("static/index.html")
	if err != nil {
		http.ServeFile(w, r, "static/index.html")
		return
	}

	// 注入首页数据，跳过API请求。
	// 日期列表与首个日期的文件都直接取自日期索引（已按倒序维护），
	// 不再遍历整个媒体文件列表。
	indexMutex.RLock()
	dateList := make([]string, len(dateKeys))
	copy(dateList, dateKeys)
	firstDate := ""
	if len(dateList) > 0 {
		firstDate = dateList[0]
	}
	var firstFiles []LightMediaFile
	if firstDate != "" {
		for _, f := range dateIndex[firstDate] {
			firstFiles = append(firstFiles, LightMediaFile{
				Path:     f.Path,
				Date:     firstDate,
				IsVideo:  f.IsVideo,
				Size:     f.Size,
				Duration: f.Duration,
			})
			if len(firstFiles) >= 100 {
				break
			}
		}
	}
	indexMutex.RUnlock()

	initData, _ := json.Marshal(map[string]interface{}{
		"dates":      dateList,
		"firstFiles": firstFiles,
	})
	inject := "<script>window.__INIT__=" + string(initData) + "</script></head>"
	html = bytes.Replace(html, []byte("</head>"), []byte(inject), 1)

	w.Write(html)
}

func datesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	indexMutex.RLock()
	dateList := make([]string, len(dateKeys))
	copy(dateList, dateKeys)
	total := len(mediaIndex)
	indexMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dates": dateList,
		"total": total,
	})
}

// 文件列表分页：每页 100 条（与前端/旧接口约定一致，不能改）。
const filesPageSize = 100

// maxPageNumber 页码上限：越界直接判为非法参数，避免 (page-1)*pageSize 整数溢出。
const maxPageNumber = 10_000_000

// parsePageParam 解析 page 查询参数：
//   - 未传 → 第 1 页
//   - 非整数 / 小于 1 / 超过上限 → 返回错误（调用方回 HTTP 400）
//
// 修复点：旧实现 strconv.Atoi 后忽略错误，也不校验最小值，page=0 或负数会
// 让切片起点变负 → 越界 panic；非数字被当成 0 页。
func parsePageParam(raw string) (int, error) {
	if raw == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("page 参数必须是整数（收到 %q）", clipParam(raw, 32))
	}
	if n < 1 {
		return 0, fmt.Errorf("page 参数必须大于等于 1（收到 %d）", n)
	}
	if n > maxPageNumber {
		return 0, fmt.Errorf("page 参数过大（上限 %d，收到 %d）", maxPageNumber, n)
	}
	return n, nil
}

// clipParam 截断回显的参数，避免超长输入进入响应。
func clipParam(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	date := strings.TrimPrefix(r.URL.Path, "/api/files/")

	page, err := parsePageParam(r.URL.Query().Get("page"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
			"date":  date,
		})
		return
	}

	// 直接命中日期索引（O(该日期的文件数)），不再逐次遍历全部媒体文件。
	// 索引切片在重建时整体替换，取得引用后即使解锁也不会被改写（不可变快照）。
	indexMutex.RLock()
	dayFiles := dateIndex[date]
	indexMutex.RUnlock()

	files := make([]LightMediaFile, 0, len(dayFiles))
	for _, f := range dayFiles {
		files = append(files, LightMediaFile{
			Path:     f.Path,
			Date:     date,
			IsVideo:  f.IsVideo,
			Size:     f.Size,
			Duration: f.Duration,
		})
	}

	pageSize := filesPageSize
	total := len(files)
	pages := (total + pageSize - 1) / pageSize

	// int64 计算边界：页码再大也不会溢出
	start := int64(page-1) * int64(pageSize)
	switch {
	case start >= int64(total):
		// 超出实际页数 → 返回空数组，不越界、不报错
		files = []LightMediaFile{}
	default:
		end := start + int64(pageSize)
		if end > int64(total) {
			end = int64(total)
		}
		files = files[int(start):int(end)]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"total": total,
		"pages": pages,
		"page":  page,
		"date":  date,
	})
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/file/")

	// 统一安全校验：目录边界 + 符号链接/junction 解析（见 pathsafe.go）
	mp, err := safeMediaPath(mediaDir, path)
	if err != nil {
		writeMediaPathError(w, r, "file", path, err)
		return
	}
	if mp.Info.IsDir() {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(mp.Rel))
	if videoExts[ext] {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Accept-Ranges", "bytes")
	} else {
		w.Header().Set("Cache-Control", "max-age=86400")
	}

	// 用解析链接后的真实路径提供服务，避免校验后被替换的链接被跟随
	http.ServeFile(w, r, mp.Real)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Files []string `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	deleted := 0
	failed := 0

	for _, fileName := range req.Files {
		// 删除专用校验：只允许普通文件或指向根内文件的符号链接；
		// 符号链接只删除链接本身（os.Remove 不跟随链接）
		mp, err := safeDeleteTarget(mediaDir, fileName)
		if err != nil {
			if !errors.Is(err, ErrPathMissing) {
				logPathReject("delete", fileName, err)
				failed++
			}
			continue
		}
		srcInfo, statErr := os.Stat(mp.Real)
		if err := os.Remove(mp.Full); err != nil {
			failed++
			log.Printf("[media] 删除失败 %s: %v", mp.Rel, err)
			continue
		}
		deleted++
		if statErr == nil {
			removeThumbCache(mp.Rel, srcInfo.Size(), srcInfo.ModTime())
		} else {
			// 拿不到内容指纹时至少清理旧命名缓存
			removeThumbCache(mp.Rel, 0, time.Time{})
		}
	}

	// 删除后走统一调度器刷新索引（扫描期间的重复请求会被合并，不会丢弃）
	scanSched.Request()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"deleted": deleted,
		"failed":  failed,
	})
}

func thumbHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	path := strings.TrimPrefix(r.URL.Path, "/thumb/")

	// 源文件必须先过统一安全校验（URL 解码在 safeMediaPath 内完成）
	mp, err := safeMediaPath(mediaDir, path)
	if err != nil {
		writeMediaPathError(w, r, "thumbnail", path, err)
		return
	}
	if mp.Info.IsDir() {
		http.NotFound(w, r)
		return
	}

	// 用「解析链接后的真实文件」的大小与修改时间构造缓存键
	srcInfo, err := os.Stat(mp.Real)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	thumbPath, err := ensureThumb(mp, srcInfo)
	if err != nil {
		log.Printf("[media] 缩略图不可用 %s: %v", mp.Rel, err)
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, thumbPath)
}

func resizeImage(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxSize && height <= maxSize {
		return img
	}

	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = (height * maxSize) / width
	} else {
		newHeight = maxSize
		newWidth = (width * maxSize) / height
	}

	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	xRatio := float64(width) / float64(newWidth)
	yRatio := float64(height) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := float64(x)*xRatio + 0.5
			srcY := float64(y)*yRatio + 0.5

			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1
			if x1 >= width {
				x1 = width - 1
			}
			if y1 >= height {
				y1 = height - 1
			}

			xFrac := srcX - float64(x0)
			yFrac := srcY - float64(y0)

			c00 := img.At(x0, y0)
			c01 := img.At(x0, y1)
			c10 := img.At(x1, y0)
			c11 := img.At(x1, y1)

			r00, g00, b00, a00 := c00.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			r := bilinearInterp(r00, r01, r10, r11, xFrac, yFrac)
			g := bilinearInterp(g00, g01, g10, g11, xFrac, yFrac)
			b := bilinearInterp(b00, b01, b10, b11, xFrac, yFrac)
			a := bilinearInterp(a00, a01, a10, a11, xFrac, yFrac)

			result.Set(x, y, color.RGBA64{
				R: uint16(r),
				G: uint16(g),
				B: uint16(b),
				A: uint16(a),
			})
		}
	}

	return result
}

func bilinearInterp(c00, c01, c10, c11 uint32, xFrac, yFrac float64) float64 {
	top := float64(c00)*(1-xFrac) + float64(c10)*xFrac
	bot := float64(c01)*(1-xFrac) + float64(c11)*xFrac
	return top*(1-yFrac) + bot*yFrac
}

// refreshHandler 请求重新扫描媒体目录并重建索引（第二阶段 · 第二部分）。
// 与 tg-station 的 /api/refresh 行为一致：异步返回「正在刷新」，由调用方通过
// /api/status 或 /api/dates 观察结果；扫描期间的重复请求合并为下一轮补扫，
// 既不丢弃请求，也绝不提前返回虚假的「已完成」。
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	started := scanSched.Request()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": "正在刷新媒体索引",
		"started": started,
		"merged":  !started,
		"running": scanSched.Running(),
		"pending": scanSched.Pending(),
	})
}
