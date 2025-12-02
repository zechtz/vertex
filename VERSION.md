# Vertex Versioning

## Automatic Versioning from Git Tags

Vertex automatically detects and uses git tags as version numbers. This integrates seamlessly with CI/CD workflows.

### Quick Start

```bash
# Development build (automatically uses git tag or "dev")
make build

# Build and show version
make version

# Manual version override (if needed)
make build VERSION=1.0.0
```

### How Version Detection Works

1. **On a tagged commit**: Uses the git tag (e.g., `v1.0.0` → version `1.0.0`)
2. **Between tags**: Uses `git describe` output (e.g., `v1.0.0-5-g1234abc`)
3. **No tags**: Falls back to `dev`

The version is automatically stripped of the `v` prefix if present.

### Using the Build Script

The build script also auto-detects git tags:

```bash
# Auto-detect version from git tags
./build.sh

# Manual override if needed
VERSION=1.0.0 ./build.sh
```

The build script creates cross-platform binaries.

## Checking Version

Once built, check the version with:

```bash
./vertex version
# or
./vertex --version
```

Example output:
```
Vertex 0.1.0
Commit: fc8869d
Built: 2025-12-02T04:59:13Z
```

## Available Make Targets

- `make help` - Show all available targets
- `make build` - Build with version info (defaults to "dev")
- `make build-release VERSION=x.x.x` - Build a release version
- `make version` - Build and display version
- `make install` - Build and install vertex
- `make clean` - Remove build artifacts
- `make dev` - Alias for `make build`

## Version Information

The version information is embedded at build time and includes:

- **Version**: Semantic version (e.g., 1.0.0) or "dev" for development builds
- **Commit**: Short git commit hash (e.g., fc8869d)
- **Built**: Build timestamp in UTC (e.g., 2025-12-02T04:59:13Z)

## Automated CI/CD Release Process

Vertex uses GitHub Actions to automatically build and release binaries when you push a version tag.

### Creating a Release

1. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

2. **Create and push a version tag**:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions automatically**:
   - Detects the tag push
   - Builds binaries for all platforms:
     - Linux (amd64, arm64)
     - macOS (amd64, arm64)
     - Windows (amd64)
   - Generates release notes from commits
   - Creates checksums
   - Publishes a GitHub Release with all binaries

4. **Users can download** the pre-built binaries from the GitHub Releases page

### Version Tag Format

Use semantic versioning with a `v` prefix:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.1.1` - Patch release (bug fixes)
- `v2.0.0-beta.1` - Pre-release

### Manual Release (Local Build)

If you need to build locally without CI/CD:

```bash
# Tag your commit
git tag -a v1.0.0 -m "Release version 1.0.0"

# Build (automatically uses the tag)
make build

# Or use build script for cross-platform binaries
./build.sh

# Verify version
./vertex version
```

### Release Checklist

- [ ] All tests passing
- [ ] Version number follows semantic versioning
- [ ] CHANGELOG updated (if you maintain one)
- [ ] Create annotated git tag: `git tag -a vX.Y.Z -m "Release version X.Y.Z"`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Verify GitHub Actions workflow completes successfully
- [ ] Check GitHub Releases page for published release
