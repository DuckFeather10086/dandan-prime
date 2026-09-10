//go:build !js && !wasm
// +build !js,!wasm

// Package matchusecase resolves media-library directories to bangumi.tv
// subjects without going through dandanplay.
//
// Why not dandanplay: it was only ever used as an *id oracle*. Its /match
// endpoint turned a file hash into an anime id, and its bangumi details
// carried a Bangumi.tv link that gave the subject id the metadata actually
// came from. Two problems followed from that.
//
// First, dandanplay's granularity became the library's granularity. All seven
// Kara no Kyoukai films share one dandanplay anime id, so they collapsed into
// one library entry -- bangumi.tv models them as seven separate subjects. The
// EVA entries hardcoded in bangumiUseCase were the same failure: a wrong
// Bangumi.tv link on the dandanplay side, patched one id at a time.
//
// Second, a file whose hash is not in dandanplay's index never matches, and
// there is nothing the user can do about it, so it sits in the library
// unmatched forever.
//
// This resolver searches bangumi.tv directly and asks an LLM to pick from the
// candidates it returns. The LLM is used for exactly one thing: deciding which
// candidate a directory refers to, including across languages, where string
// similarity is useless ("Dungeon Meshi" vs "ダンジョン飯" scores ~0 by
// character overlap but bangumi.tv's own search ranks it first). It is never
// the source of metadata -- subject ids must come from the candidate list, and
// anything it returns outside that list is rejected below.
//
// Hashes are still computed at scan time and still used for danmaku, which is
// the one thing dandanplay is genuinely authoritative about.
package matchusecase

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duckfeather10086/dandan-prime/config"
	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/bangumi"
	bangumiconstants "github.com/duckfeather10086/dandan-prime/internal/bangumi/constants"
	"github.com/duckfeather10086/dandan-prime/internal/llm"
)

const (
	candidateLimit      = 6
	candidateLimitRetry = 12

	// How many file names go into one prompt. A collection mapping can only
	// name files the model was shown, so this bounds how big a directory can
	// be fully mapped in one call; anything left over is picked up by the
	// leaf-level pass.
	promptFileCap  = 100
	searchThrottle = 350 * time.Millisecond
	sidecarName    = ".bangumi-id"
)

// SidecarName is the per-directory manual override file. A directory
// containing it is never sent to the LLM, which makes a bad match a one-time
// fix rather than something to re-litigate on every rescan. Two forms:
//
//	723            -- whole directory is this subject
//	01.mkv 233     -- one line per file: <file name> <subject id>
const SidecarName = sidecarName

type Stats struct {
	Directories int
	Series      int
	Collections int
	Sidecar     int
	Unresolved  int
	Episodes    int
	PromptTok   int
	OutputTok   int
}

type resolution struct {
	Mode      string         `json:"mode"`
	SubjectID int            `json:"subject_id"`
	Map       map[string]int `json:"map"`
	Query     string         `json:"query"`
	Reason    string         `json:"reason"`
}

const systemPrompt = `你是动画媒体库匹配器。输入是一个目录名、目录下的视频文件名、以及 bangumi.tv 搜索返回的候选条目。
判断该目录属于哪一种，并只输出 JSON：

A) 剧集/单作品：所有视频文件属于同一个 bangumi subject。
   输出 {"mode":"series","subject_id":<int>,"reason":"<20字内>"}

B) 电影合集：目录里每个正片文件是一部独立作品，各自对应不同 subject（如系列剧场版合集）。
   输出 {"mode":"collection","map":{"<文件名>":<subject_id>,...},"reason":"<20字内>"}

规则：
- subject_id 必须来自候选列表，禁止编造。
- 候选里没有正确答案时输出 {"mode":"unknown","query":"<你建议的日文搜索关键词>","reason":"..."}
- Menu/Reminder/NCOP/NCED/SP/特典 等非正片文件不要放进 map。
- 只输出 JSON，不要解释、不要代码块。`

