package query

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bit-bom/bitbom/pkg"
	"github.com/spf13/cobra"
)

type options struct {
	outputdir string
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&o.outputdir,
		"output-dir",
		"",
		"specify dir to write the output to",
	)
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	script := strings.Join(args, " ")
	// Get the storage instance (assuming a function GetStorageInstance exists)
	storage := pkg.GetStorageInstance("localhost:6379")

	execute, err := pkg.ParseAndExecute(script, storage, "")
	if err != nil {
		return fmt.Errorf("failed to parse and execute script: %w", err)
	}
	// Print dependencies
	for _, depID := range execute.ToArray() {
		node, err := storage.GetNode(depID)
		if err != nil {
			fmt.Println("Failed to get name for ID", depID, ":", err)
			continue
		}

		if o.outputdir != "" {
			data, err := json.MarshalIndent(node.Metadata, "", "	")
			if err != nil {
				return fmt.Errorf("failed to marshal node metadata: %w", err)
			}
			if _, err := os.Stat(o.outputdir); err != nil {
				return fmt.Errorf("output directory does not exist: %w", err)
			}

			filePath := filepath.Join(o.outputdir, pkg.SanitizeFilename(node.Name)+".json")
			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			_, err = file.Write(data)
			if err != nil {
				return fmt.Errorf("failed to write data to file: %w", err)
			}
		}
	}

	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "query [script]",
		Short:             "Query dependencies and dependents of a project",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
