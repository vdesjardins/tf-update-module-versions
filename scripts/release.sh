#!/bin/sh
# Release script for terraform-module-versions
# Updates version everywhere and creates a git tag

set -e

NEW_VERSION=$1

# Validate input
if [ -z "$NEW_VERSION" ]; then
    echo "Usage: scripts/release.sh VERSION"
    echo "Example: scripts/release.sh 0.2.0"
    echo ""
    echo "Current version: $(cat .version 2>/dev/null || echo 'unknown')"
    exit 1
fi

# Validate version format (semver: X.Y.Z)
if ! echo "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo "ERROR: Version must be in semver format (X.Y.Z)"
    echo "Got: $NEW_VERSION"
    exit 1
fi

# Check if version already exists as git tag
if git rev-parse "v$NEW_VERSION" >/dev/null 2>&1; then
    echo "ERROR: Tag v$NEW_VERSION already exists"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD -- 2>/dev/null; then
    echo "ERROR: Uncommitted changes detected. Commit or stash them first."
    exit 1
fi

echo "Preparing release of v$NEW_VERSION..."
echo ""

# 1. Update .version file
echo "$NEW_VERSION" > .version
echo "✓ Updated .version to $NEW_VERSION"

# 2. Update flake.nix version
sed -i.bak "s/version = \"[^\"]*\";/version = \"$NEW_VERSION\";/" flake.nix
rm -f flake.nix.bak
echo "✓ Updated flake.nix to $NEW_VERSION"

# 3. Verify versions
VERIFY_VERSION=$(cat .version 2>/dev/null)
if [ "$VERIFY_VERSION" != "$NEW_VERSION" ]; then
    echo "ERROR: Failed to update .version file"
    exit 1
fi

# 4. Run verification
echo ""
make verify-versions
echo ""

# 5. Run tests
echo "Running tests before release..."
make test >/dev/null 2>&1 || {
    echo "ERROR: Tests failed. Aborting release."
    git checkout .version flake.nix 2>/dev/null || true
    exit 1
}
echo "✓ All tests passed"
echo ""

# 6. Commit changes
echo "Committing changes..."
git add .version flake.nix
git commit -m "chore(release): bump version to v$NEW_VERSION"
echo "✓ Changes committed"

# 7. Create tag
echo "Creating git tag v$NEW_VERSION..."
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
echo "✓ Git tag created"

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║ Release v$NEW_VERSION prepared successfully!         "
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "To push the release:"
echo "  git push origin main"
echo "  git push origin v$NEW_VERSION"
echo ""
