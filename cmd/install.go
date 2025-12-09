package cmd

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install specific version",
	Run: func(cmd *cobra.Command, args []string) {
		strict, err := cmd.Flags().GetBool("strict")
		cobra.CheckErr(err)

		if len(args) < 1 {
			for _, v := range cfg.Versions {
				if strict {
					install(v)
				} else {
					install(latestVersion(v, remoteLatestVersions()))
				}
			}
		} else {
			v := args[0]

			if strict {
				install(v)
			} else {
				install(latestVersion(v, remoteLatestVersions()))
			}
		}
	},
}

func install(ver string) {
	info, err := os.Stat(filepath.Join(cfg.GorootsDir, ver))
	if os.IsNotExist(err) == false && info.IsDir() {
		log.Printf("%s is already exists.", ver)
		return
	}

	ext := "tar.gz"
	if runtime.GOARCH == "windows" {
		ext = "zip"
	}
	url := fmt.Sprintf("%s/dl/go%s.%s-%s.%s", cfg.GolangUrl, ver, runtime.GOOS, runtime.GOARCH, ext)

	res, err := http.Get(url)
	cobra.CheckErr(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		cobra.CheckErr(err)
	}(res.Body)

	archiveFile := filepath.Join(cfg.TempDir, fmt.Sprintf("god-archive.%s", ext))
	archive, err := os.Create(archiveFile)
	cobra.CheckErr(err)
	defer func(out *os.File) {
		err := out.Close()
		cobra.CheckErr(err)
	}(archive)

	_, err = io.Copy(archive, res.Body)
	cobra.CheckErr(err)

	extractDir := filepath.Join(cfg.TempDir, fmt.Sprintf("god-extract-%s", ver))
	err = archiver.Unarchive(archiveFile, extractDir)
	cobra.CheckErr(err)

	err = os.Rename(filepath.Join(extractDir, "go"), filepath.Join(cfg.GorootsDir, ver))
	cobra.CheckErr(err)

	err = os.Remove(archiveFile)
	cobra.CheckErr(err)

	err = os.Remove(extractDir)
	cobra.CheckErr(err)

	log.Printf("%s is installed.", ver)
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("strict", false, "If true, install the given version strictly")
}
