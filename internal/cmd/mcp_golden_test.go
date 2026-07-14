package cmd

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
)

var updateMCPGolden = flag.Bool("update-mcp-golden", false, "rewrite testdata/mcp_tools_golden.json from the current tool surface")

// mcpGoldenTool is the canonical, order-insensitive form of one MCP tool:
// schema (sorted required/enum) plus the child argv decomposed into path,
// flag→value map, and positionals via the real Kong model.
type mcpGoldenTool struct {
	Service     string            `json:"service"`
	Risk        string            `json:"risk"`
	Description string            `json:"description"`
	Schema      json.RawMessage   `json:"schema"`
	Path        []string          `json:"path"`
	Flags       map[string]string `json:"flags"`
	Positionals []string          `json:"positionals"`
}

func TestMCPToolSurfaceMatchesGolden(t *testing.T) {
	parser, _, err := newParserWithWriters("test", io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("newParser: %v", err)
	}
	root := parser.Model.Node

	got := map[string]mcpGoldenTool{}
	for _, spec := range mcpAllTools() {
		if _, dup := got[spec.Name]; dup {
			t.Fatalf("duplicate tool %q", spec.Name)
		}
		tool := newMCPTool(spec)
		args, buildErr := spec.BuildArgs(populatedRequest(spec))
		if buildErr != nil {
			t.Fatalf("%s: BuildArgs: %v", spec.Name, buildErr)
		}
		path, flags, positionals := canonicalMCPArgv(t, root, spec.Name, args)
		got[spec.Name] = mcpGoldenTool{
			Service:     spec.Service,
			Risk:        string(spec.Risk),
			Description: spec.Description,
			Schema:      canonicalMCPSchema(t, spec.Name, tool.InputSchema),
			Path:        path,
			Flags:       flags,
			Positionals: positionals,
		}
	}

	goldenPath := filepath.Join("testdata", "mcp_tools_golden.json")
	if *updateMCPGolden {
		data, marshalErr := json.MarshalIndent(got, "", "  ")
		if marshalErr != nil {
			t.Fatalf("marshal golden: %v", marshalErr)
		}
		if writeErr := os.WriteFile(goldenPath, append(data, '\n'), 0o644); writeErr != nil {
			t.Fatalf("write golden: %v", writeErr)
		}
		return
	}

	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run 'go test -run TestMCPToolSurfaceMatchesGolden -update-mcp-golden' once to create it): %v", err)
	}
	var want map[string]mcpGoldenTool
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatalf("unmarshal golden: %v", err)
	}

	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("tool %q missing from current surface", name)
			continue
		}
		if canonJSON(t, g) != canonJSON(t, w) {
			t.Errorf("tool %q drifted from golden:\n got: %s\nwant: %s", name, canonJSON(t, g), canonJSON(t, w))
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("unexpected new tool %q (not in golden)", name)
		}
	}
}

// canonicalMCPArgv splits a generated argv into (path, flags, positionals),
// resolving each token against the real command model so bool flags are
// distinguished from valued flags.
func canonicalMCPArgv(t *testing.T, root *kong.Node, tool string, args []string) ([]string, map[string]string, []string) {
	t.Helper()
	node := root
	var path []string
	i := 0
	for i < len(args) && args[i] != "--" && args[i][0] != '-' {
		child := commandChild(node, args[i])
		if child == nil {
			t.Fatalf("%s: unknown subcommand %q in argv %v", tool, args[i], args)
		}
		node = child
		path = append(path, args[i])
		i++
	}
	flags := map[string]string{}
	var positionals []string
	for i < len(args) {
		tok := args[i]
		if tok == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		f := commandFlag(node, tok)
		if f == nil {
			t.Fatalf("%s: unknown flag %q in argv %v", tool, tok, args)
		}
		if f.IsBool() {
			flags[tok] = ""
			i++
			continue
		}
		if i+1 >= len(args) {
			t.Fatalf("%s: flag %q missing value in argv %v", tool, tok, args)
		}
		flags[tok] = args[i+1]
		i += 2
	}
	return path, flags, positionals
}

// canonicalMCPSchema renders an input schema with sorted required and enum
// arrays so declaration-order differences do not register as drift.
func canonicalMCPSchema(t *testing.T, tool string, schema mcp.ToolInputSchema) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("%s: marshal schema: %v", tool, err)
	}
	var m map[string]any
	if unmarshalErr := json.Unmarshal(raw, &m); unmarshalErr != nil {
		t.Fatalf("%s: unmarshal schema: %v", tool, unmarshalErr)
	}
	sortAnyStrings(m, "required")
	if props, ok := m["properties"].(map[string]any); ok {
		for _, p := range props {
			if prop, ok := p.(map[string]any); ok {
				sortAnyStrings(prop, "enum")
			}
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("%s: re-marshal schema: %v", tool, err)
	}
	return out
}

func sortAnyStrings(m map[string]any, key string) {
	values, ok := m[key].([]any)
	if !ok {
		return
	}
	sort.Slice(values, func(i, j int) bool {
		a, _ := values[i].(string)
		b, _ := values[j].(string)
		return a < b
	})
}

// canonJSON normalizes a golden entry (including embedded RawMessage bytes)
// through an any round-trip so formatting differences never count as drift.
func canonJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("canon marshal: %v", err)
	}
	var x any
	if unmarshalErr := json.Unmarshal(b, &x); unmarshalErr != nil {
		t.Fatalf("canon unmarshal: %v", unmarshalErr)
	}
	out, err := json.Marshal(x)
	if err != nil {
		t.Fatalf("canon re-marshal: %v", err)
	}
	return string(out)
}
