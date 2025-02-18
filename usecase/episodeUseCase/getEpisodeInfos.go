//go:build !js && !wasm
// +build !js,!wasm

package episodeusecase

import "github.com/duckfeather10086/dandan-prime/database"

func GetEpisodeInfosByBangumiID(bnangumi_subject_id int) ([]database.EpisodeInfo, error) {
	var episodes []database.EpisodeInfo

	err := database.DB.Model(&database.EpisodeInfo{}).Where("bangumi_bangumi_id = ?", bnangumi_subject_id).Scan(&episodes).Error

	if err != nil {
		return nil, err
	}

	return episodes, nil
}
