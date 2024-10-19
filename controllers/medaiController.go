//go:build !js && !wasm
// +build !js,!wasm

package controllers

import (
	"net/http"
	"strconv"

	"github.com/duckfeather10086/dandan-prime/config"
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
