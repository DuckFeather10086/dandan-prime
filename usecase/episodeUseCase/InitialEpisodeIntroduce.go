//go:build !js && !wasm
// +build !js,!wasm

package episodeusecase

import (
	"log"
	"time"

	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/bangumi"
	"github.com/duckfeather10086/dandan-prime/internal/bangumi/constants"
)

const (
	episodePageSize     = 100
	bangumiCallThrottle = 350 * time.Millisecond
)

// PopulateEpisodeMetadata fills in per-episode titles and summaries from
// bangumi.tv for every subject in the library.
//
// This used to be impossible to notice was broken: FetchBangumiEpisodes built
// its URL by appending "/v0/episodes" to a host already ending in /v0 and had
// been returning 404 since it was written, and the function then only printed
// what it got. Nothing ever reached the database, so episode titles came
// exclusively from dandanplay's match response.
//
// Episodes are matched by number. bangumi.tv exposes both `ep` (the number
// within the episode's own type) and `sort` (the position across all of them);
// `ep` is the one that lines up with how files are numbered, with `sort` as a
// fallback for subjects that leave `ep` at zero.
func PopulateEpisodeMetadata(force bool) error {
	var subjectIDs []int
	res := database.DB.Model(&database.EpisodeInfo{}).
		Distinct().
		Where("bangumi_bangumi_id != 0").
		Pluck("bangumi_bangumi_id", &subjectIDs)
	if res.Error != nil {
		return res.Error
	}

	log.Printf("metadata: %d subjects to describe (force=%v)", len(subjectIDs), force)

	for _, subjectID := range subjectIDs {
		byNumber, err := fetchEpisodesByNumber(subjectID)
		if err != nil {
			log.Printf("metadata: subject %d: %v", subjectID, err)
			continue
		}
		if len(byNumber) == 0 {
			continue
		}

		var episodes []database.EpisodeInfo
		query := database.DB.Where("bangumi_bangumi_id = ? AND episode_no > 0", subjectID)
		if !force {
			query = query.Where("title = '' OR title IS NULL")
		}
		if err := query.Find(&episodes).Error; err != nil {
			log.Printf("metadata: subject %d: %v", subjectID, err)
			continue
		}

		for _, episode := range episodes {
			source, ok := byNumber[episode.EpisodeNo]
			if !ok {
				continue
			}

			title := source.NameCN
			if title == "" {
				title = source.Name
			}

			update := database.EpisodeInfo{
				Title:            title,
				Introduce:        source.Desc,
				AirDate:          source.Airdate,
				EpisodeBangumiID: source.ID,
			}
			if err := database.UpdateEpisodeInfoByID(episode.ID, &update); err != nil {
				log.Printf("metadata: %s: %v", episode.FileName, err)
			}
		}
	}

	return nil
}

// fetchEpisodesByNumber pages through a subject's episodes and indexes them by
// the number a file would carry.
func fetchEpisodesByNumber(subjectID int) (map[int]constants.BangumiEpisode, error) {
	byNumber := map[int]constants.BangumiEpisode{}

	for offset := 0; ; offset += episodePageSize {
		time.Sleep(bangumiCallThrottle)
		page, err := bangumi.FetchBangumiEpisodes(subjectID, episodePageSize, offset)
		if err != nil {
			return nil, err
		}
		if len(page.Data) == 0 {
			break
		}

		for _, episode := range page.Data {
			number := episode.Ep
			if number == 0 {
				number = episode.Sort
			}
			if number == 0 {
				continue
			}
			// Keep the first one seen: a subject can repeat a number across
			// episode types (a main episode 1 and a special 1), and the main
			// run comes first.
			if _, exists := byNumber[number]; !exists {
				byNumber[number] = episode
			}
		}

		if offset+len(page.Data) >= page.Total {
			break
		}
	}

	return byNumber, nil
}
