//go:build !js && !wasm
// +build !js,!wasm

package ffmpegutil

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/duckfeather10086/dandan-prime/config"
)

func GenerateThumbnail(inputFile, outputFile string, timeOffset string) (string, error) {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-ss", timeOffset,
		"-s", "480x270",
		"-preset", "ultrafast",
		"-frames:v", "1",
		outputFile)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("cmd", cmd.String())

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outputFile, nil
}

func GenerateHlsPlayList(inputFile string, episodeID, segmentDuration int) error {
	// 添加缓存清理逻辑
	cacheDir := "cache"
	if err := cleanCache(cacheDir, episodeID); err != nil {
		return fmt.Errorf("failed to clean cache: %v", err)
	}

	// Get video duration using ffprobe
	log.Println("filename", inputFile)
	durationCmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputFile)

	durationOutput, err := durationCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get video duration: %v", err)
	}

	var duration float64
	_, err = fmt.Sscanf(string(durationOutput), "%f", &duration)
	if err != nil {
		return fmt.Errorf("failed to parse video duration: %v", err)
	}

	log.Println("duration", duration)

	// 获取视频的宽度和高度
	width, height, err := getVideoDimensions(inputFile)
	if err != nil {
		return err
	}

	// 计算宽高比
	var aspectRatio string
	if width*3 == height*4 {
		aspectRatio = "4:3"
	} else if width*9 == height*16 {
		aspectRatio = "16:9"
	} else {
		aspectRatio = fmt.Sprintf("%d:%d", width, height) // 其他比例
	}

	log.Println("aspect ratio", aspectRatio)

	// Calculate the number of segments
	numSegments := int(duration) / segmentDuration
	if int(duration)%segmentDuration != 0 {
		numSegments++
	}

	log.Println("numSegments", numSegments)

	resolutions := []int{480, 720, 1080}

	metaPlaylist := "#EXTM3U\n"

	for _, res := range resolutions {
		metaPlaylist += fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=2149280,RESOLUTION=1280x720,NAME=\"%d\"\n", res)

		playList := "#EXTM3U\n#EXT-X-VERSION:6\n#EXT-X-TARGETDURATION:10\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-INDEPENDENT-SEGMENTS\n"

		for i := 0; i < numSegments; i++ {
			playList += fmt.Sprintf("#EXTINF:%v.00, \n%s/segment?id=%d&index=%d&resolution=%d&ratio=%s\n", segmentDuration, fmt.Sprintf("%s://%s", config.HLS_HOST_PROTOCOL, config.HLS_HOST_NAME), episodeID, i, res, aspectRatio)
		}
		playList += "#EXT-X-ENDLIST\n"

		// Save the playlist to cache/playlist.m3u8
		playlistPath := filepath.Join(cacheDir, fmt.Sprintf("playlist_%d.m3u8", res))
		if err := os.WriteFile(playlistPath, []byte(playList), 0644); err != nil {
			return fmt.Errorf("failed to write playlist file: %v", err)
		}
		metaPlaylist += fmt.Sprintf("%s://%s", config.HLS_HOST_PROTOCOL, config.HLS_HOST_NAME) + fmt.Sprintf("playlist_%d.m3u8", res) + "\n"

		log.Printf("Playlist saved to: %s", playlistPath)
	}

	metaPlaylistPath := filepath.Join(cacheDir, fmt.Sprintf("playlist.m3u8"))
	if err := os.WriteFile(metaPlaylistPath, []byte(metaPlaylist), 0644); err != nil {
		return fmt.Errorf("failed to write playlist file: %v", err)
	}

	// 保存当前的 episodeID
	if err := saveCurrentEpisodeID(cacheDir, episodeID); err != nil {
		return fmt.Errorf("failed to save current episode ID: %v", err)
	}

	return nil
}

