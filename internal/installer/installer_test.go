package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// goArchive builds an in-memory tar.gz archive mimicking an official
// Go release: a top-level "go/" directory containing the given files.
func goArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{
			Name: "go/" + name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestInstall(t *testing.T) {
	t.Run("downloads and extracts a release archive", func(t *testing.T) {
		archive := goArchive(t, map[string]string{
			"VERSION": "go1.22.3",
			"bin/go":  "fake go binary",
		})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write(archive); err != nil {
				t.Errorf("write archive: %v", err)
			}
		}))
		t.Cleanup(srv.Close)

		root := t.TempDir()
		gorootsDir := filepath.Join(root, "goroots")
		tempDir := filepath.Join(root, "tmp")

		if err := Install("1.22.3", srv.URL, gorootsDir, tempDir); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		got, err := os.ReadFile(filepath.Join(gorootsDir, "1.22.3", "VERSION"))
		if err != nil {
			t.Fatalf("installed VERSION file: %v", err)
		}
		if string(got) != "go1.22.3" {
			t.Errorf("VERSION content = %q, want %q", got, "go1.22.3")
		}
		if _, err := os.Stat(filepath.Join(gorootsDir, "1.22.3", "bin", "go")); err != nil {
			t.Errorf("installed bin/go: %v", err)
		}
	})

	t.Run("already installed version is a no-op", func(t *testing.T) {
		root := t.TempDir()
		gorootsDir := filepath.Join(root, "goroots")
		if err := os.MkdirAll(filepath.Join(gorootsDir, "1.22.3"), 0o755); err != nil {
			t.Fatal(err)
		}

		// The URL is never fetched because Install returns before
		// downloading.
		if err := Install("1.22.3", "http://invalid.invalid", gorootsDir, filepath.Join(root, "tmp")); err != nil {
			t.Fatalf("Install() error = %v", err)
		}
	})

	t.Run("non-200 download response is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		t.Cleanup(srv.Close)

		root := t.TempDir()
		err := Install("1.22.3", srv.URL, filepath.Join(root, "goroots"), filepath.Join(root, "tmp"))
		if err == nil {
			t.Fatal("Install() error = nil, want error")
		}
	})
}
