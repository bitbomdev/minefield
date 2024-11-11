package osv

import (
	graphData "github.com/bitbomdev/minefield/cmd/ingest/osv/graph"
	loadData "github.com/bitbomdev/minefield/cmd/ingest/osv/load"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

func New(storage graph.Storage) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "osv",
		Short:             "Ingest vulnerabilities into storage",
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(loadData.New(storage))
	cmd.AddCommand(graphData.New(storage))
	return cmd
}
