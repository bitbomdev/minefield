package helpers

import (
	"encoding/json"
	"testing"

	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/pkg/tools"
	"github.com/bitbomdev/minefield/pkg/tools/ingest"
	"github.com/stretchr/testify/assert"
)

func TestComputeAdditionalInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    *apiv1.Node
		expected string
	}{
		{
			name: "Scorecard with checks",
			input: &apiv1.Node{
				Type: tools.ScorecardType,
				Metadata: mustMarshal(ingest.ScorecardResult{
					Scorecard: ingest.ScorecardData{
						Score: 8.5,
						Checks: []ingest.Check{
							{Name: "Security", Score: 9},
							{Name: "License", Score: 8},
						},
					},
				}),
			},
			expected: "Score: 8.50\nChecks:\n- Security: 9\n- License: 8",
		},
		{
			name: "Scorecard without checks",
			input: &apiv1.Node{
				Type: tools.ScorecardType,
				Metadata: mustMarshal(ingest.ScorecardResult{
					Scorecard: ingest.ScorecardData{
						Score: 7.0,
					},
				}),
			},
			expected: "Score: 7.00",
		},
		{
			name: "Vulnerability with fixed versions",
			input: &apiv1.Node{
				Type: tools.VulnerabilityType,
				Metadata: mustMarshal(ingest.Vulnerability{
					Affected: []ingest.Affected{
						{
							Package: ingest.Package{
								Purl: "pkg:npm/example@1.0.0",
							},
							Ranges: []ingest.Range{
								{
									Events: []ingest.Event{
										{Fixed: "1.0.1"},
									},
								},
							},
						},
					},
				}),
			},
			expected: "Affected Package PURL (Package URL) : Fixed Version\n\npkg:npm/example@1.0.0 : 1.0.1",
		},
		{
			name: "Node with nil metadata",
			input: &apiv1.Node{
				Type:     tools.ScorecardType,
				Metadata: nil,
			},
			expected: "",
		},
		{
			name: "Unknown node type",
			input: &apiv1.Node{
				Type: "unknown",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeAdditionalInfo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to marshal test data
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
