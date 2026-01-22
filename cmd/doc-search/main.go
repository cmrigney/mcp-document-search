package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/cmrigney/mcp-document-search/internal/chunker"
	"github.com/cmrigney/mcp-document-search/internal/config"
	"github.com/cmrigney/mcp-document-search/internal/embeddings"
	"github.com/cmrigney/mcp-document-search/internal/fetcher"
	"github.com/cmrigney/mcp-document-search/internal/search"
	"github.com/cmrigney/mcp-document-search/internal/storage"
	"github.com/cmrigney/mcp-document-search/pkg/server"
)

func main() {
	// Enable sqlite-vec for all future database connections
	sqlite_vec.Auto()
	log.Println("sqlite-vec enabled (statically linked)")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize database
	db, err := storage.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized at: %s", cfg.DBPath)

	// Initialize embeddings client
	embClient := embeddings.NewClient(cfg.OpenAIAPIKey)
	log.Println("OpenAI embeddings client initialized")

	// Initialize chunker
	c := chunker.NewChunker(cfg.ChunkSize, cfg.Overlap)
	log.Printf("Chunker initialized (size: %d, overlap: %d)", cfg.ChunkSize, cfg.Overlap)

	// Initialize URL fetcher
	f := fetcher.NewFetcher()
	log.Println("URL fetcher initialized")

	// Initialize search service
	searchService := search.NewService(db, embClient, c, f)
	log.Println("Search service initialized")

	// Create and start MCP server
	mcpServer := server.NewServer(searchService)
	defer mcpServer.Close()

	log.Println("Starting MCP server on stdio...")

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- mcpServer.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}

	log.Println("Server shutdown complete")
}
