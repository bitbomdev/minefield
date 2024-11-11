package getMetadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"connectrpc.com/connect"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type (
	options struct {
		outputFile         string
		addr               string
		output             string
		graphServiceClient apiv1connect.GraphServiceClient
	}

	nodeOutput struct {
		Name     string                 `json:"name"`
		Type     string                 `json:"type"`
		ID       string                 `json:"id"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
)

// Constants for output formats and default values
const (
	outputFormatJSON  = "json"
	outputFormatTable = "table"
	defaultAddr       = "http://localhost:8089"
)

// New creates and returns a new cobra command for the get-metadata functionality.
// The command allows users to retrieve and display node metadata in different formats.
func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "get-metadata [node name]",
		Short:             "Outputs the node with its metadata",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)
	return cmd
}

// AddFlags adds the command-line flags to the provided cobra command.
// It configures flags for output file, server address, and output format.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.outputFile, "output-file", "", "output file")
	cmd.Flags().StringVar(&o.addr, "addr", defaultAddr, "address of the minefield server")
	cmd.Flags().StringVar(&o.output, "output", outputFormatJSON, "output format (json or table)")
}

// Run executes the get-metadata command with the provided arguments.
// It fetches the node data and formats the output according to the specified format.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	node, err := o.fetchNode(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return o.formatAndDisplayOutput(cmd, node)
}

// fetchNode retrieves node information from the server using the provided node name.
// It returns the node data or an error if the fetch operation fails.
func (o *options) fetchNode(ctx context.Context, nodeName string) (*apiv1.Node, error) {
	if nodeName == "" {
		return nil, fmt.Errorf("node name is required")
	}

	if o.graphServiceClient == nil {
		o.graphServiceClient = apiv1connect.NewGraphServiceClient(
			http.DefaultClient,
			o.addr,
		)
	}

	req := connect.NewRequest(&apiv1.GetNodeByNameRequest{
		Name: nodeName,
	})

	res, err := o.graphServiceClient.GetNodeByName(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	if res.Msg.Node == nil {
		return nil, fmt.Errorf("node not found: %s", nodeName)
	}

	return res.Msg.Node, nil
}

// formatAndDisplayOutput formats the node data according to the specified output format
// and displays or saves the result. It supports JSON and table formats.
func (o *options) formatAndDisplayOutput(cmd *cobra.Command, node *apiv1.Node) error {
	switch o.output {
	case outputFormatJSON:
		return o.handleJSONOutput(cmd, node)
	case outputFormatTable:
		return formatTable(cmd.OutOrStdout(), node)
	default:
		return fmt.Errorf("unknown output format: %s", o.output)
	}
}

// handleJSONOutput processes the node data into JSON format and either writes it
// to a file or displays it to the command output.
func (o *options) handleJSONOutput(cmd *cobra.Command, node *apiv1.Node) error {
	jsonOutput, err := formatNodeJSON(node)
	if err != nil {
		return fmt.Errorf("failed to format node as JSON: %v", err)
	}

	if o.outputFile != "" {
		if err := os.WriteFile(o.outputFile, jsonOutput, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
		return nil
	}

	cmd.Println(string(jsonOutput))
	return nil
}

// formatNodeJSON converts a node into a JSON formatted byte slice.
// It handles the conversion of metadata and ensures proper formatting of all fields.
func formatNodeJSON(node *apiv1.Node) ([]byte, error) {
	if node == nil {
		return nil, fmt.Errorf("node cannot be nil")
	}

	var metadata map[string]interface{}
	if len(node.Metadata) > 0 {
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata for node %s: %w", node.Name, err)
		}
	}

	output := nodeOutput{
		Name:     node.Name,
		Type:     node.Type,
		ID:       strconv.FormatUint(uint64(node.Id), 10),
		Metadata: metadata,
	}

	return json.MarshalIndent(output, "", "  ")
}

// formatTable writes the node information in a tabular format to the provided writer.
// It creates a table with columns for Name, Type, and ID.
func formatTable(w io.Writer, node *apiv1.Node) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Type", "ID"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)

	table.Append([]string{
		node.Name,
		node.Type,
		strconv.FormatUint(uint64(node.Id), 10),
	})

	table.Render()
	return nil
}
