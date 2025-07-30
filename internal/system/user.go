package system

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

// UserInfo represents user identification information
type UserInfo struct {
	UID int
	GID int
}

// GetCurrentUser returns the current user's UID and GID
func GetCurrentUser() (*UserInfo, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	
	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UID: %w", err)
	}
	
	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GID: %w", err)
	}
	
	return &UserInfo{
		UID: uid,
		GID: gid,
	}, nil
}

// GetCurrentUserFromEnv returns user info from environment variables (fallback)
func GetCurrentUserFromEnv() *UserInfo {
	uid := 1000 // Default fallback
	gid := 1000 // Default fallback
	
	if uidStr := os.Getenv("UID"); uidStr != "" {
		if parsedUID, err := strconv.Atoi(uidStr); err == nil {
			uid = parsedUID
		}
	}
	
	if gidStr := os.Getenv("GID"); gidStr != "" {
		if parsedGID, err := strconv.Atoi(gidStr); err == nil {
			gid = parsedGID
		}
	}
	
	return &UserInfo{
		UID: uid,
		GID: gid,
	}
}

// GetUserInfoWithFallback attempts to get current user info with environment fallback
func GetUserInfoWithFallback() *UserInfo {
	if userInfo, err := GetCurrentUser(); err == nil {
		return userInfo
	}
	
	// Fallback to environment variables
	return GetCurrentUserFromEnv()
}