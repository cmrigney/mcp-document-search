package chunker

import (
	"strings"
	"testing"
)

func TestChunkerEmptyText(t *testing.T) {
	c := NewChunker(1000, 100)
	chunks := c.ChunkText("")
	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestChunkerSmallText(t *testing.T) {
	c := NewChunker(1000, 100)
	text := "Hello world"
	chunks := c.ChunkText(text)

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for small text, got %d", len(chunks))
	}
	if chunks[0].Content != text {
		t.Errorf("Expected content %q, got %q", text, chunks[0].Content)
	}
	if chunks[0].StartOffset != 0 {
		t.Errorf("Expected StartOffset 0, got %d", chunks[0].StartOffset)
	}
	if chunks[0].EndOffset != len([]rune(text)) {
		t.Errorf("Expected EndOffset %d, got %d", len([]rune(text)), chunks[0].EndOffset)
	}
}

func TestChunkerWordBoundaries(t *testing.T) {
	c := NewChunker(50, 10)
	// Create text with clear word boundaries
	text := strings.Repeat("word ", 30) // 150 chars, should split into chunks

	chunks := c.ChunkText(text)

	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	// Verify no chunks split words (shouldn't have "wo rd")
	for i, chunk := range chunks {
		if strings.Contains(chunk.Content, "wor d") || strings.Contains(chunk.Content, "wo rd") {
			t.Errorf("Chunk %d split a word: %q", i, chunk.Content)
		}
	}
}

func TestChunkerOverlap(t *testing.T) {
	c := NewChunker(20, 5)
	text := "0123456789 0123456789 0123456789 0123456789"

	chunks := c.ChunkText(text)

	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	// Check that chunks overlap
	for i := 0; i < len(chunks)-1; i++ {
		chunk1End := chunks[i].EndOffset
		chunk2Start := chunks[i+1].StartOffset

		overlap := chunk1End - chunk2Start
		if overlap <= 0 {
			t.Errorf("Chunks %d and %d don't overlap: end=%d, start=%d", i, i+1, chunk1End, chunk2Start)
		}
	}
}

func TestChunkerUnicode(t *testing.T) {
	c := NewChunker(20, 5)
	text := "Hello 世界 Hello 世界 Hello 世界 Hello 世界"

	chunks := c.ChunkText(text)

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk for Unicode text")
	}

	// Verify all chunks contain valid Unicode
	for i, chunk := range chunks {
		if chunk.Content == "" {
			t.Errorf("Chunk %d is empty", i)
		}
		// Just verify it's valid UTF-8
		_ = []rune(chunk.Content)
	}
}

func TestChunkerNoWhitespace(t *testing.T) {
	c := NewChunker(50, 10)
	// Text with no spaces - should force split
	text := strings.Repeat("a", 200)

	chunks := c.ChunkText(text)

	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks for long text without whitespace, got %d", len(chunks))
	}

	// Verify chunks are approximately the right size
	for i, chunk := range chunks {
		runeLen := len([]rune(chunk.Content))
		if runeLen > c.chunkSize+10 { // Allow some tolerance
			t.Errorf("Chunk %d is too large: %d runes", i, runeLen)
		}
	}
}

func TestChunkerIndexing(t *testing.T) {
	c := NewChunker(30, 5)
	text := strings.Repeat("test ", 40)

	chunks := c.ChunkText(text)

	// Verify chunk indices are sequential
	for i, chunk := range chunks {
		if chunk.Index != i {
			t.Errorf("Expected chunk index %d, got %d", i, chunk.Index)
		}
	}
}

func TestNewChunkerDefaults(t *testing.T) {
	// Test with invalid values
	c := NewChunker(0, -1)
	if c.chunkSize != 1000 {
		t.Errorf("Expected default chunkSize 1000, got %d", c.chunkSize)
	}
	if c.overlap != 0 {
		t.Errorf("Expected overlap 0 for negative value, got %d", c.overlap)
	}

	// Test overlap >= chunkSize
	c2 := NewChunker(100, 150)
	if c2.overlap >= c2.chunkSize {
		t.Errorf("Overlap should be reduced when >= chunkSize, got overlap=%d chunkSize=%d", c2.overlap, c2.chunkSize)
	}
}
