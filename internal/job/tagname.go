package job

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	ikbPrefix      = "IKB"
	chunkCommentFmt = "ikb:%d" // comment encodes chunk index for custom_isp
)

var nonTagChars = regexp.MustCompile(`[^\p{Han}a-zA-Z0-9]`)

// sanitizeTag strips characters that are invalid in iKuai tag names,
// keeping Chinese characters, ASCII letters, and digits.
func sanitizeTag(raw string) string {
	return nonTagChars.ReplaceAllString(raw, "")
}

// buildTagName returns "IKB" + sanitized tag, e.g. "IKBcn" for "cn".
func buildTagName(raw string) string {
	return ikbPrefix + sanitizeTag(raw)
}

// buildChunkComment returns the comment string used to mark chunk idx (1-based).
// Custom ISP comment: "ikb:1", "ikb:2", ...
func buildChunkComment(idx int) string {
	return fmt.Sprintf(chunkCommentFmt, idx)
}

// parseChunkComment extracts the chunk index from a comment produced by buildChunkComment.
// Returns 0 if the comment is not in the expected format.
func parseChunkComment(comment string) int {
	if !strings.HasPrefix(comment, "ikb:") {
		return 0
	}
	n, err := strconv.Atoi(comment[4:])
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

// buildStreamDomainComment returns the comment used to identify a stream_domain chunk.
// Format: "IKBtag:idx", e.g. "IKBcn:1" for tag="cn", idx=1.
// Since stream_domain has no name field, the tag is embedded in the comment.
func buildStreamDomainComment(tag string, idx int) string {
	return fmt.Sprintf("%s:%d", buildTagName(tag), idx)
}

// parseStreamDomainComment extracts (tag, idx) from a stream_domain comment.
// Returns ("", 0) if the comment doesn't match the IKB pattern.
func parseStreamDomainComment(comment, tag string) int {
	prefix := buildTagName(tag) + ":"
	if !strings.HasPrefix(comment, prefix) {
		return 0
	}
	n, err := strconv.Atoi(comment[len(prefix):])
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

// splitChunks divides items into chunks of at most chunkSize entries.
func splitChunks(items []string, chunkSize int) [][]string {
	if chunkSize <= 0 {
		chunkSize = 5000
	}
	var chunks [][]string
	for i := 0; i < len(items); i += chunkSize {
		end := min(i+chunkSize, len(items))
		chunk := make([]string, end-i)
		copy(chunk, items[i:end])
		chunks = append(chunks, chunk)
	}
	return chunks
}

// dedupe removes duplicate strings while preserving order.
func dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := items[:0]
	for _, s := range items {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
