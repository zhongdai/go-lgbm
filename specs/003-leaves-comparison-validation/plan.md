# Implementation Plan: Leaves Comparison Validation

**Branch**: `003-leaves-comparison-validation` | **Date**: 2026-02-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-leaves-comparison-validation/spec.md`

## Summary

Create a standalone validation suite in a `validation/` directory that trains LightGBM models via Python, loads them in both go-lgbm and leaves, generates random inputs, compares predictions, and produces a markdown comparison report. A justfile recipe provides single-command execution. The report is committed and linked from the project README.

## Technical Context

**Language/Version**: Go 1.21+ (validation program) + Python 3.x (model generation)
**Primary Dependencies**: go-lgbm (this library), github.com/dmitryikh/leaves (comparison target), lightgbm + scikit-learn (Python, model generation)
**Storage**: File-based (model files in text format, JSON test data, markdown report)
**Testing**: The validation suite itself IS the test — it produces a pass/fail report
**Target Platform**: macOS/Linux developer machines
**Project Type**: Single project (validation subdirectory within existing repo)
**Performance Goals**: Full validation run completes in under 2 minutes
**Constraints**: Requires Python with lightgbm and scikit-learn installed
**Scale/Scope**: 4 model types x 1,000+ test cases each = 4,000+ predictions compared

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Leaves Compatibility | PASS | This feature directly validates leaves compatibility by comparing outputs |
| II. Pure Go / No CGo | PASS | The validation Go program uses only go-lgbm and leaves (both pure Go). Python is only for model generation, not inference. |
| III. LightGBM 3 and 4 Only | PASS | Models will be generated using current LightGBM (v4.x), which produces v4 format. v3 format models can be included from existing testdata. |
| IV. Test-First | PASS | The entire feature is a testing/validation tool. The validation program compares against Python-generated reference outputs. |
| V. Idiomatic Go API | PASS | Validation is a standalone program, not library API. No public API surface affected. |
| Technical: Go 1.21+ | PASS | Validation program targets Go 1.21+ |
| Technical: Minimal deps | PASS | Only adds leaves as a dependency in the validation module (separate go.mod), not in the main library |
| Technical: Immutability | PASS | No model state mutation — read-only comparison |

## Project Structure

### Documentation (this feature)

```text
specs/003-leaves-comparison-validation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
validation/
├── justfile             # Recipes: generate-models, validate, report, all
├── generate_models.py   # Python script to train and export LightGBM models
├── models/              # Generated model files (gitignored)
├── testdata/            # Generated test inputs as JSON (gitignored)
├── go.mod               # Separate module to avoid adding leaves to main go.mod
├── go.sum
├── main.go              # Validation program entry point
└── REPORT.md            # Generated comparison report (committed)
```

**Structure Decision**: The validation suite lives in a separate `validation/` directory with its own `go.mod` to keep the `leaves` dependency isolated from the main library. This prevents users of go-lgbm from pulling in leaves as a transitive dependency. Generated model files and test data are gitignored; only the final report is committed.
