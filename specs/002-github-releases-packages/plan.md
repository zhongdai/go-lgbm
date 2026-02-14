# Implementation Plan: GitHub Releases and Packages

**Branch**: `002-github-releases-packages` | **Date**: 2026-02-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-github-releases-packages/spec.md`

## Summary

Set up automated CI/CD for the go-lgbm library: a GitHub Actions CI workflow that runs tests and quality checks on PRs and pushes to main, a release workflow that creates GitHub Releases with changelogs when version tags are pushed, and alignment of the Go module path to the public GitHub repository URL for `go get` compatibility.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: GitHub Actions (CI/CD platform), Go standard toolchain
**Storage**: N/A (no data storage — CI/CD configuration files only)
**Testing**: `go test -race ./...`, `go vet ./...`
**Target Platform**: GitHub-hosted Ubuntu runners
**Project Type**: Single Go library at repository root
**Performance Goals**: CI pipeline completes within 5 minutes; releases created within 2 minutes of tag push
**Constraints**: No external CI services — GitHub Actions only; no third-party release tools
**Scale/Scope**: Single repository, single maintainer workflow

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Leaves Compatibility | PASS | No changes to library code or API |
| II. Pure Go / No CGo | PASS | CI workflows use standard Go toolchain only |
| III. LightGBM 3+4 Only | PASS | No changes to model support |
| IV. Test-First | PASS | CI enforces test-first by running `go test -race` on every PR |
| V. Idiomatic Go API | PASS | Go module path alignment follows Go conventions |

**Development Workflow compliance**:
- Conventional commits: Release changelog generated from conventional commit messages
- CI checks: `go vet`, `go test -race ./...` enforced on every PR and push to main
- Benchmarks: CI does not run benchmarks automatically (optional future addition)

All gates pass. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/002-github-releases-packages/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
.github/
└── workflows/
    ├── ci.yml           # Test + quality checks on PR and push to main
    └── release.yml      # Create GitHub Release on version tag push
go.mod                   # Module path updated to github.com/zhongdai/go-lgbm
```

**Structure Decision**: GitHub Actions workflows live in `.github/workflows/` per GitHub convention. The only source code change is updating the module path in `go.mod`. No new Go source files are created.
