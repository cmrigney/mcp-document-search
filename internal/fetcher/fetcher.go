package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// FetchResult contains the fetched content and metadata
type FetchResult struct {
	Content string
	Title   string
	Size    int
}

// Fetcher handles fetching and parsing URL content
type Fetcher struct {
	httpClient *http.Client
}

// NewFetcher creates a new fetcher with a 30-second timeout
func NewFetcher() *Fetcher {
	return &Fetcher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchURL fetches content from a URL and extracts text
func (f *Fetcher) FetchURL(ctx context.Context, urlStr string) (*FetchResult, error) {
	// Validate URL format
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (must be http or https)", parsedURL.Scheme)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent
	req.Header.Set("User-Agent", "MCP-DocSearch/1.0")

	// Execute request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !isTextContent(contentType) {
		return nil, fmt.Errorf("unsupported content type: %s (must be text/plain, text/html, or text/markdown)", contentType)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	content := string(body)
	title := ""

	// Extract text from HTML
	if strings.Contains(contentType, "text/html") {
		var extractErr error
		content, title, extractErr = extractHTMLText(content)
		if extractErr != nil {
			return nil, fmt.Errorf("failed to extract HTML text: %w", extractErr)
		}
	}

	return &FetchResult{
		Content: content,
		Title:   title,
		Size:    len(content),
	}, nil
}

// isTextContent checks if the content type is text-based
func isTextContent(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/markdown") ||
		strings.Contains(contentType, "text/")
}

// extractHTMLText extracts text content and title from HTML
func extractHTMLText(htmlContent string) (text string, title string, err error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", err
	}

	var textBuilder strings.Builder
	var titleText string

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		// Skip script and style tags
		if n.Type == html.ElementNode {
			if n.Data == "script" || n.Data == "style" {
				return
			}
			// Extract title
			if n.Data == "title" && titleText == "" {
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					titleText = strings.TrimSpace(n.FirstChild.Data)
				}
			}
		}

		// Extract text nodes
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				if textBuilder.Len() > 0 {
					textBuilder.WriteString(" ")
				}
				textBuilder.WriteString(text)
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Normalize whitespace
	extractedText := textBuilder.String()
	extractedText = normalizeWhitespace(extractedText)

	return extractedText, titleText, nil
}

// normalizeWhitespace replaces multiple spaces with single space
func normalizeWhitespace(text string) string {
	// Split by whitespace and rejoin with single spaces
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}
