package cmd

import (
	"github.com/longkey1/got/internal/installer"
	"github.com/longkey1/got/internal/version"
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
					remoteLatest, err := version.RemoteLatestVersions(cfg.GolangUrl)
					if err != nil {
						return err
					}
					versionToInstall, err = version.LatestVersion(v, remoteLatest)
					if err != nil {
						return err
					}
				}
				if err := installer.Install(versionToInstall, cfg.GolangUrl, cfg.GorootsDir, cfg.TempDir); err != nil {
					return err
				}
			}
		} else {
			v := args[0]
			var versionToInstall string
			if strict {
				versionToInstall = v
			} else {
				remoteLatest, err := version.RemoteLatestVersions(cfg.GolangUrl)
				if err != nil {
					return err
				}
				versionToInstall, err = version.LatestVersion(v, remoteLatest)
				if err != nil {
					return err
				}
			}
			if err := installer.Install(versionToInstall, cfg.GolangUrl, cfg.GorootsDir, cfg.TempDir); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("strict", false, "install the exact version specified (default: install latest patch)")
}
