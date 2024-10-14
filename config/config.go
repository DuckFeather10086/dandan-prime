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

type Config struct {
	MediaLibraryRootPath   string   `json:"media_library_root_path"`
	AllowedVideoExtensions []string `json:"allowed_video_extensions"`
	UseHLS                 bool     `json:"use_hls"`
	HLSCachePath           string   `json:"hls_cache_path"`
	Port                   int      `json:"port"`
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
