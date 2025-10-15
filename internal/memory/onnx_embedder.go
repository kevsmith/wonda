package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daulet/tokenizers"
	ort "github.com/yalue/onnxruntime_go"
)

// ONNXEmbedder implements Embedder using ONNX Runtime for in-process embeddings.
// This embedder uses the gtr-t5-base model exported to ONNX format for vec2text compatibility.
type ONNXEmbedder struct {
	tokenizer      *tokenizers.Tokenizer
	modelPath      string
	sessionOptions *ort.SessionOptions
	dimensions     int
	maxLength      int
}

// NewONNXEmbedderWithDownload creates a new ONNX embedder, downloading the model if needed.
// cacheDir is typically ~/.config/wonda/models/
// If modelURL is empty, uses the default download URL.
func NewONNXEmbedderWithDownload(cacheDir, modelURL string) (*ONNXEmbedder, error) {
	downloader := NewModelDownloader(cacheDir, modelURL)
	modelDir, err := downloader.EnsureModelAvailable()
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return NewONNXEmbedder(modelDir)
}

// NewONNXEmbedder creates a new ONNX embedder.
// modelDir should point to the directory containing model.onnx and tokenizer.json files.
func NewONNXEmbedder(modelDir string) (*ONNXEmbedder, error) {
	modelPath := filepath.Join(modelDir, "model.onnx")
	tokenizerPath := filepath.Join(modelDir, "tokenizer.json")

	// Check if files exist
	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("model file not found at %s: %w", modelPath, err)
	}
	if _, err := os.Stat(tokenizerPath); err != nil {
		return nil, fmt.Errorf("tokenizer file not found at %s: %w", tokenizerPath, err)
	}

	// Load tokenizer
	tok, err := tokenizers.FromFile(tokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokenizer: %w", err)
	}

	// Initialize ONNX Runtime
	if err := ort.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX Runtime: %w", err)
	}

	// Create session options (we'll reuse these)
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create session options: %w", err)
	}

	// Set to use CPU execution provider (can add GPU later)
	// options.AppendExecutionProviderCUDA(0) // For GPU support

	return &ONNXEmbedder{
		tokenizer:      tok,
		modelPath:      modelPath,
		sessionOptions: options,
		dimensions:     768, // gtr-t5-base has 768 dimensions
		maxLength:      512, // T5 max sequence length
	}, nil
}

// Embed generates an embedding vector for the given text.
func (e *ONNXEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Tokenize the text
	inputIDs, _ := e.tokenizer.Encode(text, true) // true = add special tokens

	// Truncate if necessary
	if len(inputIDs) > e.maxLength {
		inputIDs = inputIDs[:e.maxLength]
	}

	// Get sequence length
	seqLen := len(inputIDs)

	// Create attention mask (all 1s for valid tokens)
	attentionMask := make([]uint32, seqLen)
	for i := range attentionMask {
		attentionMask[i] = 1
	}

	// Convert to int64 for ONNX input
	inputIDsInt64 := make([]int64, seqLen)
	attentionMaskInt64 := make([]int64, seqLen)
	for i := range inputIDs {
		inputIDsInt64[i] = int64(inputIDs[i])
		attentionMaskInt64[i] = int64(attentionMask[i])
	}

	// Create input tensors
	inputShape := ort.NewShape(1, int64(seqLen))
	inputIDsTensor, err := ort.NewTensor(inputShape, inputIDsInt64)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()

	attentionMaskTensor, err := ort.NewTensor(inputShape, attentionMaskInt64)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()

	// Create output tensor (allocate empty)
	outputShape := ort.NewShape(1, int64(seqLen), int64(e.dimensions))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	// Create session with these specific tensors
	session, err := ort.NewAdvancedSession(
		e.modelPath,
		[]string{"input_ids", "attention_mask"},
		[]string{"last_hidden_state"},
		[]ort.Value{inputIDsTensor, attentionMaskTensor},
		[]ort.Value{outputTensor},
		e.sessionOptions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ONNX session: %w", err)
	}
	defer session.Destroy()

	// Run inference
	if err := session.Run(); err != nil {
		return nil, fmt.Errorf("failed to run ONNX inference: %w", err)
	}

	// Get output data
	outputData := outputTensor.GetData()

	// Apply mean pooling over sequence dimension
	// Output shape is [1, seq_len, 768]
	embedding := make([]float32, e.dimensions)
	validTokens := float32(seqLen) // all tokens are valid (no padding in our case)

	for i := 0; i < seqLen; i++ {
		for j := 0; j < e.dimensions; j++ {
			idx := i*e.dimensions + j
			embedding[j] += outputData[idx]
		}
	}

	// Divide by number of valid tokens for mean
	for j := 0; j < e.dimensions; j++ {
		embedding[j] /= validTokens
	}

	return embedding, nil
}

// Destroy cleans up resources.
func (e *ONNXEmbedder) Destroy() error {
	// Close tokenizer
	if e.tokenizer != nil {
		e.tokenizer.Close()
	}

	// Destroy session options
	if e.sessionOptions != nil {
		if err := e.sessionOptions.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy session options: %w", err)
		}
	}

	return ort.DestroyEnvironment()
}

// Dimensions returns the embedding vector size.
func (e *ONNXEmbedder) Dimensions() int {
	return e.dimensions
}
