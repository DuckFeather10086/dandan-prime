//go:build !js && !wasm
// +build !js,!wasm

package episodeusecase

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"math/rand"

	"github.com/duckfeather10086/dandan-prime/database"
	"github.com/duckfeather10086/dandan-prime/internal/dandanplay"
	"github.com/duckfeather10086/dandan-prime/internal/ffmpegutil"
)

var allowedExtensionsVideo = []string{".mkv", ".mp4"}
var allowedExtensionsSubtitle = []string{".ass", ".ssa", ".srt", ".sub"}

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
		for _, allowedExt := range allowedExtensionsVideo {
			if ext == allowedExt {
				fileExists, err := database.CheckFileExists(filepath.Base(path))
				if err != nil {
					return fmt.Errorf("error checking file existence for %s: %v", path, err)
				}
				if fileExists {
					return nil
				}

				hash, err := CalculateFileHash(path)
				if err != nil {
					return fmt.Errorf("error calculating hash for %s: %v", path, err)
				}

				// If the episode doesn't exist, create a new one
				episode := &database.EpisodeInfo{
					FileName: filepath.Base(path),
					Hash:     hash,
					FilePath: filepath.Dir(path),
					// Other fields will be filled later when we integrate with DandanPlay API
				}

				// Check if the episode already exists in the database
				_, err = database.GetEpisodeInfoByHash(hash)
				if err == nil {
					database.UpdateEpisodeInfoByHash(hash, episode)
				} else {
					if err := database.CreateEpisodeInfo(episode); err != nil {
						return fmt.Errorf("error saving episode info for %s: %v", path, err)
					}
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

	err := database.DB.Model(&database.EpisodeInfo{}).Where("info_matched =?", false).Count(&totalUnmatchedEpisodes).Error
	if err != nil {
		return err
	}

	for i := int64(0); i < totalUnmatchedEpisodes/32+1; i += 1 {
		var episodes []database.EpisodeInfo
		if err := database.DB.Where("info_matched =?", false).Limit(BATCH_SIZE).Offset(int(i) * BATCH_SIZE).Find(&episodes).Error; err != nil {
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
				episodeNO, err := strconv.Atoi(strings.TrimLeft(fmt.Sprintf("%04d", result.Result.EpisodeID%10000), "0"))
				if err != nil {
					log.Printf("Failed to convert episode number: %v", err)
					continue
				}
				database.UpdateEpisodeInfoByHash(result.FileHash, &database.EpisodeInfo{
					DandanplayBangumiID: result.Result.AnimeID,
					BangumiTitle:        result.Result.AnimeTitle,
					Title:               result.Result.EpisodeTitle,
					Type:                result.Result.Type,
					TypeDescription:     result.Result.TypeDescription,
					EpisodeDandanplayID: result.Result.EpisodeID,
					EpisodeNo:           episodeNO,
					InfoMatched:         true,
				})
			} else {
				err := database.UpdateEpisodeInfoByHash(result.FileHash, &database.EpisodeInfo{
					InfoMatched: true,
				})
				log.Println("result false, err", err)
			}
		}
	}

	fmt.Println("Finished matching episodes.")
	return nil
}

func ScanAndMatchSubtitles() error {
	var episodes []database.EpisodeInfo
	if err := database.DB.Where("subtitle_matched =?", false).Find(&episodes).Error; err != nil {
		return err
	}

	for _, episode := range episodes {
		subtitles := []string{}
		episodeDir := filepath.Dir(filepath.Join(episode.FilePath, episode.FileName))

		err := filepath.Walk(episodeDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if !contains(allowedExtensionsSubtitle, ext) {
				return nil
			}

			subtitleFileName := filepath.Base(path)
			episodeFileName := strings.TrimSuffix(episode.FileName, filepath.Ext(episode.FileName))

			if strings.HasPrefix(subtitleFileName, episodeFileName) {
				subtitles = append(subtitles, subtitleFileName)
			}

			return nil
		})

		if err != nil {
			log.Printf("error scanning subtitles for %s: %v \n", episode.FileName, err)
			continue
		}

		if len(subtitles) > 0 {
			subtitlesStr := strings.Join(subtitles, ";")
			if err := database.UpdateEpisodeInfoByHash(episode.Hash, &database.EpisodeInfo{
				Subtitles:       subtitlesStr,
				SubtitleMatched: true,
			}); err != nil {
				return fmt.Errorf("error updating subtitles for %s: %v", episode.FileName, err)
			}
		}
	}

	return nil
}

func ScanAndPrepareThumbnails() error {
	var episodes []database.EpisodeInfo
	if err := database.DB.Find(&episodes).Where("episode_dandanplay_id != 0").Where("dandanplay_bangumi_id != 0").Group("episode_dandanplay_id").Error; err != nil {
		return err
	}

	//outpoutfile := config.DefaultEpisodeThumbnailPath

	for _, episode := range episodes {
		var thumbNailCount int64
		database.DB.Model(&database.EpisodeThumbNail{}).Where("episode_id =?", episode.ID).Count(&thumbNailCount)
		if thumbNailCount > 0 {
			continue
		} else {
			// Generate random second between 0 and video duration
			randSecond := rand.Intn(int(60))
			outputFilename, err := ffmpegutil.GenerateThumbnail(fmt.Sprintf("%s/%s", episode.FilePath, episode.FileName), fmt.Sprintf("thumbnails/%d.jpg", episode.ID), fmt.Sprintf("00:00:%02d", randSecond))
			if err != nil {
				log.Printf("error generating thumbnail for %s: %v", episode.FileName, err)
				continue
			}

			if outputFilename != "" {
				episodeThumbNailInfo := database.EpisodeThumbNail{
					EpisodeID:      episode.ID,
					ThumbNailImage: outputFilename,
				}

				database.DB.Model(&database.EpisodeThumbNail{}).Save(&episodeThumbNailInfo)
			}
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	buffer := make([]byte, 16*1024*1024) // 16MB buffer

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	hash.Write(buffer[:n])

	return hex.EncodeToString(hash.Sum(nil)), nil
}
