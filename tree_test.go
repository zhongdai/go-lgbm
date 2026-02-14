package lgbm

import (
	"math"
	"testing"
)

// TestPredictLeaf_NumericalSplitLeft tests a simple numerical split where
// the feature value is less than or equal to the threshold, going left.
func TestPredictLeaf_NumericalSplitLeft(t *testing.T) {
	// Tree: 2 leaves, 1 internal node
	// split_feature=0, threshold=0.5, decision_type=0 (numerical, missing→right)
	// left_child=-1 (leaf 0), right_child=-2 (leaf 1)
	// leaf_values=[1.0, 2.0], shrinkage=1.0
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{0},
		leftChildren:  []int{-1}, // leaf index 0
		rightChildren: []int{-2}, // leaf index 1
		leafValues:    []float64{1.0, 2.0},
		shrinkage:     1.0,
	}

	// Input: features=[0.3] → should go left (0.3 <= 0.5) → return 1.0
	result := tree.predictLeaf([]float64{0.3})
	expected := 1.0

	if result != expected {
		t.Errorf("predictLeaf([0.3]) = %f; want %f", result, expected)
	}
}

// TestPredictLeaf_NumericalSplitRight tests a simple numerical split where
// the feature value is greater than the threshold, going right.
func TestPredictLeaf_NumericalSplitRight(t *testing.T) {
	// Same tree as above
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{0},
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{1.0, 2.0},
		shrinkage:     1.0,
	}

	// Input: features=[0.7] → should go right (0.7 > 0.5) → return 2.0
	result := tree.predictLeaf([]float64{0.7})
	expected := 2.0

	if result != expected {
		t.Errorf("predictLeaf([0.7]) = %f; want %f", result, expected)
	}
}

// TestPredictLeaf_NaNGoesRight tests that NaN values go to the default
// direction (right) when decision_type bit 1 is 0.
func TestPredictLeaf_NaNGoesRight(t *testing.T) {
	// Same tree, decision_type=0 (bit1=0 → missing goes right)
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{0},
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{1.0, 2.0},
		shrinkage:     1.0,
	}

	// Input: features=[NaN] → should go right → return 2.0
	result := tree.predictLeaf([]float64{math.NaN()})
	expected := 2.0

	if result != expected {
		t.Errorf("predictLeaf([NaN]) = %f; want %f", result, expected)
	}
}

// TestPredictLeaf_NaNGoesLeft tests that NaN values go left when
// decision_type bit 1 is set (decision_type=2).
func TestPredictLeaf_NaNGoesLeft(t *testing.T) {
	// Same tree but decision_type=2 (bit1=1 → missing goes left)
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{2}, // bit 1 set
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{1.0, 2.0},
		shrinkage:     1.0,
	}

	// Input: features=[NaN] → should go left → return 1.0
	result := tree.predictLeaf([]float64{math.NaN()})
	expected := 1.0

	if result != expected {
		t.Errorf("predictLeaf([NaN]) = %f; want %f", result, expected)
	}
}

// TestPredictLeaf_CategoricalSplit tests categorical splits using bitsets.
func TestPredictLeaf_CategoricalSplit(t *testing.T) {
	// Tree: 2 leaves, 1 internal node
	// split_feature=0, threshold=0 (index into cat_threshold)
	// decision_type=1 (categorical)
	// cat_boundaries=[0,1], cat_threshold=[5] (binary: 00000101, categories 0 and 2 go left)
	// left_child=-1, right_child=-2
	// leaf_values=[10.0, 20.0]
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0},
		decisionTypes: []uint8{1}, // categorical
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{10.0, 20.0},
		shrinkage:     1.0,
		catBoundaries: []int{0, 1},
		catThresholds: []uint32{5}, // 00000101 in binary: bits 0 and 2 set
	}

	tests := []struct {
		category float64
		expected float64
		desc     string
	}{
		{0.0, 10.0, "category 0 in bitset → go left"},
		{1.0, 20.0, "category 1 not in bitset → go right"},
		{2.0, 10.0, "category 2 in bitset → go left"},
	}

	for _, tc := range tests {
		result := tree.predictLeaf([]float64{tc.category})
		if result != tc.expected {
			t.Errorf("predictLeaf([%f]): %s = %f; want %f",
				tc.category, tc.desc, result, tc.expected)
		}
	}
}

