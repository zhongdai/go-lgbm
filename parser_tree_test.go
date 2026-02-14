package lgbm

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

func TestParseTree_SimpleNumericalTree(t *testing.T) {
	input := `num_leaves=4
num_cat=0
split_feature=1 2 0
split_gain=10.5 8.3 5.2
threshold=0.5 1.5 2.5
decision_type=2 2 2
left_child=1 -1 -3
right_child=2 -2 -4
leaf_value=0.1 0.2 0.3 0.4
leaf_weight=10.0 20.0 30.0 40.0
leaf_count=5 10 15 20
internal_value=0.15 0.25
internal_weight=30.0 40.0
internal_count=25 35
is_linear=0
shrinkage=1

`

	scanner := bufio.NewScanner(strings.NewReader(input))
	tr, err := parseTree(scanner)
	if err != nil {
		t.Fatalf("parseTree() error = %v", err)
	}

	if tr.numLeaves != 4 {
		t.Errorf("numLeaves = %d, want 4", tr.numLeaves)
	}

	if len(tr.splitFeatures) != 3 {
		t.Errorf("len(splitFeatures) = %d, want 3", len(tr.splitFeatures))
	}
	if tr.splitFeatures[0] != 1 || tr.splitFeatures[1] != 2 || tr.splitFeatures[2] != 0 {
		t.Errorf("splitFeatures = %v, want [1 2 0]", tr.splitFeatures)
	}

	if len(tr.thresholds) != 3 {
		t.Errorf("len(thresholds) = %d, want 3", len(tr.thresholds))
	}
	if tr.thresholds[0] != 0.5 || tr.thresholds[1] != 1.5 || tr.thresholds[2] != 2.5 {
		t.Errorf("thresholds = %v, want [0.5 1.5 2.5]", tr.thresholds)
	}

	if len(tr.decisionTypes) != 3 {
		t.Errorf("len(decisionTypes) = %d, want 3", len(tr.decisionTypes))
	}
	if tr.decisionTypes[0] != 2 || tr.decisionTypes[1] != 2 || tr.decisionTypes[2] != 2 {
		t.Errorf("decisionTypes = %v, want [2 2 2]", tr.decisionTypes)
	}

	if len(tr.leftChildren) != 3 {
		t.Errorf("len(leftChildren) = %d, want 3", len(tr.leftChildren))
	}
	if tr.leftChildren[0] != 1 || tr.leftChildren[1] != -1 || tr.leftChildren[2] != -3 {
		t.Errorf("leftChildren = %v, want [1 -1 -3]", tr.leftChildren)
	}

	if len(tr.rightChildren) != 3 {
		t.Errorf("len(rightChildren) = %d, want 3", len(tr.rightChildren))
	}
	if tr.rightChildren[0] != 2 || tr.rightChildren[1] != -2 || tr.rightChildren[2] != -4 {
		t.Errorf("rightChildren = %v, want [2 -2 -4]", tr.rightChildren)
	}

	if len(tr.leafValues) != 4 {
		t.Errorf("len(leafValues) = %d, want 4", len(tr.leafValues))
	}
	if tr.leafValues[0] != 0.1 || tr.leafValues[1] != 0.2 || tr.leafValues[2] != 0.3 || tr.leafValues[3] != 0.4 {
		t.Errorf("leafValues = %v, want [0.1 0.2 0.3 0.4]", tr.leafValues)
	}

	if tr.shrinkage != 1.0 {
		t.Errorf("shrinkage = %f, want 1.0", tr.shrinkage)
	}

	if len(tr.catBoundaries) != 0 {
		t.Errorf("len(catBoundaries) = %d, want 0", len(tr.catBoundaries))
	}

	if len(tr.catThresholds) != 0 {
		t.Errorf("len(catThresholds) = %d, want 0", len(tr.catThresholds))
	}
}

func TestParseTree_WithCategoricalSplits(t *testing.T) {
	input := `num_leaves=3
num_cat=1
split_feature=5 2
split_gain=15.5 10.2
threshold=0.0 1.5
decision_type=1 2
left_child=-1 -2
right_child=1 -3
leaf_value=0.5 0.6 0.7
leaf_weight=100.0 200.0 300.0
leaf_count=50 100 150
internal_value=0.55
internal_weight=500.0
internal_count=300
is_linear=0
shrinkage=0.5
cat_boundaries=0 2
cat_threshold=15 255

`

	scanner := bufio.NewScanner(strings.NewReader(input))
	tr, err := parseTree(scanner)
	if err != nil {
		t.Fatalf("parseTree() error = %v", err)
	}

	if tr.numLeaves != 3 {
		t.Errorf("numLeaves = %d, want 3", tr.numLeaves)
	}

	if len(tr.splitFeatures) != 2 {
		t.Errorf("len(splitFeatures) = %d, want 2", len(tr.splitFeatures))
	}

	if len(tr.leafValues) != 3 {
		t.Errorf("len(leafValues) = %d, want 3", len(tr.leafValues))
	}

	if tr.shrinkage != 0.5 {
		t.Errorf("shrinkage = %f, want 0.5", tr.shrinkage)
	}

	if len(tr.catBoundaries) != 2 {
		t.Errorf("len(catBoundaries) = %d, want 2", len(tr.catBoundaries))
	}
	if tr.catBoundaries[0] != 0 || tr.catBoundaries[1] != 2 {
		t.Errorf("catBoundaries = %v, want [0 2]", tr.catBoundaries)
	}

	if len(tr.catThresholds) != 2 {
		t.Errorf("len(catThresholds) = %d, want 2", len(tr.catThresholds))
	}
	if tr.catThresholds[0] != 15 || tr.catThresholds[1] != 255 {
		t.Errorf("catThresholds = %v, want [15 255]", tr.catThresholds)
	}
}

