package helpers

import (
	"encoding/json"
	"fmt"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/bit-bom/minefield/pkg/tools"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"strings"
)

func ComputeAdditionalInfo(node *apiv1.Node) string {
	var additionalInfo string

	switch node.Type {
	case tools.ScorecardType:
		// Unmarshal metadata into ScorecardResult
		var scorecardResult ingest.ScorecardResult
		if node.Metadata != nil {
			metadataBytes := node.Metadata
			if err := json.Unmarshal(metadataBytes, &scorecardResult); err == nil {
				var scoreDetails []string
				scoreDetails = append(scoreDetails, fmt.Sprintf("Score: %.2f", scorecardResult.Scorecard.Score))

				if len(scorecardResult.Scorecard.Checks) > 0 {
					scoreDetails = append(scoreDetails, "Checks:")
					for _, check := range scorecardResult.Scorecard.Checks {
						scoreDetails = append(scoreDetails, fmt.Sprintf("- %s: %d", check.Name, check.Score))
					}
				}

				// Join all details with newlines
				additionalInfo = strings.Join(scoreDetails, "\n")
			}
		}
	case tools.VulnerabilityType:
		// Unmarshal metadata into Vulnerability
		var vulnerability ingest.Vulnerability
		if node.Metadata != nil {
			metadataBytes := node.Metadata
			if err := json.Unmarshal(metadataBytes, &vulnerability); err == nil {
				var fixedInfo []string
				for _, affected := range vulnerability.Affected {
					for _, r := range affected.Ranges {
						for _, event := range r.Events {
							if event.Fixed != "" {
								fixedInfo = append(fixedInfo, fmt.Sprintf("%s : %s", affected.Package.Purl, event.Fixed))
							}
						}
					}
				}
				if len(fixedInfo) > 0 {
					additionalInfo = "Affected Package PURL (Package URL) : Fixed Version\n\n" + strings.Join(fixedInfo, "\n")
				}
			}
		}
	}

	return additionalInfo
}
