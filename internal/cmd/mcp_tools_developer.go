package cmd

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpAppScriptTools returns the Google Apps Script MCP tool surface (service
// "appscript").
func mcpAppScriptTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpAppScriptGetTool(),
		mcpAppScriptContentTool(),
		mcpAppScriptCreateTool(),
		mcpAppScriptRunTool(),
	}
}

func mcpAppScriptGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "appscript_get",
		Service:     "appscript",
		Risk:        mcpRiskRead,
		Description: "Get Apps Script project metadata by script ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("script_id", mcp.Description("Apps Script project ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			scriptID, err := requireMCPString(req, "script_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "appscript", "get").done(scriptID)
		},
	}
}

func mcpAppScriptContentTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "appscript_content",
		Service:     "appscript",
		Risk:        mcpRiskRead,
		Description: "Get Apps Script project source content by script ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("script_id", mcp.Description("Apps Script project ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			scriptID, err := requireMCPString(req, "script_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "appscript", "content").done(scriptID)
		},
	}
}

func mcpAppScriptCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "appscript_create",
		Service:     "appscript",
		Risk:        mcpRiskWrite,
		Description: "Create an Apps Script project. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Project title"), mcp.Required()),
			mcp.WithString("parent_id", mcp.Description("Parent Drive file ID to bind the script to")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPText(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "appscript", "create", "--title", title).
				str("parent_id", "--parent-id").done()
		},
	}
}

func mcpAppScriptRunTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "appscript_run",
		Service:     "appscript",
		Risk:        mcpRiskWrite,
		Description: "Run a deployed Apps Script function. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("script_id", mcp.Description("Apps Script project ID"), mcp.Required()),
			mcp.WithString("function", mcp.Description("Function name to run"), mcp.Required()),
			mcp.WithString("params", mcp.Description("JSON array of function parameters")),
			mcp.WithBoolean("dev_mode", mcp.Description("Run the latest saved (dev) version"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			scriptID, err := requireMCPString(req, "script_id")
			if err != nil {
				return nil, err
			}
			function, err := requireMCPString(req, "function")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "appscript", "run").
				str("params", "--params").
				flag("dev_mode", "--dev-mode").done(scriptID, function)
		},
	}
}

// mcpAPITools returns the Google Discovery API MCP tool surface (service "api").
// Discovery and describe are read-only; the generic method caller is a write
// tool and stays typed (api, version, method, params/body) rather than a free
// command bridge.
func mcpAPITools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpAPIListTool(),
		mcpAPIDescribeTool(),
		mcpAPICallTool(),
	}
}

func mcpAPIListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "api_list",
		Service:     "api",
		Risk:        mcpRiskRead,
		Description: "List Google Discovery APIs.",
		Options: []mcp.ToolOption{
			mcp.WithBoolean("all", mcp.Description("Include preview and non-preferred versions"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "api", "list").
				flag("all", "--all").done()
		},
	}
}

func mcpAPIDescribeTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "api_describe",
		Service:     "api",
		Risk:        mcpRiskRead,
		Description: "Describe a Discovery API or one of its methods.",
		Options: []mcp.ToolOption{
			mcp.WithString("api", mcp.Description("API name, e.g. drive"), mcp.Required()),
			mcp.WithString("version", mcp.Description("API version, e.g. v3"), mcp.Required()),
			mcp.WithString("method", mcp.Description("Optional method ID, e.g. drive.files.list")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			api, err := requireMCPString(req, "api")
			if err != nil {
				return nil, err
			}
			version, err := requireMCPString(req, "version")
			if err != nil {
				return nil, err
			}
			pos := []string{api, version}
			if method := strings.TrimSpace(req.GetString("method", "")); method != "" {
				pos = append(pos, method)
			}
			return mcpCommand(req, "api", "describe").done(pos...)
		},
	}
}

func mcpAPICallTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "api_call",
		Service:     "api",
		Risk:        mcpRiskWrite,
		Description: "Call a Discovery-described Google API method. Requires --allow-write; set write=true to permit mutating methods.",
		Options: []mcp.ToolOption{
			mcp.WithString("api", mcp.Description("API name, e.g. drive"), mcp.Required()),
			mcp.WithString("version", mcp.Description("API version, e.g. v3"), mcp.Required()),
			mcp.WithString("method", mcp.Description("Method ID, e.g. drive.files.list"), mcp.Required()),
			mcp.WithString("params", mcp.Description("JSON object of method parameters")),
			mcp.WithString("body", mcp.Description("JSON request body for write methods")),
			mcp.WithString("scope", mcp.Description("Override OAuth scope")),
			mcp.WithBoolean("write", mcp.Description("Permit mutating (non-GET) methods"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			api, err := requireMCPString(req, "api")
			if err != nil {
				return nil, err
			}
			version, err := requireMCPString(req, "version")
			if err != nil {
				return nil, err
			}
			method, err := requireMCPString(req, "method")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "api", "call").
				str("params", "--params").
				str("body", "--body").
				str("scope", "--scope").
				flag("write", "--allow-write").done(api, version, method)
		},
	}
}
