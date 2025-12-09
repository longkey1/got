package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Installed version list",
	Run: func(cmd *cobra.Command, args []string) {
		for _, v := range localVersions() {
			fmt.Println(v.Original())
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
