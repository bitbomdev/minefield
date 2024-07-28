package sbom

import (
	"fmt"

	"github.com/bit-bom/bitbom/pkg"
	"github.com/bit-bom/bitbom/pkg/ingest"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, args []string) error {
	// Assuming args[0] is the SBOM file path
	sbomPath := args[0]

	// Get the storage instance (assuming a function GetStorageInstance exists)
	storage := pkg.GetStorageInstance("localhost:6379")

	// Ingest SBOM
	if err := ingest.SBOM(sbomPath, storage); err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	fmt.Println("SBOM ingested successfully")
	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "sbom [sbomPath]",
		Short:             "Ingest an SBOM into the storage",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
