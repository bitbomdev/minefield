package cache

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
	clear   bool
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.clear, "clear", false, "Clear the cache instead of creating it")
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	if o.clear {
		if err := o.storage.RemoveAllCaches(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("Cache cleared successfully")
	} else {
		if err := graph.Cache(o.storage); err != nil {
			return fmt.Errorf("failed to cache: %w", err)
		}
		fmt.Println("Finished Caching")
	}
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "cache",
		Short:             "Cache all nodes or remove existing cache",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
