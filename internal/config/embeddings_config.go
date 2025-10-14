package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Embedding represents a single embedding model configuration.
type Embedding struct {
	Name       string `toml:"-"`       // Set from map key
	Provider   string `toml:"provider"` // References a provider name from [providers.*]
	Model      string `toml:"model"`
	Dimensions int    `toml:"dimensions"`
}

// Validate checks if the embedding configuration is valid.
func (e *Embedding) Validate() error {
	if e.Provider == "" {
		return fmt.Errorf("embedding '%s': provider is required", e.Name)
	}
	if e.Model == "" {
		return fmt.Errorf("embedding '%s': model is required", e.Name)
	}
	if e.Dimensions <= 0 {
		return fmt.Errorf("embedding '%s': dimensions must be positive", e.Name)
	}
	// Common embedding dimensions (sanity check)
	validDimensions := map[int]bool{
		384:  true, // sentence-transformers/all-MiniLM-L6-v2
		512:  true, // Various smaller models
		768:  true, // nomic-embed-text, BERT-base
		1024: true, // Various models
		1536: true, // OpenAI text-embedding-3-small, text-embedding-ada-002
		3072: true, // OpenAI text-embedding-3-large
	}
	if !validDimensions[e.Dimensions] {
		return fmt.Errorf("embedding '%s': unusual dimensions %d (common values: 384, 512, 768, 1024, 1536, 3072)",
			e.Name, e.Dimensions)
	}
	return nil
}

// Embeddings represents the top-level embeddings configuration.
type Embeddings struct {
	Embeddings map[string]*Embedding `toml:"embeddings"`
}

// NewEmbeddings creates an empty Embeddings configuration.
func NewEmbeddings() *Embeddings {
	return &Embeddings{
		Embeddings: make(map[string]*Embedding),
	}
}

// LoadEmbeddings creates and populates an Embeddings configuration from TOML.
func LoadEmbeddings(data []byte) (*Embeddings, error) {
	e := NewEmbeddings()
	if err := toml.Unmarshal(data, e); err != nil {
		return nil, err
	}
	for name, embedding := range e.Embeddings {
		embedding.Name = name
		if err := embedding.Validate(); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// LoadEmbeddingsFromFile loads embeddings configuration from a file path.
func LoadEmbeddingsFromFile(path string) (*Embeddings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadEmbeddings(data)
}

// Get retrieves an embedding by name, returning an error if not found.
func (e *Embeddings) Get(name string) (*Embedding, error) {
	if embedding, ok := e.Embeddings[name]; ok {
		return embedding, nil
	}
	return nil, fmt.Errorf("embedding '%s' not found", name)
}
