//go:build !js && !wasm
// +build !js,!wasm

package bangumiusecase

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
			log.Println("OnlineDatabase.Name", OnlineDatabase.Name)
			if OnlineDatabase.Name == "Bangumi.tv" {
				bangumiUrlSplit := strings.Split(OnlineDatabase.URL, "/")

				bangumiID, err = strconv.Atoi(bangumiUrlSplit[len(bangumiUrlSplit)-1])
				if err != nil {
					log.Println("Failed to convert bangumi id , err:", err)
					continue
				} else {
					break
				}

			}
		}

		if dandanPlayAnimeID == 4847 {
			bangumiID = 772
			//fix the id error for dandanplay api id 4847 (Evangelion: 1.0 You Are (Not) Alone)
		}

		if dandanPlayAnimeID == 202 {
			bangumiID = 6049
			// fix the id error for dousingplay api id 4849 (Neon Genesis Evangelion: The End of Evangelion)
		}

		fetchedBangumiInfo, err := bangumi.FetchBangumiSubjectDetails(bangumiID)
		if err != nil {
			log.Println("Failed to get bangumi info , err:", err)
			continue
		}
		log.Println(fetchedBangumiInfo)

		bangumiInfo := database.BangumiInfo{
			BangumiSubjectID:    fetchedBangumiInfo.ID,
			DandanplayBangumiID: dandanPlayAnimeID,
			Name:                fetchedBangumiInfo.Name,
			Summary:             fetchedBangumiInfo.Summary,
			Rank:                fetchedBangumiInfo.Rating.Rank,
			TotalEpisodes:       fetchedBangumiInfo.TotalEpisodes,
			RateScore:           fetchedBangumiInfo.Rating.Score,
			AirDate:             fetchedBangumiInfo.Date,
			Platform:            fetchedBangumiInfo.Platform,
		}

		database.DB.Model(&database.BangumiInfo{}).Save(&bangumiInfo)
		database.DB.Model(&database.BangumiInfo{}).Where("bangumi_subject_id =?", fetchedBangumiInfo.ID).
			Update("dandanplay_bangumi_id", dandanPlayAnimeID).
			Update("air_date", fetchedBangumiInfo.Date).
			Update("platform", fetchedBangumiInfo.Platform)

		database.DB.Model(&database.EpisodeInfo{}).Where("dandanplay_bangumi_id = ?", dandanPlayAnimeID).Update("bangumi_bangumi_id", fetchedBangumiInfo.ID)

		time.Sleep(time.Duration(0.2 * float64(time.Second)))
	}

	return nil
}

func IncrementalUpdateBangumiInfo() error {
	var dandanPlayeBangumiIDs []int

	res := database.DB.Model(&database.EpisodeInfo{}).
		Group("dandanplay_bangumi_id").
		Select("dandanplay_bangumi_id").
		Scan(&dandanPlayeBangumiIDs)

	if res.Error != nil {
		log.Println("Failed to get dandanplay bangumi ids , err:", res.Error)
		return res.Error
	}

	for _, dandanPlayeBangumiID := range dandanPlayeBangumiIDs {
		log.Println("dandanPlayAnimeID:", dandanPlayeBangumiID)
		var count int64
		database.DB.Model(&database.BangumiInfo{}).Where("dandanplay_bangumi_id =?", dandanPlayeBangumiID).Count(&count)

		if count != 0 {
			continue
		} else {
			dandanplayBangumiInfo, err := dandanplay.FetchBangumiDetails(dandanPlayeBangumiID)
			if err != nil {
				log.Println("Failed to get bangumi info , err:", err)
				continue
			}

			var bangumiID int

			for _, OnlineDatabase := range dandanplayBangumiInfo.Bangumi.OnlineDatabases {
				if OnlineDatabase.Name == "Bangumi.tv" {
					fmt.Println(OnlineDatabase.URL)

					bangumiUrlSplit := strings.Split(OnlineDatabase.URL, "/")

					bangumiID, err = strconv.Atoi(bangumiUrlSplit[len(bangumiUrlSplit)-1])

					if err != nil {
						log.Println("Failed to convert bangumi id , err:", err)
						continue
					} else {
						break
					}
				}
			}

			if dandanPlayeBangumiID == 4847 {
				bangumiID = 772
				//fix the id error for dandanplay api id 4847 (Evangelion: 1.0 You Are (Not) Alone)
			}

			if dandanPlayeBangumiID == 202 {
				bangumiID = 6049
				// fix the id error for dousingplay api id 4849 (Neon Genesis Evangelion: The End of Evangelion)
			}

			fetchedBangumiInfo, err := bangumi.FetchBangumiSubjectDetails(bangumiID)
			if err != nil {
				log.Println("Failed to get bangumi info , err:", err)
				continue
			}

			bangumiInfo := database.BangumiInfo{
				BangumiSubjectID:    fetchedBangumiInfo.ID,
				DandanplayBangumiID: dandanPlayeBangumiID,
				Name:                fetchedBangumiInfo.Name,
				Summary:             fetchedBangumiInfo.Summary,
				Rank:                fetchedBangumiInfo.Rating.Rank,
				TotalEpisodes:       fetchedBangumiInfo.TotalEpisodes,
				RateScore:           fetchedBangumiInfo.Rating.Score,
				AirDate:             fetchedBangumiInfo.Date,
				Platform:            fetchedBangumiInfo.Platform,
			}

			database.DB.Model(&database.BangumiInfo{}).Save(&bangumiInfo)
			database.DB.Model(&database.BangumiInfo{}).Where("bangumi_subject_id =?", fetchedBangumiInfo.ID).
				Update("dandanplay_bangumi_id", dandanPlayeBangumiID).
				Update("air_date", fetchedBangumiInfo.Date).
				Update("platform", fetchedBangumiInfo.Platform)

			database.DB.Model(&database.EpisodeInfo{}).Where("dandanplay_bangumi_id = ?", dandanPlayeBangumiID).Update("bangumi_bangumi_id", fetchedBangumiInfo.ID)

			time.Sleep(time.Duration(0.2 * float64(time.Second)))
		}
	}
	return nil
}
