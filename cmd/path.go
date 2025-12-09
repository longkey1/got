package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

// pathCmd represents the path command
var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Describe path",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(cfg.GorootsDir)
			return
		}

		target := args[0]

		strict, err := cmd.Flags().GetBool("strict")
		cobra.CheckErr(err)

		var path string
		if strict {
			path = filepath.Join(cfg.GorootsDir, target)
			_, err := os.Stat(path)
			if err != nil {
				log.Fatalf("Not found %s matched version\n", target)
			}
		} else {
			latest := latestVersion(target, localVersions())
			if latest == InitialVersion {
				log.Fatalf("Not found %s matched version\n", target)
			}
			path = filepath.Join(cfg.GorootsDir, latest)
		}

		fmt.Println(path)
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
	pathCmd.Flags().Bool("strict", false, "If true, return the path of the target version strictly")
}
