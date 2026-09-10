//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/duckfeather10086/dandan-prime/config"
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

	// dandanplay is now consulted only for the episode ids danmaku needs, and
	// only if credentials are configured. It is no longer how the library
	// learns what a file is, so a failure here costs danmaku, not metadata.
	if config.DANDANPLAY_API_APP_ID != "" && config.DANDANPLAY_API_APP_SECRET != "" {
		if err := episodeusecase.ScanAndMatchMedia(config.MEDIA_LIBRARY_ROOT_PATH, forceUpdate == "true"); err != nil {
			log.Printf("Error matching episodes against dandanplay: %v", err)
		}
	} else {
		log.Println("dandanplay credentials not configured; skipping danmaku id matching")
	}

	// Cheap and offline, so it runs before resolution: a parser improvement
	// gets applied without spending the LLM budget on a re-resolve.
	if _, err := matchusecase.BackfillEpisodeNumbers(forceUpdate == "true"); err != nil {
		log.Printf("Error backfilling episode numbers: %v", err)
	}

	if _, err := matchusecase.ResolveMediaLibrary(forceUpdate == "true", 0); err != nil {
		log.Printf("Error resolving bangumi subjects: %v", err)
	}

	if err := episodeusecase.PopulateEpisodeMetadata(forceUpdate == "true"); err != nil {
		log.Printf("Error populating episode metadata: %v", err)
	}

	if err := episodeusecase.ScanAndMatchSubtitles(); err != nil {
		log.Printf("Error matching subtitles: %v", err)
	}

	return nil
}

// ResolveBangumiInfo matches library directories to bangumi.tv subjects
// directly, without the dandanplay hash-match detour. It is exposed
// separately from UpdateMediaLibrary so a scan and a re-match can be run
// independently while the two matchers coexist.
// BackfillEpisodeNumbers re-derives episode numbers from file names. Separate
// from resolution because it needs neither the network nor the model.
func BackfillEpisodeNumbers(c echo.Context) error {
	force := c.QueryParam("force") == "true"

	updated, err := matchusecase.BackfillEpisodeNumbers(force)
	if err != nil {
		log.Printf("Error backfilling episode numbers: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]int{"updated": updated})
}

// PopulateEpisodeMetadata fills in per-episode titles and summaries from
// bangumi.tv for every subject in the library.
func PopulateEpisodeMetadata(c echo.Context) error {
	force := c.QueryParam("force") == "true"

	if err := episodeusecase.PopulateEpisodeMetadata(force); err != nil {
		log.Printf("Error populating episode metadata: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

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
