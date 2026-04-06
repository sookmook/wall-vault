package rtk

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	defaultMaxLines = 200
	defaultMaxBytes = 32768
	defaultHead     = 50
	defaultTail     = 50
)

// Run executes the given command, filters its output, prints the result,
// and returns the original exit code.
func Run(args []string) int {
	if len(args) == 0 {
		return 0
	}

	cmdName := filepath.Base(args[0])
	var sub string
	if len(args) > 1 {
		sub = args[1]
	}

	// execute the original command with forced English locale for consistent parsing
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	cmd.Stdin = os.Stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = ws.ExitStatus()
			} else {
				exitCode = 1
			}
		} else {
			// command not found or other exec error
			os.Stderr.WriteString(err.Error() + "\n")
			return 127
		}
	}

	// apply filters
	registry := NewRegistry()
	result := applyFilter(registry, cmdName, sub, stdout.Bytes(), stderr.Bytes(), exitCode)

	// write filtered output to stdout
	if result.Output != "" {
		os.Stdout.WriteString(result.Output)
		if !strings.HasSuffix(result.Output, "\n") {
			os.Stdout.WriteString("\n")
		}
	}

	// append stderr (errors/warnings are always valuable)
	stderrStr := strings.TrimSpace(string(stderr.Bytes()))
	if stderrStr != "" && exitCode != 0 {
		os.Stdout.WriteString(stderrStr + "\n")
	}

	return result.ExitCode
}

func applyFilter(reg *Registry, cmd, sub string, stdout, stderr []byte, exitCode int) FilterResult {
	raw := string(stdout)

	// try command-specific filter
	if f := reg.Lookup(cmd); f != nil && f.CanHandle(sub) {
		result := f.Filter(cmd, sub, stdout, stderr, exitCode)
		result.Output = ApplyPostFilter(result.Output)
		result.Output = TruncateBytes(result.Output, defaultMaxBytes)
		return result
	}

	// passthrough with truncation
	filtered := ApplyPostFilter(raw)
	filtered = Truncate(filtered, defaultHead, defaultTail)
	filtered = TruncateBytes(filtered, defaultMaxBytes)
	return FilterResult{
		Output:      filtered,
		OrigLen:     len(raw),
		FilteredLen: len(filtered),
		ExitCode:    exitCode,
	}
}
