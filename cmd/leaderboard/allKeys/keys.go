package allKeys

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

type options struct {
	storage   graph.Storage
	maxOutput int
	showInfo  bool // New field to control the display of the Info column
	saveQuery string
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-getMetadata", 10, "max getMetadata length")
	cmd.Flags().BoolVar(&o.showInfo, "show-info", true, "display the info column")
	cmd.Flags().StringVar(&o.saveQuery, "save-query", "", "save the query to a specific file")
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	httpClient := &http.Client{}
	addr := os.Getenv("BITBOMDEV_ADDR")
	if addr == "" {
		addr = "http://localhost:8089"
	}
	client := apiv1connect.NewLeaderboardServiceClient(httpClient, addr)

	ctx := context.Background()

	req := connect.NewRequest(&emptypb.Empty{})
	res, err := client.AllKeys(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get all keys: %w", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	// Use dynamic headers
	headers := []string{"Name", "Type", "ID"}
	if o.showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)

	var f *os.File
	if o.saveQuery != "" {
		f, err = os.Create(o.saveQuery)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	for index, node := range res.Msg.Nodes {
		if index >= o.maxOutput {
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

		if o.saveQuery != "" {
			f.WriteString(node.Name + "\n")
		}
	}

	table.Render()

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "allKeys",
		Short:             "returns all the keys in a random order",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
