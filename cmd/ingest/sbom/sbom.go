package sbom

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage   graph.Storage
	batchSize int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.batchSize, "batch-size", 100, "default batch node insert size")
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	sbomPath := args[0]

	// Ingest SBOM
	if err := ingest.SBOM(sbomPath, o.storage, o.batchSize); err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	fmt.Println("SBOM ingested successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "sbom [sbomPath]",
		Short:             "Ingest an SBOM into storage",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}