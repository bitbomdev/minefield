package getMetadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"connectrpc.com/connect"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct {
	storage    graph.Storage
	outputFile string
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.outputFile, "output-file", "", "output file")
}

func (o *options) Run(cmd *cobra.Command, args []string) error {
	httpClient := &http.Client{}
	addr := os.Getenv("BITBOMDEV_ADDR")
	if addr == "" {
		addr = "http://localhost:8089"
	}
	client := apiv1connect.NewGraphServiceClient(httpClient, addr)

	// Create a new context
	ctx := cmd.Context()

	// Create a new QueryRequest
	req := connect.NewRequest(&apiv1.GetNodeByNameRequest{
		Name: args[0],
	})

	res, err := client.GetNodeByName(ctx, req)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	node := res.Msg.Node
	if node == nil {
		return fmt.Errorf("node not found: %s", args[0])
	}

	// Unmarshal the metadata JSON string into a map
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(node.Metadata), &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	output := struct {
		Name     string                 `json:"name"`
		Type     string                 `json:"type"`
		ID       string                 `json:"id"`
		Metadata map[string]interface{} `json:"metadata"` // Change type to map
	}{
		Name:     node.Name,
		Type:     node.Type,
		ID:       strconv.Itoa(int(node.Id)),
		Metadata: metadata, // Use the unmarshaled map
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %v", err)
	}
	if o.outputFile != "" {
		err = os.WriteFile(o.outputFile, jsonOutput, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
	} else {
		fmt.Println(string(jsonOutput))
	}

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
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
