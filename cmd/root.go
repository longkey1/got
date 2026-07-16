package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/longkey1/got/internal/config"
	"github.com/longkey1/got/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "got",
	Short:         "The Golang downloader",
	Long:          "got is a Go version manager that allows you to easily install, list, and manage multiple versions of Go.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version.Version, version.BuildTime, version.CommitSHA)

	defaultPath, err := config.DefaultConfigPath()
	if err != nil {
		defaultPath = "~/.config/got"
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", filepath.Join(defaultPath, "config.toml")))

	rootCmd.PersistentFlags().String("gourl", config.DefaultGolangUrl, "golang url")
	if err := viper.BindPFlag("golang_url", rootCmd.PersistentFlags().Lookup("gourl")); err != nil {
		cobra.CheckErr(err)
	}

	rootCmd.PersistentFlags().String("goroots", config.DefaultGorootsDir, "goroots directory")
	if err := viper.BindPFlag("goroots_dir", rootCmd.PersistentFlags().Lookup("goroots")); err != nil {
		cobra.CheckErr(err)
	}

	rootCmd.PersistentFlags().String("temp", config.DefaultTempDir, "temp directory")
	if err := viper.BindPFlag("temp_dir", rootCmd.PersistentFlags().Lookup("temp")); err != nil {
		cobra.CheckErr(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	cobra.CheckErr(err)
}
