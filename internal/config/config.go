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
		// Default to engx-drop on the Windows side if running in WSL.
		winHome := detectWindowsHome()
		if winHome != "" {
			dropDir = filepath.Join(winHome, "Downloads", "engx-drop")
		} else {
			dropDir = filepath.Join(home, "Downloads", "engx-drop")
		}
	}

	return &Config{
		WorkspaceRoot: workspaceRoot,
		DropDir:       dropDir,
	}
}

// detectWindowsHome returns the Windows user home via /mnt/c/Users/<user>
// when running inside WSL2. Returns empty string if not in WSL.
func detectWindowsHome() string {
	// Check if we are in WSL by looking for /mnt/c/Users/
	mountBase := "/mnt/c/Users"
	entries, err := os.ReadDir(mountBase)
	if err != nil {
		return ""
	}
	// Find the first non-system user directory.
	skip := map[string]bool{
		"Public": true, "Default": true, "All Users": true,
		"Default User": true, "desktop.ini": true,
	}
	for _, e := range entries {
		if e.IsDir() && !skip[e.Name()] {
			candidate := filepath.Join(mountBase, e.Name())
			// Verify it looks like a home dir.
			if _, err := os.Stat(filepath.Join(candidate, "Downloads")); err == nil {
				return candidate
			}
		}
	}
	return ""
}
