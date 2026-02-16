# go-lgbm Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-13

## Active Technologies
- Go 1.21+ + GitHub Actions (CI/CD platform), Go standard toolchain (002-github-releases-packages)
- N/A (no data storage — CI/CD configuration files only) (002-github-releases-packages)
- Go 1.21+ (validation program) + Python 3.x (model generation) + go-lgbm (this library), github.com/dmitryikh/leaves (comparison target), lightgbm + scikit-learn (Python, model generation) (003-leaves-comparison-validation)
- File-based (model files in text format, JSON test data, markdown report) (003-leaves-comparison-validation)

- Go 1.21+ (generics available, minimum supported) + Standard library only (`bufio`, `strconv`, (001-lgbm-model-inference)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.21+ (generics available, minimum supported)

## Code Style

Go 1.21+ (generics available, minimum supported): Follow standard conventions

## Recent Changes
- 003-leaves-comparison-validation: Added Go 1.21+ (validation program) + Python 3.x (model generation) + go-lgbm (this library), github.com/dmitryikh/leaves (comparison target), lightgbm + scikit-learn (Python, model generation)
- 002-github-releases-packages: Added Go 1.21+ + GitHub Actions (CI/CD platform), Go standard toolchain

- 001-lgbm-model-inference: Added Go 1.21+ (generics available, minimum supported) + Standard library only (`bufio`, `strconv`,

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
