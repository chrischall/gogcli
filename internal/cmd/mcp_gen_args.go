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
		if req.GetBool(p.key, p.hasDef && p.def == boolTrue) {
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
