package simulations

import (
	"encoding/json"
	"strings"
)

// NoOpParser is a ResponseParser that returns the entire response as the message
// with no thinking extraction.
type NoOpParser struct{}

// Parse returns the entire response as the message with empty thinking.
func (p *NoOpParser) Parse(response string) (message string, thinking string) {
	return response, ""
}

// InBandParser extracts thinking from response text using start/end delimiters.
type InBandParser struct {
	startDelim string
	endDelim   string
}

// NewInBandParser creates a new InBandParser with the specified delimiters.
func NewInBandParser(startDelim, endDelim string) *InBandParser {
	return &InBandParser{
		startDelim: startDelim,
		endDelim:   endDelim,
	}
}

// Parse extracts thinking content between delimiters and returns the remaining text as message.
// If multiple thinking blocks exist, they are concatenated with newlines.
// If no thinking blocks are found, returns the entire response as message.
func (p *InBandParser) Parse(response string) (message string, thinking string) {
	var thinkingBlocks []string
	remaining := response

	for {
		startIdx := strings.Index(remaining, p.startDelim)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(remaining[startIdx:], p.endDelim)
		if endIdx == -1 {
			// Found start delimiter but no matching end - treat rest as message
			break
		}
		endIdx += startIdx

		// Extract thinking content (excluding delimiters)
		thinkingContent := remaining[startIdx+len(p.startDelim) : endIdx]
		thinkingBlocks = append(thinkingBlocks, strings.TrimSpace(thinkingContent))

		// Remove this thinking block from the response
		remaining = remaining[:startIdx] + remaining[endIdx+len(p.endDelim):]
	}

	message = strings.TrimSpace(remaining)
	if len(thinkingBlocks) > 0 {
		thinking = strings.Join(thinkingBlocks, "\n\n")
	}

	return message, thinking
}

// OutOfBandParser is a ResponseParser for models that return thinking in a separate
// API response field. This parser is a pass-through for the response text, as the
// actual thinking extraction is handled by the client implementation when parsing JSON.
type OutOfBandParser struct {
	fieldPath string
}

// NewOutOfBandParser creates a new OutOfBandParser with the specified field path.
func NewOutOfBandParser(fieldPath string) *OutOfBandParser {
	return &OutOfBandParser{
		fieldPath: fieldPath,
	}
}

// Parse is a pass-through that returns the response as-is.
// Actual thinking extraction happens at the client level when parsing JSON responses.
func (p *OutOfBandParser) Parse(response string) (message string, thinking string) {
	return response, ""
}

// FieldPath returns the JSON field path for extracting thinking.
// This is used by client implementations to know which field to extract.
func (p *OutOfBandParser) FieldPath() string {
	return p.fieldPath
}

// extractJSONField extracts a field value from a JSON object using a dot-notation path.
// For example, "reasoning.summary" will extract obj["reasoning"]["summary"].
// Returns empty string if the field doesn't exist or isn't a string.
func extractJSONField(jsonData []byte, fieldPath string) string {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return ""
	}

	// Split the path and traverse the object
	parts := strings.Split(fieldPath, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return ""
		}

		// If this is the last part, extract the string value
		if i == len(parts)-1 {
			if str, ok := value.(string); ok {
				return str
			}
			return ""
		}

		// Otherwise, expect a nested object
		if nested, ok := value.(map[string]interface{}); ok {
			current = nested
		} else {
			return ""
		}
	}

	return ""
}
