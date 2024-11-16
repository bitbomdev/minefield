package ingest

import (
	"github.com/bitbomdev/minefield/cmd/ingest/osv"
	"github.com/bitbomdev/minefield/cmd/ingest/sbom"
	"github.com/bitbomdev/minefield/cmd/ingest/scorecard"
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
	cmd.AddCommand(scorecard.New())
	return cmd
}
