package lgbm

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

func TestParseHeader_ValidV3Binary(t *testing.T) {
	input := `tree
version=v3.3.5.99
num_class=1
num_tree_per_iteration=1
label_index=0
max_feature_idx=9
objective=binary sigmoid:1
feature_names=Column_0 Column_1 Column_2 Column_3 Column_4 Column_5 Column_6 Column_7 Column_8 Column_9
tree_sizes=753 860 971 959 869

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	h, err := parseHeader(scanner)
	if err != nil {
		t.Fatalf("parseHeader failed: %v", err)
	}

	if h.version != "v3.3.5.99" {
		t.Errorf("version = %q, want %q", h.version, "v3.3.5.99")
	}
	if h.numClass != 1 {
		t.Errorf("numClass = %d, want 1", h.numClass)
	}
	if h.numTreePerIteration != 1 {
		t.Errorf("numTreePerIteration = %d, want 1", h.numTreePerIteration)
	}
	if h.labelIndex != 0 {
		t.Errorf("labelIndex = %d, want 0", h.labelIndex)
	}
	if h.maxFeatureIdx != 9 {
		t.Errorf("maxFeatureIdx = %d, want 9", h.maxFeatureIdx)
	}
	if h.objective != "binary sigmoid:1" {
		t.Errorf("objective = %q, want %q", h.objective, "binary sigmoid:1")
	}
	expectedFeatures := []string{"Column_0", "Column_1", "Column_2", "Column_3", "Column_4", "Column_5", "Column_6", "Column_7", "Column_8", "Column_9"}
	if len(h.featureNames) != len(expectedFeatures) {
		t.Fatalf("len(featureNames) = %d, want %d", len(h.featureNames), len(expectedFeatures))
	}
	for i, name := range expectedFeatures {
		if h.featureNames[i] != name {
			t.Errorf("featureNames[%d] = %q, want %q", i, h.featureNames[i], name)
		}
	}
	expectedTreeSizes := []int{753, 860, 971, 959, 869}
	if len(h.treeSizes) != len(expectedTreeSizes) {
		t.Fatalf("len(treeSizes) = %d, want %d", len(h.treeSizes), len(expectedTreeSizes))
	}
	for i, size := range expectedTreeSizes {
		if h.treeSizes[i] != size {
			t.Errorf("treeSizes[%d] = %d, want %d", i, h.treeSizes[i], size)
		}
	}
	if h.averageOutput {
		t.Errorf("averageOutput = true, want false")
	}
}

func TestParseHeader_ValidV4(t *testing.T) {
	input := `tree
version=v4
num_class=1
max_feature_idx=9
objective=binary sigmoid:1

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	h, err := parseHeader(scanner)
	if err != nil {
		t.Fatalf("parseHeader failed: %v", err)
	}

	if h.version != "v4" {
		t.Errorf("version = %q, want %q", h.version, "v4")
	}
	if h.numClass != 1 {
		t.Errorf("numClass = %d, want 1", h.numClass)
	}
	if h.maxFeatureIdx != 9 {
		t.Errorf("maxFeatureIdx = %d, want 9", h.maxFeatureIdx)
	}
	// num_tree_per_iteration should default to num_class
	if h.numTreePerIteration != 1 {
		t.Errorf("numTreePerIteration = %d, want 1 (defaulted to num_class)", h.numTreePerIteration)
	}
}

func TestParseHeader_WithFeatureNames(t *testing.T) {
	input := `tree
version=v3
num_class=2
max_feature_idx=2
feature_names=feat_a feat_b feat_c

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	h, err := parseHeader(scanner)
	if err != nil {
		t.Fatalf("parseHeader failed: %v", err)
	}

	expected := []string{"feat_a", "feat_b", "feat_c"}
	if len(h.featureNames) != len(expected) {
		t.Fatalf("len(featureNames) = %d, want %d", len(h.featureNames), len(expected))
	}
	for i, name := range expected {
		if h.featureNames[i] != name {
			t.Errorf("featureNames[%d] = %q, want %q", i, h.featureNames[i], name)
		}
	}
}

func TestParseHeader_WithAverageOutput(t *testing.T) {
	input := `tree
version=v3
num_class=1
max_feature_idx=5
average_output

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	h, err := parseHeader(scanner)
	if err != nil {
		t.Fatalf("parseHeader failed: %v", err)
	}

	if !h.averageOutput {
		t.Errorf("averageOutput = false, want true")
	}
}

func TestParseHeader_UnsupportedVersionV2(t *testing.T) {
	input := `tree
version=v2.0.0
num_class=1
max_feature_idx=5

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseHeader(scanner)
	if err == nil {
		t.Fatal("parseHeader succeeded, want error for v2")
	}

	var versionErr *VersionError
	if !errors.As(err, &versionErr) {
		t.Fatalf("error type = %T, want *VersionError", err)
	}
	if versionErr.Version != "v2.0.0" {
		t.Errorf("VersionError.Version = %q, want %q", versionErr.Version, "v2.0.0")
	}
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Errorf("error does not wrap ErrUnsupportedVersion")
	}
}

func TestParseHeader_UnsupportedVersionV1(t *testing.T) {
	input := `tree
version=v1.0.0
num_class=1
max_feature_idx=5

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseHeader(scanner)
	if err == nil {
		t.Fatal("parseHeader succeeded, want error for v1")
	}

	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Errorf("error does not wrap ErrUnsupportedVersion")
	}
}

func TestParseHeader_MissingNumClass(t *testing.T) {
	input := `tree
version=v3
max_feature_idx=5

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseHeader(scanner)
	if err == nil {
		t.Fatal("parseHeader succeeded, want error for missing num_class")
	}

	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("error does not wrap ErrInvalidModel")
	}
}

func TestParseHeader_MissingMaxFeatureIdx(t *testing.T) {
	input := `tree
version=v3
num_class=1

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseHeader(scanner)
	if err == nil {
		t.Fatal("parseHeader succeeded, want error for missing max_feature_idx")
	}

	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("error does not wrap ErrInvalidModel")
	}
}

func TestParseHeader_MissingTreeMagic(t *testing.T) {
	input := `version=v3
num_class=1
max_feature_idx=5

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseHeader(scanner)
	if err == nil {
		t.Fatal("parseHeader succeeded, want error for missing 'tree' magic")
	}

	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("error does not wrap ErrInvalidModel")
	}
}

func TestParseHeader_MissingVersion(t *testing.T) {
	input := `tree
num_class=1
max_feature_idx=5

`
	scanner := bufio.NewScanner(strings.NewReader(input))
	_, err := parseHeader(scanner)
	if err == nil {
		t.Fatal("parseHeader succeeded, want error for missing version")
	}

	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("error does not wrap ErrInvalidModel")
	}
}
