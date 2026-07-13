package goversion

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	hv "github.com/hashicorp/go-version"
)

// mustVersions parses raw version strings into a slice of
// *hv.Version, failing the test on invalid input.
func mustVersions(t *testing.T, raw ...string) []*hv.Version {
	t.Helper()
	versions := make([]*hv.Version, 0, len(raw))
	for _, r := range raw {
		v, err := hv.NewVersion(r)
		if err != nil {
			t.Fatalf("invalid test version %q: %v", r, err)
		}
		versions = append(versions, v)
	}
	return versions
}

// originals converts versions back to their original string form for
// comparison.
func originals(versions []*hv.Version) []string {
	out := make([]string, 0, len(versions))
	for _, v := range versions {
		out = append(out, v.Original())
	}
	return out
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

func TestLatestMinorVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string // sorted descending, as documented
		want []string
	}{
		{
			name: "keeps latest patch per minor",
			in:   []string{"1.23.4", "1.23.2", "1.22.5", "1.22.0", "1.21.0"},
			want: []string{"1.23.4", "1.22.5", "1.21.0"},
		},
		{
			name: "single version",
			in:   []string{"1.22.3"},
			want: []string{"1.22.3"},
		},
		{
			name: "already one per minor",
			in:   []string{"1.23.1", "1.22.9"},
			want: []string{"1.23.1", "1.22.9"},
		},
		{
			name: "empty input",
			in:   nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := LatestMinorVersions(mustVersions(t, tt.in...))
			if !equalStrings(originals(got), tt.want) {
				t.Errorf("LatestMinorVersions() = %v, want %v", originals(got), tt.want)
			}
		})
	}
}

func TestLatestVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ver     string
		latest  []string
		want    string
		wantErr bool
	}{
		{
			name:   "matches minor version",
			ver:    "1.22",
			latest: []string{"1.23.4", "1.22.5", "1.21.0"},
			want:   "1.22.5",
		},
		{
			name:   "full version still resolves to latest patch of its minor",
			ver:    "1.22.1",
			latest: []string{"1.23.4", "1.22.5"},
			want:   "1.22.5",
		},
		{
			name:   "picks greatest patch among same minor",
			ver:    "1.22",
			latest: []string{"1.22.1", "1.22.5", "1.22.3"},
			want:   "1.22.5",
		},
		{
			name:    "no matching minor",
			ver:     "1.19",
			latest:  []string{"1.23.4", "1.22.5"},
			wantErr: true,
		},
		{
			name:    "invalid version format",
			ver:     "not-a-version",
			latest:  []string{"1.22.5"},
			wantErr: true,
		},
		{
			name:    "empty candidate list",
			ver:     "1.22",
			latest:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := LatestVersion(tt.ver, mustVersions(t, tt.latest...))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("LatestVersion() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("LatestVersion() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("LatestVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocalVersions(t *testing.T) {
	t.Parallel()

	t.Run("returns installed versions sorted descending", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		for _, name := range []string{"1.21.0", "1.23.4", "1.22.5"} {
			if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
				t.Fatal(err)
			}
		}

		got, err := LocalVersions(dir)
		if err != nil {
			t.Fatalf("LocalVersions() error = %v", err)
		}
		want := []string{"1.23.4", "1.22.5", "1.21.0"}
		if !equalStrings(originals(got), want) {
			t.Errorf("LocalVersions() = %v, want %v", originals(got), want)
		}
	})

	t.Run("skips non-version directories and regular files", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "1.22.3"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(dir, "not-a-version"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "1.21.0"), []byte("file, not dir"), 0o644); err != nil {
			t.Fatal(err)
		}

		got, err := LocalVersions(dir)
		if err != nil {
			t.Fatalf("LocalVersions() error = %v", err)
		}
		want := []string{"1.22.3"}
		if !equalStrings(originals(got), want) {
			t.Errorf("LocalVersions() = %v, want %v", originals(got), want)
		}
	})

	t.Run("nonexistent directory yields empty list", func(t *testing.T) {
		t.Parallel()
		got, err := LocalVersions(filepath.Join(t.TempDir(), "does-not-exist"))
		if err != nil {
			t.Fatalf("LocalVersions() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("LocalVersions() = %v, want empty", originals(got))
		}
	})
}

// downloadPage builds a minimal golang.org/dl-style HTML page from
// anchor definitions.
func downloadPage(anchors ...string) string {
	page := "<html><body>"
	for _, a := range anchors {
		page += a
	}
	return page + "</body></html>"
}

// dlServer serves the given HTML body on /dl.
func dlServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dl" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRemoteVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  int
		body    string
		want    []string
		wantErr bool
	}{
		{
			name:   "parses source archive links sorted descending",
			status: http.StatusOK,
			body: downloadPage(
				`<a class="download" href="/dl/go1.21.9.src.tar.gz">go1.21.9.src.tar.gz</a>`,
				`<a class="download" href="/dl/go1.23.4.src.tar.gz">go1.23.4.src.tar.gz</a>`,
				`<a class="download" href="/dl/go1.22.5.src.tar.gz">go1.22.5.src.tar.gz</a>`,
			),
			want: []string{"1.23.4", "1.22.5", "1.21.9"},
		},
		{
			name:   "ignores binary archives and anchors without download class",
			status: http.StatusOK,
			body: downloadPage(
				`<a class="download" href="/dl/go1.22.5.src.tar.gz">src</a>`,
				`<a class="download" href="/dl/go1.22.5.linux-amd64.tar.gz">binary</a>`,
				`<a class="download" href="/dl/go1.22.5.darwin-arm64.pkg">installer</a>`,
				`<a href="/dl/go1.21.0.src.tar.gz">no class</a>`,
			),
			want: []string{"1.22.5"},
		},
		{
			name:   "empty page yields no versions",
			status: http.StatusOK,
			body:   downloadPage(),
			want:   nil,
		},
		{
			name:    "non-200 status is an error",
			status:  http.StatusInternalServerError,
			body:    "boom",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := dlServer(t, tt.status, tt.body)

			got, err := RemoteVersions(srv.URL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("RemoteVersions() = %v, want error", originals(got))
				}
				return
			}
			if err != nil {
				t.Fatalf("RemoteVersions() error = %v", err)
			}
			if !equalStrings(originals(got), tt.want) {
				t.Errorf("RemoteVersions() = %v, want %v", originals(got), tt.want)
			}
		})
	}
}

func TestRemoteLatestVersions(t *testing.T) {
	t.Parallel()

	srv := dlServer(t, http.StatusOK, downloadPage(
		`<a class="download" href="/dl/go1.23.4.src.tar.gz">a</a>`,
		`<a class="download" href="/dl/go1.23.2.src.tar.gz">b</a>`,
		`<a class="download" href="/dl/go1.22.5.src.tar.gz">c</a>`,
		`<a class="download" href="/dl/go1.22.0.src.tar.gz">d</a>`,
	))

	got, err := RemoteLatestVersions(srv.URL)
	if err != nil {
		t.Fatalf("RemoteLatestVersions() error = %v", err)
	}
	want := []string{"1.23.4", "1.22.5"}
	if !equalStrings(originals(got), want) {
		t.Errorf("RemoteLatestVersions() = %v, want %v", originals(got), want)
	}
}
