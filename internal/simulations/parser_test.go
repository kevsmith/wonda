package simulations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpParser(t *testing.T) {
	parser := &NoOpParser{}

	t.Run("returns entire response as message", func(t *testing.T) {
		response := "This is a test response"
		message, thinking := parser.Parse(response)
		assert.Equal(t, "This is a test response", message)
		assert.Equal(t, "", thinking)
	})

	t.Run("returns empty string for empty response", func(t *testing.T) {
		message, thinking := parser.Parse("")
		assert.Equal(t, "", message)
		assert.Equal(t, "", thinking)
	})
}

func TestInBandParser(t *testing.T) {
	parser := NewInBandParser("<think>", "</think>")

	t.Run("extracts single thinking block", func(t *testing.T) {
		response := "Let me think about this. <think>This is my reasoning</think> So the answer is 42."
		message, thinking := parser.Parse(response)
		assert.Equal(t, "Let me think about this.  So the answer is 42.", message)
		assert.Equal(t, "This is my reasoning", thinking)
	})

	t.Run("extracts multiple thinking blocks", func(t *testing.T) {
		response := "<think>First thought</think> Some text. <think>Second thought</think> More text."
		message, thinking := parser.Parse(response)
		assert.Equal(t, "Some text.  More text.", message)
		assert.Equal(t, "First thought\n\nSecond thought", thinking)
	})

	t.Run("handles response with no thinking blocks", func(t *testing.T) {
		response := "Just a regular response with no thinking."
		message, thinking := parser.Parse(response)
		assert.Equal(t, "Just a regular response with no thinking.", message)
		assert.Equal(t, "", thinking)
	})

	t.Run("handles thinking at start of response", func(t *testing.T) {
		response := "<think>Initial reasoning</think>The answer is yes."
		message, thinking := parser.Parse(response)
		assert.Equal(t, "The answer is yes.", message)
		assert.Equal(t, "Initial reasoning", thinking)
	})

	t.Run("handles thinking at end of response", func(t *testing.T) {
		response := "The answer is no.<think>Because of reasons</think>"
		message, thinking := parser.Parse(response)
		assert.Equal(t, "The answer is no.", message)
		assert.Equal(t, "Because of reasons", thinking)
	})

	t.Run("handles only thinking block", func(t *testing.T) {
		response := "<think>Pure reasoning with no other text</think>"
		message, thinking := parser.Parse(response)
		assert.Equal(t, "", message)
		assert.Equal(t, "Pure reasoning with no other text", thinking)
	})

	t.Run("handles unmatched start delimiter", func(t *testing.T) {
		response := "Text before <think> some thinking without end"
		message, thinking := parser.Parse(response)
		assert.Equal(t, "Text before <think> some thinking without end", message)
		assert.Equal(t, "", thinking)
	})

	t.Run("handles nested delimiters", func(t *testing.T) {
		response := "<think>Outer <think>inner</think> more outer</think> Final text."
		message, thinking := parser.Parse(response)
		// Should extract first matching pair
		assert.Contains(t, thinking, "Outer")
		assert.Contains(t, message, "Final text")
	})

	t.Run("trims whitespace from thinking and message", func(t *testing.T) {
		response := "  <think>  thinking with spaces  </think>  message with spaces  "
		message, thinking := parser.Parse(response)
		assert.Equal(t, "thinking with spaces", thinking)
		assert.Equal(t, "message with spaces", message)
	})

	t.Run("works with custom delimiters", func(t *testing.T) {
		parser := NewInBandParser("<reasoning>", "</reasoning>")
		response := "Text <reasoning>My reasoning</reasoning> More text."
		message, thinking := parser.Parse(response)
		assert.Equal(t, "Text  More text.", message)
		assert.Equal(t, "My reasoning", thinking)
	})
}

func TestOutOfBandParser(t *testing.T) {
	parser := NewOutOfBandParser("thinking")

	t.Run("returns response as-is", func(t *testing.T) {
		response := "This is the response text"
		message, thinking := parser.Parse(response)
		assert.Equal(t, "This is the response text", message)
		assert.Equal(t, "", thinking)
	})

	t.Run("has correct field path", func(t *testing.T) {
		parser := NewOutOfBandParser("reasoning.summary")
		assert.Equal(t, "reasoning.summary", parser.FieldPath())
	})
}

func TestExtractJSONField(t *testing.T) {
	t.Run("extracts top-level string field", func(t *testing.T) {
		jsonData := []byte(`{"thinking": "My thoughts", "message": "Hello"}`)
		value := extractJSONField(jsonData, "thinking")
		assert.Equal(t, "My thoughts", value)
	})

	t.Run("extracts nested string field", func(t *testing.T) {
		jsonData := []byte(`{"reasoning": {"summary": "My reasoning", "effort": "high"}}`)
		value := extractJSONField(jsonData, "reasoning.summary")
		assert.Equal(t, "My reasoning", value)
	})

	t.Run("returns empty for non-existent field", func(t *testing.T) {
		jsonData := []byte(`{"message": "Hello"}`)
		value := extractJSONField(jsonData, "thinking")
		assert.Equal(t, "", value)
	})

	t.Run("returns empty for non-existent nested field", func(t *testing.T) {
		jsonData := []byte(`{"reasoning": {"effort": "high"}}`)
		value := extractJSONField(jsonData, "reasoning.summary")
		assert.Equal(t, "", value)
	})

	t.Run("returns empty for non-string field", func(t *testing.T) {
		jsonData := []byte(`{"count": 42}`)
		value := extractJSONField(jsonData, "count")
		assert.Equal(t, "", value)
	})

	t.Run("returns empty for invalid JSON", func(t *testing.T) {
		jsonData := []byte(`{invalid json}`)
		value := extractJSONField(jsonData, "thinking")
		assert.Equal(t, "", value)
	})

	t.Run("returns empty when intermediate path is not object", func(t *testing.T) {
		jsonData := []byte(`{"reasoning": "string value"}`)
		value := extractJSONField(jsonData, "reasoning.summary")
		assert.Equal(t, "", value)
	})

	t.Run("handles deeply nested paths", func(t *testing.T) {
		jsonData := []byte(`{"a": {"b": {"c": {"d": "deep value"}}}}`)
		value := extractJSONField(jsonData, "a.b.c.d")
		assert.Equal(t, "deep value", value)
	})
}
