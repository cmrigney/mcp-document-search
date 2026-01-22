package storage

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	embeddingDimension = 1536
)

// Database handles SQLite operations with vector search
type Database struct {
	db *sql.DB
}

// Document represents a document in the database
type Document struct {
	ID          int64
	Source      string
	SourceType  string
	IndexedAt   time.Time
	ContentSize int
	ChunkCount  int
	Title       string
}

// Chunk represents a text chunk with its embedding
type Chunk struct {
	ID          int64
	DocumentID  int64
	ChunkIndex  int
	Content     string
	StartOffset int
	EndOffset   int
	Embedding   []float32
}

// SearchResult represents a search result
type SearchResult struct {
	Content    string
	Source     string
	ChunkIndex int
	Score      float64
}

// NewDatabase creates a new database connection and initializes schema
// Note: sqlite_vec.Auto() must be called before creating the database
func NewDatabase(dbPath string) (*Database, error) {
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test if vec extension is loaded (should be auto-loaded via sqlite_vec.Auto())
	var version string
	err = db.QueryRow("SELECT vec_version()").Scan(&version)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite-vec not available (did you call sqlite_vec.Auto()?): %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// runMigrations creates the database schema
func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source TEXT UNIQUE NOT NULL,
		source_type TEXT NOT NULL,
		indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		content_size INTEGER,
		chunk_count INTEGER,
		title TEXT
	);

	CREATE TABLE IF NOT EXISTS chunks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		document_id INTEGER NOT NULL,
		chunk_index INTEGER NOT NULL,
		content TEXT NOT NULL,
		start_offset INTEGER NOT NULL,
		end_offset INTEGER NOT NULL,
		embedding BLOB NOT NULL,
		FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks(document_id);
	`

	_, err := db.Exec(schema)
	return err
}

// IndexDocument stores a document with its chunks and embeddings
func (d *Database) IndexDocument(source, sourceType, title string, chunks []Chunk) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing document if it exists (for reindexing)
	_, err = tx.Exec("DELETE FROM documents WHERE source = ?", source)
	if err != nil {
		return fmt.Errorf("failed to delete existing document: %w", err)
	}

	// Calculate content size
	contentSize := 0
	for _, chunk := range chunks {
		contentSize += len(chunk.Content)
	}

	// Insert document
	result, err := tx.Exec(
		"INSERT INTO documents (source, source_type, content_size, chunk_count, title) VALUES (?, ?, ?, ?, ?)",
		source, sourceType, contentSize, len(chunks), title,
	)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	documentID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get document ID: %w", err)
	}

	// Insert chunks
	stmt, err := tx.Prepare("INSERT INTO chunks (document_id, chunk_index, content, start_offset, end_offset, embedding) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare chunk insert: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		embeddingBlob, err := serializeEmbedding(chunk.Embedding)
		if err != nil {
			return fmt.Errorf("failed to serialize embedding: %w", err)
		}

		_, err = stmt.Exec(documentID, chunk.ChunkIndex, chunk.Content, chunk.StartOffset, chunk.EndOffset, embeddingBlob)
		if err != nil {
			return fmt.Errorf("failed to insert chunk: %w", err)
		}
	}

	return tx.Commit()
}

// Search performs vector similarity search
func (d *Database) Search(queryEmbedding []float32, topK int, minScore float64, sourceFilter string) ([]SearchResult, error) {
	if len(queryEmbedding) != embeddingDimension {
		return nil, fmt.Errorf("invalid query embedding dimension: got %d, expected %d", len(queryEmbedding), embeddingDimension)
	}

	queryBlob, err := serializeEmbedding(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query embedding: %w", err)
	}

	// Build query with optional source filter
	query := `
		SELECT c.content, d.source, c.chunk_index,
		       (1 - vec_distance_cosine(c.embedding, ?)) AS score
		FROM chunks c
		JOIN documents d ON c.document_id = d.id
		WHERE 1=1
	`
	args := []interface{}{queryBlob}

	if sourceFilter != "" {
		query += " AND d.source = ?"
		args = append(args, sourceFilter)
	}

	query += " ORDER BY score DESC LIMIT ?"
	args = append(args, topK)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		err := rows.Scan(&result.Content, &result.Source, &result.ChunkIndex, &result.Score)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Filter by minimum score
		if result.Score >= minScore {
			results = append(results, result)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}

// ListDocuments returns all indexed documents with optional source type filter
func (d *Database) ListDocuments(sourceTypeFilter string) ([]Document, error) {
	query := `
		SELECT id, source, source_type, indexed_at, content_size, chunk_count, COALESCE(title, '')
		FROM documents
		WHERE 1=1
	`
	args := []interface{}{}

	if sourceTypeFilter != "" {
		query += " AND source_type = ?"
		args = append(args, sourceTypeFilter)
	}

	query += " ORDER BY indexed_at DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Source, &doc.SourceType, &doc.IndexedAt, &doc.ContentSize, &doc.ChunkCount, &doc.Title)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}

// DeleteDocument deletes a document and its chunks by source
func (d *Database) DeleteDocument(source string) error {
	result, err := d.db.Exec("DELETE FROM documents WHERE source = ?", source)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found: %s", source)
	}

	return nil
}

// DocumentExists checks if a document exists by source
func (d *Database) DocumentExists(source string) (bool, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM documents WHERE source = ?", source).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check document existence: %w", err)
	}
	return count > 0, nil
}

// serializeEmbedding converts float32 slice to binary blob
func serializeEmbedding(embedding []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, embedding)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserializeEmbedding converts binary blob to float32 slice
func deserializeEmbedding(data []byte) ([]float32, error) {
	embedding := make([]float32, embeddingDimension)
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &embedding)
	if err != nil {
		return nil, err
	}
	return embedding, nil
}
