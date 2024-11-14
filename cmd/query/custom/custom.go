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

// options holds the command-line options.
type options struct {
	maxOutput          int
	showInfo           bool
	saveQuery          string
	addr               string
	output             string
	queryServiceClient apiv1connect.QueryServiceClient
}

// AddFlags adds command-line flags to the provided cobra command.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "maximum number of results to display")
	cmd.Flags().BoolVar(&o.showInfo, "show-info", true, "display the info column")
	cmd.Flags().StringVar(&o.addr, "addr", "http://localhost:8089", "address of the minefield server")
	cmd.Flags().StringVar(&o.output, "output", "table", "output format (table or json)")
}

// Run executes the custom command with the provided arguments.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	script := strings.Join(args, " ")
	if strings.TrimSpace(script) == "" {
		return fmt.Errorf("script cannot be empty")
	}

	// Initialize client if not injected (for testing)
	if o.queryServiceClient == nil {
		o.queryServiceClient = apiv1connect.NewQueryServiceClient(
			http.DefaultClient,
			o.addr,
			connect.WithGRPC(),
			connect.WithSendGzip(),
		)
	}

	ctx := cmd.Context()
	req := connect.NewRequest(&apiv1.QueryRequest{
		Script: script,
	})

	res, err := o.queryServiceClient.Query(ctx, req)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	if len(res.Msg.Nodes) == 0 {
		return fmt.Errorf("no nodes found for script: %s", script)
	}

	switch o.output {
	case "json":
		jsonOutput, err := helpers.FormatNodeJSON(res.Msg.Nodes)
		if err != nil {
			return fmt.Errorf("failed to format nodes as JSON: %w", err)
		}
		cmd.Println(string(jsonOutput))
		return nil
	case "table":
		return formatTable(cmd.OutOrStdout(), res.Msg.Nodes, o.maxOutput, o.showInfo)
	default:
		return fmt.Errorf("unknown output format: %s", o.output)
	}
}

// formatTable formats the nodes into a table and writes it to the provided writer.
func formatTable(w io.Writer, nodes []*apiv1.Node, maxOutput int, showInfo bool) error {
	table := tablewriter.NewWriter(w)
	headers := []string{"Name", "Type", "ID"}
	if showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	count := 0
	for _, node := range nodes {
		if count >= maxOutput {
			break
		}

		row := []string{
			node.Name,
			node.Type,
			strconv.FormatUint(uint64(node.Id), 10),
		}

		if showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(node)
			row = append(row, additionalInfo)
		}

		table.Append(row)
		count++
	}

	table.Render()
	return nil
}

// New creates and returns a new Cobra command for executing custom query scripts.
func New() *cobra.Command {
	o := &options{}

	cmd := &cobra.Command{
		Use:               "custom [script]",
		Short:             "Execute a custom query script",
		Long:              "Execute a custom query script to perform tailored queries against the project's dependencies and dependents.",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	return cmd
}
