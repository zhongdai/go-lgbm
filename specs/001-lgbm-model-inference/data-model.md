# Data Model: LightGBM Model Inference Library

**Date**: 2026-02-13
**Feature**: `001-lgbm-model-inference`

## Entities

### Model

The top-level loaded model. Immutable after construction.

| Field | Type | Description |
|-------|------|-------------|
| Version | string | Format version ("v3" or "v4") |
| NumClasses | int | Output classes (1 for binary/regression/ranking) |
| NumTreesPerIteration | int | Trees per boosting round (equals NumClasses for multiclass) |
| NumFeatures | int | Expected feature count (max_feature_idx + 1) |
| Objective | ObjectiveType | Parsed objective (binary, regression, multiclass, ranking, etc.) |
| AverageOutput | bool | True for random forest (average vs sum) |
| Trees | []Tree | Ordered list of decision trees |
| FeatureNames | []string | Optional feature names from training |
| Transform | TransformFunc | Post-prediction transformation (sigmoid, softmax, identity, exp) |

**Invariants**:
- `len(Trees) % NumTreesPerIteration == 0`
- NumFeatures > 0
- NumClasses >= 1
- Trees is non-empty

### Tree

A single decision tree. All arrays are read-only after parsing.

| Field | Type | Description |
|-------|------|-------------|
| NumLeaves | int | Count of leaf nodes |
| SplitFeatures | []int | Feature index per internal node |
| Thresholds | []float64 | Split threshold per internal node |
| DecisionTypes | []uint8 | Bit flags per internal node (categorical, default direction) |
| LeftChildren | []int | Left child index per internal node |
| RightChildren | []int | Right child index per internal node |
| LeafValues | []float64 | Prediction value per leaf |
| Shrinkage | float64 | Learning rate multiplier |
| CatBoundaries | []int | Bitset boundary indices (only if categorical splits) |
| CatThresholds | []uint32 | Concatenated category bitsets (only if categorical splits) |

**Invariants**:
- `NumLeaves >= 1`
- Internal node count = `NumLeaves - 1`
- `len(SplitFeatures) == NumLeaves - 1`
- `len(LeafValues) == NumLeaves`
- Negative child indices are leaf references: leaf index = `-(child + 1)`
- Non-negative child indices are internal node references

### ObjectiveType (Enumeration)

| Value | Model Header | Transformation |
|-------|-------------|----------------|
| Binary | `binary`, `binary sigmoid:1` | Sigmoid |
| Regression | `regression`, `regression_l2`, `mean_squared_error`, `mse` | Identity |
| Multiclass | `multiclass`, `multiclassova` | Softmax |
| Ranking | `lambdarank`, `rank_xendcg` | Identity (raw scores) |
| Poisson | `poisson` | Exponential |
| Gamma | `gamma` | Exponential |
| Tweedie | `tweedie` | Exponential |

### DecisionType Bit Flags

| Bit | Mask | Name | Meaning |
|-----|------|------|---------|
| 0 | 0x01 | Categorical | 1 = categorical split, 0 = numerical split |
| 1 | 0x02 | DefaultLeft | 1 = NaN/missing goes left, 0 = goes right |

## Entity Relationships

```
Model 1──* Tree
Model 1──1 ObjectiveType
Model 1──1 TransformFunc (derived from ObjectiveType)
Tree *──* Node (implicit via arrays, not a separate struct)
```

## Prediction Flow

```
Input: []float64 (feature vector)
  │
  ├─ Validate: len(input) == Model.NumFeatures
  │
  ├─ For each Tree:
  │    ├─ Traverse from root (node 0)
  │    ├─ At each internal node:
  │    │    ├─ If NaN → follow DecisionType default direction
  │    │    ├─ If categorical → check bitset membership
  │    │    └─ If numerical → compare <= threshold
  │    ├─ Reach leaf → accumulate LeafValues[leafIdx] * Shrinkage
  │    └─ Accumulate into raw_scores[treeIdx % NumTreesPerIteration]
  │
  └─ Apply Transform(raw_scores) → output
```

## Validation Rules

| Rule | When | Error |
|------|------|-------|
| Version must be "v3" or "v4" | Model load | "unsupported LightGBM version: {version}" |
| Feature vector length must match | Predict | "expected {N} features, got {M}" |
| Tree arrays must be internally consistent | Model load | "tree {N}: array length mismatch" |
| NumLeaves >= 1 | Model load | "tree {N}: invalid num_leaves" |
| No out-of-range feature indices in splits | Model load | "tree {N}: split feature index {F} exceeds max {M}" |
