# MCP Document Search Server

A Model Context Protocol (MCP) server that provides semantic search through documents using chunked embeddings and vector similarity search.

**Notice:** This was built by Claude Code.

## Features

- **Semantic Search**: Search through indexed documents using natural language queries
- **Multiple Sources**: Index local files, URLs, or direct content
- **Vector Search**: Uses OpenAI embeddings and SQLite with sqlite-vec for efficient similarity search
- **Smart Chunking**: Chunks text with word boundary detection and configurable overlap
- **HTML Support**: Automatically extracts text from HTML pages when indexing URLs
- **Four Tools**: `search`, `index`, `list`, and `delete` for complete document management

## Architecture

- **Embeddings**: OpenAI text-embedding-3-small (1536 dimensions)
- **Vector DB**: SQLite with sqlite-vec extension
- **Similarity**: Cosine similarity
- **Chunking**: 1000 chars with 100 char overlap (configurable), respecting word boundaries

## Prerequisites

### 1. OpenAI API Key

You need an OpenAI API key:
```bash
export OPENAI_API_KEY="sk-..."
```

### 2. Go and CGO (for building from source)

- Go 1.21 or later
- CGO enabled (required for SQLite and sqlite-vec)
- C compiler (gcc or clang)

**Note**: sqlite-vec is **statically linked** into the binary, so you don't need to install it separately!

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/cmrigney/mcp-document-search.git
cd mcp-document-search

# Build
mkdir -p bin
CGO_ENABLED=1 go build -o bin/doc-search ./cmd/doc-search

# Run
export OPENAI_API_KEY="sk-..."
./bin/doc-search
```

## Quick Start with Task

The project includes a Taskfile for easy building and testing. First, install [Task](https://taskfile.dev). Then use these commands:

```bash
# Build the binary
task build

# Run all tests
task test

# Test with MCP Inspector
task inspector

# See all available tasks
task --list
```

## Configuration

The server is configured via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | Yes | - | OpenAI API key for embeddings |
| `DB_PATH` | No | `db_data/doc_search.db` | Path to SQLite database file |
| `CHUNK_SIZE` | No | `1000` | Size of text chunks in characters |
| `OVERLAP` | No | `100` | Overlap between chunks in characters |

## Usage

### With Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "doc-search": {
      "command": "/path/to/mcp-document-search/bin/doc-search",
      "env": {
        "OPENAI_API_KEY": "sk-..."
      }
    }
  }
}
```

## Tools

### 1. search

Semantic search through indexed documents.

**Arguments:**
- `query` (required): Search query
- `top_k` (optional): Number of results to return (default: 5)
- `min_score` (optional): Minimum similarity score 0-1 (default: 0.3)
- `source_filter` (optional): Filter to specific source (file path or URL)

**Example:**
```json
{
  "query": "How do I configure authentication?",
  "top_k": 3,
  "min_score": 0.3
}
```

### 2. index

Index a file, URL, or content for semantic search.

**Arguments** (provide exactly one source):
- `file_path` (optional): Path to file to index
- `url` (optional): URL to fetch and index
- `content` + `source` (optional): Direct content with source identifier
- `reindex` (optional): Force re-index if already indexed (default: false)

**Examples:**

Index a file:
```json
{
  "file_path": "/path/to/document.txt"
}
```

Index a URL:
```json
{
  "url": "https://example.com/docs/guide.html"
}
```

Index direct content:
```json
{
  "content": "This is the content to index...",
  "source": "manual-entry-1"
}
```

### 3. list

List all indexed documents with metadata.

**Arguments:**
- `source_type` (optional): Filter by type: "file" or "url" (empty for all)

**Example:**
```json
{
  "source_type": "url"
}
```

### 4. delete

Remove an indexed document from the database.

**Arguments:**
- `source` (required): Source to delete (file path or URL)

**Example:**
```json
{
  "source": "/path/to/document.txt"
}
```

## Database Schema

### documents table
- `id`: Auto-incrementing primary key
- `source`: File path or URL (unique)
- `source_type`: "file", "url", or "content"
- `indexed_at`: Timestamp when indexed
- `content_size`: Total content size in characters
- `chunk_count`: Number of chunks
- `title`: Optional title (extracted from HTML)

### chunks table
- `id`: Auto-incrementing primary key
- `document_id`: Foreign key to documents
- `chunk_index`: Index of chunk within document
- `content`: Text content of chunk
- `start_offset`: Start position in original document
- `end_offset`: End position in original document
- `embedding`: 1536-dimensional vector (6144 bytes)

## Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/chunker
```

### Project Structure

```
mcp-document-search/
├── cmd/doc-search/          # Main entry point
├── internal/
│   ├── chunker/            # Text chunking logic
│   ├── fetcher/            # URL content fetcher
│   ├── embeddings/         # OpenAI API client
│   ├── storage/            # SQLite + sqlite-vec
│   ├── search/             # Search orchestration
│   └── config/             # Configuration
└── pkg/server/             # MCP server implementation
```

## Troubleshooting

### CGO errors during build

If you see CGO-related errors:
1. Ensure `CGO_ENABLED=1`
2. Install a C compiler (gcc or clang)
3. On macOS: `xcode-select --install`

### API rate limits

If you hit OpenAI rate limits:
- The client automatically retries with exponential backoff
- Consider reducing chunk size or processing documents in smaller batches

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or pull request.
