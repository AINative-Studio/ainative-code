# CGO Cross-Compilation Notes

## Issue: SQLite Requires CGO

The project uses `github.com/mattn/go-sqlite3` which requires CGO (C bindings) to work. This creates challenges for cross-platform builds.

## Problem Discovered

During v0.1.9 release, binaries were built with `CGO_ENABLED=0`, which completely broke SQLite functionality with this error:

```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

## Current Build Strategy

### Local Builds (Works)
```bash
# Build for current platform
make build  # Uses CGO_ENABLED=1 by default
```

### Cross-Platform Builds (Limited)
```bash
# Only builds for current platform
make build-all
```

## Solutions for Cross-Platform Builds

### Option 1: Use GitHub Actions (RECOMMENDED)
GitHub Actions can build for all platforms using proper toolchains. See `.github/workflows/release.yml`.

### Option 2: Use Docker with Cross-Compilers
```bash
# Use Docker with CGO cross-compilers
docker run --rm -v "$PWD":/app -w /app \
  golang:1.25 make build-all
```

### Option 3: Switch to Pure Go SQLite
Replace `github.com/mattn/go-sqlite3` with `modernc.org/sqlite`:

**Pros:**
- No CGO required
- Works with `CGO_ENABLED=0`
- Easier cross-compilation

**Cons:**
- Slightly slower performance
- Different driver registration
- Code changes required

## Recommendation

For release builds, use GitHub Actions or CI/CD pipelines that can properly configure CGO cross-compilation toolchains for each target platform.

For local development, the current `make build` works perfectly.

## Current Workaround

The `make build-all` target now only builds for the native platform to avoid cross-compilation errors. Use CI/CD for full multi-platform releases.
