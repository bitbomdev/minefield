package leaderboard

import (
	"github.com/bit-bom/minefield/cmd/leaderboard/allKeys"
	"github.com/bit-bom/minefield/cmd/leaderboard/custom"
	"github.com/bit-bom/minefield/cmd/leaderboard/weightedNACD"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "leaderboard",
		Short:             "all the different ways to sort the ingested data",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	cmd.AddCommand(allKeys.New())
	cmd.AddCommand(weightedNACD.New())
	cmd.AddCommand(custom.New())

	return cmd
}
