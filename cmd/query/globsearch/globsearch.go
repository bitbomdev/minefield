package query

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	
	"connectrpc.com/connect"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage   graph.Storage
	maxOutput int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
}

func (o *options) Run(cmd *cobra.Command, args []string) error {
	pattern := args[0]
	httpClient := &http.Client{}
	addr := os.Getenv("BITBOMDEV_ADDR")
	if addr == "" {
		addr = "http://localhost:8089"
	}
	client := apiv1connect.NewGraphServiceClient(httpClient, addr)

	// Create a new context
	ctx := cmd.Context()

	// Create a new QueryRequest
	req := connect.NewRequest(&apiv1.GetNodesByGlobRequest{
		Pattern: pattern,
	})

	// Make the Query request
	res, err := client.GetNodesByGlob(ctx, req)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	// Print dependencies
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "ID"})

	count := 0
	for _, node := range res.Msg.Nodes {
		if count > o.maxOutput {
			break
		}

		table.Append([]string{node.Name, node.Type, strconv.Itoa(int(node.Id))})
		count++
	}

	table.Render()

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "globsearch [pattern]",
		Short:             "Search for nodes by glob pattern",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
