package cmd

import (
	"fmt"

	"github.com/longkey1/got/internal/version"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List installed Go versions",
	Long:    "Display a list of all Go versions installed on the system.",
	RunE: func(cmd *cobra.Command, args []string) error {
		versions, err := version.LocalVersions(cfg.GorootsDir)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			fmt.Println("No versions installed")
			return nil
		}

		for _, v := range versions {
			fmt.Println(v.Original())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
