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
	WorkName            string `gorm:"size:255;index"`
	Season              int    `gorm:"index"`
	EpisodeNo           int    `gorm:"index"`
	Type                string `gorm:"size:50"`
	TypeDescription     string `gorm:"size:255"`
	Rating              float32
	AirDate             string `gorm:"size:10"`
	WorkID              uint   `gorm:"index"`
	WorkTitle           string `gorm:"size:255"`
	WorkDandanplayID    int    `gorm:"index"`
	WorkBangumiID       int    `gorm:"index"`
	EpisodeDandanplayID int    `gorm:"index"`
	EpisodeBangumiID    int    `gorm:"index"`
	Introduce           string `gorm:"type:text"`
	Length              int
	FilePath            string `gorm:"size:512;uniqueIndex"`
}

type WorkInfo struct {
	gorm.Model
	Name         string `gorm:"size:255;index"`
	DandanplayID int    `gorm:"uniqueIndex"`
	BangumiID    int    `gorm:"uniqueIndex"`
	RateScore    float32
	Rank         int
	EpisodeCnt   int
	Summary      string `gorm:"type:text"`
	AirDate      string `gorm:"size:10"`
	Platform     string `gorm:"size:50"`
}
