package config

import "os"

type Config struct {
	MediaLibraryPath  string
	AllowedExtensions []string
}

func GetConfig() *Config {
	return &Config{
		MediaLibraryPath:  os.Getenv("MEDIA_LIBRARY_PATH"),
		AllowedExtensions: []string{".mkv", ".mp4"},
	}
}
