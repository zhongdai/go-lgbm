# Quickstart: go-lgbm

## Install

```bash
go get github.com/rokt/go-lgbm
```

No C compiler or system libraries required.

## Load a Model and Predict

```go
package main

import (
    "fmt"
    "log"

    lgbm "github.com/rokt/go-lgbm"
)

func main() {
    // Load a LightGBM text-format model (v3 or v4)
    model, err := lgbm.ModelFromFile("model.txt", true)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Features: %d, Classes: %d, Trees: %d\n",
        model.NFeatures(), model.NClasses(), model.NTrees())

    // Single prediction (binary classification / regression)
    features := make([]float64, model.NFeatures())
    // ... fill features ...

    prob, err := model.PredictSingle(features, 0)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Prediction: %f\n", prob)
}
```

## Multiclass Prediction

```go
output := make([]float64, model.NClasses())
err := model.Predict(features, 0, output)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Class probabilities: %v\n", output)
```

## Batch Prediction

```go
nrows := 1000
ncols := model.NFeatures()
vals := make([]float64, nrows*ncols)
// ... fill vals as row-major matrix ...

output := make([]float64, nrows*model.NClasses())
err := model.PredictDense(vals, nrows, ncols, output, 0, 0)
if err != nil {
    log.Fatal(err)
}
```

## Raw Scores (No Transformation)

```go
rawModel := model.WithRawPredictions()
rawScore, err := rawModel.PredictSingle(features, 0)
```

## Migrating from leaves

See [docs/migration-from-leaves.md](../../docs/migration-from-leaves.md)
for a detailed mapping of `leaves` API calls to `go-lgbm` equivalents.

Key changes:
- `leaves.LGEnsembleFromFile` → `lgbm.ModelFromFile`
- `leaves.Ensemble` → `lgbm.Model`
- `ensemble.PredictSingle` → `model.PredictSingle` (now returns error too)
- `ensemble.PredictDense` → `model.PredictDense` (same signature)
