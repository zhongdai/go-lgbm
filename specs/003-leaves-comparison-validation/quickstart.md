# Quickstart: Leaves Comparison Validation

## Prerequisites

- Go 1.21+
- Python 3.x with `lightgbm` and `scikit-learn` installed
- `just` command-line tool

## Run the Full Validation

From the repository root:

```bash
just validate
```

This single command:
1. Generates 4 LightGBM models (binary, multiclass, regression, ranking) using Python
2. Generates 1,000 random test inputs per model type
3. Runs the Go validation program that loads each model in both go-lgbm and leaves
4. Compares predictions and produces `validation/REPORT.md`

## Run Individual Steps

From the `validation/` directory:

```bash
# Generate models and test inputs only
just generate-models

# Run comparison only (models must exist)
just validate

# Run everything
just all
```

## View the Report

After running, open `validation/REPORT.md` or view it on GitHub via the link in the project README.

## Install Python Dependencies

```bash
pip install lightgbm scikit-learn
```

## Interpreting Results

Each model type in the report shows:
- **Test Cases**: Number of random inputs tested (1,000+)
- **Max Abs Diff**: Largest absolute difference between go-lgbm and leaves predictions
- **Mean Abs Diff**: Average absolute difference
- **Status**: PASS (within 1e-10 tolerance) or FAIL
