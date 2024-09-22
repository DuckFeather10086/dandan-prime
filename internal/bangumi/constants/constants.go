package constants

const (
	BANGUMI_API_HOST            = "https://api.bgm.tv/v0"
	BANGUMI_API_SUBJECT_DETAILS = "/subjects/"
)

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
	Infobox []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"infobox"`
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
