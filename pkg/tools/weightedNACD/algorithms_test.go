package weightedNACD

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSigmoidBasedAlgo(t *testing.T) {
	tests := []struct {
		name        string
		values      []valueAndType
		valsForType map[string]weightsForType
		want        float64
	}{
		{
			name: "simple case",
			values: []valueAndType{
				{value: 1.0, _type: "dependencies"},
			},
			valsForType: map[string]weightsForType{
				"dependencies": {weight: 1.0, k: 1.0, l: 1.0},
			},
			want: 0.6224593312018546,
		},
		{
			name: "multiple values",
			values: []valueAndType{
				{value: 3, _type: "dependencies"}, // There are 3 dependencies
				{value: 2.8, _type: "scorecard"},  // The OpenSSF Scorecard score is 7.2. We are subtracting the score from 10 to get 2.8
			},
			valsForType: map[string]weightsForType{
				"dependencies": {weight: 1.0, k: 1.0, l: 1.0},
				"scorecard":    {weight: 2.0, k: 1.0, l: 2.0},
			},
			want: 0.8801465633925935,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sigmoidBasedAlgo(test.values, test.valsForType)
			if diff := cmp.Diff(test.want, got, cmp.Comparer(compareWithNaNComparison)); diff != "" {
				t.Errorf("sigmoidBasedAlgo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRiskAlgo(t *testing.T) {
	tests := []struct {
		name        string
		weight1     float64
		criticality float64
		weight2     float64
		likelihood  float64
		want        float64
	}{
		{
			name:        "simple case",
			weight1:     0.5,
			criticality: 0.5,
			weight2:     0.5,
			likelihood:  0.5,
			want:        0.5,
		},
		{
			name:        "different weights",
			weight1:     0.7,
			criticality: 0.6,
			weight2:     0.3,
			likelihood:  0.4,
			want:        0.54,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := riskAlgo(test.weight1, test.criticality, test.weight2, test.likelihood)
			if diff := cmp.Diff(test.want, got, cmp.Comparer(compareWithNaNComparison)); diff != "" {
				t.Errorf("riskAlgo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
