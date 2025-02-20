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

var (
	client *DandanplayClient
)

type DandanplayClient struct {
	appID     string
	appSecret string
	client    *http.Client
}

// InitDandanplayClient 初始化客户端，应在程序启动时调用
func InitDandanplayClient(appID, appSecret string) {
	client = &DandanplayClient{
		appID:     appID,
		appSecret: appSecret,
		client:    &http.Client{},
	}
}

func (c *DandanplayClient) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("dandanplay client not initialized")
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-AppId", c.appID)
	req.Header.Set("X-AppSecret", c.appSecret)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.client.Do(req)
}

func BatchMatchEpisodes(episodes []database.EpisodeInfo) (constants.BatchMatchResponse, error) {
	var requests []constants.MatchRequest
	for _, episode := range episodes {
		requests = append(requests, constants.MatchRequest{
			FileName:  episode.FileName,
			FileHash:  episode.Hash,
			FileSize:  0,
			MatchMode: "hashAndFileName",
		})
	}

	jsonData, err := json.Marshal(map[string][]constants.MatchRequest{"requests": requests})
	if err != nil {
		return constants.BatchMatchResponse{}, err
	}

	resp, err := client.doRequest("POST", constants.DANDANPLAY_API_MATCH, bytes.NewBuffer(jsonData))
	if err != nil {
		return constants.BatchMatchResponse{}, err
	}
	defer resp.Body.Close()

	var matchResp constants.BatchMatchResponse
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return constants.BatchMatchResponse{}, err
	}

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
	resp, err := client.doRequest("GET", fmt.Sprintf("%s/%d", constants.DANDANPLAY_API_WORK_DETAILS, bangumiID), nil)
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

func FetchDanmakuFromDandanplay(episodeID int) (constants.DanmakuResponse, error) {
	url := fmt.Sprintf("%s/%d?withRelated=true", constants.DANDANPLAY_API_COMMENT, episodeID)

	resp, err := client.doRequest("GET", url, nil)
	if err != nil {
		return constants.DanmakuResponse{}, err
	}
	defer resp.Body.Close()

	var danmakuResp constants.DanmakuResponse
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return constants.DanmakuResponse{}, err
	}

	if err := json.Unmarshal(bodyData, &danmakuResp); err != nil {
		return constants.DanmakuResponse{}, err
	}

	return danmakuResp, nil
}