// ResolveMediaLibrary walks every directory that still has unresolved
// episodes and writes bangumi.tv subject ids onto the episode rows.
//
// It resolves per directory but assigns per file: a collection directory maps
// each film to its own subject, which is what dandanplay could not express.
// A non-zero limit stops after that many directories, so a large library can
// be worked through in controlled batches rather than one unbounded run.
// It runs in two passes. The first groups by top-level directory, which is
// one call for most works. Whatever files that leaves unassigned -- a
// franchise directory holding several distinct works in subdirectories -- is
// then resolved per leaf directory. Because both passes select only episodes
// without a subject id, the second pass naturally sees exactly the remainder,
// and costs calls only where the first pass could not finish the job.
func ResolveMediaLibrary(force bool, limit int) (Stats, error) {
	var stats Stats

	client := llm.New(config.LLM_BASE_URL, config.LLM_API_KEY, config.LLM_MODEL)

	for _, level := range []granularity{topLevel, leaf} {
		passStats, err := resolvePass(client, force, limit, level)
		if err != nil {
			return stats, err
		}
		stats.add(passStats)

		// A forced re-run has already rewritten everything at top level; a
		// leaf pass would only churn the same rows.
		if force {
			break
		}
	}

	log.Printf("resolver: done: %+v", stats)
	return stats, nil
}

func (s *Stats) add(o Stats) {
	s.Directories += o.Directories
	s.Series += o.Series
	s.Collections += o.Collections
	s.Sidecar += o.Sidecar
	s.Unresolved += o.Unresolved
	s.Episodes += o.Episodes
	s.PromptTok += o.PromptTok
	s.OutputTok += o.OutputTok
}

func resolvePass(client *llm.Client, force bool, limit int, level granularity) (Stats, error) {
	var stats Stats

	dirs, err := unresolvedDirectories(force, level)
	if err != nil {
		return stats, err
	}
	if limit > 0 && len(dirs) > limit {
		dirs = dirs[:limit]
	}
	stats.Directories = len(dirs)

	name := "top-level"
	if level == leaf {
		name = "leaf"
	}
	log.Printf("resolver: %s pass: %d directories to resolve (force=%v)", name, len(dirs), force)

	for i, dir := range dirs {
		episodes, err := episodesInDirectory(dir, force)
		if err != nil {
			log.Printf("resolver: %s: %v", dir, err)
			continue
		}
		if len(episodes) == 0 {
			continue
		}

		log.Printf("resolver: [%d/%d] %s (%d files)", i+1, len(dirs), filepath.Base(dir), len(episodes))

		if applied, err := applySidecar(dir, episodes); err != nil {
			log.Printf("resolver: %s: sidecar: %v", dir, err)
		} else if applied > 0 {
			stats.Sidecar++
			stats.Episodes += applied
			continue
		}

		if !client.Configured() {
			stats.Unresolved++
			continue
		}

		res, usage, err := resolveDirectory(client, dir, episodes)
		stats.PromptTok += usage.PromptTokens
		stats.OutputTok += usage.CompletionTokens
		if err != nil {
			log.Printf("resolver: %s: %v", dir, err)
			stats.Unresolved++
			continue
		}

		applied, err := applyResolution(dir, episodes, res)
		if err != nil {
			log.Printf("resolver: %s: %v", dir, err)
			stats.Unresolved++
			continue
		}
		if applied == 0 {
			stats.Unresolved++
			continue
		}

		switch res.Mode {
		case "series":
			stats.Series++
		case "collection":
			stats.Collections++
		}
		stats.Episodes += applied
	}

	return stats, nil
}

