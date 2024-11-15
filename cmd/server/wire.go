//go:build wireinject
// +build wireinject

package server

import (
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/google/wire"
	"github.com/spf13/cobra"
)

const redis = "redis"

func InitializeServerCommand(o *options) (*cobra.Command, error) {
	wire.Build(
		ProvideStorage,
		NewServerCommand,
	)
	return nil, nil
}

func ProvideStorage(o *options) (graph.Storage, error) {
	return o.ProvideStorage()
}
