package simulations

import (
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
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

// extractJSONField extracts a field value from a JSON object using JSONPath.
// Supports array indexing: "choices[0].message.reasoning" or "choices.0.message.reasoning"
// Returns empty string if the field doesn't exist or isn't a string.
func extractJSONField(jsonData []byte, fieldPath string) string {
	// Parse JSON data
	obj, err := oj.Parse(jsonData)
	if err != nil {
		return ""
	}

	// Convert dot-notation to JSONPath format
	// "choices.0.message.reasoning" -> "$.choices[0].message.reasoning"
	jsonPath := convertToJSONPath(fieldPath)

	// Parse and execute JSONPath
	x, err := jp.ParseString(jsonPath)
	if err != nil {
		return ""
	}

	results := x.Get(obj)
	if len(results) == 0 {
		return ""
	}

	// Return first result as string
	if str, ok := results[0].(string); ok {
		return str
	}

	return ""
}

// convertToJSONPath converts dot-notation paths to JSONPath format.
// Examples:
//
//	"choices.0.message.reasoning" -> "$.choices[0].message.reasoning"
//	"reasoning.content" -> "$.reasoning.content"
func convertToJSONPath(path string) string {
	if !strings.HasPrefix(path, "$") {
		path = "$." + path
	}

	// Convert numeric indices: .0. or .0 at end -> [0]
	parts := strings.Split(path, ".")
	var result strings.Builder

	for i, part := range parts {
		if i == 0 {
			result.WriteString(part) // Write "$"
			continue
		}

		// Check if this part is numeric
		if len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
			result.WriteString("[")
			result.WriteString(part)
			result.WriteString("]")
		} else {
			if i > 1 || (i == 1 && parts[0] == "$") {
				result.WriteString(".")
			}
			result.WriteString(part)
		}
	}

	return result.String()
}
