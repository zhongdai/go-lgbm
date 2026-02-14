# Migration Guide: leaves to go-lgbm

This guide maps the `github.com/dmitryikh/leaves` API to the equivalent `github.com/rokt/go-lgbm` API.

## Model Loading

| leaves | go-lgbm |
|--------|---------|
| `leaves.LGEnsembleFromFile(path, true)` | `lgbm.ModelFromFile(path, true)` |
| `leaves.LGEnsembleFromReader(reader, true)` | `lgbm.ModelFromReader(reader, true)` |

## Types

| leaves | go-lgbm |
|--------|---------|
| `*leaves.Ensemble` | `*lgbm.Model` |
| `leaves.EnsembleWithRawPredictions(e)` | `model.WithRawPredictions()` |

## Single Prediction

```go
// leaves
ensemble, _ := leaves.LGEnsembleFromFile("model.txt", true)
prediction := ensemble.PredictSingle(features, 0)

// go-lgbm
model, _ := lgbm.ModelFromFile("model.txt", true)
prediction, err := model.PredictSingle(features, 0)
```

Key difference: `PredictSingle` in go-lgbm returns an error as a second value. In leaves, errors were handled via panics.

## Multiclass Prediction

```go
// leaves
ensemble, _ := leaves.LGEnsembleFromFile("multiclass_model.txt", true)
predictions := make([]float64, ensemble.NOutputGroups())
ensemble.PredictSingle(features, 0, predictions)

// go-lgbm
model, _ := lgbm.ModelFromFile("multiclass_model.txt", true)
predictions := make([]float64, model.NClasses())
err := model.Predict(features, 0, predictions)
```

Key difference: Use `Predict` (not `PredictSingle`) for multiclass models. `PredictSingle` returns `ErrMulticlassNotSupported` for multiclass.

## Batch Prediction

```go
// leaves
ensemble, _ := leaves.LGEnsembleFromFile("model.txt", true)
predictions := make([]float64, nRows)
ensemble.PredictDense(features, nRows, nCols, predictions, 0, runtime.NumCPU())

// go-lgbm
model, _ := lgbm.ModelFromFile("model.txt", true)
predictions := make([]float64, nRows)
err := model.PredictDense(features, nRows, nCols, 0, 0, predictions)
```

Key differences:
- Parameter order differs: `nEstimators` before `nThreads` in go-lgbm
- `nThreads=0` means `runtime.NumCPU()` (same as leaves)
- Returns an error instead of panicking

## Raw Predictions

```go
// leaves
rawEnsemble := leaves.EnsembleWithRawPredictions(ensemble)

// go-lgbm
rawModel := model.WithRawPredictions()
```

## Metadata

| leaves | go-lgbm |
|--------|---------|
| `ensemble.NFeatures()` | `model.NFeatures()` |
| `ensemble.NOutputGroups()` | `model.NClasses()` |
| `ensemble.NEstimators()` | `model.NTrees()` |
| N/A | `model.FeatureNames()` |

## Error Handling

go-lgbm uses sentinel errors for common failure modes:

| Error | Meaning |
|-------|---------|
| `lgbm.ErrUnsupportedVersion` | Model version not v3 or v4 |
| `lgbm.ErrInvalidModel` | Malformed or truncated model file |
| `lgbm.ErrFeatureCountMismatch` | Wrong number of features passed |
| `lgbm.ErrMulticlassNotSupported` | PredictSingle called on multiclass model |

Use `errors.Is(err, lgbm.ErrFeatureCountMismatch)` for type-safe error checking.
