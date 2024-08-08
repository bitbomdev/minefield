package main

import (
	"github.com/bit-bom/minefield/cmd/root"
	"github.com/bit-bom/minefield/pkg"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		pkg.NewRedisStorageModule("localhost:6379"),
		fx.Invoke(func(storage pkg.Storage) {
			rootCmd := root.New(storage)
			if err := rootCmd.Execute(); err != nil {
				panic(err)
			}
		}),
	)

	app.Run()
}
