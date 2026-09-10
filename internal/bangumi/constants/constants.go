//go:build !js && !wasm
// +build !js,!wasm

package constants

const (
	BANGUMI_API_HOST            = "https://api.bgm.tv/v0"
	BANGUMI_API_SUBJECT_DETAILS = "/subjects"
	BANGUMI_API_EPISODES        = "/episodes"
	BANGUMI_API_SEARCH_SUBJECTS = "/search/subjects"

	// bangumi.tv asks every client to identify itself; requests without a
	// recognisable User-Agent get rate limited harder.
	BANGUMI_API_USER_AGENT = "duckfeather10086/dandan-prime"

	// bangumi.tv subject types. Filtering a search to ANIME keeps books,
	// music and games out of the candidate list, which is what a media library
	// wants almost always -- but not quite always: tokusatsu (Kamen Rider) and
	// concert footage are filed under REAL, so a search restricted to ANIME
	// returns nothing at all for them.
	SUBJECT_TYPE_BOOK  = 1
	SUBJECT_TYPE_ANIME = 2
	SUBJECT_TYPE_MUSIC = 3
	SUBJECT_TYPE_GAME  = 4
	SUBJECT_TYPE_REAL  = 6
)

// SubjectTypesVideo is the widened filter used when an anime-only search comes
// back empty: everything that can plausibly be a video file in a media
// library, still excluding books and games.
var SubjectTypesVideo = []int{SUBJECT_TYPE_ANIME, SUBJECT_TYPE_REAL, SUBJECT_TYPE_MUSIC}

// BangumiSearchSubject is one candidate row from /v0/search/subjects. It is
// deliberately narrower than BangumiSubjectResponse: the resolver only needs
// enough to tell candidates apart, and every extra field is prompt tokens.
type BangumiSearchSubject struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	NameCN   string `json:"name_cn"`
	Date     string `json:"date"`
	Platform string `json:"platform"`
}

type BangumiSearchResponse struct {
	Total int                    `json:"total"`
	Data  []BangumiSearchSubject `json:"data"`
}

type BangumiSubjectResponse struct {
	Date     string `json:"date"`
	Platform string `json:"platform"`
	Images   struct {
		Small  string `json:"small"`
		Grid   string `json:"grid"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Common string `json:"common"`
	} `json:"images"`
	Summary string `json:"summary"`
	Name    string `json:"name"`
	NameCN  string `json:"name_cn"`
	Tags    []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
	Rating struct {
		Rank  int `json:"rank"`
		Total int `json:"total"`
		Count struct {
			One   int `json:"1"`
			Two   int `json:"2"`
			Three int `json:"3"`
			Four  int `json:"4"`
			Five  int `json:"5"`
			Six   int `json:"6"`
			Seven int `json:"7"`
			Eight int `json:"8"`
			Nine  int `json:"9"`
			Ten   int `json:"10"`
		} `json:"count"`
		Score float64 `json:"score"`
	} `json:"rating"`
	TotalEpisodes int `json:"total_episodes"`
	Collection    struct {
		OnHold  int `json:"on_hold"`
		Dropped int `json:"dropped"`
		Wish    int `json:"wish"`
		Collect int `json:"collect"`
		Doing   int `json:"doing"`
	} `json:"collection"`
	ID      int  `json:"id"`
	Eps     int  `json:"eps"`
	Volumes int  `json:"volumes"`
	Series  bool `json:"series"`
	Locked  bool `json:"locked"`
	NSFW    bool `json:"nsfw"`
	Type    int  `json:"type"`
}

type BangumiEpisodesResponse struct {
	Data   []BangumiEpisode `json:"data"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type BangumiEpisode struct {
	Airdate         string `json:"airdate"`
	Name            string `json:"name"`
	NameCN          string `json:"name_cn"`
	Duration        string `json:"duration"`
	Desc            string `json:"desc"`
	Ep              int    `json:"ep"`
	Sort            int    `json:"sort"`
	ID              int    `json:"id"`
	SubjectID       int    `json:"subject_id"`
	Comment         int    `json:"comment"`
	Type            int    `json:"type"`
	Disc            int    `json:"disc"`
	DurationSeconds int    `json:"duration_seconds"`
}
