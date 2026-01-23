package cmd

import (
	"fmt"

	hv "github.com/hashicorp/go-version"
	"github.com/longkey1/got/internal/version"
	"github.com/spf13/cobra"
)

var listRemoteCmd = &cobra.Command{
	Use:     "list-remote",
	Aliases: []string{"ls-remote"},
	Short:   "List downloadable Go versions",
	Long:    "Display a list of all Go versions available for download from golang.org.",
	RunE: func(cmd *cobra.Command, args []string) error {
		latest, err := cmd.Flags().GetBool("latest")
		if err != nil {
			return err
		}

		var versions []*hv.Version
		if latest {
			versions, err = version.RemoteLatestVersions(cfg.GolangUrl)
		} else {
			versions, err = version.RemoteVersions(cfg.GolangUrl)
		}
		if err != nil {
			return err
		}

		for _, v := range versions {
			fmt.Println(v.Original())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listRemoteCmd)
	listRemoteCmd.Flags().Bool("latest", false, "show only the latest patch version for each minor version")
}
