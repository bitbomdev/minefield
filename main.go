package main

import (
	"github.com/bit-bom/minefield/cmd/root"
	"github.com/bit-bom/minefield/pkg/storage"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		storage.NewRedisStorageModule("localhost:6379"),
		fx.Invoke(func(storage storage.Storage) {
			rootCmd := root.New(storage)
			if err := rootCmd.Execute(); err != nil {
				panic(err)
			}
		}),
	)

	app.Run()
}
