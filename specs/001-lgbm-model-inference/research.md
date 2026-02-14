# LightGBM Text Model File Format Research

## Executive Summary

This document summarizes research on the LightGBM text model file format (.txt), covering structure, versioning, tree representation, and categorical feature encoding. The research is based on:
- LightGBM official source code (C++)
- dmitryikh/leaves Go library implementation
- LightGBM documentation and GitHub issues

## 1. Model File Structure

### 1.1 Overall Format

A LightGBM text model file consists of four main sections in this order:

```
[Header Section]
<blank line>
[Tree Section - repeated for each tree]
[Feature Importances Section - optional]
[Parameters Section - optional]
```

### 1.2 Header Section

The header contains metadata written as `key=value` pairs (one per line):

**Required Fields:**
- `version` - Format version identifier (v2, v3, or v4)
- `num_class` - Number of output classes (1 for binary/regression, >1 for multiclass)
- `num_tree_per_iteration` - Number of trees per boosting iteration
- `label_index` - Label column index
- `max_feature_idx` - Maximum feature index in the dataset
- `tree_sizes` - Space-delimited string indicating number of trees

**Optional Fields:**
- `objective` - Objective function specification (e.g., "binary", "regression", "multiclass", "poisson")
- `average_output` - Boolean flag (distinguishes random forest from gradient boosting)
- `feature_names` - Space-separated list of feature names
- `monotone_constraints` - Space-separated monotone constraint values
- `feature_infos` - Space-separated feature information strings

**Example:**
```
tree
version=v3
num_class=1
num_tree_per_iteration=1
label_index=0
max_feature_idx=29
objective=binary sigmoid:1
tree_sizes=1 1 1 1 1
feature_names=feature_0 feature_1 ... feature_29

```

### 1.3 Tree Section

Each tree starts with `Tree=N` where N is the zero-based tree index, followed by tree-specific fields:

**Standard Fields (All Trees):**
1. `num_leaves` - Number of leaf nodes (minimum 1)
2. `num_cat` - Count of categorical features used in this tree
3. `split_feature` - Space-separated array of feature indices for internal nodes
4. `split_gain` - Space-separated array of information gain values for splits
5. `threshold` - Space-separated array of split thresholds (double precision)
6. `decision_type` - Space-separated array of decision type flags (int8)
7. `left_child` - Space-separated array of left child node indices
8. `right_child` - Space-separated array of right child node indices
9. `leaf_value` - Space-separated array of prediction values for leaves
10. `leaf_weight` - Space-separated array of weights (sum of hessians)
11. `leaf_count` - Space-separated array of sample counts at leaves
12. `internal_value` - Space-separated array of values for internal nodes
13. `internal_weight` - Space-separated array of weights for internal nodes
14. `internal_count` - Space-separated array of sample counts at internal nodes
15. `shrinkage` - Learning rate multiplier for this tree (typically 1.0 or the learning_rate)

**Conditional Fields:**

If `num_cat > 0` (categorical splits present):
- `cat_boundaries` - Boundaries separating categorical threshold groups
- `cat_threshold` - Categorical feature thresholds as uint32 bitsets

If tree uses linear models at leaves (`is_linear=1`):
- `is_linear` - Boolean flag (0 or 1)
- `leaf_const` - Constant terms for linear models at each leaf
- `num_features` - Number of features used per leaf in linear models
- `leaf_features` - Feature indices used in linear models
- `leaf_coeff` - Linear coefficients for features

**Example:**
```
Tree=0
num_leaves=4
num_cat=0
split_feature=1 2 2
split_gain=0.568011 0.483606 0.45669
threshold=0.73144941452196321 0.90708366268745222 0.85551601478390116
decision_type=2 2 2
left_child=1 -1 -2
right_child=2 -3 -4
leaf_value=0.49510661266514339 0.50645382200299838 0.50688948369558862 0.49040602357823876
leaf_weight=326 114 39 21
leaf_count=326 114 39 21
internal_value=0.498415 0.496366 0.503957
internal_weight=0 365 135
internal_count=500 365 135
is_linear=0
shrinkage=1

```

### 1.4 Feature Importances Section (Optional)

```
feature_importances:
feature_0=123.456
feature_1=78.901
...
```

### 1.5 Parameters Section (Optional)

```
parameters:
[original training configuration parameters]
end of parameters
```

## 2. Version Differences: v3 vs v4

### 2.1 Format Version History

