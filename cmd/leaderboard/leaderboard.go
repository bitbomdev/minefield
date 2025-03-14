package leaderboard

import (
	"github.com/bitbomdev/minefield/cmd/leaderboard/custom"
	"github.com/bitbomdev/minefield/cmd/leaderboard/keys"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "leaderboard",
		Short:             "Commands to display and sort leaderboard data",
		Long:              `Commands to display and sort leaderboard data`,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	cmd.AddCommand(keys.New())
	cmd.AddCommand(custom.New())
	return cmd
}
