package query

import (
	custom "github.com/bit-bom/minefield/cmd/query/custom"
	globsearch "github.com/bit-bom/minefield/cmd/query/globsearch"
	output "github.com/bit-bom/minefield/cmd/query/output"
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
	cmd.AddCommand(globsearch.New(storage))
	cmd.AddCommand(output.New(storage))
	return cmd
}
