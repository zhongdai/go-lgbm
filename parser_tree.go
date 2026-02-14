package lgbm

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// parseTree parses a single tree section from a LightGBM text-format model.
// It assumes the "Tree=N" line has already been consumed by the caller.
// Returns a tree struct populated with all fields from the tree section.
func parseTree(scanner *bufio.Scanner) (tree, error) {
	tr := tree{
		shrinkage: 1.0, // default shrinkage value
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Empty line indicates end of tree section
		if line == "" {
			break
		}

		// Split on first '=' to get key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		var err error
		switch key {
		case "num_leaves":
			tr.numLeaves, err = strconv.Atoi(value)
			if err != nil {
				return tree{}, &ModelError{Detail: fmt.Sprintf("invalid num_leaves: %v", err)}
			}

		case "num_cat":
			// Parse but don't store num_cat directly; used to decide if we need cat arrays
			numCat, err := strconv.Atoi(value)
			if err != nil {
				return tree{}, &ModelError{Detail: fmt.Sprintf("invalid num_cat: %v", err)}
			}
			// We'll use numCat later when we encounter cat_boundaries/cat_threshold
			_ = numCat

		case "split_feature":
			if value != "" {
				tr.splitFeatures, err = parseInts(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid split_feature: %v", err)}
				}
			}

		case "split_gain":
			// Ignored for inference
			continue

		case "threshold":
			if value != "" {
				tr.thresholds, err = parseFloat64s(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid threshold: %v", err)}
				}
			}

		case "decision_type":
			if value != "" {
				tr.decisionTypes, err = parseUint8s(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid decision_type: %v", err)}
				}
			}

		case "left_child":
			if value != "" {
				tr.leftChildren, err = parseInts(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid left_child: %v", err)}
				}
			}

		case "right_child":
			if value != "" {
				tr.rightChildren, err = parseInts(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid right_child: %v", err)}
				}
			}

		case "leaf_value":
			if value != "" {
				tr.leafValues, err = parseFloat64s(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid leaf_value: %v", err)}
				}
			}

		case "leaf_weight", "leaf_count", "internal_value", "internal_weight", "internal_count":
			// Ignored for inference
			continue

		case "shrinkage":
			tr.shrinkage, err = strconv.ParseFloat(value, 64)
			if err != nil {
				return tree{}, &ModelError{Detail: fmt.Sprintf("invalid shrinkage: %v", err)}
			}

		case "cat_boundaries":
			if value != "" {
				tr.catBoundaries, err = parseInts(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid cat_boundaries: %v", err)}
				}
			}

		case "cat_threshold":
			if value != "" {
				tr.catThresholds, err = parseUint32s(value)
				if err != nil {
					return tree{}, &ModelError{Detail: fmt.Sprintf("invalid cat_threshold: %v", err)}
				}
			}

		case "is_linear":
			// Ignored for inference
			continue

		default:
			// Unknown key; ignore for forward compatibility
			continue
		}
	}

	// Validate array lengths
	expectedSplitCount := tr.numLeaves - 1
	if len(tr.splitFeatures) != expectedSplitCount {
		return tree{}, &ModelError{
			Detail: fmt.Sprintf("split_feature count mismatch: got %d, expected %d (num_leaves-1)",
				len(tr.splitFeatures), expectedSplitCount),
		}
	}

	if len(tr.leafValues) != tr.numLeaves {
		return tree{}, &ModelError{
			Detail: fmt.Sprintf("leaf_value count mismatch: got %d, expected %d (num_leaves)",
				len(tr.leafValues), tr.numLeaves),
		}
	}

	return tr, nil
}

// parseInts parses a space-separated string of integers.
func parseInts(s string) ([]int, error) {
	fields := strings.Fields(s)
	result := make([]int, len(fields))
	for i, field := range fields {
		val, err := strconv.Atoi(field)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// parseFloat64s parses a space-separated string of float64 values.
func parseFloat64s(s string) ([]float64, error) {
	fields := strings.Fields(s)
	result := make([]float64, len(fields))
	for i, field := range fields {
		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// parseUint32s parses a space-separated string of uint32 values.
func parseUint32s(s string) ([]uint32, error) {
	fields := strings.Fields(s)
	result := make([]uint32, len(fields))
	for i, field := range fields {
		val, err := strconv.ParseUint(field, 10, 32)
		if err != nil {
			return nil, err
		}
		result[i] = uint32(val)
	}
	return result, nil
}

// parseUint8s parses a space-separated string of uint8 values.
func parseUint8s(s string) ([]uint8, error) {
	fields := strings.Fields(s)
	result := make([]uint8, len(fields))
	for i, field := range fields {
		val, err := strconv.ParseUint(field, 10, 8)
		if err != nil {
			return nil, err
		}
		result[i] = uint8(val)
	}
	return result, nil
}
