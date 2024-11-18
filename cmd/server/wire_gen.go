// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

// Injectors from wire.go:

func InitializeServerCommand(o *options) (*cobra.Command, error) {
	storage, err := ProvideStorage(o)
	if err != nil {
		return nil, err
	}
	command, err := NewServerCommand(storage, o)
	if err != nil {
		return nil, err
	}
	return command, nil
}

// wire.go:

func ProvideStorage(o *options) (graph.Storage, error) {
	return o.ProvideStorage()
}