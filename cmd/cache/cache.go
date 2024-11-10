package cache

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

type options struct {
	storage graph.Storage
	clear   bool
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.clear, "clear", false, "Clear the cache instead of creating it")
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	httpClient := &http.Client{}
	addr := os.Getenv("BITBOMDEV_ADDR")
	if addr == "" {
		addr = "http://localhost:8089"
	}
	client := apiv1connect.NewCacheServiceClient(httpClient, addr)

	// Create a new context
	ctx := context.Background()

	if o.clear {
		req := connect.NewRequest(&emptypb.Empty{})
		_, err := client.Clear(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("Cache cleared successfully")
	} else {
		req := connect.NewRequest(&emptypb.Empty{})
		_, err := client.Cache(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to cache: %w", err)
		}
		fmt.Println("Graph Fully Cached")
	}
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "cache",
		Short:             "Cache all nodes or remove existing cache",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