// TestPredictLeaf_DeeperTree tests a tree with multiple internal nodes.
func TestPredictLeaf_DeeperTree(t *testing.T) {
	// Tree structure:
	//        node0 (feature 0, threshold 0.5)
	//       /                              \
	//   node1 (feature 1, threshold 0.3)   leaf2 (value 4.0)
	//    /           \
	// leaf0 (1.0)   leaf1 (2.0)
	//
	// Internal nodes: 0, 1
	// Leaves: 0, 1, 2
	tree := &tree{
		numLeaves:     3,
		splitFeatures: []int{0, 1}, // node0 splits on feature 0, node1 splits on feature 1
		thresholds:    []float64{0.5, 0.3},
		decisionTypes: []uint8{0, 0}, // both numerical
		leftChildren:  []int{1, -1},  // node0→node1, node1→leaf0
		rightChildren: []int{-3, -2}, // node0→leaf2, node1→leaf1
		leafValues:    []float64{1.0, 2.0, 4.0},
		shrinkage:     1.0,
	}

	tests := []struct {
		features []float64
		expected float64
		desc     string
	}{
		{[]float64{0.3, 0.2}, 1.0, "go left at node0, left at node1 → leaf0"},
		{[]float64{0.3, 0.4}, 2.0, "go left at node0, right at node1 → leaf1"},
		{[]float64{0.7, 0.0}, 4.0, "go right at node0 → leaf2"},
	}

	for _, tc := range tests {
		result := tree.predictLeaf(tc.features)
		if result != tc.expected {
			t.Errorf("predictLeaf(%v): %s = %f; want %f",
				tc.features, tc.desc, result, tc.expected)
		}
	}
}

// T046: NaN/infinity edge case tests
func TestPredictLeaf_PositiveInfinity(t *testing.T) {
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{0},
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{1.0, 2.0},
		shrinkage:     1.0,
	}

	// +Inf > 0.5, should go right
	result := tree.predictLeaf([]float64{math.Inf(1)})
	if result != 2.0 {
		t.Errorf("predictLeaf([+Inf]) = %f; want 2.0", result)
	}
}

func TestPredictLeaf_NegativeInfinity(t *testing.T) {
	tree := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{0},
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{1.0, 2.0},
		shrinkage:     1.0,
	}

	// -Inf <= 0.5, should go left
	result := tree.predictLeaf([]float64{math.Inf(-1)})
	if result != 1.0 {
		t.Errorf("predictLeaf([-Inf]) = %f; want 1.0", result)
	}
}

// TestPredictLeaf_LeafValueNotMultipliedByShrinkage verifies that
// predictLeaf returns the raw leaf value without shrinkage multiplication.
// LightGBM text format leaf values already incorporate the learning rate.
func TestPredictLeaf_LeafValueNotMultipliedByShrinkage(t *testing.T) {
	tr := &tree{
		numLeaves:     2,
		splitFeatures: []int{0},
		thresholds:    []float64{0.5},
		decisionTypes: []uint8{0},
		leftChildren:  []int{-1},
		rightChildren: []int{-2},
		leafValues:    []float64{10.0, 20.0},
		shrinkage:     0.1,
	}

	// Input: features=[0.3] → go left → raw leaf value 10.0 (NOT 10.0 * 0.1)
	result := tr.predictLeaf([]float64{0.3})
	expected := 10.0

	if result != expected {
		t.Errorf("predictLeaf([0.3]) = %f; want %f (shrinkage must not be applied)", result, expected)
	}
}
