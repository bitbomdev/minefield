package query

import (
	"github.com/bit-bom/minefield/cmd/query/custom"
	"github.com/bit-bom/minefield/cmd/query/getMetadata"
	"github.com/bit-bom/minefield/cmd/query/globsearch"
	"github.com/bit-bom/minefield/pkg/graph"
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
