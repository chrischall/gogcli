# MCP Annotation-Generated Tools Implementation Plan

> **For agentic workers:** Each task below is self-contained and executed by a fresh agent. Read the **Annotation Grammar Reference** and **Emission Semantics** sections before starting any task — every task depends on them. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Derive the 153 hand-written MCP tool specs from `mcp:"..."` / `mcpdesc:"..."` struct-tag annotations on the existing Kong CLI command structs, with a byte-identical tool surface, keeping only ~4 genuinely special tools hand-written.

**Architecture:** A generator walks the Kong runtime model (`parser.Model.Node`), finds command mounts tagged `mcp:"<tool>,<risk>"`, and synthesizes each `mcpToolSpec` — JSON schema from field types/tags, and a generic `BuildArgs` that assembles the child argv (flags, clamps, positionals after `--`). A golden-snapshot test captured from the current hand-written surface is the migration oracle: every migration task must keep it green. Hand-written specs are deleted service-by-service.

**Tech Stack:** Go, `github.com/alecthomas/kong` (runtime model + `Tag.Get`), `github.com/mark3labs/mcp-go`.

## Global Constraints

- Work happens in the worktree `/Users/chris/git/gogcli/.claude/worktrees/task-mcp` on branch `task/mcp`. Never `cd` out of it.
- The MCP tool surface must not change: tool names, services, risks, descriptions, schemas (types, required, defaults, enums, min/max), and the *canonical* child argv (subcommand path, flag→value set, positionals) stay identical. Flag *ordering* within argv may differ; the golden test canonicalizes it.
- The golden file `internal/cmd/testdata/mcp_tools_golden.json` is written once in Task 5 and **never regenerated afterwards**. A migration task that fails the golden test must fix its annotations, not the golden.
- The safety contract in `docs/mcp.md` § "Why this is not `gog_exec`" is preserved: fixed typed schemas, no argv passthrough, positionals after `--`, read-only by default.
- TDD: write the failing test first for all new framework code (Tasks 1–5).
- Run tests with `go test ./internal/cmd/ -run '<Pattern>' -count=1`. The full `./internal/cmd/` suite takes ~90s; run targeted tests during development and the full suite before each commit of Tasks 6–11.
- Format/lint before final commit of each task: `make fmt && make lint` (tools auto-install to `.tools/`).
- Commit at the end of each task (and at intermediate green points). Commit messages: `feat(mcp): ...` / `refactor(mcp): ...`. End every commit message with `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.

---

## Annotation Grammar Reference

Two custom struct tags, both readable at runtime via `kong.Tag.Get`:

### Mount annotation — on the command mount field (the field carrying `cmd:""`)

```
mcp:"<tool_name>,<read|write>[,service=<name>]"
mcpdesc:"<exact MCP tool description>"
```

- `<tool_name>`: the MCP tool name, e.g. `gmail_search`.
- Risk is mandatory and explicit: `read` or `write`.
- `service=` overrides the default service, which is derived from the tool name prefix up to the first `_` (e.g. `gmail_search` → `gmail`).
- `mcpdesc` carries the tool description **verbatim from the current hand-written spec**. If absent, the Kong `help:` text is used.
- The annotated command must be a **leaf** (no child commands) — the generator errors otherwise.

### Field annotation — on flags and positional args of the command struct

```
mcp:"<property_name>[,option...]"
mcpdesc:"<exact MCP property description>"
```

Fields without an `mcp` tag are **not exposed** (opt-in). Options:

| Option | Meaning |
| --- | --- |
| `required` | Force `required` in the schema (beyond what Kong implies). |
| `optional` | Force optional even if the CLI field is required. |
| `default=<v>` | MCP-side default. Goes into the schema AND is emitted into argv when the arg is absent. Bools: must be `true`/`false` (bools always get a schema default; `false` if unannotated). Integers: **must** have `default=` or `required`. |
| `min=<n>` / `max=<n>` | Integer schema `minimum`/`maximum`; runtime value is clamped into this range. |
| `enum=<v1\|v2\|...>` | Schema enum override (pipe-separated). Without it, the Kong `enum:` tag values are used when present. |
| `omitzero` | Integer flag is omitted from argv when its effective value is 0 (translates hand-written `if v > 0` guards). |
| `text` | String value is passed through untrimmed (message bodies, notes); requiredness still checks non-blank. Untagged strings are trimmed. |
| `json2d` | Value must be a literal JSON 2D array (validated + canonicalized via `requireMCPLiteralValuesJSON`); rejects `@file`/`-` stdin forms. |

Derivations when not overridden: property name maps to the CLI flag `--<kong-name>` (the annotation's property name is the MCP-visible name and may differ from the flag name); description = `mcpdesc` else Kong `help:`; type from the Go field type (`bool`→boolean, int kinds→integer, `string`/`[]string`→string); required = Kong-required (positionals, `required:""` flags) unless `optional` or `default=` given; enum from the Kong `enum:` tag.

## Emission Semantics (generic BuildArgs)

argv = subcommand path (canonical Kong names) + annotated flags in struct-field order + `--` + positionals in position order (the `--` group is omitted when no positional value is present).

- **bool**: emit bare `--flag` when the effective value (request value, else annotated default, else false) is true. Never emits `--flag=false`.
- **integer**: skipped when the arg is absent AND there is no `default=`. Otherwise value = request value or default, clamped to `[min,max]`, emitted as `--flag N` — unless `omitzero` and the value is 0.
- **string**: required → error when missing/blank; optional → emit when non-empty after trim (unless `text`); when empty and `default=` exists, the default is emitted.
- **positional**: same string resolution; non-empty values are appended after `--` in `Position` order; optional positionals are skipped when empty.

## Tools that stay hand-written (moved to `internal/cmd/mcp_tools_custom.go`)

| Tool | Why |
| --- | --- |
| `gmail_reply` | Chooses subcommand `reply` vs `reply-all` from a request bool. |
| `docs_get` | `tab`/`all_tabs` mutual exclusion + "tab provided but empty" check. |
| `docs_write` | `append`/`replace` cross-field logic with explicit failure mode. |
| `contacts_create` | "At least one of given/family/email" cross-field rule. |

If a migration task discovers another tool whose BuildArgs cannot be expressed with the grammar (e.g. a constant flag not tied to any request arg, or another cross-field rule), move it to `mcp_tools_custom.go` verbatim rather than extending the grammar, and note it in the task's commit message.

---

### Task 1: Annotation grammar parser

**Files:**
- Create: `internal/cmd/mcp_gen_annotation.go`
- Test: `internal/cmd/mcp_gen_annotation_test.go`

**Interfaces:**
- Produces: `mcpMountAnnotation{Name string; Risk mcpToolRisk; Service string}`, `parseMCPMountAnnotation(raw string) (mcpMountAnnotation, error)`, `mcpFieldAnnotation{Name string; Required, Optional bool; Default string; HasDefault bool; Enum []string; Min, Max *int; Text, JSON2D, OmitZero bool}`, `parseMCPFieldAnnotation(raw string) (mcpFieldAnnotation, error)`. `mcpToolRisk`/`mcpRiskRead`/`mcpRiskWrite` already exist in `internal/cmd/mcp.go:30`.

- [ ] **Step 1: Write the failing test**

```go
package cmd

