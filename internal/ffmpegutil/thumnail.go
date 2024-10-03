package ffmpegutil

import (
	"fmt"
	"os"
	"os/exec"
)

func GenerateThumbnail(inputFile, outputFile string, timeOffset string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-ss", timeOffset,
		"-vframes", "1",
		outputFile)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
