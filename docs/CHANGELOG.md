# Changelog

## [v1.1.0] - 2024-01-22

### 🎉 Major Improvement: Static Linking

**Breaking Change**: Removed requirement for external sqlite-vec installation

#### Added
- Static linking of sqlite-vec via `github.com/asg017/sqlite-vec-go-bindings/cgo`
- Self-contained binary with zero external extension dependencies
- Automatic sqlite-vec initialization via `sqlite_vec.Auto()`

#### Changed
- **Simplified setup**: No longer need to install sqlite-vec separately
- **Removed `SQLITE_VEC_PATH` environment variable**: No longer needed
- Updated `NewDatabase()` signature: removed `vecExtensionPath` parameter
- Updated all documentation to reflect simpler installation process

#### Benefits
- ✅ **Easier deployment**: Just build and run
- ✅ **Self-contained binary**: 14MB with everything included
- ✅ **Platform portable**: Binary works anywhere (same OS/arch)
- ✅ **Fewer errors**: One less thing to configure
- ✅ **Faster startup**: Extension already in memory

#### Migration Guide

**Before (v1.0.0)**:
```bash
# Install extension
brew install sqlite-vec

# Set path
export SQLITE_VEC_PATH=/opt/homebrew/lib/vec0.dylib

# Set API key
export OPENAI_API_KEY="sk-..."

# Run
./doc-search
```

**After (v1.1.0)**:
```bash
# Just set API key and run!
export OPENAI_API_KEY="sk-..."
./doc-search
```

If you have `SQLITE_VEC_PATH` in your environment or config, you can safely remove it.

#### Build Notes
- First build takes 30-60 seconds (compiles sqlite-vec from C source)
- Subsequent builds are cached and fast (~2-3 seconds)
- May see deprecation warnings on macOS about `sqlite3_auto_extension` (safe to ignore)

---

## [v1.0.0] - 2024-01-22

### Initial Release

#### Features
- MCP server with four tools: `search`, `index`, `list`, `delete`
- Semantic search using OpenAI text-embedding-3-small
- SQLite with sqlite-vec for vector similarity search
- Smart text chunking with word boundary detection
- HTML text extraction for URL indexing
- Support for files, URLs, and direct content
- Configurable chunk size and overlap
- Source filtering in search
- Document type tracking (file/url/content)

#### Technical Stack
- Go 1.21+
- OpenAI embeddings (1536 dimensions)
- SQLite + sqlite-vec
- Cosine similarity search
- CGO for SQLite integration

#### Tools

**search**
- Natural language semantic search
- Configurable top-k and minimum score
- Source filtering support

**index**
- Index files, URLs, or direct content
- Re-index support
- Automatic HTML text extraction

**list**
- List all indexed documents
- Filter by source type
- Full metadata display

**delete**
- Remove indexed documents
- Cascade deletion of chunks

#### Documentation
- README.md - Main documentation
- SETUP.md - Installation guide
- EXAMPLES.md - Usage examples
- IMPLEMENTATION_SUMMARY.md - Technical details
