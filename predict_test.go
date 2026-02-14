package lgbm

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
)

// goldenData holds the reference predictions from Python LightGBM.
type goldenData struct {
	Inputs         [][]float64 `json:"inputs"`
	Predictions    []float64   `json:"predictions"`     // for binary/regression/ranking (single value)
	RawPredictions []float64   `json:"raw_predictions"` // for binary raw scores
}

// goldenDataMulticlass holds multiclass reference predictions.
type goldenDataMulticlass struct {
	Inputs      [][]float64 `json:"inputs"`
	Predictions [][]float64 `json:"predictions"` // N class probabilities per input
}

func loadGolden(t *testing.T, path string) goldenData {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var g goldenData
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("failed to parse golden file %s: %v", path, err)
	}
	return g
}

func loadGoldenMulticlass(t *testing.T, path string) goldenDataMulticlass {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var g goldenDataMulticlass
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("failed to parse golden file %s: %v", path, err)
	}
	return g
}

func loadModel(t *testing.T, path string) *Model {
	t.Helper()
	m, err := ModelFromFile(path, true)
	if err != nil {
		t.Fatalf("failed to load model %s: %v", path, err)
	}
	return m
}

const relTol = 1e-6

func assertClose(t *testing.T, got, want float64, label string) {
	t.Helper()
	if want == 0 {
		if math.Abs(got) > relTol {
			t.Errorf("%s: got %v, want %v", label, got, want)
		}
		return
	}
	rel := math.Abs((got - want) / want)
	if rel > relTol {
		t.Errorf("%s: got %v, want %v (relative error %v)", label, got, want, rel)
	}
}

// T019: Golden-file test for PredictSingle with v3 binary model
func TestPredictSingle_V3Binary(t *testing.T) {
	model := loadModel(t, "testdata/v3/binary.txt")
	golden := loadGolden(t, "testdata/v3/binary.json")

	for i, input := range golden.Inputs {
		got, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.Predictions[i], "v3 binary prediction")
	}
}

// T020: Golden-file test for PredictSingle with v4 binary model
func TestPredictSingle_V4Binary(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	for i, input := range golden.Inputs {
		got, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.Predictions[i], "v4 binary prediction")
	}
}

// T021: Error-case tests
func TestPredictSingle_WrongFeatureCount(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")

	_, err := model.PredictSingle([]float64{1.0, 2.0}, 0) // model expects 10
	if !errors.Is(err, ErrFeatureCountMismatch) {
		t.Errorf("expected ErrFeatureCountMismatch, got %v", err)
	}
}

func TestPredictSingle_MulticlassModel(t *testing.T) {
	model := loadModel(t, "testdata/v4/multiclass.txt")

	features := make([]float64, model.NFeatures())
	_, err := model.PredictSingle(features, 0)
	if !errors.Is(err, ErrMulticlassNotSupported) {
		t.Errorf("expected ErrMulticlassNotSupported, got %v", err)
	}
}

// T022: WithRawPredictions test
func TestWithRawPredictions_Binary(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	rawModel := model.WithRawPredictions()

	for i, input := range golden.Inputs {
		got, err := rawModel.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.RawPredictions[i], "v4 binary raw prediction")
	}
}

// T026: Error-handling tests for model loading
func TestModelFromFile_NonexistentFile_Predict(t *testing.T) {
	_, err := ModelFromFile("nonexistent.txt", true)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestModelFromFile_V2Model_Predict(t *testing.T) {
	// Create a temporary v2 model file
	tmpFile, err := os.CreateTemp("", "v2model*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("tree\nversion=v2\nnum_class=1\nmax_feature_idx=9\n\n")
	tmpFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ModelFromFile(tmpFile.Name(), true)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Errorf("expected ErrUnsupportedVersion, got %v", err)
	}
}

// T030: Predict with single-class model (writes to output slice)
func TestPredict_BinaryModel(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	output := make([]float64, 1)
	for i, input := range golden.Inputs {
		err := model.Predict(input, 0, output)
		if err != nil {
			t.Fatalf("input %d: Predict error: %v", i, err)
		}
		assertClose(t, output[0], golden.Predictions[i], "v4 binary Predict")
	}
}

// T028: Golden-file test for regression prediction
func TestPredictSingle_V3Regression(t *testing.T) {
	model := loadModel(t, "testdata/v3/regression.txt")
	golden := loadGolden(t, "testdata/v3/regression.json")

	for i, input := range golden.Inputs {
		got, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.Predictions[i], "v3 regression prediction")
	}
}

func TestPredictSingle_V4Regression(t *testing.T) {
	model := loadModel(t, "testdata/v4/regression.txt")
	golden := loadGolden(t, "testdata/v4/regression.json")

	for i, input := range golden.Inputs {
		got, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.Predictions[i], "v4 regression prediction")
	}
}

// T029: Golden-file test for multiclass prediction
func TestPredict_V3Multiclass(t *testing.T) {
	model := loadModel(t, "testdata/v3/multiclass.txt")
	golden := loadGoldenMulticlass(t, "testdata/v3/multiclass.json")

	output := make([]float64, model.NClasses())
	for i, input := range golden.Inputs {
		err := model.Predict(input, 0, output)
		if err != nil {
			t.Fatalf("input %d: Predict error: %v", i, err)
		}
		var sum float64
		for c := 0; c < model.NClasses(); c++ {
			assertClose(t, output[c], golden.Predictions[i][c], "v3 multiclass prediction")
			sum += output[c]
		}
		if math.Abs(sum-1.0) > 1e-6 {
			t.Errorf("input %d: probabilities sum to %f, want 1.0", i, sum)
		}
	}
}

func TestPredict_V4Multiclass(t *testing.T) {
	model := loadModel(t, "testdata/v4/multiclass.txt")
	golden := loadGoldenMulticlass(t, "testdata/v4/multiclass.json")

	output := make([]float64, model.NClasses())
	for i, input := range golden.Inputs {
		err := model.Predict(input, 0, output)
		if err != nil {
			t.Fatalf("input %d: Predict error: %v", i, err)
		}
		var sum float64
		for c := 0; c < model.NClasses(); c++ {
			assertClose(t, output[c], golden.Predictions[i][c], "v4 multiclass prediction")
			sum += output[c]
		}
		if math.Abs(sum-1.0) > 1e-6 {
			t.Errorf("input %d: probabilities sum to %f, want 1.0", i, sum)
		}
	}
}

// T043: Golden-file test for ranking prediction
func TestPredictSingle_V3Ranking(t *testing.T) {
	model := loadModel(t, "testdata/v3/ranking.txt")
	golden := loadGolden(t, "testdata/v3/ranking.json")

	for i, input := range golden.Inputs {
		got, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.Predictions[i], "v3 ranking prediction")
	}
}

func TestPredictSingle_V4Ranking(t *testing.T) {
	model := loadModel(t, "testdata/v4/ranking.txt")
	golden := loadGolden(t, "testdata/v4/ranking.json")

	for i, input := range golden.Inputs {
		got, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("input %d: PredictSingle error: %v", i, err)
		}
		assertClose(t, got, golden.Predictions[i], "v4 ranking prediction")
	}
}

// Test nEstimators parameter limits trees used
func TestPredictSingle_LimitedEstimators(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	// With all trees
	full, err := model.PredictSingle(golden.Inputs[0], 0)
	if err != nil {
		t.Fatal(err)
	}

	// With 5 trees (should differ from full)
	partial, err := model.PredictSingle(golden.Inputs[0], 5)
	if err != nil {
		t.Fatal(err)
	}

	if full == partial {
		t.Error("expected different predictions with limited estimators")
	}
}
