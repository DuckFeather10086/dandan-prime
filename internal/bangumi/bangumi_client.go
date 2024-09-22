//go:build !js && !wasm
// +build !js,!wasm

package bangumi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/duckfeather10086/dandan-prime/internal/bangumi/constants"
)

func FetchBangumiSubjectDetails(subjectId int) (constants.BangumiSubjectResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%d", constants.BANGUMI_API_HOST+constants.BANGUMI_API_SUBJECT_DETAILS, subjectId))
	if err != nil {
		return constants.BangumiSubjectResponse{}, err
	}
	defer resp.Body.Close()

	var bangumiDetailsResp constants.BangumiSubjectResponse
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return constants.BangumiSubjectResponse{}, err
	}

	if err := json.Unmarshal(bodyData, &bangumiDetailsResp); err != nil {
		return constants.BangumiSubjectResponse{}, err
	}

	return bangumiDetailsResp, nil
}
