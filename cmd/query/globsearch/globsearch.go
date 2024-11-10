package globsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	maxOutput          int
	addr               string
	output             string
	showAdditionalInfo bool
	graphServiceClient apiv1connect.GraphServiceClient
}

// AddFlags adds command-line flags to the provided cobra command.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "maximum number of results to display")
	cmd.Flags().StringVar(&o.addr, "addr", "http://localhost:8089", "address of the minefield server")
	cmd.Flags().StringVar(&o.output, "output", "table", "output format (table or json)")
	cmd.Flags().BoolVar(&o.showAdditionalInfo, "show-additional-info", false, "show additional info")
}

// Run executes the globsearch command with the provided arguments.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	pattern := args[0]
	if pattern == "" {
		return fmt.Errorf("pattern is required")
	}

	// Initialize client if not injected (for testing)
	if o.graphServiceClient == nil {
		o.graphServiceClient = apiv1connect.NewGraphServiceClient(
			http.DefaultClient,
			o.addr,
		)
	}

	// Query nodes matching pattern
	res, err := o.graphServiceClient.GetNodesByGlob(
		cmd.Context(),
		connect.NewRequest(&apiv1.GetNodesByGlobRequest{Pattern: pattern}),
	)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if len(res.Msg.Nodes) == 0 {
		return fmt.Errorf("no nodes found matching pattern: %s", pattern)
	}

	// Format and display results
	switch o.output {
	case "json":
		jsonOutput, err := FormatNodeJSON(res.Msg.Nodes)
		if err != nil {
			return fmt.Errorf("failed to format nodes as JSON: %w", err)
		}
		cmd.Println(string(jsonOutput))
		return nil
	case "table":
		return formatTable(cmd.OutOrStdout(), res.Msg.Nodes, o.maxOutput, o.showAdditionalInfo)
	default:
		return fmt.Errorf("unknown output format: %s", o.output)
	}
}

// formatTable formats the nodes into a table and writes it to the provided writer.
func formatTable(w io.Writer, nodes []*apiv1.Node, maxOutput int, showAdditionalInfo bool) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Type", "ID"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	if showAdditionalInfo {
		table.SetHeader([]string{"Name", "Type", "ID", "Info"})
	}
	for i, node := range nodes {
		if i >= maxOutput {
			break
		}
		row := []string{
			node.Name,
			node.Type,
			strconv.FormatUint(uint64(node.Id), 10),
		}
		if showAdditionalInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(node)
			row = append(row, additionalInfo)
		}
		table.Append(row)
	}

	table.Render()
	return nil
}

// New returns a new cobra command for globsearch.
func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "globsearch [pattern]",
		Short:             "Search for nodes by glob pattern",
		Long:              "Search for nodes in the graph using a glob pattern",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)
	return cmd
}

type nodeOutput struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FormatNodeJSON formats the nodes as JSON.
func FormatNodeJSON(nodes []*apiv1.Node) ([]byte, error) {
	if nodes == nil {
		return nil, fmt.Errorf("nodes cannot be nil")
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	outputs := make([]nodeOutput, 0, len(nodes))
	for _, node := range nodes {
		var metadata map[string]interface{}
		if len(node.Metadata) > 0 {
			if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for node %s: %w", node.Name, err)
			}
		}

		outputs = append(outputs, nodeOutput{
			Name:     node.Name,
			Type:     node.Type,
			ID:       strconv.FormatUint(uint64(node.Id), 10),
			Metadata: metadata,
		})
	}

	return json.MarshalIndent(outputs, "", "  ")
}
