package rtk

import (
	"strings"
)

// GeneralFilter handles common system commands.
type GeneralFilter struct{}

func (g *GeneralFilter) Name() string { return "" }

func (g *GeneralFilter) CanHandle(sub string) bool { return false }

func (g *GeneralFilter) Filter(cmd, sub string, stdout, stderr []byte, exitCode int) FilterResult {
	return FilterResult{Output: string(stdout), OrigLen: len(stdout), FilteredLen: len(stdout), ExitCode: exitCode}
}

// GeneralFilters returns filters for common commands: ls, find, grep, cat.
var GeneralFilters = map[string]func(string) string{
	"find": filterFind,
	"grep": filterGrep,
	"rg":   filterGrep,
}

func filterFind(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 50 {
		return raw
	}
	return Truncate(raw, 30, 10) + "\n(" + itoa(len(lines)) + " files total)"
}

func filterGrep(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) <= 50 {
		return raw
	}
	return Truncate(raw, 30, 10) + "\n(" + itoa(len(lines)) + " matches total)"
}
