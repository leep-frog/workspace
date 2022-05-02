package main

import (
	"os"

	"github.com/leep-frog/command/sourcerer"
	"github.com/leep-frog/workspace"
)

func main() {
	os.Exit(sourcerer.Source(workspace.CLI()))
}
