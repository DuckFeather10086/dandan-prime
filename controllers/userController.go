//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/duckfeather10086/dandan-prime/database"
	bangumiUseCase "github.com/duckfeather10086/dandan-prime/usecase/bangumiUseCase"
	episodeUseCase "github.com/duckfeather10086/dandan-prime/usecase/episodeUseCase"
	userUseCase "github.com/duckfeather10086/dandan-prime/usecase/userUseCase"
	"github.com/labstack/echo/v4"
)

func GetLastWatchedInfo(c echo.Context) error {
	userID := c.QueryParam("user_id")
	log.Print("user_id", userID)
	userIDUint, err := strconv.Atoi(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	userInfo, err := userUseCase.GetUserInfoByUserId(uint(userIDUint))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user info"})
	}

	bangumiInfo, err := bangumiUseCase.GetBangumiInfo(userInfo.LastWatchedBangumiID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get bangumi info"})
	}

	episodeInfo, err := episodeUseCase.GetEpisodeInfoById(userInfo.LastWatchedEpisodeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}

	type LastWatchedBangumiResponse struct {
		AnimeTitle              string `json:"anime_title"`
		LastWatchedBangumiID    int    `json:"last_watched_bangumi_id"`
		LastWatchedEpisodeID    int    `json:"last_watched_episode_id"`
		LastWatchedEpisodeTitle string `json:"last_watched_episode_title"`
		LasteWatchedEpisodeNo   int    `json:"last_watched_episode_no"`
		ImageUrl                string `json:"imageUrl"`
		PosterUrl               string `json:"posterUrl"`
	}

	lastWatchedBangumiResponse := LastWatchedBangumiResponse{
		AnimeTitle:              bangumiInfo.Name,
		LastWatchedBangumiID:    userInfo.LastWatchedBangumiID,
		LastWatchedEpisodeID:    userInfo.LastWatchedEpisodeID,
		LastWatchedEpisodeTitle: episodeInfo.Title,
		LasteWatchedEpisodeNo:   episodeInfo.EpisodeNo,
		ImageUrl:                fmt.Sprintf("https://api.bgm.tv/v0/subjects/%v/image?type=large", bangumiInfo.BangumiSubjectID),
		PosterUrl:               fmt.Sprintf("https://api.bgm.tv/v0/subjects/%v/image?type=large", bangumiInfo.BangumiSubjectID),
	}

	return c.JSON(http.StatusOK, lastWatchedBangumiResponse)
}

func UpdateLastedWatched(c echo.Context) error {
	// Define a struct to hold the request body data
	type UpdateRequest struct {
		UserID               string `json:"user_id"`
		LastWatchedEpisodeID string `json:"last_watched_episode_id"`
	}

	reqBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read request body"})
	}
	log.Println("body", string(reqBody))

	var request UpdateRequest
	// Bind the request body to the struct
	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON in request body"})
	}

	userIDUint, err := strconv.Atoi(request.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	episodeIDInt, err := strconv.Atoi(request.LastWatchedEpisodeID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid episode ID"})
	}

	episodeInfo, err := episodeUseCase.GetEpisodeInfoById(episodeIDInt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get episode info"})
	}
	log.Println("episodeinfo:", episodeInfo.BangumiBangumiID)

	userInfo := database.UserInfo{
		LastWatchedBangumiID: int(episodeInfo.BangumiBangumiID),
		LastWatchedEpisodeID: episodeIDInt,
	}

	err = userUseCase.UpdateUserInfo(uint(userIDUint), &userInfo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update last watched info err:" + err.Error()})
	}

	err = bangumiUseCase.UpdateBangumiLastWatchedEpisode(episodeInfo.BangumiID, episodeInfo.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update bangumi last watched info err:" + err.Error()})
	}

	return c.NoContent(http.StatusOK)
}
