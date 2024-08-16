package main

import (
	"github.com/bit-bom/minefield/cmd/root"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/storages"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		storages.NewRedisStorageModule("localhost:6379"),
		fx.Invoke(func(storage graph.Storage) {
			rootCmd := root.New(storage)
			if err := rootCmd.Execute(); err != nil {
				panic(err)
			}
		}),
	)

	app.Run()
}
