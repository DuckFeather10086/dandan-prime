//go:build !js && !wasm
// +build !js,!wasm

package bangumiusecase

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/bangumi"
	"github.com/duckfeather10086/dandan-prime/internal/dandanplay"
)

func InitializeBangumiInfo() error {
	var dandanPlayeBangumiIDs []int

	res := database.DB.Model(&database.EpisodeInfo{}).
		Group("dandanplay_bangumi_id").
		Select("dandanplay_bangumi_id").
		Scan(&dandanPlayeBangumiIDs)

	if res.Error != nil {
		log.Println("Failed to get dandanplay bangumi ids , err:", res.Error)
		return res.Error
	}

	for _, dandanPlayAnimeID := range dandanPlayeBangumiIDs {

		dandanplayBangumiInfo, err := dandanplay.FetchBangumiDetails(dandanPlayAnimeID)
		if err != nil {
			log.Println("Failed to get bangumi info , err:", err)
			continue
		}

		var bangumiID int
		for _, OnlineDatabase := range dandanplayBangumiInfo.Bangumi.OnlineDatabases {
			if OnlineDatabase.Name == "bangumi" {
				fmt.Println(OnlineDatabase.URL)
			}

			bangumiUrlSplit := strings.Split(OnlineDatabase.URL, "/")

			bangumiID, err = strconv.Atoi(bangumiUrlSplit[len(bangumiUrlSplit)-1])
			if err != nil {
				log.Println("Failed to convert bangumi id , err:", err)
				continue
			} else {
				break
			}
		}

		fetchedBangumiInfo, err := bangumi.FetchBangumiSubjectDetails(bangumiID)
		if err != nil {
			log.Println("Failed to get bangumi info , err:", err)
			continue
		}

		bangumiInfo := database.BangumiInfo{
			BangumiID:     fetchedBangumiInfo.ID,
			DandanplayID:  bangumiID,
			Name:          fetchedBangumiInfo.Name,
			Summary:       fetchedBangumiInfo.Summary,
			Rank:          fetchedBangumiInfo.Rating.Rank,
			TotalEpisodes: fetchedBangumiInfo.TotalEpisodes,
			RateScore:     fetchedBangumiInfo.Rating.Score,
		}

		database.DB.Model(&database.BangumiInfo{}).Save(&bangumiInfo)

		database.DB.Model(&database.EpisodeInfo{}).Where("dandanplay_bangumi_id = ?", dandanPlayAnimeID).Update("bangumi_id", fetchedBangumiInfo.ID)
	}

	return nil
}
