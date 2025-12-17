package main

import "github.com/dwirx/ghex/cmd/ghex/commands"

// Version is set during build via ldflags
var Version = "0.0.3"

func main() {
	commands.Version = Version
	commands.Execute()
}
