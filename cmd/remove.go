package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	hv "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

var (
	removeAllOld bool
	removeDryRun bool
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove specific version",
	Run: func(cmd *cobra.Command, args []string) {
		if removeAllOld {
			installed := localVersions()
			latest := latestMinorVersions(installed)

			// Build a set of latest versions for quick lookup
			latestSet := make(map[string]bool)
			for _, v := range latest {
				latestSet[v.Original()] = true
			}

			// Filter to only versions in config
			configMinors := make(map[string]bool)
			for _, configVer := range cfg.Versions {
				target, _ := hv.NewVersion(configVer)
				seg := target.Segments()
				minorKey := fmt.Sprintf("%d.%d", seg[0], seg[1])
				configMinors[minorKey] = true
			}

			// Remove old versions that are in config but not latest
			removedCount := 0
			for _, v := range installed {
				seg := v.Segments()
				minorKey := fmt.Sprintf("%d.%d", seg[0], seg[1])

				// Skip if not in config
				if !configMinors[minorKey] {
					continue
				}

				// Skip if it's the latest
				if latestSet[v.Original()] {
					continue
				}

				versionDir := filepath.Join(cfg.GorootsDir, v.Original())
				if removeDryRun {
					fmt.Printf("Would remove %s\n", v.Original())
				} else {
					fmt.Printf("Removing %s\n", v.Original())
					err := os.RemoveAll(versionDir)
					cobra.CheckErr(err)
				}
				removedCount++
			}

			if removedCount == 0 {
				fmt.Println("No old versions to remove.")
			}
			return
		}

		if len(args) < 1 {
			log.Fatalln("requires a version argument.")
		}
		ver := args[0]
		versionDir := filepath.Join(cfg.GorootsDir, ver)
		if removeDryRun {
			fmt.Printf("Would remove %s\n", ver)
		} else {
			err := os.RemoveAll(versionDir)
			cobra.CheckErr(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVar(&removeAllOld, "all-old", false, "Remove old patch versions, keeping only the latest for each minor version in config")
	removeCmd.Flags().BoolVar(&removeDryRun, "dry-run", false, "Show what would be removed without actually removing")
}
