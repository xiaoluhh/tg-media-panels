package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type MediaFile struct {
	Path     string
	Date     time.Time
	IsVideo  bool
	Size     int64
	Duration float64 // 视频时长（秒），图片为 0；由 ffmpeg 探测并随索引持久化
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
	staticDir    string
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

	// 候选目录：exe 目录 > 当前目录 > config 目录 > 下载器脚本目录
	candidates := []string{exeDir}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if configPath != "" {
		candidates = append(candidates, filepath.Dir(configPath))
	}
	if script, err := downloader.findScript(); err == nil && script != "" {
		candidates = append(candidates, filepath.Dir(script))
	}

	for _, dir := range candidates {
		for _, name := range ffmpegNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
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
	var configFlag string
	var showConsole bool
	flag.StringVar(&configFlag, "config", "", "配置文件路径（默认: 程序目录/config.json）")
	flag.BoolVar(&showConsole, "console", false, "保留控制台窗口（默认隐藏，后台运行）")
	flag.StringVar(&mediaDir, "dir", "", "媒体目录（默认取配置 download_dir）")
	flag.IntVar(&port, "port", 8080, "HTTP 端口（默认 8080）")
	flag.StringVar(&indexPath, "index", "", "索引文件路径（默认: 程序目录/media_index.json）")
	flag.StringVar(&thumbsDir, "thumbs", "", "缩略图缓存目录（默认: 程序目录/thumbs）")
	flag.Parse()

	// Windows 下默认隐藏控制台窗口（后台运行）；-console 可保留
	if runtime.GOOS == "windows" && !showConsole {
		hideConsoleWindow()
	}

	// 加载配置：-config 可指定路径，默认 exe 目录/config.json
	configPath = configFlag
	if configPath == "" {
		configPath = defaultConfigPath()
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("[station] 配置加载失败，使用默认值: %v", err)
		cfg = &AppConfig{Raw: buildDefaultRaw(stationDefaults())}
		applyRaw(cfg)
	}
	downloader.cfg = cfg

	// ffmpeg 定位依赖 configPath/downloader（多目录搜索），须在配置加载后调用
	ffmpegPath = findFFmpeg()

	// flag 未显式设置时，用配置值覆盖默认
	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	if !setFlags["port"] && cfg.Port != 0 {
		port = cfg.Port
	}
	if !setFlags["dir"] && cfg.MediaDir != "" {
		mediaDir = cfg.MediaDir
	}
	if !setFlags["index"] {
		indexPath = cfg.IndexPath
	}
	if !setFlags["thumbs"] {
		exeDir := "."
		if exe, err := os.Executable(); err == nil {
			exeDir = filepath.Dir(exe)
		}
		thumbsDir = filepath.Join(exeDir, "thumbs")
	}

	// 前端资源目录：exe 目录/static 优先，其次进程工作目录/static。
	// 注意 filepath.Join(dir, "static") 会得到绝对路径，原先写 staticDir="static"
	// （相对路径）导致兜底分支实际依赖进程工作目录，"exe 目录优先"形同虚设——
	// 换个工作目录启动就会静默服务到另一个 index.html。
	staticDir = ""
	if exePath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "static")
		if _, statErr := os.Stat(candidate); statErr == nil {
			staticDir = candidate
		}
	}
	if staticDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			candidate := filepath.Join(cwd, "static")
			if _, statErr := os.Stat(candidate); statErr == nil {
				staticDir = candidate
			}
		}
	}
	if staticDir == "" {
		staticDir = "static"
	}
	log.Printf("[station] 前端资源目录: %s", staticDir)

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

	go func() {
		scanFiles()
		go generateAllThumbnails()
		for range time.Tick(scanInterval) {
			scanFiles()
			go generateAllThumbnails()
		}
	}()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/dates", datesHandler)
	http.HandleFunc("/api/files/", filesHandler)
	http.HandleFunc("/api/file/", fileHandler)
	http.HandleFunc("/api/delete", deleteHandler)
	http.HandleFunc("/thumb/", thumbHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	// 整合功能：下载器管理 / 设置 / 刷新（必须在 ListenAndServe 前挂载）
	registerAPIRoutes()

	// 配置开启自动启动时，拉起下载器
	if cfg.Station.AutoStart {
		go func() {
			if err := downloader.Start(); err != nil {
				log.Printf("[station] 自动启动下载器失败: %v", err)
			}
		}()
	}

	// 代理由用户常开的 Clash Verge（127.0.0.1:7897）负责，内置 mihomo 自启动已于 2026-10-09 移除。

	fmt.Println()
	fmt.Println("=============================================")
	fmt.Println("      🎛️  TG Station - 下载 + 媒体 一体机")
	fmt.Println("=============================================")
	fmt.Println()
	fmt.Println("✅ 启动成功！")
	fmt.Println("📁 媒体文件夹：", mediaDir)
	fmt.Println("🌐 访问地址：http://localhost:" + strconv.Itoa(port))
	fmt.Println("🎧 下载器：/api/listen/start 启动 | /api/listen/stop 停止")
	fmt.Println("⚙️  设置面板：页面右上角 ⚙️ 按钮（或 /api/settings）")
	fmt.Println()

	log.Printf("Server starting on :%d, media dir: %s", port, mediaDir)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 路由 "/" 是 Go 的兜底 pattern：未注册的路径也会落到这里。
	// 不注册的 /api/* 与 /thumb/* 必须返回 404，否则接口写错/已下线会被
	// 伪装成"200 + 首页 HTML"，前端 stationFetch 会把它当成成功响应。
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/thumb/") {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")

	html, err := os.ReadFile(filepath.Join(staticDir, "index.html"))
	if err != nil {
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		return
	}

	// 注入首页数据，跳过API请求
	indexMutex.RLock()
	firstDate := ""
	var firstFiles []LightMediaFile
	dates := make(map[string]bool)
	for _, f := range mediaIndex {
		d := f.Date.Format("2006-01-02")
		dates[d] = true
		if firstDate == "" || d > firstDate {
			firstDate = d
		}
	}
	if firstDate != "" {
		for _, f := range mediaIndex {
			if f.Date.Format("2006-01-02") == firstDate {
				firstFiles = append(firstFiles, LightMediaFile{
					Path:    f.Path,
					Date:    f.Date.Format("2006-01-02"),
					IsVideo: f.IsVideo,
					Size:    f.Size,
					Duration: f.Duration,
				})
				if len(firstFiles) >= 100 {
					break
				}
			}
		}
	}
	dateList := make([]string, 0, len(dates))
	for d := range dates {
		dateList = append(dateList, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dateList)))
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
	defer indexMutex.RUnlock()

	dates := make(map[string]bool)
	for _, f := range mediaIndex {
		dates[f.Date.Format("2006-01-02")] = true
	}

	dateList := make([]string, 0, len(dates))
	for d := range dates {
		dateList = append(dateList, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dateList)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dates": dateList,
		"total": len(mediaIndex),
	})
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	date := strings.TrimPrefix(r.URL.Path, "/api/files/")
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	indexMutex.RLock()
	defer indexMutex.RUnlock()

	var files []LightMediaFile
	targetDate := date
	for _, f := range mediaIndex {
		if f.Date.Format("2006-01-02") == targetDate {
			files = append(files, LightMediaFile{
				Path:    f.Path,
				Date:    f.Date.Format("2006-01-02"),
				IsVideo: f.IsVideo,
				Size:    f.Size,
				Duration: f.Duration,
			})
		}
	}

	pageSize := 100
	total := len(files)
	pages := (total + pageSize - 1) / pageSize
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		files = []LightMediaFile{}
	} else if end > total {
		files = files[start:]
	} else {
		files = files[start:end]
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
	fullPath := filepath.Join(mediaDir, filepath.FromSlash(path))
	absMedia, _ := filepath.Abs(mediaDir)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absMedia) {
		http.NotFound(w, r)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	if videoExts[ext] {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Accept-Ranges", "bytes")
	} else {
		w.Header().Set("Cache-Control", "max-age=86400")
	}

	http.ServeFile(w, r, fullPath)
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
		filePath := filepath.Join(mediaDir, filepath.FromSlash(fileName))
		safeName := strings.ReplaceAll(fileName, "/", "_")
		thumbPath := filepath.Join(thumbsDir, safeName+".thumb.jpg")

		absMedia, _ := filepath.Abs(mediaDir)
		absFull, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFull, absMedia) {
			continue
		}

		if _, err := os.Stat(filePath); err == nil {
			if err := os.Remove(filePath); err == nil {
				deleted++
				os.Remove(thumbPath)
			} else {
				failed++
				log.Printf("Failed to delete %s: %v", filePath, err)
			}
		}
	}

	go scanFiles()

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
	// URL解码：前端用encodeURIComponent，这里解码回原始中文路径
	if decoded, err := url.PathUnescape(path); err == nil {
		path = decoded
	}
	safeName := strings.ReplaceAll(path, "/", "_")
	thumbPath := filepath.Join(thumbsDir, safeName+".thumb.jpg")
	fullPath := filepath.Join(mediaDir, filepath.FromSlash(path))

	absMedia, _ := filepath.Abs(mediaDir)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absMedia) {
		http.NotFound(w, r)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	needRegenerate := false
	thumbInfo, err1 := os.Stat(thumbPath)
	origInfo, err2 := os.Stat(fullPath)

	if err1 != nil || thumbInfo == nil {
		needRegenerate = true
	} else if err2 == nil && origInfo != nil && origInfo.ModTime().After(thumbInfo.ModTime()) {
		needRegenerate = true
	}

	if !needRegenerate {
		http.ServeFile(w, r, thumbPath)
		return
	}

	ext := strings.ToLower(filepath.Ext(path))

	var img image.Image
	if videoExts[ext] {
		img = generateVideoThumbnail(fullPath)
	} else {
		img = generateImageThumbnail(fullPath)
	}

	if img == nil {
		http.NotFound(w, r)
		return
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	os.WriteFile(thumbPath, buf.Bytes(), 0644)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(buf.Bytes())
}

func generateImageThumbnail(imagePath string) image.Image {
	cmd := exec.Command(ffmpegPath,
		"-i", imagePath,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale='if(gt(iw,ih),%d,-2)':'if(gt(iw,ih),-2,%d)':flags=lanczos", thumbSize, thumbSize),
		"-q:v", "2",
		"-f", "image2pipe",
		"-")
	cmd.SysProcAttr = childProcAttr()
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err == nil {
		img, _, err := image.Decode(&out)
		if err == nil {
			return img
		}
	}

	file, err := os.Open(imagePath)
	if err != nil {
		log.Printf("Failed to open image: %v", err)
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Failed to decode image: %v, path: %s", err, imagePath)
		return nil
	}

	return resizeImage(img, thumbSize)
}

func generateVideoThumbnail(videoPath string) image.Image {
	timePoints := []string{"00:00:01", "00:00:03", "00:00:05", "00:00:10"}

	for _, timePoint := range timePoints {
		cmd := exec.Command(ffmpegPath,
			"-ss", timePoint,
			"-i", videoPath,
			"-vframes", "1",
			"-vf", fmt.Sprintf("scale='if(gt(iw,ih),%d,-2)':'if(gt(iw,ih),-2,%d)':flags=lanczos", thumbSize, thumbSize),
			"-q:v", "2",
			"-f", "image2pipe",
			"-")
		cmd.SysProcAttr = childProcAttr()
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			img, _, err := image.Decode(&out)
			if err == nil {
				return img
			}
		}
	}

	log.Printf("Failed to generate thumbnail for video: %s", videoPath)
	return nil
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

var thumbGenRunning int32

func generateAllThumbnails() {
	// 防止重复触发导致 ffmpeg 进程堆积：下载器频繁调用 /api/refresh，
	// 若上一次生成尚未完成，直接跳过本次触发。
	if !atomic.CompareAndSwapInt32(&thumbGenRunning, 0, 1) {
		log.Printf("缩略图生成已在进行中，跳过本次触发")
		return
	}
	defer atomic.StoreInt32(&thumbGenRunning, 0)

	indexMutex.RLock()
	files := make([]*MediaFile, len(mediaIndex))
	copy(files, mediaIndex)
	indexMutex.RUnlock()

	workerCount := 2
	jobs := make(chan string, len(files))
	results := make(chan bool, len(files))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				safeName := strings.ReplaceAll(path, "/", "_")
				thumbPath := filepath.Join(thumbsDir, safeName+".thumb.jpg")
				fullPath := filepath.Join(mediaDir, filepath.FromSlash(path))

				needGenerate := false
				thumbInfo, err1 := os.Stat(thumbPath)
				origInfo, err2 := os.Stat(fullPath)

				if err2 != nil || origInfo == nil {
					results <- false
					continue
				}

				if err1 != nil || thumbInfo == nil {
					needGenerate = true
				} else if origInfo.ModTime().After(thumbInfo.ModTime()) {
					needGenerate = true
				}

				if !needGenerate {
					results <- false
					continue
				}

				ext := strings.ToLower(filepath.Ext(path))

				var img image.Image
				if videoExts[ext] {
					img = generateVideoThumbnail(fullPath)
				} else {
					img = generateImageThumbnail(fullPath)
				}

				if img != nil {
					f, err := os.Create(thumbPath)
					if err == nil {
						jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
						f.Close()
						results <- true
					} else {
						results <- false
					}
				} else {
					results <- false
				}
			}
		}()
	}

	for _, f := range files {
		jobs <- f.Path
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	generated := 0
	skipped := 0
	for result := range results {
		if result {
			generated++
		} else {
			skipped++
		}
	}

	log.Printf("缩略图预生成完成：新生成 %d 个，跳过 %d 个", generated, skipped)
}

