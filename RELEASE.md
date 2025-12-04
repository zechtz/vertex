# Release Guide

## Quick Release Process

1. **Commit all changes**
   ```bash
   git add .
   git commit -m "feat: your changes"
   ```

2. **Create and push version tag**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

3. **Wait for GitHub Actions**
   - Workflow automatically builds binaries for all platforms
   - Creates GitHub Release with binaries and release notes
   - Check: https://github.com/zechtz/vertex/actions

4. **Done!**
   - Users can download from: https://github.com/zechtz/vertex/releases

## What Happens Automatically

When you push a tag (e.g., `v1.0.0`):

1. **GitHub Actions triggers** (`.github/workflows/release.yml`)
2. **Builds binaries** for:
   - Linux (amd64, arm64)
   - macOS (Intel, Apple Silicon)
   - Windows (amd64)
3. **Generates SHA256 checksums**
4. **Creates release notes** from git commits
5. **Publishes GitHub Release** with all artifacts

## Local Testing

Test the build before tagging:

```bash
# Build locally with current git tag
make build
./vertex version

# Build with specific version (override)
make build VERSION=1.0.0-test
./vertex version
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- `v1.0.0` - Major release (breaking changes)
- `v1.1.0` - Minor release (new features, backward compatible)
- `v1.1.1` - Patch release (bug fixes)
- `v1.0.0-beta.1` - Pre-release

## Example Workflow

```bash
# 1. Make changes
vim internal/services/manager.go

# 2. Test locally
make build
./vertex version

# 3. Commit
git add .
git commit -m "feat: add new service management feature"

# 4. Tag and push
git tag -a v1.2.0 -m "Release version 1.2.0 - Add service management"
git push origin main
git push origin v1.2.0

# 5. Monitor GitHub Actions
# Visit: https://github.com/zechtz/vertex/actions

# 6. Release is ready!
# Visit: https://github.com/zechtz/vertex/releases/latest
```

## Troubleshooting

### Build fails in GitHub Actions

- Check the Actions tab for detailed logs
- Common issues:
  - Tests failing
  - Dependencies not available
  - Go version mismatch

### Version not detected

```bash
# Check git tags
git tag -l

# Check current version
git describe --tags

# If no tags exist, create one
git tag -a v0.1.0 -m "Initial release"
```

### Need to re-release

If you need to fix a release:

```bash
# Delete the tag locally and remotely
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Delete the GitHub Release (via web UI)
# Make your fixes, commit, and re-tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

## Manual Release (Without CI/CD)

If you need to create a release manually:

```bash
# Build all platforms
./build.sh

# This creates:
# - vertex-linux-amd64
# - vertex-linux-arm64
# - vertex-darwin-amd64
# - vertex-darwin-arm64
# - vertex-windows-amd64.exe

# Create checksums
sha256sum vertex-* > checksums.txt

# Manually create GitHub Release and upload files
```