// resolveDirectory asks bangumi.tv for candidates and the LLM to choose. When
// the first keyword returns nothing usable the LLM is asked for a Japanese
// keyword instead and the search is retried once -- rule-based keyword
// extraction fails on a fair share of release-group names, and this recovers
// those without another parsing rule.
func resolveDirectory(client *llm.Client, dir string, episodes []database.EpisodeInfo) (resolution, llm.Usage, error) {
	var total llm.Usage

	// Pool the candidates from every plausible keyword rather than betting on
	// one. A directory yields at most a handful, and the extra searches are
	// cheaper than a wasted LLM round trip on an empty candidate list.
	keywords := SearchKeywords(dir)

	seen := map[int]bool{}
	var candidates []bangumiconstants.BangumiSearchSubject
	for _, kw := range keywords {
		found, err := searchCandidates(kw, candidateLimit)
		if err != nil {
			log.Printf("resolver: %s: search %q: %v", dir, kw, err)
			continue
		}
		for _, c := range found {
			if seen[c.ID] {
				continue
			}
			seen[c.ID] = true
			candidates = append(candidates, c)
		}
	}

	res, usage, err := ask(client, dir, episodes, candidates)
	total.PromptTokens += usage.PromptTokens
	total.CompletionTokens += usage.CompletionTokens
	if err != nil {
		return resolution{}, total, err
	}

	if res.Mode == "unknown" {
		// Retry with the type filter widened, using BOTH the original keyword
		// and whatever the model suggested.
		//
		// The original keyword goes first and matters most: an anime-only
		// search returns nothing for tokusatsu, so a directory can fail pass 1
		// purely because of the filter while its keyword was fine all along.
		// Kamen Rider ZEZTZ is exactly that -- the keyword finds subject 545304
		// immediately once the filter allows live action, whereas the model,
		// having seen no candidates at all, can only guess a broader Japanese
		// title ("仮面ライダー") that buries the answer under the whole franchise.
		retryKeywords := append([]string{}, keywords...)
		if q := strings.TrimSpace(res.Query); q != "" {
			retryKeywords = append(retryKeywords, q)
		}
		log.Printf("resolver: %s: retrying widened with %q (%s)", filepath.Base(dir), retryKeywords, res.Reason)

		seen := map[int]bool{}
		var retryCandidates []bangumiconstants.BangumiSearchSubject
		for _, kw := range retryKeywords {
			found, err := searchCandidatesOfType(kw, candidateLimitRetry, bangumiconstants.SubjectTypesVideo)
			if err != nil {
				log.Printf("resolver: %s: retry search %q: %v", dir, kw, err)
				continue
			}
			for _, c := range found {
				if seen[c.ID] {
					continue
				}
				seen[c.ID] = true
				retryCandidates = append(retryCandidates, c)
			}
		}
		if len(retryCandidates) == 0 {
			return res, total, nil
		}

		res2, usage2, err := ask(client, dir, episodes, retryCandidates)
		total.PromptTokens += usage2.PromptTokens
		total.CompletionTokens += usage2.CompletionTokens
		if err != nil {
			return res, total, err
		}
		return res2, total, validate(res2, retryCandidates)
	}

	return res, total, validate(res, candidates)
}

// validate rejects any subject id the model produced that was not in the
// candidate list, so a hallucinated id can never reach the database.
func validate(res resolution, candidates []bangumiconstants.BangumiSearchSubject) error {
	allowed := make(map[int]bool, len(candidates))
	for _, c := range candidates {
		allowed[c.ID] = true
	}

	switch res.Mode {
	case "series":
		if !allowed[res.SubjectID] {
			return fmt.Errorf("subject %d not among candidates", res.SubjectID)
		}
	case "collection":
		for file, id := range res.Map {
			if !allowed[id] {
				return fmt.Errorf("subject %d (for %q) not among candidates", id, file)
			}
		}
	}
	return nil
}

func ask(client *llm.Client, dir string, episodes []database.EpisodeInfo, candidates []bangumiconstants.BangumiSearchSubject) (resolution, llm.Usage, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "目录名:\n  %s\n\n视频文件(%d个):\n", filepath.Base(dir), len(episodes))
	for i, ep := range episodes {
		if i >= promptFileCap {
			fmt.Fprintf(&sb, "  ...另有 %d 个文件\n", len(episodes)-i)
			break
		}
		fmt.Fprintf(&sb, "  %s\n", relativeName(dir, ep))
	}
	sb.WriteString("\nbangumi.tv 候选:\n")
	if len(candidates) == 0 {
		sb.WriteString("  (无)\n")
	}
	for _, c := range candidates {
		fmt.Fprintf(&sb, "  id=%d | %s | %s | %s\n", c.ID, c.Name, c.NameCN, c.Date)
	}

	text, usage, err := client.Chat(systemPrompt, sb.String())
	if err != nil {
		return resolution{}, usage, err
	}

	var res resolution
	if err := json.Unmarshal([]byte(text), &res); err != nil {
		return resolution{}, usage, fmt.Errorf("unparseable model output %q: %v", text, err)
	}
	return res, usage, nil
}

