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
	return DB.AutoMigrate(&EpisodeInfo{}, &BangumiInfo{}, &EpisodeThumbNail{}, &UserInfo{})
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

func DeleteEpisodeInfoByHash(hash string) error {
	return DB.Where("hash =?", hash).Delete(&EpisodeInfo{}).Error
}

func CheckFileExists(fileName string) (bool, error) {
	var count int64
	err := DB.Model(&EpisodeInfo{}).Where("file_name =?", fileName).Count(&count).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	if count == 0 {
		return false, nil
	}

	return true, err
}

func GetEpisodeInfoByHash(hash string) (*EpisodeInfo, error) {
	var episode EpisodeInfo
	err := DB.Where("hash = ?", hash).First(&episode).Error
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

func InitUserInfo(userInfo *UserInfo) error {
	return DB.Model(userInfo).Save(userInfo).Error
}

func GetUserInfoByUserId(id uint) (*UserInfo, error) {
	var userInfo UserInfo
	err := DB.Where("id =?", id).First(&userInfo).Error
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func UpdateUserInfoByUserId(userID uint, userInfo *UserInfo) error {
	return DB.Model(&UserInfo{}).Where("id =?", userID).Updates(userInfo).Error
}
