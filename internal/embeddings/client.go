package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	openAIEmbeddingsURL = "https://api.openai.com/v1/embeddings"
	embeddingModel      = "text-embedding-3-small"
	maxBatchSize        = 100
	embeddingDimension  = 1536
)

// Client handles OpenAI embeddings API calls
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new embeddings client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// embeddingRequest represents the OpenAI API request
type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// embeddingResponse represents the OpenAI API response
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *apiError `json:"error,omitempty"`
}

// apiError represents an error from the OpenAI API
type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Embed generates embeddings for multiple texts
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	var allEmbeddings [][]float32

	// Process in batches
	for i := 0; i < len(texts); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := c.embedBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to embed batch [%d:%d]: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)

		// Add delay between batches to avoid rate limiting
		if end < len(texts) {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return allEmbeddings, nil
}

// embedBatch generates embeddings for a single batch
func (c *Client) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	// Create request payload
	reqBody := embeddingRequest{
		Input: texts,
		Model: embeddingModel,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", openAIEmbeddingsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Execute request with retry on rate limit
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		// Check for rate limit error
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt < maxRetries-1 {
				// Exponential backoff
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		}
		break
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if embResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", embResp.Error.Message, embResp.Error.Type)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s, body: %s", resp.StatusCode, resp.Status, string(body))
	}

	// Extract embeddings in correct order
	embeddings := make([][]float32, len(texts))
	for _, item := range embResp.Data {
		if item.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index: %d", item.Index)
		}
		if len(item.Embedding) != embeddingDimension {
			return nil, fmt.Errorf("unexpected embedding dimension: got %d, expected %d", len(item.Embedding), embeddingDimension)
		}
		embeddings[item.Index] = item.Embedding
	}

	// Verify all embeddings were received
	for i, emb := range embeddings {
		if emb == nil {
			return nil, fmt.Errorf("missing embedding for index %d", i)
		}
	}

	return embeddings, nil
}
