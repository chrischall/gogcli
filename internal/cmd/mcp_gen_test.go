package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
)

type mcpGenWalkFixture struct {
	Widgets struct {
		List struct {
			Query []string `arg:"" name:"query" help:"Search query" mcp:"query" mcpdesc:"Widget query"`
			Max   int64    `name:"max" help:"Max results" default:"10" mcp:"max,default=10,min=1,max=100"`
			All   bool     `name:"all" help:"All pages" mcp:"all"`
		} `cmd:"" name:"list" help:"List widgets" mcp:"widgets_list,read" mcpdesc:"List widgets with filters."`
		Delete struct {
			ID string `arg:"" name:"id" help:"Widget ID" mcp:"widget_id"`
		} `cmd:"" name:"delete" help:"Delete a widget" mcp:"widgets_delete,write"`
		Hidden struct {
			ID string `arg:"" name:"id" help:"Widget ID"`
		} `cmd:"" name:"hidden" help:"Not exported"`
	} `cmd:"" name:"widgets" help:"Widget ops"`
}

func TestMCPSpecsFromModel(t *testing.T) {
	var cli mcpGenWalkFixture
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	specs, err := mcpSpecsFromModel(parser.Model.Node)
	if err != nil {
		t.Fatalf("mcpSpecsFromModel: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("specs = %d, want 2 (hidden command must not be exported)", len(specs))
	}

	list := specs[0]
	if list.Name != "widgets_list" || list.Service != "widgets" || list.Risk != mcpRiskRead {
		t.Fatalf("spec = %#v", list)
	}
	if list.Description != "List widgets with filters." {
		t.Fatalf("description = %q", list.Description)
	}
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: list.Name, Arguments: map[string]any{"query": "x", "all": true}}}
	args, err := list.BuildArgs(req)
	if err != nil {
		t.Fatalf("BuildArgs: %v", err)
	}
	want := []string{"widgets", "list", "--max", "10", "--all", "--", "x"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}

	del := specs[1]
	if del.Name != "widgets_delete" || del.Risk != mcpRiskWrite {
		t.Fatalf("delete spec = %#v", del)
	}
	if del.Description != "Delete a widget" {
		t.Fatalf("fallback to kong help failed: %q", del.Description)
	}
}

type mcpGenNonLeafFixture struct {
	Widgets struct {
		Sub struct {
			Leaf struct{} `cmd:"" name:"leaf" help:"leaf"`
		} `cmd:"" name:"sub" help:"Non-leaf" mcp:"widgets_sub,read"`
	} `cmd:"" name:"widgets" help:"Widget ops"`
}

func TestMCPSpecsFromModelRejectsNonLeaf(t *testing.T) {
	var cli mcpGenNonLeafFixture
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	if _, err := mcpSpecsFromModel(parser.Model.Node); err == nil || !strings.Contains(err.Error(), "leaf") {
		t.Fatalf("expected non-leaf error, got %v", err)
	}
}

func TestMCPGeneratedToolsAgainstRealGrammar(t *testing.T) {
	specs, err := mcpGeneratedTools()
	if err != nil {
		t.Fatalf("mcpGeneratedTools must succeed on the real grammar: %v", err)
	}
	_ = specs // Non-empty once services are migrated; must never error.
}

func TestMCPAllToolNamesUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, tool := range mcpAllTools() {
		if seen[tool.Name] {
			t.Fatalf("duplicate MCP tool name %q", tool.Name)
		}
		seen[tool.Name] = true
	}
}
