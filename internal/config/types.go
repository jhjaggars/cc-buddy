package config

import "time"

// Environment represents a development environment with its associated resources
type Environment struct {
	Name          string    `json:"name"`
	Branch        string    `json:"branch"`
	WorktreePath  string    `json:"worktree_path"`
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	VolumeName    string    `json:"volume_name"`
	Created       time.Time `json:"created"`
	Status        string    `json:"status"`
}

// Config holds user configuration settings
type Config struct {
	WorktreeDir   string `json:"worktree_dir"`
	Runtime       string `json:"runtime"`       // "docker" or "podman"
	Containerfile string `json:"containerfile"` // path to containerfile
	ExposeAll     bool   `json:"expose_all"`    // expose all container ports
}

// State represents the persistent application state
type State struct {
	Environments []Environment `json:"environments"`
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		WorktreeDir:   ".worktrees",
		Runtime:       "auto", // auto-detect
		Containerfile: "Containerfile.dev",
		ExposeAll:     false,
	}
}