//go:build !js && !wasm
// +build !js,!wasm

package filesacnusecase

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/dandanplay"
)

var allowedExtensions = []string{".mkv", ".mp4"}

const BATCH_SIZE = 32

func ScanAndSaveMedia(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, allowedExt := range allowedExtensions {
			if ext == allowedExt {
				hash, err := CalculateFileHash(path)
				if err != nil {
					return fmt.Errorf("error calculating hash for %s: %v", path, err)
				}

				// Check if the episode already exists in the database
				_, err = database.GetEpisodeInfoByHash(hash)
				if err == nil {
					fmt.Printf("File already exists in database: %s\n", path)
					return nil
				}

				// If the episode doesn't exist, create a new one
				episode := &database.EpisodeInfo{
					FileName: filepath.Base(path),
					Hash:     hash,
					FilePath: path,
					// Other fields will be filled later when we integrate with DandanPlay API
				}

				if err := database.SaveEpisodeInfo(episode); err != nil {
					return fmt.Errorf("error saving episode info for %s: %v", path, err)
				}

				fmt.Printf("Saved new episode: %s\n", path)
				break
			}
		}
		return nil
	})
}

func ScanAndMatchMedia(rootPath string) error {
	var totalUnmatchedEpisodes int64

	err := database.DB.Model(&database.EpisodeInfo{}).Count(&totalUnmatchedEpisodes).Error
	if err != nil {
		return err
	}

	for i := int64(0); i < totalUnmatchedEpisodes/32+1; i += 1 {
		var episodes []database.EpisodeInfo
		if err := database.DB.Limit(BATCH_SIZE).Offset(int(i) * BATCH_SIZE).Find(&episodes).Error; err != nil {
			return err
		}

		if len(episodes) == 0 {
			fmt.Println("No new episodes to match.")
			return nil
		}

		fmt.Printf("Matching %d/%d episodes with DandanPlay API...\n", len(episodes)*int(i), totalUnmatchedEpisodes)
		matchResp, err := dandanplay.BatchMatchEpisodes(episodes)
		if err != nil {
			log.Printf("Error matching episodes: %v", err)
			return err
		}

		for _, result := range matchResp.Matches {
			if result.Success {
				log.Printf("Matched episode: %s", result.Result.EpisodeTitle)
				database.UpdateEpisodeInfoByHash(result.FileHash, &database.EpisodeInfo{
					WorkDandanplayID:    result.Result.AnimeID,
					WorkTitle:           result.Result.AnimeTitle,
					Title:               result.Result.EpisodeTitle,
					Type:                result.Result.Type,
					TypeDescription:     result.Result.TypeDescription,
					EpisodeDandanplayID: result.Result.EpisodeID,
				})
			}
		}
	}

	fmt.Println("Finished matching episodes.")
	return nil
}

func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.CopyN(hash, file, 16*1024*1024); err != nil && err != io.EOF {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
