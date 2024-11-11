package cache

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	DefaultAddr = "http://localhost:8089" // Default address of the minefield server
)

// options for the cache command
type options struct {
	clear bool   // Clear all cached graph data
	addr  string // Address of the minefield server

	cacheServiceClient apiv1connect.CacheServiceClient
}

// AddFlags adds command-line flags to the provided cobra command.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.clear, "clear", false, "Clear all cached graph data")
	cmd.Flags().StringVar(&o.addr, "addr", DefaultAddr, "Address of the minefield server")
}

// Run executes the cache command with the provided arguments.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	// Initialize dependencies if not injected (for testing)
	if err := o.initDependencies(); err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	ctx := cmd.Context()

	if o.clear {
		return o.clearCache(ctx)
	}

	return o.populateCache(ctx)
}

// initDependencies initializes dependencies if they are not already set.
func (o *options) initDependencies() error {
	if o.cacheServiceClient == nil {
		o.cacheServiceClient = apiv1connect.NewCacheServiceClient(
			http.DefaultClient,
			o.addr,
		)
	}
	return nil
}

// clearCache clears the cache by calling the CacheService's Clear method.
func (o *options) clearCache(ctx context.Context) error {
	req := connect.NewRequest(&emptypb.Empty{})
	_, err := o.cacheServiceClient.Clear(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	fmt.Println("Cache cleared successfully")
	return nil
}

// populateCache populates the cache by calling the CacheService's Cache method.
func (o *options) populateCache(ctx context.Context) error {
	req := connect.NewRequest(&emptypb.Empty{})
	_, err := o.cacheServiceClient.Cache(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to populate cache: %w", err)
	}
	fmt.Println("Graph fully cached")
	return nil
}

// New returns a new cobra command for the cache command.
func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "cache",
		Short:             "Cache all nodes or remove existing cache",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
