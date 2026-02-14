package lgbm

import (
	"math"
	"strings"
)

// ObjectiveType identifies the LightGBM training objective.
type ObjectiveType int

const (
	ObjectiveBinary     ObjectiveType = iota // binary classification
	ObjectiveRegression                      // regression (L2, L1, huber, fair, etc.)
	ObjectiveMulticlass                      // multiclass / multiclassova
	ObjectiveRanking                         // lambdarank, rank_xendcg
	ObjectivePoisson                         // poisson regression
	ObjectiveGamma                           // gamma regression
	ObjectiveTweedie                         // tweedie regression
)

// TransformFunc applies a post-prediction transformation to raw tree
// scores. The function receives raw scores and writes transformed
// values into out. raw and out may alias for in-place transforms.
type TransformFunc func(raw []float64, out []float64)

// parseObjective maps the objective string from a model header to an
// ObjectiveType. The objective string may contain parameters after the
// name (e.g. "binary sigmoid:1", "multiclass num_class:3").
func parseObjective(s string) (ObjectiveType, error) {
	name := strings.Fields(s)
	if len(name) == 0 {
		return ObjectiveRegression, nil // default to regression if empty
	}

	switch strings.ToLower(name[0]) {
	case "binary", "cross_entropy":
		return ObjectiveBinary, nil
	case "multiclass", "multiclassova", "multi_logloss", "softmax",
		"multiclass_ova", "ova", "ovr":
		return ObjectiveMulticlass, nil
	case "lambdarank", "rank_xendcg", "rank":
		return ObjectiveRanking, nil
	case "poisson":
		return ObjectivePoisson, nil
	case "gamma":
		return ObjectiveGamma, nil
	case "tweedie":
		return ObjectiveTweedie, nil
	case "regression", "regression_l2", "regression_l1",
		"mean_squared_error", "mse", "l2", "l1",
		"mean_absolute_error", "mae",
		"huber", "fair", "quantile", "mape",
		"custom":
		return ObjectiveRegression, nil
	default:
		// Unknown objectives default to regression (raw output).
		return ObjectiveRegression, nil
	}
}

// transformForObjective returns the appropriate TransformFunc for the
// given objective type.
func transformForObjective(obj ObjectiveType) TransformFunc {
	switch obj {
	case ObjectiveBinary:
		return transformSigmoid
	case ObjectiveMulticlass:
		return transformSoftmax
	case ObjectivePoisson, ObjectiveGamma, ObjectiveTweedie:
		return transformExponential
	default:
		return transformIdentity
	}
}

// transformIdentity copies raw scores to output unchanged.
func transformIdentity(raw []float64, out []float64) {
	copy(out, raw)
}

// transformSigmoid applies the logistic sigmoid: 1/(1+exp(-x)).
func transformSigmoid(raw []float64, out []float64) {
	out[0] = sigmoid(raw[0])
}

// transformSoftmax applies softmax normalization across all classes.
func transformSoftmax(raw []float64, out []float64) {
	maxVal := raw[0]
	for _, v := range raw[1:] {
		if v > maxVal {
			maxVal = v
		}
	}

	var sum float64
	for i, v := range raw {
		out[i] = math.Exp(v - maxVal)
		sum += out[i]
	}
	for i := range out[:len(raw)] {
		out[i] /= sum
	}
}

// transformExponential applies exp(x) to each raw score.
func transformExponential(raw []float64, out []float64) {
	for i, v := range raw {
		out[i] = math.Exp(v)
	}
}

// sigmoid computes 1/(1+exp(-x)).
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}
