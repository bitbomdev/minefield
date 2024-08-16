package allKeys

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bit-bom/minefield/pkg/storage"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage   storage.Storage
	maxOutput int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	keys, err := o.storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("failed to query keys: %w", err)
	}

	// Print dependencies
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "ID"})

	for index, key := range keys {
		if index > o.maxOutput {
			break
		}
		node, err := o.storage.GetNode(key)
		if err != nil {
			fmt.Println("Failed to get name for ID:", err)
			continue
		}
		table.Append([]string{node.Name, node.Type, strconv.Itoa(int(node.ID))})
	}

	table.Render()

	return nil
}

func New(storage storage.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "allKeys",
		Short:             "returns all the keys in a random order",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
