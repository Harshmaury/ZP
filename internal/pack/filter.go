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
	FilterGo                         // *.go files only
	FilterYAML                       // *.yaml, *.yml files only
	FilterAPI                        // handler + server + middleware
	FilterCore                       // non-API, non-cmd Go files
)

// FilterName returns the human-readable filter name for ZIP naming.
func FilterName(f FilterMode) string {
	switch f {
	case FilterHandlers:
		return "handlers"
	case FilterGo:
		return "go"
	case FilterYAML:
		return "yaml"
	case FilterAPI:
		return "api"
	case FilterCore:
		return "core"
	default:
		return "full"
	}
}

// defaultExcludes are always applied regardless of filter mode.
var defaultExcludes = []string{
	".git", "vendor", "node_modules",
	".DS_Store", "*.exe", "*.dll", "*.so",
	"*.tmp", "*.log",
}

// Collector walks a project directory and returns matching file paths.
type Collector struct {
	root    string
	mode    FilterMode
	ignores []string
}

// NewCollector creates a Collector for the given project root and filter.
func NewCollector(root string, mode FilterMode, extraIgnores []string) *Collector {
	ignores := append(defaultExcludes, extraIgnores...)
	return &Collector{root: root, mode: mode, ignores: ignores}
}

// Collect returns all matching file paths relative to root.
func (c *Collector) Collect() ([]string, error) {
	var files []string
	err := filepath.WalkDir(c.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(c.root, path)
		if err != nil {
			return err
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
		if c.matches(rel) {
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

// isExcluded returns true if the path matches any exclusion pattern.
func (c *Collector) isExcluded(rel string, isDir bool) bool {
	base := filepath.Base(rel)
	for _, pattern := range c.ignores {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		if strings.Contains(rel, pattern) && isDir {
			return true
		}
	}
	return false
}

// matches returns true if the file passes the active filter.
func (c *Collector) matches(rel string) bool {
	switch c.mode {
	case FilterHandlers:
		return strings.Contains(rel, filepath.Join("internal", "api", "handler")) &&
			strings.HasSuffix(rel, ".go")

	case FilterGo:
		return strings.HasSuffix(rel, ".go")

	case FilterYAML:
		return strings.HasSuffix(rel, ".yaml") || strings.HasSuffix(rel, ".yml")

	case FilterAPI:
		inAPI := strings.Contains(rel, filepath.Join("internal", "api"))
		return inAPI && strings.HasSuffix(rel, ".go")

	case FilterCore:
		inInternal := strings.HasPrefix(rel, "internal")
		inAPI := strings.Contains(rel, filepath.Join("internal", "api"))
		inCmd := strings.HasPrefix(rel, "cmd")
		return strings.HasSuffix(rel, ".go") && inInternal && !inAPI && !inCmd

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
