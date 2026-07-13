package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/longkey1/got/internal/config"
)

// setupGoroots creates a goroots directory populated with the given
// installed version directories and points the package-level cfg at
// it, restoring the previous cfg afterwards.
func setupGoroots(t *testing.T, installed, configVersions []string) string {
	t.Helper()
	gorootsDir := t.TempDir()
	for _, v := range installed {
		if err := os.Mkdir(filepath.Join(gorootsDir, v), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	orig := cfg
	t.Cleanup(func() { cfg = orig })
	cfg = &config.Config{
		GorootsDir: gorootsDir,
		Versions:   configVersions,
	}
	return gorootsDir
}

// remaining lists the version directories left in gorootsDir, sorted
// lexicographically.
func remaining(t *testing.T, gorootsDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(gorootsDir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// equalStrings compares two string slices, treating nil and empty as
// equal.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// These tests mutate the package-level cfg, so they must not run in
// parallel.
func TestRemoveVersion(t *testing.T) {
	t.Run("removes an installed version", func(t *testing.T) {
		gorootsDir := setupGoroots(t, []string{"1.22.3", "1.21.0"}, nil)

		if err := removeVersion("1.22.3", false); err != nil {
			t.Fatalf("removeVersion() error = %v", err)
		}
		if want := []string{"1.21.0"}; !equalStrings(remaining(t, gorootsDir), want) {
			t.Errorf("remaining versions = %v, want %v", remaining(t, gorootsDir), want)
		}
	})

	t.Run("dry-run keeps the version", func(t *testing.T) {
		gorootsDir := setupGoroots(t, []string{"1.22.3"}, nil)

		if err := removeVersion("1.22.3", true); err != nil {
			t.Fatalf("removeVersion() error = %v", err)
		}
		if want := []string{"1.22.3"}; !equalStrings(remaining(t, gorootsDir), want) {
			t.Errorf("remaining versions = %v, want %v", remaining(t, gorootsDir), want)
		}
	})

	t.Run("not installed version is an error", func(t *testing.T) {
		setupGoroots(t, nil, nil)

		if err := removeVersion("1.99.0", false); err == nil {
			t.Fatal("removeVersion() error = nil, want error")
		}
	})
}

func TestRemoveAllOldVersions(t *testing.T) {
	t.Run("removes old patches of configured minors only", func(t *testing.T) {
		gorootsDir := setupGoroots(t,
			[]string{"1.23.4", "1.23.2", "1.22.5", "1.22.0", "1.21.5", "1.21.0"},
			[]string{"1.23", "1.22"},
		)

		if err := removeAllOldVersions(false); err != nil {
			t.Fatalf("removeAllOldVersions() error = %v", err)
		}
		// 1.23.2 and 1.22.0 are old patches of configured minors;
		// 1.21.x is not in the config and stays untouched.
		want := []string{"1.21.0", "1.21.5", "1.22.5", "1.23.4"}
		if got := remaining(t, gorootsDir); !equalStrings(got, want) {
			t.Errorf("remaining versions = %v, want %v", got, want)
		}
	})

	t.Run("dry-run removes nothing", func(t *testing.T) {
		gorootsDir := setupGoroots(t,
			[]string{"1.23.4", "1.23.2"},
			[]string{"1.23"},
		)

		if err := removeAllOldVersions(true); err != nil {
			t.Fatalf("removeAllOldVersions() error = %v", err)
		}
		want := []string{"1.23.2", "1.23.4"}
		if got := remaining(t, gorootsDir); !equalStrings(got, want) {
			t.Errorf("remaining versions = %v, want %v", got, want)
		}
	})

	t.Run("nothing to remove when only latest patches are installed", func(t *testing.T) {
		gorootsDir := setupGoroots(t,
			[]string{"1.23.4", "1.22.5"},
			[]string{"1.23", "1.22"},
		)

		if err := removeAllOldVersions(false); err != nil {
			t.Fatalf("removeAllOldVersions() error = %v", err)
		}
		want := []string{"1.22.5", "1.23.4"}
		if got := remaining(t, gorootsDir); !equalStrings(got, want) {
			t.Errorf("remaining versions = %v, want %v", got, want)
		}
	})
}
