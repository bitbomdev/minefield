package root

import (
	"github.com/bit-bom/bitbom/cmd/allKeys"
	"github.com/bit-bom/bitbom/cmd/cache"
	"github.com/bit-bom/bitbom/cmd/ingest"
	"github.com/bit-bom/bitbom/cmd/query"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "bitbom",
		Short:             "graphing SBOM's with the power of roaring bitmaps",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	cmd.AddCommand(query.New())
	cmd.AddCommand(ingest.New())
	cmd.AddCommand(allKeys.New())
	cmd.AddCommand(cache.New())

	return cmd
}
