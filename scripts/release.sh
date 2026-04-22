#!/bin/bash
# Release script for kmgr
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v0.1.0

set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

# Remove 'v' prefix if present for some operations
VERSION_NO_V="${VERSION#v}"

echo "🚀 Preparing release $VERSION"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo "❌ Must be on main branch to release"
    exit 1
fi

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    echo "❌ Working directory is not clean"
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^${VERSION}$"; then
    echo "❌ Tag $VERSION already exists"
    exit 1
fi

echo "✅ Pre-release checks passed"

# Update version in CHANGELOG.md (move from Unreleased to version)
if [[ -f "CHANGELOG.md" ]]; then
    sed -i.bak "s/## \[Unreleased\]/## [${VERSION_NO_V}] - $(date +%Y-%m-%d)/" CHANGELOG.md
    rm CHANGELOG.md.bak
    echo "✅ Updated CHANGELOG.md"
fi

# Commit version bump if any changes
if [[ -n $(git status --porcelain) ]]; then
    git add CHANGELOG.md
    git commit -m "chore: prepare release $VERSION"
    echo "✅ Committed version changes"
fi

# Create and push tag
git tag -a "$VERSION" -m "Release $VERSION"
git push origin main
git push origin "$VERSION"

echo "✅ Tag $VERSION created and pushed"

# Build release binaries
make release-build

echo "✅ Release binaries built in bin/"
ls -lh bin/

echo ""
echo "🎉 Release $VERSION prepared!"
echo ""
echo "Next steps:"
echo "1. Wait for GitHub Actions to complete the release"
echo "2. Test the binaries from the GitHub release"
echo "3. Update Homebrew formula if needed"
echo "4. Announce the release"