package agent

import (
	"strings"
)

// XMLStructuredDataExtractor handles the extraction of XML structured data
// delimited by specific XML tags from a single LLM response,
// supporting basic nesting and attributes in the target tag.
type XMLStructuredDataExtractor struct {
	tagName string // The target XML tag name (e.g., "results")
}

// NewXMLStructuredDataExtractor creates a new XMLStructuredDataExtractor with the specified tag name.
func NewXMLStructuredDataExtractor(tagName string) *XMLStructuredDataExtractor {
	return &XMLStructuredDataExtractor{
		tagName: tagName,
	}
}

// Extract parses the response, extracting data between the first occurrence
// of <tagName...> and its corresponding </tagName>, respecting nesting.
// It returns a string containing the extracted data, or an empty string if
// the tag is not found or not properly closed.
func (e *XMLStructuredDataExtractor) Extract(response string) string {
	openTagStart := "<" + e.tagName
	closeTag := "</" + e.tagName + ">"

	startIdx := -1
	contentStartIdx := -1
	depth := 0
	searchFrom := 0

	for searchFrom < len(response) {
		firstOpenTagIdx := strings.Index(response[searchFrom:], openTagStart)
		firstCloseTagIdx := strings.Index(response[searchFrom:], closeTag)

		// If neither tag is found in the remaining string, break
		if firstOpenTagIdx == -1 && firstCloseTagIdx == -1 {
			break
		}

		// Adjust indices to be relative to the original response string
		if firstOpenTagIdx != -1 {
			firstOpenTagIdx += searchFrom
		}
		if firstCloseTagIdx != -1 {
			firstCloseTagIdx += searchFrom
		}

		// Determine which tag comes first
		if firstOpenTagIdx != -1 && (firstCloseTagIdx == -1 || firstOpenTagIdx < firstCloseTagIdx) {
			// Found an opening tag <tagName...>
			// Check if it's the actual tag (could be <tagNameSuffix>)
			tagEndCharIdx := firstOpenTagIdx + len(openTagStart)
			if tagEndCharIdx < len(response) {
				c := response[tagEndCharIdx]
				// Allow space or > after tag name for attributes
				if c == ' ' || c == '>' {
					if depth == 0 {
						// Found the *first* opening tag at depth 0
						startIdx = firstOpenTagIdx
						// Find the actual end of the opening tag '>'
						tagContentStartIdx := strings.Index(response[startIdx:], ">")
						if tagContentStartIdx != -1 {
							contentStartIdx = startIdx + tagContentStartIdx + 1
						} else {
							// Malformed opening tag, abort
							return ""
						}
					}
					depth++
					searchFrom = firstOpenTagIdx + len(openTagStart) // Continue search after <tagName
				} else {
					// False positive (e.g., <tagNameSuffix>), continue search after this point
					searchFrom = firstOpenTagIdx + 1
				}
			} else {
				// Tag starts but string ends, break
				break
			}
		} else if firstCloseTagIdx != -1 {
			// Found a closing tag </tagName>
			if depth > 0 {
				depth--
				if depth == 0 && contentStartIdx != -1 {
					// Found the matching closing tag for the first opening tag
					contentEndIdx := firstCloseTagIdx
					return strings.TrimSpace(response[contentStartIdx:contentEndIdx])
				}
			}
			searchFrom = firstCloseTagIdx + len(closeTag) // Continue search after </tagName>
		} else {
			// Should not happen based on initial check, but safety break
			break
		}
	}

	// If loop finishes and depth > 0, or contentStartIdx was never set, tag was not found or properly closed
	return ""
}

// ExtractMultipleData extracts data for multiple different XML tag pairs from the same response.
// The tagNames parameter specifies which tags to look for.
// Returns a map of tag names to their extracted content.
func ExtractMultipleData(response string, tagNames []string) map[string]string {
	results := make(map[string]string)

	for _, tagName := range tagNames {
		extractor := NewXMLStructuredDataExtractor(tagName)
		content := extractor.Extract(response)
		if content != "" {
			results[tagName] = content
		}
	}

	return results
}
