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
