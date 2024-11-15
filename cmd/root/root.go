package root

import (
	"fmt"
	"net/http"

	"github.com/bitbomdev/minefield/cmd/cache"
	"github.com/bitbomdev/minefield/cmd/ingest"
	"github.com/bitbomdev/minefield/cmd/leaderboard"
	"github.com/bitbomdev/minefield/cmd/query"
	"github.com/bitbomdev/minefield/cmd/server"
	"github.com/spf13/cobra"
)

type Options struct {
	PprofAddr    string
	PprofEnabled bool
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&o.PprofEnabled, "pprof", false, "Enable pprof server")
	cmd.PersistentFlags().StringVar(&o.PprofAddr, "pprof-addr", "localhost:6060", "Address for pprof server")
}

func New() (*cobra.Command, error) {
	o := &Options{}
	rootCmd := &cobra.Command{
		Use:               "minefield",
		Short:             "Graphing SBOM's with the power of roaring bitmaps",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			o.AddFlags(cmd)
			if o.PprofEnabled {
				go func() {
					fmt.Printf("Starting pprof server on %s\n", o.PprofAddr)
					http.ListenAndServe(o.PprofAddr, nil)
				}()
			}
		},
	}

	rootCmd.AddCommand(query.New())
	rootCmd.AddCommand(ingest.New())
	rootCmd.AddCommand(cache.New())
	rootCmd.AddCommand(leaderboard.New())
	rootCmd.AddCommand(server.New())

	return rootCmd, nil
}
