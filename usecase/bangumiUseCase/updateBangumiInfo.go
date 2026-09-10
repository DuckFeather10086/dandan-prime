//go:build !js && !wasm
// +build !js,!wasm

package bangumiusecase

import (
	"log"

	"github.com/duckfeather10086/dandan-prime/database"
)

// UpdateBangumiInfo updates a row by its primary key.
func UpdateBangumiInfo(id uint, bangumiInfo *database.BangumiInfo) error {
	err := database.DB.Model(&database.BangumiInfo{}).Where("id = ?", id).Updates(bangumiInfo).Error
	if err != nil {
		log.Println("Failed to update bangumi info, err:", err)
		return err
	}
	return nil
}

// UpdateBangumiLastWatchedEpisode records the episode last watched for a
// bangumi.tv subject.
//
// It takes the subject id, not the BangumiInfo primary key. The caller used to
// pass EpisodeInfo.BangumiID, a column nothing has ever written, so the lookup
// was always "id = 0" and this silently updated no rows -- which is why the
// per-bangumi resume position never worked. The subject id is what every other
// path here keys on.
func UpdateBangumiLastWatchedEpisode(bangumiSubjectID int, episodeID uint) error {
	if bangumiSubjectID == 0 {
		return nil
	}

	err := database.DB.Model(&database.BangumiInfo{}).
		Where("bangumi_subject_id = ?", bangumiSubjectID).
		Update("last_watched_episode_id", episodeID).Error
	if err != nil {
		log.Println("Failed to update bangumi last watched episode, err:", err)
		return err
	}
	return nil
}
