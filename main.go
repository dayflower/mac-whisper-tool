package main

import (
	"github.com/dayflower/mac-whisper-tool/cmd"
)

// version is populated at release build time by GoReleaser via -ldflags -X.
var version = "dev"

func main() {
	cmd.Execute(version)
}
