package lgbm

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"testing"
)

// TestModelFromReader_ValidV4Binary tests loading a valid v4 binary model.
func TestModelFromReader_ValidV4Binary(t *testing.T) {
	file, err := os.Open("testdata/v4/binary.txt")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	model, err := parseModel(reader)
	if err != nil {
		t.Fatalf("parseModel() failed: %v", err)
	}

	if model.NFeatures() != 10 {
		t.Errorf("NFeatures() = %d, want 10", model.NFeatures())
	}
	if model.NClasses() != 1 {
		t.Errorf("NClasses() = %d, want 1", model.NClasses())
	}
	if model.NTrees() != 20 {
		t.Errorf("NTrees() = %d, want 20", model.NTrees())
	}
	if model.version != "v4" {
		t.Errorf("version = %q, want %q", model.version, "v4")
	}
	if model.objective != ObjectiveBinary {
		t.Errorf("objective = %v, want ObjectiveBinary", model.objective)
	}
}

// TestModelFromReader_ValidV3Binary tests loading a valid v3 binary model.
func TestModelFromReader_ValidV3Binary(t *testing.T) {
	file, err := os.Open("testdata/v3/binary.txt")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	model, err := parseModel(reader)
	if err != nil {
		t.Fatalf("parseModel() failed: %v", err)
	}

	if model.NFeatures() != 10 {
		t.Errorf("NFeatures() = %d, want 10", model.NFeatures())
	}
	if model.NClasses() != 1 {
		t.Errorf("NClasses() = %d, want 1", model.NClasses())
	}
	if model.NTrees() != 20 {
		t.Errorf("NTrees() = %d, want 20", model.NTrees())
	}
	if model.version != "v3.3.5.99" {
		t.Errorf("version = %q, want %q", model.version, "v3.3.5.99")
	}
	if model.objective != ObjectiveBinary {
		t.Errorf("objective = %v, want ObjectiveBinary", model.objective)
	}
}

// TestModelFromReader_EmptyInput tests error handling for empty input.
func TestModelFromReader_EmptyInput(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(""))
	_, err := parseModel(reader)
	if err == nil {
		t.Fatal("parseModel() succeeded for empty input, want error")
	}
	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("parseModel() error = %v, want ErrInvalidModel", err)
	}
}

// TestModelFromReader_UnsupportedVersion tests error handling for unsupported version.
func TestModelFromReader_UnsupportedVersion(t *testing.T) {
	input := `tree
version=v2
num_class=1
max_feature_idx=5
objective=binary
`
	reader := bufio.NewReader(strings.NewReader(input))
	_, err := parseModel(reader)
	if err == nil {
		t.Fatal("parseModel() succeeded for v2, want error")
	}
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Errorf("parseModel() error = %v, want ErrUnsupportedVersion", err)
	}
}

// TestModelFromReader_ZeroTrees tests error handling for model with 0 trees.
func TestModelFromReader_ZeroTrees(t *testing.T) {
	input := `tree
version=v3
num_class=1
max_feature_idx=5
objective=binary

end of trees
`
	reader := bufio.NewReader(strings.NewReader(input))
	_, err := parseModel(reader)
	if err == nil {
		t.Fatal("parseModel() succeeded for 0 trees, want error")
	}
	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("parseModel() error = %v, want ErrInvalidModel", err)
	}
}

// TestModelFromFile_ValidFile tests the file-loading wrapper.
func TestModelFromFile_ValidFile(t *testing.T) {
	model, err := modelFromFile("testdata/v4/binary.txt", true)
	if err != nil {
		t.Fatalf("modelFromFile() failed: %v", err)
	}

	if model.NFeatures() != 10 {
		t.Errorf("NFeatures() = %d, want 10", model.NFeatures())
	}
	if model.NTrees() != 20 {
		t.Errorf("NTrees() = %d, want 20", model.NTrees())
	}
}

// TestModelFromFile_NonexistentFile tests error handling for missing file.
func TestModelFromFile_NonexistentFile(t *testing.T) {
	_, err := modelFromFile("nonexistent.txt", true)
	if err == nil {
		t.Fatal("modelFromFile() succeeded for nonexistent file, want error")
	}
}

// TestModelFromFile_NoTransformation tests loading with transformation disabled.
func TestModelFromFile_NoTransformation(t *testing.T) {
	model, err := modelFromFile("testdata/v4/binary.txt", false)
	if err != nil {
		t.Fatalf("modelFromFile() failed: %v", err)
	}

	// Verify that transform is identity (raw scores)
	if model.transform == nil {
		t.Fatal("transform is nil, want identity function")
	}
}

// TestPublicAPI_ModelFromFile tests the public ModelFromFile function.
func TestPublicAPI_ModelFromFile(t *testing.T) {
	model, err := ModelFromFile("testdata/v4/binary.txt", true)
	if err != nil {
		t.Fatalf("ModelFromFile() failed: %v", err)
	}

	if model.NFeatures() != 10 {
		t.Errorf("NFeatures() = %d, want 10", model.NFeatures())
	}
	if model.NTrees() != 20 {
		t.Errorf("NTrees() = %d, want 20", model.NTrees())
	}
}

// TestPublicAPI_ModelFromReader tests the public ModelFromReader function.
func TestPublicAPI_ModelFromReader(t *testing.T) {
	file, err := os.Open("testdata/v4/binary.txt")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	model, err := ModelFromReader(reader, true)
	if err != nil {
		t.Fatalf("ModelFromReader() failed: %v", err)
	}

	if model.NFeatures() != 10 {
		t.Errorf("NFeatures() = %d, want 10", model.NFeatures())
	}
	if model.NTrees() != 20 {
		t.Errorf("NTrees() = %d, want 20", model.NTrees())
	}
}

// TestPublicAPI_ModelFromReader_NoTransform tests loading with transformation disabled.
func TestPublicAPI_ModelFromReader_NoTransform(t *testing.T) {
	file, err := os.Open("testdata/v4/binary.txt")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	model, err := ModelFromReader(reader, false)
	if err != nil {
		t.Fatalf("ModelFromReader() failed: %v", err)
	}

	if model.transform == nil {
		t.Fatal("transform is nil, want identity function")
	}
}
