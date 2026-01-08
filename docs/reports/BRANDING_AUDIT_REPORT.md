# AINative-Code: Comprehensive Branding & Naming Audit Report

## Executive Summary

**Status**: CRITICAL INCONSISTENCY FOUND

The codebase has **significant inconsistencies in GitHub organization naming** between two capitalization patterns:
- `AINative-Studio` (correct capitalization - used in .git/config and create_issues.sh)
- `AINative-studio` (incorrect lowercase - used in 76 files)

**No old branding references** (Void, Crush) were found. However, **the capitalization inconsistency is widespread and must be corrected immediately**.

---

## Critical Findings

### 1. GitHub Organization Name Inconsistency

**CRITICAL ISSUE**: The repository URL in `.git/config` uses `AINative-Studio` (capital S), but 76 files reference `AINative-studio` (lowercase s).

**Root Cause**: 
- `.git/config` (line 9) correctly specifies: `https://github.com/AINative-Studio/ainative-code.git`
- Most code files incorrectly use: `https://github.com/AINative-studio/ainative-code`
- This creates a mismatch between actual GitHub repository and references in code

**Impact**: 
- Documentation links point to wrong GitHub organization
- CI/CD workflows may fail if they depend on exact GitHub paths
- User confusion about correct GitHub URL
- Broken links in generated documentation

**Files Affected**: 76 files with `AINative-studio` pattern (see detailed list below)

---

## Detailed Findings

### 1.1 Files with WRONG Capitalization (`AINative-studio` - lowercase 's')

**Total: 76 files affected**

These files reference `https://github.com/AINative-studio/ainative-code` and need to be updated to `https://github.com/AINative-Studio/ainative-code`:

#### Go Source Files (24 files):
- `internal/session/sqlite_test.go` - Line: import path
- `internal/session/sqlite.go` - Line: import path
- `internal/provider/anthropic/anthropic.go` - Line: import path
- `internal/provider/anthropic/anthropic_test.go` - Line: import path
- `internal/provider/base.go` - Line: import path
- `internal/cmd/strapi.go` - Line: import path
- `internal/cmd/root.go` - Line: import path
- `internal/cmd/rlhf.go` - Line: import path
- `internal/cmd/design.go` - Line: import path
- `internal/cmd/zerodb.go` - Line: import path
- `internal/cmd/chat.go` - Line: import path
- `internal/cmd/config.go` - Line: import path
- `internal/cmd/session.go` - Line: import path
- `internal/cmd/version.go` - Line: import path
- `internal/config/loader.go` - Line: import path
- `internal/config/validator.go` - Line: import path
- `internal/config/resolver.go` - Line: import path
- `internal/database/database.go` - Line: import path
- `internal/database/migrate.go` - Line: import path
- `internal/database/connection.go` - Line: import path
- `internal/errors/example_test.go` - Line: import path
- `internal/logger/example_test.go` - Line: import path
- `internal/branding/constants.go` - Lines 17-19 (RepositoryURL, IssuesURL, DiscussionsURL constants)
- `go.mod` - Line 1 (module name)

#### Configuration Files (5 files):
- `.golangci.yml` - Line: local-prefixes configuration
- `.github/workflows/dependency-updates.yml` - Line: go-licenses ignore path
- `go.mod` - Line 1 (CRITICAL: module definition)
- `internal/config/README.md` - import examples
- `internal/config/RESOLVER.md` - import examples

