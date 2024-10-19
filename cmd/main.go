//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"log"

	"github.com/duckfeather10086/dandan-prime/config"
	"github.com/duckfeather10086/dandan-prime/controllers"
	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	//Initialize database
	if err := database.InitDatabase("media_library.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := config.InitConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
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

	e := echo.New()

	// w := ffmpegutil.NewWorker()
	// w.CurrentEpisodeID = 0
	// go w.Start()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	// 配置静态文件服务
	e.Static("/videos", config.MEDIA_LIBRARY_ROOT_PATH)
	e.Static("/subtitles", config.MEDIA_LIBRARY_ROOT_PATH)

	// 配置视频流媒体服务
	e.GET("/stream/:filename", controllers.ServeHLSPlayListHandler)

	e.GET("/segment", controllers.ServeHLSSegmentHandler)

	e.GET("/api/bangumi/:bangumi_subject_id/contents", controllers.GetBangumiContentsByBangumiID)

	e.GET("/api/bangumi/list", controllers.GetBangumiInfoList)

	e.GET("/api/bangumi/episode/:id", controllers.GetEpisodeInfoByID)

	e.GET("/api/bangumi/danmaku/:episode_id", controllers.GetDanmakuByDandanplayEpisodeID)

	e.POST("/api/playlist/:id", controllers.InitPlayListHandler)

	e.POST("/api/hls", controllers.SetHlsEnable)

	e.GET("/api/hls/enabled", controllers.GetHlsEnable)

	e.GET("/api/last_watched", controllers.GetLastWatchedInfo)
	e.POST("/api/last_watched", controllers.UpdateLastedWatched)

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
