// @zp-project: zp
// @zp-path: cmd/zp/main.go
// zp v2.0.0 — developer packaging tool for the engx platform.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Harshmaury/ZP/internal/config"
	"github.com/Harshmaury/ZP/internal/manifest"
	"github.com/Harshmaury/ZP/internal/pack"
	"github.com/Harshmaury/ZP/internal/registry"
)

const version = "2.0.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "\nzp error: %v\n\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg := config.Load()
	opts, positional, err := parseArgs(args)
	if err != nil {
		return err
	}

	// --out flag overrides drop dir for this run.
	if opts.outDir != "" {
		cfg.DropDir = opts.outDir
	}

	if len(positional) == 0 {
		return runCurrent(cfg, opts)
	}

	switch positional[0] {
	case "help", "--help", "-h":
		printHelp(cfg)
		return nil
	case "version", "--version":
		fmt.Printf("zp v%s\n", version)
		return nil
	case "list", "ls":
		return runList(cfg)
	case "status", "st":
		return runStatus(cfg)
	case "all":
		return runAll(cfg, opts)
	case "dev":
		if len(positional) < 2 {
			return fmt.Errorf("usage: zp dev <project-id>")
		}
		return runDev(cfg, positional[1])
	default:
		return runProjects(cfg, positional, opts)
	}
}

// ── OPTIONS ───────────────────────────────────────────────────────────────────

type options struct {
	filter pack.FilterMode
	outDir string   // --out <path>
	path   string   // --path <dir> (explicit project root, no nexus.yaml needed)
}

func parseArgs(args []string) (options, []string, error) {
	opts := options{filter: pack.FilterFull}
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--out" || arg == "-o":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--out requires a path argument")
			}
			i++
			opts.outDir = args[i]
		case arg == "--path" || arg == "-p":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--path requires a directory argument")
			}
			i++
			opts.path = args[i]
		case strings.HasPrefix(arg, "--out="):
			opts.outDir = strings.TrimPrefix(arg, "--out=")
		case strings.HasPrefix(arg, "--path="):
			opts.path = strings.TrimPrefix(arg, "--path=")
		default:
			if f, ok := pack.ParseFilter(arg); ok {
				opts.filter = f
			} else {
				positional = append(positional, arg)
			}
		}
	}
	return opts, positional, nil
}

// ── COMMANDS ──────────────────────────────────────────────────────────────────

