//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

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
	RateScore           float64   `json:"rate_score"`
	TotalEpisodes       int       `json:"total_episodes"`
	AirDate             string    `json:"air_date"`
	Platform            string    `json:"platform"`
	Title               string    `json:"title"`
	Directory           string    `json:"directory"`
	Contents            []Content `json:"contents"`
}

// Content 结构体用于存储作品目录下的内容信息
type Content struct {
	Title        string `json:"title"`
	Type         string `json:"type"`         // 例如: "video", "image", "subtitle" 等
	Introduction string `json:"introduction"` //
	FileName     string `json:"file_name"`
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

	contents := make([]Content, 0, len(episodeInfos))

	for _, episodeInfo := range episodeInfos {

		contents = append(contents, Content{
			Title:        episodeInfo.Title,
			Type:         episodeInfo.TypeDescription,
			Introduction: episodeInfo.Introduce,
			FileName:     episodeInfo.FileName,
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
		Contents:      contents,
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
