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
	// load vuln data into storage
	progress := func(count int, path string) {
		fmt.Printf("\r\033[K%s", printProgress(count, path))
	}
	if _, err := o.VulnerabilitiesToStorage(args[0], progress); err != nil {
		return fmt.Errorf("failed to load vuln data: %w", err)
	}
	fmt.Println("\nVulnerabilities loaded successfully")

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "load [zip file or vuln dir or vuln file]",
		Short:             "Load vuln data into storage",
		RunE:              o.Run,
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;36mIngested %d vulnerabilities\033[0m | \033[1;34mCurrent: %s\033[0m", count, tools.TruncateString(path, 50))
}

// VulnerabilitiesToStorage takes in a dir or file path and dumps the vulns into the storage. The vulns can either be json files or zip files
func (o *options) VulnerabilitiesToStorage(path string, progress func(count int, id string)) (int, error) {
	return ingest.LoadDataFromPath(o.storage, path, ingest.LoadVulnerabilities, progress)
}
