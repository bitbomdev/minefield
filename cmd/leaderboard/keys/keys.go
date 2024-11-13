package keys

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	v1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

// options defines the command-line options for the allKeys command.
type options struct {
	maxOutput int
	showInfo  bool
	addr      string
	output    string
	client    apiv1connect.LeaderboardServiceClient
}

// AddFlags adds command-line flags to the provided cobra command.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&o.maxOutput, "max-output", "m", 10, "Specify the maximum number of keys to display")
	cmd.Flags().BoolVarP(&o.showInfo, "show-info", "i", true, "Toggle display of additional information for each key")
	cmd.Flags().StringVarP(&o.addr, "addr", "a", "http://localhost:8089", "Address of the Minefield server")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "Output format (table or json)")
}

// Run executes the allKeys command.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	// Initialize HTTP client and LeaderboardServiceClient if not injected
	if o.client == nil {
		httpClient := &http.Client{}
		if o.addr == "" {
			o.addr = "http://localhost:8089"
		}
		o.client = apiv1connect.NewLeaderboardServiceClient(httpClient, o.addr)
	}

	ctx := cmd.Context()

	// Create and send the request
	req := connect.NewRequest(&emptypb.Empty{})
	res, err := o.client.AllKeys(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to retrieve all keys: %w", err)
	}

	// Handle output format
	switch o.output {
	case "json":
		jsonOutput, err := helpers.FormatNodeJSON(res.Msg.Nodes)
		if err != nil {
			return fmt.Errorf("failed to format nodes as JSON: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(jsonOutput)
		if err != nil {
			return fmt.Errorf("failed to write JSON output: %w", err)
		}
	case "table":
		if err := o.renderTable(cmd.OutOrStdout(), res, o.showInfo, o.maxOutput); err != nil {
			return fmt.Errorf("failed to render table: %w", err)
		}
	default:
		return fmt.Errorf("invalid output format specified: %s", o.output)
	}

	return nil
}

// New initializes and returns a new Cobra command for retrieving all leaderboard keys.
// It sets up command usage, descriptions, and binds the necessary flags.
func New() *cobra.Command {
	const (
		use   = "keys"
		short = "Retrieve and display all keys from the database without it being cached"
		long  = `This command fetches all keys from the leaderboard service and displays them
in a neatly formatted table. You can specify the maximum number of keys to display,
toggle additional information, and choose the output format (table or json).`
	)
	o := &options{}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,

		RunE:              o.Run,
		DisableAutoGenTag: true,
	}

	// Add command-line flags to the command
	o.AddFlags(cmd)

	return cmd
}

// renderTable renders the nodes in a table format.
func (o *options) renderTable(w io.Writer, resp *connect.Response[v1.AllKeysResponse], showInfo bool, maxOutput int) error {
	const (
		headerName = "Name"
		headerType = "Type"
		headerID   = "ID"
		headerInfo = "Info"
	)

	if w == nil {
		return fmt.Errorf("writer is nil")
	}
	if resp == nil || resp.Msg == nil || resp.Msg.Nodes == nil {
		return fmt.Errorf("nodes data is invalid")
	}

	// Setup table writer
	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	// Define headers dynamically based on showInfo flag
	headers := []string{headerName, headerType, headerID}
	if showInfo {
		headers = append(headers, headerInfo)
	}
	table.SetHeader(headers)

	for index, node := range resp.Msg.Nodes {
		if index >= maxOutput {
			break
		}

		row := []string{
			node.Name,
			node.Type,
			strconv.Itoa(int(node.Id)),
		}

		if showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(node)
			row = append(row, additionalInfo)
		}

		table.Append(row)
	}

	table.Render()
	return nil
}
