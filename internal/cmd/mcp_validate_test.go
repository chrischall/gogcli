package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
)

// skipPopulateKeys maps a tool name to option keys left unset when synthesizing
// a fully populated request, because setting them together with their siblings
// trips a deliberate mutual-exclusion guard in BuildArgs (e.g. docs_get
// tab/all_tabs and docs_write append/replace).
var skipPopulateKeys = map[string]map[string]bool{
	"docs_get":   {"all_tabs": true},
	"docs_write": {"replace": true},
}

// TestMCPToolArgsResolveToRealCommands builds every MCP tool's argv from a
// synthetic request and verifies the subcommand path and every flag resolve
// against the real gog command model. This catches typos in subcommand names
// and flag names across the whole typed surface.
func TestMCPToolArgsResolveToRealCommands(t *testing.T) {
	parser, _, err := newParserWithWriters("test", io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("newParser: %v", err)
	}
	root := parser.Model.Node

	for _, spec := range mcpAllTools() {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			args, buildErr := spec.BuildArgs(populatedRequest(spec))
			if buildErr != nil {
				t.Fatalf("BuildArgs: %v", buildErr)
			}

			pre := args
			for i, arg := range args {
				if arg == "--" {
					pre = args[:i]
					break
				}
			}

			node := root
			flagStart := len(pre)
			for i, tok := range pre {
				if strings.HasPrefix(tok, "-") {
					flagStart = i
					break
				}
				child := commandChild(node, tok)
				if child == nil {
					t.Fatalf("unknown subcommand %q in argv %v", tok, args)
				}
				node = child
			}

			if node == root {
				t.Fatalf("no subcommand resolved from argv %v", args)
			}
			if hasCommandChildren(node) {
				t.Fatalf("argv %v resolves to non-leaf command %q", args, node.Name)
			}

			for i := flagStart; i < len(pre); i++ {
				tok := pre[i]
				if !strings.HasPrefix(tok, "-") {
					continue // a flag value, not a flag
				}
				if commandFlag(node, tok) == nil {
					t.Fatalf("unknown flag %q for command %q (argv %v)", tok, node.Name, args)
				}
			}
		})
	}
}

func hasCommandChildren(node *kong.Node) bool {
	for _, child := range node.Children {
		if child.Type == kong.CommandNode {
			return true
		}
	}
	return false
}

// populatedRequest synthesizes a CallToolRequest with every declared option set
// to a type-appropriate dummy value, so BuildArgs emits each optional flag.
func populatedRequest(spec mcpToolSpec) mcp.CallToolRequest {
	tool := newMCPTool(spec)
	skip := skipPopulateKeys[spec.Name]
	arguments := map[string]any{}
	for name, raw := range tool.InputSchema.Properties {
		if skip[name] {
			continue
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			arguments[name] = "x"
			continue
		}
		switch prop["type"] {
		case "boolean":
			arguments[name] = true
		case "integer", "number":
			arguments[name] = 1
		default:
			if strings.Contains(name, "values_json") {
				arguments[name] = "[[1]]"
			} else if enum, ok := prop["enum"].([]any); ok && len(enum) > 0 {
				arguments[name] = enum[0]
			} else {
				arguments[name] = "x"
			}
		}
	}
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Name: spec.Name, Arguments: arguments}}
}
