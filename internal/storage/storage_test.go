package storage

import (
	"testing"
)

func TestSerializeDeserializeEmbedding(t *testing.T) {
	// Create a test embedding
	original := make([]float32, embeddingDimension)
	for i := range original {
		original[i] = float32(i) * 0.001
	}

	// Serialize
	blob, err := serializeEmbedding(original)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Check blob size (1536 floats * 4 bytes = 6144 bytes)
	expectedSize := embeddingDimension * 4
	if len(blob) != expectedSize {
		t.Errorf("Expected blob size %d, got %d", expectedSize, len(blob))
	}

	// Deserialize
	deserialized, err := deserializeEmbedding(blob)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	// Compare
	if len(deserialized) != len(original) {
		t.Errorf("Expected %d dimensions, got %d", len(original), len(deserialized))
	}

	for i := range original {
		if deserialized[i] != original[i] {
			t.Errorf("Mismatch at index %d: expected %f, got %f", i, original[i], deserialized[i])
			break
		}
	}
}

func TestRunMigrations(t *testing.T) {
	// Note: This test would require sqlite-vec to be installed
	// For CI/CD, you might want to skip this or use a mock
	t.Skip("Skipping database test - requires sqlite-vec installation")
}

// Additional integration tests would go here
// They would require sqlite-vec to be properly installed
