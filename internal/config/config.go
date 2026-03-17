// @zp-project: zp
// @zp-path: internal/config/config.go
package config

import (
	"os"
	"path/filepath"
)

// Config holds resolved zp configuration.
type Config struct {
	WorkspaceRoot string
	DropDir       string
}

// Load resolves zp configuration from environment and defaults.
func Load() *Config {
	home, _ := os.UserHomeDir()

	workspaceRoot := os.Getenv("ZP_WORKSPACE")
	if workspaceRoot == "" {
		workspaceRoot = filepath.Join(home, "workspace")
	}

	dropDir := os.Getenv("ZP_DROP_DIR")
	if dropDir == "" {
		dropDir = filepath.Join(home, "Downloads", "nexus-drop")
	}

	return &Config{
		WorkspaceRoot: workspaceRoot,
		DropDir:       dropDir,
	}
}
