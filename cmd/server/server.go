package server

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	storage     graph.Storage
	concurrency int32
	addr        string
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Int32Var(&o.concurrency, "concurrency", 10, "Maximum number of concurrent operations for leaderboard operations")
	cmd.Flags().StringVar(&o.addr, "addr", "localhost:8089", "Network address and port for the server (e.g. localhost:8089)")
}

func (o *options) Run(cmd *cobra.Command, args []string) error {
	server, err := o.setupServer()
	if err != nil {
		return err
	}
	return o.startServer(server)
}

func (o *options) setupServer() (*http.Server, error) {
	if o.concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than zero")
	}

	serviceAddr := o.addr
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
	path, handler = apiv1connect.NewHealthServiceHandler(newService)
	mux.Handle(path, handler)

	server := &http.Server{
		Addr:    serviceAddr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	return server, nil
}

func (o *options) startServer(server *http.Server) error {
	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server is starting on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("ListenAndServe(): %s\n", err)
		}
	}()

	<-stop
	log.Println("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server gracefully stopped")
	return nil
}

// New returns a new cobra command for the server.
func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "server",
		Short:             "Start the minefield server for graph operations and queries",
		Args:              cobra.ExactArgs(0),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
