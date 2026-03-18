// @zp-project: zp
// @zp-path: internal/pack/filter.go
// Package pack implements file collection and ZIP creation for zp.
package pack

import (
	"os"
	"path/filepath"
	"strings"
)

// FilterMode controls which files are included in the package.
type FilterMode int

const (
	FilterFull     FilterMode = iota // all non-excluded files
	FilterHandlers                   // internal/api/handler/ only
	FilterGo                         // *.go files (+ go.mod, go.sum always)
	FilterYAML                       // *.yaml, *.yml files only
	FilterAPI                        // handler + server + middleware
	FilterCore                       // internal/ non-API Go + pkg/ layer
	FilterPkg                        // pkg/ layer only
	FilterStore                      // internal/store/ only
	FilterConfig                     // internal/config/ + *.yaml
)

// FilterName returns the human-readable filter name used in ZIP naming.
func FilterName(f FilterMode) string {
	names := map[FilterMode]string{
		FilterHandlers: "handlers",
		FilterGo:       "go",
		FilterYAML:     "yaml",
		FilterAPI:      "api",
		FilterCore:     "core",
		FilterPkg:      "pkg",
		FilterStore:    "store",
		FilterConfig:   "config",
	}
	if n, ok := names[f]; ok {
		return n
	}
	return "full"
}

// ParseFilter converts a flag string to a FilterMode.
// Returns FilterFull and false if unrecognised.
func ParseFilter(flag string) (FilterMode, bool) {
	filters := map[string]FilterMode{
		"-H":       FilterHandlers,
		"-go":      FilterGo,
		"-yaml":    FilterYAML,
		"-api":     FilterAPI,
		"-core":    FilterCore,
		"-pkg":     FilterPkg,
		"-store":   FilterStore,
		"-config":  FilterConfig,
	}
	f, ok := filters[flag]
	return f, ok
}

// alwaysIncluded are files always included regardless of filter mode.
var alwaysIncluded = map[string]bool{
	"go.mod":      true,
	"go.sum":      true,
	"nexus.yaml":  true,
	".zpignore":   true,
}

// defaultExcludes are always skipped.
// Dirs starting with _ or . are also skipped (handled in isExcluded).
var defaultExcludes = []string{
	".git", "vendor", "node_modules",
	".DS_Store", "*.exe", "*.dll", "*.so",
	"*.tmp", "*.log", "dist", "build",
}

// Collector walks a project directory and returns matching file paths.
type Collector struct {
	root    string
	mode    FilterMode
	ignores []string
}

// NewCollector creates a Collector for the given project root and filter.
func NewCollector(root string, mode FilterMode, extraIgnores []string) *Collector {
	ignores := make([]string, len(defaultExcludes))
	copy(ignores, defaultExcludes)
	ignores = append(ignores, extraIgnores...)
	return &Collector{root: root, mode: mode, ignores: ignores}
}

// Collect returns all matching file paths relative to root.
func (c *Collector) Collect() ([]string, error) {
	var files []string
	err := filepath.WalkDir(c.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		rel, err := filepath.Rel(c.root, path)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		if c.isExcluded(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(rel)
		if alwaysIncluded[base] || c.matches(rel) {
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

// isExcluded returns true if the path should be skipped.
func (c *Collector) isExcluded(rel string, isDir bool) bool {
	base := filepath.Base(rel)

	// Skip dirs/files starting with _ or . (backups, hidden)
	if len(base) > 0 && (base[0] == '_' || base[0] == '.') {
		return true
	}

	for _, pattern := range c.ignores {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		// For directory segments — check if any part of the path matches.
		if isDir && strings.EqualFold(base, pattern) {
			return true
		}
	}
	return false
}

// matches returns true if the file passes the active filter.
func (c *Collector) matches(rel string) bool {
	sep := string(filepath.Separator)

	switch c.mode {
	case FilterHandlers:
		return strings.Contains(rel, filepath.Join("internal", "api", "handler")) &&
			strings.HasSuffix(rel, ".go")

	case FilterGo:
		return strings.HasSuffix(rel, ".go")

	case FilterYAML:
		return strings.HasSuffix(rel, ".yaml") ||
			strings.HasSuffix(rel, ".yml")

	case FilterAPI:
		return strings.Contains(rel, sep+"api"+sep) &&
			strings.HasSuffix(rel, ".go")

	case FilterCore:
		inInternal := strings.HasPrefix(rel, "internal")
		inPkg := strings.HasPrefix(rel, "pkg")
		inAPI := strings.Contains(rel, sep+"api"+sep)
		inCmd := strings.HasPrefix(rel, "cmd")
		isGo := strings.HasSuffix(rel, ".go")
		return isGo && (inInternal || inPkg) && !inAPI && !inCmd

	case FilterPkg:
		return strings.HasPrefix(rel, "pkg") &&
			strings.HasSuffix(rel, ".go")

	case FilterStore:
		return strings.Contains(rel, sep+"store"+sep) &&
			strings.HasSuffix(rel, ".go")

	case FilterConfig:
		isConfig := strings.Contains(rel, sep+"config"+sep) &&
			strings.HasSuffix(rel, ".go")
		isYAML := strings.HasSuffix(rel, ".yaml") ||
			strings.HasSuffix(rel, ".yml")
		return isConfig || isYAML

	default: // FilterFull
		return true
	}
}

// LoadZPIgnore reads .zpignore from root and returns ignore patterns.
func LoadZPIgnore(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, ".zpignore"))
	if err != nil {
		return nil
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns
}
