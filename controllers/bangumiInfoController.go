//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/duckfeather10086/dandan-prime/config"
	"github.com/duckfeather10086/dandan-prime/internal/dandanplay"
	bangumiusecase "github.com/duckfeather10086/dandan-prime/usecase/bangumiUseCase"
	episodeusecase "github.com/duckfeather10086/dandan-prime/usecase/episodeUseCase"
	"github.com/labstack/echo/v4"
)

type BangumiInfo struct {
	ID                   int             `json:"id"`
	BangumiSubjectID     int             `json:"bangumi_subject_id"`
	DandanplayBangumiID  int             `json:"dandanpaly_bangumi_id"`
	ImageURL             string          `json:"image_url"`
	Summary              string          `json:"summary"`
	RateScore            float64         `json:"rate_score"`
	TotalEpisodes        int             `json:"total_episodes"`
	AirDate              string          `json:"air_date"`
	Platform             string          `json:"platform"`
	Title                string          `json:"title"`
	Directory            string          `json:"directory"`
	Episodes             json.RawMessage `json:"episodes"`
	LastWatchedEpisodeID uint            `json:"last_watched_episode_id"`
	UnknownEpisodes      []Episode       `json:"unknown_episodes"`
}

// Episode 结构体用于存储作品目录下的内容信息
type Episode struct {
	ID                  uint     `json:"id"`
	EpisodeNo           int      `json:"episode_no,omitempty"`
	DandanplayEpisodeID int      `json:"dandanpaly_episode_id"`
	Title               string   `json:"title"`
	Type                string   `json:"type"`         // 例如: "video", "image", "subtitle" 等
	Introduction        string   `json:"introduction"` //
	FileName            string   `json:"file_name"`
	Subtitles           []string `json:"subtitles"`
	FilePath            string   `json:"file_path"`
	LastWatchedAt       int      `json:"last_watched_at"`
}

type DanmakuInfo struct {
	Danmakus []Danmaku `json:"danmakus"`
}

type Danmaku struct {
	Time  int    `json:"time"`
	Text  string `json:"text"`
	Color string `json:"color"`
	Type  string `json:"type"` // top' | 'bottom' | 'scroll'
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

	episodeMap := make(map[string][]Episode)
	unknownEpisodes := []Episode{}

	for _, episodeInfo := range episodeInfos {
		subtitles := []string{}
		if episodeInfo.Subtitles != "" {
			subtitles = strings.Split(episodeInfo.Subtitles, ";")
		}

		episode := Episode{
			ID:                  episodeInfo.ID,
			DandanplayEpisodeID: episodeInfo.EpisodeDandanplayID,
			Title:               episodeInfo.Title,
			Type:                episodeInfo.TypeDescription,
			Introduction:        episodeInfo.Introduce,
			FileName:            episodeInfo.FileName,
			FilePath:            episodeInfo.FilePath,
			Subtitles:           subtitles,
		}

		if episodeInfo.EpisodeDandanplayID != 0 {
			key := strconv.Itoa(episodeInfo.EpisodeDandanplayID)
			episodeMap[key] = append(episodeMap[key], episode)
		} else {
			unknownEpisodes = append(unknownEpisodes, episode)
		}
	}

	// Sort episodes in each array, putting those with subtitles first
	for key, episodes := range episodeMap {
		sort.Slice(episodes, func(i, j int) bool {
			return len(episodes[i].Subtitles) > len(episodes[j].Subtitles)
		})
		episodeMap[key] = episodes
	}

	// Sort unknown episodes, putting those with subtitles first
	sort.Slice(unknownEpisodes, func(i, j int) bool {
		return len(unknownEpisodes[i].Subtitles) > len(unknownEpisodes[j].Subtitles)
	})

	// Add unknown episodes to the map with a special key
	if len(unknownEpisodes) > 0 {
		episodeMap["unknown"] = unknownEpisodes
	}

	// Marshal the episode map to JSON
	episodesJSON, err := json.Marshal(episodeMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal episodes"})
	}

	bangumiInfo, err := bangumiusecase.GetBangumiInfo(bangumiSubjectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get bangumi info"})
	}
	var respBangumiInfo BangumiInfo

	if len(episodeInfos) > 0 {
		fullPath := episodeInfos[0].FilePath
		relativePath, err := filepath.Rel(config.MEDIA_LIBRARY_ROOT_PATH, fullPath)
		if err != nil {
			fmt.Println("Error:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to Handing Path"})
		}
		respBangumiInfo = BangumiInfo{
			ID:                   bangumiInfo.BangumiSubjectID,
			Title:                bangumiInfo.Name,
			Directory:            relativePath,
			ImageURL:             fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d/image?type=large", bangumiInfo.BangumiSubjectID),
			Summary:              bangumiInfo.Summary,
			RateScore:            bangumiInfo.RateScore,
			TotalEpisodes:        bangumiInfo.TotalEpisodes,
			AirDate:              bangumiInfo.AirDate,
			Platform:             bangumiInfo.Platform,
			LastWatchedEpisodeID: bangumiInfo.LastWatchedEpisodeID,
			Episodes:             episodesJSON,
		}

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

	// // Remove the leading path from episodeInfo.FilePath
	episodeInfo.FilePath = strings.TrimPrefix(episodeInfo.FilePath, config.MEDIA_LIBRARY_ROOT_PATH)

	subtitles := []string{}
	if episodeInfo.Subtitles != "" {
		subtitles = strings.Split(episodeInfo.Subtitles, ";")
	}

	resEpisodeINfo := Episode{
		ID:                  episodeInfo.ID,
		EpisodeNo:           episodeInfo.EpisodeNo,
		DandanplayEpisodeID: episodeInfo.EpisodeDandanplayID,
		Title:               episodeInfo.Title,
		Type:                episodeInfo.TypeDescription,
		Introduction:        episodeInfo.Introduce,
		FileName:            episodeInfo.FileName,
		FilePath:            episodeInfo.FilePath,
		Subtitles:           subtitles,
		LastWatchedAt:       episodeInfo.LastWatchedAt,
	}

	return c.JSON(http.StatusOK, resEpisodeINfo)
}

