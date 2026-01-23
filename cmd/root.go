package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	hv "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultGolangUrl  = "https://golang.org"
	DefaultGorootsDir = "goroots"
	DefaultTempDir    = "tmp"
	InitialVersion    = "0.0.0"
)

type Config struct {
	GolangUrl  string   `mapstructure:"golang_url"`
	GorootsDir string   `mapstructure:"goroots_dir"`
	TempDir    string   `mapstructure:"temp_dir"`
	Versions   []string `mapstructure:"versions"`
}

var (
	cfgFile string
	cfg     Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "got",
	Short: "The Golang downloader",
	Long:  "got is a Go version manager that allows you to easily install, list, and manage multiple versions of Go.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", filepath.Join(defaultConfigPath(), "config.toml")))

	rootCmd.PersistentFlags().String("gourl", DefaultGolangUrl, "golang url")
	if err := viper.BindPFlag("golang_url", rootCmd.PersistentFlags().Lookup("gourl")); err != nil {
		cobra.CheckErr(err)
	}

	rootCmd.PersistentFlags().String("goroots", DefaultGorootsDir, "goroots directory")
	if err := viper.BindPFlag("goroots_dir", rootCmd.PersistentFlags().Lookup("goroots")); err != nil {
		cobra.CheckErr(err)
	}

	rootCmd.PersistentFlags().String("temp", DefaultTempDir, "temp directory")
	if err := viper.BindPFlag("temp_dir", rootCmd.PersistentFlags().Lookup("temp")); err != nil {
		cobra.CheckErr(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// SetDefault must be called before ReadInConfig
	viper.SetDefault("golang_url", DefaultGolangUrl)
	viper.SetDefault("goroots_dir", filepath.Join(defaultConfigPath(), DefaultGorootsDir))
	viper.SetDefault("temp_dir", filepath.Join(defaultConfigPath(), DefaultTempDir))
	viper.SetDefault("versions", []string{})

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in config directory with name ".got" (without extension).
		viper.AddConfigPath(defaultConfigPath())
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		// If --config flag was explicitly set, fail on error
		if cfgFile != "" {
			cobra.CheckErr(fmt.Errorf("failed to read config file: %w", err))
		}
		// Otherwise, it's okay if the default config file doesn't exist
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		cobra.CheckErr(fmt.Errorf("failed to unmarshal config: %w", err))
	}
}

func defaultConfigPath() string {
	config, err := os.UserConfigDir()
	cobra.CheckErr(err)

	if runtime.GOOS == "darwin" {
		config = os.Getenv("XDG_CONFIG_HOME")
		if config == "" {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)

			if home == "" {
				_, err := fmt.Fprintln(os.Stderr, "$XDG_CONFIG_HOME or $HOME are not defined")
				cobra.CheckErr(err)
			}
			config = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(config, "got")
}

func remoteVersions() ([]*hv.Version, error) {
	res, err := http.Get(cfg.GolangUrl + "/dl")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions from %s/dl: %w", cfg.GolangUrl, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var versionsRaw []string
	doc.Find("a.download").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if !exists || !strings.HasSuffix(url, "src.tar.gz") {
			return
		}
		reg := regexp.MustCompile(`/dl/go([0-9.]+)\.src\.tar\.gz$`)
		ver := reg.FindStringSubmatch(url)
		if len(ver) > 1 {
			versionsRaw = append(versionsRaw, ver[1])
		}
	})

	versions := make([]*hv.Version, 0, len(versionsRaw))
	for _, raw := range versionsRaw {
		v, err := hv.NewVersion(raw)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(sort.Reverse(hv.Collection(versions)))

	return versions, nil
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func remoteLatestVersions() ([]*hv.Version, error) {
	versions, err := remoteVersions()
	if err != nil {
		return nil, err
	}
	return latestMinorVersions(versions), nil
}

// latestMinorVersions returns the latest patch version for each minor version.
// Input must be sorted in descending order.
func latestMinorVersions(versions []*hv.Version) []*hv.Version {
	latestMap := make(map[string]*hv.Version)
	for _, v := range versions {
		seg := v.Segments()
		minorKey := fmt.Sprintf("%d.%d", seg[0], seg[1])
		if _, exists := latestMap[minorKey]; exists {
			continue
		}
		latestMap[minorKey] = v
	}

	var result []*hv.Version
	for _, v := range latestMap {
		result = append(result, v)
	}

	sort.Sort(sort.Reverse(hv.Collection(result)))

	return result
}

func localVersions() ([]*hv.Version, error) {
	files, err := os.ReadDir(cfg.GorootsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*hv.Version{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", cfg.GorootsDir, err)
	}

	versions := make([]*hv.Version, 0)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		v, err := hv.NewVersion(file.Name())
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(sort.Reverse(hv.Collection(versions)))

	return versions, nil
}

func latestVersion(ver string, latestVersions []*hv.Version) (string, error) {
	target, err := hv.NewVersion(ver)
	if err != nil {
		return "", fmt.Errorf("invalid version format: %s", ver)
	}
	seg := target.Segments()

	latest, err := hv.NewVersion(InitialVersion)
	if err != nil {
		return "", err
	}

	for _, v := range latestVersions {
		if latest.GreaterThan(v) {
			continue
		}
		segl := v.Segments()
		if seg[0] == segl[0] && seg[1] == segl[1] {
			latest = v
		}
	}

	if latest.Original() == InitialVersion {
		return "", fmt.Errorf("no matching version found for %s", ver)
	}

	return latest.Original(), nil
}
