package cache

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg"
	"github.com/spf13/cobra"
)

type options struct {
	storage pkg.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	if err := pkg.Cache(o.storage); err != nil {
		return fmt.Errorf("failed to cache: %w", err)
	}

	fmt.Println("Finished Caching")
	return nil
}

func New(storage pkg.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "cache",
		Short:             "Cache all nodes",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
