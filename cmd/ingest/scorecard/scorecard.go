package scorecard

import (
	"fmt"

	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/bitbomdev/minefield/pkg/tools"
	"github.com/bitbomdev/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, args []string) error {
	scorecardPath := args[0]

	result, err := ingest.LoadDataFromPath(o.storage, scorecardPath)
	if err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	for index, data := range result {
		if err := ingest.Scorecards(o.storage, data.Data); err != nil {
			return fmt.Errorf("failed to ingest Scorecard: %w", err)
		}
		// Clear the line by overwriting with spaces
		fmt.Printf("\r\033[1;36m%-80s\033[0m", " ")
		fmt.Printf("\r\033[1;36mIngested %d/%d Scorecards\033[0m | \033[1;34m%s\033[0m", index+1, len(result), tools.TruncateString(data.Path, 50))
	}

	fmt.Println("\nScorecards ingested successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "scorecard [path to scorecard file/dir]",
		Short:             "Graph scorecard data into the graph, and connect it to existing library nodes",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;36mIngested %d SBOMs\033[0m | \033[1;34m%s\033[0m", count, tools.TruncateString(path, 50))
}
