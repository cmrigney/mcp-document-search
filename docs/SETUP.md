# Setup Guide

This guide will help you set up the MCP Document Search Server. Setup is much simpler now because **sqlite-vec is statically linked** into the binary!

## Quick Start

### Step 1: Get OpenAI API Key

1. Go to https://platform.openai.com/api-keys
2. Create a new API key
3. Copy the key (starts with `sk-...`)

### Step 2: Set Environment Variable

```bash
export OPENAI_API_KEY="sk-..."
```

### Step 3: Build the Server

#### Option A: Using Task (Recommended)

Install [Task](https://taskfile.dev) if you haven't already:
```bash
# macOS
brew install go-task

# Linux/other
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
```

Then build:
```bash
cd mcp-document-search
task build
```

#### Option B: Manual Build

```bash
cd mcp-document-search
mkdir -p bin
CGO_ENABLED=1 go build -o bin/doc-search ./cmd/doc-search
```

**Note**: The first build will be slower (30-60 seconds) as it compiles sqlite-vec, but subsequent builds will be much faster due to caching.

### Step 4: Run the Server

```bash
./bin/doc-search
```

You should see:
```
sqlite-vec enabled (statically linked)
Database initialized at: db_data/doc_search.db
OpenAI embeddings client initialized
Chunker initialized (size: 1000, overlap: 100)
URL fetcher initialized
Search service initialized
Starting MCP server on stdio...
MCP Doc Search server started on stdio
```

That's it! No separate extension installation required.

## Testing the Server

### Option A: With MCP Inspector

Using Task:
```bash
task inspector
```

Or manually:
```bash
npx @modelcontextprotocol/inspector ./bin/doc-search
```

This will open a web interface where you can test all four tools.

### Option B: With Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "doc-search": {
      "command": "/absolute/path/to/mcp-document-search/bin/doc-search",
      "env": {
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

Restart Claude Desktop and the server should be available.

## Try Your First Search

### 1. Create a test document

```bash
echo "The quick brown fox jumps over the lazy dog. This is a test document for semantic search." > test.txt
```

### 2. Index it using the `index` tool

```json
{
  "file_path": "/absolute/path/to/test.txt"
}
```

### 3. Search using the `search` tool

```json
{
  "query": "What does the fox do?",
  "top_k": 3
}
```

You should see results with similarity scores.

## Build Requirements

### For Building from Source

- **Go**: Version 1.21 or later
- **CGO**: Must be enabled (`CGO_ENABLED=1`)
- **C Compiler**: gcc or clang

#### Installing Build Tools

**macOS:**
```bash
# Install Xcode Command Line Tools (includes clang)
xcode-select --install

# Or install full Xcode from the App Store
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y build-essential golang
```

**Fedora/RHEL:**
```bash
sudo dnf install -y gcc golang
```

### For Running the Binary

Once built, the binary is **self-contained** and only needs:
- The OpenAI API key
- No external dependencies or extensions!

## Configuration Options

All configuration is via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | ✅ Yes | - | OpenAI API key for embeddings |
| `DB_PATH` | ❌ No | `db_data/doc_search.db` | Path to SQLite database file |
| `CHUNK_SIZE` | ❌ No | `1000` | Characters per chunk |
| `OVERLAP` | ❌ No | `100` | Overlap between chunks |

Example with custom settings:

```bash
export OPENAI_API_KEY="sk-..."
export DB_PATH="/var/lib/doc-search/search.db"
export CHUNK_SIZE=500
export OVERLAP=50
./bin/doc-search
```

## Troubleshooting

### "CGO_ENABLED" errors during build

Make sure CGO is enabled:
```bash
export CGO_ENABLED=1
mkdir -p bin
go build -o bin/doc-search ./cmd/doc-search
```

### "C compiler not found"

**macOS:**
```bash
xcode-select --install
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# Fedora/RHEL
sudo dnf install gcc
```

### "OPENAI_API_KEY environment variable is required"

Set your API key:
```bash
export OPENAI_API_KEY="sk-..."
```

### Build is very slow

The **first build** is slow (30-60 seconds) because it compiles sqlite-vec from source. This is normal! Subsequent builds will be much faster because Go caches the compiled extension.

### "Permission denied" when running

Make the binary executable:
```bash
chmod +x bin/doc-search
```

### API Rate Limiting

If you hit OpenAI rate limits:
- The server automatically retries with exponential backoff
- You can reduce `CHUNK_SIZE` to create fewer chunks per document
- Process large documents one at a time

## What Changed from External Extension?

**Before (external extension required):**
1. ❌ Install sqlite-vec via Homebrew or compile from source
2. ❌ Set `SQLITE_VEC_PATH` environment variable
3. ❌ Ensure extension loads correctly
4. ❌ Different setup per platform

**Now (statically linked):**
1. ✅ Build the binary once
2. ✅ Run anywhere (just needs OpenAI API key)
3. ✅ No platform-specific extension paths
4. ✅ Self-contained and portable

## Binary Distribution

If you want to distribute the built binary:

1. Build it:
   ```bash
   mkdir -p bin
   CGO_ENABLED=1 go build -o bin/doc-search ./cmd/doc-search
   ```

2. The resulting binary (14MB) is **self-contained** and includes:
   - All Go code
   - SQLite
   - sqlite-vec extension
   - All dependencies

3. Users only need to:
   - Download the binary
   - Set `OPENAI_API_KEY`
   - Run it!

No installation required on the target machine (beyond standard system libraries).

## Next Steps

- Check out [EXAMPLES.md](EXAMPLES.md) for detailed usage examples
- Read [README.md](README.md) for full documentation
- Index your first documents and start searching!