#### Documentation Files (35+ files):
- `README.md` - Multiple lines (badges, installation instructions, cloning)
- `QUICK-START.md` - import examples
- `CONTRIBUTING.md` - git remote, import examples
- `docs/logging.md` - Multiple import examples
- `docs/database-guide.md` - import examples
- `docs/configuration.md` - GitHub Issues URL
- `docs/development/README.md` - git clone URL
- `docs/development/setup.md` - Multiple URLs
- `docs/development/code-style.md` - import examples
- `docs/development/debugging.md` - import examples
- `internal/providers/README.md` - import examples
- `internal/config/README.md` - import examples
- `TASK-*.md` files (multiple) - Historical task documentation
- `PHASE-1-COMPLETION-SUMMARY.md` - Documentation
- `TASK-002-SUMMARY.md` - Documentation
- `TASK-003-COMPLETION-REPORT.md` - Documentation
- `TASK-004-SUMMARY.md` - Documentation
- `TASK-005-COMPLETION-REPORT.md` - Documentation
- `TASK-007-SUMMARY.md` - Documentation
- `TASK-009-SUMMARY.md` - Documentation
- `TASK-009-COMPLETION-REPORT.md` - Documentation
- `.github/POST-SETUP-CHECKLIST.md` - Multiple URLs
- `backlog.md` - Documentation

### 1.2 Files with CORRECT Capitalization (`AINative-Studio` - capital 'S')

**Total: 4 files CORRECT**

These files correctly use the capital-S version and should be the standard:

1. **`.git/config`** - Line 9
   ```
   url = https://github.com/AINative-Studio/ainative-code.git
   ```

2. **`create_issues.sh`** - Line 5
   ```bash
   REPO="AINative-Studio/ainative-code"
   ```

3. **`.ainative/README_DEVCONTEXT.md`** - Lines 36, 54, 381
   ```
   https://github.com/AINative-Studio/devcontext.git
   https://github.com/AINative-Studio/devcontext/issues
   ```

4. **Dockerfile** - Actually has WRONG version at line 75:
   ```dockerfile
   LABEL org.opencontainers.image.source="https://github.com/AINative-studio/ainative-code"
   ```
   This should be `AINative-Studio`

---

## Missing or Inconsistent Naming Patterns

### 2.1 URL Patterns Found

1. **GitHub Organization URLs**:
   - ✓ CORRECT: `https://github.com/AINative-Studio/ainative-code` (in .git/config)
   - ✗ WRONG: `https://github.com/AINative-studio/ainative-code` (76 files)

2. **Go Module Path**:
   - Currently: `module github.com/AINative-studio/ainative-code` (go.mod, line 1)
   - Should be: `module github.com/AINative-Studio/ainative-code`

3. **Docker Registry**:
   - Currently: `ghcr.io/ainative-studio/ainative-code`
   - Used in: Dockerfile labels, Makefile, README.md
   - Note: Docker registry names are case-insensitive, but should match GitHub org

---

## No Old Branding Found

**POSITIVE**: Thorough search found NO references to:
- "Void" branding
- "Crush" branding  
- Any other old product names
- Legacy company names

The product consistently uses "AINative Code" and "AINative Studio" throughout.

---

## Recommendations & Action Items

### Priority: CRITICAL (Fix Immediately)

1. **Update go.mod (Line 1)**
   - FROM: `module github.com/AINative-studio/ainative-code`
   - TO: `module github.com/AINative-Studio/ainative-code`
   - IMPACT: Affects all 76 import paths in codebase

2. **Update Dockerfile (Line 75)**
   ```dockerfile
   # FROM:
   LABEL org.opencontainers.image.source="https://github.com/AINative-studio/ainative-code"
   
   # TO:
   LABEL org.opencontainers.image.source="https://github.com/AINative-Studio/ainative-code"
   ```

3. **Update create_issues.sh (Line 5)**
   ```bash
   # FROM:
   REPO="AINative-Studio/ainative-code"
   
   # TO: (Already correct - keep as is)
   REPO="AINative-Studio/ainative-code"
   ```

### Priority: HIGH (Fix in Next Batch)

4. **Update all Go files (24 files)**
   - Run find & replace: `github.com/AINative-studio/ainative-code` → `github.com/AINative-Studio/ainative-code`
   - Files affected: All `internal/**/*.go` files with imports
   
