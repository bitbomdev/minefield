package cache

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	// Get the storage instance (assuming a function GetStorageInstance exists)
	storage := pkg.GetStorageInstance("localhost:6379")

	if err := pkg.Cache(storage); err != nil {
		return fmt.Errorf("failed to cache: %w", err)
	}

	fmt.Println("Finished Caching")
	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "cache",
		Short:             "Cache all nodes",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
