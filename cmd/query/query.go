package query

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bit-bom/minefield/pkg"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage   pkg.Storage
	outputdir string
	maxOutput int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.outputdir, "output-dir", "", "specify dir to write the output to")
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	script := strings.Join(args, " ")

	execute, err := pkg.ParseAndExecute(script, o.storage, "")
	if err != nil {
		return fmt.Errorf("failed to parse and execute script: %w", err)
	}
	// Print dependencies
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "ID"})

	for index, key := range execute.ToArray() {
		if index > o.maxOutput {
			break
		}
		node, err := o.storage.GetNode(key)
		if err != nil {
			fmt.Println("Failed to get name for ID:", err)
			continue
		}

		table.Append([]string{node.Name, node.Type, strconv.Itoa(int(node.ID))})

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

	table.Render()

	return nil
}

func New(storage pkg.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
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
