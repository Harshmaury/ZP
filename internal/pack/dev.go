// @zp-project: zp
// @zp-path: internal/pack/dev.go
// Dev isolation mode — creates a clean sandbox for focused development.
package pack

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Harshmaury/ZP/internal/manifest"
	"github.com/Harshmaury/ZP/internal/registry"
)

// DevResult is the outcome of a dev isolation operation.
type DevResult struct {
	SandboxDir string
	ProjectID  string
	CopiedFiles int
	Contracts  []string // depends_on project IDs whose contracts were included
}

// CreateDevSandbox creates an isolated development environment in /tmp/zp-dev/.
// It copies the full project and the nexus.yaml of each depends_on project.
func CreateDevSandbox(m *manifest.Manifest, workspaceRoot string) (*DevResult, error) {
	ts := time.Now().Format("20060102-1504")
	sandboxDir := filepath.Join(os.TempDir(), "zp-dev", fmt.Sprintf("%s-%s", m.ID, ts))

	projectDest := filepath.Join(sandboxDir, m.ID)
	if err := os.MkdirAll(projectDest, 0755); err != nil {
		return nil, fmt.Errorf("create sandbox: %w", err)
	}

	// Copy full project.
	ignores := LoadZPIgnore(m.RootDir)
	collector := NewCollector(m.RootDir, FilterFull, ignores)
	files, err := collector.Collect()
	if err != nil {
		return nil, fmt.Errorf("collect project files: %w", err)
	}

	for _, rel := range files {
		if err := copyFile(
			filepath.Join(m.RootDir, rel),
			filepath.Join(projectDest, rel),
		); err != nil {
			return nil, fmt.Errorf("copy %s: %w", rel, err)
		}
	}

	// Copy nexus.yaml from each depends_on project.
	var contracts []string
	contractsDir := filepath.Join(sandboxDir, "contracts")
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return nil, fmt.Errorf("create contracts dir: %w", err)
	}

	for _, dep := range m.DependsOn {
		dep := strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		entry, err := registry.Find(workspaceRoot, dep)
		if err != nil || entry == nil {
			// Dependency not found — note it but don't fail.
			contracts = append(contracts, dep+" (not found)")
			continue
		}
		src := filepath.Join(entry.RootDir, "nexus.yaml")
		dst := filepath.Join(contractsDir, dep+"-nexus.yaml")
		if err := copyFile(src, dst); err == nil {
			contracts = append(contracts, dep)
		}
	}

	// Write README.
	readme := buildDevREADME(m, sandboxDir, contracts)
	if err := os.WriteFile(filepath.Join(sandboxDir, "README.md"), []byte(readme), 0644); err != nil {
		return nil, fmt.Errorf("write README: %w", err)
	}

	return &DevResult{
		SandboxDir:  sandboxDir,
		ProjectID:   m.ID,
		CopiedFiles: len(files),
		Contracts:   contracts,
	}, nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func buildDevREADME(m *manifest.Manifest, sandboxDir string, contracts []string) string {
	var sb strings.Builder
	sb.WriteString("# zp dev sandbox\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s (%s)\n", m.Name, m.ID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n", time.Now().Format("2006-01-02 15:04")))
	sb.WriteString(fmt.Sprintf("**Sandbox:** %s\n\n", sandboxDir))
	sb.WriteString("## Contents\n\n")
	sb.WriteString(fmt.Sprintf("- `%s/` — full project copy\n", m.ID))
	if len(contracts) > 0 {
		sb.WriteString("- `contracts/` — nexus.yaml from dependency projects:\n")
		for _, c := range contracts {
			sb.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}
	sb.WriteString("\n## Usage\n\n")
	sb.WriteString(fmt.Sprintf("```bash\ncd %s/%s\ngo build ./...\n```\n", sandboxDir, m.ID))
	sb.WriteString("\nThis sandbox is isolated from the live workspace.\n")
	sb.WriteString("Changes here do not affect the original project.\n")
	return sb.String()
}
