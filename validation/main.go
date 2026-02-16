// Package main implements a validation program that compares predictions
// from go-lgbm and leaves libraries to verify they produce identical output.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	lgbm "github.com/zhongdai/go-lgbm"

	"github.com/dmitryikh/leaves"
)

const tolerance = 1e-10

type testData struct {
	Inputs    [][]float64 `json:"inputs"`
	NFeatures int         `json:"n_features"`
	NClasses  int         `json:"n_classes"`
}

type modelResult struct {
	Name        string
	TestCases   int
	MaxAbsDiff    float64
	MeanAbsDiff   float64
	Pass          bool
	Error         string
	LeavesUnsupported bool
}

type modelConfig struct {
	Name       string
	ModelFile  string
	DataFile   string
	Multiclass bool
}

func loadTestData(path string) (testData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return testData{}, fmt.Errorf("read test data %s: %w", path, err)
	}
	var td testData
	if err := json.Unmarshal(raw, &td); err != nil {
		return testData{}, fmt.Errorf("parse test data %s: %w", path, err)
	}
	return td, nil
}

func compareModel(cfg modelConfig) modelResult {
	// Load with go-lgbm
	goModel, err := lgbm.ModelFromFile(cfg.ModelFile, true)
	if err != nil {
		return modelResult{Name: cfg.Name, Error: fmt.Sprintf("go-lgbm load: %v", err)}
	}

	// Load with leaves
	leavesModel, err := leaves.LGEnsembleFromFile(cfg.ModelFile, true)
	if err != nil {
		return modelResult{
			Name:              cfg.Name,
			Error:             fmt.Sprintf("leaves cannot load: %v", err),
			LeavesUnsupported: true,
		}
	}

	// Load test data
	td, err := loadTestData(cfg.DataFile)
	if err != nil {
		return modelResult{Name: cfg.Name, Error: fmt.Sprintf("test data: %v", err)}
	}

	var maxDiff, sumDiff float64
	totalComparisons := 0

	for _, input := range td.Inputs {
		if cfg.Multiclass {
			nClasses := goModel.NClasses()
			goOut := make([]float64, nClasses)
			if err := goModel.Predict(input, 0, goOut); err != nil {
				return modelResult{Name: cfg.Name, Error: fmt.Sprintf("go-lgbm predict: %v", err)}
			}

			leavesOut := make([]float64, leavesModel.NOutputGroups())
			if err := leavesModel.Predict(input, 0, leavesOut); err != nil {
				return modelResult{Name: cfg.Name, Error: fmt.Sprintf("leaves predict: %v", err)}
			}

			for c := 0; c < nClasses; c++ {
				diff := math.Abs(goOut[c] - leavesOut[c])
				if diff > maxDiff {
					maxDiff = diff
				}
				sumDiff += diff
				totalComparisons++
			}
		} else {
			goPred, err := goModel.PredictSingle(input, 0)
			if err != nil {
				return modelResult{Name: cfg.Name, Error: fmt.Sprintf("go-lgbm predict: %v", err)}
			}

			leavesPred := leavesModel.PredictSingle(input, 0)

			diff := math.Abs(goPred - leavesPred)
			if diff > maxDiff {
				maxDiff = diff
			}
			sumDiff += diff
			totalComparisons++
		}
	}

	meanDiff := 0.0
	if totalComparisons > 0 {
		meanDiff = sumDiff / float64(totalComparisons)
	}

	return modelResult{
		Name:        cfg.Name,
		TestCases:   len(td.Inputs),
		MaxAbsDiff:  maxDiff,
		MeanAbsDiff: meanDiff,
		Pass:        maxDiff <= tolerance,
	}
}

