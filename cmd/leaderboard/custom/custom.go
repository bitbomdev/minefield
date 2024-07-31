package custom

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/bit-bom/bitbom/pkg"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	all bool
}

type query struct {
	node   *pkg.Node
	output []uint32
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.all, "all", false, "show the queries output for each node")
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	storage := pkg.GetStorageInstance("localhost:6379")

	uncachedNodes, err := storage.ToBeCached()
	if err != nil {
		return err
	}
	if len(uncachedNodes) != 0 {
		return fmt.Errorf("cannot use sorted leaderboards without caching")
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("failed to query keys: %w", err)
	}

	// Print dependencies
	queries := []query{}

	for _, key := range keys {
		node, err := storage.GetNode(key)
		if err != nil {
			return err
		}

		if node.Name == "" {
			continue
		}

		execute, err := pkg.ParseAndExecute(args[0], storage, node.Name)

		queries = append(queries, query{node: node, output: execute.ToArray()})
	}

	sort.Slice(queries, func(i, j int) bool {
		return len(queries[i].output) > len(queries[j].output)
	})

	table := tablewriter.NewWriter(os.Stdout)

	if o.all {
		table.SetHeader([]string{"Name", "Type", "ID", "Query"})
	} else {
		table.SetHeader([]string{"Name", "Type", "ID", "QueryLength"})
	}

	for _, q := range queries {
		if o.all {
			table.Append([]string{q.node.Name, q.node.Type, strconv.Itoa(int(q.node.ID)), fmt.Sprint(q.output)})
		} else {
			table.Append([]string{q.node.Name, q.node.Type, strconv.Itoa(int(q.node.ID)), fmt.Sprint(len(q.output))})
		}
	}

	table.Render()

	return nil
}

func New() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:               "custom [script]",
		Short:             "returns all the keys based on the fed in script",
		Args:              cobra.ExactArgs(1),
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
