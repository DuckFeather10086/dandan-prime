//go:build !js && !wasm
// +build !js,!wasm

package episodeusecase

import "github.com/duckfeather10086/dandan-prime/database"

func GetEpisodeInfoById(id int) (database.EpisodeInfo, error) {
	var episode database.EpisodeInfo

	err := database.DB.Model(&database.EpisodeInfo{}).Where("id = ?", id).Find(&episode).Error

	if err != nil {
		return database.EpisodeInfo{}, err
	}

	return episode, nil
}