/*
ffmpeg -ss 00:10:00 -i yurucamp_06.mkv -t 00:05:00 -c:v libx264 -c:a aac -threads 4 -preset ultrafast -f hls -hls_time 10 -hls_playlist_type vod -hls_segment_filename "output%03d.ts"
*/

func GenerateHlsSegment(inputFile string, startIndex, resolution, segmentDuration int, ratio string, w io.Writer) error {
	log.Println("startIndex", startIndex)
	// Create cache directory if it doesn't exist
	cacheDir := config.HLS_CACHE_PATH
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Check if the segment already exists in cache
	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("segment_%05d_%d.ts", startIndex, resolution))
	if _, err := os.Stat(cacheFile); err == nil {
		log.Println("cache file found", startIndex)
		// Segment exists in cache, read and return it
		cachedData, err := os.ReadFile(cacheFile)
		if err != nil {
			return fmt.Errorf("failed to read cached segment: %v", err)
		}
		_, err = w.Write(cachedData)
		return err
	}

	dimensionRatio := "16:9"
	if ratio == "4:3" {
		dimensionRatio = "4:3"
	}

	var outputWidth int
	if dimensionRatio == "16:9" {
		outputWidth = resolution * 16 / 9
	} else if dimensionRatio == "4:3" {
		outputWidth = resolution * 4 / 3
	}

	if outputWidth%2 != 0 {
		outputWidth += 1
	}

	// Segment doesn't exist, generate it
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%v.00", startIndex*segmentDuration),
		"-i", inputFile,
		"-t", fmt.Sprintf("%v.00", segmentDuration),
		"-async", "1",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-ac", "2",
		"-threads", "4",
		"-pix_fmt", "yuv420p",
		"-force_key_frames", "expr:gte(t,n_forced*5.000)",
		"-preset", "ultrafast",
		"-vf", fmt.Sprintf("scale=%d:%d", outputWidth, resolution),
		"-f", "ssegment",
		"-segment_time", fmt.Sprintf("%v.00", segmentDuration),
		"-initial_offset", fmt.Sprintf("%v.00", startIndex*segmentDuration),
		"pipe:out%05d.ts")

	log.Println("cmd", cmd.String())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("FFmpeg error output:\n%s", stderr.String())
		return fmt.Errorf("command failed: %v", err)
	}

	// Save the segment to cache
	if err := os.WriteFile(cacheFile, stdout.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write segment to cache: %v", err)
	}

	// Write the segment to the provided writer
	_, err := w.Write(stdout.Bytes())
	return err
}

func FormatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remainingSeconds)
}

func getVideoDimensions(inputFile string) (int, int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		inputFile)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get video dimensions: %v", err)
	}

	var width, height int
	_, err = fmt.Sscanf(string(output), "%d,%d", &width, &height)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse video dimensions: %v", err)
	}

	return width, height, nil
}

func GenerateMultipleThumbnails(inputFile, outputPattern string, interval string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vf", fmt.Sprintf("fps=1/%s", interval),
		outputPattern)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// 新增函数：清理缓存
func cleanCache(cacheDir string, currentEpisodeID int) error {
	// 读取上一次的 episodeID
	lastEpisodeID, err := readLastEpisodeID(cacheDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// 如果 episodeID 不匹配，删除所有缓存文件
	if lastEpisodeID != currentEpisodeID {
		err := os.RemoveAll(cacheDir)
		if err != nil {
			return fmt.Errorf("failed to remove cache directory: %v", err)
		}
		// 重新创建缓存目录
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %v", err)
		}
	}

	return nil
}

// 新增函数：读取上一次的 episodeID
func readLastEpisodeID(cacheDir string) (int, error) {
	data, err := os.ReadFile(filepath.Join(cacheDir, "current_episode.txt"))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// 新增函数：保存当前的 episodeID
func saveCurrentEpisodeID(cacheDir string, episodeID int) error {
	return os.WriteFile(filepath.Join(cacheDir, "current_episode.txt"), []byte(strconv.Itoa(episodeID)), 0644)
}