// runCurrent packages the project in the current working directory.
func runCurrent(cfg *config.Config, opts options) error {
	// --path flag bypasses nexus.yaml requirement.
	if opts.path != "" {
		abs, err := filepath.Abs(opts.path)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		id := filepath.Base(abs)
		return packageOne(abs, id, opts.filter, cfg.DropDir)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	m, err := manifest.Load(cwd)
	if err != nil {
		return fmt.Errorf(
			"no nexus.yaml in current directory\n  specify a project: zp <id>\n  or use: zp --path <dir>")
	}
	return packageOne(m.RootDir, m.ID, opts.filter, cfg.DropDir)
}

// runProjects packages one or more named projects.
func runProjects(cfg *config.Config, ids []string, opts options) error {
	if len(ids) == 1 {
		root, id, err := resolveProject(cfg, ids[0])
		if err != nil {
			return err
		}
		return packageOne(root, id, opts.filter, cfg.DropDir)
	}

	// Multiple projects — combined ZIP.
	var projects []struct{ Root, ID string }
	for _, id := range ids {
		root, resolvedID, err := resolveProject(cfg, id)
		if err != nil {
			return fmt.Errorf("project %q: %w", id, err)
		}
		projects = append(projects, struct{ Root, ID string }{root, resolvedID})
	}

	result, err := pack.BuildMultiZIP(projects, opts.filter, cfg.DropDir)
	if err != nil {
		return err
	}
	printResult(result)
	return nil
}

// runAll packages all projects discovered in the workspace.
func runAll(cfg *config.Config, opts options) error {
	entries, err := registry.Scan(cfg.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("scan workspace: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no projects with nexus.yaml found under %s", cfg.WorkspaceRoot)
	}

	fmt.Printf("zp: packaging %d projects...\n\n", len(entries))
	ok, skipped := 0, 0
	for _, e := range entries {
		result, err := pack.BuildZIP(e.RootDir, e.ID, opts.filter, cfg.DropDir)
		if err != nil {
			fmt.Printf("  %-20s error: %v\n", e.ID, err)
			skipped++
			continue
		}
		fmt.Printf("  %-20s %4d files → %s\n", e.ID, result.FileCount, filepath.Base(result.ZipPath))
		ok++
	}
	fmt.Printf("\n%d packaged, %d skipped → %s\n", ok, skipped, cfg.DropDir)
	return nil
}

// runList lists all discovered projects in the workspace.
func runList(cfg *config.Config) error {
	entries, err := registry.Scan(cfg.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("scan workspace: %w", err)
	}
	if len(entries) == 0 {
		fmt.Printf("no projects found under %s\n", cfg.WorkspaceRoot)
		return nil
	}

	fmt.Printf("\nProjects in %s\n\n", cfg.WorkspaceRoot)
	fmt.Printf("  %-22s %-16s %s\n", "ID", "TYPE", "PATH")
	fmt.Printf("  %s\n", strings.Repeat("─", 72))
	for _, e := range entries {
		rel, _ := filepath.Rel(cfg.WorkspaceRoot, e.RootDir)
		fmt.Printf("  %-22s %-16s %s\n", e.ID, e.Type, rel)
	}
	fmt.Printf("\n%d project(s) found\n\n", len(entries))
	return nil
}

// runStatus shows all projects and their last ZIP timestamp.
func runStatus(cfg *config.Config) error {
	entries, err := registry.Scan(cfg.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("scan workspace: %w", err)
	}

	// Find last ZIP per project in drop dir.
	lastZIP := map[string]time.Time{}
	if zipFiles, err := filepath.Glob(filepath.Join(cfg.DropDir, "*.zip")); err == nil {
		for _, zf := range zipFiles {
			base := filepath.Base(zf)
			// ZIP names: <id>-<filter>-<YYYYMMDD>-<HHMM>.zip
			// Match project ID prefix.
			for _, e := range entries {
				if strings.HasPrefix(base, e.ID+"-") {
					info, err := os.Stat(zf)
					if err == nil {
						if info.ModTime().After(lastZIP[e.ID]) {
							lastZIP[e.ID] = info.ModTime()
						}
					}
				}
			}
		}
	}

	fmt.Printf("\nzp status — workspace: %s\n", cfg.WorkspaceRoot)
	fmt.Printf("           drop dir:  %s\n\n", cfg.DropDir)
	fmt.Printf("  %-22s %-16s %s\n", "ID", "TYPE", "LAST ZIP")
	fmt.Printf("  %s\n", strings.Repeat("─", 64))

	for _, e := range entries {
		last := "never"
		if t, ok := lastZIP[e.ID]; ok {
			last = t.Format("2006-01-02 15:04")
		}
		fmt.Printf("  %-22s %-16s %s\n", e.ID, e.Type, last)
	}
	fmt.Printf("\n%d project(s)\n\n", len(entries))
	return nil
}

// runDev creates an isolated development sandbox.
func runDev(cfg *config.Config, id string) error {
	root, _, err := resolveProject(cfg, id)
	if err != nil {
		return err
	}
	m, err := manifest.Load(root)
	if err != nil {
		return err
	}
	fmt.Printf("zp: creating dev sandbox for %s...\n", id)
	result, err := pack.CreateDevSandbox(m, cfg.WorkspaceRoot)
	if err != nil {
		return err
	}
	fmt.Printf("\n  %-16s %s\n", "sandbox", result.SandboxDir)
	fmt.Printf("  %-16s %d files\n", "copied", result.CopiedFiles)
	if len(result.Contracts) > 0 {
		fmt.Printf("  %-16s %s\n", "contracts", strings.Join(result.Contracts, ", "))
	}
	fmt.Printf("\n  cd %s/%s\n\n", result.SandboxDir, result.ProjectID)
	return nil
}

// ── HELPERS ───────────────────────────────────────────────────────────────────

// resolveProject finds a project root by ID using registry scan.
// Falls back to manifest.LoadFromID for backwards compat.
func resolveProject(cfg *config.Config, id string) (root, resolvedID string, err error) {
	entry, err := registry.Find(cfg.WorkspaceRoot, id)
	if err != nil {
		return "", "", err
	}
	if entry != nil {
		return entry.RootDir, entry.ID, nil
	}
	// Fallback: try manifest direct path resolution.
	m, err := manifest.LoadFromID(cfg.WorkspaceRoot, id)
	if err != nil {
		return "", "", fmt.Errorf("project %q not found — run 'zp list' to see available projects", id)
	}
	return m.RootDir, m.ID, nil
}

func packageOne(root, id string, filter pack.FilterMode, outDir string) error {
	result, err := pack.BuildZIP(root, id, filter, outDir)
	if err != nil {
		return err
	}
	printResult(result)
	return nil
}

func printResult(r *pack.ZipResult) {
	fmt.Printf("\n  %-10s %s\n", "project", r.ProjectID)
	fmt.Printf("  %-10s %s\n", "filter", pack.FilterName(r.FilterMode))
	fmt.Printf("  %-10s %d files\n", "packed", r.FileCount)
	fmt.Printf("  %-10s %s\n\n", "output", r.ZipPath)
}

// printHelp prints clean, structured help.
func printHelp(cfg *config.Config) {
	fmt.Printf(`
zp v%s — developer packaging tool

USAGE
  zp                       package current project (reads nexus.yaml)
  zp <id>                  package project by id
  zp <id> <id> ...         package multiple projects into one zip
  zp all                   package all discovered platform projects
  zp list                  list all discovered projects
  zp status                show all projects + last zip timestamp
  zp dev <id>              create isolated dev sandbox
  zp version               show version
  zp help                  show this help

FILTERS  (combine with any command)
  -H                       handlers only    internal/api/handler/
  -go                      Go source files only
  -yaml                    YAML / config files only
  -api                     full API layer   handler + server + middleware
  -core                    core logic       internal/ + pkg/ (non-API)
  -pkg                     pkg/ layer only
  -store                   internal/store/ only
  -config                  internal/config/ + YAML files

FLAGS
  --out <dir>              override output directory for this run
  --path <dir>             use explicit project root (no nexus.yaml needed)

EXAMPLES
  zp                       package current project
  zp nexus                 package nexus
  zp atlas forge -api      package atlas + forge, API layer only
  zp nexus -H              nexus handlers only
  zp all                   package every project with nexus.yaml
  zp all -go               all projects, Go files only
  zp list                  discover all projects in workspace
  zp status                show projects + last zip date
  zp dev forge             isolated forge sandbox → /tmp/zp-dev/
  zp --path ~/workspace/developer-platform    package by explicit path
  zp nexus --out /tmp      output to /tmp instead of drop dir

CONFIGURATION
  drop dir:    %s
  workspace:   %s

  Override with environment variables:
    export ZP_DROP_DIR=/mnt/c/Users/harsh/Downloads/engx-drop
    export ZP_WORKSPACE=~/workspace

`, version, cfg.DropDir, cfg.WorkspaceRoot)
}
