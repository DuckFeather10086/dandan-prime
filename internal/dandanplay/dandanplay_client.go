//go:build !js && !wasm
// +build !js,!wasm

package dandanplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/dandanplay/constants"
)

func BatchMatchEpisodes(episodes []database.EpisodeInfo) (constants.BatchMatchResponse, error) {
	var requests []constants.MatchRequest
	for _, episode := range episodes {
		requests = append(requests, constants.MatchRequest{
			FileName:  episode.FileName,
			FileHash:  episode.Hash,
			FileSize:  0, // We don't have this information in our current model
			MatchMode: "hashAndFileName",
		})
	}

	jsonData, err := json.Marshal(map[string][]constants.MatchRequest{"requests": requests})
	if err != nil {
		return constants.BatchMatchResponse{}, err
	}

	resp, err := http.Post(constants.DANDANPLAY_API_MATCH, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return constants.BatchMatchResponse{}, err
	}
	defer resp.Body.Close()

	var matchResp constants.BatchMatchResponse
	bodyData, err := io.ReadAll(resp.Body)

	if err := json.Unmarshal(bodyData, &matchResp); err != nil {
		return constants.BatchMatchResponse{}, err
	}

	if !matchResp.Success {
		log.Printf("API request was not successful: %v", matchResp)

		return constants.BatchMatchResponse{}, fmt.Errorf("API request was not successful")
	}

	return matchResp, nil
}

func FetchBangumiDetails(bangumiID int) (constants.BangumiDetailsResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%d", constants.DANDANPLAY_API_WORK_DETAILS, bangumiID))
	if err != nil {
		return constants.BangumiDetailsResponse{}, err
	}
	defer resp.Body.Close()

	var bangumiDetailsResp constants.BangumiDetailsResponse
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return constants.BangumiDetailsResponse{}, err
	}

	if err := json.Unmarshal(bodyData, &bangumiDetailsResp); err != nil {
		return constants.BangumiDetailsResponse{}, err
	}

	return bangumiDetailsResp, nil
}
