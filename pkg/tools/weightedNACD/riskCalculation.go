package weightedNACD

import (
	"fmt"
	"math"
	"sort"

	"github.com/bitbomdev/minefield/pkg/graph"
)

const (
	dependencies = "dependencies"
	scorecard    = "scorecard"
)

type PkgAndValue struct {
	Id          uint32
	Risk        float64
	Criticality float64
	Likelihood  float64
}

type Weights struct {
	Dependencies *struct {
		Weight float64 `json:"weight"`
		K      float64 `json:"k"`
		L      float64 `json:"l"`
	} `json:"dependencies,omitempty"`
	Scorecard *struct {
		Weight float64 `json:"weight"`
		K      float64 `json:"k"`
		L      float64 `json:"l"`
	} `json:"scorecard,omitempty"`
	CriticalityWeight float64 `json:"criticalityWeight"`
	LikelihoodWeight  float64 `json:"likelihoodWeight"`
}

func WeightedNACD(storage graph.Storage, weights Weights, progress func(int, int)) ([]*PkgAndValue, error) {
	weightsForEachType := map[string]weightsForType{}

	if weights.Dependencies != nil {
		weightsForEachType[dependencies] = weightsForType{
			weight: weights.Dependencies.Weight,
			k:      weights.Dependencies.K,
			l:      weights.Dependencies.L,
		}
	}

	if weights.Scorecard != nil {
		weightsForEachType[scorecard] = weightsForType{
			weight: weights.Scorecard.Weight,
			k:      weights.Scorecard.K,
			l:      weights.Scorecard.L,
		}
	}

	var scoresPerPkg []*PkgAndValue

	ids, err := storage.GetAllKeys()
	if err != nil {
		return nil, fmt.Errorf("error getting all leaderboard: %w", err)
	}

	for index, id := range ids {
		node, err := storage.GetNode(id)
		if err != nil {
			return nil, fmt.Errorf("error getting node with Id %d: %w", id, err)
		}

		// We can really only calculate the algo on package nodes
		if node.Type == "library" {
			deps, err := node.QueryDependencies(storage)
			if err != nil {
				return nil, fmt.Errorf("error querying dependencies for node with Id %d: %w", id, err)
			}

			var valsAndTypesForCriticality []valueAndType
			var valsAndTypesForLikelihood []valueAndType

			valsAndTypesForCriticality = append(valsAndTypesForCriticality, valueAndType{value: float64(len(deps.ToArray())), _type: dependencies})
			// TODO: Add the Scorecard data to the likelihood (The Scorecard score has to be subtracted from 10)

			criticality := sigmoidBasedAlgo(valsAndTypesForCriticality, weightsForEachType)
			likelihood := sigmoidBasedAlgo(valsAndTypesForLikelihood, weightsForEachType)

			risk := riskAlgo(weights.CriticalityWeight, criticality, weights.LikelihoodWeight, likelihood)

			scoresPerPkg = append(scoresPerPkg, &PkgAndValue{Id: id, Risk: risk, Criticality: criticality, Likelihood: likelihood})
		}

		// Update progress
		if progress != nil {
			progress(index+1, len(ids))
		}
	}

	sort.Slice(scoresPerPkg, func(i, j int) bool {
		if math.IsNaN(scoresPerPkg[i].Risk) && !math.IsNaN(scoresPerPkg[j].Risk) {
			return false
		}
		if !math.IsNaN(scoresPerPkg[i].Risk) && math.IsNaN(scoresPerPkg[j].Risk) {
			return true
		}
		if math.IsNaN(scoresPerPkg[i].Risk) && math.IsNaN(scoresPerPkg[j].Risk) {
			return scoresPerPkg[i].Criticality > scoresPerPkg[j].Criticality
		}
		return scoresPerPkg[i].Risk > scoresPerPkg[j].Risk
	})

	return scoresPerPkg, nil
}
