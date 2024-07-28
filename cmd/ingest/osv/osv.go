package osv

import (
	"fmt"

	"github.com/bit-bom/bitbom/pkg"
	"github.com/bit-bom/bitbom/pkg/ingest"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	// Get the storage instance (assuming a function GetStorageInstance exists)
	storage := pkg.GetStorageInstance("localhost:6379")

	// Ingest SBOM
	if err := ingest.Vulnerabilities(storage); err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	fmt.Println("Vulnerabilities ingested successfully")
	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "osv",
		Short:             "Ingest vulnerabilities into the storage",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
