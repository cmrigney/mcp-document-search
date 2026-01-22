package chunker

import (
	"unicode"
)

// Chunk represents a text chunk with its metadata
type Chunk struct {
	Content     string
	Index       int
	StartOffset int
	EndOffset   int
}

// Chunker handles text chunking with word boundary support
type Chunker struct {
	chunkSize int
	overlap   int
}

// NewChunker creates a new chunker with specified chunk size and overlap
func NewChunker(chunkSize, overlap int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 2
	}
	return &Chunker{
		chunkSize: chunkSize,
		overlap:   overlap,
	}
}

// ChunkText splits text into overlapping chunks respecting word boundaries
func (c *Chunker) ChunkText(text string) []Chunk {
	if text == "" {
		return []Chunk{}
	}

	// Convert to runes for proper Unicode handling
	runes := []rune(text)
	totalLen := len(runes)

	if totalLen <= c.chunkSize {
		// Text fits in single chunk
		return []Chunk{{
			Content:     text,
			Index:       0,
			StartOffset: 0,
			EndOffset:   totalLen,
		}}
	}

	var chunks []Chunk
	chunkIndex := 0
	position := 0

	for position < totalLen {
		// Determine chunk end position
		endPos := position + c.chunkSize
		if endPos > totalLen {
			endPos = totalLen
		}

		// Find word boundary if not at end
		actualEnd := endPos
		if endPos < totalLen {
			actualEnd = c.findWordBoundary(runes, position, endPos)
		}

		// Extract chunk content
		chunkRunes := runes[position:actualEnd]
		content := string(chunkRunes)

		chunks = append(chunks, Chunk{
			Content:     content,
			Index:       chunkIndex,
			StartOffset: position,
			EndOffset:   actualEnd,
		})

		// Move to next chunk position with overlap
		// If we're at the end, we're done
		if actualEnd >= totalLen {
			break
		}

		// Next chunk starts at (current end - overlap)
		position = actualEnd - c.overlap
		if position < 0 {
			position = 0
		}

		chunkIndex++
	}

	return chunks
}

// findWordBoundary scans backward from endPos to find the last whitespace
// within maxScanBack characters, respecting word boundaries
func (c *Chunker) findWordBoundary(runes []rune, startPos, endPos int) int {
	maxScanBack := 100
	if maxScanBack > c.chunkSize {
		maxScanBack = c.chunkSize
	}

	// Scan backward from endPos
	scanStart := endPos - maxScanBack
	if scanStart < startPos {
		scanStart = startPos
	}

	// Look for last whitespace character
	lastWhitespace := -1
	for i := endPos - 1; i >= scanStart; i-- {
		if unicode.IsSpace(runes[i]) {
			lastWhitespace = i
			break
		}
	}

	// If we found whitespace, split after it (skip the whitespace)
	if lastWhitespace >= 0 {
		// Skip consecutive whitespace
		splitPos := lastWhitespace + 1
		for splitPos < endPos && unicode.IsSpace(runes[splitPos]) {
			splitPos++
		}
		return splitPos
	}

	// No whitespace found, force split at endPos
	return endPos
}
