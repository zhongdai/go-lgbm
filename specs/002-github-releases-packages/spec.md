# Feature Specification: GitHub Releases and Packages

**Feature Branch**: `002-github-releases-packages`
**Created**: 2026-02-13
**Status**: Draft
**Input**: User description: "add support for github releases and packages"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automated Release on Version Tag (Priority: P1)

As a library maintainer, when I push a semantic version tag (e.g. `v1.0.0`) to the repository, the system automatically creates a GitHub Release with a generated changelog so that consumers can discover and use new versions without manual release steps.

**Why this priority**: Without automated releases, every version requires manual GitHub Release creation and changelog assembly. This is the core value — tagging a version should be all that's needed to publish.

**Independent Test**: Push a tag matching `v*.*.*` to the repository and verify a GitHub Release is created with the correct version, a changelog derived from commit history, and the source archive attached.

**Acceptance Scenarios**:

1. **Given** a commit on the main branch, **When** the maintainer pushes a tag `v1.0.0`, **Then** a GitHub Release is created with title "v1.0.0", a changelog body listing commits since the previous tag, and source archives attached automatically.
2. **Given** a tag `v1.0.0` already exists and new commits are merged, **When** the maintainer pushes tag `v1.1.0`, **Then** the changelog for v1.1.0 lists only commits between v1.0.0 and v1.1.0.
3. **Given** a tag that does not match the `v*.*.*` pattern (e.g. `test-build`), **When** it is pushed, **Then** no release is created.

---

### User Story 2 - Go Module Availability via GitHub (Priority: P2)

As a Go developer wanting to use go-lgbm, when a new version is released I can immediately fetch it via `go get` using the repository's GitHub URL, so that I can depend on stable, tagged versions of the library.

**Why this priority**: Go modules rely on tagged versions in the source repository. The Go module proxy discovers versions from git tags. Proper tagging and release hygiene ensures `go get` works seamlessly.

**Independent Test**: After a release is created, run `go get github.com/zhongdai/go-lgbm@v1.0.0` from a fresh project and verify the module is fetched successfully with the correct version.

**Acceptance Scenarios**:

1. **Given** a GitHub Release for `v1.0.0` exists with a proper semantic version tag, **When** a developer runs `go get github.com/zhongdai/go-lgbm@v1.0.0`, **Then** the Go module proxy resolves and downloads the correct version.
2. **Given** the module path in `go.mod` matches the GitHub repository URL, **When** a new tag is pushed, **Then** the Go module proxy indexes the new version within its standard propagation window.

---

### User Story 3 - CI Validation Before Release (Priority: P3)

As a library maintainer, I want all tests and quality checks to run automatically on every pull request and push to main, so that only validated code gets released.

**Why this priority**: Releases should never contain broken code. CI validation gates ensure that tagged versions have passed all quality checks before a release is created.

**Independent Test**: Open a pull request with a failing test and verify the CI check reports failure. Then fix the test and verify CI passes before the PR can be merged.

**Acceptance Scenarios**:

1. **Given** a pull request targeting the main branch, **When** it is opened or updated, **Then** the system runs the full test suite and reports pass/fail status on the PR.
2. **Given** a push to the main branch, **When** it is received, **Then** the system runs the full test suite and code quality checks.
3. **Given** a pull request where CI checks have failed, **When** a reviewer views the PR, **Then** the failing checks are clearly visible and the PR cannot be merged until checks pass.

---

### Edge Cases

- What happens when a tag is pushed that duplicates an existing release version? The system should skip release creation and report a warning rather than fail.
- What happens when the changelog generation finds no commits between the previous tag and the new tag? The release should still be created with a note indicating no changes.
- What happens when the `go.mod` module path does not match the GitHub repository URL? The Go module proxy will not correctly resolve versions — this must be validated in CI.
- What happens when a pre-release tag is pushed (e.g. `v1.0.0-rc.1`)? It should create a GitHub Release marked as a pre-release.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST automatically create a GitHub Release when a semantic version tag (`v*.*.*`) is pushed to the repository.
- **FR-002**: Each GitHub Release MUST include a changelog generated from commit messages between the current and previous version tags.
- **FR-003**: Pre-release tags (e.g. `v1.0.0-rc.1`, `v2.0.0-beta.1`) MUST create GitHub Releases marked as "pre-release".
- **FR-004**: The system MUST run the full test suite with race detection on every pull request targeting the main branch.
- **FR-005**: The system MUST run the full test suite on every push to the main branch.
- **FR-006**: The system MUST validate that the Go module path in `go.mod` is consistent with the repository URL.
- **FR-007**: The system MUST run code quality checks (vet, formatting) on pull requests and pushes to main.
- **FR-008**: Tagged releases MUST only be created from commits that have passed all CI checks on the main branch.
- **FR-009**: The Go module MUST be discoverable and installable via `go get` using the repository's public URL after a version tag is pushed.
- **FR-010**: The system MUST NOT create a release for tags that do not match the semantic version pattern.

### Key Entities

- **Release**: A versioned snapshot of the library associated with a git tag, containing a changelog and source archives. Key attributes: version tag, changelog body, pre-release flag, creation timestamp.
- **CI Pipeline**: An automated validation process triggered by repository events (push, pull request, tag). Key attributes: trigger event, check suite (tests, vet, format), pass/fail status.
- **Go Module Version**: A tagged version of the library resolvable by the Go module proxy. Key attributes: module path, version tag, compatibility with `go get`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of semantic version tags pushed to the repository result in a corresponding GitHub Release within 5 minutes.
- **SC-002**: Every GitHub Release includes a complete changelog covering all commits since the previous version.
- **SC-003**: All pull requests receive automated test and quality check results before merge.
- **SC-004**: Developers can install any released version via `go get` within 15 minutes of release creation.
- **SC-005**: Zero releases are created from code that has not passed the full test suite.
- **SC-006**: Pre-release versions are clearly distinguished from stable releases.

## Scope

### In Scope

- Automated GitHub Release creation on version tags
- Changelog generation from commit history
- CI pipeline for tests and quality checks on PRs and main branch pushes
- Go module version availability via standard `go get`
- Pre-release support

### Out of Scope

- Binary artifact distribution (the library is source-only, consumed via `go get`)
- Container image publishing
- Notifications or announcements beyond the GitHub Release itself
- Branch protection rule configuration (assumed to be managed separately)
- Code signing or provenance attestation

## Assumptions

- The repository is hosted on GitHub at `github.com/zhongdai/go-lgbm` and is publicly accessible.
- The `go.mod` module path will be updated to match the public GitHub URL if it differs.
- Semantic versioning (semver) is the versioning strategy, starting from `v0.1.0` or `v1.0.0`.
- GitHub's built-in release changelog generation (from commit messages and PR titles) is sufficient — no custom changelog tool is required.
- The Go module proxy (proxy.golang.org) will automatically index new tagged versions from the public repository.
- The CI pipeline runs on GitHub-hosted runners with Go pre-installed.

## Dependencies

- A public GitHub repository with push access for the maintainer.
- GitHub Actions or equivalent CI service enabled on the repository.
- The Go module proxy must be able to access the repository (public access required).
