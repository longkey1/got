package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	hv "github.com/hashicorp/go-version"
	"github.com/longkey1/got/internal/version"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [version]",
	Short: "Remove a specific Go version",
	Long: `Remove an installed Go version from the system.
Use --all-old flag to remove old patch versions while keeping the latest for each minor version.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		allOld, err := cmd.Flags().GetBool("all-old")
		if err != nil {
			return err
		}
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		if allOld {
			return removeAllOldVersions(dryRun)
		}

		if len(args) < 1 {
			return fmt.Errorf("requires a version argument (or use --all-old flag)")
		}

		return removeVersion(args[0], dryRun)
	},
}

func removeAllOldVersions(dryRun bool) error {
	installed, err := version.LocalVersions(cfg.GorootsDir)
	if err != nil {
		return err
	}
	latest := version.LatestMinorVersions(installed)

	// Build a set of latest versions for quick lookup
	latestSet := make(map[string]bool)
	for _, v := range latest {
		latestSet[v.Original()] = true
	}

	// Filter to only versions in config
	configMinors := make(map[string]bool)
	for _, configVer := range cfg.Versions {
		target, err := hv.NewVersion(configVer)
		if err != nil {
			continue
		}
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
		if dryRun {
			fmt.Printf("Would remove %s\n", v.Original())
		} else {
			fmt.Printf("Removing %s\n", v.Original())
			if err := os.RemoveAll(versionDir); err != nil {
				return fmt.Errorf("failed to remove %s: %w", v.Original(), err)
			}
		}
		removedCount++
	}

	if removedCount == 0 {
		fmt.Println("No old versions to remove")
	}
	return nil
}

func removeVersion(ver string, dryRun bool) error {
	versionDir := filepath.Join(cfg.GorootsDir, ver)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", ver)
	}

	if dryRun {
		fmt.Printf("Would remove %s\n", ver)
		return nil
	}

	fmt.Printf("Removing %s\n", ver)
	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", ver, err)
	}
	fmt.Printf("Go %s removed successfully\n", ver)
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().Bool("all-old", false, "remove old patch versions, keeping only the latest for each minor version in config")
	removeCmd.Flags().Bool("dry-run", false, "show what would be removed without actually removing")
}
