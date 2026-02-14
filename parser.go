package lgbm

import (
	"bufio"
	"os"
	"strings"
)

// parseModel reads a LightGBM text-format model from a buffered reader.
// It parses the header, reads all trees, and constructs a Model.
func parseModel(reader *bufio.Reader) (*Model, error) {
	scanner := bufio.NewScanner(reader)

	// Parse header section
	h, err := parseHeader(scanner)
	if err != nil {
		return nil, err
	}

	// Parse trees
	var trees []tree
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for end of trees section
		if strings.HasPrefix(line, "end of trees") ||
			strings.HasPrefix(line, "feature_names") ||
			strings.HasPrefix(line, "feature_importances") ||
			strings.HasPrefix(line, "feature importances") ||
			strings.HasPrefix(line, "parameters") {
			break
		}

		// Parse tree if line starts with "Tree="
		if strings.HasPrefix(line, "Tree=") {
			tr, err := parseTree(scanner)
			if err != nil {
				return nil, err
			}
			trees = append(trees, tr)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, &ModelError{Detail: "failed to read model: " + err.Error()}
	}

	// Validate model has trees
	if len(trees) == 0 {
		return nil, &ModelError{Detail: "model has no trees"}
	}

	// Validate trees count is multiple of numTreePerIteration
	if len(trees)%h.numTreePerIteration != 0 {
		return nil, &ModelError{
			Detail: "tree count not a multiple of num_tree_per_iteration",
		}
	}

	// Determine objective type
	objective, err := parseObjective(h.objective)
	if err != nil {
		return nil, err
	}

	// Determine transformation function
	transform := transformForObjective(objective)

	// Calculate number of features (max_feature_idx + 1)
	numFeatures := h.maxFeatureIdx + 1

	model := &Model{
		version:              h.version,
		numClasses:           h.numClass,
		numTreesPerIteration: h.numTreePerIteration,
		numFeatures:          numFeatures,
		objective:            objective,
		averageOutput:        h.averageOutput,
		trees:                trees,
		featureNames:         h.featureNames,
		transform:            transform,
	}

	return model, nil
}

// modelFromFile loads a LightGBM text-format model from a file.
// If loadTransformation is false, the raw tree scores are returned
// without any transformation.
func modelFromFile(filename string, loadTransformation bool) (*Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	model, err := parseModel(reader)
	if err != nil {
		return nil, err
	}

	// Override transform if loadTransformation is false
	if !loadTransformation {
		model.transform = transformIdentity
	}

	return model, nil
}