- **v2**: Used in early LightGBM versions (pre-v3.0)
- **v3**: Current stable format used in LightGBM 3.x
- **v4**: Introduced in LightGBM 4.0.0

### 2.2 Compatibility Issues

**Key Finding:** Model files created with LightGBM v3.3.0 are **not compatible** with v4.0.0.

**Root Cause:**
The incompatibility stems from parameter schema changes, not structural format changes:
- LightGBM v4.0.0 introduced stricter parameter validation
- When loading v3 models, v4 fails to handle deprecated/removed parameters:
  - `max_conflict_rate`
  - `sparse_threshold`
  - `max_position`
  - `lambdamart_norm`

**Resolution:**
- Pull Request #6126 fixed this by implementing logic to "ignore unknown parameters when loading from model file"
- LightGBM v4.0.0+ can now gracefully skip unrecognized settings from older models

### 2.3 Format Differences

**Text Format Structure:** The core text format structure (header fields, tree serialization) remains **largely compatible** between v3 and v4.

**Key Changes:**
1. **Header version field**: `version=v3` → `version=v4`
2. **Parameter handling**: v4 is more permissive, ignoring unknown parameters
3. **API changes**: Python API method `setinfo()` replaced by `set_field()` (not model format change)

**Parser Compatibility:**
The dmitryikh/leaves library accepts both v2 and v3 formats with identical parsing logic, suggesting compatible structures. The library would need updates to handle v4-specific changes if any exist beyond parameter handling.

## 3. Tree Representation

### 3.1 Node Indexing System

**Internal Nodes:** Non-negative integers (0, 1, 2, ...)
**Leaf Nodes:** Negative integers that index into the `leaf_value` array
- `-1` → `leaf_value[0]`
- `-2` → `leaf_value[1]`
- `-3` → `leaf_value[2]`
- etc.

### 3.2 Tree Traversal Logic

For prediction, navigate from root (node 0):
1. Check if node index is negative → it's a leaf, return `leaf_value[-(index+1)]`
2. If node index is non-negative, it's a split node:
   - Get `feature_idx = split_feature[node]`
   - Get `threshold_val = threshold[node]`
   - Get `decision = decision_type[node]`
   - Check feature value against threshold using decision type
   - Navigate to `left_child[node]` or `right_child[node]`

### 3.3 Split Representation

**Numerical Splits:**
- Use `<=` comparison (LightGBM standard)
- `threshold[i]` contains the split point as a double
- If `feature_value <= threshold`, go left; otherwise go right

**Feature Indices:**
- `split_feature[i]` references the column index in the input data
- Must match the feature order used during training

### 3.4 Decision Type Encoding (decision_type field)

The `decision_type` is an int8 bit field encoding multiple properties:

**Bit Masks (from LightGBM source code):**
```cpp
#define kCategoricalMask (1)    // Bit 0: 1 = categorical split, 0 = numerical split
#define kDefaultLeftMask (2)    // Bit 1: 1 = missing values go left, 0 = go right
```

**Encoding:**
- `decision_type & 1`: Check if categorical (1) or numerical (0)
- `decision_type & 2`: Check default direction for missing values

**Common Values:**
- `2` = Numerical split, missing values go left
- `0` = Numerical split, missing values go right
- `3` = Categorical split, missing values go left
- `1` = Categorical split, missing values go right

### 3.5 Missing Value Handling

**Three mechanisms control missing data routing:**

1. **`missingZero` flag**: Treats zero as missing
2. **`missingNan` flag**: Treats NaN as missing
3. **`defaultLeft` flag** (from decision_type bit 1): Routes missing values leftward when set

**Default Behavior:**
- LightGBM uses NA/NaN to represent missing values by default
- During training, the best missing value direction is learned per split
- The learned direction is encoded in the `decision_type` field

**Configuration:**
- `use_missing=true` (default): Enable missing value handling
- `zero_as_missing=true`: Treat zeros as missing
- `use_missing=false`: Disable special handling

## 4. Categorical Split Encoding

### 4.1 Overview

LightGBM handles categorical features natively without requiring one-hot encoding. Categorical splits partition categories into two subsets using bitset encoding.

### 4.2 Bitset Representation

**Storage Format:**
- Each category is represented by a bit in a bitset
- Bitsets are stored as arrays of uint32 values
- Category k goes left if bit k in the bitset is set to 1

**Example:**
If categories {0, 2, 5} go left and {1, 3, 4} go right:
- Bitset: `00100101` (binary) = bit 0, 2, 5 are set
- Stored as uint32 values in `cat_threshold`

