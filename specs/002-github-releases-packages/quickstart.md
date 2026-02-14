# Quickstart: GitHub Releases and Packages

## For Library Consumers

### Install a specific version

```bash
go get github.com/zhongdai/go-lgbm@v1.0.0
```

### Install the latest version

```bash
go get github.com/zhongdai/go-lgbm@latest
```

### Use in your project

```go
import lgbm "github.com/zhongdai/go-lgbm"

model, err := lgbm.ModelFromFile("model.txt", true)
```

## For Library Maintainers

### Creating a release

1. Ensure all changes are merged to `main` and CI is green.
2. Tag the release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

3. The release workflow automatically:
   - Creates a GitHub Release with changelog
   - Marks pre-releases (if tag contains `-`, e.g. `v1.0.0-rc.1`)

### Creating a pre-release

```bash
git tag v1.1.0-rc.1
git push origin v1.1.0-rc.1
```

This creates a GitHub Release marked as "pre-release".

### Verifying the release

After pushing a tag:
1. Check the [Releases page](https://github.com/zhongdai/go-lgbm/releases) for the new release
2. Verify the changelog lists the expected commits
3. Test `go get` from a separate project:

```bash
mkdir /tmp/test-lgbm && cd /tmp/test-lgbm
go mod init test
go get github.com/zhongdai/go-lgbm@v1.0.0
```

### CI checks

Every pull request and push to `main` automatically runs:
- `go vet ./...` — static analysis
- `go test -race -count=1 ./...` — full test suite with race detection

Check the "Actions" tab on GitHub for CI status.
