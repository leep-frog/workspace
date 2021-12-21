package main

import (
	"github.com/leep-frog/command/sourcerer"
	"github.com/leep-frog/workspace"
)

func main() {
	sourcerer.Source(workspace.CLI())
}
