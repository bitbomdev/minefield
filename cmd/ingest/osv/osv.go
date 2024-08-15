package osv

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	// Ingest SBOM
	if err := ingest.Vulnerabilities(o.storage); err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	fmt.Println("Vulnerabilities ingested successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "osv",
		Short:             "Ingest vulnerabilities into storage",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
