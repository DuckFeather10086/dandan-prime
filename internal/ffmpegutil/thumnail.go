//go:build !js && !wasm
// +build !js,!wasm

package ffmpegutil

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

func GenerateHlsCache(inputFile, timeOffset string, segmentDuration, segmentIndex int) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-f", "hls",
		"-threads", "8",
		"-preset", "ultrafast",
		"-hls_time", fmt.Sprintf("%d", segmentDuration),
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", fmt.Sprintf("output%05d.ts", segmentIndex),
		"output.m3u8")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("cmd", cmd.String())

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// func ExtractFrameToMemory(inputFile, timeStamp, size, outputFileName string) ([]byte, error) {
// 	// Prepare the ffmpeg command
// 	cmd := exec.Command("ffmpeg", "-i", inputFile, "-ss", timeStamp, "-s", size, "-frames:v", "1", outputFileName)
// 	println(cmd.String())

// 	// Execute the ffmpeg command
// 	if err := cmd.Run(); err != nil {
// 		return nil, fmt.Errorf("failed to execute ffmpeg command: %v", err)
// 	}

// 	// Read the generated image file into memory
// 	imageData, err := ioutil.ReadFile(outputPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read the generated image: %v", err)
// 	}

// 	return imageData, nil
// }

func GenerateMultipleThumbnails(inputFile, outputPattern string, interval string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vf", fmt.Sprintf("fps=1/%s", interval),
		outputPattern)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
