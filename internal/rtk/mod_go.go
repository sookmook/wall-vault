package rtk

import (
	"strings"
)

// GoFilter handles go command output filtering.
type GoFilter struct{}

func (g *GoFilter) Name() string { return "go" }

func (g *GoFilter) CanHandle(sub string) bool {
	switch sub {
	case "test", "build", "vet":
		return true
	}
	return false
}

func (g *GoFilter) Filter(cmd, sub string, stdout, stderr []byte, exitCode int) FilterResult {
	// go test outputs to stdout, errors to stderr
	raw := string(stdout) + string(stderr)
	var filtered string

	switch sub {
	case "test":
		filtered = filterGoTest(raw)
	case "build":
		filtered = filterGoBuild(raw, exitCode)
	case "vet":
		filtered = filterGoBuild(raw, exitCode) // same format
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

// filterGoTest hides passing tests, shows failures and summary.
func filterGoTest(raw string) string {
	lines := strings.Split(raw, "\n")
	var failures []string
	var summaries []string
	passCount := 0
	failCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "--- PASS"):
			passCount++
		case strings.HasPrefix(trimmed, "--- FAIL"):
			failCount++
			failures = append(failures, line)
		case strings.HasPrefix(trimmed, "FAIL") || strings.HasPrefix(trimmed, "ok"):
			summaries = append(summaries, line)
		case strings.HasPrefix(trimmed, "=== RUN") || strings.HasPrefix(trimmed, "=== PAUSE") || strings.HasPrefix(trimmed, "=== CONT"):
			// skip run/pause/cont markers for passing tests
			continue
		case failCount > 0 || (len(failures) > 0 && !strings.HasPrefix(trimmed, "---")):
			// include output after a failure
			if trimmed != "" {
				failures = append(failures, line)
			}
		}
	}

	var out []string
	if passCount > 0 {
		out = append(out, "PASS: "+itoa(passCount)+" tests")
	}
	if len(failures) > 0 {
		out = append(out, strings.Join(failures, "\n"))
	}
	if len(summaries) > 0 {
		out = append(out, strings.Join(summaries, "\n"))
	}
	if len(out) == 0 {
		return raw
	}
	return strings.Join(out, "\n")
}

// filterGoBuild shows only error/warning lines.
func filterGoBuild(raw string, exitCode int) string {
	if exitCode == 0 {
		return "build ok"
	}
	// show error lines only
	lines := strings.Split(raw, "\n")
	var errors []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		errors = append(errors, line)
	}
	return strings.Join(errors, "\n")
}
