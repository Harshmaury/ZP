// @zp-project: zp
// @zp-path: internal/pack/zipper.go
// Zipper creates platform ZIPs with enforced naming convention.
package pack

import (
	"archive/zip"

	"github.com/Harshmaury/ZP/internal/gate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ZipResult is the outcome of a zip operation.
type ZipResult struct {
	ZipPath    string
	FileCount  int
	ProjectID  string
	FilterMode FilterMode
}

// BuildZIP collects files from projectRoot and writes a ZIP to outDir.
// Naming: <projectID>-<filter>-<YYYYMMDD>-<HHMM>.zip
func BuildZIP(projectRoot, projectID string, mode FilterMode, outDir string) (*ZipResult, error) {
	ignores := LoadZPIgnore(projectRoot)
	collector := NewCollector(projectRoot, mode, ignores)

	files, err := collector.Collect()
	if err != nil {
		return nil, fmt.Errorf("collect files: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files matched filter %q in %s", FilterName(mode), projectRoot)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	zipName := buildZipName(projectID, mode)
	zipPath := filepath.Join(outDir, zipName)

	if err := writeZIP(zipPath, projectRoot, files); err != nil {
		return nil, fmt.Errorf("write zip: %w", err)
	}

	return &ZipResult{
		ZipPath:    zipPath,
		FileCount:  len(files),
		ProjectID:  projectID,
		FilterMode: mode,
	}, nil
}

// BuildMultiZIP packages multiple projects into a single ZIP.
// Naming: <id1>-<id2>-full-<YYYYMMDD>-<HHMM>.zip
func BuildMultiZIP(projects []struct{ Root, ID string }, mode FilterMode, outDir string) (*ZipResult, error) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	combinedID := ""
	for i, p := range projects {
		if i > 0 {
			combinedID += "-"
		}
		combinedID += p.ID
	}

	zipName := buildZipName(combinedID, mode)
	zipPath := filepath.Join(outDir, zipName)

	if err := gate.CheckPackaging(projectRoot); err != nil {
		return nil, err
	}

	if err := gate.CheckPackaging(projectRoot); err != nil {
		return nil, err
	}

	zf, err := os.Create(zipPath)
	if err != nil {
		return nil, fmt.Errorf("create zip file: %w", err)
	}
	defer zf.Close()

	w :=
 zip.NewWriter(zf)
	defer w.Close()

	totalFiles := 0
	for _, p := range projects {
		ignores := LoadZPIgnore(p.Root)
		collector := NewCollector(p.Root, mode, ignores)
		files, err := collector.Collect()
		if err != nil {
			return nil, fmt.Errorf("collect %s: %w", p.ID, err)
		}
		for _, rel := range files {
			// Prefix with project ID to avoid collisions.
			if err := addFileToZip(w, p.Root, rel, p.ID+"/"+rel); err != nil {
				return nil, err
			}
			totalFiles++
		}
	}

	return &ZipResult{
		ZipPath:    zipPath,
		FileCount:  totalFiles,
		ProjectID:  combinedID,
		FilterMode: mode,
	}, nil
}

// buildZipName enforces the platform ZIP naming convention.
func buildZipName(projectID string, mode FilterMode) string {
	ts := time.Now().Format("20060102-1504")
	return fmt.Sprintf("%s-%s-%s.zip", projectID, FilterName(mode), ts)
}

// writeZIP creates a ZIP archive from files relative to root.
func writeZIP(zipPath, root string, files []string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()

	w := zip.NewWriter(zf)
	defer w.Close()

	for _, rel := range files {
		if err := addFileToZip(w, root, rel, rel); err != nil {
			return err
		}
	}
	return nil
}

// addFileToZip adds one file to a zip.Writer with the given archive path.
func addFileToZip(w *zip.Writer, root, rel, archivePath string) error {
	srcPath := filepath.Join(root, rel)
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", srcPath, err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = archivePath
	header.Method = zip.Deflate

	dst, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}
