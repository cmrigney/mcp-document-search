package search

import (
	"context"
	"fmt"
	"os"

	"github.com/yourusername/test-doc-mcp/internal/chunker"
	"github.com/yourusername/test-doc-mcp/internal/embeddings"
	"github.com/yourusername/test-doc-mcp/internal/fetcher"
	"github.com/yourusername/test-doc-mcp/internal/storage"
)

// Service orchestrates search operations
type Service struct {
	db             *storage.Database
	embeddingClient *embeddings.Client
	chunker        *chunker.Chunker
	fetcher        *fetcher.Fetcher
}

// NewService creates a new search service
func NewService(db *storage.Database, embClient *embeddings.Client, c *chunker.Chunker, f *fetcher.Fetcher) *Service {
	return &Service{
		db:             db,
		embeddingClient: embClient,
		chunker:        c,
		fetcher:        f,
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query        string
	TopK         int
	MinScore     float64
	SourceFilter string
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results []SearchResultItem
	Count   int
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	Content    string  `json:"content"`
	Source     string  `json:"source"`
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score"`
}

// IndexRequest represents an index request
type IndexRequest struct {
	FilePath string
	URL      string
	Content  string
	Source   string
	Reindex  bool
}

// IndexResponse represents an index response
type IndexResponse struct {
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	ChunkCount int    `json:"chunk_count"`
	Message    string `json:"message"`
}

// ListRequest represents a list request
type ListRequest struct {
	SourceType string
}

// ListResponse represents a list response
type ListResponse struct {
	Documents []DocumentInfo `json:"documents"`
	Count     int            `json:"count"`
}

// DocumentInfo represents document metadata
type DocumentInfo struct {
	Source      string `json:"source"`
	SourceType  string `json:"source_type"`
	ChunkCount  int    `json:"chunk_count"`
	ContentSize int    `json:"content_size"`
	IndexedAt   string `json:"indexed_at"`
	Title       string `json:"title,omitempty"`
}

// DeleteRequest represents a delete request
type DeleteRequest struct {
	Source string
}

// DeleteResponse represents a delete response
type DeleteResponse struct {
	Source  string `json:"source"`
	Deleted bool   `json:"deleted"`
	Message string `json:"message"`
}

// Search performs semantic search
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	// Set defaults
	if req.TopK <= 0 {
		req.TopK = 5
	}

	// Embed query
	embeddings, err := s.embeddingClient.Embed(ctx, []string{req.Query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned for query")
	}

	queryEmbedding := embeddings[0]

	// Search database
	results, err := s.db.Search(queryEmbedding, req.TopK, req.MinScore, req.SourceFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to search database: %w", err)
	}

	// Convert to response format
	items := make([]SearchResultItem, len(results))
	for i, result := range results {
		items[i] = SearchResultItem{
			Content:    result.Content,
			Source:     result.Source,
			ChunkIndex: result.ChunkIndex,
			Score:      result.Score,
		}
	}

	return &SearchResponse{
		Results: items,
		Count:   len(items),
	}, nil
}

// Index indexes content for search
func (s *Service) Index(ctx context.Context, req IndexRequest) (*IndexResponse, error) {
	// Validate exactly one source is provided
	hasFilePath := req.FilePath != ""
	hasURL := req.URL != ""
	hasContent := req.Content != "" && req.Source != ""

	sourceCount := 0
	if hasFilePath {
		sourceCount++
	}
	if hasURL {
		sourceCount++
	}
	if hasContent {
		sourceCount++
	}

	if sourceCount == 0 {
		return nil, fmt.Errorf("must provide exactly one of: file_path, url, or (content + source)")
	}
	if sourceCount > 1 {
		return nil, fmt.Errorf("provide exactly one of: file_path, url, or (content + source)")
	}

	var content, source, sourceType, title string
	var err error

	// Fetch content based on source type
	if hasFilePath {
		source = req.FilePath
		sourceType = "file"
		contentBytes, readErr := os.ReadFile(req.FilePath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read file: %w", readErr)
		}
		content = string(contentBytes)
	} else if hasURL {
		source = req.URL
		sourceType = "url"
		fetchResult, fetchErr := s.fetcher.FetchURL(ctx, req.URL)
		if fetchErr != nil {
			return nil, fmt.Errorf("failed to fetch URL: %w", fetchErr)
		}
		content = fetchResult.Content
		title = fetchResult.Title
	} else {
		// Direct content
		source = req.Source
		sourceType = "content"
		content = req.Content
	}

	// Check if already indexed
	if !req.Reindex {
		exists, err := s.db.DocumentExists(source)
		if err != nil {
			return nil, fmt.Errorf("failed to check if document exists: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("document already indexed: %s (use reindex=true to force re-indexing)", source)
		}
	}

	// Chunk content
	chunks := s.chunker.ChunkText(content)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks generated from content (is the content empty?)")
	}

	// Extract chunk texts for embedding
	chunkTexts := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkTexts[i] = chunk.Content
	}

	// Embed all chunks
	chunkEmbeddings, err := s.embeddingClient.Embed(ctx, chunkTexts)
	if err != nil {
		return nil, fmt.Errorf("failed to embed chunks: %w", err)
	}

	if len(chunkEmbeddings) != len(chunks) {
		return nil, fmt.Errorf("embedding count mismatch: got %d embeddings for %d chunks", len(chunkEmbeddings), len(chunks))
	}

	// Create storage chunks
	storageChunks := make([]storage.Chunk, len(chunks))
	for i, chunk := range chunks {
		storageChunks[i] = storage.Chunk{
			ChunkIndex:  chunk.Index,
			Content:     chunk.Content,
			StartOffset: chunk.StartOffset,
			EndOffset:   chunk.EndOffset,
			Embedding:   chunkEmbeddings[i],
		}
	}

	// Store in database
	err = s.db.IndexDocument(source, sourceType, title, storageChunks)
	if err != nil {
		return nil, fmt.Errorf("failed to store document: %w", err)
	}

	return &IndexResponse{
		Source:     source,
		SourceType: sourceType,
		ChunkCount: len(chunks),
		Message:    fmt.Sprintf("Successfully indexed %s (%d chunks)", source, len(chunks)),
	}, nil
}

// List returns all indexed documents
func (s *Service) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	documents, err := s.db.ListDocuments(req.SourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	docInfos := make([]DocumentInfo, len(documents))
	for i, doc := range documents {
		docInfos[i] = DocumentInfo{
			Source:      doc.Source,
			SourceType:  doc.SourceType,
			ChunkCount:  doc.ChunkCount,
			ContentSize: doc.ContentSize,
			IndexedAt:   doc.IndexedAt.Format("2006-01-02 15:04:05"),
			Title:       doc.Title,
		}
	}

	return &ListResponse{
		Documents: docInfos,
		Count:     len(docInfos),
	}, nil
}

// Delete removes an indexed document
func (s *Service) Delete(ctx context.Context, req DeleteRequest) (*DeleteResponse, error) {
	err := s.db.DeleteDocument(req.Source)
	if err != nil {
		return &DeleteResponse{
			Source:  req.Source,
			Deleted: false,
			Message: fmt.Sprintf("Failed to delete: %v", err),
		}, nil
	}

	return &DeleteResponse{
		Source:  req.Source,
		Deleted: true,
		Message: fmt.Sprintf("Successfully deleted %s", req.Source),
	}, nil
}
