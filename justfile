# go-lgbm project recipes

# Run all tests with race detector
test:
    go test -race -count=1 ./...

# Run tests with verbose output
test-verbose:
    go test -race -count=1 -v ./...

# Run tests with coverage report
coverage:
    go test -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Run go vet and format check
lint:
    go vet ./...
    @UNFORMATTED=$$(gofmt -l .); \
    if [ -n "$$UNFORMATTED" ]; then \
        echo "Unformatted files:"; \
        echo "$$UNFORMATTED"; \
        exit 1; \
    fi
    @echo "All checks passed"

# Format all Go files
fmt:
    gofmt -w .

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Show current version (latest git tag)
version:
    @git tag --sort=-v:refname | head -1

# Bump patch version (e.g. v0.1.0 -> v0.1.1) and push tag
patch:
    #!/usr/bin/env bash
    set -euo pipefail
    latest=$(git tag --sort=-v:refname | head -1)
    if [ -z "$latest" ]; then
        echo "No existing tags found"
        exit 1
    fi
    # Strip leading 'v'
    ver=${latest#v}
    IFS='.' read -r major minor patch <<< "$ver"
    new="v${major}.${minor}.$((patch + 1))"
    echo "Bumping ${latest} -> ${new}"
    git tag "$new"
    git push origin "$new"
    echo "Tagged and pushed ${new}"

# Bump minor version (e.g. v0.1.0 -> v0.2.0) and push tag
minor:
    #!/usr/bin/env bash
    set -euo pipefail
    latest=$(git tag --sort=-v:refname | head -1)
    if [ -z "$latest" ]; then
        echo "No existing tags found"
        exit 1
    fi
    ver=${latest#v}
    IFS='.' read -r major minor patch <<< "$ver"
    new="v${major}.$((minor + 1)).0"
    echo "Bumping ${latest} -> ${new}"
    git tag "$new"
    git push origin "$new"
    echo "Tagged and pushed ${new}"

# Bump major version (e.g. v0.1.0 -> v1.0.0) and push tag
major:
    #!/usr/bin/env bash
    set -euo pipefail
    latest=$(git tag --sort=-v:refname | head -1)
    if [ -z "$latest" ]; then
        echo "No existing tags found"
        exit 1
    fi
    ver=${latest#v}
    IFS='.' read -r major minor patch <<< "$ver"
    new="v$((major + 1)).0.0"
    echo "Bumping ${latest} -> ${new}"
    git tag "$new"
    git push origin "$new"
    echo "Tagged and pushed ${new}"

# Run leaves comparison validation (generate models + compare predictions)
validate:
    just --justfile validation/justfile all

# Run all checks (lint + test) — useful before tagging
check: lint test
