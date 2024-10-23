package query

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/RoaringBitmap/roaring"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools"
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage        graph.Storage
	outputdir      string
	visualizerAddr string
	maxOutput      int
	visualize      bool
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.outputdir, "output-dir", "", "specify dir to write the output to")
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
	cmd.Flags().BoolVar(&o.visualize, "visualize", false, "visualize the query")
	cmd.Flags().StringVar(&o.visualizerAddr, "addr", "8081", "address to run the visualizer on")
}

func (o *options) Run(cmd *cobra.Command, args []string) error {
	script := strings.Join(args, " ")
	httpClient := &http.Client{}
	addr := os.Getenv("BITBOMDEV_ADDR")
	if addr == "" {
		addr = "http://localhost:8089"
	}
	client := apiv1connect.NewQueryServiceClient(httpClient, addr)

	// Create a new context
	ctx := cmd.Context()

	// Create a new QueryRequest
	req := connect.NewRequest(&apiv1.QueryRequest{
		Script: script,
	})

	// Make the Query request
	res, err := client.Query(ctx, req)
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

		if o.outputdir != "" {
			data, err := json.MarshalIndent(node.Metadata, "", "	")
			if err != nil {
				return fmt.Errorf("failed to marshal node metadata: %w", err)
			}
			if _, err := os.Stat(o.outputdir); err != nil {
				return fmt.Errorf("output directory does not exist: %w", err)
			}

			filePath := filepath.Join(o.outputdir, tools.SanitizeFilename(node.Name)+".json")
			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			_, err = file.Write(data)
			if err != nil {
				return fmt.Errorf("failed to write data to file: %w", err)
			}
		}
		count++
	}

	if o.visualize {
		server := &http.Server{
			Addr: ":" + o.visualizerAddr,
		}

		ids := roaring.New()

		for _, node := range res.Msg.Nodes {
			ids.Add(node.Id)
		}

		shutdown, err := graph.RunGraphVisualizer(o.storage, ids, script, server)
		if err != nil {
			return err
		}
		defer shutdown()

		fmt.Println("Press Enter to stop the server and continue...")
		if _, err := bufio.NewReader(os.Stdin).ReadBytes('\n'); err != nil {
			return err
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
		Use:               "custom [script]",
		Short:             "Quer dependencies and dependents of a project",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
