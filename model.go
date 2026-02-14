package lgbm

// Model represents a loaded LightGBM model with its ensemble of trees
// and associated metadata. It provides methods for prediction and
// inspection of model properties.
type Model struct {
	// version is the LightGBM model format version (e.g. "v3", "v4").
	version string

	// numClasses is the number of output classes.
	// 1 for binary/regression, >1 for multiclass.
	numClasses int

	// numTreesPerIteration is the number of trees trained per boosting iteration.
	// Equals 1 for binary/regression, equals numClasses for multiclass.
	numTreesPerIteration int

	// numFeatures is the expected number of input features.
	numFeatures int

	// objective identifies the training objective (binary, regression, multiclass, etc.).
	objective ObjectiveType

	// averageOutput indicates whether to average tree outputs (true for ranking).
	averageOutput bool

	// trees is the ensemble of decision trees.
	trees []tree

	// featureNames stores the names of input features, if available.
	featureNames []string

	// transform is the post-prediction transformation function
	// (e.g. sigmoid for binary, softmax for multiclass).
	transform TransformFunc
}

// NFeatures returns the number of input features expected by the model.
func (m *Model) NFeatures() int {
	return m.numFeatures
}

// NClasses returns the number of output classes.
// Returns 1 for binary classification and regression, >1 for multiclass.
func (m *Model) NClasses() int {
	return m.numClasses
}

// NTrees returns the total number of trees in the ensemble.
func (m *Model) NTrees() int {
	return len(m.trees)
}

// FeatureNames returns a copy of the feature names slice.
// Returns nil if feature names were not present in the model file.
func (m *Model) FeatureNames() []string {
	if len(m.featureNames) == 0 {
		return nil
	}
	names := make([]string, len(m.featureNames))
	copy(names, m.featureNames)
	return names
}
