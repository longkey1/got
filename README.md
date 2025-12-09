# god

god is The golang downloader.


## Usage

```
$ god [command]
```

## Available Commands

- completion  Generate the autocompletion script for the specified shell
- help        Help about any command
- install     Install specific version
- list        Installed version list
- list-remote Downloadable version list
- path        Describe path
- remove      Remove specific version

Use "god [command] --help" for more information about a command.

## Configration

`path/to/god/config.toml`

```
golang_url = "https://golang.org"
goroots_dir = "/path/to/god/goroots"
temp_dir" = "/path/to/god/tmp"
versions = [
  "1.21",
  "1.20",
  "1.19",
  "1.18",
  "1.17",
  "1.16",
]
```