// applyResolution writes subject ids onto the episode rows. Episode numbers
// come from the filenames, except in a collection where each file is a
// separate one-episode work.
func applyResolution(dir string, episodes []database.EpisodeInfo, res resolution) (int, error) {
	switch res.Mode {
	case "series":
		if res.SubjectID == 0 {
			return 0, fmt.Errorf("series with no subject id")
		}
		n := 0
		for _, ep := range episodes {
			update := database.EpisodeInfo{
				BangumiBangumiID: res.SubjectID,
				BangumiMatched:   true,
			}
			if no, ok := EpisodeNumber(ep.FileName); ok {
				update.EpisodeNo = no
			}
			if err := database.UpdateEpisodeInfoByID(ep.ID, &update); err != nil {
				log.Printf("resolver: %s: %v", ep.FileName, err)
				continue
			}
			n++
		}
		return n, ensureBangumiInfo(res.SubjectID)

	case "collection":
		if len(res.Map) == 0 {
			return 0, fmt.Errorf("collection with empty map")
		}
		byName := make(map[string]database.EpisodeInfo, len(episodes)*2)
		for _, ep := range episodes {
			byName[relativeName(dir, ep)] = ep
			// Accept a bare base name too: the model is given relative paths
			// but a flat directory makes the two identical anyway.
			byName[ep.FileName] = ep
		}
		subjects := map[int]bool{}
		n := 0
		for file, subjectID := range res.Map {
			ep, ok := byName[file]
			if !ok {
				log.Printf("resolver: %s: mapped file %q not in directory", filepath.Base(dir), file)
				continue
			}
			// Each film is its own subject, so it is episode 1 of that subject
			// rather than episode N of the collection.
			update := database.EpisodeInfo{
				BangumiBangumiID: subjectID,
				EpisodeNo:        1,
				BangumiMatched:   true,
			}
			if err := database.UpdateEpisodeInfoByID(ep.ID, &update); err != nil {
				log.Printf("resolver: %s: %v", file, err)
				continue
			}
			subjects[subjectID] = true
			n++
		}
		for subjectID := range subjects {
			if err := ensureBangumiInfo(subjectID); err != nil {
				log.Printf("resolver: subject %d: %v", subjectID, err)
			}
		}
		return n, nil

	case "unknown":
		return 0, nil
	}
	return 0, fmt.Errorf("unknown mode %q", res.Mode)
}

// applySidecar honours a manual override, and returns how many rows it wrote.
func applySidecar(dir string, episodes []database.EpisodeInfo) (int, error) {
	data, err := os.ReadFile(filepath.Join(dir, sidecarName))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	perFile := map[string]int{}
	wholeDir := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if id, err := strconv.Atoi(line); err == nil {
			wholeDir = id
			continue
		}
		idx := strings.LastIndexAny(line, " \t")
		if idx < 0 {
			return 0, fmt.Errorf("malformed sidecar line %q", line)
		}
		id, err := strconv.Atoi(strings.TrimSpace(line[idx+1:]))
		if err != nil {
			return 0, fmt.Errorf("malformed sidecar line %q", line)
		}
		perFile[strings.TrimSpace(line[:idx])] = id
	}

	if wholeDir != 0 {
		return applyResolution(dir, episodes, resolution{Mode: "series", SubjectID: wholeDir})
	}
	if len(perFile) > 0 {
		return applyResolution(dir, episodes, resolution{Mode: "collection", Map: perFile})
	}
	return 0, nil
}

