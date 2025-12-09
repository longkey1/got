package cmd

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	hv "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

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
	Use:     "got",
	Short:   "The Golang downloader",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", filepath.Join(defaultConfigPath(), "config.toml")))

	rootCmd.PersistentFlags().String("gourl", DefaultGorootsDir, "golang url")
	err := viper.BindPFlag("golang_url", rootCmd.PersistentFlags().Lookup("gourl"))
	cobra.CheckErr(err)

	rootCmd.PersistentFlags().String("goroots", DefaultGorootsDir, "goroots directory")
	err = viper.BindPFlag("goroots_dir", rootCmd.PersistentFlags().Lookup("goroots"))
	cobra.CheckErr(err)

	rootCmd.PersistentFlags().String("temp", DefaultGorootsDir, "temp directory")
	err = viper.BindPFlag("temp_dir", rootCmd.PersistentFlags().Lookup("temp"))
	cobra.CheckErr(err)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

	cobra.CheckErr(viper.Unmarshal(&cfg))
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

func remoteVersions() []*hv.Version {
	res, err := http.Get(cfg.GolangUrl + "/dl")
	cobra.CheckErr(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		cobra.CheckErr(err)
	}(res.Body)

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	cobra.CheckErr(err)

	var versionsRaw []string
	doc.Find("a.download").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		url, _ := s.Attr("href")
		if strings.HasSuffix(url, "src.tar.gz") == false {
			return
		}
		reg := regexp.MustCompile(`/dl/go([0-9.]+)\.src\.tar\.gz$`)
		ver := reg.FindStringSubmatch(url)
		if len(ver) > 1 {
			versionsRaw = append(versionsRaw, ver[1])
		}
	})

	versions := make([]*hv.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, _ := hv.NewVersion(raw)
		versions[i] = v
	}

	sort.Sort(sort.Reverse(hv.Collection(versions)))

	return versions
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func remoteLatestVersions() []*hv.Version {
	return latestMinorVersions(remoteVersions())
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

func localVersions() []*hv.Version {
	files, err := ioutil.ReadDir(cfg.GorootsDir)
	cobra.CheckErr(err)

	var versionsRaw []string
	for _, file := range files {
		if file.IsDir() == false {
			continue
		}
		versionsRaw = append(versionsRaw, file.Name())
	}

	versions := make([]*hv.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, _ := hv.NewVersion(raw)
		versions[i] = v
	}

	sort.Sort(sort.Reverse(hv.Collection(versions)))

	return versions
}

func latestVersion(ver string, latestVersions []*hv.Version) string {
	target, _ := hv.NewVersion(ver)
	seg := target.Segments()

	latest, err := hv.NewVersion(InitialVersion)
	cobra.CheckErr(err)

	for _, v := range latestVersions {
		if latest.GreaterThan(v) {
			continue
		}
		segl := v.Segments()
		if seg[0] == segl[0] && seg[1] == segl[1] {
			latest = v
			continue
		}
	}

	return latest.Original()
}
