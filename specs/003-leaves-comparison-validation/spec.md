# Feature Specification: Leaves Comparison Validation

**Feature Branch**: `003-leaves-comparison-validation`
**Created**: 2026-02-15
**Status**: Draft
**Input**: User description: "Create a validation Go program to check the output from go-lgbm and leaves are exact same. Make a new directory, use Python to generate real models, load the same model using go-lgbm and leaves, random create input candidates, make predictions, compare. Use a justfile to easy re-run the test, and generate the report in markdown. Add a section on the readme to link to the comparison report."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Full Comparison Validation (Priority: P1)

As a developer evaluating go-lgbm, I want to run a single command that generates models, makes predictions using both go-lgbm and leaves, compares the results, and produces a report — so I can verify that go-lgbm produces identical output to the established leaves library.

**Why this priority**: This is the core value of the feature. Without the ability to run the comparison and see results, nothing else matters.

**Independent Test**: Can be fully tested by running the validation command and verifying a comparison report is generated with pass/fail results for each model type.

**Acceptance Scenarios**:

1. **Given** the validation directory exists with all required scripts and programs, **When** the user runs the validation command, **Then** models are generated, predictions are made by both libraries, results are compared, and a markdown report is produced.
2. **Given** a successful validation run, **When** the user inspects the report, **Then** each model type shows the number of test cases, maximum difference, and a clear pass/fail status.
3. **Given** both libraries produce identical predictions, **When** the comparison runs, **Then** all model types report as "PASS" with zero or near-zero differences.

---

### User Story 2 - View Comparison Report (Priority: P2)

As a developer considering adopting go-lgbm, I want to view a pre-generated comparison report linked from the project README — so I can see the validation evidence without running anything myself.

**Why this priority**: The report provides trust and transparency. Linking it from the README makes validation results discoverable to all visitors.

**Independent Test**: Can be tested by checking that the README contains a link to the comparison report and that the report is present and readable in the repository.

**Acceptance Scenarios**:

1. **Given** a completed validation run, **When** the report is committed to the repository, **Then** it is accessible via the link in the README.
2. **Given** the README, **When** a user looks for validation evidence, **Then** they find a clearly labeled section linking to the comparison report.

---

### User Story 3 - Re-run Validation After Changes (Priority: P3)

As a maintainer making changes to go-lgbm, I want to re-run the full validation suite with a single command — so I can verify that my changes haven't introduced prediction differences.

**Why this priority**: Ongoing regression prevention is important but depends on the initial validation infrastructure being in place first.

**Independent Test**: Can be tested by modifying a prediction function, re-running validation, and confirming the report reflects any differences introduced.

**Acceptance Scenarios**:

1. **Given** a previous validation run exists, **When** the user re-runs the validation command, **Then** new models are generated, fresh predictions are made, and a new report replaces the old one.
2. **Given** a code change that introduces a prediction difference, **When** validation is re-run, **Then** the report clearly shows which model types have differences and the magnitude.

---

### Edge Cases

- What happens when the model generation step fails (e.g., missing dependencies)?
  - The validation process MUST report the failure clearly and stop, rather than producing a misleading report.
- What happens when one library crashes on a specific model but the other succeeds?
  - The report MUST record the failure for that model type and continue testing remaining models.
- How does the system handle floating-point differences at the edge of precision?
  - A configurable tolerance threshold determines pass/fail. Differences within tolerance are reported but marked as PASS.
- What happens when models have varying numbers of features or classes?
  - Each model type defines its own feature count and class count. Random inputs are generated to match each model's requirements.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The validation suite MUST generate real trained models covering at minimum: binary classification, multiclass classification, regression, and ranking.
- **FR-002**: The validation suite MUST generate random input candidates appropriate for each model type (correct number of features, realistic value ranges).
- **FR-003**: The validation suite MUST load each generated model in both go-lgbm and leaves and produce predictions for the same set of input candidates.
- **FR-004**: The validation suite MUST compare predictions from both libraries numerically and determine pass/fail based on a tolerance threshold.
- **FR-005**: The validation suite MUST produce a markdown-formatted comparison report containing: model type, number of test cases, maximum absolute difference, mean absolute difference, and pass/fail status.
- **FR-006**: The validation suite MUST be runnable with a single command via a justfile recipe.
- **FR-007**: The comparison report MUST be committed to the repository and linked from the project README.
- **FR-008**: The validation suite MUST use a tolerance of 1e-10 for floating-point comparison (effectively requiring exact match within floating-point representation limits).
- **FR-009**: The validation suite MUST test at least 1,000 random input candidates per model type.
- **FR-010**: The validation suite MUST test both single predictions and batch predictions where applicable.

### Key Entities

- **Model**: A trained LightGBM model file in text format. Attributes: objective type (binary, multiclass, regression, ranking), number of features, number of classes (for multiclass), number of trees.
- **Test Case**: A single set of input features paired with predictions from both libraries. Attributes: input vector, go-lgbm prediction, leaves prediction, absolute difference.
- **Comparison Report**: A markdown document summarizing validation results across all model types. Attributes: model type sections, aggregate statistics, pass/fail verdicts, timestamp.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All model types (binary, multiclass, regression, ranking) produce identical predictions (within 1e-10 tolerance) between go-lgbm and leaves across 100% of test cases.
- **SC-002**: The validation suite completes a full run (model generation, prediction, comparison, report) in under 2 minutes.
- **SC-003**: The comparison report is human-readable and contains per-model-type statistics with clear pass/fail indicators.
- **SC-004**: A new developer can run the full validation with a single command after cloning the repository.
- **SC-005**: The README links to the comparison report within one click from the main project page.
- **SC-006**: At least 1,000 test cases per model type are validated, covering 4+ model types.

## Assumptions

- Python with LightGBM and scikit-learn is available in the environment for model generation.
- The leaves library (`github.com/dmitryikh/leaves`) is available as a Go dependency for comparison.
- The justfile tool is available on the developer's machine.
- Both libraries can load the same LightGBM text model format without conversion.
- Random inputs use a fixed seed for reproducibility across runs.
