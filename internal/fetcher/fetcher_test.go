package fetcher

import (
	"context"
	"testing"
)

func TestExtractHTMLText(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		expectedText  string
		expectedTitle string
	}{
		{
			name: "simple HTML",
			html: `<html><head><title>Test Page</title></head><body><p>Hello World</p></body></html>`,
			expectedText:  "Test Page Hello World",
			expectedTitle: "Test Page",
		},
		{
			name: "HTML with script and style",
			html: `<html><head><title>Page</title><script>alert('test');</script><style>body{color:red;}</style></head><body><p>Content</p></body></html>`,
			expectedText:  "Page Content",
			expectedTitle: "Page",
		},
		{
			name: "HTML with multiple paragraphs",
			html: `<html><body><p>First paragraph</p><p>Second paragraph</p></body></html>`,
			expectedText:  "First paragraph Second paragraph",
			expectedTitle: "",
		},
		{
			name: "HTML with extra whitespace",
			html: `<html><body><p>Text   with    extra     spaces</p></body></html>`,
			expectedText:  "Text with extra spaces",
			expectedTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, title, err := extractHTMLText(tt.html)
			if err != nil {
				t.Fatalf("extractHTMLText failed: %v", err)
			}
			if text != tt.expectedText {
				t.Errorf("Expected text %q, got %q", tt.expectedText, text)
			}
			if title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello  world", "hello world"},
		{"hello\n\nworld", "hello world"},
		{"  hello   world  ", "hello world"},
		{"hello\tworld", "hello world"},
	}

	for _, tt := range tests {
		result := normalizeWhitespace(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeWhitespace(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsTextContent(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"text/html; charset=utf-8", true},
		{"text/plain", true},
		{"text/markdown", true},
		{"application/json", false},
		{"image/png", false},
		{"text/html", true},
		{"TEXT/HTML", true}, // Case insensitive
	}

	for _, tt := range tests {
		result := isTextContent(tt.contentType)
		if result != tt.expected {
			t.Errorf("isTextContent(%q) = %v, expected %v", tt.contentType, result, tt.expected)
		}
	}
}

func TestFetcherInvalidURL(t *testing.T) {
	f := NewFetcher()
	ctx := context.Background()

	// Test invalid URL
	_, err := f.FetchURL(ctx, "not a valid url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Test unsupported scheme
	_, err = f.FetchURL(ctx, "ftp://example.com")
	if err == nil {
		t.Error("Expected error for unsupported scheme")
	}
}
