package lgbm

import (
	"testing"
)

// TestIntegration_LoadAndInspectModel verifies the full model loading pipeline.
func TestIntegration_LoadAndInspectModel(t *testing.T) {
	// Load a v4 binary model
	model, err := ModelFromFile("testdata/v4/binary.txt", true)
	if err != nil {
		t.Fatalf("ModelFromFile() failed: %v", err)
	}

	// Verify model metadata
	if model.NFeatures() != 10 {
		t.Errorf("NFeatures() = %d, want 10", model.NFeatures())
	}
	if model.NClasses() != 1 {
		t.Errorf("NClasses() = %d, want 1", model.NClasses())
	}
	if model.NTrees() != 20 {
		t.Errorf("NTrees() = %d, want 20", model.NTrees())
	}

	// Verify feature names
	featureNames := model.FeatureNames()
	if len(featureNames) != 10 {
		t.Errorf("len(FeatureNames()) = %d, want 10", len(featureNames))
	}
	if featureNames[0] != "Column_0" {
		t.Errorf("FeatureNames()[0] = %q, want %q", featureNames[0], "Column_0")
	}

	// Verify objective is set correctly
	if model.objective != ObjectiveBinary {
		t.Errorf("objective = %v, want ObjectiveBinary", model.objective)
	}

	// Verify transform function is set (for binary, should be sigmoid)
	if model.transform == nil {
		t.Fatal("transform is nil, want sigmoid function")
	}
}

// TestIntegration_LoadMulticlass verifies loading a multiclass model.
func TestIntegration_LoadMulticlass(t *testing.T) {
	model, err := ModelFromFile("testdata/v4/multiclass.txt", true)
	if err != nil {
		t.Fatalf("ModelFromFile() failed: %v", err)
	}

	// Multiclass has 3 classes
	if model.NClasses() != 3 {
		t.Errorf("NClasses() = %d, want 3", model.NClasses())
	}

	// Verify objective
	if model.objective != ObjectiveMulticlass {
		t.Errorf("objective = %v, want ObjectiveMulticlass", model.objective)
	}
}

// TestIntegration_LoadRegression verifies loading a regression model.
func TestIntegration_LoadRegression(t *testing.T) {
	model, err := ModelFromFile("testdata/v4/regression.txt", true)
	if err != nil {
		t.Fatalf("ModelFromFile() failed: %v", err)
	}

	// Regression has 1 class
	if model.NClasses() != 1 {
		t.Errorf("NClasses() = %d, want 1", model.NClasses())
	}

	// Verify objective
	if model.objective != ObjectiveRegression {
		t.Errorf("objective = %v, want ObjectiveRegression", model.objective)
	}
}

// TestIntegration_RawTransform verifies loading with transformation disabled.
func TestIntegration_RawTransform(t *testing.T) {
	model, err := ModelFromFile("testdata/v4/binary.txt", false)
	if err != nil {
		t.Fatalf("ModelFromFile() failed: %v", err)
	}

	// Transform should be identity
	raw := []float64{2.5}
	out := make([]float64, 1)
	model.transform(raw, out)

	if out[0] != 2.5 {
		t.Errorf("transform(2.5) = %f, want 2.5 (identity)", out[0])
	}
}
