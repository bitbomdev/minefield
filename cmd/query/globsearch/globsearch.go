package globsearch

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"connectrpc.com/connect"
	"github.com/bit-bom/minefield/cmd/helpers"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage   graph.Storage
	maxOutput int
	showInfo  bool // New field to control the display of the Info column
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "maximum number of results to display")
	cmd.Flags().BoolVar(&o.showInfo, "show-info", true, "display the info column")
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

	if len(res.Msg.Nodes) == 0 {
		fmt.Println("No nodes found matching pattern:", pattern)
		return nil
	}

	// Initialize the table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	// Dynamically set the header based on the showInfo flag
	headers := []string{"Name", "Type", "ID"}
	if o.showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)

	count := 0
	for _, node := range res.Msg.Nodes {
		if count >= o.maxOutput {
			break
		}

		// Build the common row data
		row := []string{
			node.Name,
			node.Type,
			strconv.Itoa(int(node.Id)),
		}

		// If showInfo is true, compute the additionalInfo and append it
		if o.showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(node)
			row = append(row, additionalInfo)
		}

		// Append the row to the table
		table.Append(row)
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
