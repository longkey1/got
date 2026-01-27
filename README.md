# got

`got` is a Golang version manager that downloads, installs, and manages multiple versions of Go on a single system.

## Features

- Install multiple Go versions side by side
- List installed and available versions
- Automatically installs the latest patch version by default
- Remove old versions with flexible filtering options
- Cross-platform support for Linux and macOS
- Support for various architectures including Raspberry Pi

## Supported Platforms

- **Linux**: x86_64 (amd64), ARM64, ARMv6, ARMv7
- **macOS**: x86_64 (Intel), ARM64 (Apple Silicon)

### Raspberry Pi Support

- Raspberry Pi 1, 2: ARMv6 (32-bit)
- Raspberry Pi 3, 4, 5: ARMv7 (32-bit) or ARM64 (64-bit)

## Installation

Download the latest release from the [releases page](https://github.com/longkey1/got/releases) and extract the binary:

```bash
# Example for Linux x86_64
wget https://github.com/longkey1/got/releases/latest/download/got_Linux_x86_64.tar.gz
tar xzf got_Linux_x86_64.tar.gz
sudo mv got /usr/local/bin/
```

## Usage

```bash
got [command]
```

### Available Commands

- `completion` - Generate the autocompletion script for the specified shell
- `help` - Help about any command
- `install` - Install specific version
  - `--strict` - Install exact version (default: install latest patch)
- `list` - List installed versions
- `list-remote` - List downloadable versions
  - `--latest` - Show only latest patch per minor version
- `path` - Show path information
- `remove` - Remove specific version
  - `--all-old` - Remove old patch versions, keeping only the latest for each minor version in config
  - `--dry-run` - Show what would be removed without actually removing

Use `got [command] --help` for more information about a command.

### Examples

```bash
# Install the latest patch version of Go 1.23
got install 1.23

# Install exact version
got install --strict 1.23.5

# List installed versions
got list

# List available remote versions
got list-remote

# Show only latest patch versions
got list-remote --latest

# Remove specific version
got remove 1.22.3

# Remove all old patch versions (keep latest only)
got remove --all-old

# Preview what would be removed
got remove --all-old --dry-run

# Show path information
got path 1.23
```

## Configuration

Default configuration file location: `~/.config/got/config.toml`

```toml
golang_url = "https://golang.org"
goroots_dir = "/home/user/.config/got/goroots"
temp_dir = "/home/user/.config/got/tmp"
versions = [
  "1.23",
  "1.22",
]
```

### Configuration Options

- `golang_url` - Base URL for downloading Go releases (default: https://golang.org)
- `goroots_dir` - Directory to store installed Go versions
- `temp_dir` - Temporary directory for downloads
- `versions` - List of Go minor versions to manage (used by `remove --all-old`)

## License

MIT
