// Package rtk implements token reduction for shell command output.
// It filters and compresses command output to minimize LLM token usage.
package rtk

import (
	"regexp"
	"strings"
)

// FilterResult holds the output after filtering.
type FilterResult struct {
	Output      string
	OrigLen     int
	FilteredLen int
	ExitCode    int
}

// CommandFilter is the interface every command module implements.
type CommandFilter interface {
	Name() string
	CanHandle(subcommand string) bool
	Filter(cmd, sub string, stdout, stderr []byte, exitCode int) FilterResult
}

// ApplyPostFilter applies regex-based post-processing:
// - ANSI escape removal
// - blank line collapsing
// - duplicate line aggregation
func ApplyPostFilter(s string) string {
	// strip ANSI escape codes
	s = reANSI.ReplaceAllString(s, "")

	// collapse multiple blank lines into one
	s = reBlankLines.ReplaceAllString(s, "\n\n")

	// deduplicate consecutive identical lines
	s = dedup(s)

	return strings.TrimSpace(s)
}

// Truncate limits output to head+tail lines with an omission marker.
func Truncate(s string, headLines, tailLines int) string {
	lines := strings.Split(s, "\n")
	total := len(lines)
	keep := headLines + tailLines
	if total <= keep {
		return s
	}
	var out []string
	out = append(out, lines[:headLines]...)
	omitted := total - keep
	out = append(out, strings.Repeat(" ", 4)+"... ("+itoa(omitted)+" lines omitted) ...")
	out = append(out, lines[total-tailLines:]...)
	return strings.Join(out, "\n")
}

// TruncateBytes enforces a maximum byte limit.
func TruncateBytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "\n... (output truncated at " + itoa(maxBytes) + " bytes)"
}

var (
	reANSI      = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	reBlankLines = regexp.MustCompile(`\n{3,}`)
)

func dedup(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) < 3 {
		return s
	}
	var out []string
	prev := ""
	count := 0
	for _, line := range lines {
		if line == prev && line != "" {
			count++
			continue
		}
		if count > 0 {
			out = append(out, prev+" (x"+itoa(count+1)+")")
			count = 0
		} else if prev != "" || len(out) > 0 {
			out = append(out, prev)
		}
		prev = line
	}
	if count > 0 {
		out = append(out, prev+" (x"+itoa(count+1)+")")
	} else {
		out = append(out, prev)
	}
	return strings.Join(out, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
