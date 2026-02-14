# Research: GitHub Releases and Packages

## R1: GitHub Actions for Go Projects

**Decision**: Use GitHub Actions with the official `actions/setup-go` action and `actions/checkout`.

**Rationale**: GitHub Actions is the native CI/CD platform for GitHub repositories. It requires no external service, no API keys, and is free for public repositories. The `actions/setup-go` action provides caching of Go modules and build artifacts out of the box.

**Alternatives considered**:
- CircleCI, Travis CI: External services requiring separate accounts and configuration. Unnecessary for a public Go library.
- Self-hosted runners: Adds maintenance burden with no benefit for a small open-source library.

## R2: Release Automation Approach

**Decision**: Use GitHub's built-in release creation via the `softprops/action-gh-release` action, triggered by tag pushes matching `v*.*.*`.

**Rationale**: This is the most widely used GitHub Actions release action (14k+ stars). It supports automatic changelog generation from git history, pre-release detection, and is well-maintained.

**Alternatives considered**:
- `actions/create-release` (archived): GitHub's official action is archived and no longer maintained.
- GoReleaser: Full-featured but designed for distributing compiled binaries. Overkill for a source-only Go library that is consumed via `go get`.
- Manual `gh release create` in a script: Works but lacks the declarative configuration and community support of the action.

## R3: Go Module Path and Proxy

**Decision**: Update `go.mod` module path from `github.com/rokt/go-lgbm` to `github.com/zhongdai/go-lgbm` to match the public GitHub repository URL.

**Rationale**: The Go module proxy (proxy.golang.org) discovers modules by their `go.mod` module path. If the path doesn't match the repository URL, `go get` will fail. Since the public repository is `github.com/zhongdai/go-lgbm`, the module path must match.

**Alternatives considered**:
- Keep `github.com/rokt/go-lgbm` and use a vanity import path: Requires running a redirect server, adds complexity. Not justified for a personal project.
- Use `go.mod` replace directives: Only works locally, not for published modules.

## R4: Changelog Generation Strategy

**Decision**: Use GitHub's automatically generated release notes feature (`generate_release_notes: true` in the release action). This creates changelogs from PR titles and commit messages since the previous tag.

**Rationale**: Zero configuration needed. Works well with conventional commit messages. GitHub categorizes changes by label (features, fixes, etc.) automatically.

**Alternatives considered**:
- git-cliff: Powerful but adds a dependency and requires configuration.
- Manual changelog maintenance (CHANGELOG.md): Prone to drift, extra manual work.
- `conventional-changelog` tools: Node.js dependency, unnecessary for a Go project.

## R5: CI Workflow Design

**Decision**: Single CI workflow file (`ci.yml`) triggered on PR and push to main. Steps: checkout, setup Go with caching, run `go vet`, run `go test -race -count=1 ./...`.

**Rationale**: Keeps the CI simple and fast. Race detection catches concurrency bugs. `go vet` catches common mistakes. Module caching speeds up repeated builds.

**Alternatives considered**:
- Matrix builds across multiple Go versions: Could add later, but for now targeting Go 1.21+ is sufficient.
- Adding `staticcheck` or `golangci-lint`: Useful but not in the constitution's required CI checks. Can be added as a follow-up.

## R6: Pre-release Detection

**Decision**: Detect pre-release versions by checking if the tag contains a hyphen (e.g. `v1.0.0-rc.1`). The `softprops/action-gh-release` action supports a `prerelease` flag that can be set conditionally.

**Rationale**: Semantic versioning defines pre-release versions as those with a hyphen suffix. This is a standard convention that requires no custom parsing.
