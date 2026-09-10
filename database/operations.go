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

// UpdateEpisodeInfoByID is the safe way to write to one episode. Hash is no
// longer unique, so UpdateEpisodeInfoByHash can touch several rows.
func UpdateEpisodeInfoByID(id uint, episode *EpisodeInfo) error {
	return DB.Model(&EpisodeInfo{}).Where("id = ?", id).Updates(episode).Error
}

func UpdateEpisodeInfoByHash(hash string, episode *EpisodeInfo) error {
	return DB.Model(&EpisodeInfo{}).Where("hash = ?", hash).Updates(episode).Error
}

func DeleteEpisodeInfoByHash(hash string) error {
	return DB.Where("hash =?", hash).Delete(&EpisodeInfo{}).Error
}

// CheckFileExists reports whether a file at this exact location is already
// known.
//
// It takes the directory as well as the name on purpose: keying on the base
// name alone silently drops every file whose name repeats elsewhere in the
// library, and release layouts repeat names constantly -- four different
// "Menu.mkv", or the same episode present in both a TV folder and a batch
// folder. Those files were skipped by the scanner and never appeared in the
// library at all.
func CheckFileExists(dirPath string, fileName string) (bool, error) {
	var count int64
	err := DB.Model(&EpisodeInfo{}).
		Where("file_path = ? AND file_name = ?", dirPath, fileName).
		Count(&count).Error
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
