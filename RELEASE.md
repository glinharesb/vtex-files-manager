# Release Guide

This document describes how to create a new release of VTEX Files Manager.

## Prerequisites

1. Have push permissions on the repository
2. Git configured locally
3. GoReleaser installed (optional, only for local testing)

## Release Process

### 1. Prepare the Release

Make sure all changes are committed and the `main` branch is up to date:

```bash
git checkout main
git pull origin main
```

### 2. Run Tests Locally

```bash
# Run tests
go test ./...

# Check build
go build -o vtex-files-manager .
go build -o vfm .

# Validate GoReleaser configuration (optional)
goreleaser check
```

### 3. Update the CHANGELOG (Optional)

If you maintain a `CHANGELOG.md` file, update it with the new version changes.

### 4. Create and Publish the Tag

```bash
# Define the version (following Semantic Versioning)
# MAJOR.MINOR.PATCH
# - MAJOR: incompatible API changes
# - MINOR: new features (compatible)
# - PATCH: bug fixes

VERSION="v1.0.0"  # Adjust as needed

# Create annotated tag
git tag -a $VERSION -m "Release $VERSION"

# Verify the tag
git tag -l $VERSION

# Publish the tag to GitHub
git push origin $VERSION
```

### 5. Wait for GitHub Actions

After pushing the tag, GitHub Actions will automatically:

1. **Run tests** on multiple platforms
2. **Compile binaries** for:
   - Linux (amd64, arm64)
   - macOS (amd64 Intel, arm64 Apple Silicon)
   - Windows (amd64)
3. **Generate checksums** SHA256
4. **Create the release** on GitHub with:
   - Automatic release notes
   - Attached binaries
   - Compressed files (.tar.gz, .zip)

Monitor progress at: `https://github.com/glinharesb/vtex-files-manager/actions`

### 6. Verify the Release

1. Go to: `https://github.com/glinharesb/vtex-files-manager/releases/latest`
2. Check if all files are present:
   - `vtex-files-manager_VERSION_Linux_x86_64.tar.gz`
   - `vtex-files-manager_VERSION_Linux_arm64.tar.gz`
   - `vtex-files-manager_VERSION_Darwin_x86_64.tar.gz`
   - `vtex-files-manager_VERSION_Darwin_arm64.tar.gz`
   - `vtex-files-manager_VERSION_Windows_x86_64.zip`
   - `checksums.txt`
3. Test one of the downloaded binaries

### 7. Edit Release Notes (Optional)

If necessary, edit the automatically generated release notes to add:
- Important highlights
- Breaking changes
- Migration instructions
- Contributor acknowledgments

## Complete Example

```bash
# 1. Update main
git checkout main
git pull origin main

# 2. Run tests
go test ./...

# 3. Create and publish tag
git tag -a v1.2.3 -m "Release v1.2.3

New Features:
- Add support for new file formats
- Improve batch upload performance

Bug Fixes:
- Fix timeout error on large files
- Fix special character encoding
"

git push origin v1.2.3

# 4. Wait for GitHub Actions to complete
# Visit: https://github.com/glinharesb/vtex-files-manager/actions

# 5. Verify release
# Visit: https://github.com/glinharesb/vtex-files-manager/releases/latest
```

## Semantic Versioning

Follow the [Semantic Versioning](https://semver.org/) standard:

- **v1.0.0** → Initial release
- **v1.0.1** → Bug fixes (PATCH)
- **v1.1.0** → New compatible feature (MINOR)
- **v2.0.0** → Breaking changes (MAJOR)

### Examples

| Change | Version |
|--------|---------|
| Bug fix | `v1.0.0` → `v1.0.1` |
| New optional flag | `v1.0.0` → `v1.1.0` |
| Rename command | `v1.0.0` → `v2.0.0` |
| Add new command | `v1.0.0` → `v1.1.0` |
| Remove command | `v1.0.0` → `v2.0.0` |

## Test Release Locally

To test the release process locally without publishing:

```bash
# Install GoReleaser
brew install goreleaser  # macOS
# or
go install github.com/goreleaser/goreleaser@latest

# Run local build (snapshot)
goreleaser release --snapshot --clean

# Binaries will be in ./dist/
ls -lh dist/
```

## Revert a Release

If something goes wrong:

```bash
# Delete tag locally
git tag -d v1.2.3

# Delete tag remotely
git push origin :refs/tags/v1.2.3

# Delete the release on GitHub
# Visit: https://github.com/glinharesb/vtex-files-manager/releases
# Click "Delete" on the unwanted release
```

## Troubleshooting

### GoReleaser fails in CI

**Problem**: Error during build in GitHub Actions

**Solution**:
1. Check logs: `https://github.com/glinharesb/vtex-files-manager/actions`
2. Test locally: `goreleaser release --snapshot --clean`
3. Validate config: `goreleaser check`

### Tag already exists

**Problem**: `error: tag 'v1.0.0' already exists`

**Solution**:
```bash
# Delete existing tag
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Create new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Corrupted binaries

**Problem**: Downloaded binaries don't work

**Solution**:
1. Verify checksums: `sha256sum -c checksums.txt`
2. Redo the release by deleting the tag and creating it again

## Additional Automation

To further automate the process, you can create a script:

```bash
#!/bin/bash
# scripts/release.sh

set -e

if [ $# -eq 0 ]; then
    echo "Usage: ./scripts/release.sh v1.2.3"
    exit 1
fi

VERSION=$1

echo "=== Creating release $VERSION ==="

# Validate version
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must follow format vX.Y.Z"
    exit 1
fi

# Run tests
echo "Running tests..."
go test ./...

# Create tag
echo "Creating tag $VERSION..."
git tag -a $VERSION -m "Release $VERSION"

# Push tag
echo "Pushing tag to GitHub..."
git push origin $VERSION

echo "✅ Release $VERSION created!"
echo "Check progress at: https://github.com/glinharesb/vtex-files-manager/actions"
echo "Release will be available at: https://github.com/glinharesb/vtex-files-manager/releases/tag/$VERSION"
```

Usage:
```bash
chmod +x scripts/release.sh
./scripts/release.sh v1.2.3
```
