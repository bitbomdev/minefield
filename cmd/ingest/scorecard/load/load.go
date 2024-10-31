package load

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, args []string) error {
	progress := func(count int, id string) {
		fmt.Printf("\r\033[KIngested %d scorecards\033[0m | \033[1;34mCurrent: %s\033[0m", count, id)
	}
	if _, err := o.ScorecardToStorage(args[0], progress); err != nil {
		return fmt.Errorf("failed to load scorecard data: %w", err)
	}
	fmt.Println("\nScorecards loaded successfully")

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "load [scorecard file or directory]",
		Short:             "Load scorecard data into storage",
		RunE:              o.Run,
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

// ScorecardToStorage takes in a dir or file path and dumps the scorecard data into the storage.
// The data ingested can either be json files or zip files
func (o *options) ScorecardToStorage(path string, progress func(count int, id string)) (int, error) {
	return ingest.LoadDataFromPath(o.storage, path, ingest.LoadScorecard, progress)
}
