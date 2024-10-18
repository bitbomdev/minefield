package custom

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/gen/api/v1/apiv1connect"
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
	script := strings.Join(args, " ")
	httpClient := &http.Client{}
	addr := os.Getenv("BITBOMDEV_ADDR")
	if addr == "" {
		addr = "http://localhost:8089"
	}
	client := apiv1connect.NewLeaderboardServiceClient(httpClient, addr)

	// Create a new context
	ctx := context.Background()

	// Create a new Leaderboard request
	req := connect.NewRequest(&apiv1.CustomLeaderboardRequest{
		Script: script,
	})

	// Make the Leaderboard request
	res, err := client.CustomLeaderboard(ctx, req)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)

	for index, q := range res.Msg.Queries {
		if index > o.maxOutput {
			break
		}
		if o.all {
			table.Append([]string{q.Node.Name, q.Node.Type, strconv.Itoa(int(q.Node.Id)), fmt.Sprint(q.Output)})
		} else {
			table.Append([]string{q.Node.Name, q.Node.Type, strconv.Itoa(int(q.Node.Id)), fmt.Sprint(len(q.Output))})
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
		Args:              cobra.MinimumNArgs(1),
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
