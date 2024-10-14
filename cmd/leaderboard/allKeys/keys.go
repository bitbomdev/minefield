package allKeys

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"connectrpc.com/connect"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

type options struct {
	storage   graph.Storage
	maxOutput int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
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
	table.SetHeader([]string{"Name", "Type", "ID"})

	for index, node := range res.Msg.Nodes {
		if index > o.maxOutput {
			break
		}
		table.Append([]string{node.Name, node.Type, strconv.Itoa(int(node.Id))})
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
