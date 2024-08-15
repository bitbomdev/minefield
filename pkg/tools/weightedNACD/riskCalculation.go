package weightedNACD

import (
	"fmt"
	"math"
	"sort"

	"github.com/bit-bom/minefield/pkg/storage"
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
	CriticalityWeight float64 `json:"criticalityWeight"`
	LikelihoodWeight  float64 `json:"likelihoodWeight"`
	Dependencies      *struct {
		Weight float64 `json:"weight"`
		K      float64 `json:"k"`
		L      float64 `json:"l"`
	} `json:"dependencies,omitempty"`
	Scorecard *struct {
		Weight float64 `json:"weight"`
		K      float64 `json:"k"`
		L      float64 `json:"l"`
	} `json:"scorecard,omitempty"`
}

func WeightedNACD(storage storage.Storage, weights Weights) ([]*PkgAndValue, error) {
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

	for _, id := range ids {
		node, err := storage.GetNode(id)
		if err != nil {
			return nil, fmt.Errorf("error getting node with Id %d: %w", id, err)
		}

		// We can really only calculate the algo on package nodes
		if node.Type == "PACKAGE" {
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

			// Round values to the second decimal place
			criticality = math.Round(criticality*100) / 100
			likelihood = math.Round(likelihood*100) / 100

			risk := riskAlgo(weights.CriticalityWeight, criticality, weights.LikelihoodWeight, likelihood)
			risk = math.Round(risk*100) / 100

			scoresPerPkg = append(scoresPerPkg, &PkgAndValue{Id: id, Risk: risk, Criticality: criticality, Likelihood: likelihood})
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
