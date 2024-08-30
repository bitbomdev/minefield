package osv

import (
	"fmt"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	// Ingest SBOM
	progress := func(count int, path string) {
		fmt.Printf("\r\033[K%s", printProgress(count, path))
	}
	if err := ingest.Vulnerabilities(o.storage, progress); err != nil {
		return fmt.Errorf("failed to ingest SBOM: %w", err)
	}

	fmt.Println("Vulnerabilities ingested successfully")
	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "osv",
		Short:             "Ingest vulnerabilities into storage",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;33mIngested %d vulnerabilities\033[0m | Current: %s", count, tools.TruncateString(path, 50))
}
