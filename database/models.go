package database

import "gorm.io/gorm"

type EpisodeInfo struct {
	gorm.Model
	Name                string
	Title               string
	Hash                string
	WorkName            string
	Season              int
	EpisodeNo           int
	Type                string
	TypeDescription     string
	Rating              float32
	AirDate             string
	WorkID              uint
	WorkTitle           string
	WorkDandanplayID    int
	WorkBangumiID       int
	EpisodeDandanplayID int
	EpisodeBangumiID    int
	Introduce           string
	Length              int
	Filename            string
	FilePath            string
}

type WorkInfo struct {
	gorm.Model
	Name         string
	DandanplayID int
	BangumiID    int
	RateScore    float32
	Rank         int
	EpisodeCnt   int
	Summary      string
	AirDate      string
	Platform     string
}
