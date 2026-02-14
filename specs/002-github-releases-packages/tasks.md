# Tasks: GitHub Releases and Packages

**Input**: Design documents from `/specs/002-github-releases-packages/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, quickstart.md

**Tests**: Not applicable — this feature is CI/CD infrastructure (YAML workflow files), not library code.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Workflows**: `.github/workflows/` at repository root
- **Module config**: `go.mod` at repository root

---

## Phase 1: Setup

**Purpose**: Create the `.github/workflows/` directory structure

- [x] T001 Create `.github/workflows/` directory at repository root

**Checkpoint**: Directory structure ready for workflow files.

---

## Phase 2: User Story 1 - Automated Release on Version Tag (Priority: P1) MVP

**Goal**: When a semantic version tag (`v*.*.*`) is pushed, automatically create a GitHub Release with a generated changelog.

**Independent Test**: Push a tag `v0.1.0` to the repository and verify a GitHub Release appears with changelog and source archives.

### Implementation for User Story 1

- [x] T002 [US1] Create release workflow in `.github/workflows/release.yml` — trigger on push of tags matching `v*.*.*`, use `softprops/action-gh-release` action with `generate_release_notes: true`, set `prerelease` flag when tag contains a hyphen (e.g. `v1.0.0-rc.1`)
- [x] T003 [US1] Verify release workflow handles edge cases — ensure workflow does not trigger on non-semver tags, and that duplicate tag pushes do not cause failures
- [ ] T004 [US1] Push initial version tag `v0.1.0` to test the release workflow end-to-end and verify GitHub Release is created with changelog

**Checkpoint**: Pushing a version tag creates a GitHub Release automatically.

---

## Phase 3: User Story 2 - Go Module Availability (Priority: P2)

**Goal**: Ensure the Go module is discoverable and installable via `go get` using the public GitHub repository URL.

**Independent Test**: Run `go get github.com/zhongdai/go-lgbm@v0.1.0` from a fresh Go project and verify it resolves correctly.

### Implementation for User Story 2

- [x] T005 [US2] Update module path in `go.mod` from `github.com/rokt/go-lgbm` to `github.com/zhongdai/go-lgbm`
- [x] T006 [US2] Update all import references in `*_test.go` and source files if any use the full module path
- [x] T007 [US2] Run `go test -race ./...` to verify module path change does not break existing tests
- [ ] T008 [US2] Verify Go module proxy indexes the version by running `go get github.com/zhongdai/go-lgbm@v0.1.0` from a temporary project after the v0.1.0 release exists

**Checkpoint**: The module is installable via `go get` with the public GitHub URL.

---

## Phase 4: User Story 3 - CI Validation Before Release (Priority: P3)

**Goal**: Run tests and quality checks on every PR and push to main, so only validated code gets released.

**Independent Test**: Open a PR with a deliberate test failure and verify CI reports failure on the PR.

### Implementation for User Story 3

- [x] T009 [US3] Create CI workflow in `.github/workflows/ci.yml` — trigger on pull_request targeting main and push to main, steps: checkout, setup Go 1.21+ with module caching, run `go vet ./...`, run `go test -race -count=1 ./...`
- [x] T010 [US3] Add module path validation step to CI workflow — verify `go.mod` module path matches `github.com/zhongdai/go-lgbm`
- [ ] T011 [US3] Push the CI workflow to main and verify it runs on a test PR

**Checkpoint**: Every PR and push to main runs the full test suite and quality checks.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [x] T012 [P] Update README.md to add CI status badge from `.github/workflows/ci.yml`
- [x] T013 [P] Update README.md installation section to use `github.com/zhongdai/go-lgbm` module path
- [x] T014 Verify all workflows are syntactically valid by pushing to a branch and checking GitHub Actions tab
- [ ] T015 Run quickstart.md validation — verify `go get github.com/zhongdai/go-lgbm@v0.1.0` works after the first release

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (US1 - Release)**: Depends on Phase 1
- **Phase 3 (US2 - Go Module)**: Depends on Phase 2 (needs a released version tag to test `go get`)
- **Phase 4 (US3 - CI)**: Depends on Phase 1 only (can run parallel to US1)
- **Phase 5 (Polish)**: Depends on Phases 2-4

### User Story Dependencies

- **US1 (Release)**: Independent — only needs the `.github/workflows/` directory
- **US2 (Go Module)**: Depends on US1 — needs a version tag/release to verify `go get`
- **US3 (CI)**: Independent of US1 and US2 — can be implemented in parallel with US1

### Parallel Opportunities

- T002 (release workflow) and T009 (CI workflow) can be created in parallel — different files
- T005 (go.mod update) can start as soon as T001 is done — independent of workflow files
- T012 and T013 (README updates) can run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create directory structure (T001)
2. Complete Phase 2: Release workflow (T002-T004)
3. **STOP and VALIDATE**: Push `v0.1.0` tag and verify GitHub Release is created
4. This is a shippable MVP — maintainer can create releases by tagging

### Incremental Delivery

1. Setup → Directory ready
2. US1 (Release workflow) → Tags create GitHub Releases (MVP!)
3. US2 (Go Module) → Module path aligned, `go get` works
4. US3 (CI) → PRs and pushes validated automatically
5. Polish → Badges, docs, final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- This feature involves NO Go library code changes — only workflow YAML files, `go.mod` path, and README updates
- The `go.mod` module path change (T005) will require updating the README import paths as well (T013)
- The first tag push (T004) is both a test and the initial public release
- Commit after each task or logical group
