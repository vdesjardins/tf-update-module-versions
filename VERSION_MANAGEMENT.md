# Version Management Strategy Guide

## Current State Analysis

### Where Versions Appear

1. **go.mod** - Go version requirement
   ```
   go 1.25.4
   ```

2. **flake.nix** (package definition) - Package version
   ```nix
   version = "0.1.0";  # Line 20
   ```

3. **flake.nix** (dev shell) - Go compiler version
   ```nix
   go_1_25  # Line 49
   ```

4. **Makefile** - Dynamic version from git tags
   ```makefile
   VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
   ```

5. **main.go** - Version variables (injected at build time)
   ```go
   var Version = "dev"  # Injected via ldflags
   ```

---

## Issue 1: Go Version Sync (go.mod vs flake.nix)

### Current Setup (GOOD)
```
go.mod:     go 1.25.4
flake.nix:  go_1_25    ← Major version only
```

### Is It Important? **SOMEWHAT**
- **Yes**: Keep go.mod and Go compiler in sync
- **Go patch versions** (1.25.0 → 1.25.4) are compatible within same minor
- **Minor version mismatch** (go 1.25 in mod, go 1.24 in flake) can cause issues
- **Major version mismatch** (go 2.x in mod, go 1.x in flake) WILL break builds

### Best Practices

#### Option 1: Manual (Current - Simple)
Keep go.mod major.minor in sync with flake.nix `go_X_Y`:
- Check when upgrading: `grep "^go" go.mod` vs `grep "go_" flake.nix`
- Works fine for stable versions
- Low maintenance

#### Option 2: Automated Check (Recommended)
Add to Makefile to verify sync:

```makefile
# In Makefile
check-go-version:
	@GO_MOD_VERSION=$$(grep "^go " go.mod | cut -d' ' -f2 | cut -d. -f1-2) && \
	FLAKE_VERSION=$$(grep "go_" flake.nix | sed 's/.*go_//' | tr '_' '.' | cut -d' ' -f1 | cut -d. -f1-2) && \
	if [ "$$GO_MOD_VERSION" != "$$FLAKE_VERSION" ]; then \
		echo "ERROR: Go version mismatch!"; \
		echo "  go.mod:   $$GO_MOD_VERSION"; \
		echo "  flake.nix: $$FLAKE_VERSION"; \
		exit 1; \
	fi; \
	echo "✓ Go versions in sync"

build: check-go-version $(BIN_DIR)
```

#### Option 3: Derivation Version (Advanced - Not Recommended)
Extract from go.mod in flake.nix:
```nix
# This is complex and fragile - avoid
goVersion = builtins.head (builtins.match "go ([0-9.]+).*" (builtins.readFile ./go.mod));
```

### Recommendation for This Project
**Keep it simple**: Manually verify when upgrading Go
```bash
# When you upgrade Go, run this check:
grep "^go" go.mod
grep "go_" flake.nix
# Ensure major.minor versions match
```

---

## Issue 2: Package Version Management (Harder Problem)

### Current State
```
flake.nix:    version = "0.1.0";         (hardcoded)
Makefile:     VERSION ?= git describe    (dynamic)
main.go:      var Version = "dev"        (injected)
```

### The Problem
When you release v0.2.0, you need to change:
1. Git tag: `git tag v0.2.0`
2. flake.nix: `version = "0.2.0"`
3. (Makefile/main.go automatically pick up from git tag)

**Easy to forget flake.nix update.**

### Solution: Use Git Tags as Single Source of Truth

#### Option 1: Extract Version from Git Tag in flake.nix (Recommended)

Update flake.nix:
```nix
{
  description = "Terraform Module Version Management Tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      # Extract version from git tag or use dev
      version = 
        if (self.sourceInfo ? rev) then
          let
            describe = builtins.tryEval (builtins.readFile "${self}/.version");
          in
            if describe.success then describe.value
            else "0.1.0-dev"
        else
          "0.1.0-dev";
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages.default = pkgs.buildGoModule {
          pname = "terraform-module-versions";
          inherit version;  # Use extracted version
          # ... rest of config ...
        };
      }
    );
}
```