### 4.3 cat_threshold and cat_boundaries

**`cat_threshold`:**
- Array of uint32 values representing bitsets
- Contains all categorical split bitsets concatenated

**`cat_boundaries`:**
- Array of integers defining boundaries between different nodes' bitsets
- Example: `cat_boundaries = [0, 3, 7]` means:
  - Node 0's bitset: `cat_threshold[0:3]` (3 uint32 values = 96 bits)
  - Node 1's bitset: `cat_threshold[3:7]` (4 uint32 values = 128 bits)

**Bitset Size Calculation:**
- Each uint32 represents 32 bits (32 categories)
- For N categories, need ceil(N/32) uint32 values
- A bitset with n uint32 values can represent up to 32*n categories

### 4.4 Category Membership Check

To check if category value `k` goes left at node `i`:

```go
// Get bitset boundaries for node i
start := cat_boundaries[i]
end := cat_boundaries[i+1]
bitset := cat_threshold[start:end]

// Check if bit k is set
uint32_idx := k / 32
bit_idx := k % 32
goes_left := (bitset[uint32_idx] & (1 << bit_idx)) != 0
```

### 4.5 Categorical Feature Requirements

From LightGBM documentation:
- Features must be encoded as **non-negative integers**
- Values must be less than `Int32.MaxValue` (2,147,483,647)
- Best performance with **contiguous range starting from zero**
- Negative values are treated as missing
- Floating-point values are rounded toward zero

### 4.6 Optimal Split Algorithm

LightGBM uses the **Fisher (1958)** algorithm:
1. Sort histogram by accumulated values (`sum_gradient / sum_hessian`)
2. Find best split on sorted histogram (like numerical features)
3. This often outperforms one-hot encoding and other categorical methods

## 5. Objective Types and Transformations

### 5.1 Objective Field Format

The `objective` header field specifies the loss function and output transformation:

**Common Formats:**
- `regression` - Mean squared error, no transformation
- `binary` or `binary sigmoid:1` - Logistic transformation
- `multiclass num_class:N` - Softmax transformation
- `poisson` - Exponential transformation
- `gamma` - Exponential transformation
- `tweedie` - Exponential transformation

### 5.2 Transformation Functions

**No Transformation (Raw Output):**
- Regression tasks
- Output = raw tree prediction sum

**Logistic (Sigmoid) Transformation:**
- Binary classification
- `output = 1 / (1 + exp(-raw_score))`
- Converts log-odds to probability [0, 1]

**Exponential Transformation:**
- Poisson, Gamma, Tweedie objectives
- `output = exp(raw_score)`
- Ensures positive predictions

**Softmax Transformation:**
- Multiclass classification
- `output[i] = exp(raw_score[i]) / sum(exp(raw_score[j]) for all j)`
- Converts to probability distribution across classes

### 5.3 Model Type Identification

**Binary Classification:**
- `num_class=1`
- `num_tree_per_iteration=1`
- `objective` contains "binary"

**Multiclass Classification:**
- `num_class=K` (number of classes)
- `num_tree_per_iteration=K`
- `objective` contains "multiclass"
- Trees are grouped: first K trees for iteration 0, next K for iteration 1, etc.

**Regression:**
- `num_class=1`
- `num_tree_per_iteration=1`
- `objective` contains "regression" or similar (poisson, gamma, tweedie)

**Ranking:**
- Uses specialized objective like "lambdarank"
- May have ranking-specific parameters

### 5.4 average_output Flag

**Purpose:** Distinguishes Random Forest from Gradient Boosting

**When present:**
- Random Forest: `average_output` is set (divide by number of trees)
- Gradient Boosting: `average_output` is absent (sum all trees)

## 6. Key Header Fields Summary

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Format version (v2, v3, or v4) |
| `num_class` | int | Number of classes (1 for binary/regression) |
| `num_tree_per_iteration` | int | Trees per boosting round (typically equals num_class) |
| `label_index` | int | Index of label column in training data |
| `max_feature_idx` | int | Maximum feature index (number of features - 1) |
| `objective` | string | Loss function and transformation specification |
| `average_output` | flag | If present, average predictions (random forest) |
| `feature_names` | string | Space-separated feature names |
| `monotone_constraints` | string | Space-separated constraint values |
| `tree_sizes` | string | Space-separated tree sizes |

## 7. Implementation Considerations

