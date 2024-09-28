//go:build !js && !wasm
// +build !js,!wasm

package bangumiusecase

import (
	"log"

	"github.com/duckfeather10086/dandan-prime/database"
)

func GetBangumiInfo(bangumiSubjectID int) (database.BangumiInfo, error) {
	log.Println("Getting bangumi info for bangumi subject ID:", bangumiSubjectID)

	var bangumi database.BangumiInfo
	err := database.DB.Where("bangumi_subject_id = ?", bangumiSubjectID).First(&bangumi).Error

	if err != nil {
		return bangumi, err
	}

	return bangumi, nil
}

func GetAllBangumiInfo() ([]database.BangumiInfo, error) {
	var bangumiList []database.BangumiInfo
	err := database.DB.Where("bangumi_subject_id != 0").Group("bangumi_subject_id").Find(&bangumiList).Error

	if err != nil {
		return nil, err
	}

	return bangumiList, nil
}
