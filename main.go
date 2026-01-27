package main

import (
	"github.com/longkey1/got/cmd"
	"github.com/longkey1/got/internal/version"
)

var (
	ver    = "dev"
	commit = "unknown"
	date   = "unknown"
)

func main() {
	version.Version = ver
	version.CommitSHA = commit
	version.BuildTime = date
	cmd.Execute()
}
