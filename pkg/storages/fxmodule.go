package storages

import (
	"github.com/bit-bom/minefield/pkg/graph"
	"go.uber.org/fx"
)

func NewRedisStorageModule(addr string) fx.Option {
	return fx.Provide(
		func() graph.Storage {
			return NewRedisStorage(addr)
		},
	)
}