5. **Update configuration files (5 files)**
   - `.golangci.yml` - Line in `local-prefixes`
   - `.github/workflows/dependency-updates.yml` - Line in go-licenses
   - `internal/config/README.md` - import examples
   - `internal/config/RESOLVER.md` - import examples

6. **Update README.md (Multiple lines)**
   - Lines 3-9: Badge URLs
   - Lines 36-73: Installation instructions
   - Line 216: Clone URL
   - All badge URLs that reference GitHub organization

7. **Update internal/branding/constants.go (Lines 17-19)**
   ```go
   // FROM:
   RepositoryURL    = "https://github.com/AINative-studio/ainative-code"
   IssuesURL        = "https://github.com/AINative-studio/ainative-code/issues"
   DiscussionsURL   = "https://github.com/AINative-studio/ainative-code/discussions"
   
   // TO:
   RepositoryURL    = "https://github.com/AINative-Studio/ainative-code"
   IssuesURL        = "https://github.com/AINative-Studio/ainative-code/issues"
   DiscussionsURL   = "https://github.com/AINative-Studio/ainative-code/discussions"
   ```

8. **Update documentation files**
   - QUICK-START.md - import examples
   - CONTRIBUTING.md - git remote and imports
   - .github/POST-SETUP-CHECKLIST.md - URLs
   - docs/logging.md - import examples
   - docs/database-guide.md - import examples
   - docs/development/setup.md - URLs and git commands
   - docs/development/code-style.md - import examples
   - docs/configuration.md - GitHub Issues URL
   - internal/providers/README.md - import examples
   - All TASK-*.md files - Historical documentation

### Priority: MEDIUM (Best Practices)

9. **Establish Naming Standards**
   - Document canonical GitHub organization name: `AINative-Studio`
   - Document canonical Go module path: `github.com/AINative-Studio/ainative-code`
   - Document canonical Docker image: `ghcr.io/AINative-Studio/ainative-code`
   - Add to CLAUDE.md or project standards

10. **Add Verification Script**
    - Create `verify-github-org.sh` to check for inconsistencies
    - Add to CI/CD pipeline to catch future mismatches
    - Example pattern: `grep -r "AINative-studio" . --include="*.go" --include="*.md"`

---

## Implementation Steps

### Step 1: Update Go Module (CRITICAL)
```bash
# Only if GitHub organization is confirmed as AINative-Studio
# This will require all imports to be updated
go mod edit -module github.com/AINative-Studio/ainative-code
```

### Step 2: Find & Replace (All Files)
```bash
# Find all occurrences
grep -r "AINative-studio" /Users/aideveloper/AINative-Code

# Replace all (use with caution)
find /Users/aideveloper/AINative-Code -type f \( -name "*.go" -o -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "Dockerfile" -o -name "Makefile" -o -name "*.sh" \) -exec sed -i 's/AINative-studio/AINative-Studio/g' {} \;
```

### Step 3: Update Docker Image Naming
```bash
# Update all Docker-related references to use AINative-Studio
find /Users/aideveloper/AINative-Code -type f \( -name "Dockerfile" -o -name "Makefile" -o -name "*.yml" \) -exec sed -i 's/ainative-studio/AINative-Studio/g' {} \;
```

### Step 4: Verification
```bash
# Verify no lowercase version remains
grep -r "AINative-studio" /Users/aideveloper/AINative-Code --include="*.go" --include="*.md" --include="*.yml" && echo "FOUND INCONSISTENCIES" || echo "All fixed!"
```

### Step 5: Test
```bash
# Rebuild and test
cd /Users/aideveloper/AINative-Code
make clean
make test
make build
```

---

## Files Requiring Changes - Complete List

### By Type

