//go:build !js && !wasm
// +build !js,!wasm

package config

import (
	"encoding/json"
	"os"
)

const ()

var DEFAULT_ALLOWED_EXTENSIONS []string
var HLS_ENABLE bool
var MEDIA_LIBRARY_ROOT_PATH string
var PORT int
var HLS_HOST_PROTOCOL string
var HLS_HOST_NAME string
var HLS_CACHE_PATH string
var DANDANPLAY_API_APP_ID string
var DANDANPLAY_API_APP_SECRET string

type Config struct {
	MediaLibraryRootPath   string   `json:"media_library_root_path"`
	AllowedVideoExtensions []string `json:"allowed_video_extensions"`
	UseHLS                 bool     `json:"use_hls"`
	HLSCachePath           string   `json:"hls_cache_path"`
	Port                   int      `json:"port"`
	HLSHostName            string   `json:"hls_host_name"`
	HLSHostProtocol        string   `json:"hls_host_protocol"`
	DandanPlayAppID        string   `json:"dandan_play_app_id"`
	DandanPlayAppSecret    string   `json:"dandan_play_app_secret"`
}

func InitConfig() error {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	var config Config

	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	DEFAULT_ALLOWED_EXTENSIONS = config.AllowedVideoExtensions
	MEDIA_LIBRARY_ROOT_PATH = config.MediaLibraryRootPath
	PORT = config.Port
	HLS_ENABLE = config.UseHLS

	HLS_HOST_NAME = config.HLSHostName
	HLS_HOST_PROTOCOL = config.HLSHostProtocol
	HLS_CACHE_PATH = config.HLSCachePath

	DANDANPLAY_API_APP_ID = config.DandanPlayAppID
	DANDANPLAY_API_APP_SECRET = config.DandanPlayAppSecret

	return nil
}

func SetHlsEnabled(enabled bool) error {
	HLS_ENABLE = enabled

	file, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	var config Config

	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	config.UseHLS = enabled

	data, err := json.MarshalIndent(config, "", "  ")

	err = os.WriteFile("cmd/config.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}
