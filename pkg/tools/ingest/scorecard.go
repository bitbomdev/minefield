package ingest

import (
	"encoding/json"
	"fmt"

	"strings"

	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/bitbomdev/minefield/pkg/tools"
	"github.com/package-url/packageurl-go"
)

type Repo struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

type Scorecard struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type ScorecardData struct {
	Date      string    `json:"date"`
	Repo      Repo      `json:"repo"`
	Scorecard Scorecard `json:"scorecard"`
	Score     float64   `json:"score"`
	Checks    []Check
	PURL      string `json:"purl"`
}

type Check struct {
	Name   string
	Score  int
	Reason string
}

type ScorecardResult struct {
	PURL      string        `json:"purl"`
	Success   bool          `json:"success,omitempty"`
	Scorecard ScorecardData `json:"scorecard,omitempty"`
	Error     string        `json:"error,omitempty"`
	GitHubURL string        `json:"github_url,omitempty"`
}

// Scorecard processes the Scorecard JSON data and stores it in the graph.
func Scorecards(storage graph.Storage, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	var results []ScorecardResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("failed to decode Scorecard data: %w", err)
	}

	scorecardResults := map[string][]ScorecardResult{}

	for _, result := range results {
		if !result.Success {
			continue
		}
		purl, err := packageurl.FromString(result.PURL)
		if err != nil {
			return err
		}

		scorecardResults[purl.Name] = append(scorecardResults[purl.Name], result)
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		return err
	}

	nodes, err := storage.GetNodes(keys)
	if err != nil {
		return fmt.Errorf("failed to get nodes from storage: %w", err)
	}

	for _, node := range nodes {
		if node.Type == tools.LibraryType && strings.HasPrefix(node.Name, pkg) {
			purl, err := PURLToPackage(node.Name)
			if err != nil {
				continue
			}

			scorecardData, ok := scorecardResults[purl.Name]
			if !ok {
				continue
			}

			for _, scorecardResult := range scorecardData {

				if scorecardResult.Success {
					scorecardPurl, err := PURLToPackage(scorecardResult.PURL)
					if err != nil {
						continue
					}

					// The scorecard data is found based on the packages name, but then we need
					// to check whether the scorecard data is for the current packages version
					if scorecardPurl.Version == purl.Version {
						scorecardNode, err := graph.AddNode(storage, tools.ScorecardType, scorecardResult, getScorecardNodeName(scorecardResult.PURL))
						if err != nil {
							return fmt.Errorf("failed to add Scorecard node to storage: %w", err)
						}

						if err := node.SetDependency(storage, scorecardNode); err != nil {
							return fmt.Errorf("failed to add dependency edge to Scorecard node: %w", err)
						}
					}
				}
			}
		}
	}

	return nil
}

func getScorecardNodeName(name string) string {
	return "scorecard:" + name
}
