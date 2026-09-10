//go:build !js && !wasm
// +build !js,!wasm

package bangumi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/duckfeather10086/dandan-prime/internal/bangumi/constants"
)

func FetchBangumiSubjectDetails(subjectId int) (constants.BangumiSubjectResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%d", constants.BANGUMI_API_HOST+constants.BANGUMI_API_SUBJECT_DETAILS, subjectId), nil)
	if err != nil {
		return constants.BangumiSubjectResponse{}, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("User-Agent", constants.BANGUMI_API_USER_AGENT)

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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return constants.BangumiEpisodesResponse{}, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("User-Agent", constants.BANGUMI_API_USER_AGENT)

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

// SearchBangumiSubjects looks a keyword up on bangumi.tv and returns the
// candidates in the order bangumi.tv ranked them.
//
// The order matters and must not be re-sorted by string similarity: directory
// names are usually romaji or English while the subjects are Japanese/Chinese,
// so character-level similarity between the two is close to meaningless
// ("Dungeon Meshi" vs "ダンジョン飯" scores ~0). bangumi.tv's own search already
// resolves that cross-language step, and deciding between the candidates it
// returns is the job handed to the LLM in matchusecase.
func SearchBangumiSubjects(keyword string, limit int) ([]constants.BangumiSearchSubject, error) {
	return SearchBangumiSubjectsOfType(keyword, limit, []int{constants.SUBJECT_TYPE_ANIME})
}

// SearchBangumiSubjectsOfType is SearchBangumiSubjects with an explicit subject
// type filter. Pass constants.SubjectTypesVideo to also reach tokusatsu and
// concert footage, which an anime-only search cannot see.
func SearchBangumiSubjectsOfType(keyword string, limit int, types []int) ([]constants.BangumiSearchSubject, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 6
	}
	if len(types) == 0 {
		types = []int{constants.SUBJECT_TYPE_ANIME}
	}

	payload, err := json.Marshal(map[string]interface{}{
		"keyword": keyword,
		"filter":  map[string]interface{}{"type": types},
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s?limit=%d", constants.BANGUMI_API_HOST, constants.BANGUMI_API_SEARCH_SUBJECTS, limit)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("User-Agent", constants.BANGUMI_API_USER_AGENT)

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bangumi search %q: status %d: %s", keyword, resp.StatusCode, truncate(bodyData, 200))
	}

	var searchResp constants.BangumiSearchResponse
	if err := json.Unmarshal(bodyData, &searchResp); err != nil {
		return nil, fmt.Errorf("bangumi search %q: %v", keyword, err)
	}

	return searchResp.Data, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
