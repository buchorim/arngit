// Package main is the entry point for the ArnGit CLI.
package main

import (
	"fmt"
	"os"

	"github.com/arfrfrr/arngit/internal/command"
	"github.com/arfrfrr/arngit/internal/core"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialize engine
	engine, err := core.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	defer engine.Close()

	// Set version info
	engine.SetVersion(Version, BuildTime, GitCommit)

	// Create command router
	router := command.NewRouter(engine)

	// Execute
	if err := router.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
