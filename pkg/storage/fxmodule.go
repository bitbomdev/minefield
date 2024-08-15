package storage

import (
	"go.uber.org/fx"
)

func NewRedisStorageModule(addr string) fx.Option {
	return fx.Provide(
		func() Storage {
			return NewRedisStorage(addr)
		},
	)
}
