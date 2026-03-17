// @zp-project: zp
// @zp-path: cmd/zp/main.go
// zp — developer packaging tool for the engx platform.
//
// Usage:
//   zp                     package current project
//   zp <id> [id...]        package one or more projects by id
//   zp all                 package all platform projects
//   zp dev <id>            create isolated dev sandbox
//   zp help                show this help
//
// Filters (combine with any command):
//   -H     handlers only
//   -go    Go files only
//   -yaml  YAML/config files only
//   -api   full API layer
//   -core  core logic only
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Harshmaury/ZP/internal/config"
	"github.com/Harshmaury/ZP/internal/manifest"
	"github.com/Harshmaury/ZP/internal/pack"
)

const version = "1.0.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "zp: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg := config.Load()

	// Parse filter flag and remaining args.
	filter, positional := parseArgs(args)

	if len(positional) == 0 {
		return runCurrent(cfg, filter)
	}

	switch positional[0] {
	case "help", "--help", "-h":
		printHelp(cfg)
		return nil
	case "all":
		return runAll(cfg, filter)
	case "dev":
		if len(positional) < 2 {
			return fmt.Errorf("usage: zp dev <project-id>")
		}
		return runDev(cfg, positional[1])
	case "version", "--version":
		fmt.Printf("zp v%s\n", version)
		return nil
	default:
		return runProjects(cfg, positional, filter)
	}
}

// runCurrent packages the project in the current working directory.
func runCurrent(cfg *config.Config, filter pack.FilterMode) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	m, err := manifest.Load(cwd)
	if err != nil {
		return fmt.Errorf("no nexus.yaml found in current directory\n  run from a project root or specify a project id: zp <id>")
	}
	return packageOne(m.RootDir, m.ID, filter, cfg.DropDir)
}

// runProjects packages one or more named projects.
func runProjects(cfg *config.Config, ids []string, filter pack.FilterMode) error {
	if len(ids) == 1 {
		m, err := manifest.LoadFromID(cfg.WorkspaceRoot, ids[0])
		if err != nil {
			return err
		}
		return packageOne(m.RootDir, m.ID, filter, cfg.DropDir)
	}

	// Multiple projects — build combined ZIP.
	var projects []struct{ Root, ID string }
	for _, id := range ids {
		m, err := manifest.LoadFromID(cfg.WorkspaceRoot, id)
		if err != nil {
			return fmt.Errorf("project %q: %w", id, err)
		}
		projects = append(projects, struct{ Root, ID string }{m.RootDir, m.ID})
	}

	result, err := pack.BuildMultiZIP(projects, filter, cfg.DropDir)
	if err != nil {
		return err
	}
	printResult(result)
	return nil
}

// runAll packages every known platform project.
func runAll(cfg *config.Config, filter pack.FilterMode) error {
	platformProjects := []string{
		"nexus", "atlas", "forge",
		"metrics", "navigator", "guardian", "observer", "sentinel",
	}
	fmt.Printf("zp: packaging %d platform projects...\n\n", len(platformProjects))
	ok, skipped := 0, 0
	for _, id := range platformProjects {
		m, err := manifest.LoadFromID(cfg.WorkspaceRoot, id)
		if err != nil {
			fmt.Printf("  %-12s skipped (%v)\n", id, err)
			skipped++
			continue
		}
		result, err := pack.BuildZIP(m.RootDir, m.ID, filter, cfg.DropDir)
		if err != nil {
			fmt.Printf("  %-12s error: %v\n", id, err)
			skipped++
			continue
		}
		fmt.Printf("  %-12s %4d files → %s\n", id, result.FileCount, filepath.Base(result.ZipPath))
		ok++
	}
	fmt.Printf("\n%d packaged, %d skipped → %s\n", ok, skipped, cfg.DropDir)
	return nil
}

// runDev creates an isolated development sandbox.
func runDev(cfg *config.Config, id string) error {
	m, err := manifest.LoadFromID(cfg.WorkspaceRoot, id)
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
		fmt.Printf("  %-16s %v\n", "contracts", result.Contracts)
	}
	fmt.Printf("\nReady:\n  cd %s/%s\n", result.SandboxDir, result.ProjectID)
	return nil
}

// packageOne packages a single project and prints the result.
func packageOne(root, id string, filter pack.FilterMode, outDir string) error {
	result, err := pack.BuildZIP(root, id, filter, outDir)
	if err != nil {
		return err
	}
	printResult(result)
	return nil
}

// printResult prints a clean summary of a zip operation.
func printResult(r *pack.ZipResult) {
	fmt.Printf("\n  %-10s %s\n", "project", r.ProjectID)
	fmt.Printf("  %-10s %s\n", "filter", pack.FilterName(r.FilterMode))
	fmt.Printf("  %-10s %d files\n", "packed", r.FileCount)
	fmt.Printf("  %-10s %s\n\n", "output", r.ZipPath)
}

// parseArgs separates filter flags from positional arguments.
func parseArgs(args []string) (pack.FilterMode, []string) {
	filter := pack.FilterFull
	var positional []string
	for _, arg := range args {
		switch arg {
		case "-H":
			filter = pack.FilterHandlers
		case "-go":
			filter = pack.FilterGo
		case "-yaml":
			filter = pack.FilterYAML
		case "-api":
			filter = pack.FilterAPI
		case "-core":
			filter = pack.FilterCore
		default:
			positional = append(positional, arg)
		}
	}
	return filter, positional
}

// printHelp prints clean, structured help output.
func printHelp(cfg *config.Config) {
	fmt.Printf(`
zp v%s — developer packaging tool

USAGE
  zp                     package current project (reads nexus.yaml)
  zp <id>                package project by id
  zp <id> <id> ...       package multiple projects into one zip
  zp all                 package all platform projects
  zp dev <id>            create isolated dev sandbox
  zp version             show version

FILTERS
  -H                     handlers only   (internal/api/handler/)
  -go                    Go source files only
  -yaml                  YAML / config files only
  -api                   full API layer  (handler + server + middleware)
  -core                  core logic      (non-API, non-cmd)

EXAMPLES
  zp                     package current project, all files
  zp nexus               package nexus project
  zp atlas forge -api    package atlas + forge, API layer only
  zp nexus -H            package nexus handlers only
  zp all -go             package all projects, Go files only
  zp dev forge           create isolated forge sandbox in /tmp/zp-dev/

OUTPUT
  drop dir:   %s
  override:   export ZP_DROP_DIR=<path>

WORKSPACE
  root:       %s
  override:   export ZP_WORKSPACE=<path>

`, version, cfg.DropDir, cfg.WorkspaceRoot)
}
