//go:build !js && !wasm
// +build !js,!wasm

package episodeusecase

import (
	"fmt"
	"log"

	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/bangumi"
)

func InitialEpisodeIntroduce() error {
	var bangumiSubjectIDs []int

	res := database.DB.Model(&database.EpisodeInfo{}).
		Group("bangumi_bangumi_id").
		Select("bangumi_bangumi_id").
		Scan(&bangumiSubjectIDs)

	if res.Error != nil {
		log.Println("Failed to get dandanplay bangumi ids , err:", res.Error)
		return res.Error
	}

	for _, bangumiSubjectID := range bangumiSubjectIDs {
		fetchedBangumiEpisodeInfo, err := bangumi.FetchBangumiEpisodes(bangumiSubjectID, 100, 0)
		if err != nil {
			log.Println("Failed to get bangumi episodes , err:", err)
			continue
		}

		for _, bangumiEpisode := range fetchedBangumiEpisodeInfo.Data {
			fmt.Println(bangumiEpisode)
		}
	}

	return nil
}
