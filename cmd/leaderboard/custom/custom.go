package custom

import (
	"context"
	"fmt"
	"github.com/bitbomdev/minefield/cmd/helpers"
	"net/http"
	"os"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage   graph.Storage
	all       bool
	maxOutput int
	showInfo  bool // New field to control the display of the Info column
	saveQuery string
}

type query struct {
	node   *graph.Node
	output []uint32
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.all, "all", false, "show the queries getMetadata for each node")
	cmd.Flags().IntVar(&o.maxOutput, "max-getMetadata", 10, "max getMetadata length")
	cmd.Flags().BoolVar(&o.showInfo, "show-info", true, "display the info column")
	cmd.Flags().StringVar(&o.saveQuery, "save-query", "", "save the query to a specific file")
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
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	headers := []string{"Name", "Type", "ID", "Output"}
	if o.showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)

	var f *os.File
	if o.saveQuery != "" {
		f, err = os.Create(o.saveQuery)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	for index, q := range res.Msg.Queries {
		if index >= o.maxOutput {
			break
		}

		// Determine the Output value
		var output string
		if o.all {
			output = fmt.Sprint(q.Output)
		} else {
			output = fmt.Sprint(len(q.Output))
		}

		// Build the common row data
		row := []string{
			q.Node.Name,
			q.Node.Type,
			strconv.Itoa(int(q.Node.Id)),
			output,
		}

		// If showInfo is true, compute the additionalInfo and append it
		if o.showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(q.Node)
			row = append(row, additionalInfo)
		}

		// Append the row to the table
		table.Append(row)

		if o.saveQuery != "" {
			f.WriteString(q.Node.Name + "\n")
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
