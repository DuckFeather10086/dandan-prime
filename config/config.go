//go:build !js && !wasm
// +build !js,!wasm

package config

const (
	MEDIA_LIBRARY_ROOT_PATH = "/WDBLUE_1"
	PORT                    = 1234
)

var DEFAULT_ALLOWED_EXTENSIONS = []string{".mkv", ".mp4"}
var HLS_ENABLE = false

func SetHlsEnabled(enabled bool) {
	HLS_ENABLE = enabled
}