#### Option 2: Keep Version in File (Simplest)

Create `.version` file:
```
0.1.0
```

Update flake.nix:
```nix
version = builtins.readFile ./.version;
```

Update Makefile:
```makefile
VERSION ?= $(shell cat .version 2>/dev/null || git describe --tags 2>/dev/null || echo "dev")
```

This way, all files read from `.version`:
- flake.nix
- Makefile
- CI/CD pipelines
- Release scripts

#### Option 3: Release Script (Most Reliable)

Create `scripts/release.sh`:
```bash
#!/bin/bash
set -e

NEW_VERSION=$1

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: scripts/release.sh VERSION"
    echo "Example: scripts/release.sh 0.2.0"
    exit 1
fi

echo "Releasing version $NEW_VERSION..."

# 1. Update .version file
echo "$NEW_VERSION" > .version

# 2. Update flake.nix if using hardcoded version
sed -i.bak "s/version = \"[^\"]*\"/version = \"$NEW_VERSION\"/" flake.nix

# 3. Commit and tag
git add .version flake.nix go.mod go.sum
git commit -m "chore: release v$NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"

# 4. Show what to do next
echo ""
echo "✓ Version updated to $NEW_VERSION"
echo "✓ Git tag created"
echo ""
echo "To push the release:"
echo "  git push origin main"
echo "  git push origin v$NEW_VERSION"
```

Usage:
```bash
chmod +x scripts/release.sh
./scripts/release.sh 0.2.0
```

---

## Recommended Implementation for This Project

### Step 1: Create `.version` File (Single Source of Truth)
```bash
echo "0.1.0" > .version
```

### Step 2: Update flake.nix to Read It
```nix
{
  description = "Terraform Module Version Management Tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      version = builtins.readFile ./.version;
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages.default = pkgs.buildGoModule {
          pname = "terraform-module-versions";
          version = builtins.replaceStrings ["\n"] [""] version;
          # ... rest of config ...
        };
      }
    );
}
```

### Step 3: Update Makefile to Read It
```makefile
# Get version from .version file, fallback to git, then "dev"
VERSION ?= $(shell cat .version 2>/dev/null || git describe --tags 2>/dev/null || echo "dev")
```

### Step 4: Create Release Script
See Option 3 above.

### Step 5: Add to .gitignore
```
# Version tracking
.version
```

---

## Summary: What to Do Now

### For Go Version (go.mod ↔ flake.nix)

**Current state: ACCEPTABLE**
- go.mod: 1.25.4
- flake.nix: go_1_25

**Action**: Nothing required now, but when upgrading:
1. Update go.mod: `go mod tidy` (when you use newer features)
2. Check flake.nix matches major.minor version
3. Consider adding make target to verify

### For Package Version Management (RECOMMENDED TO FIX)

**Current state: ERROR-PRONE**
- flake.nix hardcoded: 0.1.0
- Makefile dynamic: git describe
- Conflict between sources

**Action**: Implement Option 2 (simplest)
1. Create `.version` file with "0.1.0"
2. Update flake.nix to read it: `version = builtins.readFile ./.version;`
3. Update Makefile to read it
4. When releasing: Edit `.version`, commit, tag

**Or implement Option 3** (most professional)
1. Create release script
2. Script handles all version updates automatically
3. You just run: `./scripts/release.sh 0.2.0`
4. Script updates everything and creates git tag

---

## Checking Current Sync

```bash
# Check go versions
echo "go.mod version:"
grep "^go" go.mod

echo "flake.nix go compiler:"
grep "go_" flake.nix | grep buildInputs -A 10

# Check package versions
echo "flake.nix package version:"
grep "version =" flake.nix

echo "Makefile VERSION:"
grep "VERSION ?=" Makefile
```

---

## Going Forward: Recommended Changes

1. **Immediate** (Optional): Add make target to verify Go versions
2. **Soon** (Recommended): Implement `.version` file approach
3. **When first release**: Use release script for v0.2.0+

This ensures:
- ✅ Single source of truth (`.version`)
- ✅ No manual updates missed
- ✅ Consistent across all tools
- ✅ Automated release process
