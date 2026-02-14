package lgbm

import "fmt"

// PredictSingle returns a single prediction for models with one output
// class (binary classification, regression, ranking). Returns 0 and an
// error for multiclass models — use Predict instead.
//
// features must have length equal to NFeatures().
// nEstimators limits the number of trees used (0 = all trees).
func (m *Model) PredictSingle(features []float64, nEstimators int) (float64, error) {
	if err := m.validateFeatures(features); err != nil {
		return 0, err
	}

	if m.numClasses > 1 {
		return 0, ErrMulticlassNotSupported
	}

	raw := m.predictRaw(features, nEstimators)

	out := make([]float64, 1)
	m.transform(raw, out)
	return out[0], nil
}

// Predict writes prediction(s) into the provided output slice.
// For single-class models, output must have length >= 1.
// For multiclass models, output must have length >= NClasses().
//
// features must have length equal to NFeatures().
// nEstimators limits the number of trees used (0 = all trees).
func (m *Model) Predict(features []float64, nEstimators int, output []float64) error {
	if err := m.validateFeatures(features); err != nil {
		return err
	}

	required := m.numClasses
	if required == 1 {
		required = 1
	}
	if len(output) < required {
		return fmt.Errorf("%w: output slice length %d, need at least %d",
			ErrInvalidModel, len(output), required)
	}

	raw := m.predictRaw(features, nEstimators)
	m.transform(raw, output)
	return nil
}

// WithRawPredictions returns a new Model that bypasses the output
// transformation, returning raw tree scores instead. The returned
// Model shares tree data with the original (no deep copy).
func (m *Model) WithRawPredictions() *Model {
	return &Model{
		version:              m.version,
		numClasses:           m.numClasses,
		numTreesPerIteration: m.numTreesPerIteration,
		numFeatures:          m.numFeatures,
		objective:            m.objective,
		averageOutput:        m.averageOutput,
		trees:                m.trees, // shared, not copied
		featureNames:         m.featureNames,
		transform:            transformIdentity,
	}
}

// predictRaw accumulates raw tree scores across the ensemble.
// Returns a slice of length numTreesPerIteration (1 for single-class,
// numClasses for multiclass).
func (m *Model) predictRaw(features []float64, nEstimators int) []float64 {
	nGroups := m.numTreesPerIteration
	raw := make([]float64, nGroups)

	maxTrees := len(m.trees)
	if nEstimators > 0 {
		limit := nEstimators * nGroups
		if limit < maxTrees {
			maxTrees = limit
		}
	}

	for i := 0; i < maxTrees; i++ {
		classIdx := i % nGroups
		raw[classIdx] += m.trees[i].predictLeaf(features)
	}

	if m.averageOutput && maxTrees > 0 {
		iterations := float64(maxTrees / nGroups)
		for i := range raw {
			raw[i] /= iterations
		}
	}

	return raw
}

// validateFeatures checks that the feature vector has the correct length.
func (m *Model) validateFeatures(features []float64) error {
	if len(features) != m.numFeatures {
		return fmt.Errorf("%w: model expects %d features, got %d",
			ErrFeatureCountMismatch, m.numFeatures, len(features))
	}
	return nil
}
