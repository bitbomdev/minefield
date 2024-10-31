package scorecard

import (
	sc_graph "github.com/bit-bom/minefield/cmd/ingest/scorecard/graph"
	"github.com/bit-bom/minefield/cmd/ingest/scorecard/load"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

func New(storage graph.Storage) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scorecard",
		Short: "Ingest OpenSSF Scorecard data into storage",
	}

	cmd.AddCommand(load.New(storage))
	cmd.AddCommand(sc_graph.New(storage))
	return cmd
}