var durationRe = regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+(?:\.\d+)?)`)

// probeDuration 用 ffmpeg -i 读取视频时长（秒）。失败返回 0。
func probeDuration(path string) float64 {
	cmd := exec.Command(ffmpegPath, "-i", path)
	cmd.SysProcAttr = childProcAttr()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_ = cmd.Run() // -i 不指定输出会以非零码退出，属预期；只需解析 stderr
	m := durationRe.FindStringSubmatch(stderr.String())
	if m == nil {
		return 0
	}
	h, _ := strconv.ParseFloat(m[1], 64)
	mi, _ := strconv.ParseFloat(m[2], 64)
	s, _ := strconv.ParseFloat(m[3], 64)
	return h*3600 + mi*60 + s
}

// fillDurations 并发探测视频时长，补齐 files 中缺失的 Duration（秒）。
func fillDurations(files []*MediaFile) {
	var pending []*MediaFile
	for _, f := range files {
		if f.IsVideo && f.Duration <= 0 {
			pending = append(pending, f)
		}
	}
	if len(pending) == 0 {
		return
	}
	workerCount := 4
	if len(pending) < workerCount {
		workerCount = len(pending)
	}
	jobs := make(chan *MediaFile, len(pending))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				f.Duration = probeDuration(filepath.Join(mediaDir, filepath.FromSlash(f.Path)))
			}
		}()
	}
	for _, f := range pending {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
}

// scanFiles 的重入保护。
// scanFiles 内部会调用 fillDurations（对每个缺失时长的视频跑一次 ffmpeg -i），
// 首次扫描 5700+ 个视频要十几分钟。这段时间里任何再次触发的扫描（每小时 tick、
// 前端 F5 的 POST /api/refresh、删除后的重扫）都会与它叠加：实测出现 67 个并发
// ffmpeg、磁盘被打饱和、/api/status 从 250ms 退化到 4.5 秒、网页卡死。
// 因此同一时刻只允许一个扫描在跑，后来的直接跳过（索引本来就会在下一轮补上）。
var scanRunning int32

func scanFiles() {
	if !atomic.CompareAndSwapInt32(&scanRunning, 0, 1) {
		log.Printf("[station] 扫描已在进行中，跳过本次触发")
		return
	}
	defer atomic.StoreInt32(&scanRunning, 0)

	scanFilesLocked()
}

func scanFilesLocked() {
	// 复用上次索引中同路径同大小的视频时长，避免每次扫描都重新 ffmpeg 探测
	indexMutex.RLock()
	prevByPath := make(map[string]*MediaFile, len(mediaIndex))
	for _, f := range mediaIndex {
		prevByPath[f.Path] = f
	}
	indexMutex.RUnlock()

	var files []*MediaFile
	err := filepath.WalkDir(mediaDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == thumbsDir {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !imageExts[ext] && !videoExts[ext] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(mediaDir, path)
		mf := &MediaFile{
			Path:    filepath.ToSlash(relPath),
			Date:    info.ModTime(),
			IsVideo: videoExts[ext],
			Size:    info.Size(),
		}
		if mf.IsVideo {
			if p, ok := prevByPath[mf.Path]; ok && p.Size == info.Size() && p.Duration > 0 {
				mf.Duration = p.Duration
			}
		}
		files = append(files, mf)
		return nil
	})

	if err != nil {
		log.Printf("Scan error: %v", err)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Date.After(files[j].Date)
	})

	// 补齐缺失的视频时长（新文件或大小变化后重新探测）
	fillDurations(files)

	indexMutex.Lock()
	mediaIndex = files
	lastScan = time.Now()
	indexMutex.Unlock()

	saveIndex()
	log.Printf("Scanned %d files", len(files))
}

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

	indexMutex.Lock()
	mediaIndex = files
	indexMutex.Unlock()

	log.Printf("Loaded index with %d files", len(files))
}

func saveIndex() {
	indexMutex.RLock()
	data, err := json.Marshal(mediaIndex)
	indexMutex.RUnlock()

	if err != nil {
		log.Printf("Failed to marshal index: %v", err)
		return
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		log.Printf("Failed to save index: %v", err)
	}
}
