package cmd

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
)

type mcpGenFieldFixture struct {
	Widget struct {
		Query  []string `arg:"" name:"query" help:"Search query" mcp:"query" mcpdesc:"Widget search query"`
		Max    int64    `name:"max" help:"Max results" default:"10" mcp:"max,default=10,min=1,max=100" mcpdesc:"Maximum results"`
		All    bool     `name:"all" help:"All pages" mcp:"all"`
		Body   string   `name:"body" help:"Body text" mcp:"body,required,text"`
		Render string   `name:"render" help:"Render option" enum:"A,B" default:"A" mcp:"render"`
		Skip   string   `name:"skip" help:"Not exposed"`
	} `cmd:"" name:"widget" help:"Widget ops"`
}

func mcpGenFieldFixtureNode(t *testing.T) *kong.Node {
	t.Helper()
	var cli mcpGenFieldFixture
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}
	return parser.Model.Children[0]
}

func fixtureFlag(t *testing.T, node *kong.Node, name string) *kong.Value {
	t.Helper()
	for _, f := range node.Flags {
		if f.Name == name {
			return f.Value
		}
	}
	t.Fatalf("flag %q not found", name)
	return nil
}

func optionSchema(t *testing.T, opt mcp.ToolOption) map[string]any {
	t.Helper()
	tool := mcp.NewTool("probe", opt)
	raw, err := json.Marshal(tool.InputSchema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	return schema
}

func TestMCPGenFieldInteger(t *testing.T) {
	node := mcpGenFieldFixtureNode(t)
	param, opt, err := mcpGenField(fixtureFlag(t, node, "max"), true)
	if err != nil {
		t.Fatalf("mcpGenField: %v", err)
	}
	if param.key != "max" || param.flag != "--max" || param.kind != mcpGenInt {
		t.Fatalf("param = %#v", param)
	}
	if !param.hasDef || param.def != "10" || *param.min != 1 || *param.max != 100 {
		t.Fatalf("param bounds = %#v", param)
	}
	schema := optionSchema(t, opt)
	prop := schema["properties"].(map[string]any)["max"].(map[string]any)
	if prop["type"] != "integer" || prop["default"] != float64(10) || prop["minimum"] != float64(1) || prop["maximum"] != float64(100) {
		t.Fatalf("schema prop = %#v", prop)
	}
	if prop["description"] != "Maximum results" {
		t.Fatalf("description from mcpdesc = %q", prop["description"])
	}
}

func TestMCPGenFieldPositionalRequiredString(t *testing.T) {
	node := mcpGenFieldFixtureNode(t)
	if len(node.Positional) != 1 {
		t.Fatalf("positional count = %d", len(node.Positional))
	}
	param, opt, err := mcpGenField(node.Positional[0], false)
	if err != nil {
		t.Fatalf("mcpGenField: %v", err)
	}
	if param.flag != "" || param.kind != mcpGenString || !param.required {
		t.Fatalf("param = %#v", param)
	}
	schema := optionSchema(t, opt)
	required, _ := schema["required"].([]any)
	if len(required) != 1 || required[0] != "query" {
		t.Fatalf("required = %#v", required)
	}
}

func TestMCPGenFieldBoolDefaultsFalse(t *testing.T) {
	node := mcpGenFieldFixtureNode(t)
	param, opt, err := mcpGenField(fixtureFlag(t, node, "all"), true)
	if err != nil {
		t.Fatalf("mcpGenField: %v", err)
	}
	if param.kind != mcpGenBool || param.required {
		t.Fatalf("param = %#v", param)
	}
	prop := optionSchema(t, opt)["properties"].(map[string]any)["all"].(map[string]any)
	if prop["type"] != "boolean" || prop["default"] != false {
		t.Fatalf("prop = %#v", prop)
	}
}

func TestMCPGenFieldTextAndEnumAndSkip(t *testing.T) {
	node := mcpGenFieldFixtureNode(t)

	param, _, err := mcpGenField(fixtureFlag(t, node, "body"), true)
	if err != nil || !param.text || !param.required {
		t.Fatalf("body param = %#v, err = %v", param, err)
	}

	_, opt, err := mcpGenField(fixtureFlag(t, node, "render"), true)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	prop := optionSchema(t, opt)["properties"].(map[string]any)["render"].(map[string]any)
	enum, _ := prop["enum"].([]any)
	if len(enum) != 2 || enum[0] != "A" || enum[1] != "B" {
		t.Fatalf("enum from kong tag = %#v", prop)
	}

	param, opt, err = mcpGenField(fixtureFlag(t, node, "skip"), true)
	if param != nil || opt != nil || err != nil {
		t.Fatalf("untagged field must be skipped: %#v, %#v, %v", param, opt, err)
	}
}
