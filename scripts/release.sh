#!/bin/bash
# Release automation script for VTEX Files Manager
# Usage: ./scripts/release.sh v1.2.3

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Version argument required${NC}"
    echo "Usage: ./scripts/release.sh v1.2.3"
    echo ""
    echo "Examples:"
    echo "  ./scripts/release.sh v1.0.0   # Major release"
    echo "  ./scripts/release.sh v1.1.0   # Minor release (new features)"
    echo "  ./scripts/release.sh v1.0.1   # Patch release (bug fixes)"
    exit 1
fi

VERSION=$1

echo -e "${GREEN}=== VTEX Files Manager Release Process ===${NC}"
echo ""
echo "Version: $VERSION"
echo ""

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Version must follow format vX.Y.Z (e.g., v1.2.3)${NC}"
    exit 1
fi

# Check if on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${YELLOW}Warning: Not on main branch (current: $CURRENT_BRANCH)${NC}"
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}Error: Working directory is not clean${NC}"
    echo "Commit or stash your changes first"
    git status --short
    exit 1
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag $VERSION already exists${NC}"
    echo "To delete and recreate:"
    echo "  git tag -d $VERSION"
    echo "  git push origin :refs/tags/$VERSION"
    exit 1
fi

# Pull latest changes
echo -e "${YELLOW}Pulling latest changes...${NC}"
git pull origin "$CURRENT_BRANCH"

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
if ! go test ./...; then
    echo -e "${RED}Tests failed. Fix errors before releasing.${NC}"
    exit 1
fi

# Build binaries to verify
echo -e "${YELLOW}Building binaries...${NC}"
if ! go build -o vtex-files-manager .; then
    echo -e "${RED}Build failed. Fix errors before releasing.${NC}"
    exit 1
fi

if ! go build -o vfm .; then
    echo -e "${RED}Build failed. Fix errors before releasing.${NC}"
    exit 1
fi

# Clean up built binaries
rm -f vtex-files-manager vfm

# Check GoReleaser config (if goreleaser is installed)
if command -v goreleaser &> /dev/null; then
    echo -e "${YELLOW}Validating GoReleaser config...${NC}"
    if ! goreleaser check; then
        echo -e "${RED}GoReleaser config is invalid. Fix errors before releasing.${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}Note: goreleaser not installed, skipping config validation${NC}"
fi

# Confirm release
echo ""
echo -e "${GREEN}All checks passed!${NC}"
echo ""
echo "About to create and push tag: $VERSION"
echo "This will trigger an automated release build on GitHub Actions."
echo ""
read -p "Proceed with release? [y/N] " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled"
    exit 1
fi

# Create annotated tag
echo -e "${YELLOW}Creating tag $VERSION...${NC}"
git tag -a "$VERSION" -m "Release $VERSION"

# Push tag
echo -e "${YELLOW}Pushing tag to GitHub...${NC}"
git push origin "$VERSION"

# Success
echo ""
echo -e "${GREEN}✅ Release $VERSION created successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Check GitHub Actions: https://github.com/glinharesb/vtex-files-manager/actions"
echo "  2. Monitor release build progress"
echo "  3. Verify release: https://github.com/glinharesb/vtex-files-manager/releases/tag/$VERSION"
echo ""
echo "The release will be available at:"
echo "  https://github.com/glinharesb/vtex-files-manager/releases/latest"
echo ""
