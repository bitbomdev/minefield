package start_service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	service "github.com/bitbomdev/minefield/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type options struct {
	storage graph.Storage

	concurrency int32
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Int32Var(&o.concurrency, "concurrency", 10, "Number of concurrent operations a given API call can make")
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	if o.concurrency <= 0 {
		return fmt.Errorf("Concurrency must be greater than zero")
	}
	serviceAddr := os.Getenv("BITBOMDEV_ADDR")
	if serviceAddr == "" {
		serviceAddr = "localhost:8089"
	}
	newService := service.NewService(o.storage, o.concurrency)
	mux := http.NewServeMux()
	path, handler := apiv1connect.NewQueryServiceHandler(newService)
	mux.Handle(path, handler)
	path, handler = apiv1connect.NewLeaderboardServiceHandler(newService)
	mux.Handle(path, handler)
	path, handler = apiv1connect.NewCacheServiceHandler(newService)
	mux.Handle(path, handler)
	path, handler = apiv1connect.NewGraphServiceHandler(newService)
	mux.Handle(path, handler)

	server := &http.Server{
		Addr:    serviceAddr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Server is starting on %s\n", serviceAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("ListenAndServe(): %s\n", err)
		}
	}()

	<-stop
	fmt.Println("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	fmt.Println("Server gracefully stopped")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "start-service",
		Short:             "start the server for interactions with the graph",
		Args:              cobra.ExactArgs(0),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