import (
	"strings"
	"testing"
)

func TestParseMCPMountAnnotation(t *testing.T) {
	ann, err := parseMCPMountAnnotation("gmail_search,read")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ann.Name != "gmail_search" || ann.Risk != mcpRiskRead || ann.Service != "gmail" {
		t.Fatalf("ann = %#v", ann)
	}

	ann, err = parseMCPMountAnnotation("searchconsole_query,read,service=searchconsole")
	if err != nil {
		t.Fatalf("parse with service: %v", err)
	}
	if ann.Service != "searchconsole" {
		t.Fatalf("service = %q", ann.Service)
	}

	for _, bad := range []string{"", "gmail_search", "gmail_search,destroy", "noservicename,read", "gmail_search,read,bogus=1"} {
		if _, err := parseMCPMountAnnotation(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestParseMCPFieldAnnotation(t *testing.T) {
	ann, err := parseMCPFieldAnnotation("max,default=10,min=1,max=100")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ann.Name != "max" || !ann.HasDefault || ann.Default != "10" || *ann.Min != 1 || *ann.Max != 100 {
		t.Fatalf("ann = %#v", ann)
	}

	ann, err = parseMCPFieldAnnotation("render,enum=RAW|USER_ENTERED,default=USER_ENTERED")
	if err != nil {
		t.Fatalf("parse enum: %v", err)
	}
	if len(ann.Enum) != 2 || ann.Enum[0] != "RAW" || ann.Enum[1] != "USER_ENTERED" {
		t.Fatalf("enum = %#v", ann.Enum)
	}

	ann, err = parseMCPFieldAnnotation("body,required,text")
	if err != nil || !ann.Required || !ann.Text {
		t.Fatalf("ann = %#v, err = %v", ann, err)
	}

	ann, err = parseMCPFieldAnnotation("values_json,required,json2d")
	if err != nil || !ann.JSON2D {
		t.Fatalf("ann = %#v, err = %v", ann, err)
	}

	ann, err = parseMCPFieldAnnotation("days,default=0,min=0,max=31,omitzero")
	if err != nil || !ann.OmitZero {
		t.Fatalf("ann = %#v, err = %v", ann, err)
	}

	for _, bad := range []string{"", "x,required,optional", "x,min=2,max=1", "x,min=abc", "x,unknown"} {
		if _, err := parseMCPFieldAnnotation(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
	if _, err := parseMCPFieldAnnotation("x,bogus"); err == nil || !strings.Contains(err.Error(), "unknown option") {
		t.Fatalf("unknown option error = %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run 'TestParseMCP' -count=1`
Expected: FAIL — `undefined: parseMCPMountAnnotation`

- [ ] **Step 3: Write the implementation**

```go
package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

// mcpMountAnnotation is parsed from a command mount field's `mcp:"..."` tag
// (e.g. `mcp:"gmail_search,read"`) and marks a leaf CLI command for export as
// a typed MCP tool.
type mcpMountAnnotation struct {
	Name    string
	Risk    mcpToolRisk
	Service string
}

func parseMCPMountAnnotation(raw string) (mcpMountAnnotation, error) {
	var ann mcpMountAnnotation
	parts := strings.Split(raw, ",")
	if len(parts) < 2 {
		return ann, fmt.Errorf("mcp mount tag %q: want <tool_name>,<read|write>[,service=<name>]", raw)
	}
	ann.Name = strings.TrimSpace(parts[0])
	if ann.Name == "" {
		return ann, fmt.Errorf("mcp mount tag %q: empty tool name", raw)
	}
	switch strings.TrimSpace(parts[1]) {
	case "read":
		ann.Risk = mcpRiskRead
	case "write":
		ann.Risk = mcpRiskWrite
	default:
		return ann, fmt.Errorf("mcp mount tag %q: risk must be read or write", raw)
	}
	for _, part := range parts[2:] {
		part = strings.TrimSpace(part)
		if service, ok := strings.CutPrefix(part, "service="); ok && service != "" {
			ann.Service = service
			continue
		}
		return ann, fmt.Errorf("mcp mount tag %q: unknown option %q", raw, part)
	}
	if ann.Service == "" {
		idx := strings.Index(ann.Name, "_")
		if idx <= 0 {
			return ann, fmt.Errorf("mcp mount tag %q: cannot derive service from tool name, add service=", raw)
		}
		ann.Service = ann.Name[:idx]
	}
	return ann, nil
}

// mcpFieldAnnotation is parsed from a flag or positional field's `mcp:"..."`
// tag. The first item is the MCP property name; the rest are options that
// shape the schema and the generic argv emission.
type mcpFieldAnnotation struct {
	Name       string
	Required   bool
	Optional   bool
	Default    string
	HasDefault bool
	Enum       []string
	Min        *int
	Max        *int
	Text       bool
	JSON2D     bool
	OmitZero   bool
}

func parseMCPFieldAnnotation(raw string) (mcpFieldAnnotation, error) {
	var ann mcpFieldAnnotation
	parts := strings.Split(raw, ",")
	ann.Name = strings.TrimSpace(parts[0])
	if ann.Name == "" {
		return ann, fmt.Errorf("mcp field tag %q: first item must be the MCP property name", raw)
	}
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		switch {
		case part == "required":
			ann.Required = true
		case part == "optional":
			ann.Optional = true
		case part == "text":
			ann.Text = true
		case part == "json2d":
			ann.JSON2D = true
		case part == "omitzero":
			ann.OmitZero = true
		case strings.HasPrefix(part, "default="):
			ann.Default = strings.TrimPrefix(part, "default=")
			ann.HasDefault = true
		case strings.HasPrefix(part, "enum="):
			ann.Enum = strings.Split(strings.TrimPrefix(part, "enum="), "|")
		case strings.HasPrefix(part, "min="):
			n, err := strconv.Atoi(strings.TrimPrefix(part, "min="))
			if err != nil {
				return ann, fmt.Errorf("mcp field tag %q: invalid min: %w", raw, err)
			}
			ann.Min = &n
		case strings.HasPrefix(part, "max="):
			n, err := strconv.Atoi(strings.TrimPrefix(part, "max="))
			if err != nil {
				return ann, fmt.Errorf("mcp field tag %q: invalid max: %w", raw, err)
			}
			ann.Max = &n
		default:
			return ann, fmt.Errorf("mcp field tag %q: unknown option %q", raw, part)
		}
	}
	if ann.Required && ann.Optional {
		return ann, fmt.Errorf("mcp field tag %q: required and optional are mutually exclusive", raw)
	}
	if ann.Min != nil && ann.Max != nil && *ann.Min > *ann.Max {
		return ann, fmt.Errorf("mcp field tag %q: min greater than max", raw)
	}
	return ann, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run 'TestParseMCP' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/mcp_gen_annotation.go internal/cmd/mcp_gen_annotation_test.go
git commit -m "feat(mcp): parse mcp struct-tag annotation grammar"
```

---

### Task 2: Field synthesis — kong.Value → schema option + argv parameter

**Files:**
- Create: `internal/cmd/mcp_gen_field.go`
- Test: `internal/cmd/mcp_gen_field_test.go`

**Interfaces:**
- Consumes: `parseMCPFieldAnnotation` (Task 1).
- Produces: `mcpGenKind` (`mcpGenBool`, `mcpGenInt`, `mcpGenString`), `mcpGenParam{key, flag string; position int; kind mcpGenKind; required bool; def string; hasDef bool; min, max *int; text, json2d, omitZero bool}`, and `mcpGenField(v *kong.Value, isFlag bool) (*mcpGenParam, mcp.ToolOption, error)` — returns `(nil, nil, nil)` for values with no `mcp` tag.

- [ ] **Step 1: Write the failing test**

Build a real Kong model in the test so `kong.Value` instances are authentic:

```go
package cmd

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/kong"
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
	return parser.Model.Node.Children[0]
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
```

Note: the `Render` field has a Kong `default:"A"` but the mcp annotation has no `default=`, so the MCP schema gets no default and the param has `hasDef == false` — MCP defaults are always explicit in the annotation. Also note `Render` is optional in the schema even though it carries a Kong default.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run 'TestMCPGenField' -count=1`
Expected: FAIL — `undefined: mcpGenField`

- [ ] **Step 3: Write the implementation**

```go
package cmd

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
)

type mcpGenKind int

const (
	mcpGenBool mcpGenKind = iota
	mcpGenInt
	mcpGenString
)

// mcpGenParam is the argv-emission recipe for one annotated CLI value; the
// generic BuildArgs (mcp_gen_args.go) interprets a slice of these.
type mcpGenParam struct {
	key      string // MCP property name
	flag     string // "--<kong-name>"; empty for positionals
	position int    // ordering for positionals
	kind     mcpGenKind
	required bool
	def      string
	hasDef   bool
	min, max *int
	text     bool
	json2d   bool
	omitZero bool
}

// mcpGenField converts one annotated Kong value into its MCP schema option and
// argv parameter. Values without an `mcp` tag return (nil, nil, nil).
func mcpGenField(v *kong.Value, isFlag bool) (*mcpGenParam, mcp.ToolOption, error) {
	if v == nil || v.Tag == nil {
		return nil, nil, nil
	}
	raw := v.Tag.Get("mcp")
	if raw == "" {
		return nil, nil, nil
	}
	ann, err := parseMCPFieldAnnotation(raw)
	if err != nil {
		return nil, nil, err
	}
	kind, err := mcpGenKindOf(v)
	if err != nil {
		return nil, nil, fmt.Errorf("field %s: %w", ann.Name, err)
	}
	desc := v.Tag.Get("mcpdesc")
	if desc == "" {
		desc = v.Help
	}
	required := ann.Required || (!ann.Optional && !ann.HasDefault && v.Required)

	param := &mcpGenParam{
		key:      ann.Name,
		position: v.Position,
		kind:     kind,
		required: required,
		def:      ann.Default,
		hasDef:   ann.HasDefault,
		min:      ann.Min,
		max:      ann.Max,
		text:     ann.Text,
		json2d:   ann.JSON2D,
		omitZero: ann.OmitZero,
	}
	if isFlag {
		param.flag = "--" + v.Name
	}

	props := []mcp.PropertyOption{mcp.Description(desc)}
	if required {
		props = append(props, mcp.Required())
	}
	var opt mcp.ToolOption
	switch kind {
	case mcpGenBool:
		def := false
		if ann.HasDefault {
			switch ann.Default {
			case "true":
				def = true
			case "false":
			default:
				return nil, nil, fmt.Errorf("field %s: bool default must be true or false", ann.Name)
			}
		}
		props = append(props, mcp.DefaultBool(def))
		opt = mcp.WithBoolean(ann.Name, props...)
	case mcpGenInt:
		if !ann.HasDefault && !required {
			return nil, nil, fmt.Errorf("field %s: integer needs default= or required", ann.Name)
		}
		if ann.HasDefault {
			n, convErr := strconv.Atoi(ann.Default)
			if convErr != nil {
				return nil, nil, fmt.Errorf("field %s: invalid integer default: %w", ann.Name, convErr)
			}
			props = append(props, mcp.DefaultNumber(float64(n)))
		}
		if ann.Min != nil {
			props = append(props, mcp.Min(float64(*ann.Min)))
		}
		if ann.Max != nil {
			props = append(props, mcp.Max(float64(*ann.Max)))
		}
		opt = mcp.WithInteger(ann.Name, props...)
	case mcpGenString:
		if ann.HasDefault {
			props = append(props, mcp.DefaultString(ann.Default))
		}
		enum := ann.Enum
		if len(enum) == 0 && v.Enum != "" {
			enum = v.EnumSlice()
		}
		if len(enum) > 0 {
			props = append(props, mcp.Enum(enum...))
		}
		opt = mcp.WithString(ann.Name, props...)
	}
	return param, opt, nil
}

func mcpGenKindOf(v *kong.Value) (mcpGenKind, error) {
	t := v.Target.Type()
	for t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool:
		return mcpGenBool, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return mcpGenInt, nil
	case reflect.String:
		return mcpGenString, nil
	default:
		return 0, fmt.Errorf("unsupported field type %s", t.Kind())
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run 'TestMCPGenField' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/mcp_gen_field.go internal/cmd/mcp_gen_field_test.go
git commit -m "feat(mcp): synthesize schema options and argv params from annotated kong values"
```

---

### Task 3: Generic BuildArgs

**Files:**
- Create: `internal/cmd/mcp_gen_args.go`
- Test: `internal/cmd/mcp_gen_args_test.go`

**Interfaces:**
- Consumes: `mcpGenParam`/`mcpGenKind` (Task 2); `requireMCPString` (`mcp.go:319`), `requireMCPText`, `requireMCPLiteralValuesJSON` (`mcp_tools.go`).
- Produces: `mcpGenBuildArgs(path []string, params []mcpGenParam) func(mcp.CallToolRequest) ([]string, error)`.

- [ ] **Step 1: Write the failing test**

```go
package cmd

import (
	"reflect"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func genArgsRequest(arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "probe", Arguments: arguments}}
}

func intPtr(n int) *int { return &n }

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run 'TestMCPGenBuildArgs' -count=1`
Expected: FAIL — `undefined: mcpGenBuildArgs`

- [ ] **Step 3: Write the implementation**

```go
package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpGenBuildArgs returns a generic BuildArgs for an annotated command: the
// subcommand path, then annotated flags in declaration order, then positional
// values after `--`. Emission rules are documented in the plan's "Emission
// Semantics" and mirror the hand-written builders they replace.
func mcpGenBuildArgs(path []string, params []mcpGenParam) func(mcp.CallToolRequest) ([]string, error) {
	var flagParams, posParams []mcpGenParam
	for _, p := range params {
		if p.flag != "" {
			flagParams = append(flagParams, p)
		} else {
			posParams = append(posParams, p)
		}
	}
	sort.SliceStable(posParams, func(i, j int) bool { return posParams[i].position < posParams[j].position })

	return func(req mcp.CallToolRequest) ([]string, error) {
		args := append([]string(nil), path...)
		for i := range flagParams {
			toks, err := flagParams[i].tokens(req)
			if err != nil {
				return nil, err
			}
			args = append(args, toks...)
		}
		var pos []string
		for i := range posParams {
			value, err := posParams[i].stringValue(req)
			if err != nil {
				return nil, err
			}
			if value != "" {
				pos = append(pos, value)
			}
		}
		if len(pos) > 0 {
			args = append(args, "--")
			args = append(args, pos...)
		}
		return args, nil
	}
}

func (p *mcpGenParam) tokens(req mcp.CallToolRequest) ([]string, error) {
	switch p.kind {
	case mcpGenBool:
		if req.GetBool(p.key, p.hasDef && p.def == "true") {
			return []string{p.flag}, nil
		}
		return nil, nil
	case mcpGenInt:
		if _, present := req.GetArguments()[p.key]; !present && !p.hasDef {
			return nil, nil
		}
		def := 0
		if p.hasDef {
			def, _ = strconv.Atoi(p.def)
		}
		value := req.GetInt(p.key, def)
		if p.min != nil && value < *p.min {
			value = *p.min
		}
		if p.max != nil && value > *p.max {
			value = *p.max
		}
		if p.omitZero && value == 0 {
			return nil, nil
		}
		return []string{p.flag, strconv.Itoa(value)}, nil
	case mcpGenString:
		value, err := p.stringValue(req)
		if err != nil {
			return nil, err
		}
		if value == "" {
			return nil, nil
		}
		return []string{p.flag, value}, nil
	}
	return nil, fmt.Errorf("unsupported parameter kind for %s", p.key)
}

func (p *mcpGenParam) stringValue(req mcp.CallToolRequest) (string, error) {
	if p.kind != mcpGenString {
		return "", fmt.Errorf("parameter %s is not a string", p.key)
	}
	if p.json2d {
		if !p.required && strings.TrimSpace(req.GetString(p.key, "")) == "" {
			return "", nil
		}
		return requireMCPLiteralValuesJSON(req, p.key)
	}
	if p.required {
		if p.text {
			return requireMCPText(req, p.key)
		}
		return requireMCPString(req, p.key)
	}
	value := req.GetString(p.key, "")
	if !p.text {
		value = strings.TrimSpace(value)
	}
	if value == "" && p.hasDef {
		return p.def, nil
	}
	return value, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run 'TestMCPGenBuildArgs' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/mcp_gen_args.go internal/cmd/mcp_gen_args_test.go
git commit -m "feat(mcp): generic argv builder for annotated commands"
```

---

### Task 4: Model walker, registry integration, custom-tools home

**Files:**
- Create: `internal/cmd/mcp_gen.go`, `internal/cmd/mcp_tools_custom.go`
- Modify: `internal/cmd/mcp_tools.go` (mcpAllTools), `internal/cmd/mcp.go` (McpCmd.Run preflight)
- Test: `internal/cmd/mcp_gen_test.go`

**Interfaces:**
- Consumes: `parseMCPMountAnnotation` (Task 1), `mcpGenField` (Task 2), `mcpGenBuildArgs` (Task 3), `newParserWithWriters` (`root.go:482`).
- Produces: `mcpGeneratedTools() ([]mcpToolSpec, error)` (memoized), `mcpSpecsFromModel(root *kong.Node) ([]mcpToolSpec, error)`, `mcpCustomTools() []mcpToolSpec` (initially empty).

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run 'TestMCPSpecsFromModel|TestMCPGeneratedTools|TestMCPAllToolNamesUnique' -count=1`
Expected: FAIL — `undefined: mcpSpecsFromModel`

- [ ] **Step 3: Write the implementation**

`internal/cmd/mcp_gen.go`:

```go
package cmd

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	mcpGenOnce  sync.Once
	mcpGenSpecs []mcpToolSpec
	mcpGenErr   error
)

// mcpGeneratedTools derives MCP tool specs from mcp/mcpdesc annotations on the
// real command grammar. The result is memoized: the grammar is static for the
// lifetime of the process.
func mcpGeneratedTools() ([]mcpToolSpec, error) {
	mcpGenOnce.Do(func() {
		parser, _, err := newParserWithWriters("gog", io.Discard, io.Discard)
		if err != nil {
			mcpGenErr = fmt.Errorf("build command model: %w", err)
			return
		}
		mcpGenSpecs, mcpGenErr = mcpSpecsFromModel(parser.Model.Node)
	})
	return mcpGenSpecs, mcpGenErr
}

func mcpSpecsFromModel(root *kong.Node) ([]mcpToolSpec, error) {
	var specs []mcpToolSpec
	var walk func(node *kong.Node, path []string) error
	walk = func(node *kong.Node, path []string) error {
		for _, child := range node.Children {
			if child == nil || child.Type != kong.CommandNode {
				continue
			}
			childPath := append(append([]string(nil), path...), child.Name)
			raw := ""
			if child.Tag != nil {
				raw = child.Tag.Get("mcp")
			}
			if raw != "" {
				spec, err := mcpSpecFromNode(child, childPath, raw)
				if err != nil {
					return err
				}
				specs = append(specs, spec)
			}
			if err := walk(child, childPath); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, nil); err != nil {
		return nil, err
	}
	return specs, nil
}

func mcpSpecFromNode(node *kong.Node, path []string, raw string) (mcpToolSpec, error) {
	command := strings.Join(path, " ")
	mount, err := parseMCPMountAnnotation(raw)
	if err != nil {
		return mcpToolSpec{}, fmt.Errorf("command %q: %w", command, err)
	}
	for _, child := range node.Children {
		if child != nil && child.Type == kong.CommandNode {
			return mcpToolSpec{}, fmt.Errorf("mcp tool %s: command %q is not a leaf", mount.Name, command)
		}
	}
	description := node.Tag.Get("mcpdesc")
	if description == "" {
		description = node.Help
	}

	var opts []mcp.ToolOption
	var params []mcpGenParam
	seen := map[string]bool{}
	add := func(v *kong.Value, isFlag bool) error {
		param, opt, fieldErr := mcpGenField(v, isFlag)
		if fieldErr != nil {
			return fmt.Errorf("mcp tool %s: %w", mount.Name, fieldErr)
		}
		if param == nil {
			return nil
		}
		if seen[param.key] {
			return fmt.Errorf("mcp tool %s: duplicate property %q", mount.Name, param.key)
		}
		seen[param.key] = true
		params = append(params, *param)
		opts = append(opts, opt)
		return nil
	}
	for _, flag := range node.Flags {
		if flag == nil {
			continue
		}
		if err := add(flag.Value, true); err != nil {
			return mcpToolSpec{}, err
		}
	}
	for _, positional := range node.Positional {
		if err := add(positional, false); err != nil {
			return mcpToolSpec{}, err
		}
	}
	return mcpToolSpec{
		Name:        mount.Name,
		Service:     mount.Service,
		Risk:        mount.Risk,
		Description: description,
		Options:     opts,
		BuildArgs:   mcpGenBuildArgs(path, params),
	}, nil
}
```

`internal/cmd/mcp_tools_custom.go`:

```go
package cmd

// mcpCustomTools returns the hand-written tool specs whose request-to-argv
// mapping cannot be expressed with mcp/mcpdesc annotations (conditional
// subcommand selection, cross-field validation). Every other tool is generated
// from the annotated command grammar; see mcp_gen.go.
func mcpCustomTools() []mcpToolSpec {
	return nil
}
```

In `internal/cmd/mcp_tools.go`, change `mcpAllTools` to append custom + generated groups (keep the existing hand-written group calls for now):

```go
func mcpAllTools() []mcpToolSpec {
	generated, err := mcpGeneratedTools()
	if err != nil {
		// Annotation mistakes are programmer errors caught by tests; a broken
		// grammar must never serve a partial tool surface.
		panic(fmt.Sprintf("invalid mcp annotations: %v", err))
	}
	groups := [][]mcpToolSpec{
		mcpGmailTools(),
		mcpDriveTools(),
		// ... keep every existing group line unchanged ...
		mcpAPITools(),
		mcpCustomTools(),
	}
	groups = append(groups, generated)
	var all []mcpToolSpec
	for _, group := range groups {
		all = append(all, group...)
	}
	return all
}
```

(`fmt` is already imported in `mcp_tools.go`.)

In `internal/cmd/mcp.go` `McpCmd.Run`, add a preflight right after the `MaxOutputBytes` validation so a bad annotation yields a clean error instead of a panic:

```go
	if _, genErr := mcpGeneratedTools(); genErr != nil {
		return genErr
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cmd/ -run 'TestMCPSpecsFromModel|TestMCPGeneratedTools|TestMCPAllToolNamesUnique|TestMCPEnabledTools' -count=1`
Expected: PASS (generated set is empty so far; existing enabled-tools tests unaffected)

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/mcp_gen.go internal/cmd/mcp_gen_test.go internal/cmd/mcp_tools_custom.go internal/cmd/mcp_tools.go internal/cmd/mcp.go
git commit -m "feat(mcp): generate tool specs from annotated command grammar"
```

---

### Task 5: Golden snapshot of the tool surface

**Files:**
- Create: `internal/cmd/mcp_golden_test.go`, `internal/cmd/testdata/mcp_tools_golden.json`

**Interfaces:**
- Consumes: `mcpAllTools`, `newMCPTool` (`mcp.go:126`), `populatedRequest`/`skipPopulateKeys` (`mcp_validate_test.go`), `commandChild`/`commandFlag` (`arg_rewrite.go:78,105`), `newParserWithWriters`.
- Produces: `TestMCPToolSurfaceMatchesGolden` and the committed golden file. This is the oracle every migration task must keep green.

- [ ] **Step 1: Write the test**

```go
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
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s: unmarshal schema: %v", tool, err)
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
	if err := json.Unmarshal(b, &x); err != nil {
		t.Fatalf("canon unmarshal: %v", err)
	}
	out, err := json.Marshal(x)
	if err != nil {
		t.Fatalf("canon re-marshal: %v", err)
	}
	return string(out)
}
```

- [ ] **Step 2: Generate the golden file from the current hand-written surface**

```bash
mkdir -p internal/cmd/testdata
go test ./internal/cmd/ -run TestMCPToolSurfaceMatchesGolden -count=1 -update-mcp-golden
```

Expected: PASS (file written). Inspect: `python3 -c "import json;d=json.load(open('internal/cmd/testdata/mcp_tools_golden.json'));print(len(d))"` → `153`.

- [ ] **Step 3: Verify the test passes against the committed golden**

Run: `go test ./internal/cmd/ -run TestMCPToolSurfaceMatchesGolden -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/cmd/mcp_golden_test.go internal/cmd/testdata/mcp_tools_golden.json
git commit -m "test(mcp): golden snapshot of the full MCP tool surface"
```

**After this commit the golden file is frozen. No later task may pass `-update-mcp-golden`.**

---

### Migration task template (applies to Tasks 6–10)

Each migration task converts one batch of services from hand-written specs to annotations. Procedure per tool:

1. Open the tool's hand-written spec (file listed in the task). From its `BuildArgs` argv, identify the CLI command: follow the path tokens through the command mounts starting at `CLI` in `internal/cmd/root.go` (e.g. `{"gmail","messages","search"}` → `GmailCmd.Messages` → `GmailMessagesCmd.Search` → `GmailMessagesSearchCmd`).
2. On the mount field add `mcp:"<tool>,<risk>"` and `mcpdesc:"<Description verbatim>"`.
3. For each schema option, find the struct field whose Kong `name:` matches the flag that `BuildArgs` emits for it (or the positional it maps to). Add `mcp:"<mcp_property_name>[,options]"` and `mcpdesc:"<option Description verbatim>"`. Translation table:
   - `mcp.Required()` → `required` (omit when the field is a Kong-required positional — that's automatic; check the golden if unsure).
   - `mcp.DefaultNumber(N)`/`DefaultString(S)` → `default=N`. `mcp.DefaultBool(false)` is automatic; `DefaultBool(true)` → `default=true`.
   - `mcp.Min(N)`/`mcp.Max(N)` → `min=N`/`max=N`.
   - `mcp.Enum(...)` → automatic when the CLI field has a matching `enum:` tag; otherwise `enum=A|B|C`.
   - `.num(key, flag, def, hi)` in fluent builders → `default=def,min=1,max=hi`.
   - Hand-written `if v > 0 { append(...) }` int guards → add `omitzero`.
   - Values read with `requireMCPText` (bodies, notes — whitespace preserved) → `text`.
   - Values read with `requireMCPLiteralValuesJSON` → `json2d`.
   - MCP property names that differ from the flag name (e.g. `message_id` for a positional) are simply the annotation's first item.
4. Tools listed as custom for the batch: move their spec functions **verbatim** into `internal/cmd/mcp_tools_custom.go` and add them to `mcpCustomTools()`.
5. Delete the migrated hand-written spec functions and their entries in the service group function; when a group function is empty, delete it and its line in `mcpAllTools`. Delete files that become empty.
6. Verify:

```bash
go test ./internal/cmd/ -run 'TestMCPToolSurfaceMatchesGolden|TestMCPToolArgsResolveToRealCommands|TestMCPAllToolNamesUnique' -count=1
```

Expected: PASS. A golden diff means an annotation is wrong — fix the annotation, never the golden. Then run the full package suite and lint before committing:

```bash
go test ./internal/cmd/ -count=1 && make fmt && make lint
```

7. Commit: `git add -A && git commit -m "refactor(mcp): generate <services> tools from annotations"`.

**Worked example** (part of Task 6, shown here for every task's reference). Hand-written `mcpGmailSearchThreadsTool` (`mcp_tools_gmail.go:31`) has argv path `gmail search`, options `query` (required), `max` (default 10, min 1, max 100), `oldest` (bool). The CLI command is `GmailCmd.Search` → `GmailSearchCmd` (`gmail_search.go:15`). Annotations:

```go
// internal/cmd/gmail.go — mount field:
Search GmailSearchCmd `cmd:"" name:"search" aliases:"find,query,ls,list" group:"Read" help:"Search threads using Gmail query syntax" mcp:"gmail_search_threads,read" mcpdesc:"Search Gmail threads with Gmail query syntax. Returns thread summaries."`

// internal/cmd/gmail_search.go — annotated fields (others untouched):
Query  []string `arg:"" name:"query" help:"Search query" mcp:"query" mcpdesc:"Gmail search query, e.g. newer_than:7d from:person@example.com"`
Max    int64    `name:"max" aliases:"limit" help:"Max results" default:"10" mcp:"max,default=10,min=1,max=100" mcpdesc:"Maximum results"`
Oldest bool     `name:"oldest" help:"Show first message date instead of last" mcp:"oldest" mcpdesc:"Sort oldest first"`
```

Then delete `mcpGmailSearchThreadsTool` and its registration line in `mcpGmailTools()`.

---

### Task 6: Migrate Gmail (pilot, 18 tools)

**Files:**
- Modify: `internal/cmd/gmail.go`, `internal/cmd/gmail_messages.go`, `internal/cmd/gmail_search.go`, `internal/cmd/gmail_send.go`, `internal/cmd/gmail_drafts.go`, `internal/cmd/gmail_labels.go`, `internal/cmd/gmail_thread.go`, `internal/cmd/gmail_get.go`, `internal/cmd/gmail_history.go` (actual struct locations may vary — find them by following mounts from `GmailCmd`), `internal/cmd/mcp_tools.go`, `internal/cmd/mcp_tools_gmail.go` (delete at the end), `internal/cmd/mcp_tools_custom.go`
- Test oracle: `internal/cmd/testdata/mcp_tools_golden.json` (frozen)

**Interfaces:**
- Consumes: the migration template above; annotation grammar; `mcpCustomTools()` from Task 4.
- Produces: all 18 Gmail tools served from annotations except `gmail_reply`, which moves verbatim (with its `reply`/`reply-all` switch) into `mcp_tools_custom.go`.

- [ ] **Step 1:** Read the Annotation Grammar Reference, Emission Semantics, and migration template sections of this plan. Read `internal/cmd/mcp_tools_gmail.go` and the Gmail entries in `internal/cmd/mcp_tools.go` (`gmail_search` at `mcp_tools.go:99`, `gmail_get_message`, `gmail_get_thread`).
- [ ] **Step 2:** Migrate the generated 17: `gmail_search` (path `gmail messages search`), `gmail_search_threads` (`gmail search`, worked example above), `gmail_get_message` (`gmail get`; note `sanitize_content` → `default=true`), `gmail_get_thread` (`gmail thread get`; `sanitize_content` → `default=true`), `gmail_list_labels`, `gmail_get_label`, `gmail_list_drafts`, `gmail_get_draft`, `gmail_history`, `gmail_send` (`to`/`subject` required; `subject` and `body` use `requireMCPText` → `text`; note `subject`+`body` required + `text`), `gmail_forward`, `gmail_create_draft`, `gmail_send_draft`, `gmail_modify_message`, `gmail_modify_thread`, `gmail_trash`, `gmail_create_label`. Apply the template per tool; run the golden test after each few tools to catch translation errors early.
- [ ] **Step 3:** Move `mcpGmailReplyTool` verbatim into `mcp_tools_custom.go`, register it in `mcpCustomTools()`, and remove it from `mcpGmailTools()`.
- [ ] **Step 4:** Delete `mcpGmailTools()` and all remaining Gmail spec functions (`mcp_tools_gmail.go` should end up deleted; the three Gmail specs in `mcp_tools.go` deleted; the `mcpGmailTools(),` line removed from `mcpAllTools`).
- [ ] **Step 5:** Verify per template step 6 (golden + validate + unique + full package suite + lint).
- [ ] **Step 6:** Commit: `refactor(mcp): generate gmail tools from annotations`.

---

### Task 7: Migrate Drive, Docs, Sheets, Slides (~36 tools)

**Files:**
- Modify: the Drive/Docs/Sheets/Slides command-struct files (find via mounts from `DriveCmd`, `DocsCmd`, `SheetsCmd`, `SlidesCmd` in `root.go`), `internal/cmd/mcp_tools.go`, `internal/cmd/mcp_tools_custom.go`
- Delete when empty: `internal/cmd/mcp_tools_drive.go`, `internal/cmd/mcp_tools_docs.go`, `internal/cmd/mcp_tools_sheets.go`, `internal/cmd/mcp_tools_slides.go`

**Interfaces:**
- Consumes: migration template; `json2d` annotation for Sheets value writes.
- Produces: all Drive (15), Docs (8), Sheets (9), Slides (8) tools annotation-generated except `docs_get` and `docs_write` (custom).

- [ ] **Step 1:** Read the template sections plus `mcp_tools_drive.go`, `mcp_tools_docs.go`, `mcp_tools_sheets.go`, `mcp_tools_slides.go`, and the Drive/Docs/Sheets specs inside `mcp_tools.go` (`drive_search`, `drive_get`, `docs_get`, `sheets_read_range`, `docs_write`, `sheets_update_range`).
- [ ] **Step 2:** Move `mcpDocsGetTool` and `mcpDocsWriteTool` verbatim to `mcp_tools_custom.go` (they carry cross-field rules), register in `mcpCustomTools()`. Keep their `skipPopulateKeys` entries in `mcp_validate_test.go` untouched.
- [ ] **Step 3:** Migrate the rest per template. Notes: `sheets_update_range` and `sheets_append` use `values_json` → `mcp:"values_json,required,json2d"` and `input` → `mcp:"input,default=USER_ENTERED"` with `enum=RAW|USER_ENTERED` if the CLI field lacks a matching `enum:` tag; `mcpSheetsValuesArgs` and both hand-written sheets write specs are then deletable. `drive_search`'s `parent` is a plain optional string.
- [ ] **Step 4:** Delete migrated spec functions, group functions, empty files, and their `mcpAllTools` lines.
- [ ] **Step 5:** Verify per template step 6.
- [ ] **Step 6:** Commit: `refactor(mcp): generate drive, docs, sheets, slides tools from annotations`.

---

### Task 8: Migrate Calendar, Contacts, People, Tasks, Chat, Keep, Meet, Forms (~47 tools)

**Files:**
- Modify: the corresponding command-struct files (find via mounts in `root.go`), `internal/cmd/mcp_tools.go`, `internal/cmd/mcp_tools_custom.go`
- Delete when empty: `internal/cmd/mcp_tools_calendar.go`, `internal/cmd/mcp_tools_workspace.go`

**Interfaces:**
- Consumes: migration template; `omitzero` for `calendar_events.days`.
- Produces: Calendar (9), Contacts (7), People (3), Tasks (8), Chat (6), Keep (4), Meet (4), Forms (6) generated, except `contacts_create` (custom).

- [ ] **Step 1:** Read the template sections plus `mcp_tools_calendar.go`, `mcp_tools_workspace.go`, and `calendar_events` in `mcp_tools.go:296`.
- [ ] **Step 2:** Move `contacts_create` (the spec with the "provide at least one of given, family, or email" rule at `mcp_tools_workspace.go:120`) verbatim to `mcp_tools_custom.go`.
- [ ] **Step 3:** Migrate the rest per template. Notes: `calendar_events` — `days` gets `mcp:"days,default=0,min=0,max=31,omitzero"`, `max` gets `default=10,min=1,max=250`, `calendar_id` is an optional positional; `today`/`tomorrow` are plain bools.
- [ ] **Step 4:** Delete migrated spec functions, group functions, empty files, and their `mcpAllTools` lines.
- [ ] **Step 5:** Verify per template step 6.
- [ ] **Step 6:** Commit: `refactor(mcp): generate calendar and workspace service tools from annotations`.

---

### Task 9: Migrate Classroom, Photos, Maps, YouTube, Search Console (~30 tools)

**Files:**
- Modify: the corresponding command-struct files, `internal/cmd/mcp_tools.go`
- Delete when empty: `internal/cmd/mcp_tools_more.go`

**Interfaces:**
- Consumes: migration template. `searchconsole_*` tools need `service=searchconsole` only if the tool name prefix doesn't already derive it (it does — no override needed).
- Produces: Classroom (10), Photos (3), Maps (6), YouTube (8), Search Console (3) generated; no customs expected.

- [ ] **Step 1:** Read the template sections plus `mcp_tools_more.go`.
- [ ] **Step 2:** Migrate all tools per template. If any tool turns out to be inexpressible (constant flags, conditional paths), move it to `mcp_tools_custom.go` verbatim and say so in the commit message.
- [ ] **Step 3:** Delete migrated spec functions, group functions, the file when empty, and the `mcpAllTools` lines.
- [ ] **Step 4:** Verify per template step 6.
- [ ] **Step 5:** Commit: `refactor(mcp): generate classroom, photos, maps, youtube, searchconsole tools from annotations`.

---

### Task 10: Migrate Admin, Groups, Apps Script, Discovery API (~18 tools)

**Files:**
- Modify: the corresponding command-struct files, `internal/cmd/mcp_tools.go`
- Delete when empty: `internal/cmd/mcp_tools_admin.go`, `internal/cmd/mcp_tools_developer.go`

**Interfaces:**
- Consumes: migration template.
- Produces: Admin (9), Groups (2), Apps Script (4), API (3) generated; no customs expected.

- [ ] **Step 1:** Read the template sections plus `mcp_tools_admin.go` and `mcp_tools_developer.go`.
- [ ] **Step 2:** Migrate all tools per template (same escape-hatch rule as Task 9).
- [ ] **Step 3:** Delete migrated spec functions, group functions, empty files, and the `mcpAllTools` lines.
- [ ] **Step 4:** Verify per template step 6.
- [ ] **Step 5:** Commit: `refactor(mcp): generate admin, groups, appscript, api tools from annotations`.

---

### Task 11: Cleanup, docs, changelog, full verification

**Files:**
- Modify: `internal/cmd/mcp_tools.go` (final shape), `docs/mcp.md`, `CHANGELOG.md`
- Possibly delete: unused helpers flagged by deadcode

**Interfaces:**
- Consumes: everything prior. After Tasks 6–10, `mcpAllTools` should reduce to `mcpCustomTools()` + generated.

- [ ] **Step 1:** Simplify `mcpAllTools` to its final shape (update the doc comment accordingly):

```go
// mcpAllTools returns the full typed MCP tool surface: specs generated from
// mcp/mcpdesc annotations on the command grammar (mcp_gen.go) plus the few
// hand-written specs in mcp_tools_custom.go. --allow-tool selectors (gmail,
// drive.*, ...) map onto each spec's Service; --allow-write gates Risk.
func mcpAllTools() []mcpToolSpec {
	generated, err := mcpGeneratedTools()
	if err != nil {
		// Annotation mistakes are programmer errors caught by tests; a broken
		// grammar must never serve a partial tool surface.
		panic(fmt.Sprintf("invalid mcp annotations: %v", err))
	}
	return append(append([]mcpToolSpec(nil), mcpCustomTools()...), generated...)
}
```

- [ ] **Step 2:** Run `make deadcode` and remove now-unused helpers (e.g. the fluent `mcpArgs` builder in `mcp_tools.go` if no custom tool uses it — `gmail_reply` and `contacts_create` likely still do; keep whatever is referenced). `requireMCPString`, `requireMCPText`, `requireMCPLiteralValuesJSON`, `clampMCPInt` stay (generator/custom use them — verify `clampMCPInt` is still referenced; if only tests use it, inline or delete).
- [ ] **Step 3:** Add to `docs/mcp.md`, at the end of the "Tools by area" section, this paragraph:

```markdown
Tool definitions are generated from annotations on the CLI command grammar, so
each tool's schema (types, defaults, enums, bounds) cannot drift from the
command it invokes. A handful of tools with cross-field rules remain
hand-written. The registered surface is unchanged by this mechanism; it remains
fixed, typed, and allowlisted as described above.
```

- [ ] **Step 4:** Add a `CHANGELOG.md` entry under the current unreleased heading (match the file's existing format): `- MCP: tool specs are now generated from command-grammar annotations (identical tool surface, verified by golden snapshot).`
- [ ] **Step 5:** Full verification:

```bash
make fmt && make lint && make deadcode
go test ./... -count=1
```

Expected: all PASS. The golden test still passes with 153 tools, `git grep -l 'mcpToolSpec{' internal/cmd/` shows only `mcp_tools_custom.go` (plus `mcp_gen.go`'s constructor and tests).

- [ ] **Step 6:** Commit: `refactor(mcp): finish annotation-generated tool surface; docs and changelog`.

---

## Self-Review Notes

- Spec coverage: grammar (Tasks 1–3), generation+registry (Task 4), no-regression oracle (Task 5), full migration (Tasks 6–10), cleanup/docs (Task 11). The four cross-field tools are explicitly routed to `mcp_tools_custom.go`.
- Type consistency: `mcpGenParam` fields (`key/flag/position/kind/required/def/hasDef/min/max/text/json2d/omitZero`) are used identically in Tasks 2, 3; `mcpSpecsFromModel`/`mcpGeneratedTools`/`mcpCustomTools` names match across Tasks 4, 5, 11.
- Known cosmetic changes (accepted): tool ordering in `--list-tools` output and argv flag ordering may differ; the golden test canonicalizes both, and MCP clients are order-insensitive.
