// @zp-project: zp
// @zp-path: internal/manifest/manifest.go
// Package manifest reads nexus.yaml and resolves project metadata.
package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest is the parsed nexus.yaml descriptor.
type Manifest struct {
	Name      string   `yaml:"name"`
	ID        string   `yaml:"id"`
	Type      string   `yaml:"type"`
	Language  string   `yaml:"language"`
	Version   string   `yaml:"version"`
	DependsOn []string `yaml:"depends_on"`
	RootDir   string   `yaml:"-"` // populated after load
}

// Load reads nexus.yaml from dir and returns a Manifest.
func Load(dir string) (*Manifest, error) {
	path := filepath.Join(dir, "nexus.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read nexus.yaml in %s: %w", dir, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse nexus.yaml: %w", err)
	}
	if m.ID == "" {
		return nil, fmt.Errorf("nexus.yaml missing required field: id")
	}
	m.RootDir = dir
	return &m, nil
}

// LoadFromID searches workspaceRoot for a project with the given ID.
func LoadFromID(workspaceRoot, id string) (*Manifest, error) {
	// Search common platform project paths.
	searchDirs := []string{
		filepath.Join(workspaceRoot, "projects", "apps", id),
		filepath.Join(workspaceRoot, "projects", "tools", id),
		filepath.Join(workspaceRoot, id),
	}
	for _, dir := range searchDirs {
		if _, err := os.Stat(filepath.Join(dir, "nexus.yaml")); err == nil {
			return Load(dir)
		}
	}
	return nil, fmt.Errorf("project %q not found under %s", id, workspaceRoot)
}
