package root

import (
	"github.com/bit-bom/minefield/cmd/cache"
	"github.com/bit-bom/minefield/cmd/ingest"
	"github.com/bit-bom/minefield/cmd/leaderboard"
	"github.com/bit-bom/minefield/cmd/query"
	start_service "github.com/bit-bom/minefield/cmd/start-service"
	"github.com/bit-bom/minefield/pkg/graph"
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

	cmd.AddCommand(query.New(storage))
	cmd.AddCommand(ingest.New(storage))
	cmd.AddCommand(cache.New(storage))
	cmd.AddCommand(leaderboard.New(storage))
	cmd.AddCommand(start_service.New(storage))

	return cmd
}
