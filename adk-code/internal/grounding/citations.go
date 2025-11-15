package grounding

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"google.golang.org/genai"
)

// CitationFormatter handles the formatting and rendering of grounded responses
// with inline citations and source attribution.
type CitationFormatter struct {
	// IncludeLinks adds hyperlinks to citation references
	IncludeLinks bool
	// IncludeMetadata adds source metadata (title, domain) to citations
	IncludeMetadata bool
	// CitationStyle determines how citations are rendered: "inline", "footnote", or "mixed"
	CitationStyle string
	// IncludeSourcesList adds a consolidated sources section at the end
	IncludeSourcesList bool
	// MaxSourcesPerSegment limits the number of citations per text segment
	MaxSourcesPerSegment int
}

// NewCitationFormatter creates a new citation formatter with sensible defaults
func NewCitationFormatter() *CitationFormatter {
	return &CitationFormatter{
		IncludeLinks:         true,
		IncludeMetadata:      true,
		CitationStyle:        "inline",
		IncludeSourcesList:   true,
		MaxSourcesPerSegment: 5,
	}
}

// FormatWithCitations adds inline citations to the response text based on grounding metadata
// It processes grounding supports and chunks to insert citation references.
func (cf *CitationFormatter) FormatWithCitations(
	responseText string,
	groundingMetadata *genai.GroundingMetadata,
) string {
	if groundingMetadata == nil || len(groundingMetadata.GroundingSupports) == 0 {
		return responseText
	}

	supports := groundingMetadata.GroundingSupports
	chunks := groundingMetadata.GroundingChunks

	// Sort supports by end_index in descending order to avoid index shifting
	sortedSupports := supports
	sort.Slice(sortedSupports, func(i, j int) bool {
		return sortedSupports[i].Segment.EndIndex > sortedSupports[j].Segment.EndIndex
	})

	result := responseText
	for _, support := range sortedSupports {
		if support.Segment == nil {
			continue
		}

		endIdx := int(support.Segment.EndIndex)
		if endIdx > len(result) {
			endIdx = len(result)
		}

		// Build citation string for this segment
		citationStr := cf.buildCitationString(support.GroundingChunkIndices, chunks)
		if citationStr == "" {
			continue
		}

		// Insert citation at segment end
		result = result[:endIdx] + citationStr + result[endIdx:]
	}

	// Optionally add a sources list at the end
	if cf.IncludeSourcesList && len(chunks) > 0 {
		sourcesList := cf.buildSourcesList(chunks)
		result = result + "\n\n" + sourcesList
	}

	return result
}

// buildCitationString creates formatted citation references for a set of chunk indices
func (cf *CitationFormatter) buildCitationString(chunkIndices []int32, chunks []*genai.GroundingChunk) string {
	if len(chunkIndices) == 0 {
		return ""
	}

	// Limit number of citations per segment
	limit := cf.MaxSourcesPerSegment
	if limit > len(chunkIndices) {
		limit = len(chunkIndices)
	}

	var citations []string
	for i := 0; i < limit; i++ {
		idx := chunkIndices[i]
		if idx < 0 || int(idx) >= len(chunks) {
			continue
		}

		chunk := chunks[idx]
		citation := cf.formatSingleCitation(int(idx)+1, chunk)
		if citation != "" {
			citations = append(citations, citation)
		}
	}

	if len(citations) == 0 {
		return ""
	}

	// Format based on style preference
	switch cf.CitationStyle {
	case "footnote":
		return fmt.Sprintf("[%s]", strings.Join(citations, ","))
	case "mixed":
		// Combine inline and superscript style
		return fmt.Sprintf("^%s", strings.Join(citations, ","))
	default: // "inline"
		return " " + strings.Join(citations, " ")
	}
}

// formatSingleCitation formats a single citation reference
func (cf *CitationFormatter) formatSingleCitation(index int, chunk *genai.GroundingChunk) string {
	if chunk == nil || chunk.Web == nil {
		return ""
	}

	uri := chunk.Web.URI
	title := chunk.Web.Title

	if cf.IncludeLinks {
		if cf.IncludeMetadata && title != "" {
			domain := extractDomain(uri)
			return fmt.Sprintf("[%d](%s) %s", index, uri, domain)
		}
		return fmt.Sprintf("[%d](%s)", index, uri)
	}

	if cf.IncludeMetadata {
		if title != "" {
			return fmt.Sprintf("[%d: %s]", index, title)
		}
		return fmt.Sprintf("[%d: %s]", index, extractDomain(uri))
	}

	return fmt.Sprintf("[%d]", index)
}

