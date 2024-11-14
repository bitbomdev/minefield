package root

import (
	"github.com/bitbomdev/minefield/cmd/cache"
	"github.com/bitbomdev/minefield/cmd/ingest"
	"github.com/bitbomdev/minefield/cmd/leaderboard"
	"github.com/bitbomdev/minefield/cmd/query"
	"github.com/bitbomdev/minefield/cmd/server"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct {
	pprofAddr    string
	pprofEnabled bool
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&o.pprofEnabled, "pprof", false, "enable pprof server")
	cmd.PersistentFlags().StringVar(&o.pprofAddr, "pprof-addr", "localhost:6060", "address for pprof server")
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "minefield",
		Short:             "graphing SBOM's with the power of roaring bitmaps",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	cmd.AddCommand(query.New())
	cmd.AddCommand(ingest.New(storage))
	cmd.AddCommand(cache.New())
	cmd.AddCommand(leaderboard.New())
	cmd.AddCommand(server.New(storage))

	return cmd
}
