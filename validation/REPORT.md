# go-lgbm vs leaves Comparison Report

**Generated**: 2026-02-16 00:49:28 UTC
**Tolerance**: 1e-10

## Summary

| Model Type | Test Cases | Max Abs Diff | Mean Abs Diff | Status |
|------------|-----------|-------------|--------------|--------|
| Binary Classification | 1000 | 0.00e+00 | 0.00e+00 | PASS |
| Multiclass Classification | 1000 | 5.55e-16 | 2.14e-17 | PASS |
| Regression | 1000 | 0.00e+00 | 0.00e+00 | PASS |
| Ranking | - | - | - | SKIP (leaves unsupported) |

## Overall Result

**ALL COMPARABLE TESTS PASSED** — go-lgbm produces identical predictions to leaves for all model types that leaves supports.

Note: Some model types were skipped because the leaves library does not support them. go-lgbm supports these model types independently.

## Details

### Binary Classification

- **Test cases**: 1000
- **Max absolute difference**: 0.00e+00
- **Mean absolute difference**: 0.00e+00
- **Status**: PASS

### Multiclass Classification

- **Test cases**: 1000
- **Max absolute difference**: 5.55e-16
- **Mean absolute difference**: 2.14e-17
- **Status**: PASS

### Regression

- **Test cases**: 1000
- **Max absolute difference**: 0.00e+00
- **Mean absolute difference**: 0.00e+00
- **Status**: PASS

### Ranking

**Skipped**: leaves cannot load: unexpected objective field: 'lambdarank'. go-lgbm loads this model successfully but leaves does not support it.