func TestParseTree_SingleLeaf(t *testing.T) {
	input := `num_leaves=1
num_cat=0
split_feature=
split_gain=
threshold=
decision_type=
left_child=
right_child=
leaf_value=0.123
leaf_weight=50.0
leaf_count=100
internal_value=
internal_weight=
internal_count=
is_linear=0
shrinkage=1.0

`

	scanner := bufio.NewScanner(strings.NewReader(input))
	tr, err := parseTree(scanner)
	if err != nil {
		t.Fatalf("parseTree() error = %v", err)
	}

	if tr.numLeaves != 1 {
		t.Errorf("numLeaves = %d, want 1", tr.numLeaves)
	}

	if len(tr.splitFeatures) != 0 {
		t.Errorf("len(splitFeatures) = %d, want 0", len(tr.splitFeatures))
	}

	if len(tr.leafValues) != 1 {
		t.Errorf("len(leafValues) = %d, want 1", len(tr.leafValues))
	}

	if tr.leafValues[0] != 0.123 {
		t.Errorf("leafValues[0] = %f, want 0.123", tr.leafValues[0])
	}
}

func TestParseTree_InvalidLeafValueCount(t *testing.T) {
	input := `num_leaves=4
num_cat=0
split_feature=1 2 0
split_gain=10.5 8.3 5.2
threshold=0.5 1.5 2.5
decision_type=2 2 2
left_child=1 -1 -3
right_child=2 -2 -4
leaf_value=0.1 0.2 0.3
leaf_weight=10.0 20.0 30.0
leaf_count=5 10 15
internal_value=0.15 0.25
internal_weight=30.0 40.0
internal_count=25 35
is_linear=0
shrinkage=1

`

	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseTree(scanner)
	if err == nil {
		t.Fatal("parseTree() expected error for mismatched leaf_value count, got nil")
	}

	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("parseTree() error = %v, want ErrInvalidModel", err)
	}
}

func TestParseTree_InvalidSplitFeatureCount(t *testing.T) {
	input := `num_leaves=4
num_cat=0
split_feature=1 2
split_gain=10.5 8.3
threshold=0.5 1.5
decision_type=2 2
left_child=1 -1
right_child=2 -2
leaf_value=0.1 0.2 0.3 0.4
leaf_weight=10.0 20.0 30.0 40.0
leaf_count=5 10 15 20
internal_value=0.15 0.25
internal_weight=30.0 40.0
internal_count=25 35
is_linear=0
shrinkage=1

`

	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseTree(scanner)
	if err == nil {
		t.Fatal("parseTree() expected error for mismatched split_feature count, got nil")
	}

	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("parseTree() error = %v, want ErrInvalidModel", err)
	}
}

func TestParseTree_RealWorldExample(t *testing.T) {
	// Example from actual LightGBM v4 model file
	input := `num_leaves=6
num_cat=0
split_feature=1 0 0 1 0
split_gain=63.6598 57.4799 21.4371 3.42323 1.42109e-14
threshold=-0.15407353588631145 -0.56557689594403493 0.5049099755936578 0.2261832294637203 0.56460871160744486
decision_type=2 2 2 2 2
left_child=2 -2 -1 -3 -5
right_child=1 3 -4 4 -6
leaf_value=-0.16407629560554576 -0.11961406596818872 0.14513830141789172 0.05029516871742562 0.2360837711306675 0.23608377113066753
leaf_weight=13.994399905204775 6.7472999542951611 5.4977999627590206 6.9971999526023856 11.495399922132494 5.247899964451789
leaf_count=56 27 22 28 46 21
internal_value=0.0400053 0.136044 -0.0926191 0.213603 0.236084
internal_weight=49.98 28.9884 20.9916 22.2411 16.7433
internal_count=200 116 84 89 67
is_linear=0
shrinkage=1

`

	scanner := bufio.NewScanner(strings.NewReader(input))
	tr, err := parseTree(scanner)
	if err != nil {
		t.Fatalf("parseTree() error = %v", err)
	}

	if tr.numLeaves != 6 {
		t.Errorf("numLeaves = %d, want 6", tr.numLeaves)
	}

	if len(tr.splitFeatures) != 5 {
		t.Errorf("len(splitFeatures) = %d, want 5", len(tr.splitFeatures))
	}

	if len(tr.leafValues) != 6 {
		t.Errorf("len(leafValues) = %d, want 6", len(tr.leafValues))
	}

	// Verify last threshold value
	expectedThreshold4 := 0.56460871160744486
	if tr.thresholds[4] != expectedThreshold4 {
		t.Errorf("thresholds[4] = %e, want %e", tr.thresholds[4], expectedThreshold4)
	}

	// Verify negative leaf values are parsed correctly
	if tr.leafValues[0] != -0.16407629560554576 {
		t.Errorf("leafValues[0] = %f, want -0.16407629560554576", tr.leafValues[0])
	}
}
