package main

import (
	"github.com/vdesjardins/terraform-module-versions/cmd/tf-update-module-versions/cmd"
)

// Version information injected at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Set version info in the command package
	cmd.SetVersion(Version, Commit, BuildTime)

	// Execute the command
	cmd.Execute()
}
