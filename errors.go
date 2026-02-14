// Package lgbm provides pure-Go LightGBM model loading and inference.
package lgbm

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by model loading and prediction functions.
var (
	// ErrUnsupportedVersion indicates the model file uses a LightGBM
	// version that this library does not support (only v3 and v4).
	ErrUnsupportedVersion = errors.New("lgbm: unsupported LightGBM version")

	// ErrInvalidModel indicates the model file is malformed, truncated,
	// or missing required fields.
	ErrInvalidModel = errors.New("lgbm: invalid model")

	// ErrFeatureCountMismatch indicates the feature vector length does
	// not match the model's expected feature count.
	ErrFeatureCountMismatch = errors.New("lgbm: feature count mismatch")

	// ErrMulticlassNotSupported indicates PredictSingle was called on
	// a multiclass model (use Predict instead).
	ErrMulticlassNotSupported = errors.New("lgbm: PredictSingle not supported for multiclass models, use Predict")
)

// VersionError wraps ErrUnsupportedVersion with the detected version string.
type VersionError struct {
	Version string
}

func (e *VersionError) Error() string {
	return fmt.Sprintf("%v: %q (only v3 and v4 are supported)", ErrUnsupportedVersion, e.Version)
}

func (e *VersionError) Unwrap() error {
	return ErrUnsupportedVersion
}

// ModelError wraps ErrInvalidModel with a descriptive message.
type ModelError struct {
	Detail string
}

func (e *ModelError) Error() string {
	return fmt.Sprintf("%v: %s", ErrInvalidModel, e.Detail)
}

func (e *ModelError) Unwrap() error {
	return ErrInvalidModel
}
