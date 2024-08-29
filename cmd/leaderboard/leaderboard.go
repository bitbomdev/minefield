package leaderboard

import (
	"github.com/bit-bom/minefield/cmd/leaderboard/allKeys"
	"github.com/bit-bom/minefield/cmd/leaderboard/custom"
	"github.com/bit-bom/minefield/cmd/leaderboard/weightedNACD"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct {
	maxConcurrency int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxConcurrency, "max-concurrency", 25, "maximum number of concurrent goroutines")
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "leaderboard",
		Short:             "all the different ways to sort the ingested data",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	cmd.AddCommand(allKeys.New(storage))
	cmd.AddCommand(weightedNACD.New(storage))
	cmd.AddCommand(custom.New(storage, o.maxConcurrency))

	return cmd
}
