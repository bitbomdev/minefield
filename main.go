package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // Import for side-effect

	"github.com/bitbomdev/minefield/cmd/root"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/bitbomdev/minefield/pkg/storages"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func main() {
	var pprofEnabled bool
	var pprofAddr string

	app := fx.New(
		storages.NewRedisStorageModule("localhost:6379"),
		fx.Invoke(func(storage graph.Storage, shutdowner fx.Shutdowner) {
			rootCmd := root.New(storage)
			rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
				pprofEnabled, _ = cmd.Flags().GetBool("pprof")
				pprofAddr, _ = cmd.Flags().GetString("pprof-addr")
				if pprofEnabled {
					go func() {
						fmt.Printf("Starting pprof server on %s\n", pprofAddr)
						http.ListenAndServe(pprofAddr, nil)
					}()
				}
			}
			if err := rootCmd.Execute(); err != nil {
				panic(err)
			}
			if err := shutdowner.Shutdown(); err != nil {
				panic(fmt.Sprintf("Failed to shutdown fx err = %s", err))
			}
		}),
		fx.NopLogger,
	)

	app.Run()
}
