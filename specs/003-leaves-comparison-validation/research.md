# Research: Leaves Comparison Validation

## Decision 1: Separate Go Module for Validation

**Decision**: Use a separate `go.mod` in `validation/` rather than adding leaves to the main module.

**Rationale**: If leaves were added to the main `go.mod`, every user who `go get`s go-lgbm would also pull in the leaves dependency tree. A separate module keeps the validation tool self-contained and the main library dependency-free.

**Alternatives considered**:
- Add leaves to main go.mod as a test dependency → Still appears in go.sum, confuses users
- Use a build tag to isolate → Complex, easy to forget the tag
- **Separate go.mod** → Clean isolation, standard Go multi-module pattern ✓

## Decision 2: Python Model Generation Strategy

**Decision**: Use a single Python script (`generate_models.py`) that trains 4 model types (binary, multiclass, regression, ranking) using scikit-learn synthetic datasets and exports them in LightGBM text format.

**Rationale**: Synthetic data ensures reproducibility (fixed random seed) and avoids shipping large datasets. LightGBM's Python API can generate all 4 model types with minimal code.

**Alternatives considered**:
- Ship pre-trained models in the repo → Large files, can't regenerate with different params
- Use real datasets (e.g., UCI) → Download dependency, larger scope
- **Synthetic data with fixed seed** → Reproducible, small, fast ✓

## Decision 3: Test Input Generation

**Decision**: The Python script also generates random test inputs (1,000 per model) as JSON files with a fixed seed. The Go program reads these JSON files to ensure both libraries see exactly the same inputs.

**Rationale**: Generating inputs in Python and saving to JSON guarantees bit-exact input data for both Go programs. Generating in Go would risk subtle floating-point differences between Python and Go random number generators.

**Alternatives considered**:
- Generate inputs in Go → Risk of different float representations
- Hardcode a small set of inputs → Insufficient coverage
- **Generate in Python, save as JSON** → Guarantees identical inputs ✓

## Decision 4: Leaves API Compatibility

**Decision**: Use `github.com/dmitryikh/leaves` with its `LGEnsembleFromReader` or `LGEnsembleFromFile` API to load models, and `PredictSingle` for predictions, mirroring how go-lgbm is used.

**Rationale**: The leaves library uses `LGEnsemble` as its model type. Both libraries load the same text format. Comparing `PredictSingle` output is the most direct validation.

**Alternatives considered**:
- Compare batch predictions only → Misses single-prediction path
- Compare raw scores only → Misses transform (sigmoid/softmax) differences
- **Compare both transformed predictions** → Most comprehensive ✓

## Decision 5: Report Format

**Decision**: Generate a markdown report (`REPORT.md`) with a summary table per model type showing: model name, test count, max absolute diff, mean absolute diff, and pass/fail. Include metadata (timestamp, library versions, tolerance).

**Rationale**: Markdown renders natively on GitHub, making the report immediately readable when linked from README. Per-model-type tables make it easy to identify which model types match.

**Alternatives considered**:
- JSON report → Not human-readable on GitHub
- HTML report → Doesn't render on GitHub
- **Markdown** → Native GitHub rendering ✓

## Decision 6: Justfile Integration

**Decision**: Create a `validation/justfile` with recipes: `generate-models` (runs Python), `validate` (runs Go comparison), and `all` (runs both in sequence). The root `justfile` gets a `validate` recipe that delegates to `validation/justfile`.

**Rationale**: Separating the justfile into the validation directory keeps concerns isolated. The root justfile recipe provides a convenient single entry point.

**Alternatives considered**:
- Put all recipes in root justfile → Mixes validation and library concerns
- Use a Makefile → Less readable, project already uses justfile
- **Validation justfile + root recipe** → Clean separation ✓
