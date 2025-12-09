# got

got is The golang downloader.


## Usage

```
$ got [command]
```

## Available Commands

- completion  Generate the autocompletion script for the specified shell
- help        Help about any command
- install     Install specific version
- list        Installed version list
- list-remote Downloadable version list
- path        Describe path
- remove      Remove specific version

Use "got [command] --help" for more information about a command.

## Configration

`path/to/got/config.toml`

```
golang_url = "https://golang.org"
goroots_dir = "/path/to/got/goroots"
temp_dir" = "/path/to/got/tmp"
versions = [
  "1.21",
  "1.20",
  "1.19",
  "1.18",
  "1.17",
  "1.16",
]
```
