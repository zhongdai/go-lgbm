# Tasks: Leaves Comparison Validation

**Input**: Design documents from `/specs/003-leaves-comparison-validation/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, quickstart.md

**Tests**: Not applicable — the entire feature IS a validation/comparison test suite.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Validation suite**: `validation/` at repository root
- **Models output**: `validation/models/` (gitignored)
- **Test data output**: `validation/testdata/` (gitignored)
- **Report**: `validation/REPORT.md` (committed)

---

## Phase 1: Setup

**Purpose**: Create the validation directory structure and initialize the Go module

- [x] T001 Create `validation/` directory structure: `validation/models/`, `validation/testdata/`
- [x] T002 Initialize separate Go module in `validation/go.mod` with module path `github.com/zhongdai/go-lgbm/validation`, adding dependencies on `github.com/zhongdai/go-lgbm` (local replace directive) and `github.com/dmitryikh/leaves`
- [x] T003 Add `.gitignore` entries in `validation/.gitignore` for `models/` and `testdata/` directories

**Checkpoint**: Validation directory ready with isolated Go module.

---

## Phase 2: User Story 1 - Run Full Comparison Validation (Priority: P1) MVP

**Goal**: A single command generates models, runs predictions through both libraries, compares results, and produces a markdown report.

**Independent Test**: Run `just validate` from the repo root and verify `validation/REPORT.md` is produced with pass/fail results for 4 model types.

### Implementation for User Story 1

- [x] T004 [P] [US1] Create Python model generation script in `validation/generate_models.py` — train 4 LightGBM models (binary classification, multiclass classification, regression, ranking) using scikit-learn synthetic datasets with fixed random seed, export each as text format to `validation/models/`, and generate 1,000 random test inputs per model as JSON to `validation/testdata/`
- [x] T005 [P] [US1] Create the Go validation program entry point in `validation/main.go` — load each model file using both go-lgbm (`ModelFromFile`) and leaves (`LGEnsembleFromFile`), read test inputs from JSON, run predictions through both libraries, compute max/mean absolute differences, determine pass/fail (tolerance 1e-10), and write results to `validation/REPORT.md` in markdown format
- [x] T006 [US1] Create `validation/justfile` with recipes: `generate-models` (runs `python3 generate_models.py`), `validate` (runs `go run .`), and `all` (runs generate-models then validate)
- [x] T007 [US1] Add `validate` recipe to root `justfile` that runs `just --justfile validation/justfile all`
- [x] T008 [US1] Run full validation end-to-end: execute `just validate` from repo root, verify `validation/REPORT.md` is generated with all 4 model types showing PASS status

**Checkpoint**: Single-command validation produces a complete comparison report. This is the shippable MVP.

---

## Phase 3: User Story 2 - View Comparison Report (Priority: P2)

**Goal**: The comparison report is committed to the repository and linked from the project README.

**Independent Test**: Check that README contains a "Validation" section with a working link to `validation/REPORT.md`.

### Implementation for User Story 2

- [x] T009 [US2] Commit the generated `validation/REPORT.md` to the repository
- [x] T010 [US2] Add a "Validation" section to `README.md` with a brief description and link to `validation/REPORT.md`

**Checkpoint**: The comparison report is discoverable from the README.

---

## Phase 4: User Story 3 - Re-run Validation After Changes (Priority: P3)

**Goal**: Maintainers can re-run the validation suite to detect regressions.

**Independent Test**: Modify the validation tolerance, re-run, and verify the report updates.

### Implementation for User Story 3

- [x] T011 [US3] Verify re-run behavior: run `just validate` a second time and confirm `validation/REPORT.md` is regenerated with a fresh timestamp, replacing the previous report
- [x] T012 [US3] Add error handling in `validation/main.go` for missing models or test data — print a clear message directing the user to run model generation first
- [x] T013 [US3] Add error handling in `validation/main.go` for individual model failures — record the failure in the report and continue with remaining models

**Checkpoint**: The validation suite is robust for repeated use.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup and validation

- [x] T014 [P] Verify `validation/.gitignore` correctly excludes `models/` and `testdata/` but includes `REPORT.md`
- [x] T015 [P] Verify `validation/go.mod` replace directive works correctly when run from repository root
- [x] T016 Run quickstart.md validation — follow the quickstart steps from a clean state and verify everything works

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (US1 - Full Validation)**: Depends on Phase 1
- **Phase 3 (US2 - Report in README)**: Depends on Phase 2 (needs a generated report to commit)
- **Phase 4 (US3 - Re-run Robustness)**: Depends on Phase 2 (needs working validation to test re-runs)
- **Phase 5 (Polish)**: Depends on Phases 2-4

### User Story Dependencies

- **US1 (Full Validation)**: Independent — only needs setup directory and Go module
- **US2 (Report in README)**: Depends on US1 — needs a generated report to commit and link
- **US3 (Re-run Robustness)**: Depends on US1 — needs working validation to test error handling

### Parallel Opportunities

- T004 (Python script) and T005 (Go program) can be created in parallel — different files, different languages
- T006 (validation justfile) and T007 (root justfile) can run in parallel after T004+T005
- T014 and T015 (polish tasks) can run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create directory and Go module (T001-T003)
2. Complete Phase 2: Python script + Go program + justfile (T004-T008)
3. **STOP and VALIDATE**: Run `just validate` and confirm all 4 model types show PASS
4. This is a shippable MVP — the comparison is proven

### Incremental Delivery

1. Setup → Directory and module ready
2. US1 (Full Validation) → Single-command comparison with report (MVP!)
3. US2 (Report in README) → Report committed and linked for visibility
4. US3 (Re-run Robustness) → Error handling for ongoing use
5. Polish → Final verification

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- The Python script and Go program are the two core deliverables — everything else is wiring
- The `validation/go.mod` uses a `replace` directive to point to the local go-lgbm module (`replace github.com/zhongdai/go-lgbm => ../`)
- Generated model files and test data are gitignored; only `REPORT.md` is committed
- The Python script uses a fixed random seed (42) for reproducibility
- Tolerance of 1e-10 is effectively exact match within floating-point representation
