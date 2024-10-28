package graph

import (
	"fmt"
	"github.com/bit-bom/minefield/pkg/tools"
	"github.com/bit-bom/minefield/pkg/tools/ingest"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	// Ingest scorecard
	progress := func(count int, path string) {
		fmt.Printf("\r\033[K%s", printProgress(count, path))
	}
	if err := ingest.Scorecards(o.storage, progress); err != nil {
		return fmt.Errorf("failed to graph scorecard data: %w", err)
	}
	fmt.Println("\nScorecard data graphed successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "graph",
		Short:             "Graph Scorecard data into the graph and connect it to existing library nodes",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)
	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;36mGraphed %d scorecards\033[0m | \033[1;34mCurrent: %s\033[0m", count, tools.TruncateString(path, 50))
}
