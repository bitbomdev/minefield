package cache

import (
	"fmt"
	"strings"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
	clear   bool
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.clear, "clear", false, "Clear the cache instead of creating it")
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	if o.clear {
		if err := o.storage.RemoveAllCaches(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("Cache cleared successfully")
	} else {
		if err := graph.Cache(o.storage, printProgress, printProgress); err != nil {
			return fmt.Errorf("failed to cache: %w", err)
		}
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

func printProgress(progress, total int, dependents bool) {
	if total == 0 {
		fmt.Println("Progress total cannot be zero.")
		return
	}
	barLength := 40
	progressRatio := float64(progress) / float64(total)
	progressBar := int(progressRatio * float64(barLength))

	bar := "\033[1;36m" + strings.Repeat("=", progressBar)
	if progressBar < barLength {
		bar += ">"
	}
	bar += strings.Repeat(" ", max(0, barLength-progressBar-1)) + "\033[0m"

	percentage := fmt.Sprintf("\033[1;34m%3d%%\033[0m", int(progressRatio*100))

	if dependents {
		fmt.Printf("\r Caching dependents [%s] %s \033[1;34m(%d/%d)\033[0m", bar, percentage, progress, total)
	} else {
		fmt.Printf("\r Caching dependencies [%s] %s \033[1;34m(%d/%d)\033[0m", bar, percentage, progress, total)
	}
	if progress == total {
		fmt.Println()
	}
}
