package cmd

import (
	"reflect"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func genArgsRequest(arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "probe", Arguments: arguments}}
}

func TestMCPGenBuildArgsFullEmission(t *testing.T) {
	build := mcpGenBuildArgs([]string{"gmail", "search"}, []mcpGenParam{
		{key: "max", flag: "--max", kind: mcpGenInt, def: "10", hasDef: true, min: intPtr(1), max: intPtr(100)},
		{key: "oldest", flag: "--oldest", kind: mcpGenBool},
		{key: "since", flag: "--since", kind: mcpGenString},
		{key: "query", kind: mcpGenString, required: true, position: 0},
	})

	args, err := build(genArgsRequest(map[string]any{"query": " q ", "oldest": true, "max": 500}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := []string{"gmail", "search", "--max", "100", "--oldest", "--", "q"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}

	// Defaults: absent int with default emits it; absent bool/string emit nothing.
	args, err = build(genArgsRequest(map[string]any{"query": "q"}))
	if err != nil {
		t.Fatalf("build defaults: %v", err)
	}
	want = []string{"gmail", "search", "--max", "10", "--", "q"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}

	// Missing required string errors.
	if _, err = build(genArgsRequest(map[string]any{})); err == nil {
		t.Fatal("expected error for missing query")
	}
	if _, err = build(genArgsRequest(map[string]any{"query": "  "})); err == nil {
		t.Fatal("expected error for blank query")
	}
}

func TestMCPGenBuildArgsOmitZeroAndAbsentInt(t *testing.T) {
	build := mcpGenBuildArgs([]string{"calendar", "events"}, []mcpGenParam{
		{key: "days", flag: "--days", kind: mcpGenInt, def: "0", hasDef: true, min: intPtr(0), max: intPtr(31), omitZero: true},
		{key: "offset", flag: "--offset", kind: mcpGenInt, required: true},
	})

	args, err := build(genArgsRequest(map[string]any{"days": 0, "offset": 5}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := []string{"calendar", "events", "--offset", "5"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}

	// Required int without default: absent → omitted (schema validation is the gate).
	args, err = build(genArgsRequest(map[string]any{"days": 7}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want = []string{"calendar", "events", "--days", "7"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}

func TestMCPGenBuildArgsStringModes(t *testing.T) {
	build := mcpGenBuildArgs([]string{"docs", "write"}, []mcpGenParam{
		{key: "body", flag: "--body", kind: mcpGenString, required: true, text: true},
		{key: "input", flag: "--input", kind: mcpGenString, def: "USER_ENTERED", hasDef: true},
		{key: "values_json", flag: "--values-json", kind: mcpGenString, required: true, json2d: true},
	})

	args, err := build(genArgsRequest(map[string]any{"body": " keep spaces ", "values_json": " [[1, \"a\"]] "}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := []string{"docs", "write", "--body", " keep spaces ", "--input", "USER_ENTERED", "--values-json", `[[1,"a"]]`}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}

	if _, err = build(genArgsRequest(map[string]any{"body": "b", "values_json": "@file.json"})); err == nil {
		t.Fatal("expected json2d rejection of @file input")
	}

	// Bool with default=true emits unless explicitly disabled.
	sanitize := mcpGenBuildArgs([]string{"gmail", "get"}, []mcpGenParam{
		{key: "sanitize_content", flag: "--sanitize-content", kind: mcpGenBool, def: "true", hasDef: true},
	})
	args, err = sanitize(genArgsRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want = []string{"gmail", "get", "--sanitize-content"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	args, err = sanitize(genArgsRequest(map[string]any{"sanitize_content": false}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want = []string{"gmail", "get"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}

func TestMCPGenBuildArgsPositionalOrderAndOptional(t *testing.T) {
	build := mcpGenBuildArgs([]string{"sheets", "get"}, []mcpGenParam{
		{key: "range", kind: mcpGenString, required: true, position: 1},
		{key: "spreadsheet_id", kind: mcpGenString, required: true, position: 0},
		{key: "calendar_id", kind: mcpGenString, position: 2},
	})
	args, err := build(genArgsRequest(map[string]any{"range": "A1:B2", "spreadsheet_id": "sid"}))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := []string{"sheets", "get", "--", "sid", "A1:B2"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
}
