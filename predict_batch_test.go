package lgbm

import (
	"errors"
	"sync"
	"testing"
)

// T035: Correctness test — PredictDense matches PredictSingle for binary model
func TestPredictDense_BinaryCorrectness(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	nRows := len(golden.Inputs)
	nCols := model.NFeatures()

	// Flatten inputs into dense matrix
	features := make([]float64, nRows*nCols)
	for i, row := range golden.Inputs {
		copy(features[i*nCols:], row)
	}

	// PredictDense
	output := make([]float64, nRows)
	err := model.PredictDense(features, nRows, nCols, 0, 1, output)
	if err != nil {
		t.Fatalf("PredictDense error: %v", err)
	}

	// Compare with PredictSingle
	for i, input := range golden.Inputs {
		single, err := model.PredictSingle(input, 0)
		if err != nil {
			t.Fatalf("PredictSingle error: %v", err)
		}
		if output[i] != single {
			t.Errorf("row %d: PredictDense=%f, PredictSingle=%f", i, output[i], single)
		}
	}
}

// T036: Correctness test — PredictDense matches Predict for multiclass model
func TestPredictDense_MulticlassCorrectness(t *testing.T) {
	model := loadModel(t, "testdata/v4/multiclass.txt")
	golden := loadGoldenMulticlass(t, "testdata/v4/multiclass.json")

	nRows := len(golden.Inputs)
	nCols := model.NFeatures()
	nClasses := model.NClasses()

	// Flatten inputs
	features := make([]float64, nRows*nCols)
	for i, row := range golden.Inputs {
		copy(features[i*nCols:], row)
	}

	// PredictDense
	output := make([]float64, nRows*nClasses)
	err := model.PredictDense(features, nRows, nCols, 0, 1, output)
	if err != nil {
		t.Fatalf("PredictDense error: %v", err)
	}

	// Compare with Predict per row
	singleOutput := make([]float64, nClasses)
	for i, input := range golden.Inputs {
		err := model.Predict(input, 0, singleOutput)
		if err != nil {
			t.Fatalf("Predict error: %v", err)
		}
		for c := 0; c < nClasses; c++ {
			if output[i*nClasses+c] != singleOutput[c] {
				t.Errorf("row %d class %d: PredictDense=%f, Predict=%f",
					i, c, output[i*nClasses+c], singleOutput[c])
			}
		}
	}
}

// T037: Concurrency test — multiple goroutines calling PredictDense
func TestPredictDense_Concurrency(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	nRows := len(golden.Inputs)
	nCols := model.NFeatures()

	features := make([]float64, nRows*nCols)
	for i, row := range golden.Inputs {
		copy(features[i*nCols:], row)
	}

	// Get reference results
	reference := make([]float64, nRows)
	err := model.PredictDense(features, nRows, nCols, 0, 1, reference)
	if err != nil {
		t.Fatalf("reference PredictDense error: %v", err)
	}

	// Run 100 concurrent goroutines
	var wg sync.WaitGroup
	for g := 0; g < 100; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out := make([]float64, nRows)
			err := model.PredictDense(features, nRows, nCols, 0, 0, out)
			if err != nil {
				t.Errorf("concurrent PredictDense error: %v", err)
				return
			}
			for i := range reference {
				if out[i] != reference[i] {
					t.Errorf("concurrent result mismatch at row %d", i)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// T038: Input validation tests for PredictDense
func TestPredictDense_WrongColumnCount(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")

	features := make([]float64, 10)
	output := make([]float64, 1)
	err := model.PredictDense(features, 1, 5, 0, 1, output) // model expects 10 cols
	if !errors.Is(err, ErrFeatureCountMismatch) {
		t.Errorf("expected ErrFeatureCountMismatch, got %v", err)
	}
}

func TestPredictDense_OutputTooShort(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")

	features := make([]float64, 20) // 2 rows * 10 cols
	output := make([]float64, 1)    // too short for 2 rows
	err := model.PredictDense(features, 2, 10, 0, 1, output)
	if !errors.Is(err, ErrInvalidModel) {
		t.Errorf("expected ErrInvalidModel, got %v", err)
	}
}

func TestPredictDense_ZeroRows(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")

	err := model.PredictDense(nil, 0, 10, 0, 1, nil)
	if err != nil {
		t.Errorf("expected nil error for zero rows, got %v", err)
	}
}

// Test parallel PredictDense produces same results as single-threaded
func TestPredictDense_ParallelMatchesSingleThreaded(t *testing.T) {
	model := loadModel(t, "testdata/v4/binary.txt")
	golden := loadGolden(t, "testdata/v4/binary.json")

	nRows := len(golden.Inputs)
	nCols := model.NFeatures()

	features := make([]float64, nRows*nCols)
	for i, row := range golden.Inputs {
		copy(features[i*nCols:], row)
	}

	single := make([]float64, nRows)
	err := model.PredictDense(features, nRows, nCols, 0, 1, single)
	if err != nil {
		t.Fatal(err)
	}

	parallel := make([]float64, nRows)
	err = model.PredictDense(features, nRows, nCols, 0, 4, parallel)
	if err != nil {
		t.Fatal(err)
	}

	for i := range single {
		if single[i] != parallel[i] {
			t.Errorf("row %d: single=%f, parallel=%f", i, single[i], parallel[i])
		}
	}
}
