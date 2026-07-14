package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpAllTools returns every typed MCP tool, grouped by service area. Each area
// contributes its own slice so the surface stays "gated by area": --allow-tool
// selectors such as gmail, drive.*, or calendar map directly onto these groups,
// and --allow-write hides the write tools within each area until requested.
func mcpAllTools() []mcpToolSpec {
	generated, err := mcpGeneratedTools()
	if err != nil {
		// Annotation mistakes are programmer errors caught by tests; a broken
		// grammar must never serve a partial tool surface.
		panic(fmt.Sprintf("invalid mcp annotations: %v", err))
	}
	groups := [][]mcpToolSpec{
		mcpCustomTools(),
		generated,
	}
	var all []mcpToolSpec
	for _, group := range groups {
		all = append(all, group...)
	}
	return all
}

// mcpArgs is a small fluent helper for assembling a child gog command's argv
// from a typed MCP request. Optional flags are only appended when present, and
// positional arguments are emitted after a `--` separator so model-supplied
// values can never be parsed as flags.
type mcpArgs struct {
	req mcp.CallToolRequest
	out []string
	err error
}

func mcpCommand(req mcp.CallToolRequest, sub ...string) *mcpArgs {
	return &mcpArgs{req: req, out: append([]string{}, sub...)}
}

// str appends "--flag value" when the named string arg is present and non-empty.
func (a *mcpArgs) str(key, flag string) *mcpArgs {
	if v := strings.TrimSpace(a.req.GetString(key, "")); v != "" {
		a.out = append(a.out, flag, v)
	}
	return a
}

// done finalizes the argv, appending positional arguments after `--`.
func (a *mcpArgs) done(pos ...string) ([]string, error) {
	if a.err != nil {
		return nil, a.err
	}
	out := a.out
	if len(pos) > 0 {
		out = append(out, "--")
		out = append(out, pos...)
	}
	return out, nil
}

func requireMCPText(req mcp.CallToolRequest, key string) (string, error) {
	value, err := req.RequireString(key)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("empty %s", key)
	}
	return value, nil
}

func requireMCPLiteralValuesJSON(req mcp.CallToolRequest, key string) (string, error) {
	value, err := requireMCPText(req, key)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "-" || strings.HasPrefix(trimmed, "@") {
		return "", fmt.Errorf("%s must be literal JSON, not stdin or @file input", key)
	}
	var rows [][]any
	dec := json.NewDecoder(bytes.NewReader([]byte(trimmed)))
	dec.UseNumber()
	if unmarshalErr := dec.Decode(&rows); unmarshalErr != nil {
		return "", fmt.Errorf("invalid %s JSON 2D array: %w", key, unmarshalErr)
	}
	var extra any
	if extraErr := dec.Decode(&extra); extraErr != io.EOF {
		return "", fmt.Errorf("invalid %s JSON 2D array: trailing content", key)
	}
	canonical, err := json.Marshal(rows)
	if err != nil {
		return "", fmt.Errorf("canonicalize %s: %w", key, err)
	}
	return string(canonical), nil
}
