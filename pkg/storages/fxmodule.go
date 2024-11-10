package storages

import (
	"github.com/bitbomdev/minefield/pkg/graph"
	"go.uber.org/fx"
)

func NewRedisStorageModule(addr string) fx.Option {
	return fx.Provide(
		func() graph.Storage {
			return NewRedisStorage(addr)
		},
	)
}
