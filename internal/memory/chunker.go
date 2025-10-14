package memory

import (
	"strings"
	"unicode"
)

// ChunkText splits text into smaller chunks suitable for embedding.
// Uses sentence-based chunking with a maximum character limit.
func ChunkText(text string, maxChars int) []string {
	if text == "" {
		return []string{}
	}

	// Trim whitespace
	text = strings.TrimSpace(text)

	// If text is shorter than max, return as single chunk
	if len(text) <= maxChars {
		return []string{text}
	}

	// Split into sentences
	sentences := splitSentences(text)

	chunks := make([]string, 0)
	currentChunk := ""

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// If adding this sentence would exceed max, start new chunk
		if len(currentChunk)+len(sentence)+1 > maxChars && currentChunk != "" {
			chunks = append(chunks, currentChunk)
			currentChunk = sentence
		} else {
			// Add to current chunk
			if currentChunk != "" {
				currentChunk += " "
			}
			currentChunk += sentence
		}
	}

	// Add final chunk
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// splitSentences splits text into sentences based on punctuation.
func splitSentences(text string) []string {
	sentences := make([]string, 0)
	current := ""

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		current += string(r)

		// Check for sentence-ending punctuation
		if r == '.' || r == '!' || r == '?' {
			// Look ahead to see if this is really end of sentence
			if i+1 < len(runes) {
				next := runes[i+1]
				// If followed by whitespace and capital letter, it's likely a sentence end
				if unicode.IsSpace(next) {
					if i+2 < len(runes) && unicode.IsUpper(runes[i+2]) {
						sentences = append(sentences, strings.TrimSpace(current))
						current = ""
						continue
					}
				}
			} else {
				// End of text
				sentences = append(sentences, strings.TrimSpace(current))
				current = ""
			}
		}
	}

	// Add any remaining text
	if current != "" {
		sentences = append(sentences, strings.TrimSpace(current))
	}

	return sentences
}

// ChunkByLines splits text by newlines, respecting max character limit.
// Useful for structured text like bullet points.
func ChunkByLines(text string, maxChars int) []string {
	if text == "" {
		return []string{}
	}

	lines := strings.Split(text, "\n")
	chunks := make([]string, 0)
	currentChunk := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// If adding this line would exceed max, start new chunk
		if len(currentChunk)+len(line)+1 > maxChars && currentChunk != "" {
			chunks = append(chunks, currentChunk)
			currentChunk = line
		} else {
			// Add to current chunk
			if currentChunk != "" {
				currentChunk += "\n"
			}
			currentChunk += line
		}
	}

	// Add final chunk
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
