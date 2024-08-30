package custom

import (
	"container/heap"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage   graph.Storage
	all       bool
	maxOutput int
}

type query struct {
	node   *graph.Node
	output []uint32
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.all, "all", false, "show the queries output for each node")
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
}

func (o *options) Run(_ *cobra.Command, args []string) error {
	uncachedNodes, err := o.storage.ToBeCached()
	if err != nil {
		return err
	}
	if len(uncachedNodes) != 0 {
		return fmt.Errorf("cannot use sorted leaderboards without caching")
	}

	keys, err := o.storage.GetAllKeys()
	if err != nil {
		return fmt.Errorf("failed to query keys: %w", err)
	}

	// Print dependencies
	queries := []query{}

	nodes, err := o.storage.GetNodes(keys)
	if err != nil {
		return fmt.Errorf("failed to batch query nodes from keys: %w", err)
	}

	caches, err := o.storage.GetCaches(keys)
	if err != nil {
		return fmt.Errorf("failed to batch query caches from keys: %w", err)
	}

	cacheStack, err := o.storage.ToBeCached()
	if err != nil {
		return err
	}

	h := &queryHeap{}
	heap.Init(h)

	for index := 0; index < len(keys); index++ {
		node := nodes[keys[index]]
		if node.Name == "" {
			continue
		}

		execute, err := graph.ParseAndExecute(args[0], o.storage, node.Name, nodes, caches, len(cacheStack) == 0)
		if err != nil {
			return err
		}

		output := execute.ToArray()
		heap.Push(h, &query{node: node, output: output})
		printProgress(index+1, len(nodes))
	}

	queries = make([]query, h.Len())
	for i := len(queries) - 1; i >= 0; i-- {
		queries[i] = *heap.Pop(h).(*query)
	}

	table := tablewriter.NewWriter(os.Stdout)

	if o.all {
		table.SetHeader([]string{"Name", "Type", "ID", "Query"})
	} else {
		table.SetHeader([]string{"Name", "Type", "ID", "QueryLength"})
	}

	for index, q := range queries {
		if index > o.maxOutput {
			break
		}
		if o.all {
			table.Append([]string{q.node.Name, q.node.Type, strconv.Itoa(int(q.node.ID)), fmt.Sprint(q.output)})
		} else {
			table.Append([]string{q.node.Name, q.node.Type, strconv.Itoa(int(q.node.ID)), fmt.Sprint(len(q.output))})
		}
	}

	table.Render()

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
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

type queryHeap []*query

func (h queryHeap) Len() int { return len(h) }
func (h queryHeap) Less(i, j int) bool {
	return len(h[i].output) < len(h[j].output)
}
func (h queryHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *queryHeap) Push(x interface{}) {
	*h = append(*h, x.(*query))
}

func (h *queryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func printProgress(progress, total int) {
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

	fmt.Printf("\r[%s] %s of the queries computed \033[1;34m(%d/%d)\033[0m", bar, percentage, progress, total)

	if progress == total {
		fmt.Println()
	}
}
