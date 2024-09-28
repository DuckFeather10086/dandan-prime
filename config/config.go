//go:build !js && !wasm
// +build !js,!wasm

package config

const (
	DefaultMediaLibraryPath = "/WDBLUE_1"
)

var DefaultAllowedExtensions = []string{".mkv", ".mp4"}

// func GetConfig() *Config {
// 	return &Config{
// 		MediaLibraryPath:  os.Getenv("MEDIA_LIBRARY_PATH"),
// 		AllowedExtensions: []string{".mkv", ".mp4"},
// 	}
// }
