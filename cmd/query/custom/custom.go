package custom

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/RoaringBitmap/roaring"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage        graph.Storage
	visualizerAddr string
	maxOutput      int
	visualize      bool
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
	cmd.Flags().BoolVar(&o.visualize, "visualize", false, "visualize the query")
	cmd.Flags().StringVar(&o.visualizerAddr, "addr", "8081", "address to run the visualizer on")
}

func (o *options) Run(cmd *cobra.Command, args []string) error {
	script := strings.Join(args, " ")

	if strings.TrimSpace(script) == "" {
		return fmt.Errorf("script cannot be empty")
	}
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
		if count >= o.maxOutput {
			break
		}

		table.Append([]string{node.Name, node.Type, strconv.Itoa(int(node.Id))})
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
		Args:              cobra.MinimumNArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
