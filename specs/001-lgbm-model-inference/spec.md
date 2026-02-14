# Feature Specification: LightGBM Model Inference Library

**Feature Branch**: `001-lgbm-model-inference`
**Created**: 2026-02-13
**Status**: Draft
**Input**: User description: "Create a pure-Go LightGBM model loading and inference library as a replacement for github.com/dmitryikh/leaves, supporting LightGBM v3 and v4 model formats with API compatibility."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Load and Predict with a Binary Classification Model (Priority: P1)

A developer has a LightGBM binary classification model (trained with
LightGBM v3 or v4) saved as a text model file. They want to load
the model in their Go service and run predictions on individual
feature vectors, getting probability scores back. Today they use
`leaves` for this; they need the same workflow with `go-lgbm`.

**Why this priority**: Binary classification is the most common
LightGBM use case in production (fraud detection, click prediction,
churn scoring). Without this, the library has no value.

**Independent Test**: Load a known binary classification model file,
pass a feature vector, and verify the predicted probability matches
the reference output from Python LightGBM within 1e-6 tolerance.

**Acceptance Scenarios**:

1. **Given** a LightGBM v3 binary classification text model file,
   **When** a developer loads it and calls predict with a valid
   feature vector, **Then** the returned probability matches the
   Python LightGBM reference output within 1e-6 relative error.
2. **Given** a LightGBM v4 binary classification text model file,
   **When** a developer loads it and calls predict with a valid
   feature vector, **Then** the returned probability matches the
   Python LightGBM reference output within 1e-6 relative error.
3. **Given** an invalid or corrupted model file, **When** a
   developer attempts to load it, **Then** a descriptive error is
   returned (not a panic).

---

### User Story 2 - Load and Predict with Regression and Multiclass Models (Priority: P2)

A developer has LightGBM models for regression (continuous output)
or multiclass classification (multiple class probabilities). They
want to load these models and run predictions, getting raw values
or class probabilities back respectively.

**Why this priority**: Regression and multiclass are the next most
common objective types after binary classification. Supporting
these covers the vast majority of real-world LightGBM deployments.

**Independent Test**: Load known regression and multiclass model
files, predict on test vectors, and verify outputs match Python
LightGBM references within tolerance.

**Acceptance Scenarios**:

1. **Given** a LightGBM v3 or v4 regression model, **When** a
   developer loads it and predicts on a feature vector, **Then**
   the predicted value matches the Python reference within 1e-6.
2. **Given** a LightGBM v3 or v4 multiclass model with N classes,
   **When** a developer predicts on a feature vector, **Then** N
   class probabilities are returned matching the Python reference.
3. **Given** a multiclass model, **When** predictions are run on
   a batch of feature vectors, **Then** each prediction matches
   the reference output independently.

---

### User Story 3 - Batch Prediction for High-Throughput Serving (Priority: P3)

A developer operating a high-throughput Go service needs to predict
on batches of feature vectors efficiently. They want to pass
multiple rows at once and receive all predictions back, with
performance comparable to or better than `leaves`.

**Why this priority**: Production inference services typically
process requests in batches for throughput. This is critical for
adoption in latency-sensitive environments but depends on the
single-prediction path working first.

**Independent Test**: Load a model, predict on a batch of 1000
feature vectors, verify correctness against per-row predictions,
and benchmark throughput against `leaves`.

**Acceptance Scenarios**:

1. **Given** a loaded model and a batch of 1000 feature vectors,
   **When** batch predict is called, **Then** each result matches
   the corresponding single-row prediction exactly.
2. **Given** a loaded model, **When** batch predict is called
   concurrently from multiple goroutines, **Then** all predictions
   are correct and no data races occur.
3. **Given** a loaded model and a batch of feature vectors, **When**
   batch predict is benchmarked, **Then** throughput is within 20%
   of `leaves` on the same hardware.

---

### User Story 4 - Ranking Model Prediction (Priority: P4)

A developer uses LightGBM ranking models (lambdarank, lambdamart)
for search result ordering or recommendation scoring. They want to
load ranking models and predict relevance scores.

**Why this priority**: Ranking is a specialized but important
LightGBM use case. It completes objective-type coverage but is
less common than classification and regression.

