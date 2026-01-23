package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install a specific Go version",
	Long: `Install a Go version from golang.org.
If no version is specified, installs all versions listed in the config file.
By default, installs the latest patch version for the specified minor version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		strict, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}

		if len(args) < 1 {
			// Install all versions from config
			for _, v := range cfg.Versions {
				var versionToInstall string
				if strict {
					versionToInstall = v
				} else {
					remoteLatest, err := remoteLatestVersions()
					if err != nil {
						return err
					}
					versionToInstall, err = latestVersion(v, remoteLatest)
					if err != nil {
						return err
					}
				}
				if err := install(versionToInstall); err != nil {
					return err
				}
			}
		} else {
			v := args[0]
			var versionToInstall string
			if strict {
				versionToInstall = v
			} else {
				remoteLatest, err := remoteLatestVersions()
				if err != nil {
					return err
				}
				versionToInstall, err = latestVersion(v, remoteLatest)
				if err != nil {
					return err
				}
			}
			if err := install(versionToInstall); err != nil {
				return err
			}
		}
		return nil
	},
}

func install(ver string) error {
	targetDir := filepath.Join(cfg.GorootsDir, ver)
	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		fmt.Printf("%s is already installed\n", ver)
		return nil
	}

	// Ensure directories exist
	if err := os.MkdirAll(cfg.GorootsDir, 0755); err != nil {
		return fmt.Errorf("failed to create goroots directory: %w", err)
	}
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	url := fmt.Sprintf("%s/dl/go%s.%s-%s.%s", cfg.GolangUrl, ver, runtime.GOOS, runtime.GOARCH, ext)

	fmt.Printf("Downloading Go %s from %s...\n", ver, url)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("download failed with status: %d %s", res.StatusCode, res.Status)
	}

	archiveFile := filepath.Join(cfg.TempDir, fmt.Sprintf("got-archive-%s.%s", ver, ext))
	archive, err := os.Create(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer archive.Close()
	defer os.Remove(archiveFile)

	if _, err = io.Copy(archive, res.Body); err != nil {
		return fmt.Errorf("failed to save archive: %w", err)
	}
	archive.Close()

	fmt.Printf("Extracting...\n")
	extractDir := filepath.Join(cfg.TempDir, fmt.Sprintf("got-extract-%s", ver))
	if err = archiver.Unarchive(archiveFile, extractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	defer os.RemoveAll(extractDir)

	if err = os.Rename(filepath.Join(extractDir, "go"), targetDir); err != nil {
		return fmt.Errorf("failed to move extracted files: %w", err)
	}

	fmt.Printf("Go %s installed successfully\n", ver)
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("strict", false, "install the exact version specified (default: install latest patch)")
}
