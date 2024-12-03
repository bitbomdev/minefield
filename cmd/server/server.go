package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	connectcors "connectrpc.com/cors"
	service "github.com/bitbomdev/minefield/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/bitbomdev/minefield/pkg/storages"
	chromadb "github.com/philippgille/chromem-go"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type options struct {
	storage      graph.Storage
	concurrency  int32
	addr         string
	StorageType  string
	StorageAddr  string
	StoragePath  string
	UseInMemory  bool
	CORS         []string
	UseOpenAILLM  bool
	VectorDBPath  string
}

const (
	defaultConcurrency = 10
	defaultAddr        = "localhost:8089"
	redisStorageType   = "redis"
	sqliteStorageType  = "sqlite"
)

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Int32Var(&o.concurrency, "concurrency", defaultConcurrency, "Maximum number of concurrent operations for leaderboard operations")
	cmd.Flags().StringVar(&o.addr, "addr", defaultAddr, "Network address and port for the server (e.g. localhost:8089)")
	cmd.Flags().StringVar(&o.StorageType, "storage-type", sqliteStorageType, "Type of storage to use (e.g., redis, sqlite)")
	cmd.Flags().StringVar(&o.StorageAddr, "storage-addr", "localhost:6379", "Address for redis storage backend")
	cmd.Flags().StringVar(&o.StoragePath, "storage-path", "", "Path to the SQLite database file")
	cmd.Flags().BoolVar(&o.UseInMemory, "use-in-memory", true, "Use in-memory SQLite database")
	cmd.Flags().StringSliceVar(
		&o.CORS,
		"cors",
		[]string{"http://localhost:8089"},
		"Allowed origins for CORS (e.g., 'https://app.bitbom.dev')",
	)
	cmd.Flags().BoolVar(&o.UseOpenAILLM, "use-openai-llm", false, "Use OpenAI LLM for graph analysis")
	cmd.Flags().StringVar(&o.VectorDBPath, "vector-db-path", "./db", "Path to the vector database")
}

func (o *options) ProvideStorage() (graph.Storage, error) {
	switch o.StorageType {
	case redisStorageType:
		return storages.NewRedisStorage(o.StorageAddr)
	case sqliteStorageType:
		return storages.NewSQLStorage(o.StoragePath, o.UseInMemory)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", o.StorageType)
	}
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	var err error
	o.storage, err = o.ProvideStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	server, err := o.setupServer()
	if err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}
	return o.startServer(server)
}

func (o *options) PersistentPreRunE(_ *cobra.Command, _ []string) error {
	if o.StorageType != redisStorageType && o.StorageType != sqliteStorageType {
		return fmt.Errorf("invalid storage-type %q: must be one of [redis, sqlite]", o.StorageType)
	}

	if o.StorageType == sqliteStorageType && o.StoragePath == "" {
		if !o.UseInMemory {
			return fmt.Errorf("storage-path is required when using SQLite with file-based storage")
		}
	}

	if o.StorageType == redisStorageType && o.StorageAddr == "" {
		return fmt.Errorf("storage-addr is required when using Redis (format: host:port)")
	}

	return nil
}

func (o *options) setupServer() (*http.Server, error) {
	if o.concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than zero")
	}

	serviceAddr := o.addr
	if serviceAddr == "" {
		serviceAddr = defaultAddr
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
	path, handler = apiv1connect.NewIngestServiceHandler(newService)
	mux.Handle(path, handler)

	server := &http.Server{
		Addr:    serviceAddr,
		Handler: h2c.NewHandler(withCORS(mux, o), &http2.Server{}),
	}

	return server, nil
}

