package osv

import (
	"fmt"

	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/bitbomdev/minefield/pkg/tools"
	"github.com/bitbomdev/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	// Ingest vulnerabilities
	progress := func(count int, path string) {
		fmt.Printf("\r\033[K%s", printProgress(count, path))
	}
	if err := ingest.Vulnerabilities(o.storage, progress); err != nil {
		return fmt.Errorf("failed to graph vuln data: %w", err)
	}
	fmt.Println("\nVulnerabilities graphed successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "graph",
		Short:             "Graph vuln data into the graph, and connect it to existing library nodes",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;36mGraphed %d vulnerabilities\033[0m | \033[1;34mCurrent: %s\033[0m", count, tools.TruncateString(path, 50))
}
