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
