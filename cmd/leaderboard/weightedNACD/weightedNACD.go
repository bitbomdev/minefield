package weightedNACD

import (
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-json"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools/weightedNACD"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type options struct {
	storage     graph.Storage
	weightsFile string
	maxOutput   int
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.weightsFile, "weights", "cmd/leaderboard/weightedNACD/defaultWeights.json", "path to the JSON file with weights (optional, default weights will be used if not provided)")
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "max output length")
}

func (o *options) validateWeights(weights *weightedNACD.Weights) error {
	if weights.CriticalityWeight == 0 {
		return fmt.Errorf("criticalityWeight is required")
	}
	if weights.LikelihoodWeight == 0 {
		return fmt.Errorf("likelihoodWeight is required")
	}
	if weights.Dependencies != nil {
		if weights.Dependencies.Weight == 0 || weights.Dependencies.K == 0 || weights.Dependencies.L == 0 {
			return fmt.Errorf("if dependencies is specified then all fields in dependencies are required")
		}
	}
	if weights.Scorecard != nil {
		if weights.Scorecard.Weight == 0 || weights.Scorecard.K == 0 || weights.Scorecard.L == 0 {
			return fmt.Errorf("if scorecard is specified then all fields in scorecard are required")
		}
	}
	return nil
}

func (o *options) Run(_ *cobra.Command, _ []string) error {
	uncachedNodes, err := o.storage.ToBeCached()
	if err != nil {
		return err
	}
	if len(uncachedNodes) != 0 {
		return fmt.Errorf("cannot use sorted leaderboards without caching")
	}

	file, err := os.Open(o.weightsFile)
	if err != nil {
		return fmt.Errorf("failed to open weights file: %w", err)
	}
	defer file.Close()

	var weights weightedNACD.Weights
	if err := json.NewDecoder(file).Decode(&weights); err != nil {
		return fmt.Errorf("failed to decode weights file: %w", err)
	}

	if err := o.validateWeights(&weights); err != nil {
		return err
	}

	results, err := weightedNACD.WeightedNACD(o.storage, weights)
	if err != nil {
		return fmt.Errorf("failed to calculate weighted NACD: %w", err)
	}

	// Print results as a table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Package", "Risk", "Criticality", "Likelihood"})

	for index, result := range results {
		if index > o.maxOutput {
			break
		}
		node, err := o.storage.GetNode(result.Id)
		if err != nil {
			fmt.Println("Failed to get node for ID:", err)
			continue
		}
		table.Append([]string{strconv.Itoa(int(node.ID)), fmt.Sprintf("%f", result.Risk), fmt.Sprintf("%f", result.Criticality), fmt.Sprintf("%f", result.Likelihood)})
	}

	table.Render()

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:   "weightedNACD",
		Short: "calculates the risk of all packages",
		Long: "calculates the risk of all packages, the risk of a package is based on https://docs.google.com/document/d/1Xb86MrKFQZQNq9rCQb08Dk1b5HU7nzLHkzfjBvbndeM/edit?usp=sharing" +
			"if a package doesn't have a risk (this is mainly because the OpenSSF Scorecard score is unavailable for the package), then we compare it to other packages based on its criticality. " +
			"If one package has a risk and the other one doesn't, then the one with the risk will always rank higher up in the leaderboard. ",
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}
