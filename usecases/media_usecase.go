//go:build !js && !wasm
// +build !js,!wasm

package usecases

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/duckfeather10086/dandan-prime/database"
	"gorm.io/gorm"
)

type MediaUsecase struct {
	db *gorm.DB
}

type DandanplayRequest struct {
	FileName  string `json:"fileName"`
	FileHash  string `json:"fileHash"`
	FileSize  int64  `json:"fileSize"`
	MatchMode string `json:"matchMode"`
}

type DandanplayResponse struct {
	Matches []struct {
		EpisodeID       int    `json:"episodeId"`
		AnimeID         int    `json:"animeId"`
		AnimeTitle      string `json:"animeTitle"`
		EpisodeTitle    string `json:"episodeTitle"`
		Type            string `json:"type"`
		TypeDescription string `json:"typeDescription"`
		Episode         int    `json:"episode"`
	} `json:"matches"`
}

func (mu *MediaUsecase) callDandanplayAPI(requests []DandanplayRequest) ([]DandanplayResponse, error) {
	client := &http.Client{}
	url := "https://api.dandanplay.net/api/v2/match/batch"

	requestBody, err := json.Marshal(map[string][]DandanplayRequest{"requests": requests})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responses []DandanplayResponse
	err = json.NewDecoder(resp.Body).Decode(&responses)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func (mu *MediaUsecase) createOrUpdateEpisode(filePath string, match struct {
	EpisodeID       int    `json:"episodeId"`
	AnimeID         int    `json:"animeId"`
	AnimeTitle      string `json:"animeTitle"`
	EpisodeTitle    string `json:"episodeTitle"`
	Type            string `json:"type"`
	TypeDescription string `json:"typeDescription"`
	Episode         int    `json:"episode"`
}) error {
	var episode database.EpisodeInfo
	result := mu.db.Where(database.EpisodeInfo{EpisodeDandanplayID: match.EpisodeID}).FirstOrCreate(&episode)
	if result.Error != nil {
		return result.Error
	}

	episode.FileName = filepath.Base(filePath)
	episode.Title = match.EpisodeTitle
	episode.BangumiTitle = match.AnimeTitle
	episode.EpisodeNo = match.Episode
	episode.Type = match.Type
	episode.TypeDescription = match.TypeDescription
	episode.DandanplayBangumiID = match.AnimeID
	episode.EpisodeDandanplayID = match.EpisodeID
	episode.FilePath = filePath

	result = mu.db.Save(&episode)
	if result.Error != nil {
		return result.Error
	}

	return mu.createOrUpdateBangumi(match.AnimeID, match.AnimeTitle)
}

func (mu *MediaUsecase) createOrUpdateBangumi(animeID int, animeTitle string) error {
	var bangumi database.BangumiInfo
	result := mu.db.Where(database.BangumiInfo{DandanplayBangumiID: animeID}).FirstOrCreate(&bangumi)
	if result.Error != nil {
		return result.Error
	}

	bangumi.Name = animeTitle
	bangumi.DandanplayBangumiID = animeID

	result = mu.db.Save(&bangumi)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (mu *MediaUsecase) GetEpisodes() ([]database.EpisodeInfo, error) {
	var episodes []database.EpisodeInfo
	result := mu.db.Find(&episodes)
	return episodes, result.Error
}

func (mu *MediaUsecase) GetBangumis() ([]database.BangumiInfo, error) {
	var bangumis []database.BangumiInfo
	result := mu.db.Find(&bangumis)
	return bangumis, result.Error
}

func getMediaFiles(root string, allowedExtensions []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, ext := range allowedExtensions {
				if filepath.Ext(path) == ext {
					files = append(files, path)
					break
				}
			}
		}
		return nil
	})
	return files, err
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return getFileHash(file)
}

func getFileHash(file *os.File) (string, error) {
	hash := md5.New()
	chunkSize := 16 * 1024 * 1024 // 16MB
	buf := make([]byte, chunkSize)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	hash.Write(buf[:n])
	return hex.EncodeToString(hash.Sum(nil)), nil
}