// ensureBangumiInfo fetches subject metadata from bangumi.tv and upserts the
// BangumiInfo row, keyed on the subject id rather than the dandanplay id.
func ensureBangumiInfo(subjectID int) error {
	var existing database.BangumiInfo
	err := database.DB.Where("bangumi_subject_id = ?", subjectID).First(&existing).Error
	if err == nil && existing.Name != "" {
		return nil
	}

	details, err := bangumi.FetchBangumiSubjectDetails(subjectID)
	if err != nil {
		return err
	}
	if details.ID == 0 {
		return fmt.Errorf("subject %d: empty response", subjectID)
	}

	name := details.NameCN
	if name == "" {
		name = details.Name
	}

	info := database.BangumiInfo{
		BangumiSubjectID: details.ID,
		Name:             name,
		Summary:          details.Summary,
		Rank:             details.Rating.Rank,
		RateScore:        details.Rating.Score,
		TotalEpisodes:    details.TotalEpisodes,
		AirDate:          details.Date,
		Platform:         details.Platform,
	}
	if existing.ID != 0 {
		info.Model = existing.Model
	}

	time.Sleep(searchThrottle)
	return database.DB.Save(&info).Error
}

func searchCandidates(keyword string, limit int) ([]bangumiconstants.BangumiSearchSubject, error) {
	time.Sleep(searchThrottle)
	return bangumi.SearchBangumiSubjects(keyword, limit)
}

// searchCandidatesOfType widens the search beyond animation. It is used only
// on the retry, after an anime-only search has already come back without the
// answer: tokusatsu and concert video are filed under other subject types, and
// a library holds both.
func searchCandidatesOfType(keyword string, limit int, types []int) ([]bangumiconstants.BangumiSearchSubject, error) {
	time.Sleep(searchThrottle)
	return bangumi.SearchBangumiSubjectsOfType(keyword, limit, types)
}

// unresolvedDirectories returns the *top-level* directories under the library
// root that still need resolving, not the leaf directories the files sit in.
//
// EpisodeInfo.FilePath is a leaf, and release layouts nest freely --
// "Evangelion 3.0+1.01/EXTRA/PV&CM", "Cowboy Bebop .../TV (1S26E)_.../[...]".
// Resolving per leaf would split one work across several library entries and
// spend an LLM call on each fragment. The top-level directory is the unit a
// human filed the work under, so it is the unit resolved here; the LLM is
// shown paths relative to it and can still tell a nested TV folder from a
// nested Film folder.
type granularity int

const (
	// topLevel groups by the first directory under the library root: the unit a
	// human filed a work under. It resolves most of a library in one call each.
	topLevel granularity = iota
	// leaf groups by the directory a file actually sits in. Needed where a
	// top-level directory holds a whole franchise -- "Strike the Blood" with
	// six seasons and two OVA arcs in subdirectories, each its own subject --
	// which no single subject id can describe.
	leaf
)

func unresolvedDirectories(force bool, level granularity) ([]string, error) {
	var leaves []string
	query := database.DB.Model(&database.EpisodeInfo{}).
		Distinct().
		Where("file_path != ''")
	if !force {
		query = query.Where("bangumi_bangumi_id = 0 OR bangumi_bangumi_id IS NULL")
	}
	if err := query.Pluck("file_path", &leaves).Error; err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var dirs []string
	for _, dir := range leaves {
		if level == topLevel {
			dir = topLevelDirectory(dir)
		}
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs, nil
}

// topLevelDirectory maps a leaf directory to the first level under the media
// library root. A file sitting directly in the root maps to its own directory.
func topLevelDirectory(leaf string) string {
	root := filepath.Clean(config.MEDIA_LIBRARY_ROOT_PATH)
	clean := filepath.Clean(leaf)

	rel, err := filepath.Rel(root, clean)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return clean
	}
	first := strings.Split(rel, string(os.PathSeparator))[0]
	return filepath.Join(root, first)
}

