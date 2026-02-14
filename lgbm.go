package lgbm

import (
	"bufio"
)

// ModelFromFile loads a LightGBM text-format model from the given file.
// If loadTransformation is true, the appropriate output transformation
// (sigmoid, softmax, etc.) is derived from the model's objective.
// If false, raw tree scores are returned.
func ModelFromFile(filename string, loadTransformation bool) (*Model, error) {
	return modelFromFile(filename, loadTransformation)
}

// ModelFromReader loads a LightGBM text-format model from a buffered reader.
// If loadTransformation is true, the appropriate output transformation
// (sigmoid, softmax, etc.) is derived from the model's objective.
// If false, raw tree scores are returned.
func ModelFromReader(reader *bufio.Reader, loadTransformation bool) (*Model, error) {
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
