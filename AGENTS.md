# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, etc.) when working with code in this repository.

## Project Overview

`got` is a Go version manager built with Cobra: it downloads official Go releases from golang.org, installs multiple versions side by side under a goroots directory, and lists/removes/resolves them. Versions are specified as minors (`1.23`) and resolved to the latest patch by default; `--strict` opts into exact-version handling. Supported platforms are Linux and macOS (amd64, arm64, armv6/armv7).

## Build Commands

```sh
make build   # Build binary to ./bin/got (name read from .product_name)
make test    # Run tests (go test ./...)
make fmt     # Format code
make vet     # Vet code
make tidy    # Tidy dependencies
make clean   # Remove build artifacts
```

## Release

```sh
make release type=patch|minor|major            # dry run (default)
make release type=patch dryrun=false           # create and push tag
make re-release [tag=vX.Y.Z] dryrun=false      # re-release an existing tag
```

Pushing a `v*` tag triggers `.github/workflows/gorelease.yml`, which builds multi-platform binaries (linux/darwin, amd64/arm64/armv6/armv7) with GoReleaser and uploads them to GitHub Releases. Version metadata is injected via ldflags into `internal/version`.

## Architecture

- `main.go` — entry point; assigns the ldflags-injected `ver`/`commit`/`date` to `internal/version` package variables, then calls `cmd.Execute()`
- `cmd/` — Cobra commands. `root.go` defines persistent flags (`--config`, `--gourl`, `--goroots`, `--temp`; the latter three are bound to viper keys `golang_url`/`goroots_dir`/`temp_dir`) and loads the config into the package-level `cfg` via `cobra.OnInitialize`
  - `install.go` — `got install [version]`: resolves the latest patch for the given minor via `goversion.RemoteLatestVersions` + `LatestVersion` (skipped with `--strict`), then delegates to `installer.Install`. With no argument, installs every version listed in the config's `versions`
  - `list.go` — `got list` (alias `ls`): prints installed versions from `goversion.LocalVersions`
  - `list-remote.go` — `got list-remote` (alias `ls-remote`): prints downloadable versions; `--latest` keeps only the newest patch per minor
  - `remove.go` — `got remove [version]`: removes one installed version (`removeVersion`), or with `--all-old` removes old patch versions of minors listed in the config, keeping the latest per minor (`removeAllOldVersions`); `--dry-run` prints what would be removed without deleting
  - `path.go` — `got path [version]`: prints the goroots directory (no argument), the resolved latest local patch directory for a minor, or the exact directory with `--strict` (error if not installed)
- `internal/config/` — configuration (`Config`: `golang_url`, `goroots_dir`, `temp_dir`, `versions`). `Load` uses viper: reads the explicit `--config` file (error if unreadable) or `<config-dir>/config.toml` (silently optional), applies defaults, and honors matching environment variables via `AutomaticEnv`. `DefaultConfigPath` returns `os.UserConfigDir()/got`, except on darwin where it uses `$XDG_CONFIG_HOME/got` (falling back to `$HOME/.config/got`) so Linux and macOS share the same layout
- `internal/goversion/` — version logic on `hashicorp/go-version` values, always sorted descending: `RemoteVersions` scrapes `<golang_url>/dl` with goquery, collecting `a.download` anchors whose href matches `/dl/go<ver>.src.tar.gz`; `RemoteLatestVersions`/`LatestMinorVersions` keep the newest patch per minor; `LocalVersions` reads version-named subdirectories of the goroots directory (missing directory means empty, non-version names are skipped); `LatestVersion` resolves a requested version to the newest candidate sharing its major.minor
- `internal/installer/` — `Install(ver, golangUrl, gorootsDir, tempDir)`: no-op if `<gorootsDir>/<ver>` already exists; otherwise downloads `<golangUrl>/dl/go<ver>.<GOOS>-<GOARCH>.tar.gz` (`.zip` on windows) into the temp directory, extracts it with `mholt/archiver`, and moves the archive's `go/` directory to `<gorootsDir>/<ver>`; download and extract artifacts are cleaned up via defers
- `internal/version/` — build-time version info (`Version`, `CommitSHA`, `BuildTime`) injected via ldflags; `Info()`/`Short()` formatting

## Key behavior

- Version arguments are treated as minors by default: `got install 1.23` installs the latest 1.23.x, and `got path 1.23` resolves to the newest installed 1.23.x. `--strict` (on `install` and `path`) uses the exact string instead
- The config file (`~/.config/got/config.toml` by default) is optional; flags, matching environment variables, and built-in defaults cover everything. `versions` drives `got install` with no argument and scopes `got remove --all-old` — minors not listed there are never removed
- Installed versions are plain directories named after the version under `goroots_dir`; `got path` output is meant for wiring `GOROOT`/`PATH`
- Remote version discovery is HTML scraping of the golang.org download page, keyed on source-tarball links, so it needs no API token

## Testing

Tests use only the standard library (`testing`, `httptest`), are table-driven with `t.Run` subtests, and live beside the code under test:

- `internal/goversion/version_test.go` — `LatestMinorVersions`, `LatestVersion`, `LocalVersions` (via `t.TempDir()`), and `RemoteVersions`/`RemoteLatestVersions` against an `httptest.Server` serving static download-page HTML
- `internal/config/config_test.go` — `DefaultConfigPath` (env-driven, `t.Setenv`) and `Load` with explicit TOML files (defaults, overrides, missing-file error); not parallel because viper is a global singleton
- `internal/version/version_test.go` — `Info`/`Short` formatting with save/restore of the package variables
- `internal/installer/installer_test.go` — full download/extract flow against an `httptest.Server` serving an in-memory tar.gz, the already-installed no-op, and the non-200 error path
- `cmd/remove_test.go` — `removeVersion` and `removeAllOldVersions` against a temporary goroots directory with an injected package-level `cfg`; not parallel because `cfg` is package state

No network access is required: HTTP-dependent code is tested with `httptest` servers only.
