package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeFile creates a file with the given content, creating parent
// directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// DefaultConfigPath depends on process-wide state (environment
// variables), so these subtests use t.Setenv and must not run in
// parallel.
func TestDefaultConfigPath(t *testing.T) {
	t.Run("XDG_CONFIG_HOME is respected", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))
		t.Setenv("HOME", filepath.Join(root, "home"))

		got, err := DefaultConfigPath()
		if err != nil {
			t.Fatalf("DefaultConfigPath() error = %v", err)
		}
		if want := filepath.Join(root, "xdg", "got"); got != want {
			t.Errorf("DefaultConfigPath() = %q, want %q", got, want)
		}
	})

	t.Run("falls back to HOME/.config without XDG_CONFIG_HOME", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", filepath.Join(root, "home"))

		got, err := DefaultConfigPath()
		if err != nil {
			t.Fatalf("DefaultConfigPath() error = %v", err)
		}
		if want := filepath.Join(root, "home", ".config", "got"); got != want {
			t.Errorf("DefaultConfigPath() = %q, want %q", got, want)
		}
	})
}

// Load goes through the global viper instance, so these subtests must
// not run in parallel.
func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    func(defaultPath string) Config
	}{
		{
			name: "all fields set",
			content: `
golang_url = "https://example.com"
goroots_dir = "/opt/goroots"
temp_dir = "/opt/tmp"
versions = ["1.23", "1.22"]
`,
			want: func(string) Config {
				return Config{
					GolangUrl:  "https://example.com",
					GorootsDir: "/opt/goroots",
					TempDir:    "/opt/tmp",
					Versions:   []string{"1.23", "1.22"},
				}
			},
		},
		{
			name:    "missing fields fall back to defaults",
			content: `goroots_dir = "/opt/goroots"` + "\n",
			want: func(defaultPath string) Config {
				return Config{
					GolangUrl:  DefaultGolangUrl,
					GorootsDir: "/opt/goroots",
					TempDir:    filepath.Join(defaultPath, DefaultTempDir),
					Versions:   []string{},
				}
			},
		},
		{
			name:    "empty file uses all defaults",
			content: "",
			want: func(defaultPath string) Config {
				return Config{
					GolangUrl:  DefaultGolangUrl,
					GorootsDir: filepath.Join(defaultPath, DefaultGorootsDir),
					TempDir:    filepath.Join(defaultPath, DefaultTempDir),
					Versions:   []string{},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile := filepath.Join(t.TempDir(), "config.toml")
			writeFile(t, cfgFile, tt.content)

			got, err := Load(cfgFile)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			defaultPath, err := DefaultConfigPath()
			if err != nil {
				t.Fatalf("DefaultConfigPath() error = %v", err)
			}
			want := tt.want(defaultPath)
			if !reflect.DeepEqual(*got, want) {
				t.Errorf("Load() = %+v, want %+v", *got, want)
			}
		})
	}

	t.Run("explicit config file that does not exist is an error", func(t *testing.T) {
		if _, err := Load(filepath.Join(t.TempDir(), "missing.toml")); err == nil {
			t.Fatal("Load() error = nil, want error")
		}
	})
}
