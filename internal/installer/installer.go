package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver/v3"
)

// Install downloads and installs the specified Go version.
func Install(ver, golangUrl, gorootsDir, tempDir string) error {
	targetDir := filepath.Join(gorootsDir, ver)
	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		fmt.Printf("%s is already installed\n", ver)
		return nil
	}

	// Ensure directories exist
	if err := os.MkdirAll(gorootsDir, 0755); err != nil {
		return fmt.Errorf("failed to create goroots directory: %w", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	url := fmt.Sprintf("%s/dl/go%s.%s-%s.%s", golangUrl, ver, runtime.GOOS, runtime.GOARCH, ext)

	fmt.Printf("Downloading Go %s from %s...\n", ver, url)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		return fmt.Errorf("download failed with status: %d %s", res.StatusCode, res.Status)
	}

	archiveFile := filepath.Join(tempDir, fmt.Sprintf("got-archive-%s.%s", ver, ext))
	archive, err := os.Create(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer func() { _ = archive.Close() }()
	defer func() { _ = os.Remove(archiveFile) }()

	if _, err = io.Copy(archive, res.Body); err != nil {
		return fmt.Errorf("failed to save archive: %w", err)
	}
	_ = archive.Close()

	fmt.Printf("Extracting...\n")
	extractDir := filepath.Join(tempDir, fmt.Sprintf("got-extract-%s", ver))
	if err = archiver.Unarchive(archiveFile, extractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	defer func() { _ = os.RemoveAll(extractDir) }()

	if err = os.Rename(filepath.Join(extractDir, "go"), targetDir); err != nil {
		return fmt.Errorf("failed to move extracted files: %w", err)
	}

	fmt.Printf("Go %s installed successfully\n", ver)
	return nil
}