### 7.1 Parser Requirements

A complete parser must:
1. Read header parameters until blank line
2. Validate required fields (version, num_class, num_tree_per_iteration, etc.)
3. Parse each tree section starting with `Tree=N`
4. Handle both numerical and categorical splits
5. Support missing value handling via decision_type
6. Apply appropriate transformation based on objective
7. Handle both v3 and v4 formats (ignore unknown parameters in v4)

### 7.2 Tree Array Sizes

For a tree with `n` internal nodes and `m` leaves (where `m = num_leaves`):
- Total nodes: `n + m`
- Relationship: `m = n + 1` (binary tree property)

**Array Lengths:**
- `split_feature`: length `n`
- `split_gain`: length `n`
- `threshold`: length `n`
- `decision_type`: length `n`
- `left_child`: length `n`
- `right_child`: length `n`
- `leaf_value`: length `m`
- `leaf_weight`: length `m`
- `leaf_count`: length `m`
- `internal_value`: length `n`
- `internal_weight`: length `n`
- `internal_count`: length `n`

### 7.3 Prediction Algorithm

```
function predict(model, features):
    raw_scores = array of size num_class, initialized to 0

    for tree_idx from 0 to total_trees:
        class_idx = tree_idx % num_tree_per_iteration
        node = 0  // start at root

        while node >= 0:  // while not a leaf
            feature_idx = tree.split_feature[node]
            threshold = tree.threshold[node]
            decision = tree.decision_type[node]
            feature_value = features[feature_idx]

            // Handle missing values
            if is_missing(feature_value, decision):
                if decision & kDefaultLeftMask:
                    node = tree.left_child[node]
                else:
                    node = tree.right_child[node]
                continue

            // Handle categorical splits
            if decision & kCategoricalMask:
                if category_goes_left(feature_value, node, tree):
                    node = tree.left_child[node]
                else:
                    node = tree.right_child[node]
            else:
                // Numerical split
                if feature_value <= threshold:
                    node = tree.left_child[node]
                else:
                    node = tree.right_child[node]

        // node is negative, it's a leaf
        leaf_idx = -(node + 1)
        raw_scores[class_idx] += tree.leaf_value[leaf_idx] * tree.shrinkage

    // Apply transformation based on objective
    return apply_transformation(raw_scores, model.objective, model.average_output)
```

## 8. Reference Implementation: dmitryikh/leaves

The [leaves Go library](https://github.com/dmitryikh/leaves) provides a complete reference implementation:

**Key Files:**
- `lgensemble_io.go`: Header and model-level parsing
- `lgtree.go`: Tree structure and traversal
- `testdata/`: Example model files for testing

**Features:**
- Supports LightGBM v2 and v3 formats
- Handles categorical splits with bitsets
- Implements missing value logic
- Applies transformations (sigmoid, softmax, exp)
- Pure Go implementation with no C dependencies

**Limitations:**
- May not support v4-specific features yet
- Limited to prediction (no training)

## 9. Additional Resources

### Documentation
- [LightGBM Official Docs - Advanced Topics](https://lightgbm.readthedocs.io/en/latest/Advanced-Topics.html)
- [LightGBM Parameters](https://lightgbm.readthedocs.io/en/latest/Parameters.html)
- [LightGBM Features](https://lightgbm.readthedocs.io/en/latest/Features.html)

### Source Code
- [gbdt_model_text.cpp](https://github.com/microsoft/LightGBM/blob/master/src/boosting/gbdt_model_text.cpp) - Model serialization
- [tree.cpp](https://github.com/microsoft/LightGBM/blob/master/src/io/tree.cpp) - Tree ToString() method
- [tree.h](https://github.com/Microsoft/LightGBM/blob/master/include/LightGBM/tree.h) - Tree structure definition

### GitHub Issues
- [v3 model file not compatible with v4 #6082](https://github.com/microsoft/LightGBM/issues/6082)
- [Missing values: NaN prediction #6139](https://github.com/microsoft/LightGBM/issues/6139)
- [Large model size and cat_threshold #4237](https://github.com/microsoft/LightGBM/issues/4237)

### Libraries
- [dmitryikh/leaves](https://github.com/dmitryikh/leaves) - Go implementation
- [leaves documentation](https://pkg.go.dev/github.com/dmitryikh/leaves)

---

**Document Version:** 1.0
**Date:** 2026-02-13
**Research Duration:** Comprehensive analysis of LightGBM source code, documentation, and reference implementations
