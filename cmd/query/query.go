package query

import (
	"github.com/bitbomdev/minefield/cmd/query/custom"
	"github.com/bitbomdev/minefield/cmd/query/getMetadata"
	"github.com/bitbomdev/minefield/cmd/query/globsearch"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

func New(storage graph.Storage) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "query",
		Short:             "Query dependencies and dependents of a project",
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(custom.New(storage))
	cmd.AddCommand(getMetadata.New(storage))
	cmd.AddCommand(globsearch.New())
	return cmd
}
