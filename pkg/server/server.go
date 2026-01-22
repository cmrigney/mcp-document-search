package server

import (
	"context"
	"fmt"
	"log"

	"github.com/cmrigney/mcp-document-search/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the MCP server
type Server struct {
	mcpServer     *mcp.Server
	searchService *search.Service
}

// NewServer creates a new MCP server
func NewServer(searchService *search.Service) *Server {
	s := &Server{
		searchService: searchService,
	}

	// Create MCP server implementation
	impl := &mcp.Implementation{
		Name:    "doc-search",
		Version: "1.0.0",
	}

	// Create the MCP server
	mcpServer := mcp.NewServer(impl, nil)

	// Register tools
	s.registerTools(mcpServer)

	s.mcpServer = mcpServer
	return s
}

// registerTools registers all MCP tools
func (s *Server) registerTools(mcpServer *mcp.Server) {
	// Search tool
	searchTool := &mcp.Tool{
		Name:        "search",
		Description: "Semantic search through indexed documents using vector similarity",
	}
	mcp.AddTool(mcpServer, searchTool, s.handleSearch)

	// Index tool
	indexTool := &mcp.Tool{
		Name:        "index",
		Description: "Index a file, URL, or content for semantic search. Provide exactly one of: file_path, url, or (content + source)",
	}
	mcp.AddTool(mcpServer, indexTool, s.handleIndex)

	// List tool
	listTool := &mcp.Tool{
		Name:        "list",
		Description: "List all indexed documents with metadata",
	}
	mcp.AddTool(mcpServer, listTool, s.handleList)

	// Delete tool
	deleteTool := &mcp.Tool{
		Name:        "delete",
		Description: "Remove an indexed document from the database",
	}
	mcp.AddTool(mcpServer, deleteTool, s.handleDelete)
}

// Run starts the MCP server on stdio transport
func (s *Server) Run(ctx context.Context) error {
	transport := &mcp.StdioTransport{}

	log.Println("MCP Doc Search server started on stdio")

	// Run the server
	if err := s.mcpServer.Run(ctx, transport); err != nil {
		return fmt.Errorf("failed to run MCP server: %w", err)
	}

	return nil
}

// Close cleans up server resources
func (s *Server) Close() error {
	// No explicit close needed for the new SDK
	return nil
}
