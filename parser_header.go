package lgbm

import (
	"bufio"
	"strconv"
	"strings"
)

// header holds parsed metadata from the model file header.
type header struct {
	version             string
	numClass            int
	numTreePerIteration int
	labelIndex          int
	maxFeatureIdx       int
	objective           string
	averageOutput       bool
	featureNames        []string
	treeSizes           []int
}

// parseHeader reads and parses the header section of a LightGBM model file.
// The header consists of a "tree" magic line followed by key=value pairs,
// terminated by a blank line.
func parseHeader(scanner *bufio.Scanner) (header, error) {
	h := header{}

	// Read the magic "tree" identifier line.
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return h, &ModelError{Detail: "failed to read magic line: " + err.Error()}
		}
		return h, &ModelError{Detail: "empty model file"}
	}

	line := strings.TrimSpace(scanner.Text())
	if line != "tree" {
		return h, &ModelError{Detail: "expected 'tree' magic line, got: " + line}
	}

	// Parse key=value pairs until blank line.
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			// Handle keys without values (e.g., "average_output").
			switch line {
			case "average_output":
				h.averageOutput = true
			}
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "version":
			h.version = value
		case "num_class":
			val, err := strconv.Atoi(value)
			if err != nil {
				return h, &ModelError{Detail: "invalid num_class: " + err.Error()}
			}
			h.numClass = val
		case "num_tree_per_iteration":
			val, err := strconv.Atoi(value)
			if err != nil {
				return h, &ModelError{Detail: "invalid num_tree_per_iteration: " + err.Error()}
			}
			h.numTreePerIteration = val
		case "label_index":
			val, err := strconv.Atoi(value)
			if err != nil {
				return h, &ModelError{Detail: "invalid label_index: " + err.Error()}
			}
			h.labelIndex = val
		case "max_feature_idx":
			val, err := strconv.Atoi(value)
			if err != nil {
				return h, &ModelError{Detail: "invalid max_feature_idx: " + err.Error()}
			}
			h.maxFeatureIdx = val
		case "objective":
			h.objective = value
		case "feature_names":
			h.featureNames = strings.Fields(value)
		case "tree_sizes":
			sizes := strings.Fields(value)
			h.treeSizes = make([]int, 0, len(sizes))
			for _, s := range sizes {
				val, err := strconv.Atoi(s)
				if err != nil {
					return h, &ModelError{Detail: "invalid tree_sizes: " + err.Error()}
				}
				h.treeSizes = append(h.treeSizes, val)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return h, &ModelError{Detail: "failed to read header: " + err.Error()}
	}

	// Validate required fields.
	if h.version == "" {
		return h, &ModelError{Detail: "missing required field: version"}
	}
	if h.numClass == 0 {
		return h, &ModelError{Detail: "missing required field: num_class"}
	}
	if h.maxFeatureIdx == 0 {
		return h, &ModelError{Detail: "missing required field: max_feature_idx"}
	}

	// Validate version (must be v3 or v4).
	if !strings.HasPrefix(h.version, "v3") && !strings.HasPrefix(h.version, "v4") {
		return h, &VersionError{Version: h.version}
	}

	// Default num_tree_per_iteration to num_class if not present.
	if h.numTreePerIteration == 0 {
		h.numTreePerIteration = h.numClass
	}

	return h, nil
}
