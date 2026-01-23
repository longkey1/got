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

### Command Structure (Cobra-based)
- `cmd/root.go` - Root command and shared configuration logic
  - Handles config file loading (TOML format in `~/.config/got/config.toml`)
  - Defines `Config` struct with: `golang_url`, `goroots_dir`, `temp_dir`, `versions`
  - Implements version fetching from golang.org/dl
  - Provides `remoteVersions()`, `localVersions()`, `latestVersion()` helper functions
- `cmd/install.go` - Downloads and installs Go versions
  - Supports `--strict` flag to install exact version (default: latest patch)
  - Downloads from golang.org/dl based on OS and architecture
  - Extracts to `goroots_dir/VERSION` directory
- `cmd/list.go` - Lists installed versions
- `cmd/list-remote.go` - Lists downloadable versions from golang.org
- `cmd/remove.go` - Removes installed versions
  - Supports `--all-old` flag to remove old patch versions (keeps latest for each minor version in config)
  - Supports `--dry-run` flag to preview what would be removed
- `cmd/path.go` - Shows path information
- `main.go` - Entry point, sets version info from ldflags

### Configuration System
- Uses Viper for configuration management
- Default config location: `~/.config/got/config.toml` (Linux/Windows) or `~/.config/got/config.toml` (macOS with XDG_CONFIG_HOME)
- Can be overridden with `--config` flag
- Supports environment variables via `viper.AutomaticEnv()`
- Default values set before reading config file

### Version Management Logic
- Uses `hashicorp/go-version` for version parsing and comparison
- Web scraping golang.org/dl with `goquery` to find available versions
- `latestMinorVersions()` function filters to latest patch for each minor version
- `latestVersion()` function finds latest patch for a given minor version string

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
