package embeddings

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	if client == nil {
		t.Error("Expected non-nil client")
	}
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got %q", client.apiKey)
	}
}

func TestEmbedEmptyTexts(t *testing.T) {
	client := NewClient("test-api-key")
	ctx := context.Background()

	embeddings, err := client.Embed(ctx, []string{})
	if err != nil {
		t.Errorf("Expected no error for empty texts, got %v", err)
	}
	if len(embeddings) != 0 {
		t.Errorf("Expected empty embeddings slice, got %d embeddings", len(embeddings))
	}
}

// Note: Actual API tests would require a valid API key and would make real API calls
// For integration testing, you would use a valid OPENAI_API_KEY and test against the real API
