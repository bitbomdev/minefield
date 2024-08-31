package weightedNACD

import "math"

type weightsForType struct {
	weight, k, l float64
}

type valueAndType struct {
	_type string
	value float64
}

// sigmoidBasedAlgo is based on this alternative algorithm https://docs.google.com/document/d/1Xb86MrKFQZQNq9rCQb08Dk1b5HU7nzLHkzfjBvbndeM/edit?pli=1#heading=h.n6yhyn5znyw3
func sigmoidBasedAlgo(values []valueAndType, valsForType map[string]weightsForType) float64 {
	sum := float64(0)
	for _, value := range values {
		sum += valsForType[value._type].weight
	}

	sum2 := float64(0)
	for i := range values {
		vals := valsForType[values[i]._type]
		sum2 += vals.weight / (1 + math.Pow(math.E, -vals.k*(values[i].value-vals.l/2)))
	}

	return 1 / sum * sum2
}

// Risk is because of this https://docs.google.com/document/d/1Xb86MrKFQZQNq9rCQb08Dk1b5HU7nzLHkzfjBvbndeM/edit?pli=1#heading=h.z6k227z3me9k
func riskAlgo(weight1, criticality, weight2, likelihood float64) float64 {
	return weight1*criticality + weight2*likelihood
}
