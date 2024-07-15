package query

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/bit-bom/bitbom/pkg"
)

type options struct{}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, args []string) error {
	script := strings.Join(args, " ")
	// Get the storage instance (assuming a function GetStorageInstance exists)
	storage := pkg.GetStorageInstance("localhost:6379")

	execute, err := pkg.ParseAndExecute(script, storage)
	if err != nil {
		return err
	}
	// Print dependencies
	for _, depID := range execute.ToArray() {
		depNode, err := storage.GetNode(depID)
		if err != nil {
			fmt.Println("Failed to get name for ID", depID, ":", err)
			continue
		}
		fmt.Println(depNode.Type, depNode.Name)
	}

	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "query [script]",
		Short:             "Query dependencies and dependents of a project",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
