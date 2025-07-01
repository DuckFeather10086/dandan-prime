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
		Where("dandanplay_bangumi_id IS NOT NULL").
		Where("bangumi_matched =?", false).
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

		if fetchedBangumiInfo.ID != 0 {
			database.DB.Model(&database.EpisodeInfo{}).Where("dandanplay_bangumi_id = ?", dandanPlayAnimeID).Update("bangumi_bangumi_id", fetchedBangumiInfo.ID).Update("bangumi_matched", true)
		}

		time.Sleep(time.Duration(0.2 * float64(time.Second)))
	}

	if err := updateMissingBangumiIDs(); err != nil {
		log.Printf("Failed to update missing bangumi IDs: %v", err)
	}

	return nil
}

func updateMissingBangumiIDs() error {
	// Find all dandanplay_bangumi_id groups that have episodes with bangumi_bangumi_id = 0
	var dandanplayGroups []struct {
		DandanplayBangumiID int `gorm:"column:dandanplay_bangumi_id"`
	}

	err := database.DB.Model(&database.EpisodeInfo{}).
		Select("dandanplay_bangumi_id").
		Where("dandanplay_bangumi_id != 0 AND bangumi_bangumi_id = 0").
		Group("dandanplay_bangumi_id").
		Find(&dandanplayGroups).Error

	if err != nil {
		return fmt.Errorf("failed to find dandanplay groups: %v", err)
	}

	for _, group := range dandanplayGroups {
		// Try to find a non-zero bangumi_bangumi_id for this dandanplay_bangumi_id
		var targetBangumiID int
		err := database.DB.Model(&database.EpisodeInfo{}).
			Select("bangumi_bangumi_id").
			Where("dandanplay_bangumi_id = ? AND bangumi_bangumi_id != 0", group.DandanplayBangumiID).
			Limit(1).
			Pluck("bangumi_bangumi_id", &targetBangumiID).Error

		if err == nil && targetBangumiID != 0 {
			// Found a valid bangumi_bangumi_id, update all episodes with the same dandanplay_bangumi_id
			err = database.DB.Model(&database.EpisodeInfo{}).
				Where("dandanplay_bangumi_id = ? AND bangumi_bangumi_id = 0", group.DandanplayBangumiID).
				Updates(map[string]interface{}{
					"bangumi_bangumi_id": targetBangumiID,
					"bangumi_matched":    true,
				}).Error

			if err != nil {
				log.Printf("Failed to update episodes for dandanplay_bangumi_id %d: %v", group.DandanplayBangumiID, err)
			}
		} else {
			// No valid bangumi_bangumi_id found for this dandanplay_bangumi_id, set is_episode_matched to false
			err = database.DB.Model(&database.EpisodeInfo{}).
				Where("dandanplay_bangumi_id = ?", group.DandanplayBangumiID).
				Update("bangumi_matched", false).Error

			if err != nil {
				log.Printf("Failed to update bangumi_matched for dandanplay_bangumi_id %d: %v", group.DandanplayBangumiID, err)
			}
		}
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
				// fix the id error for dandanplay api id 4849 (Neon Genesis Evangelion: The End of Evangelion)
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

func UpdateBangumiInfo(bangumiID uint, bangumiInfo *database.BangumiInfo) error {
	err := database.DB.Model(&database.BangumiInfo{}).Where("id =?", bangumiID).Updates(bangumiInfo).Error
	if err != nil {
		log.Println("Failed to update bangumi info, err:", err)
		return err
	}
	return nil
}

func UpdateBangumiLastWatchedEpisode(bangumiID uint, episodeID uint) error {
	err := database.DB.Model(&database.BangumiInfo{}).Where("id =?", bangumiID).Update("last_watched_episode_id", episodeID).Error
	if err != nil {
		log.Println("Failed to update bangumi last watched episode, err:", err)
		return err
	}
	return nil
}
