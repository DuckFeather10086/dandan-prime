// internal/database/models.go

package database

import (
	"gorm.io/gorm"
)

type EpisodeInfo struct {
	gorm.Model
	FileName            string `gorm:"size:255;index"`
	Title               string `gorm:"size:255"`
	Hash                string `gorm:"size:32;uniqueIndex"`
	BangumiName         string `gorm:"size:255;index"`
	Season              int    `gorm:"index"`
	EpisodeNo           int    `gorm:"index;not null;default:0"`
	Type                string `gorm:"size:50"`
	TypeDescription     string `gorm:"size:255"`
	Rating              float32
	AirDate             string `gorm:"size:10"`
	BangumiID           uint   `gorm:"index"`
	BangumiTitle        string `gorm:"size:255"`
	DandanplayBangumiID int    `gorm:"index;not null;default:0"`
	BangumiBangumiID    int    `gorm:"index;not null;default:0"`
	EpisodeDandanplayID int    `gorm:"index;not null;default:0"`
	EpisodeBangumiID    int    `gorm:"index;not null;default:0"`
	Introduce           string `gorm:"type:text"`
	Subtitles           string `gorm:"type:text"`
	Length              int
	LastWatchedAt       int
	InfoMatched         bool   `gorm:"default:false"`
	BangumiMatched      bool   `gorm:"default:false"`
	SubtitleMatched     bool   `gorm:"default:false"`
	FilePath            string `gorm:"size:512;index"`
}

type BangumiInfo struct {
	gorm.Model
	Name                 string `gorm:"size:255;index"`
	DandanplayBangumiID  int    `gorm:"index"`
	BangumiSubjectID     int    `gorm:"index"`
	RateScore            float64
	Rank                 int
	TotalEpisodes        int
	Summary              string `gorm:"type:text"`
	AirDate              string `gorm:"size:20"`
	Platform             string `gorm:"size:50"`
	LastWatchedEpisodeID uint
}

type EpisodeThumbNail struct {
	gorm.Model
	EpisodeID      uint   `gorm:"episode_id"`
	ThumbNailImage string `gorm:"thumb_nail_image"`
}

type UserInfo struct {
	gorm.Model
	Username             string `gorm:"size:255;uniqueIndex"`
	Password             string `gorm:"size:255;not null"`
	Email                string `gorm:"size:255;uniqueIndex"`
	Avatar               string `gorm:"size:255"`
	LastWatchedBangumiID int
	LastWatchedEpisodeID int
}
