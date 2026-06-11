// Command precommit is the cross-platform pre-commit gate invoked by the
// Claude Code (.claude/settings.json) and Codex (.codex/hooks.json) agent
// hooks. Agents pass the intercepted tool call as JSON on stdin; this gates
// only `git commit` invocations, blocking the commit when the code is
// unformatted or any make step fails.
//
// It checks formatting with gofmt directly (rather than `make fmt-ci`, whose
// shell recipe is Linux/CI-only) and then runs `make`, whose targets are single
// go/gosec commands valid on every platform (its own `go fmt` prerequisite is a
// no-op once the gofmt check above has passed). Invoked via `go run`, the hook
// command carries no shell-specific syntax and needs neither bash nor jq,
// behaving identically on Windows, macOS, and Linux.
//
// Blocking is signalled by a decision object on stdout (exit 0), not a non-zero
// exit code: `go run` collapses a program's exit code to 1, which neither agent
// treats as a block. The emitted object carries both agents' schemas — Claude
// reads hookSpecificOutput.permissionDecision, Codex reads decision — so a
// single payload blocks both.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if !isGitCommit() {
		return
	}

	root, err := repoRoot()
	if err != nil {
		block("precommit: cannot locate repo root: " + err.Error())
	}

	files, err := unformatted(root)
	if err != nil {
		block("precommit: gofmt failed: " + err.Error())
	}
	if len(files) > 0 {
		block("Files need formatting (run 'make fmt'):\n" + string(files))
	}

	cmd := exec.Command("make")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		block("make pipeline failed; fix the issues below and retry:\n" + string(out))
	}
}

// block writes a PreToolUse deny decision to stdout (understood by both Claude
// Code and Codex) and exits 0. See the package comment for why a decision
// object is used instead of a non-zero exit code.
func block(reason string) {
	payload := map[string]any{
		"decision": "block",
		"reason":   reason,
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, "precommit:", reason)
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, string(b))
	os.Exit(0)
}

// isGitCommit reports whether the agent's intercepted command runs `git commit`.
// The hook receives the tool call as JSON on stdin; the filter lives here (rather
// than in a matcher) so it also catches commits chained into a larger command.
// Malformed or empty stdin fails safe (gates the commit) rather than letting an
// unchecked commit through.
func isGitCommit() bool {
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(bytes.TrimSpace(data)) == 0 {
		return true
	}
	var payload struct {
		ToolInput struct {
			Command string `json:"command"`
		} `json:"tool_input"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return true
	}
	return containsGitCommit(payload.ToolInput.Command)
}

// containsGitCommit reports whether any segment of a (possibly chained) shell
// command invokes `git commit`, e.g. "git add -A && git commit -m x" or
// "cd sub && git commit". The command is split on shell separators (&, |, ;,
// newline) and each segment is checked for a leading `git commit`.
func containsGitCommit(command string) bool {
	segments := strings.FieldsFunc(command, func(r rune) bool {
		return r == '&' || r == '|' || r == ';' || r == '\n'
	})
	for _, segment := range segments {
		if strings.HasPrefix(strings.TrimSpace(segment), "git commit") {
			return true
		}
	}
	return strings.Contains(command, "git commit")
}

// unformatted returns the gofmt -l output (one path per line) for the module
// rooted at root; empty output means everything is formatted.
func unformatted(root string) ([]byte, error) {
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = root
	return cmd.Output()
}

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
