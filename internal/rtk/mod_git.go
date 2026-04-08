package rtk

import (
	"regexp"
	"strings"
)

// GitFilter handles git command output filtering.
type GitFilter struct{}

func (g *GitFilter) Name() string { return "git" }

func (g *GitFilter) CanHandle(sub string) bool {
	switch sub {
	case "status", "diff", "log", "push", "pull", "fetch", "branch", "stash":
		return true
	}
	return false
}

func (g *GitFilter) Filter(cmd, sub string, stdout, stderr []byte, exitCode int) FilterResult {
	raw := string(stdout)
	var filtered string

	switch sub {
	case "status":
		filtered = filterGitStatus(raw)
	case "diff":
		filtered = filterGitDiff(raw)
	case "log":
		filtered = filterGitLog(raw)
	case "push", "pull", "fetch":
		filtered = filterGitTransfer(raw)
	case "branch":
		filtered = raw // already compact
	case "stash":
		filtered = raw
	default:
		filtered = raw
	}

	return FilterResult{
		Output:      filtered,
		OrigLen:     len(raw),
		FilteredLen: len(filtered),
		ExitCode:    exitCode,
	}
}

// filterGitStatus extracts only the file change summary.
func filterGitStatus(raw string) string {
	lines := strings.Split(raw, "\n")
	var staged, unstaged, untracked []string
	section := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		switch {
		case strings.Contains(line, "Changes to be committed"):
			section = "staged"
		case strings.Contains(line, "Changes not staged"):
			section = "unstaged"
		case strings.Contains(line, "Untracked files"):
			section = "untracked"
		case strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "        "):
			// file entry
			entry := strings.TrimSpace(line)
			switch section {
			case "staged":
				staged = append(staged, entry)
			case "unstaged":
				unstaged = append(unstaged, entry)
			case "untracked":
				untracked = append(untracked, entry)
			}
		}
	}

	var parts []string
	if len(staged) > 0 {
		parts = append(parts, "staged: "+strings.Join(staged, ", "))
	}
	if len(unstaged) > 0 {
		parts = append(parts, "unstaged: "+strings.Join(unstaged, ", "))
	}
	if len(untracked) > 0 {
		parts = append(parts, "untracked: "+strings.Join(untracked, ", "))
	}
	if len(parts) == 0 {
		// check for clean status
		if strings.Contains(raw, "nothing to commit") {
			return "clean (nothing to commit)"
		}
		return raw
	}
	return strings.Join(parts, "\n")
}

// filterGitDiff keeps file headers and changed lines, reduces context.
func filterGitDiff(raw string) string {
	lines := strings.Split(raw, "\n")
	var out []string
	contextCount := 0
	maxContext := 3

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			out = append(out, line)
			contextCount = 0
		case strings.HasPrefix(line, "--- "), strings.HasPrefix(line, "+++ "):
			out = append(out, line)
		case strings.HasPrefix(line, "@@"):
			out = append(out, line)
			contextCount = 0
		case strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-"):
			out = append(out, line)
			contextCount = 0
		default:
			// context line — keep up to maxContext
			contextCount++
			if contextCount <= maxContext {
				out = append(out, line)
			}
		}
	}
	return strings.Join(out, "\n")
}

var reGitLogCommit = regexp.MustCompile(`^commit [0-9a-f]{40}`)

// filterGitLog keeps commit hash + first message line only.
func filterGitLog(raw string) string {
	lines := strings.Split(raw, "\n")
	var out []string
	inCommit := false
	gotMessage := false

	for _, line := range lines {
		if reGitLogCommit.MatchString(line) && len(line) >= 19 {
			hash := line[7:19] // short hash
			out = append(out, hash)
			inCommit = true
			gotMessage = false
			continue
		}
		if inCommit && len(out) > 0 {
			if strings.HasPrefix(line, "    ") && !gotMessage {
				out[len(out)-1] += " " + strings.TrimSpace(line)
				gotMessage = true
			}
			if line == "" && gotMessage {
				inCommit = false
			}
		}
	}
	return strings.Join(out, "\n")
}

// filterGitTransfer extracts summary line from push/pull/fetch.
func filterGitTransfer(raw string) string {
	lines := strings.Split(raw, "\n")
	var summary []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// skip delta/compression progress lines
		if strings.Contains(line, "Compressing") || strings.Contains(line, "Writing") ||
			strings.Contains(line, "Counting") || strings.Contains(line, "Delta") ||
			strings.Contains(line, "remote:") {
			continue
		}
		summary = append(summary, trimmed)
	}
	if len(summary) == 0 {
		return strings.TrimSpace(raw)
	}
	return strings.Join(summary, "\n")
}
