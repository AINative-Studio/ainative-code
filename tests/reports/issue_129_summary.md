# Issue #129 - Quick Summary

## Problem
The `--json` flag for `zerodb table` commands was declared but never registered.

## Before Fix
```bash
$ ainative-code zerodb table list --json
Error: unknown flag: --json
```

## After Fix
```bash
$ ainative-code zerodb table list --json
[
  {
    "id": "tbl_abc123",
    "name": "users",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

## The Fix
**File:** `internal/cmd/zerodb_table.go`

**Changed 2 lines (line 238-239):**
```go
// Before:
// Note: --json flag is inherited from parent zerodbCmd
}

// After:
// Table output flags - register --json flag for all table commands
zerodbTableCmd.PersistentFlags().BoolVar(&tableOutputJSON, "json", false, "output in JSON format")
}
```

## Impact
✅ All 6 subcommands now support --json flag:
1. `zerodb table create --json`
2. `zerodb table list --json`
3. `zerodb table insert --json`
4. `zerodb table query --json`
5. `zerodb table update --json`
6. `zerodb table delete --json`

## Testing
- ✅ 18 integration tests - ALL PASSED
- ✅ 6 help output tests - ALL PASSED
- ✅ Shell script validation - ALL PASSED
- ✅ No regressions

## Status
🎉 **FIXED AND TESTED** - Ready for merge
