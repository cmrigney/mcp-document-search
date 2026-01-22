# Usage Examples

This document provides practical examples of using the MCP Document Search Server.

## Example 1: Index and Search a Local File

### 1. Create a test document

```bash
cat > golang-guide.txt << 'EOF'
Go is a statically typed, compiled programming language designed at Google.
It is syntactically similar to C, but with memory safety, garbage collection,
structural typing, and CSP-style concurrency. Go was designed to improve
programming productivity in an era of multicore processors, networked systems,
and large codebases.

The Go programming language is often referred to as Golang because of its
domain name, golang.org. Go is used by many companies including Google,
Uber, Dropbox, and Docker for building scalable, high-performance applications.
EOF
```

### 2. Index the file

Tool: `index`
```json
{
  "file_path": "/absolute/path/to/golang-guide.txt"
}
```

Response:
```json
{
  "source": "/absolute/path/to/golang-guide.txt",
  "source_type": "file",
  "chunk_count": 1,
  "message": "Successfully indexed /absolute/path/to/golang-guide.txt (1 chunks)"
}
```

### 3. Search for information

Tool: `search`
```json
{
  "query": "What companies use Go?",
  "top_k": 2
}
```

Response:
```json
{
  "results": [
    {
      "content": "Go is used by many companies including Google, Uber, Dropbox, and Docker...",
      "source": "/absolute/path/to/golang-guide.txt",
      "chunk_index": 0,
      "score": 0.87
    }
  ],
  "count": 1
}
```

## Example 2: Index a URL

### Index a documentation page

Tool: `index`
```json
{
  "url": "https://golang.org/doc/effective_go"
}
```

Response:
```json
{
  "source": "https://golang.org/doc/effective_go",
  "source_type": "url",
  "chunk_count": 45,
  "message": "Successfully indexed https://golang.org/doc/effective_go (45 chunks)"
}
```

### Search across all indexed documents

Tool: `search`
```json
{
  "query": "How do I handle errors in Go?",
  "top_k": 5,
  "min_score": 0.7
}
```

This will search across both the local file and the URL.

## Example 3: Filter Search by Source

Search only in the URL we indexed:

Tool: `search`
```json
{
  "query": "error handling",
  "top_k": 3,
  "source_filter": "https://golang.org/doc/effective_go"
}
```

## Example 4: List All Indexed Documents

### List all documents

Tool: `list`
```json
{}
```

Response:
```json
{
  "documents": [
    {
      "source": "https://golang.org/doc/effective_go",
      "source_type": "url",
      "chunk_count": 45,
      "content_size": 52341,
      "indexed_at": "2024-01-22 10:30:15",
      "title": "Effective Go"
    },
    {
      "source": "/absolute/path/to/golang-guide.txt",
      "source_type": "file",
      "chunk_count": 1,
      "content_size": 542,
      "indexed_at": "2024-01-22 10:25:10",
      "title": ""
    }
  ],
  "count": 2
}
```

### List only files

Tool: `list`
```json
{
  "source_type": "file"
}
```

### List only URLs

Tool: `list`
```json
{
  "source_type": "url"
}
```

## Example 5: Index Direct Content

Useful for indexing generated or dynamic content:

Tool: `index`
```json
{
  "content": "This is custom content that I want to index. It could be generated dynamically or come from a database.",
  "source": "custom-content-id-123"
}
```

## Example 6: Re-index a Document

If you've updated a file and want to re-index it:

Tool: `index`
```json
{
  "file_path": "/absolute/path/to/golang-guide.txt",
  "reindex": true
}
```

Without `reindex: true`, you'll get an error if the document is already indexed.

## Example 7: Delete a Document

Remove a document from the index:

Tool: `delete`
```json
{
  "source": "/absolute/path/to/golang-guide.txt"
}
```

Response:
```json
{
  "source": "/absolute/path/to/golang-guide.txt",
  "deleted": true,
  "message": "Successfully deleted /absolute/path/to/golang-guide.txt"
}
```

## Example 8: Semantic Search vs Keyword Search

The power of semantic search is that it understands meaning, not just keywords.

### Index a document about databases

```bash
cat > databases.txt << 'EOF'
PostgreSQL is a powerful, open source object-relational database system.
It has more than 30 years of active development. PostgreSQL is known for
its reliability, feature robustness, and performance. It supports both
SQL (relational) and JSON (non-relational) querying.
EOF
```

Tool: `index`
```json
{
  "file_path": "/absolute/path/to/databases.txt"
}
```

### Search with natural language

Tool: `search`
```json
{
  "query": "What are some reliable database systems?",
  "top_k": 3
}
```

Even though the word "reliable" appears once, and "database systems" is phrased differently, the semantic search will find this content because it understands the meaning.

## Example 9: Building a Documentation Search

### Index multiple documentation pages

```json
[
  {"url": "https://docs.example.com/getting-started"},
  {"url": "https://docs.example.com/api-reference"},
  {"url": "https://docs.example.com/deployment"},
  {"url": "https://docs.example.com/troubleshooting"}
]
```

### Search for answers

Users can ask natural questions:
- "How do I deploy the application?"
- "What API endpoints are available?"
- "My service won't start, what should I check?"

The semantic search will find relevant sections even if the exact words aren't used.

## Example 10: Adjusting Search Sensitivity

### High precision (only very relevant results)

Tool: `search`
```json
{
  "query": "Go concurrency patterns",
  "top_k": 5,
  "min_score": 0.85
}
```

### More exploratory (include loosely related results)

Tool: `search`
```json
{
  "query": "Go concurrency patterns",
  "top_k": 10,
  "min_score": 0.6
}
```

## Tips for Best Results

1. **Index granularly**: Break large documents into logical sections before indexing
2. **Use descriptive sources**: When using `content` + `source`, use meaningful source IDs
3. **Adjust chunk size**: For technical docs with code, smaller chunks (500-800) work better
4. **Re-index regularly**: Keep your index fresh by re-indexing updated documents
5. **Filter by source**: Use `source_filter` when you know which document to search
6. **Tune min_score**: Experiment with `min_score` to find the right balance for your use case

## Common Patterns

### Pattern 1: Knowledge Base

```
1. Index all documentation pages
2. Users ask questions in natural language
3. System returns relevant sections with sources
4. Users can click through to full documents
```

### Pattern 2: Code Search

```
1. Index README files, API docs, and code comments
2. Developers search: "How do I authenticate?"
3. System returns examples and relevant code sections
4. Developers find answers faster than browsing
```

### Pattern 3: Research Assistant

```
1. Index research papers, articles, and notes
2. Researcher asks: "What studies discuss X?"
3. System returns relevant passages from multiple sources
4. Researcher discovers connections across documents
```
