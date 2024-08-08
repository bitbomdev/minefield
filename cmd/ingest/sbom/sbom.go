package sbom

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg"
	"github.com/bit-bom/minefield/pkg/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage pkg.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, args []string) error {
	sbomPath := args[0]

	// Ingest SBOM
	if err := ingest.SBOM(sbomPath, o.storage); err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	fmt.Println("SBOM ingested successfully")
	return nil
}

func New(storage pkg.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
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
