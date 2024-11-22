package scorecard

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/spf13/cobra"
)

type options struct {
	addr                string // Address of the minefield server
	ingestServiceClient apiv1connect.IngestServiceClient
}

const (
	DefaultAddr = "http://localhost:8089" // Default address of the minefield server
)

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.addr, "addr", DefaultAddr, "Address of the minefield server")
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	if o.ingestServiceClient == nil {
		o.ingestServiceClient = apiv1connect.NewIngestServiceClient(
			http.DefaultClient,
			o.addr,
		)
	}
	scorecardPath := args[0]

	result, err := helpers.LoadDataFromPath(scorecardPath)
	if err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	for index, data := range result {
		req := connect.NewRequest(&apiv1.IngestScorecardRequest{
			Scorecard: data.Data,
		})
		if _, err := o.ingestServiceClient.IngestScorecard(context.Background(), req); err != nil {
			return fmt.Errorf("failed to ingest Scorecard: %w", err)
		}
		// Clear the line by overwriting with spaces
		fmt.Printf("\r\033[1;36m%-80s\033[0m", " ")
		fmt.Printf("\r\033[1;36mIngested %d/%d Scorecards\033[0m | \033[1;34m%s\033[0m", index+1, len(result), helpers.TruncateString(data.Path, 50))
	}

	fmt.Println("\nScorecards ingested successfully")
	return nil
}

func New() *cobra.Command {
	o := &options{}
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
	return fmt.Sprintf("\033[1;36mIngested %d SBOMs\033[0m | \033[1;34m%s\033[0m", count, helpers.TruncateString(path, 50))
}
