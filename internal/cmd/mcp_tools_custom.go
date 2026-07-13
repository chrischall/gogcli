package cmd

// mcpCustomTools returns the hand-written tool specs whose request-to-argv
// mapping cannot be expressed with mcp/mcpdesc annotations (conditional
// subcommand selection, cross-field validation). Every other tool is generated
// from the annotated command grammar; see mcp_gen.go.
func mcpCustomTools() []mcpToolSpec {
	return nil
}
