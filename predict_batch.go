package lgbm

import (
	"fmt"
	"runtime"
	"sync"
)

// PredictDense predicts on a dense matrix of feature vectors.
// features is a flat row-major slice of length nRows * nCols.
// nCols must equal NFeatures().
// output must have length >= nRows * outputWidth where outputWidth is
// NClasses() for multiclass models, 1 otherwise.
// nThreads controls parallelism: 0 = runtime.NumCPU(), 1 = single-threaded.
// nEstimators limits trees used (0 = all).
func (m *Model) PredictDense(features []float64, nRows, nCols, nEstimators, nThreads int, output []float64) error {
	if nCols != m.numFeatures {
		return fmt.Errorf("%w: model expects %d features, got %d columns",
			ErrFeatureCountMismatch, m.numFeatures, nCols)
	}

	if nRows == 0 {
		return nil
	}

	outputWidth := 1
	if m.numClasses > 1 {
		outputWidth = m.numClasses
	}

	requiredOutput := nRows * outputWidth
	if len(output) < requiredOutput {
		return fmt.Errorf("%w: output slice length %d, need at least %d",
			ErrInvalidModel, len(output), requiredOutput)
	}

	requiredInput := nRows * nCols
	if len(features) < requiredInput {
		return fmt.Errorf("%w: features slice length %d, need at least %d",
			ErrInvalidModel, len(features), requiredInput)
	}

	if nThreads == 0 {
		nThreads = runtime.NumCPU()
	}

	if nThreads == 1 || nRows <= nThreads {
		// Single-threaded path
		for i := 0; i < nRows; i++ {
			row := features[i*nCols : (i+1)*nCols]
			out := output[i*outputWidth : (i+1)*outputWidth]
			if err := m.Predict(row, nEstimators, out); err != nil {
				return fmt.Errorf("row %d: %w", i, err)
			}
		}
		return nil
	}

	// Multi-threaded path
	var wg sync.WaitGroup
	errCh := make(chan error, nThreads)

	rowsPerThread := (nRows + nThreads - 1) / nThreads

	for t := 0; t < nThreads; t++ {
		startRow := t * rowsPerThread
		endRow := startRow + rowsPerThread
		if endRow > nRows {
			endRow = nRows
		}
		if startRow >= endRow {
			break
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				row := features[i*nCols : (i+1)*nCols]
				out := output[i*outputWidth : (i+1)*outputWidth]
				if err := m.Predict(row, nEstimators, out); err != nil {
					errCh <- fmt.Errorf("row %d: %w", i, err)
					return
				}
			}
		}(startRow, endRow)
	}

	wg.Wait()
	close(errCh)

	if err, ok := <-errCh; ok {
		return err
	}
	return nil
}