// buildSourcesList creates a consolidated sources section
func (cf *CitationFormatter) buildSourcesList(chunks []*genai.GroundingChunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "### Sources")
	lines = append(lines, "")

	for i, chunk := range chunks {
		if chunk != nil && chunk.Web != nil {
			uri := chunk.Web.URI
			title := chunk.Web.Title
			if title == "" {
				title = extractDomain(uri)
			}
			lines = append(lines, fmt.Sprintf("%d. **%s**  \n   %s", i+1, title, uri))
		}
	}

	return strings.Join(lines, "\n")
}

// extractDomain extracts the domain from a URI
func extractDomain(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	if parsed.Host != "" {
		return parsed.Host
	}
	return uri
}

// GroundingInfo provides summary information about a grounded response
type GroundingInfo struct {
	// IsGrounded indicates if the response has grounding metadata
	IsGrounded bool
	// SourceCount is the number of sources used
	SourceCount int
	// SearchQueriesUsed is the list of search queries executed
	SearchQueriesUsed []string
	// Domains is the list of unique domains in sources
	Domains []string
	// SourceTitles is the list of source titles
	SourceTitles []string
}

// ExtractGroundingInfo extracts summary information from grounding metadata
func ExtractGroundingInfo(groundingMetadata *genai.GroundingMetadata) GroundingInfo {
	info := GroundingInfo{
		IsGrounded:        groundingMetadata != nil,
		SearchQueriesUsed: []string{},
		Domains:           []string{},
		SourceTitles:      []string{},
	}

	if groundingMetadata == nil {
		return info
	}

	info.SearchQueriesUsed = groundingMetadata.WebSearchQueries
	info.SourceCount = len(groundingMetadata.GroundingChunks)

	domainsMap := make(map[string]bool)
	for _, chunk := range groundingMetadata.GroundingChunks {
		if chunk.Web != nil {
			domain := extractDomain(chunk.Web.URI)
			if !domainsMap[domain] {
				info.Domains = append(info.Domains, domain)
				domainsMap[domain] = true
			}

			if chunk.Web.Title != "" {
				info.SourceTitles = append(info.SourceTitles, chunk.Web.Title)
			}
		}
	}

	sort.Strings(info.Domains)

	return info
}

// FormatGroundingInfo formats grounding information for display
func FormatGroundingInfo(info GroundingInfo) string {
	if !info.IsGrounded {
		return "⚠ Response not grounded in web search (using model knowledge)"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("✓ Response grounded in %d source(s)", info.SourceCount))

	if len(info.SearchQueriesUsed) > 0 {
		lines = append(lines, fmt.Sprintf("  Search queries: %s", strings.Join(info.SearchQueriesUsed, ", ")))
	}

	if len(info.Domains) > 0 {
		lines = append(lines, fmt.Sprintf("  Domains: %s", strings.Join(info.Domains, ", ")))
	}

	return strings.Join(lines, "\n")
}

// ValidateGroundingMetadata validates that grounding metadata is structurally sound
func ValidateGroundingMetadata(groundingMetadata *genai.GroundingMetadata) error {
	if groundingMetadata == nil {
		return nil // Not an error, just not grounded
	}

	// Check for segment-to-chunk consistency
	for _, support := range groundingMetadata.GroundingSupports {
		if support.Segment == nil {
			return fmt.Errorf("grounding support has nil segment")
		}

		for _, chunkIdx := range support.GroundingChunkIndices {
			if chunkIdx < 0 || int(chunkIdx) >= len(groundingMetadata.GroundingChunks) {
				return fmt.Errorf("chunk index %d out of bounds (have %d chunks)",
					chunkIdx, len(groundingMetadata.GroundingChunks))
			}
		}
	}

	// Check for valid chunk URIs
	for i, chunk := range groundingMetadata.GroundingChunks {
		if chunk.Web == nil {
			return fmt.Errorf("chunk %d has nil web source", i)
		}
		if chunk.Web.URI == "" {
			return fmt.Errorf("chunk %d has empty URI", i)
		}
	}

	return nil
}
