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

func (o *options) Run(_ *cobra.Command, args []string) error {
	vulnsPath := args[0]
	// Ingest vulnerabilities
	result, err := ingest.LoadDataFromPath(o.storage, vulnsPath)
	if err != nil {
		return fmt.Errorf("failed to load vulnerabilities: %w", err)
	}
	for index, data := range result {
		if err := ingest.Vulnerabilities(o.storage, data.Data); err != nil {
			return fmt.Errorf("failed to graph vuln data: %w", err)
		}
		// Clear the line by overwriting with spaces
		fmt.Printf("\r\033[1;36m%-80s\033[0m", " ")
		fmt.Printf("\r\033[K\033[1;36mIngested %d/%d vulnerabilities\033[0m | \033[1;34mCurrent: %s\033[0m", index+1, len(result), tools.TruncateString(data.Path, 50))
	}
	fmt.Println("\nVulnerabilities ingested successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "osv [path to vulnerability file/dir]",
		Short:             "Graph vulnerability data into the graph, and connect it to existing library nodes",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;36mGraphed %d vulnerabilities\033[0m | \033[1;34mCurrent: %s\033[0m", count, tools.TruncateString(path, 50))
}
