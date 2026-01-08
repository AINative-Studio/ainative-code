# AINative-Code: Detailed Changes Required by File

## Critical Files - Line-by-Line Changes Needed

### 1. go.mod (CRITICAL - BLOCKING)
**File**: `/Users/aideveloper/AINative-Code/go.mod`
**Current Line 1**:
```
module github.com/AINative-studio/ainative-code
```
**Required Change**:
```
module github.com/AINative-Studio/ainative-code
```

---

### 2. internal/branding/constants.go (CRITICAL)
**File**: `/Users/aideveloper/AINative-Code/internal/branding/constants.go`

**Current Lines 17-19**:
```go
RepositoryURL    = "https://github.com/AINative-studio/ainative-code"
IssuesURL        = "https://github.com/AINative-studio/ainative-code/issues"
DiscussionsURL   = "https://github.com/AINative-studio/ainative-code/discussions"
```

**Required Changes**:
```go
RepositoryURL    = "https://github.com/AINative-Studio/ainative-code"
IssuesURL        = "https://github.com/AINative-Studio/ainative-code/issues"
DiscussionsURL   = "https://github.com/AINative-Studio/ainative-code/discussions"
```

---

### 3. Dockerfile
**File**: `/Users/aideveloper/AINative-Code/Dockerfile`

**Current Line 75**:
```dockerfile
LABEL org.opencontainers.image.source="https://github.com/AINative-studio/ainative-code"
```

**Required Change**:
```dockerfile
LABEL org.opencontainers.image.source="https://github.com/AINative-Studio/ainative-code"
```

---

### 4. README.md
**File**: `/Users/aideveloper/AINative-Code/README.md`

**Lines to Update**:
- Line 3: Badge URL - change `AINative-studio` to `AINative-Studio`
- Line 4: Badge URL - change `AINative-studio` to `AINative-Studio`
- Line 5: Badge URL - change `AINative-studio` to `AINative-Studio`
- Line 6: Badge URL - change `AINative-studio` to `AINative-Studio`
- Line 8: Badge URL - change `AINative-studio` to `AINative-Studio`
- Line 9: Badge URL - change `AINative-studio` to `AINative-Studio`
- Line 36: Download URL - change `AINative-studio` to `AINative-Studio`
- Line 41: Download URL - change `AINative-studio` to `AINative-Studio`
- Line 50: Download URL - change `AINative-studio` to `AINative-Studio`
- Line 55: Download URL - change `AINative-studio` to `AINative-Studio`
- Line 64: Download URL - change `AINative-studio` to `AINative-Studio`
- Line 216: git clone - change `AINative-studio` to `AINative-Studio`

---

### 5. .golangci.yml
**File**: `/Users/aideveloper/AINative-Code/.golangci.yml`

**Change**:
Search for: `github.com/AINative-studio/ainative-code`
Replace with: `github.com/AINative-Studio/ainative-code`

---

### 6. .github/workflows/dependency-updates.yml
**File**: `/Users/aideveloper/AINative-Code/.github/workflows/dependency-updates.yml`

**Change**:
Search for: `github.com/AINative-studio/ainative-code`
Replace with: `github.com/AINative-Studio/ainative-code`

---

### 7. Documentation Files - Bulk Updates

**Files to update with search & replace**:
- `QUICK-START.md`
- `CONTRIBUTING.md`
- `.github/POST-SETUP-CHECKLIST.md`
- `docs/logging.md`
- `docs/database-guide.md`
- `docs/configuration.md`
- `docs/development/README.md`
- `docs/development/setup.md`
- `docs/development/code-style.md`
- `docs/development/debugging.md`
- `internal/config/README.md`
- `internal/config/RESOLVER.md`
- `internal/providers/README.md`
- All `TASK-*.md` files
- `PHASE-1-COMPLETION-SUMMARY.md`
- `backlog.md`

**Search & Replace Pattern**:
```
FROM: github.com/AINative-studio/ainative-code
TO:   github.com/AINative-Studio/ainative-code
```

---

### 8. Go Source Files (All import paths affected)

**All files under `internal/` will auto-update imports when go.mod is updated**:

When you change go.mod line 1, the following files will need their imports updated:
- `internal/session/sqlite_test.go`
- `internal/session/sqlite.go`
- `internal/provider/anthropic/anthropic.go`
- `internal/provider/anthropic/anthropic_test.go`
- `internal/provider/base.go`
- `internal/cmd/strapi.go`
- `internal/cmd/root.go`
- `internal/cmd/rlhf.go`
- `internal/cmd/design.go`
- `internal/cmd/zerodb.go`
- `internal/cmd/chat.go`
- `internal/cmd/config.go`
- `internal/cmd/session.go`
- `internal/cmd/version.go`
- `internal/config/loader.go`
- `internal/config/validator.go`
- `internal/config/resolver.go`
- `internal/database/database.go`
- `internal/database/migrate.go`
- `internal/database/connection.go`
- `internal/errors/example_test.go`
- `internal/logger/example_test.go`

**These can be updated via**:
```bash
go mod edit -module github.com/AINative-Studio/ainative-code
go mod tidy
```

---

## Recommended Update Order

### Phase 1: Critical Files (Must fix first)
1. `go.mod` - Change module name
2. `internal/branding/constants.go` - Update URL constants
3. Run `go mod edit -module github.com/AINative-Studio/ainative-code && go mod tidy`

### Phase 2: Configuration & Build Files
1. `Dockerfile`
2. `.golangci.yml`
3. `.github/workflows/dependency-updates.yml`

### Phase 3: Documentation
1. `README.md`
2. All `.md` files in `docs/` and `internal/`
3. Task and phase documentation files

### Phase 4: Verification
1. Run tests: `make test`
2. Build: `make build`
3. Verify no `AINative-studio` remains: `grep -r "AINative-studio" . --include="*.go" --include="*.md"`

---

## Automated Fix Command

**BACKUP FIRST**, then run:

```bash
cd /Users/aideveloper/AINative-Code

# Step 1: Update module name
go mod edit -module github.com/AINative-Studio/ainative-code
go mod tidy

# Step 2: Replace in all files
find . -type f \( -name "*.go" -o -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "Dockerfile" -o -name "Makefile" -o -name "*.sh" \) -not -path "./.git/*" -exec sed -i 's/AINative-studio/AINative-Studio/g' {} \;

# Step 3: Verify
grep -r "AINative-studio" . --include="*.go" --include="*.md" 2>/dev/null | wc -l
# Should return 0

# Step 4: Test
make clean
make test
make build
```

---

## Summary Statistics

| Category | Count | Status |
|----------|-------|--------|
| Files needing updates | 76+ | Search & Replace |
| Critical files | 3 | High Priority |
| Documentation files | 40+ | Medium Priority |
| Go source files | 24 | Auto-fixed by go.mod |
| **Total impact** | **76+** | One org name change |

