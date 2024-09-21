package usecases

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/duckfeather10086/dandan-prime/config"
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

func (mu *MediaUsecase) ScrapeDandanplay() error {
	cfg := config.GetConfig()
	files, err := getMediaFiles(cfg.MediaLibraryPath, cfg.AllowedExtensions)
	if err != nil {
		return err
	}

	for i := 0; i < len(files); i += 32 {
		end := i + 32
		if end > len(files) {
			end = len(files)
		}
		batch := files[i:end]

		requests := make([]DandanplayRequest, len(batch))
		for j, file := range batch {
			hash, err := calculateFileHash(file)
			if err != nil {
				continue
			}
			fileInfo, err := os.Stat(file)
			if err != nil {
				continue
			}
			requests[j] = DandanplayRequest{
				FileName:  filepath.Base(file),
				FileHash:  hash,
				FileSize:  fileInfo.Size(),
				MatchMode: "hashAndFileName",
			}
		}

		responses, err := mu.callDandanplayAPI(requests)
		if err != nil {
			return err
		}

		for j, response := range responses {
			if len(response.Matches) > 0 {
				match := response.Matches[0]
				err := mu.createOrUpdateEpisode(batch[j], match)
				if err != nil {
					fmt.Printf("Error updating episode: %v\n", err)
				}
			}
		}
	}

	return nil
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
	episode.WorkName = match.AnimeTitle
	episode.EpisodeNo = match.Episode
	episode.Type = match.Type
	episode.TypeDescription = match.TypeDescription
	episode.WorkDandanplayID = match.AnimeID
	episode.EpisodeDandanplayID = match.EpisodeID
	episode.FilePath = filePath

	result = mu.db.Save(&episode)
	if result.Error != nil {
		return result.Error
	}

	return mu.createOrUpdateWork(match.AnimeID, match.AnimeTitle)
}

func (mu *MediaUsecase) createOrUpdateWork(animeID int, animeTitle string) error {
	var work database.WorkInfo
	result := mu.db.Where(database.WorkInfo{DandanplayID: animeID}).FirstOrCreate(&work)
	if result.Error != nil {
		return result.Error
	}

	work.Name = animeTitle
	work.DandanplayID = animeID

	result = mu.db.Save(&work)
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

func (mu *MediaUsecase) GetWorks() ([]database.WorkInfo, error) {
	var works []database.WorkInfo
	result := mu.db.Find(&works)
	return works, result.Error
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
