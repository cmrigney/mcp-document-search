# Static Linking: sqlite-vec Integration

## Overview

The MCP Document Search Server now uses **statically linked** sqlite-vec instead of requiring a separate extension installation. This significantly simplifies deployment and improves the user experience.

## Technical Implementation

### Dependencies

Added the sqlite-vec Go bindings:
```go
import sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
```

Package: `github.com/asg017/sqlite-vec-go-bindings v0.1.6`

### Initialization

In `cmd/doc-search/main.go`:
```go
func main() {
    // Enable sqlite-vec for all future database connections
    sqlite_vec.Auto()
    log.Println("sqlite-vec enabled (statically linked)")

    // ... rest of initialization
}
```

The `sqlite_vec.Auto()` function registers sqlite-vec as an auto-extension, making it available to all SQLite connections opened after this call.

### Database Layer Changes

**Before** (`internal/storage/database.go`):
```go
func NewDatabase(dbPath, vecExtensionPath string) (*Database, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Load external extension
    if vecExtensionPath != "" {
        _, err = db.Exec("SELECT load_extension(?)", vecExtensionPath)
        if err != nil {
            return nil, err
        }
    }
    // ...
}
```

**After**:
```go
func NewDatabase(dbPath string) (*Database, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Extension is auto-loaded via sqlite_vec.Auto()
    // Just verify it's available
    var version string
    err = db.QueryRow("SELECT vec_version()").Scan(&version)
    if err != nil {
        return nil, fmt.Errorf("sqlite-vec not available: %w", err)
    }
    // ...
}
```

### Configuration Simplification

**Removed from `internal/config/config.go`**:
```go
type Config struct {
    // ... other fields
    SQLiteVecPath  string  // REMOVED
}
```

Users no longer need to set `SQLITE_VEC_PATH` environment variable.

## Build Process

### First Build

The first build compiles sqlite-vec from source (C code):

```bash
CGO_ENABLED=1 go build -o doc-search ./cmd/doc-search
```

This takes **30-60 seconds** on first run because:
1. CGO compiles the C source for sqlite-vec
2. All C objects are built and linked
3. The complete extension is embedded in the binary

### Subsequent Builds

Go caches the compiled extension, so rebuilds are fast (2-3 seconds) unless:
- You run `go clean -cache`
- The sqlite-vec-go-bindings version changes
- You switch platforms/architectures

### Build Warnings (macOS)

You may see deprecation warnings on macOS:
```
warning: 'sqlite3_auto_extension' is deprecated: first deprecated in macOS 10.10
```

These are **expected and safe to ignore**. They come from SQLite's recommendation against process-global auto extensions on Apple platforms, but the Go bindings handle this correctly.

## Binary Characteristics

### Size
- **14MB** (includes Go runtime, SQLite, sqlite-vec, and all dependencies)
- Larger than external-extension version (~12MB) but much more portable

### Portability
The binary is **platform-specific** but **self-contained**:
- ✅ Can run on any machine with the same OS/architecture
- ✅ No external .so/.dylib files needed
- ✅ No installation steps for users
- ❌ Must build separately for different platforms (Linux, macOS, Windows)

### Cross-Compilation

For cross-platform builds with CGO:
```bash
# Linux from macOS
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o doc-search-linux ./cmd/doc-search

# Requires cross-compilation toolchain (e.g., musl-cross)
```

## Performance

### Embedding Performance
**No performance difference** from external extension:
- Same sqlite-vec C code
- Same vector algorithms
- Same query performance
- Same memory usage

### Startup Performance
**Slightly faster** than external extension:
- No file I/O to load .so/.dylib
- Extension is already in memory
- No dynamic linking overhead

## Maintenance

### Updating sqlite-vec

To update to a newer version:
```bash
go get -u github.com/asg017/sqlite-vec-go-bindings/cgo
go mod tidy
CGO_ENABLED=1 go build -o doc-search ./cmd/doc-search
```

This will:
1. Fetch the new bindings (which includes the latest sqlite-vec C code)
2. Rebuild with the updated extension
3. Invalidate Go's build cache for this dependency

### Version Management

Check the current sqlite-vec version at runtime:
```sql
SELECT vec_version();
```

Or in code:
```go
var version string
db.QueryRow("SELECT vec_version()").Scan(&version)
log.Printf("sqlite-vec version: %s", version)
```

## Deployment Scenarios

### Scenario 1: Single Binary Distribution

**Use Case**: Distribute to non-technical users

**Steps**:
1. Build the binary
2. Create a simple wrapper script:
   ```bash
   #!/bin/bash
   export OPENAI_API_KEY="${OPENAI_API_KEY:-sk-...}"
   exec ./doc-search
   ```
3. Package as ZIP or tar.gz
4. Users just extract and run

**Benefits**: Zero installation, works immediately

### Scenario 2: Docker Container

**Use Case**: Cloud deployment

**Dockerfile**:
```dockerfile
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 go build -o doc-search ./cmd/doc-search

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/doc-search /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/doc-search"]
```

**Benefits**: Reproducible builds, no host dependencies

### Scenario 3: System Installation

**Use Case**: System-wide installation via package manager

**Create a .deb or .rpm**:
- Binary goes to `/usr/local/bin/doc-search`
- Systemd service file
- Configuration in `/etc/doc-search/config.env`

**Benefits**: Familiar installation process, managed updates

## Troubleshooting

### "undefined reference" errors during build

**Cause**: CGO is disabled or C compiler not found

**Solution**:
```bash
export CGO_ENABLED=1
# Install compiler if needed
xcode-select --install  # macOS
sudo apt install build-essential  # Linux
```

### "sqlite-vec not available" at runtime

**Cause**: `sqlite_vec.Auto()` not called before database connection

**Solution**: Ensure main.go calls `sqlite_vec.Auto()` before creating database:
```go
func main() {
    sqlite_vec.Auto()  // Must be first!
    // ... then create database
}
```

### Build is very slow

**Expected on first build**. The extension compiles from C source which takes time.

**If every build is slow**:
- Check if Go build cache is disabled
- Verify `go env GOCACHE` points to a writable directory
- Try `go clean -cache` then rebuild (to reset corrupted cache)

## Comparison: External vs Static

| Aspect | External Extension | Static Linking |
|--------|-------------------|----------------|
| **Setup** | Install extension, set env var | Just build binary |
| **Binary Size** | 12MB | 14MB |
| **Dependencies** | Requires .so/.dylib file | Self-contained |
| **Portability** | Platform + extension paths | Platform-specific binary only |
| **Updates** | Update extension separately | Rebuild binary |
| **Startup** | Load from disk | Already in memory |
| **Debugging** | Two components | Single component |
| **User Experience** | Complex | Simple |

## References

- [sqlite-vec Go Documentation](https://alexgarcia.xyz/sqlite-vec/go.html)
- [sqlite-vec GitHub](https://github.com/asg017/sqlite-vec)
- [Go CGO Documentation](https://golang.org/cmd/cgo/)
