package sacnner

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var allowedExtensions = []string{".mkv", ".mp4"}

// ScanMediaFiles scans the given directory for media files and prints their hashes
func ScanMediaFiles(rootPath string) error {
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
				fmt.Printf("File: %s\nHash: %s\n\n", path, hash)
				break
			}
		}
		return nil
	})
}

// ScanAndPrintHashes scans the given directory and prints file hashes
func ScanAndPrintHashes(rootPath string) {
	fmt.Printf("Scanning directory: %s\n\n", rootPath)
	err := ScanMediaFiles(rootPath)
	if err != nil {
		fmt.Printf("Error scanning files: %v\n", err)
	}
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
