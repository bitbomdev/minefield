package sbom

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/bitbomdev/minefield/pkg/tools"
	"github.com/spf13/cobra"
)

type options struct {
	addr string // Address of the minefield server

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
	sbomPath := args[0]
	// Ingest SBOM
	result, err := helpers.LoadDataFromPath(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	for index, data := range result {
		req := connect.NewRequest(&apiv1.IngestSBOMRequest{
			Sbom: data.Data,
		})
		if _, err := o.ingestServiceClient.IngestSBOM(context.Background(), req); err != nil {
			return fmt.Errorf("failed to ingest SBOM: %w", err)
		}
		// Clear the line by overwriting with spaces
		fmt.Printf("\r\033[1;36m%-80s\033[0m", " ")
		fmt.Printf("\r\033[1;36mIngested %d/%d SBOMs\033[0m | \033[1;34m%s\033[0m", index+1, len(result), tools.TruncateString(data.Path, 50))
	}

	fmt.Println("\nSBOMs ingested successfully")
	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "sbom [path to sbom file/dir]",
		Short:             "Ingest an sbom into the graph ",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