func writeReport(results []modelResult, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# go-lgbm vs leaves Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("**Tolerance**: %.0e\n\n", tolerance))

	// Summary table
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Model Type | Test Cases | Max Abs Diff | Mean Abs Diff | Status |\n")
	sb.WriteString("|------------|-----------|-------------|--------------|--------|\n")

	allPass := true
	hasUnsupported := false
	for _, r := range results {
		if r.LeavesUnsupported {
			sb.WriteString(fmt.Sprintf("| %s | - | - | - | SKIP (leaves unsupported) |\n", r.Name))
			hasUnsupported = true
			continue
		}
		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("| %s | - | - | - | FAIL (error) |\n", r.Name))
			allPass = false
			continue
		}
		status := "PASS"
		if !r.Pass {
			status = "FAIL"
			allPass = false
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %.2e | %.2e | %s |\n",
			r.Name, r.TestCases, r.MaxAbsDiff, r.MeanAbsDiff, status))
	}

	sb.WriteString("\n## Overall Result\n\n")
	if allPass {
		sb.WriteString("**ALL COMPARABLE TESTS PASSED** — go-lgbm produces identical predictions to leaves for all model types that leaves supports.\n")
		if hasUnsupported {
			sb.WriteString("\nNote: Some model types were skipped because the leaves library does not support them. go-lgbm supports these model types independently.\n")
		}
	} else {
		sb.WriteString("**SOME TESTS FAILED** — see details above.\n")
	}

	// Detail sections
	sb.WriteString("\n## Details\n\n")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("### %s\n\n", r.Name))
		if r.LeavesUnsupported {
			sb.WriteString(fmt.Sprintf("**Skipped**: %s. go-lgbm loads this model successfully but leaves does not support it.\n\n", r.Error))
			continue
		}
		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("**Error**: %s\n\n", r.Error))
			continue
		}
		sb.WriteString(fmt.Sprintf("- **Test cases**: %d\n", r.TestCases))
		sb.WriteString(fmt.Sprintf("- **Max absolute difference**: %.2e\n", r.MaxAbsDiff))
		sb.WriteString(fmt.Sprintf("- **Mean absolute difference**: %.2e\n", r.MeanAbsDiff))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", map[bool]string{true: "PASS", false: "FAIL"}[r.Pass]))
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

func main() {
	modelsDir := "models"
	testdataDir := "testdata"
	reportPath := "REPORT.md"

	configs := []modelConfig{
		{
			Name:       "Binary Classification",
			ModelFile:  filepath.Join(modelsDir, "binary.txt"),
			DataFile:   filepath.Join(testdataDir, "binary.json"),
			Multiclass: false,
		},
		{
			Name:       "Multiclass Classification",
			ModelFile:  filepath.Join(modelsDir, "multiclass.txt"),
			DataFile:   filepath.Join(testdataDir, "multiclass.json"),
			Multiclass: true,
		},
		{
			Name:       "Regression",
			ModelFile:  filepath.Join(modelsDir, "regression.txt"),
			DataFile:   filepath.Join(testdataDir, "regression.json"),
			Multiclass: false,
		},
		{
			Name:       "Ranking",
			ModelFile:  filepath.Join(modelsDir, "ranking.txt"),
			DataFile:   filepath.Join(testdataDir, "ranking.json"),
			Multiclass: false,
		},
	}

	// Check that models exist
	for _, cfg := range configs {
		if _, err := os.Stat(cfg.ModelFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Model file %s not found.\n", cfg.ModelFile)
			fmt.Fprintf(os.Stderr, "Run 'just generate-models' first to generate models and test data.\n")
			os.Exit(1)
		}
		if _, err := os.Stat(cfg.DataFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Test data file %s not found.\n", cfg.DataFile)
			fmt.Fprintf(os.Stderr, "Run 'just generate-models' first to generate models and test data.\n")
			os.Exit(1)
		}
	}

	fmt.Println("Running validation...")
	var results []modelResult
	for _, cfg := range configs {
		fmt.Printf("  Comparing %s...\n", cfg.Name)
		result := compareModel(cfg)
		if result.LeavesUnsupported {
			fmt.Printf("    SKIP (leaves unsupported: %s)\n", result.Error)
		} else if result.Error != "" {
			fmt.Printf("    ERROR: %s\n", result.Error)
		} else if result.Pass {
			fmt.Printf("    PASS (max diff: %.2e)\n", result.MaxAbsDiff)
		} else {
			fmt.Printf("    FAIL (max diff: %.2e)\n", result.MaxAbsDiff)
		}
		results = append(results, result)
	}

	if err := writeReport(results, reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nReport written to %s\n", reportPath)
}
