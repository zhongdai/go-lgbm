package lgbm

import (
	"encoding/json"
	"os"
	"testing"
)

func BenchmarkPredictSingle(b *testing.B) {
	model := loadModelBench(b, "testdata/v4/binary.txt")
	golden := loadGoldenBench(b, "testdata/v4/binary.json")
	input := golden.Inputs[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = model.PredictSingle(input, 0)
	}
}

func BenchmarkPredictDense_1Thread(b *testing.B) {
	model := loadModelBench(b, "testdata/v4/binary.txt")
	golden := loadGoldenBench(b, "testdata/v4/binary.json")

	nRows := 1000
	nCols := model.NFeatures()
	features := makeDenseMatrix(golden.Inputs, nRows, nCols)
	output := make([]float64, nRows)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.PredictDense(features, nRows, nCols, 0, 1, output)
	}
}

func BenchmarkPredictDense_NumCPU(b *testing.B) {
	model := loadModelBench(b, "testdata/v4/binary.txt")
	golden := loadGoldenBench(b, "testdata/v4/binary.json")

	nRows := 1000
	nCols := model.NFeatures()
	features := makeDenseMatrix(golden.Inputs, nRows, nCols)
	output := make([]float64, nRows)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.PredictDense(features, nRows, nCols, 0, 0, output)
	}
}

func BenchmarkPredictDense_Multiclass_NumCPU(b *testing.B) {
	model := loadModelBench(b, "testdata/v4/multiclass.txt")
	golden := loadGoldenMulticlassBench(b, "testdata/v4/multiclass.json")

	nRows := 1000
	nCols := model.NFeatures()
	nClasses := model.NClasses()
	features := makeDenseMatrixMulti(golden.Inputs, nRows, nCols)
	output := make([]float64, nRows*nClasses)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.PredictDense(features, nRows, nCols, 0, 0, output)
	}
}

// Helper: load model for benchmarks
func loadModelBench(b *testing.B, path string) *Model {
	b.Helper()
	m, err := ModelFromFile(path, true)
	if err != nil {
		b.Fatalf("failed to load model %s: %v", path, err)
	}
	return m
}

func loadGoldenBench(b *testing.B, path string) goldenData {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var g goldenData
	if err := json.Unmarshal(data, &g); err != nil {
		b.Fatalf("failed to parse golden file %s: %v", path, err)
	}
	return g
}

func loadGoldenMulticlassBench(b *testing.B, path string) goldenDataMulticlass {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var g goldenDataMulticlass
	if err := json.Unmarshal(data, &g); err != nil {
		b.Fatalf("failed to parse golden file %s: %v", path, err)
	}
	return g
}

// makeDenseMatrix creates a flat row-major matrix by repeating golden inputs
func makeDenseMatrix(inputs [][]float64, nRows, nCols int) []float64 {
	features := make([]float64, nRows*nCols)
	for i := 0; i < nRows; i++ {
		copy(features[i*nCols:], inputs[i%len(inputs)])
	}
	return features
}

func makeDenseMatrixMulti(inputs [][]float64, nRows, nCols int) []float64 {
	features := make([]float64, nRows*nCols)
	for i := 0; i < nRows; i++ {
		copy(features[i*nCols:], inputs[i%len(inputs)])
	}
	return features
}
