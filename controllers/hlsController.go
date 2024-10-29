//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/duckfeather10086/dandan-prime/internal/ffmpegutil"
	episodeusecase "github.com/duckfeather10086/dandan-prime/usecase/episodeUsecase"
	"github.com/labstack/echo/v4"
)

type HlsSegmentInfo struct {
	SegmentIndex    int    `json:"segment_index"`
	SegmentDuration int    `json:"segment_duration"`
	PixelFormat     string `json:"pixel_format"`
	Size            int    `json:"size"`
}

func ServeHLSPlayListHandler(c echo.Context) error {
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

func ServeHLSSegmentHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.QueryParam("id"))
	log.Print("id", id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid episode ID"})
	}

	startIndex, err := strconv.Atoi(c.QueryParam("index"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid start index")
	}

	resolution := 1080
	if c.QueryParam("resolution") != "" {
		resolution, err = strconv.Atoi(c.QueryParam("resolution"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid resolution")
		}
	}

	ratio := "16:9"
	if c.QueryParam("ratio") != "" {
		ratio = c.QueryParam("ratio")
	}

	c.Response().Header().Set("Content-Type", "video/MP2T")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=segment_%d.ts", startIndex))

	c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	c.Response().Header().Set("Accept-Ranges", "bytes")

	episodeInfo, err := episodeusecase.GetEpisodeInfoById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	segmentationDuration := 5
	err = ffmpegutil.GenerateHlsSegment(episodeInfo.FilePath+"/"+episodeInfo.FileName, startIndex, resolution, segmentationDuration, ratio, c.Response().Writer)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating HLS segment: %v", err))
	}

	return nil
}

func InitPlayListHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	log.Print("id", id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid episode ID"})
	}

	episodeInfo, err := episodeusecase.GetEpisodeInfoById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	log.Println("FileName", episodeInfo.FilePath+episodeInfo.FileName)
	err = ffmpegutil.GenerateHlsPlayList(episodeInfo.FilePath+"/"+episodeInfo.FileName, id, 5)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate episode playlist. err: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"playlist_path": "cache" + "playlist.m3u8",
	})
}

func GenerateHlsSegments(c echo.Context, worker *ffmpegutil.Worker) error {
	type RequestBody struct {
		ID         string `json:"id"`
		StartIndex int    `json:"start_index"`
	}
	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read request body"})
	}

	log.Println("body", string(body))
	// Unmarshal the JSON data
	var reqBody RequestBody
	if err := json.Unmarshal(body, &reqBody); err != nil {
		log.Println("error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON in request body"})
	}

	log.Printf("id: %s, start_index: %d", reqBody.ID, reqBody.StartIndex)

	episodeIdInt, err := strconv.Atoi(reqBody.ID)
	if err != nil {
		log.Println("Error converting episode ID to int")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to convert episode ID to int"})
	}

	episodeInfo, err := episodeusecase.GetEpisodeInfoById(episodeIdInt)
	if err != nil {
		log.Println("Error getting episode info from", reqBody.ID)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	params := ffmpegutil.HlsSegmentsParams{
		InputFile:       episodeInfo.FilePath + "/" + episodeInfo.FileName,
		StartIndex:      reqBody.StartIndex,
		TotalSegments:   5,
		SegmentDuration: 10,
	}

	worker.Process(params)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}
