//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/duckfeather10086/dandan-prime/config"
	bangumiusecase "github.com/duckfeather10086/dandan-prime/usecase/bangumiUsecase"
	episodeusecase "github.com/duckfeather10086/dandan-prime/usecase/episodeUsecase"
	"github.com/labstack/echo/v4"
)

type BangumiInfo struct {
	ID                  int       `json:"id"`
	BangumiSubjectID    int       `json:"bangumi_subject_id"`
	DandanplayBangumiID int       `json:"dandanpaly_bangumi_id"`
	ImageURL            string    `json:"image_url"`
	Summary             string    `json:"summary"`
	RateScore           float64   `json:"rate_score"`
	TotalEpisodes       int       `json:"total_episodes"`
	AirDate             string    `json:"air_date"`
	Platform            string    `json:"platform"`
	Title               string    `json:"title"`
	Directory           string    `json:"directory"`
	Episodes            []Episode `json:"episodes"`
}

// Episode 结构体用于存储作品目录下的内容信息
type Episode struct {
	ID                  uint     `json:"id"`
	DandanplayEpisodeID int      `json:"dandanpaly_episode_id"`
	Title               string   `json:"title"`
	Type                string   `json:"type"`         // 例如: "video", "image", "subtitle" 等
	Introduction        string   `json:"introduction"` //
	FileName            string   `json:"file_name"`
	Subtitles           []string `json:"subtitles"`
	FilePath            string   `json:"file_path"`
}

func GetBangumiContentsByBangumiID(c echo.Context) error {
	bangumiSubjectIDStr := c.Param("bangumi_subject_id")

	bangumiSubjectID, err := strconv.Atoi(bangumiSubjectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bangumi subject ID"})
	}

	episodeInfos, err := episodeusecase.GetEpisodeInfosByBangumiID(bangumiSubjectID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	episodes := make([]Episode, 0, len(episodeInfos))

	for _, episodeInfo := range episodeInfos {
		subtitles := []string{}
		if episodeInfo.Subtitles != "" {
			subtitles = strings.Split(episodeInfo.Subtitles, ";")
		}

		episodes = append(episodes, Episode{
			ID:                  episodeInfo.ID,
			DandanplayEpisodeID: episodeInfo.EpisodeDandanplayID,
			Title:               episodeInfo.Title,
			Type:                episodeInfo.TypeDescription,
			Introduction:        episodeInfo.Introduce,
			FileName:            episodeInfo.FileName,
			FilePath:            episodeInfo.FilePath,
			Subtitles:           subtitles,
		})
	}

	bangumiInfo, err := bangumiusecase.GetBangumiInfo(bangumiSubjectID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get bangumi info"})
	}

	fullPath := episodeInfos[0].FilePath
	relativePath, err := filepath.Rel(config.DefaultMediaLibraryPath, fullPath)
	if err != nil {
		fmt.Println("Error:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to Handing Path"})
	}

	respBangumiInfo := BangumiInfo{
		ID:            bangumiInfo.BangumiSubjectID,
		Title:         bangumiInfo.Name,
		Directory:     relativePath,
		ImageURL:      fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d/image?type=large", bangumiInfo.BangumiSubjectID),
		Summary:       bangumiInfo.Summary,
		Episodes:      episodes,
		RateScore:     bangumiInfo.RateScore,
		TotalEpisodes: bangumiInfo.TotalEpisodes,
		AirDate:       bangumiInfo.AirDate,
		Platform:      bangumiInfo.Platform,
	}

	return c.JSON(http.StatusOK, respBangumiInfo)
}

func GetBangumiInfoList(c echo.Context) error {
	bangumiInfo, err := bangumiusecase.GetAllBangumiInfo()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get bangumi info"})
	}

	var bangumiList []BangumiInfo
	for _, bangumi := range bangumiInfo {
		bangumiList = append(bangumiList, BangumiInfo{
			ID:                  bangumi.BangumiSubjectID,
			BangumiSubjectID:    bangumi.BangumiSubjectID,
			ImageURL:            fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d/image?type=medium", bangumi.BangumiSubjectID),
			Title:               bangumi.Name,
			DandanplayBangumiID: bangumi.DandanplayBangumiID,
			RateScore:           bangumi.RateScore,
			TotalEpisodes:       bangumi.TotalEpisodes,
			AirDate:             bangumi.AirDate,
			Platform:            bangumi.Platform,
		})
	}

	return c.JSON(http.StatusOK, bangumiList)
}

func GetEpisodeInfoByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid episode ID"})
	}

	episodeInfo, err := episodeusecase.GetEpisodeInfoById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	// Remove the leading path from episodeInfo.FilePath
	episodeInfo.FilePath = filepath.Base(episodeInfo.FilePath)

	subtitles := []string{}
	if episodeInfo.Subtitles != "" {
		subtitles = strings.Split(episodeInfo.Subtitles, ";")
	}

	resEpisodeINfo := Episode{
		ID:                  episodeInfo.ID,
		DandanplayEpisodeID: episodeInfo.EpisodeDandanplayID,
		Title:               episodeInfo.Title,
		Type:                episodeInfo.TypeDescription,
		Introduction:        episodeInfo.Introduce,
		FileName:            episodeInfo.FileName,
		FilePath:            episodeInfo.FilePath,
		Subtitles:           subtitles,
	}

	return c.JSON(http.StatusOK, resEpisodeINfo)
}
