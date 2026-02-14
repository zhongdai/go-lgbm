<h1 align="center">
  🌳 go-lgbm
</h1>

<p align="center">
  Pure Go library for LightGBM model loading and inference.<br>
  No CGo. No dependencies. Just Go.
</p>

<p align="center">
  <a href="https://github.com/zhongdai/go-lgbm/actions/workflows/ci.yml"><img src="https://github.com/zhongdai/go-lgbm/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/zhongdai/go-lgbm"><img src="https://pkg.go.dev/badge/github.com/zhongdai/go-lgbm.svg" alt="Go Reference"></a>
</p>

---

## Introduction

**go-lgbm** loads LightGBM text-format models and runs inference entirely in Go — no C bindings, no external libraries. It supports LightGBM **v3** and **v4** model formats with all major objective types:

- **Binary classification** (sigmoid)
- **Multiclass classification** (softmax)
- **Regression** (identity)
- **Ranking** (lambdarank, rank_xendcg)
- **Poisson / Gamma / Tweedie** (exponential)

It is designed as a modern, maintained replacement for [github.com/dmitryikh/leaves](https://github.com/dmitryikh/leaves) with a compatible API surface. See the [migration guide](docs/migration-from-leaves.md) for a detailed mapping.

### Key features

- **Pure Go** — deploy anywhere Go runs, no CGo compilation required
- **LightGBM v3 + v4** — supports both current model format versions
- **Batch prediction** — `PredictDense` with goroutine-based parallelism
- **Zero dependencies** — only the Go standard library
- **Thoroughly tested** — 87%+ coverage, golden-file verified against Python LightGBM

## Installation

```bash
go get github.com/zhongdai/go-lgbm
```

Requires **Go 1.21** or later.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	lgbm "github.com/zhongdai/go-lgbm"
)

func main() {
	// Load a LightGBM text model
	model, err := lgbm.ModelFromFile("model.txt", true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Features: %d, Trees: %d\n", model.NFeatures(), model.NTrees())

	// Predict a single sample
	features := []float64{0.5, 1.2, -0.3, 0.8, 0.1, -1.5, 0.7, 2.1, -0.9, 0.4}
	prob, err := model.PredictSingle(features, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Prediction: %.6f\n", prob)
}
```

## Usage Examples

### Binary Classification

```go
model, _ := lgbm.ModelFromFile("binary_model.txt", true)

features := []float64{0.5, 1.2, -0.3, 0.8, 0.1, -1.5, 0.7, 2.1, -0.9, 0.4}

// Returns probability after sigmoid transform
prob, err := model.PredictSingle(features, 0) // 0 = use all trees
```

### Multiclass Classification

```go
model, _ := lgbm.ModelFromFile("multiclass_model.txt", true)

features := []float64{0.5, 1.2, -0.3, 0.8, 0.1, -1.5, 0.7, 2.1, -0.9, 0.4}

// Output probabilities for each class (softmax applied)
output := make([]float64, model.NClasses())
err := model.Predict(features, 0, output)
// output = [0.85, 0.10, 0.05]  (sums to 1.0)
```

### Batch Prediction with Parallelism

```go
model, _ := lgbm.ModelFromFile("model.txt", true)

nRows := 10000
nCols := model.NFeatures()
features := make([]float64, nRows*nCols) // flat row-major matrix
// ... fill features ...

output := make([]float64, nRows)
err := model.PredictDense(
	features,
	nRows, nCols,
	0, // nEstimators: 0 = all trees
	0, // nThreads: 0 = runtime.NumCPU()
	output,
)
```

### Raw Predictions (No Transform)

```go
model, _ := lgbm.ModelFromFile("binary_model.txt", true)
rawModel := model.WithRawPredictions()

// Returns raw log-odds (no sigmoid)
logOdds, err := rawModel.PredictSingle(features, 0)
```

### Limiting Trees (Early Stopping)

```go
// Use only the first 50 trees for prediction
prob, err := model.PredictSingle(features, 50)
```

## API Reference

### Model Loading

| Function | Description |
|----------|-------------|
| `ModelFromFile(path, loadTransform) (*Model, error)` | Load model from a text file |
| `ModelFromReader(reader, loadTransform) (*Model, error)` | Load model from a `*bufio.Reader` |

When `loadTransform` is `true`, the appropriate output transformation (sigmoid, softmax, etc.) is applied based on the model's objective. When `false`, raw tree scores are returned.

### Prediction

| Method | Description |
|--------|-------------|
| `PredictSingle(features, nEstimators) (float64, error)` | Single prediction for binary/regression/ranking models |
| `Predict(features, nEstimators, output) error` | Prediction into output slice (works for all model types including multiclass) |
| `PredictDense(features, nRows, nCols, nEstimators, nThreads, output) error` | Batch prediction on a flat row-major feature matrix |
| `WithRawPredictions() *Model` | Returns a new Model that skips the output transformation |

### Model Inspection

| Method | Description |
|--------|-------------|
| `NFeatures() int` | Number of input features |
| `NClasses() int` | Number of output classes (1 for binary/regression) |
| `NTrees() int` | Total number of trees in the ensemble |
| `FeatureNames() []string` | Feature names from the model file (nil if absent) |

### Errors

| Error | Meaning |
|-------|---------|
| `ErrUnsupportedVersion` | Model version is not v3 or v4 |
| `ErrInvalidModel` | Malformed or truncated model file |
| `ErrFeatureCountMismatch` | Feature vector length does not match model |
| `ErrMulticlassNotSupported` | `PredictSingle` called on a multiclass model |

All errors support `errors.Is()` for type-safe checking.

## Benchmarks

Measured on Apple M1 Pro with a 20-tree binary classification model:

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| PredictSingle | 181 | 16 | 2 |
| PredictDense (1 thread, 1000 rows) | 177,152 | 8,008 | 1,001 |
| PredictDense (NumCPU, 1000 rows) | 89,566 | 9,736 | 1,024 |

## Migrating from leaves

See the full [migration guide](docs/migration-from-leaves.md) for a detailed API mapping. The short version:

```go
// Before (leaves)
ensemble, _ := leaves.LGEnsembleFromFile("model.txt", true)
pred := ensemble.PredictSingle(features, 0)

// After (go-lgbm)
model, _ := lgbm.ModelFromFile("model.txt", true)
pred, err := model.PredictSingle(features, 0)
```

## Disclaimer

> **Warning**: This project was built entirely through vibe coding with [Claude Code](https://claude.ai/claude-code). While it is thoroughly tested (87%+ coverage, golden-file verified against Python LightGBM), use it at your own risk. Review the code and run your own validation before using in production.

## License

MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgements

Built with [Claude Code](https://claude.ai/claude-code) by Anthropic.
