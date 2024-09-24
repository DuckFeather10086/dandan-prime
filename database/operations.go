//go:build !js && !wasm
// +build !js,!wasm

package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase(dbPath string) error {
	var err error
	DB, err = gorm.Open(

		sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// Migrate the schema
	return DB.AutoMigrate(&EpisodeInfo{}, &BangumiInfo{})
}

func CreateEpisodeInfo(episode *EpisodeInfo) error {
	return DB.Save(episode).Error
}

func UpdateEpisodeInfo(episode *EpisodeInfo) error {
	return DB.Model(episode).Updates(episode).Error
}

func UpdateEpisodeInfoByHash(hash string, episode *EpisodeInfo) error {
	return DB.Model(&EpisodeInfo{}).Where("hash = ?", hash).Updates(episode).Error
}

func GetEpisodeInfoByHash(hash string) (*EpisodeInfo, error) {
	var episode EpisodeInfo
	err := DB.Where("hash = ?", hash).First(&episode).Error
	if err != nil {
		return nil, err
	}
	return &episode, nil
}
