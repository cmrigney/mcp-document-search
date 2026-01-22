package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/test-doc-mcp/internal/search"
)

// handleSearch handles the search tool
func (s *Server) handleSearch(ctx context.Context, request *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	// Validate query
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}

	// Set defaults
	if args.TopK <= 0 {
		args.TopK = 5
	}

	minScore := 0.3
	if args.MinScore != nil {
		minScore = *args.MinScore
	}

	// Execute search
	searchReq := search.SearchRequest{
		Query:        args.Query,
		TopK:         args.TopK,
		MinScore:     minScore,
		SourceFilter: args.SourceFilter,
	}

	resp, err := s.searchService.Search(ctx, searchReq)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}

	// Format response as JSON
	resultJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// handleIndex handles the index tool
func (s *Server) handleIndex(ctx context.Context, request *mcp.CallToolRequest, args IndexArgs) (*mcp.CallToolResult, any, error) {
	// Validate exactly one source
	hasFilePath := args.FilePath != ""
	hasURL := args.URL != ""
	hasContent := args.Content != "" && args.Source != ""

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
		return nil, nil, fmt.Errorf("must provide exactly one of: file_path, url, or (content + source)")
	}
	if sourceCount > 1 {
		return nil, nil, fmt.Errorf("provide exactly one of: file_path, url, or (content + source)")
	}

	// Execute index
	indexReq := search.IndexRequest{
		FilePath: args.FilePath,
		URL:      args.URL,
		Content:  args.Content,
		Source:   args.Source,
		Reindex:  args.Reindex,
	}

	resp, err := s.searchService.Index(ctx, indexReq)
	if err != nil {
		return nil, nil, fmt.Errorf("indexing failed: %w", err)
	}

	// Format response as JSON
	resultJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// handleList handles the list tool
func (s *Server) handleList(ctx context.Context, request *mcp.CallToolRequest, args ListArgs) (*mcp.CallToolResult, any, error) {
	// Execute list
	listReq := search.ListRequest{
		SourceType: args.SourceType,
	}

	resp, err := s.searchService.List(ctx, listReq)
	if err != nil {
		return nil, nil, fmt.Errorf("list failed: %w", err)
	}

	// Format response as JSON
	resultJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// handleDelete handles the delete tool
func (s *Server) handleDelete(ctx context.Context, request *mcp.CallToolRequest, args DeleteArgs) (*mcp.CallToolResult, any, error) {
	// Validate source
	if args.Source == "" {
		return nil, nil, fmt.Errorf("source is required")
	}

	// Execute delete
	deleteReq := search.DeleteRequest{
		Source: args.Source,
	}

	resp, err := s.searchService.Delete(ctx, deleteReq)
	if err != nil {
		return nil, nil, fmt.Errorf("delete failed: %w", err)
	}

	// Format response as JSON
	resultJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to format results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}
