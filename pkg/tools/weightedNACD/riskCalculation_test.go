package weightedNACD

import (
	"math"
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

// compareWithNaNComparison is a custom comparer for float64 values that treats NaN values as equal.
func compareWithNaNComparison(x, y float64) bool {
	if math.IsNaN(x) && math.IsNaN(y) {
		return true
	}
	return x == y
}

func TestWeightedNACD(t *testing.T) {
	tests := []struct {
		name    string
		weights Weights
		want    []*PkgAndValue
		wantErr bool
	}{
		{
			name: "default",
			weights: Weights{
				CriticalityWeight: 0.5,
				LikelihoodWeight:  0.5,
				Dependencies: &struct {
					Weight float64 `json:"weight"`
					K      float64 `json:"k"`
					L      float64 `json:"l"`
				}{
					Weight: 1.0,
					K:      1.0,
					L:      1.0,
				},
			},
			want: []*PkgAndValue{
				{Id: 1, Risk: math.NaN(), Criticality: 0.8175744761936437, Likelihood: math.NaN()},
				{Id: 2, Risk: math.NaN(), Criticality: 0.6224593312018546, Likelihood: math.NaN()},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := graph.NewMockStorage()
			// Add mock nodes to storages
			node1, err := graph.AddNode(storage, "library", "metadata1", "pkg:generic/dep1@1.0.0")
			assert.NoError(t, err)
			node2, err := graph.AddNode(storage, "library", "metadata2", "pkg:generic/dep2@1.0.0")
			assert.NoError(t, err)

			// Set dependencies
			err = node1.SetDependency(storage, node2)
			assert.NoError(t, err)

			got, err := WeightedNACD(storage, test.weights, nil)
			if (err != nil) != test.wantErr {
				t.Errorf("WeightedNACD() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if diff := cmp.Diff(test.want, got, cmp.Comparer(compareWithNaNComparison)); diff != "" {
				t.Errorf("WeightedNACD() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
