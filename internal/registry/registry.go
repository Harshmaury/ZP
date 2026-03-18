// @zp-project: zp
// @zp-path: internal/registry/registry.go
// Package registry dynamically discovers all platform projects by scanning
// the workspace for nexus.yaml files. Replaces the hardcoded project list.
package registry

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/Harshmaury/ZP/internal/manifest"
)

// Entry is a discovered project in the workspace.
type Entry struct {
	ID      string
	Name    string
	Type    string
	RootDir string
}

// Scan walks the workspace root and returns all projects with nexus.yaml.
// Search depth is limited to 4 levels to avoid traversing deep trees.
func Scan(workspaceRoot string) ([]*Entry, error) {
	var entries []*Entry
	seen := map[string]bool{}

	err := walkDepth(workspaceRoot, 4, func(path string, d os.DirEntry) error {
		if d.IsDir() || d.Name() != "nexus.yaml" {
			return nil
		}
		dir := filepath.Dir(path)
		if seen[dir] {
			return nil
		}
		seen[dir] = true

		m, err := manifest.Load(dir)
		if err != nil {
			return nil // skip malformed nexus.yaml — don't fail scan
		}
		entries = append(entries, &Entry{
			ID:      m.ID,
			Name:    m.Name,
			Type:    m.Type,
			RootDir: dir,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
	return entries, nil
}

// Find returns the entry with the given ID, or nil if not found.
func Find(workspaceRoot, id string) (*Entry, error) {
	entries, err := Scan(workspaceRoot)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, nil
}

// walkDepth walks dir up to maxDepth levels deep.
func walkDepth(root string, maxDepth int, fn func(path string, d os.DirEntry) error) error {
	return walkDir(root, 0, maxDepth, fn)
}

func walkDir(dir string, depth, maxDepth int, fn func(string, os.DirEntry) error) error {
	if depth > maxDepth {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // skip unreadable dirs
	}
	for _, d := range entries {
		path := filepath.Join(dir, d.Name())
		if err := fn(path, d); err != nil {
			if err == fs.SkipDir {
				continue
			}
			return err
		}
		if d.IsDir() && !shouldSkipDir(d.Name()) {
			if err := walkDir(path, depth+1, maxDepth, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// shouldSkipDir returns true for directories that should never be scanned.
func shouldSkipDir(name string) bool {
	if len(name) > 0 && (name[0] == '.' || name[0] == '_') {
		return true
	}
	skip := map[string]bool{
		"vendor": true, "node_modules": true,
		"testdata": true, "dist": true, "build": true,
	}
	return skip[name]
}
