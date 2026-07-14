package judge

import (
	"strings"
)

const maxOutputBytes = 8 * 1024

// CompareOutput performs strict text comparison after normalization.
// Normalization: CRLF/CR → LF, trailing whitespace per line, trailing blank lines.
func CompareOutput(actual, expected string) bool {
	return normalize(actual) == normalize(expected)
}

func normalize(s string) string {
	// 1. Normalize line endings: \r\n and \r → \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// 2. Split into lines, trim trailing whitespace per line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// 3. Remove trailing blank lines
	end := len(lines)
	for end > 0 && lines[end-1] == "" {
		end--
	}
	lines = lines[:end]

	return strings.Join(lines, "\n")
}

// TruncateOutput truncates output to maxOutputBytes (8KB), appending a marker if truncated.
func TruncateOutput(output string) string {
	if len(output) <= maxOutputBytes {
		return output
	}
	return output[:maxOutputBytes] + "\n... [truncated]"
}

// SanitizePath removes host filesystem paths from error output.
func SanitizePath(output string) string {
	// Common path patterns to strip: /home/..., /tmp/..., /usr/..., /var/..., C:\...
	// Replace with generic placeholder
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		lines[i] = sanitizeLine(line)
	}
	return strings.Join(lines, "\n")
}

func sanitizeLine(line string) string {
	// Simple approach: replace common path prefixes
	prefixes := []string{"/home/", "/tmp/", "/usr/", "/var/", "/opt/", "C:\\"}
	result := line
	for _, prefix := range prefixes {
		idx := strings.Index(result, prefix)
		for idx != -1 {
			// Find end of path (space or end of string)
			end := idx
			for end < len(result) && result[end] != ' ' && result[end] != '\t' {
				end++
			}
			result = result[:idx] + "[path]" + result[end:]
			idx = strings.Index(result, prefix)
		}
	}
	return result
}
