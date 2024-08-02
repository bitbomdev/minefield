package allKeys

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bit-bom/minefield/pkg"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	storage := pkg.GetStorageInstance("localhost:6379")

	keys, err := storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("failed to query keys: %w", err)
	}

	// Print dependencies
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "ID"})

	for _, key := range keys {
		node, err := storage.GetNode(key)
		if err != nil {
			fmt.Println("Failed to get name for ID:", err)
			continue
		}
		table.Append([]string{node.Name, node.Type, strconv.Itoa(int(node.ID))})
	}

	table.Render()

	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "allKeys",
		Short:             "returns all the keys in a random order",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
