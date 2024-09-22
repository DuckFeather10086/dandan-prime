//go:build !js && !wasm
// +build !js,!wasm

package bangumi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/duckfeather10086/dandan-prime/internal/bangumi/constants"
)

func FetchBangumiSubjectDetails(subjectId int) (constants.BangumiSubjectResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%d", constants.BANGUMI_API_HOST+constants.BANGUMI_API_SUBJECT_DETAILS, subjectId), nil)
	if err != nil {
		return constants.BangumiSubjectResponse{}, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("User-Agent", "dandan-prime")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("error in fetching errors1: err", err)
		return constants.BangumiSubjectResponse{}, err
	}
	defer resp.Body.Close()

	var bangumiDetailsResp constants.BangumiSubjectResponse
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("error in fetching errors2: err", err)
		return constants.BangumiSubjectResponse{}, err
	}

	if err := json.Unmarshal(bodyData, &bangumiDetailsResp); err != nil {
		log.Println("error in fetching errors3: err", err)
		return constants.BangumiSubjectResponse{}, err
	}

	return bangumiDetailsResp, nil
}

func FetchBangumiEpisodes(subjectId int, limit int, offset int) (constants.BangumiEpisodesResponse, error) {
	url := fmt.Sprintf("%s%s?subject_id=%d&limit=%d&offset=%d", constants.BANGUMI_API_HOST, constants.BANGUMI_API_EPISODES, subjectId, limit, offset)
	resp, err := http.Get(url)
	if err != nil {
		return constants.BangumiEpisodesResponse{}, err
	}
	defer resp.Body.Close()

	var bangumiEpisodesResp constants.BangumiEpisodesResponse
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return constants.BangumiEpisodesResponse{}, err
	}

	if err := json.Unmarshal(bodyData, &bangumiEpisodesResp); err != nil {
		return constants.BangumiEpisodesResponse{}, err
	}

	return bangumiEpisodesResp, nil
}
