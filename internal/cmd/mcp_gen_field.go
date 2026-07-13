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
			case boolTrue:
				def = true
			case boolFalse:
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
