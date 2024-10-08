//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/duckfeather10086/dandan-prime/config"
	"github.com/duckfeather10086/dandan-prime/controllers"
	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type AnimeData struct {
	AnimeTitle string `json:"animeTitle"`
	ImageURL   string `json:"imageUrl"`
}

type APIResponse struct {
	BangumiList []AnimeData `json:"bangumiList"`
}

func fetchAnimeData() ([]AnimeData, error) {
	resp, err := http.Get("https://api.dandanplay.net/api/v2/bangumi/shin")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, err
	}

	return apiResp.BangumiList, nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	animeData, err := fetchAnimeData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(animeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func main() {
	//Initialize database
	if err := database.InitDatabase("media_library.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// if err := episodeusecase.ScanAndSaveMedia(config.DefaultMediaLibraryPath); err != nil {
	// 	log.Printf("Error scanning and matching media: %v", err)
	// }

	// if err := episodeusecase.ScanAndMatchMedia(config.DefaultMediaLibraryPath); err != nil {
	// 	log.Printf("Error scanning and matching media: %v", err)
	// }

	// err := bangumiusecase.InitializeBangumiInfo()
	// if err != nil {
	// 	log.Printf("Error scanning and matching media: %v", err)
	// }

	// err = episodeusecase.ScanAndMatchSubtitles()
	// if err != nil {
	// 	log.Printf("Error scanning and matching media: %v", err)
	// }

	// err := episodeusecase.ScanAndPrepareThumbnails()
	// if err != nil {
	// 	log.Printf("Error scanning and matching media: %v", err)
	// }

	//filesacnner.ScanAndSaveMedia(mediaLibraryPath)

	// filePath := "O:\\dandan-backend\\[Airota&Nekomoe kissaten&VCB-Studio] Yuru Camp Season 2 [01][Ma10p_1080p][x265_flac].mkv"
	// info, err := getVideoInfo(filePath)
	// if err != nil {
	// 	fmt.Printf("Error: %v\n", err)
	// 	return
	// }

	// jsonData, err := json.MarshalIndent(info, "", "  ")
	// if err != nil {
	// 	fmt.Printf("Error marshaling JSON: %v\n", err)
	// 	return
	// }

	// fmt.Println(string(jsonData))

	// mux := http.NewServeMux()
	// mux.HandleFunc("/api/anime", handleRequest)

	// handler := cors.Default().Handler(mux)

	// fmt.Println("Server is running on http://localhost:8080")
	// log.Fatal(http.ListenAndServe(":8080", handler))

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	// 配置静态文件服务
	e.Static("/videos", config.DefaultMediaLibraryPath)
	e.Static("/subtitles", config.DefaultMediaLibraryPath)

	// 配置视频流媒体服务
	e.GET("/stream/:filename", serveHLSHandler)

	e.GET("/api/bangumi/:bangumi_subject_id/contents", controllers.GetBangumiContentsByBangumiID)

	e.GET("/api/bangumi/list", controllers.GetBangumiInfoList)

	e.GET("/api/bangumi/episode/:id", controllers.GetEpisodeInfoByID)

	e.GET("/api/bangumi/danmaku/:episode_id", controllers.GetDanmakuByDandanplayEpisodeID)

	// 启动服务器
	if err := e.Start(":1234"); err != nil {
		log.Fatal(err)
	}

}

type VideoInfo struct {
	FileName      string  `json:"fileName"`
	FileHash      string  `json:"fileHash"`
	FileSize      int64   `json:"fileSize"`
	VideoDuration float64 `json:"videoDuration"`
	MatchMode     string  `json:"matchMode"`
}

func getVideoInfo(filePath string) (VideoInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return VideoInfo{}, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return VideoInfo{}, err
	}

	hash, err := getFileHash(file)
	if err != nil {
		return VideoInfo{}, err
	}

	//1c7ac8df48699785872bea85a9e82c60

	return VideoInfo{
		FileName:      filepath.Base(filePath),
		FileHash:      hash,
		FileSize:      fileInfo.Size(),
		VideoDuration: 24,
		MatchMode:     "hashAndFileName",
	}, nil
}

func getFileHash(file *os.File) (string, error) {
	hash := md5.New()
	chunkSize := 16 * 1024 * 1024 // 16MB
	buf := make([]byte, chunkSize)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	hash.Write(buf[:n])
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func streamHandler(c echo.Context) error {
	filename := c.Param("filename")
	videoPath := filepath.Join("./videos", filename)

	// 检查文件是否存在
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Video not found"})
	}

	// 设置响应头
	c.Response().Header().Set(echo.HeaderContentType, "video/mp4")
	c.Response().Header().Set(echo.HeaderContentDisposition, "inline; filename="+filename)
	return c.File(videoPath)
}
func serveHLSHandler(c echo.Context) error {
	filename := c.Param("filename")
	fileExt := filepath.Ext(filename)
	filePath := filepath.Join("./cache", filename)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "File not found"})
	}

	var contentType string
	switch fileExt {
	case ".m3u8":
		contentType = "application/vnd.apple.mpegurl"
	case ".ts":
		contentType = "video/mp2t"
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid file type"})
	}

	// Set appropriate headers
	c.Response().Header().Set(echo.HeaderContentType, contentType)
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Cache-Control", "no-cache")

	return c.File(filePath)
}
