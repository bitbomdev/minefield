package ingest

import (
	"github.com/bit-bom/minefield/cmd/ingest/osv"
	"github.com/bit-bom/minefield/cmd/ingest/sbom"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ingest",
		Short:             "ingest metadata into the graph",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(osv.New())
	cmd.AddCommand(sbom.New())
	return cmd
}
