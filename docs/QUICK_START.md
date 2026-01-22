# Quick Start Guide

## Prerequisites
- Go 1.21+
- C compiler (gcc/clang)
- OpenAI API key

## 1. Setup

```bash
# Clone the repository
cd test-doc-mcp

# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."
```

## 2. Build (Choose One)

### With Task (Recommended)
```bash
# Install Task first (if needed)
brew install go-task  # macOS

# Build
task build
```

### Without Task
```bash
mkdir -p bin
CGO_ENABLED=1 go build -o bin/doc-search ./cmd/doc-search
```

## 3. Test with MCP Inspector

### With Task
```bash
task inspector
```

### Without Task
```bash
npx @modelcontextprotocol/inspector ./bin/doc-search
```

## 4. Try It Out

In the MCP Inspector web interface:

### Index a test file
```json
{
  "file_path": "/path/to/your/file.txt"
}
```

### Search
```json
{
  "query": "What is this about?",
  "top_k": 3
}
```

### List documents
```json
{}
```

### Delete a document
```json
{
  "source": "/path/to/your/file.txt"
}
```

## 5. Use with Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "doc-search": {
      "command": "/absolute/path/to/test-doc-mcp/bin/doc-search",
      "env": {
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

Restart Claude Desktop.

## Task Commands

```bash
task build         # Build the binary
task test          # Run tests (verbose)
task test-short    # Run tests (quiet)
task inspector     # Test with MCP Inspector
task clean         # Clean build artifacts
task --list        # Show all tasks
```

## Troubleshooting

### "OPENAI_API_KEY environment variable is required"
```bash
export OPENAI_API_KEY="sk-..."
```

### "CGO_ENABLED" errors
```bash
# macOS
xcode-select --install

# Linux
sudo apt install build-essential
```

### Build is slow
First build takes 30-60 seconds (compiles sqlite-vec). Later builds are cached and fast.

## Next Steps

- Read [README.md](README.md) for full documentation
- See [EXAMPLES.md](EXAMPLES.md) for usage examples
- Check [TASKFILE_USAGE.md](TASKFILE_USAGE.md) for all Task commands
