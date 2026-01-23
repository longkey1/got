# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`got` is a Golang version manager written in Go. It downloads, installs, and manages multiple versions of Go on a single system.

## Key Commands

### Building and Development
```bash
go build                    # Build the binary
go run main.go [command]    # Run without building
go test ./...               # Run all tests
go mod tidy                 # Clean up dependencies
```

### Release Management
```bash
make release type=patch dryrun=true   # Show what would happen
make release type=patch dryrun=false  # Create and push new patch version
make release type=minor dryrun=false  # Increment minor version
make release type=major dryrun=false  # Increment major version

make re-release tag=v1.2.3 dryrun=false  # Re-release a specific tag
make re-release dryrun=false             # Re-release latest tag
```

Release process:
1. Pushes to master branch
2. Creates and pushes a new git tag
3. GitHub Actions automatically builds binaries via GoReleaser for multiple platforms (Linux, Darwin, Windows; amd64, arm64, arm)

### Tool Installation
```bash
make tools  # Install goreleaser
```

## Architecture

### Project Structure
- `main.go` - Entry point, sets version info from ldflags
- `cmd/` - Cobra command definitions
  - `cmd/root.go` - Root command and CLI initialization
  - `cmd/install.go` - Install command
  - `cmd/list.go` - List installed versions command
  - `cmd/list-remote.go` - List remote versions command
  - `cmd/remove.go` - Remove command
  - `cmd/path.go` - Path display command
- `internal/` - Internal packages (not importable by external projects)
  - `internal/config/` - Configuration management
  - `internal/version/` - Version fetching and management
  - `internal/installer/` - Installation logic

### Command Structure (Cobra-based)
- `cmd/root.go` - Root command and CLI initialization
  - Initializes configuration using `internal/config`
  - Sets up persistent flags and Viper bindings
  - Provides `SetVersionInfo()` for version injection
- `cmd/install.go` - Downloads and installs Go versions
  - Supports `--strict` flag to install exact version (default: latest patch)
  - Uses `internal/version` to resolve versions
  - Uses `internal/installer` to download and extract
- `cmd/list.go` - Lists installed versions
  - Uses `internal/version.LocalVersions()`
- `cmd/list-remote.go` - Lists downloadable versions from golang.org
  - Supports `--latest` flag to show only latest patch per minor version
  - Uses `internal/version.RemoteVersions()` or `RemoteLatestVersions()`
- `cmd/remove.go` - Removes installed versions
  - Supports `--all-old` flag to remove old patch versions (keeps latest for each minor version in config)
  - Supports `--dry-run` flag to preview what would be removed
  - Uses `internal/version.LocalVersions()` and `LatestMinorVersions()`
- `cmd/path.go` - Shows path information
  - Uses `internal/version` to resolve version paths

### Internal Packages

#### internal/config
- `Config` struct with: `golang_url`, `goroots_dir`, `temp_dir`, `versions`
- `DefaultConfigPath()` - Returns default config directory path
- `Load(cfgFile string)` - Loads configuration from file or defaults
- Uses Viper for configuration management
- Default config location: `~/.config/got/config.toml`
- Supports environment variables via `viper.AutomaticEnv()`

#### internal/version
- `RemoteVersions(golangUrl)` - Fetches all available versions from golang.org/dl
- `RemoteLatestVersions(golangUrl)` - Fetches latest patch per minor version from remote
- `LocalVersions(gorootsDir)` - Lists installed versions
- `LatestVersion(ver, versions)` - Finds latest patch for given minor version
- `LatestMinorVersions(versions)` - Filters to latest patch per minor version
- Uses `hashicorp/go-version` for version parsing and comparison
- Web scraping golang.org/dl with `goquery` to find available versions

#### internal/installer
- `Install(ver, golangUrl, gorootsDir, tempDir)` - Downloads and installs a Go version
- Downloads from golang.org/dl based on OS and architecture
- Extracts to `goroots_dir/VERSION` directory
- Handles both tar.gz (Linux/macOS) and zip (Windows) formats
- Uses `mholt/archiver/v3` for extraction

### Key Dependencies
- `spf13/cobra` - CLI framework
- `spf13/viper` - Configuration management
- `PuerkitoBio/goquery` - HTML parsing for version scraping
- `hashicorp/go-version` - Semantic version handling
- `mholt/archiver/v3` - Archive extraction (tar.gz/zip)

## Configuration Example

```toml
golang_url = "https://golang.org"
goroots_dir = "/home/user/.config/got/goroots"
temp_dir = "/home/user/.config/got/tmp"
versions = [
  "1.23",
  "1.22",
]
```

## Development Guidelines

### Adding New Commands

When adding new commands, follow these Cobra best practices:

1. **Use `RunE` instead of `Run`** for proper error handling:
   ```go
   RunE: func(cmd *cobra.Command, args []string) error {
       // ... your logic
       return nil  // or return error
   }
   ```

2. **Return errors instead of using `log.Fatal()` or `os.Exit()`**:
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }

   // Bad
   if err != nil {
       log.Fatalf("operation failed: %v", err)
   }
   ```

3. **Retrieve flags in the command handler** (avoid global flag variables):
   ```go
   flag, err := cmd.Flags().GetBool("flag-name")
   if err != nil {
       return err
   }
   ```

4. **Add comprehensive documentation**:
   ```go
   var myCmd = &cobra.Command{
       Use:   "mycommand [args]",
       Short: "Brief description",
       Long:  "Detailed description of what this command does.",
       Args:  cobra.MaximumNArgs(1),
       RunE:  func(cmd *cobra.Command, args []string) error { ... },
   }
   ```

5. **Wrap errors with context using `%w`**:
   ```go
   return fmt.Errorf("failed to parse config: %w", err)
   ```

### Helper Function Patterns

All helper functions that can fail should return errors:
```go
func helperFunction() ([]*Type, error) {
    // ... logic
    if err != nil {
        return nil, fmt.Errorf("descriptive error: %w", err)
    }
    return result, nil
}
```

## Notes

- Version info (version, commit, date) is injected at build time via ldflags in `.goreleaser.yaml`
- The `install` command defaults to installing the latest patch version unless `--strict` is used
- The `remove --all-old` command only removes versions that are in the config's `versions` list
- All functions that interact with external resources (HTTP, filesystem) return errors
- Use `os.ReadDir()` instead of deprecated `io/ioutil.ReadDir()`