func (o *options) startServer(server *http.Server) error {
	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	if o.UseOpenAILLM {
		db, err := chromadb.NewPersistentDB(o.VectorDBPath, false)
		if err != nil {
			log.Fatal("failed to initialize ChromaDB:", err)
		}

		c, err := db.CreateCollection("knowledge-base", nil, nil)
		if err != nil {
			log.Fatal("failed to create collection in ChromaDB:", err)
		}

		// Initialize ChromaDB documents
		err = c.AddDocuments(context.Background(), []chromadb.Document{
			{
				ID:      "1",
				Content: "To query dependencies of a package, use the format, and only output the query: dependencies library pkg:<package_name>. All three are necessary.",
			},
			{
				ID:      "2",
				Content: "To find dependents of a package, use the format, and only output the query: dependents library pkg:<package_name>. All three are necessary.",
			},
			{
				ID:      "3",
				Content: "For vulnerabilities related to a package, use the format, and only output the query: dependencies vuln pkg:<package_name>. All three are necessary.",
			},
			{
				ID:      "4",
				Content: "Combine queries using logical operators like and, or, and xor. All three are necessary.",
			},
			{
				ID:      "5",
				Content: "When using 'and', both conditions must be true. For example, and only output the query: dependencies library pkg:A and dependencies library pkg:B.",
			},
			{
				ID:      "6",
				Content: "When using 'or', at least one of the conditions must be true. For example, and only output the query: dependencies library pkg:A or dependencies library pkg:B.",
			},
			{
				ID:      "7",
				Content: "Use 'xor' to indicate that only one of the conditions can be true. For example, and only output the query: dependencies library pkg:A xor dependencies library pkg:B.",
			},
			{
				ID:      "8",
				Content: "You can chain multiple queries together. For example, and only output the query: dependencies library pkg:A and dependents library pkg:B or vulnerabilities vuln pkg:C.",
			},
			{
				ID:      "9",
				Content: "Package names can include versioning information. For example, and only output the query: dependencies library pkg:example-lib@1.0.0.",
			},
			{
				ID:      "10",
				Content: "Ensure that all keywords are used correctly. The keywords are: dependencies, dependents, library, vuln, xor, or, and.",
			},
			{
				ID:      "11",
				Content: "If a query does not specify a package name, it cannot be processed. Always include a package name in your queries.",
			},
			{
				ID:      "12",
				Content: "To check for multiple vulnerabilities across different packages, you can use, and only output the query: dependencies vuln pkg:A or dependencies vuln pkg:B.",
			},
			{
				ID:      "13",
				Content: "When using 'or', 'and', or 'xor', to take the answer of multiple queries and use an binary operator on the whole result with another query, or another set of queries, you can wrap a set of queries in brackets (), these can also be nested. Example: ((dependencies library pkg:A and dependencies library pkg:B) or (dependencies library pkg:C and dependencies library pkg:D)) and (dependents library pkg:E).",
			},
			{
				ID:      "14",
				Content: "Leaderboard queries are a different type of query, they run queries for every single node in the graph and return a sorted list based of length of the result.",
			},
			{
				ID:      "15",
				Content: "To run a leaderboard query, it is quite similar to a regular query, but instead of runing somthing like depedencies library pkg:A, you would run dependencies library. This would create a leaderboard which is sorted by each node's dependencies of type library.",
			},
			{
				ID:      "16",
				Content: "Leaderboards format are basicaly the same as a query, just if you do not include the node name for the last part of the query, it fills it with the node we are using for the leaderboard, which is every single node in the leaderboard. This means that to make a proper leaderboard we should have at least one part that does not include a node name, we can still combine this with another query. For example: (dependencies library) and (dependents library pkg:github.com/bitbomdev/minefield) would create a leaderboard sorted by the number of dependencies a project has that is shared with minefield. We can also repeat this 2 part query multiple times, for example : dependencies library and dependents library would work as well.",
			},
			{
				ID:      "17",
				Content: "If the user, or you are not sure about what the node's name is, you can use the glob pattern to search for nodes. For example, if you want to search for all nodes that start with 'github.com/bitbomdev', you can use the pattern 'github.com/bitbomdev*'. Try to lean get as many nodes as possible, so if they tell you the name is mineifield, and maybe the org is bitbomdev, you can use '*minefield*', since it will match all nodes that contain minefield, since they are not sure about the org.",
			},
			{
				ID:      "18",
				Content: "When glob seaching never assume the position of anything, so wrap everything can in ** on both sides.",
			},
		}, runtime.NumCPU())
		if err != nil {
			log.Fatal("failed to add documents to ChromaDB:", err)
		}
	}

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
func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "server",
		Short:             "Start the minefield server for graph operations and queries",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: o.PersistentPreRunE,
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)
	return cmd
}

func NewServerCommand(storage graph.Storage, o *options) (*cobra.Command, error) {
	o.storage = storage
	cmd := &cobra.Command{
		Use:               "server",
		Short:             "Start the minefield server for graph operations and queries",
		Args:              cobra.ExactArgs(0),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)
	return cmd, nil
}

// withCORS adds CORS support to a Connect HTTP handler.
func withCORS(h http.Handler, o *options) http.Handler {
	middleware := cors.New(cors.Options{
		AllowedOrigins:   o.CORS,
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: true,
		MaxAge:           3600,
	})
	return middleware.Handler(h)
}
