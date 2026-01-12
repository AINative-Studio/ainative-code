# CRITICAL BUG: Binary Compiled Without CGO Breaks SQLite

## Severity: CRITICAL 🚨

## Discovery
Found during user acceptance testing of v0.1.9 release binary.

## Problem
The release binaries are compiled with `CGO_ENABLED=0`, which completely breaks SQLite functionality. This affects ALL session-related commands.

## Impact
- ❌ `session list` - BROKEN
- ❌ `session search` - BROKEN
- ❌ `session export` - BROKEN
- ❌ `session delete` - BROKEN
- ❌ `session show` - BROKEN
- ❌ Any command that uses SQLite database - BROKEN

## Error Message
```
Error: failed to open database: failed to initialize database:
[DB_CONNECTION_FAILED] Failed to connect to database '/Users/aideveloper/.ainative/ainative.db':
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

## Root Cause
**File**: `Makefile` line 62

```makefile
# WRONG - Current implementation
GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $$OUTPUT_NAME $(CMD_DIR);
```

The build-all target explicitly sets `CGO_ENABLED=0` for all platform builds, which disables CGO and breaks go-sqlite3.

## Why CGO_ENABLED=0 Was Used
Likely reasons:
1. Attempt to create static binaries
2. Avoid cross-compilation issues
3. Simplify builds

## Problem with CGO_ENABLED=0
- go-sqlite3 is a CGO-based library
- It requires CGO to compile C bindings to SQLite
- With CGO_ENABLED=0, it falls back to a stub that does nothing
- All database operations fail immediately

## Solution Options

### Option 1: Enable CGO for All Platforms (RECOMMENDED)
```makefile
# Build with CGO enabled
GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=1 $(GOBUILD) -tags "$(SQLITE_TAGS)" $(LDFLAGS) -o $$OUTPUT_NAME $(CMD_DIR);
```

**Pros:**
- Works correctly
- Full SQLite support
- Consistent behavior across platforms

**Cons:**
- Requires C compiler for each target platform
- Cross-compilation is more complex
- Binaries are dynamically linked (larger, platform-specific)

### Option 2: Use Pure Go SQLite (modernc.org/sqlite)
Replace github.com/mattn/go-sqlite3 with modernc.org/sqlite

**Pros:**
- Pure Go, no CGO required
- Works with CGO_ENABLED=0
- Easier cross-compilation

**Cons:**
- Slightly slower than native SQLite
- Different driver name
- Code changes required

### Option 3: Hybrid Approach
- Build native platform with CGO_ENABLED=1
- Use Docker for cross-compilation with proper CGO setup

## Immediate Fix Required

**Priority**: P0 - BLOCKER
**Affects**: ALL v0.1.9 release binaries
**Action**:
1. Fix Makefile to enable CGO
2. Rebuild all binaries
3. Replace v0.1.9 release assets
4. Notify testers

## Testing Commands
```bash
# Test that SQLite works
./ainative-code-darwin-arm64 session list

# Should NOT return:
# "Binary was compiled with 'CGO_ENABLED=0'"
```

## Related Files
- `Makefile` (line 62 - build-all target)
- `Makefile` (line 50 - regular build uses CGO correctly)

## Note
The regular `build` target (line 50) does NOT disable CGO, which is why local development builds work fine. Only the `build-all` target for releases has this issue.

## Discovered By
User acceptance testing before v0.1.9 deployment
Date: 2026-01-11
