package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/longkey1/got/internal/version"
	"github.com/spf13/cobra"
)

// pathCmd represents the path command
var pathCmd = &cobra.Command{
	Use:   "path [version]",
	Short: "Show the installation path for a Go version",
	Long: `Display the installation path for a specific Go version.
If no version is specified, shows the goroots directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println(cfg.GorootsDir)
			return nil
		}

		target := args[0]

		strict, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return err
		}

		var path string
		if strict {
			path = filepath.Join(cfg.GorootsDir, target)
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("version %s not found", target)
			}
		} else {
			versions, err := version.LocalVersions(cfg.GorootsDir)
			if err != nil {
				return err
			}
			latest, err := version.LatestVersion(target, versions)
			if err != nil {
				return err
			}
			path = filepath.Join(cfg.GorootsDir, latest)
		}

		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
	pathCmd.Flags().Bool("strict", false, "match the exact version specified")
}
