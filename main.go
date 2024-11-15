package main

import (
	"fmt"
	"os"

	"github.com/bitbomdev/minefield/cmd/root"
)

func main() {
	rootCmd, err := root.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing root command: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
