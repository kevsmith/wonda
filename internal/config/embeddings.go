package config

// Embedding model configuration
const (
	// RequiredEmbeddingModel is the embedding model required for memory operations.
	// Currently using nomic-embed-text for MVP (available via LM Studio).
	// NOTE: This model is NOT vec2text compatible. For cognitive distortion features,
	// we'll need to switch to sentence-transformers/gtr-t5-base in the future.
	RequiredEmbeddingModel = "nomic-ai/nomic-embed-text-v1.5-GGUF"

	// RequiredEmbeddingDimensions is the expected dimensionality of embedding vectors.
	RequiredEmbeddingDimensions = 768

	// EmbeddingMaxTokens is the maximum token length for embedding input.
	EmbeddingMaxTokens = 8192 // nomic-embed-text supports 8192 tokens
)
