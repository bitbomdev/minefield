package main

import (
	"os"
	"runtime/debug"

	"github.com/bit-bom/bitbom/cmd/root"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			os.Exit(1) //nolint:gocritic
		}
	}()
	rootCmd := root.New()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1) //nolint:gocritic
	}
}
