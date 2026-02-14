package lgbm

import (
	"math"
	"testing"
)

const epsilon = 1e-9

// TestSigmoid tests the sigmoid function directly.
func TestSigmoid(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
		desc     string
	}{
		{0.0, 0.5, "sigmoid(0) should be 0.5"},
		{10.0, 0.9999546021312976, "sigmoid(large positive) should be close to 1.0"},
		{-10.0, 0.00004539786870243442, "sigmoid(large negative) should be close to 0.0"},
	}

	for _, tc := range tests {
		result := sigmoid(tc.input)
		if math.Abs(result-tc.expected) > epsilon {
			t.Errorf("%s: sigmoid(%f) = %f; want %f",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestTransformSigmoid tests the transformSigmoid function.
func TestTransformSigmoid(t *testing.T) {
	raw := []float64{0.0}
	out := make([]float64, 1)

	transformSigmoid(raw, out)

	expected := 0.5
	if math.Abs(out[0]-expected) > epsilon {
		t.Errorf("transformSigmoid([0.0]) = %f; want %f", out[0], expected)
	}
}

// TestTransformSoftmax tests the softmax transformation.
func TestTransformSoftmax(t *testing.T) {
	t.Run("standard inputs", func(t *testing.T) {
		raw := []float64{1.0, 2.0, 3.0}
		out := make([]float64, 3)

		transformSoftmax(raw, out)

		// Verify sum is approximately 1.0
		sum := 0.0
		for _, v := range out {
			sum += v
		}
		if math.Abs(sum-1.0) > epsilon {
			t.Errorf("sum of softmax outputs = %f; want 1.0", sum)
		}

		// Verify ordering is preserved (higher input → higher output)
		if out[0] >= out[1] || out[1] >= out[2] {
			t.Errorf("softmax should preserve ordering: got %v", out)
		}
	})

	t.Run("equal inputs", func(t *testing.T) {
		raw := []float64{1.0, 1.0, 1.0}
		out := make([]float64, 3)

		transformSoftmax(raw, out)

		expected := 1.0 / 3.0
		for i, v := range out {
			if math.Abs(v-expected) > epsilon {
				t.Errorf("softmax([1,1,1])[%d] = %f; want %f", i, v, expected)
			}
		}
	})

	t.Run("large values without overflow", func(t *testing.T) {
		raw := []float64{1000.0, 1001.0, 1002.0}
		out := make([]float64, 3)

		transformSoftmax(raw, out)

		// Should not produce NaN or Inf
		for i, v := range out {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("softmax produced invalid value at index %d: %f", i, v)
			}
		}

		// Sum should still be 1.0
		sum := 0.0
		for _, v := range out {
			sum += v
		}
		if math.Abs(sum-1.0) > epsilon {
			t.Errorf("sum of softmax outputs with large values = %f; want 1.0", sum)
		}
	})
}

// TestTransformIdentity tests the identity transformation.
func TestTransformIdentity(t *testing.T) {
	raw := []float64{1.5, 2.5, 3.5}
	out := make([]float64, 3)

	transformIdentity(raw, out)

	for i := range raw {
		if out[i] != raw[i] {
			t.Errorf("transformIdentity: out[%d] = %f; want %f", i, out[i], raw[i])
		}
	}
}

// TestTransformExponential tests the exponential transformation.
func TestTransformExponential(t *testing.T) {
	tests := []struct {
		raw      []float64
		expected []float64
		desc     string
	}{
		{
			raw:      []float64{0.0},
			expected: []float64{1.0},
			desc:     "exp(0) should be 1.0",
		},
		{
			raw:      []float64{1.0},
			expected: []float64{math.E},
			desc:     "exp(1) should be e",
		},
		{
			raw:      []float64{2.0, 3.0},
			expected: []float64{math.Exp(2.0), math.Exp(3.0)},
			desc:     "exp on multiple values",
		},
	}

	for _, tc := range tests {
		out := make([]float64, len(tc.raw))
		transformExponential(tc.raw, out)

		for i := range tc.expected {
			if math.Abs(out[i]-tc.expected[i]) > epsilon {
				t.Errorf("%s: out[%d] = %f; want %f",
					tc.desc, i, out[i], tc.expected[i])
			}
		}
	}
}

// TestParseObjective tests the objective string parsing.
func TestParseObjective(t *testing.T) {
	tests := []struct {
		input    string
		expected ObjectiveType
		desc     string
	}{
		{"binary", ObjectiveBinary, "binary objective"},
		{"binary sigmoid:1", ObjectiveBinary, "binary with parameters"},
		{"cross_entropy", ObjectiveBinary, "cross_entropy alias"},
		{"regression", ObjectiveRegression, "regression objective"},
		{"regression_l2", ObjectiveRegression, "regression_l2 variant"},
		{"multiclass", ObjectiveMulticlass, "multiclass objective"},
		{"multiclass num_class:3", ObjectiveMulticlass, "multiclass with parameters"},
		{"softmax", ObjectiveMulticlass, "softmax alias"},
		{"lambdarank", ObjectiveRanking, "lambdarank objective"},
		{"rank_xendcg", ObjectiveRanking, "rank_xendcg alias"},
		{"poisson", ObjectivePoisson, "poisson objective"},
		{"gamma", ObjectiveGamma, "gamma objective"},
		{"tweedie", ObjectiveTweedie, "tweedie objective"},
		{"", ObjectiveRegression, "empty string defaults to regression"},
		{"unknown", ObjectiveRegression, "unknown objective defaults to regression"},
	}

	for _, tc := range tests {
		result, err := parseObjective(tc.input)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.desc, err)
		}
		if result != tc.expected {
			t.Errorf("%s: parseObjective(%q) = %v; want %v",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestTransformForObjective tests that the correct transform is returned
// for each objective type.
func TestTransformForObjective(t *testing.T) {
	t.Run("binary objective applies sigmoid", func(t *testing.T) {
		transform := transformForObjective(ObjectiveBinary)
		raw := []float64{0.0}
		out := make([]float64, 1)

		transform(raw, out)

		expected := 0.5
		if math.Abs(out[0]-expected) > epsilon {
			t.Errorf("binary transform: out[0] = %f; want %f", out[0], expected)
		}
	})

	t.Run("regression objective applies identity", func(t *testing.T) {
		transform := transformForObjective(ObjectiveRegression)
		raw := []float64{1.5, 2.5}
		out := make([]float64, 2)

		transform(raw, out)

		for i := range raw {
			if out[i] != raw[i] {
				t.Errorf("regression transform: out[%d] = %f; want %f",
					i, out[i], raw[i])
			}
		}
	})

	t.Run("multiclass objective applies softmax", func(t *testing.T) {
		transform := transformForObjective(ObjectiveMulticlass)
		raw := []float64{1.0, 2.0, 3.0}
		out := make([]float64, 3)

		transform(raw, out)

		sum := 0.0
		for _, v := range out {
			sum += v
		}
		if math.Abs(sum-1.0) > epsilon {
			t.Errorf("multiclass transform: sum = %f; want 1.0", sum)
		}
	})

	t.Run("poisson objective applies exponential", func(t *testing.T) {
		transform := transformForObjective(ObjectivePoisson)
		raw := []float64{0.0}
		out := make([]float64, 1)

		transform(raw, out)

		expected := 1.0
		if math.Abs(out[0]-expected) > epsilon {
			t.Errorf("poisson transform: out[0] = %f; want %f", out[0], expected)
		}
	})

	t.Run("ranking objective applies identity", func(t *testing.T) {
		transform := transformForObjective(ObjectiveRanking)
		raw := []float64{1.5}
		out := make([]float64, 1)

		transform(raw, out)

		if out[0] != raw[0] {
			t.Errorf("ranking transform: out[0] = %f; want %f", out[0], raw[0])
		}
	})
}