// episodesInDirectory collects every episode at or below dir.
func episodesInDirectory(dir string, force bool) ([]database.EpisodeInfo, error) {
	var episodes []database.EpisodeInfo
	query := database.DB.Where("file_path = ? OR file_path LIKE ?", dir, dir+string(os.PathSeparator)+"%")
	if !force {
		query = query.Where("bangumi_bangumi_id = 0 OR bangumi_bangumi_id IS NULL")
	}
	if err := query.Order("file_path, file_name").Find(&episodes).Error; err != nil {
		return nil, err
	}
	return episodes, nil
}

// relativeName is how a file is named in the prompt and in a collection map:
// relative to the directory being resolved, so nested layouts stay
// distinguishable and two files with the same base name cannot collide.
func relativeName(dir string, ep database.EpisodeInfo) string {
	full := filepath.Join(ep.FilePath, ep.FileName)
	rel, err := filepath.Rel(dir, full)
	if err != nil {
		return ep.FileName
	}
	return rel
}

// BackfillEpisodeNumbers re-derives episode numbers from file names for rows
// that do not have one.
//
// Episode numbers are written during resolution, but they come from the file
// name rather than from bangumi.tv or the model, so an improvement to the
// parser should not require re-resolving the library and spending the LLM
// budget again. This applies the current parser on its own: no network, no
// model, idempotent.
//
// force also revisits rows that already have a number, for when the parser
// changed its mind rather than merely learned something new.
func BackfillEpisodeNumbers(force bool) (int, error) {
	query := database.DB.Model(&database.EpisodeInfo{})
	if !force {
		query = query.Where("episode_no = 0")
	}

	var episodes []database.EpisodeInfo
	if err := query.Find(&episodes).Error; err != nil {
		return 0, err
	}

	updated := 0
	for _, episode := range episodes {
		number, ok := EpisodeNumber(episode.FileName)
		if !ok || number == episode.EpisodeNo {
			continue
		}
		if err := database.UpdateEpisodeInfoByID(episode.ID, &database.EpisodeInfo{EpisodeNo: number}); err != nil {
			log.Printf("backfill: %s: %v", episode.FileName, err)
			continue
		}
		updated++
	}

	// A film has no number in its file name because there is nothing to
	// number. Left at zero it lands in the season page's "unknown" bucket and
	// renders as bonus material, which is wrong for the feature itself.
	films, err := numberSingleEpisodeWorks(force)
	if err != nil {
		return updated, err
	}
	updated += films

	log.Printf("backfill: set an episode number on %d rows (%d of them single-episode works)", updated, films)
	return updated, nil
}

// numberSingleEpisodeWorks assigns episode 1 to a subject whose library
// holds exactly one feature file for it.
//
// The obvious test -- bangumi.tv's total_episodes -- does not work: it counts
// bonus entries, so 言の葉の庭, a single film, reports 2. Counting the feature
// files actually present is both simpler and self-evident: if a subject has
// one feature file and that file carries no number, it is that subject's only
// episode.
func numberSingleEpisodeWorks(force bool) (int, error) {
	var episodes []database.EpisodeInfo
	query := database.DB.Where("bangumi_bangumi_id != 0")
	if !force {
		query = query.Where("episode_no = 0")
	}
	if err := query.Find(&episodes).Error; err != nil {
		return 0, err
	}

	// Count the feature files per subject across the whole library, not just
	// the rows selected above, so a subject that already has numbered
	// episodes is not mistaken for a film.
	var all []database.EpisodeInfo
	if err := database.DB.Where("bangumi_bangumi_id != 0").Find(&all).Error; err != nil {
		return 0, err
	}
	features := map[int]int{}
	for _, episode := range all {
		if !IsExtra(episode.FileName) {
			features[episode.BangumiBangumiID]++
		}
	}

	updated := 0
	for _, episode := range episodes {
		if IsExtra(episode.FileName) || episode.EpisodeNo == 1 {
			continue
		}
		if features[episode.BangumiBangumiID] != 1 {
			continue
		}
		if err := database.UpdateEpisodeInfoByID(episode.ID, &database.EpisodeInfo{EpisodeNo: 1}); err != nil {
			log.Printf("backfill: %s: %v", episode.FileName, err)
			continue
		}
		updated++
	}
	return updated, nil
}
