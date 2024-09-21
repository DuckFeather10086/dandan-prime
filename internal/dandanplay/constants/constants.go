//go:build !js && !wasm
// +build !js,!wasm

package constants

// Constants for fetching work details
const DANDANPLAY_API_WORK_DETAILS = "https://api.dandanplay.net/api/v2/bangumi"

type Title struct {
	Language string `json:"language"`
	Title    string `json:"title"`
}

type Episode struct {
	SeasonID      *int    `json:"seasonId"`
	EpisodeID     int     `json:"episodeId"`
	EpisodeTitle  string  `json:"episodeTitle"`
	EpisodeNumber string  `json:"episodeNumber"`
	LastWatched   *string `json:"lastWatched"`
	AirDate       *string `json:"airDate"`
}

type RelatedAnime struct {
	AnimeID       int     `json:"animeId"`
	BangumiID     string  `json:"bangumiId"`
	AnimeTitle    string  `json:"animeTitle"`
	ImageURL      string  `json:"imageUrl"`
	SearchKeyword string  `json:"searchKeyword"`
	IsOnAir       bool    `json:"isOnAir"`
	AirDay        int     `json:"airDay"`
	IsFavorited   bool    `json:"isFavorited"`
	IsRestricted  bool    `json:"isRestricted"`
	Rating        float64 `json:"rating"`
}

type Tag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type OnlineDatabase struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type BangumiInfo struct {
	Type            string           `json:"type"`
	TypeDescription string           `json:"typeDescription"`
	Titles          []Title          `json:"titles"`
	Seasons         []interface{}    `json:"seasons"`
	Episodes        []Episode        `json:"episodes"`
	Summary         string           `json:"summary"`
	Metadata        []string         `json:"metadata"`
	BangumiURL      string           `json:"bangumiUrl"`
	UserRating      int              `json:"userRating"`
	FavoriteStatus  *string          `json:"favoriteStatus"`
	Comment         *string          `json:"comment"`
	OnlineDatabases []OnlineDatabase `json:"onlineDatabases"`
	AnimeID         int              `json:"animeId"`
	BangumiID       string           `json:"bangumiId"`
	AnimeTitle      string           `json:"animeTitle"`
	ImageURL        string           `json:"imageUrl"`
	SearchKeyword   string           `json:"searchKeyword"`
	IsOnAir         bool             `json:"isOnAir"`
	AirDay          int              `json:"airDay"`
	IsFavorited     bool             `json:"isFavorited"`
	IsRestricted    bool             `json:"isRestricted"`
	Rating          float64          `json:"rating"`
}

type WorkDetailsResponse struct {
	Bangumi      BangumiInfo `json:"bangumi"`
	ErrorCode    int         `json:"errorCode"`
	Success      bool        `json:"success"`
	ErrorMessage string      `json:"errorMessage"`
}

// Constants for matching episodes
const DANDANPLAY_API_MATCH = "https://api.dandanplay.net/api/v2/match/batch"

type MatchRequest struct {
	FileName  string `json:"fileName"`
	FileHash  string `json:"fileHash"`
	FileSize  int64  `json:"fileSize"`
	MatchMode string `json:"matchMode"`
}

type MatchResult struct {
	EpisodeID       int    `json:"episodeId"`
	AnimeID         int    `json:"animeId"`
	AnimeTitle      string `json:"animeTitle"`
	EpisodeTitle    string `json:"episodeTitle"`
	Type            string `json:"type"`
	TypeDescription string `json:"typeDescription"`
	Season          int    `json:"season"`
	EpisodeNo       int    `json:"episodeNo"`
}

type MatchResults struct {
	Result   *MatchResult `json:"matchResult"`
	Success  bool         `json:"success"`
	FileHash string       `json:"fileHash"`
}

type BatchMatchResponse struct {
	Success bool           `json:"success"`
	Matches []MatchResults `json:"results"`
}
