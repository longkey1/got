package version

import (
	"fmt"
	"runtime"
	"testing"
)

// setBuildInfo overrides the build-time variables for the duration of
// a test and restores them afterwards.
func setBuildInfo(t *testing.T, ver, commit, buildTime string) {
	t.Helper()
	origVersion, origCommit, origBuildTime := Version, CommitSHA, BuildTime
	t.Cleanup(func() {
		Version, CommitSHA, BuildTime = origVersion, origCommit, origBuildTime
	})
	Version, CommitSHA, BuildTime = ver, commit, buildTime
}

func TestInfo(t *testing.T) {
	setBuildInfo(t, "v1.2.3", "abc1234", "2026-01-02T03:04:05Z")

	want := fmt.Sprintf(
		"Version: v1.2.3\nCommit: abc1234\nBuild Time: 2026-01-02T03:04:05Z\nGo Version: %s",
		runtime.Version(),
	)
	if got := Info(); got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestShort(t *testing.T) {
	setBuildInfo(t, "v9.9.9", "unknown", "unknown")

	if got := Short(); got != "v9.9.9" {
		t.Errorf("Short() = %q, want %q", got, "v9.9.9")
	}
}
