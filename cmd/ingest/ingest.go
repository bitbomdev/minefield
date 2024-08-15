package ingest

import (
	"github.com/bit-bom/minefield/cmd/ingest/osv"
	"github.com/bit-bom/minefield/cmd/ingest/sbom"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {
}

func New(storage graph.Storage) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ingest",
		Short:             "ingest metadata into the graph",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(osv.New(storage))
	cmd.AddCommand(sbom.New(storage))
	return cmd
}