**Independent Test**: Load a ranking model, predict on grouped
feature vectors, verify scores match Python reference.

**Acceptance Scenarios**:

1. **Given** a LightGBM v3 or v4 ranking model, **When** a
   developer loads it and predicts on feature vectors, **Then**
   relevance scores match the Python reference within 1e-6.

---

### Edge Cases

- What happens when a model file is from LightGBM v2 or earlier?
  The loader MUST return a clear error indicating the version is
  unsupported rather than producing silent incorrect results.
- What happens when a feature vector has a different number of
  features than the model expects? A descriptive error MUST be
  returned.
- What happens when feature values contain NaN or infinity? The
  library MUST handle these the same way LightGBM does (NaN goes
  to the default child direction in each tree split).
- What happens when the model file is truncated or partially
  written? The loader MUST detect corruption and return an error.
- What happens when categorical features are present in the model?
  The library MUST support categorical splits as defined in the
  LightGBM text model format.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Library MUST parse LightGBM text-format model files
  produced by LightGBM v3.x and v4.x.
- **FR-002**: Library MUST reject model files from unsupported
  LightGBM versions with a clear error message including the
  detected version.
- **FR-003**: Library MUST support binary classification objective
  (sigmoid transformation of raw predictions).
- **FR-004**: Library MUST support regression objectives (identity
  transformation of raw predictions).
- **FR-005**: Library MUST support multiclass classification
  (softmax transformation across class outputs).
- **FR-006**: Library MUST support ranking objectives (raw
  relevance score output).
- **FR-007**: Library MUST support both numerical and categorical
  feature splits during tree traversal.
- **FR-008**: Library MUST handle NaN feature values by following
  the default child direction stored in each tree node.
- **FR-009**: Library MUST support single-row prediction given a
  dense feature vector (slice of float64).
- **FR-010**: Library MUST support batch prediction given multiple
  feature vectors.
- **FR-011**: Loaded models MUST be safe for concurrent prediction
  calls from multiple goroutines without external synchronization.
- **FR-012**: Library MUST be implemented in pure Go with no CGo
  dependencies.
- **FR-013**: All prediction outputs MUST match the official
  LightGBM Python library output within 1e-6 relative error for
  the same model and input.
- **FR-014**: Library MUST validate feature vector length against
  the model's expected feature count and return an error on
  mismatch.

### Key Entities

- **Model**: A loaded LightGBM model consisting of metadata
  (version, objective type, number of features, number of classes)
  and a collection of decision trees. Immutable after loading.
- **Tree**: A single decision tree with internal nodes (split
  conditions on features) and leaf nodes (prediction values).
  Part of a Model.
- **Node**: A decision point in a tree, containing a feature
  index, threshold or categorical set, and child references.
  Leaf nodes contain output values.
- **Prediction**: The output of running inference — a single
  float64 for binary/regression/ranking, or a slice of float64
  for multiclass.

### Assumptions

- The primary model format is LightGBM text format (`.txt` or
  model string). Binary `.bin` format is out of scope for the
  initial release.
- Feature input is always dense float64 vectors. Sparse input
  support may be added later but is not required initially.
- The library is inference-only; training is out of scope.
- Performance parity target is within 20% of `leaves` throughput.
  Exceeding `leaves` performance is desirable but not required.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All golden-file prediction tests pass — outputs
  match Python LightGBM reference within 1e-6 relative error for
  binary classification, regression, multiclass, and ranking
  across both v3 and v4 model formats.
- **SC-002**: A developer can migrate from `leaves` to `go-lgbm`
  by changing import paths and making minimal API adjustments,
  with a documented migration guide.
- **SC-003**: Batch prediction throughput (rows per second) is
  within 20% of `leaves` on the same model and hardware, verified
  by reproducible benchmarks.
- **SC-004**: The library produces correct results under concurrent
  access — 100 goroutines predicting simultaneously on the same
  model with zero data races (verified by race detector).
- **SC-005**: The library installs with a single package manager
  command on Linux, macOS, and Windows without any system-level
  dependencies or build tools beyond the language toolchain.
- **SC-006**: Test coverage is at or above 80% across the entire
  library.
