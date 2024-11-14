package query

import (
	"github.com/bitbomdev/minefield/cmd/query/custom"
	"github.com/bitbomdev/minefield/cmd/query/getMetadata"
	"github.com/bitbomdev/minefield/cmd/query/globsearch"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query dependencies and dependents of a project",
		Long:  "A comprehensive set of commands to query dependencies and dependents of a project, enabling detailed data retrieval and analysis.",
		DisableAutoGenTag: true,
	}

	// Add subcommands
	cmd.AddCommand(custom.New())
	cmd.AddCommand(getMetadata.New())
	cmd.AddCommand(globsearch.New())

	return cmd
}
