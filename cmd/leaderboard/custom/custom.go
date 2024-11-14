package custom

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// options defines the command-line options for the custom command.
type options struct {
	all       bool
	maxOutput int
	showInfo  bool
	saveQuery string
	addr      string
	output    string
	client    apiv1connect.LeaderboardServiceClient
}

// AddFlags adds command-line flags to the provided cobra command.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.all, "all", false, "show the queries output for each node")
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max number of outputs to display")
	cmd.Flags().BoolVar(&o.showInfo, "show-info", true, "display the info column")
	cmd.Flags().StringVarP(&o.addr, "addr", "a", "http://localhost:8089", "Address of the Minefield server")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "Output format (table or json)")
}

// Run executes the custom command.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	script := strings.Join(args, " ")

	// Initialize HTTP client and LeaderboardServiceClient if not injected
	if o.client == nil {
		httpClient := &http.Client{}
		if o.addr == "" {
			o.addr = "http://localhost:8089"
		}
		o.client = apiv1connect.NewLeaderboardServiceClient(httpClient, o.addr, connect.WithGRPC(), connect.WithSendGzip())
	}

	ctx := cmd.Context()

	// Create and send the request
	req := connect.NewRequest(&apiv1.CustomLeaderboardRequest{
		Script: script,
	})
	res, err := o.client.CustomLeaderboard(ctx, req)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	// Handle output format
	switch o.output {
	case "json":
		jsonOutput, err := helpers.FormatCustomQueriesJSON(res.Msg.Queries)
		if err != nil {
			return fmt.Errorf("failed to format queries as JSON: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(jsonOutput)
		if err != nil {
			return fmt.Errorf("failed to write JSON output: %w", err)
		}
	case "table":
		if err := o.renderTable(cmd.OutOrStdout(), res, o.showInfo, o.maxOutput, o.all); err != nil {
			return fmt.Errorf("failed to render table: %w", err)
		}
	default:
		return fmt.Errorf("invalid output format specified: %s", o.output)
	}

	return nil
}

// renderTable renders the custom queries in a table format.
func (o *options) renderTable(w io.Writer, res *connect.Response[apiv1.CustomLeaderboardResponse], showInfo bool, maxOutput int, all bool) error {
	if w == nil {
		return fmt.Errorf("writer is nil")
	}
	if res == nil || res.Msg == nil || res.Msg.Queries == nil {
		return fmt.Errorf("queries data is invalid")
	}

	if len(res.Msg.Queries) == 0 {
		fmt.Fprintln(w, "No data available")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	headers := []string{"Name", "Type", "ID", "Output"}
	if showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)

	for index, q := range res.Msg.Queries {
		if index >= maxOutput {
			break
		}

		// Determine the Output value
		var output string
		if all {
			output = fmt.Sprint(q.Output)
		} else {
			output = fmt.Sprint(len(q.Output))
		}

		// Build the common row data
		row := []string{
			q.Node.Name,
			q.Node.Type,
			strconv.Itoa(int(q.Node.Id)),
			output,
		}

		// If showInfo is true, compute the additionalInfo and append it
		if showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(q.Node)
			row = append(row, additionalInfo)
		}

		// Append the row to the table
		table.Append(row)

	}
	table.Render()
	return nil
}

// New initializes and returns a new Cobra command for the custom leaderboard.
func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "custom [script]",
		Short:             "Returns all the keys based on the provided script",
		Args:              cobra.MinimumNArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
