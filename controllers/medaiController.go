//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/duckfeather10086/dandan-prime/config"
	bangumiusecase "github.com/duckfeather10086/dandan-prime/usecase/bangumiUseCase"
	episodeusecase "github.com/duckfeather10086/dandan-prime/usecase/episodeUseCase"
	matchusecase "github.com/duckfeather10086/dandan-prime/usecase/matchUseCase"
	"github.com/labstack/echo/v4"
)

func InitMediaLibrary(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func SetHlsEnable(c echo.Context) error {
	hlsEnabled, err := strconv.ParseBool(c.QueryParam("enable"))

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid enable parameter"})
	}

	config.SetHlsEnabled(hlsEnabled)

	return nil
}

func GetHlsEnable(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"hls_enabled": config.HLS_ENABLE,
	})
}

func UpdateMediaLibrary(c echo.Context) error {
	forceUpdate := c.QueryParam("force_update")
	log.Println("forceUpdate", forceUpdate)
	if err := episodeusecase.ScanAndSaveMedia(config.MEDIA_LIBRARY_ROOT_PATH); err != nil {
		log.Printf("Error scanning and matching media: %v", err)
	}

	if err := episodeusecase.ScanAndMatchMedia(config.MEDIA_LIBRARY_ROOT_PATH, forceUpdate == "true"); err != nil {
		log.Printf("Error scanning and matching media: %v", err)
	}

	err := bangumiusecase.InitializeBangumiInfo()
	if err != nil {
		log.Printf("Error scanning and matching media: %v", err)
	}

	err = episodeusecase.ScanAndMatchSubtitles()
	if err != nil {
		log.Printf("Error scanning and matching media: %v", err)
	}

	return nil
}

// ResolveBangumiInfo matches library directories to bangumi.tv subjects
// directly, without the dandanplay hash-match detour. It is exposed
// separately from UpdateMediaLibrary so a scan and a re-match can be run
// independently while the two matchers coexist.
func ResolveBangumiInfo(c echo.Context) error {
	force := c.QueryParam("force") == "true"
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	stats, err := matchusecase.ResolveMediaLibrary(force, limit)
	if err != nil {
		log.Printf("Error resolving bangumi info: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, stats)
}
