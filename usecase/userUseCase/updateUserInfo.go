//go:build !js && !wasm
// +build !js,!wasm

package userusecase

import "github.com/duckfeather10086/dandan-prime/database"

func GetUserInfoByUserId(userId uint) (*database.UserInfo, error) {
	// Implement this function to fetch user info from database
	userInfo, err := database.GetUserInfoByUserId(uint(userId))

	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func UpdateUserInfo(userID uint, userinfo *database.UserInfo) error {
	return database.UpdateUserInfoByUserId(userID, userinfo)
}
