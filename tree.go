package lgbm

import "math"

// tree represents a single decision tree in a LightGBM ensemble.
// It stores the tree structure (splits and leaves) and supports
// both numerical and categorical split conditions.
type tree struct {
	// numLeaves is the count of leaf nodes in this tree.
	numLeaves int

	// splitFeatures[i] is the feature index used for the split at internal node i.
	splitFeatures []int

	// thresholds[i] is the threshold value for the split at internal node i.
	// For numerical splits: go left if feature_value <= threshold.
	// For categorical splits: threshold is unused (check catThresholds instead).
	thresholds []float64

	// decisionTypes[i] encodes the split type for internal node i.
	// 0 = numerical (<=), 1 = categorical (bitset membership).
	decisionTypes []uint8

	// leftChildren[i] is the index of the left child of internal node i.
	// Negative values indicate leaf nodes: actual leaf index = ^leftChildren[i].
	leftChildren []int

	// rightChildren[i] is the index of the right child of internal node i.
	// Negative values indicate leaf nodes: actual leaf index = ^rightChildren[i].
	rightChildren []int

	// leafValues[i] is the output value of leaf i.
	leafValues []float64

	// shrinkage is the learning rate multiplier applied to this tree's output.
	shrinkage float64

	// catBoundaries[i] and catBoundaries[i+1] define the range in catThresholds
	// for the categorical split at internal node i. Only used for categorical splits.
	catBoundaries []int

	// catThresholds stores concatenated bitsets for all categorical splits.
	// Each uint32 represents 32 categories. A set bit means "go left" for that category.
	catThresholds []uint32
}

// predictLeaf traverses the tree with the given feature values and
// returns the leaf value. The leaf values in the LightGBM text format
// already incorporate the learning rate, so no shrinkage multiplication
// is applied during prediction.
func (t *tree) predictLeaf(features []float64) float64 {
	node := 0 // Start at root

	// Traverse tree until we reach a leaf (negative node index)
	for node >= 0 {
		featureIdx := t.splitFeatures[node]
		val := features[featureIdx]

		var goLeft bool

		// Handle NaN values
		if math.IsNaN(val) {
			// If decision_type bit 1 is set (& 2 != 0), missing goes left
			// Otherwise, missing goes right
			goLeft = (t.decisionTypes[node] & 2) != 0
		} else if (t.decisionTypes[node] & 1) != 0 {
			// Categorical split
			category := int(val)
			catIdx := int(t.thresholds[node])
			start := t.catBoundaries[catIdx]
			end := t.catBoundaries[catIdx+1]

			// Check if category bit is set in bitset
			goLeft = isCategoryInBitset(category, t.catThresholds[start:end])
		} else {
			// Numerical split
			goLeft = val <= t.thresholds[node]
		}

		// Navigate to next node
		if goLeft {
			node = t.leftChildren[node]
		} else {
			node = t.rightChildren[node]
		}
	}

	// node is negative, so leaf index = -(node + 1)
	leafIdx := -(node + 1)
	return t.leafValues[leafIdx]
}

// isCategoryInBitset checks if a category is set in the bitset.
// bitset is a slice of uint32, where each uint32 represents 32 categories.
func isCategoryInBitset(category int, bitset []uint32) bool {
	wordIdx := category / 32
	bitIdx := uint(category % 32)

	if wordIdx >= len(bitset) {
		return false
	}

	return (bitset[wordIdx] & (1 << bitIdx)) != 0
}
