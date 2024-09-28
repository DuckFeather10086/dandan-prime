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
	FilePath            string `gorm:"size:512;index"`
}

type BangumiInfo struct {
	gorm.Model
	Name                string `gorm:"size:255;index"`
	DandanplayBangumiID int    `gorm:"index"`
	BangumiSubjectID    int    `gorm:"index"`
	RateScore           float64
	Rank                int
	TotalEpisodes       int
	Summary             string `gorm:"type:text"`
	AirDate             string `gorm:"size:20"`
	Platform            string `gorm:"size:50"`
}