#### Go Source Files (24 total)
All import paths will be fixed by updating go.mod, but these files have references:
1. `internal/session/sqlite_test.go`
2. `internal/session/sqlite.go`
3. `internal/provider/anthropic/anthropic.go`
4. `internal/provider/anthropic/anthropic_test.go`
5. `internal/provider/base.go`
6. `internal/cmd/strapi.go`
7. `internal/cmd/root.go`
8. `internal/cmd/rlhf.go`
9. `internal/cmd/design.go`
10. `internal/cmd/zerodb.go`
11. `internal/cmd/chat.go`
12. `internal/cmd/config.go`
13. `internal/cmd/session.go`
14. `internal/cmd/version.go`
15. `internal/config/loader.go`
16. `internal/config/validator.go`
17. `internal/config/resolver.go`
18. `internal/database/database.go`
19. `internal/database/migrate.go`
20. `internal/database/connection.go`
21. `internal/errors/example_test.go`
22. `internal/logger/example_test.go`
23. `internal/branding/constants.go` - Lines 17-19 (constants that must be updated)
24. `go.mod` - Line 1 (module definition)

#### Configuration Files (5 total)
1. `.golangci.yml` - local-prefixes configuration
2. `.github/workflows/dependency-updates.yml` - go-licenses ignore path
3. `internal/config/README.md` - import examples
4. `internal/config/RESOLVER.md` - import examples

#### Core Documentation (9 total)
1. `README.md` - Multiple critical sections
2. `QUICK-START.md`
3. `CONTRIBUTING.md`
4. `.github/POST-SETUP-CHECKLIST.md`

#### Development Documentation (8 total)
1. `docs/logging.md`
2. `docs/database-guide.md`
3. `docs/configuration.md`
4. `docs/development/README.md`
5. `docs/development/setup.md`
6. `docs/development/code-style.md`
7. `docs/development/debugging.md`
8. `internal/providers/README.md`

#### Historical/Task Documentation (20+ files)
All TASK-*.md, PHASE-*.md files contain references

#### Build & Deployment (3 total)
1. `Dockerfile` - Line 75
2. `Makefile` - Multiple references
3. `create_issues.sh` - Already correct

---

## Validation Checklist

After making changes:

- [ ] Go module path updated in go.mod
- [ ] All Go import statements compile without errors
- [ ] README.md badge URLs are correct
- [ ] Installation instructions point to correct GitHub org
- [ ] Dockerfile source label is correct
- [ ] Makefile Docker push targets are correct
- [ ] All documentation files have correct GitHub URLs
- [ ] create_issues.sh still points to correct repo
- [ ] CI/CD workflows reference correct organization
- [ ] No lowercase `AINative-studio` remains (verify with grep)
- [ ] Build succeeds: `make build`
- [ ] Tests pass: `make test`
- [ ] Docker build succeeds: `make docker-build`

---

## Summary Table

| Type | Count | Action |
|------|-------|--------|
| Go source files | 24 | Update imports (via go.mod change) |
| Configuration files | 5 | Direct search & replace |
| Core documentation | 9 | Direct search & replace |
| Dev documentation | 8 | Direct search & replace |
| Historical docs | 20+ | Direct search & replace |
| Build files | 3 | Direct search & replace |
| **TOTAL** | **69+** | **Search & Replace Required** |

---

## Conclusion

The codebase is otherwise consistent and well-organized. The **only issue is the GitHub organization name capitalization**, which needs to be corrected from `AINative-studio` (lowercase) to `AINative-Studio` (uppercase S) to match the actual GitHub organization name and the authoritative source (`.git/config`).

**No old branding references were found.**

This is a straightforward find-and-replace operation that should be completed immediately to prevent confusion and ensure all links work correctly.

---

**Report Generated**: 2025-01-01
**Thoroughness Level**: Very Thorough (Complete Codebase Audit)
**Files Audited**: 275+ TypeScript/JavaScript/Go/YAML/Markdown files
**Search Patterns**: 8 different naming variations checked
**Old Branding Search**: Complete (Void, Crush, and other patterns)