func GetDanmakuByDandanplayEpisodeID(c echo.Context) error {
	episodeIDStr := c.Param("episode_id")
	log.Println(episodeIDStr)
	episodeID, err := strconv.Atoi(episodeIDStr)
	log.Println(episodeID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid episode ID"})
	}

	episodeInfo, err := episodeusecase.GetEpisodeInfoById(episodeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	danmakuFromDandanplay, err := dandanplay.FetchDanmakuFromDandanplay(episodeInfo.EpisodeDandanplayID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch danmaku from dandanplay"})
	}

	// Convert dandanplay danmaku to BulletOption format
	bulletOptions := make([]Danmaku, 0, len(danmakuFromDandanplay.Comments))
	if danmakuFromDandanplay.Comments != nil {
		for _, d := range danmakuFromDandanplay.Comments {
			parts := strings.Split(d.P, ",")
			if len(parts) < 3 {
				continue // Skip invalid danmaku
			}

			time, _ := strconv.ParseFloat(parts[0], 64)
			mode, _ := strconv.Atoi(parts[1])
			color, _ := strconv.Atoi(parts[2])

			danmaku := Danmaku{
				Time:  int(time), // Convert to milliseconds
				Text:  d.M,
				Color: fmt.Sprintf("#%06X", color),
				Type:  getDanmakuType(mode),
			}

			// Find the correct position to insert the new danmaku
			insertIndex := sort.Search(len(bulletOptions), func(i int) bool {
				return bulletOptions[i].Time > danmaku.Time
			})

			// Insert the new danmaku at the correct position
			bulletOptions = append(bulletOptions[:insertIndex], append([]Danmaku{danmaku}, bulletOptions[insertIndex:]...)...)
		}
	}

	response := DanmakuInfo{
		Danmakus: bulletOptions,
	}

	return c.JSON(http.StatusOK, response)
}

func GetDanmakuForDplayerByDandanplayEpisodeID(c echo.Context) error {
	id, err := strconv.Atoi(c.QueryParam("id"))
	log.Print("id ", id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid episode ID"})
	}

	episodeInfo, err := episodeusecase.GetEpisodeInfoById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}
	danmakuFromDandanplay, err := dandanplay.FetchDanmakuFromDandanplay(episodeInfo.EpisodeDandanplayID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch danmaku"})
	}

	// Convert to Dplayer format
	dplayerData := make([][]interface{}, 0)
	if danmakuFromDandanplay.Comments != nil {
		for _, d := range danmakuFromDandanplay.Comments {
			parts := strings.Split(d.P, ",")
			if len(parts) < 3 {
				continue
			}

			time, _ := strconv.ParseFloat(parts[0], 64)
			color, _ := strconv.ParseInt(parts[2], 10, 64)

			// DPlayer format ：[time(second），0，font color, author name，content]
			dplayerData = append(dplayerData, []interface{}{
				time,                     // time
				0,                        // fixed value 0
				strconv.Itoa(int(color)), // color
				"anonymous",              // author name
				d.M,                      // content
			})
		}
	}

	response := map[string]interface{}{
		"code":    0,
		"data":    dplayerData,
		"version": 3,
	}

	return c.JSON(http.StatusOK, response)
}

func DeleteBangumi(c echo.Context) error {
	idStr := c.Param("bangumi_subject_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid bangumi ID"})
	}

	err = bangumiusecase.DeleteBangumi(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete bangumi"})
	}

	return c.NoContent(http.StatusNoContent)
}

func getDanmakuType(mode int) string {
	switch mode {
	case 4:
		return "bottom"
	case 5:
		return "top"
	default:
		return "scroll"
	}
}
