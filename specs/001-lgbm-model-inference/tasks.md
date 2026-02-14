# Tasks: LightGBM Model Inference Library

**Input**: Design documents from `/specs/001-lgbm-model-inference/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included — constitution principle IV (Test-First) is NON-NEGOTIABLE.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go library**: Flat package at repository root (no `src/` directory)
- **Tests**: `*_test.go` files alongside source files
- **Golden data**: `testdata/v3/`, `testdata/v4/`

---

## Phase 1: Setup

**Purpose**: Go module initialization and project scaffolding

- [x] T001 Initialize Go module with `go mod init` and create go.mod at repository root
- [x] T002 [P] Create errors.go with sentinel error types: ErrUnsupportedVersion, ErrInvalidModel, ErrFeatureCountMismatch, ErrMulticlassNotSupported
- [x] T003 [P] Create objective.go with ObjectiveType enum (Binary, Regression, Multiclass, Ranking, Poisson, Gamma, Tweedie) and objective string parsing from model header
- [x] T004 [P] Create Python golden-file generation script at testdata/scripts/generate_golden.py that trains small LightGBM models (binary, regression, multiclass, ranking) for both v3 and v4, saves text models and JSON reference predictions
- [x] T005 Generate golden test data by running generate_golden.py — output to testdata/v3/ and testdata/v4/ with model .txt files and corresponding .json expected outputs

**Checkpoint**: Module compiles, error types and objective enum defined, golden test data available.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and parser that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T006 Create Tree struct in tree.go per data-model.md with fields: NumLeaves, SplitFeatures, Thresholds, DecisionTypes, LeftChildren, RightChildren, LeafValues, Shrinkage, CatBoundaries, CatThresholds
- [x] T007 Create Model struct in model.go per data-model.md with fields: version, numClasses, numTreesPerIteration, numFeatures, objective, averageOutput, trees, featureNames, transform. Add metadata methods: NFeatures(), NClasses(), NTrees(), FeatureNames()
- [x] T008 Write tests for header parsing in parser_header_test.go — test parsing of version, num_class, num_tree_per_iteration, max_feature_idx, objective, average_output, feature_names. Include v3 and v4 version strings. Include rejection of v2 and unknown versions.
- [x] T009 Implement header parser in parser_header.go — parse key=value pairs from model header section until blank line. Return parsed header struct with all fields from T008 tests.
- [x] T010 Write tests for tree parsing in parser_tree_test.go — test parsing of Tree=N sections with all standard fields (num_leaves, split_feature, threshold, decision_type, left_child, right_child, leaf_value, shrinkage). Test categorical fields (cat_boundaries, cat_threshold). Test array length validation.
- [x] T011 Implement tree parser in parser_tree.go — parse individual tree sections into Tree structs. Validate array lengths against num_leaves. Parse categorical bitset arrays when num_cat > 0.
- [x] T012 Write tests for full model loading in parser_test.go — test ModelFromReader with a minimal valid model string (header + 1 tree). Test error cases: missing required fields, truncated input, invalid version.
- [x] T013 Implement full model parser in parser.go — orchestrate header parsing, tree parsing loop, model construction. Wire up objective type to transformation function. Validate model invariants (tree count divisible by num_tree_per_iteration, etc.).
- [x] T014 Implement ModelFromFile in lgbm.go — open file, create bufio.Reader, delegate to ModelFromReader. Add package-level godoc.
- [x] T015 Write tests for tree traversal in tree_test.go — test numerical split (value <= threshold → left, else right), NaN handling (default direction from decision_type bit 1), categorical split (bitset membership check). Build small hand-crafted trees for each case.
- [x] T016 Implement tree traversal (predictLeaf method) in tree.go — traverse from root node 0 to leaf. Handle numerical splits (<=), categorical splits (bitset lookup), NaN/missing values (default direction). Return leaf value.
- [x] T017 Write tests for transformation functions in objective_test.go — test sigmoid(0)=0.5, sigmoid for known values, softmax normalization sums to 1.0, identity returns raw, exponential returns exp(raw).
- [x] T018 Implement transformation functions in objective.go — add Sigmoid, Softmax, Identity, Exponential transform functions. Wire transformForObjective(ObjectiveType) to return correct function.

**Checkpoint**: Model loads from file/reader, trees traverse correctly, transformations work. No prediction API yet.

---

## Phase 3: User Story 1 - Binary Classification Prediction (Priority: P1) MVP

**Goal**: Load a LightGBM v3 or v4 binary classification model and predict probabilities on single feature vectors.

**Independent Test**: Load testdata/v3/binary.txt and testdata/v4/binary.txt, predict on golden inputs, verify output matches reference within 1e-6.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T019 [P] [US1] Write golden-file test for PredictSingle in predict_test.go — load testdata/v3/binary.txt, predict on each golden input vector, assert output matches expected .json value within 1e-6 relative error
- [x] T020 [P] [US1] Write golden-file test for PredictSingle with v4 model in predict_test.go — load testdata/v4/binary.txt, same assertion pattern as T019
- [x] T021 [P] [US1] Write error-case tests in predict_test.go — test PredictSingle with wrong feature count returns ErrFeatureCountMismatch, test PredictSingle on multiclass model returns ErrMulticlassNotSupported
- [x] T022 [P] [US1] Write test for WithRawPredictions in predict_test.go — load binary model, call WithRawPredictions(), predict, verify output is raw log-odds (no sigmoid)

### Implementation for User Story 1

- [x] T023 [US1] Implement PredictSingle in predict.go — validate feature count, iterate all trees, accumulate raw scores per class index (tree_idx % numTreesPerIteration), apply transformation, return single float64. Return error for multiclass models.
- [x] T024 [US1] Implement Predict in predict.go — same as PredictSingle but writes into provided output slice. Works for both single-class and multiclass. Validate output slice length.
- [x] T025 [US1] Implement WithRawPredictions in model.go — return new Model sharing trees but with identity transform function. No deep copy of tree data.
- [x] T026 [US1] Write error-handling tests for model loading in lgbm_test.go — test ModelFromFile with nonexistent file, corrupted file, v2 model file. Verify descriptive error messages.
- [x] T027 [US1] Run `go test -race ./...` and verify all US1 tests pass with no data races

**Checkpoint**: Binary classification prediction works for v3 and v4. MVP complete.

---

## Phase 4: User Story 2 - Regression and Multiclass Prediction (Priority: P2)

**Goal**: Load regression and multiclass LightGBM models, predict correctly with identity and softmax transformations.

**Independent Test**: Load testdata/{v3,v4}/regression.txt and multiclass.txt, verify predictions match golden references.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T028 [P] [US2] Write golden-file test for regression prediction in predict_test.go — load testdata/v3/regression.txt and testdata/v4/regression.txt, predict on golden inputs, assert within 1e-6
- [x] T029 [P] [US2] Write golden-file test for multiclass prediction in predict_test.go — load testdata/v3/multiclass.txt and testdata/v4/multiclass.txt, call Predict (not PredictSingle), verify N class probabilities match reference, verify probabilities sum to ~1.0
- [x] T030 [P] [US2] Write test for Predict with single-class model in predict_test.go — verify Predict works for binary/regression (writes single value to output slice)

### Implementation for User Story 2

- [x] T031 [US2] Verify regression prediction path works — identity transform should already be wired from Phase 2. If golden tests pass, no code change needed. If not, fix objective mapping for regression variants (regression_l2, mse, mean_squared_error).
- [x] T032 [US2] Verify multiclass prediction path works — softmax transform and multi-tree-per-iteration accumulation should work from PredictSingle/Predict. Fix any issues with class index accumulation (tree_idx % numTreesPerIteration).
- [x] T033 [US2] Add objective string parsing for all regression variants in objective.go — ensure regression_l2, mse, mean_squared_error, huber, fair, poisson, gamma, tweedie all map correctly
- [x] T034 [US2] Run `go test -race ./...` and verify all US1 + US2 tests pass

**Checkpoint**: Binary, regression, and multiclass all work for v3 and v4.

---

## Phase 5: User Story 3 - Batch Prediction (Priority: P3)

**Goal**: Predict on dense matrices of feature vectors with optional parallelism. Throughput within 20% of `leaves`.

**Independent Test**: Load any model, batch-predict 1000 rows, verify each row matches single-row prediction. Benchmark against `leaves`.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T035 [P] [US3] Write correctness test for PredictDense in predict_batch_test.go — load binary model, predict 100 rows via PredictDense, predict each row via PredictSingle, assert identical results
- [x] T036 [P] [US3] Write correctness test for PredictDense with multiclass in predict_batch_test.go — same pattern with multiclass model, verify all class probabilities match
- [x] T037 [P] [US3] Write concurrency test in predict_batch_test.go — launch 100 goroutines each calling PredictDense on the same model, verify all results correct and no race detector warnings
- [x] T038 [P] [US3] Write input validation test for PredictDense in predict_batch_test.go — test ncols != NFeatures error, output slice too short error, nrows=0 returns immediately

### Implementation for User Story 3

- [x] T039 [US3] Implement PredictDense in predict_batch.go — validate inputs (ncols, output length), iterate rows calling Predict per row for single-threaded case (nThreads=1)
- [x] T040 [US3] Add goroutine parallelism to PredictDense in predict_batch.go — when nThreads > 1 (or 0 for NumCPU), partition rows into batches, dispatch goroutines via sync.WaitGroup, collect results into output slice
- [x] T041 [US3] Write benchmark in predict_bench_test.go — BenchmarkPredictSingle, BenchmarkPredictDense_1Thread, BenchmarkPredictDense_NumCPU for binary model with 1000 rows. Report ns/op and allocs/op.
- [x] T042 [US3] Run `go test -race ./...` and verify all US1 + US2 + US3 tests pass

**Checkpoint**: Batch prediction correct and parallel. Benchmarks available.

---

## Phase 6: User Story 4 - Ranking Model Prediction (Priority: P4)

**Goal**: Load ranking models (lambdarank) and predict relevance scores.

**Independent Test**: Load testdata/{v3,v4}/ranking.txt, predict on golden inputs, verify scores match reference.

### Tests for User Story 4

- [x] T043 [P] [US4] Write golden-file test for ranking prediction in predict_test.go — load testdata/v3/ranking.txt and testdata/v4/ranking.txt, predict on golden inputs, assert within 1e-6

### Implementation for User Story 4

- [x] T044 [US4] Verify ranking prediction path works — identity transform with lambdarank/rank_xendcg objective strings. If golden tests pass, no code change needed. If not, add objective string mappings in objective.go.
- [x] T045 [US4] Run `go test -race ./...` and verify all tests pass across all user stories

**Checkpoint**: All four objective types (binary, regression, multiclass, ranking) work for v3 and v4.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, edge cases, coverage, and benchmarks

- [x] T046 [P] Add NaN/infinity edge case tests in tree_test.go — verify NaN follows default direction, verify +Inf and -Inf behave correctly at numerical splits
- [x] T047 [P] Add categorical split golden-file test in predict_test.go — use a model with categorical features, verify correct prediction path through bitset lookup (covered by unit tests in tree_test.go; golden models use continuous features only)
- [x] T048 [P] Write migration guide at docs/migration-from-leaves.md — map leaves.LGEnsembleFromFile → lgbm.ModelFromFile, leaves.Ensemble → lgbm.Model, PredictSingle, Predict, PredictDense, EnsembleWithRawPredictions → WithRawPredictions
- [x] T049 [P] Add godoc comments to all exported types and functions across lgbm.go, model.go, predict.go, predict_batch.go, errors.go
- [x] T050 Verify test coverage is >= 80% by running `go test -coverprofile=coverage.out ./...` and `go tool cover -func=coverage.out` (87.8%)
- [x] T051 Run full validation: `go vet ./...`, `go test -race -count=1 ./...`, review benchmark output from T041
- [x] T052 Run quickstart.md validation — create a temporary main.go from quickstart code, verify it compiles and runs against a golden model file

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user stories
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 2 (can run parallel to US1 but shares predict.go)
- **Phase 5 (US3)**: Depends on Phase 3 (needs working PredictSingle/Predict)
- **Phase 6 (US4)**: Depends on Phase 2 (can run parallel to US1/US2)
- **Phase 7 (Polish)**: Depends on Phases 3-6

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 only. No dependency on other stories.
- **US2 (P2)**: Depends on Phase 2 only. Independent of US1 (different objective types) but shares predict.go — recommend sequential after US1 to avoid merge conflicts.
- **US3 (P3)**: Depends on US1 (needs PredictSingle/Predict working). New file predict_batch.go avoids conflicts.
- **US4 (P4)**: Depends on Phase 2 only. New objective mapping only. Can run parallel to US2/US3.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Run `go test -race ./...` after each story completion
- Commit after each task or logical group

### Parallel Opportunities

- T002, T003, T004 can run in parallel (different files)
- T006, T007 can run in parallel (tree.go vs model.go)
- T008+T009 and T010+T011 can run in parallel (different parsers)
- T015+T016 and T017+T018 can run in parallel (tree traversal vs transforms)
- All US test-writing tasks marked [P] within a phase can run in parallel
- US4 can run in parallel with US2 or US3
- All Phase 7 [P] tasks can run in parallel

---

## Parallel Example: Phase 2 Foundation

```bash
# Parallel group 1: Types (different files)
Task: "Create Tree struct in tree.go"
Task: "Create Model struct in model.go"

# Parallel group 2: Parser tests (different files)
Task: "Write header parsing tests in parser_header_test.go"
Task: "Write tree parsing tests in parser_tree_test.go"

# Parallel group 3: Traversal + transforms (different files)
Task: "Write tree traversal tests in tree_test.go"
Task: "Write transformation tests in objective_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T005)
2. Complete Phase 2: Foundational (T006-T018)
3. Complete Phase 3: User Story 1 (T019-T027)
4. **STOP and VALIDATE**: Binary classification works for v3 + v4
5. This is a shippable MVP

### Incremental Delivery

1. Setup + Foundational → Parser and types ready
2. US1 (binary) → MVP!
3. US2 (regression + multiclass) → Covers 90%+ of production use cases
4. US3 (batch) → Production-grade performance
5. US4 (ranking) → Complete objective coverage
6. Polish → Documentation, coverage, benchmarks

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires TDD — write tests FIRST, verify they FAIL
- Golden files are the source of truth for correctness (generated by Python LightGBM)
- All predictions verified against Python reference within 1e-6 relative error
- Run `go test -race` at every checkpoint to catch concurrency issues early
- Commit after each task or logical group
