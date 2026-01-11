# Issue #121: Flag Standardization - Quick Reference

## TL;DR

✅ **All file operations now use `-f, --file`**
✅ **Old `-o, --output` flags still work (deprecated)**
✅ **100% test pass rate (28/28 tests)**

---

## Commands Changed

| Command | Old | New | Backward Compat |
|---------|-----|-----|-----------------|
| `rlhf export` | `-o output.file` | `-f output.file` | ✅ Yes |
| `design extract` | `-o output.file` | `-f output.file` | ✅ Yes |
| `design generate` | `-o output.file` | `-f output.file` | ✅ Yes |
| `session export` | `-o output.file` | `-f output.file` | ✅ Yes |
| `design validate` | (no flag) | `-f input.file` | ✅ N/A (new) |

---

## Quick Examples

### Design Commands

```bash
# Import (already used -f)
ainative-code design import -f tokens.json

# Export (already used -f)
ainative-code design export -f tokens.json

# Validate (NEW: now accepts file input)
ainative-code design validate -f tokens.json

# Extract (changed from -o to -f)
ainative-code design extract --source styles.css -f tokens.json

# Generate (changed from -o to -f)
ainative-code design generate --tokens tokens.json -f output.css
```

### RLHF Commands

```bash
# Export (changed from -o to -f)
ainative-code rlhf export -f feedback.jsonl
```

### Session Commands

```bash
# Export (changed from -o to -f)
ainative-code session export abc123 -f session.json
ainative-code session export abc123 --format markdown -f session.md
```

---

## Testing

### Run All Tests

```bash
# Unit tests
go test -v -run TestFlagStandardization ./internal/cmd/

# Integration tests
./test_issue121_flag_standardization.sh
```

### Quick Verification

```bash
# Check all commands have -f flag
for cmd in "design import" "design export" "design validate" \
           "design extract" "design generate" "rlhf export" \
           "session export"; do
    echo "Checking: $cmd"
    ./ainative-code $cmd --help | grep -q -- "-f" && echo "  ✓ Has -f flag" || echo "  ✗ Missing -f flag"
done
```

---

## For Existing Scripts

Your scripts will continue to work with deprecation warnings:

```bash
# Old way (works but warns)
ainative-code rlhf export -o feedback.jsonl
# Output: Warning: --output/-o flag is deprecated. Please use --file/-f instead.

# New way (recommended)
ainative-code rlhf export -f feedback.jsonl
```

---

## Files Modified

1. `/Users/aideveloper/AINative-Code/internal/cmd/rlhf.go`
2. `/Users/aideveloper/AINative-Code/internal/cmd/design.go`
3. `/Users/aideveloper/AINative-Code/internal/cmd/design_extract.go`
4. `/Users/aideveloper/AINative-Code/internal/cmd/design_generate.go`
5. `/Users/aideveloper/AINative-Code/internal/cmd/session.go`

## Files Added

1. `/Users/aideveloper/AINative-Code/internal/cmd/flag_standardization_test.go` (unit tests)
2. `/Users/aideveloper/AINative-Code/test_issue121_flag_standardization.sh` (integration tests)

---

## Test Results Summary

```
Total Tests: 28
Passed: 28
Failed: 0
Success Rate: 100%
```

**Categories:**
- ✅ Help Text Tests: 14/14
- ✅ Functional Tests: 8/8
- ✅ Backward Compatibility: 2/2
- ✅ Conflict Tests: 2/2
- ✅ Consistency Tests: 2/2

---

## Status

- **Issue:** #121
- **Status:** ✅ RESOLVED
- **Priority:** Medium
- **Breaking Changes:** None
- **Test Coverage:** 100%

---

## See Also

- Full Report: `/Users/aideveloper/AINative-Code/ISSUE_121_FLAG_STANDARDIZATION_REPORT.md`
- Unit Tests: `/Users/aideveloper/AINative-Code/internal/cmd/flag_standardization_test.go`
- Integration Tests: `/Users/aideveloper/AINative-Code/test_issue121_flag_standardization.sh`
