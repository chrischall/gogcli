package cmd

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpSheetsTools returns the Google Sheets MCP tool surface (service "sheets").
func mcpSheetsTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpSheetsReadRangeTool(),
		mcpSheetsMetadataTool(),
		mcpSheetsListTablesTool(),
		mcpSheetsListNamedRangesTool(),
		mcpSheetsUpdateRangeTool(),
		mcpSheetsAppendTool(),
		mcpSheetsClearTool(),
		mcpSheetsCreateTool(),
		mcpSheetsAddTabTool(),
	}
}

func mcpSheetsMetadataTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_metadata",
		Service:     "sheets",
		Risk:        mcpRiskRead,
		Description: "Get spreadsheet metadata (sheets, titles, dimensions).",
		Options: []mcp.ToolOption{
			mcp.WithString("spreadsheet_id", mcp.Description("Google Sheets spreadsheet ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			spreadsheetID, err := requireMCPString(req, "spreadsheet_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "sheets", "metadata").done(spreadsheetID)
		},
	}
}

func mcpSheetsListTablesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_list_tables",
		Service:     "sheets",
		Risk:        mcpRiskRead,
		Description: "List tables defined in a spreadsheet.",
		Options: []mcp.ToolOption{
			mcp.WithString("spreadsheet_id", mcp.Description("Google Sheets spreadsheet ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			spreadsheetID, err := requireMCPString(req, "spreadsheet_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "sheets", "table", "list").done(spreadsheetID)
		},
	}
}

func mcpSheetsListNamedRangesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_list_named_ranges",
		Service:     "sheets",
		Risk:        mcpRiskRead,
		Description: "List named ranges in a spreadsheet.",
		Options: []mcp.ToolOption{
			mcp.WithString("spreadsheet_id", mcp.Description("Google Sheets spreadsheet ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			spreadsheetID, err := requireMCPString(req, "spreadsheet_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "sheets", "named-ranges", "list").done(spreadsheetID)
		},
	}
}

func mcpSheetsAppendTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_append",
		Service:     "sheets",
		Risk:        mcpRiskWrite,
		Description: "Append rows to a Google Sheets range from a literal JSON 2D array. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("spreadsheet_id", mcp.Description("Google Sheets spreadsheet ID"), mcp.Required()),
			mcp.WithString("range", mcp.Description("A1 notation or named range"), mcp.Required()),
			mcp.WithString("values_json", mcp.Description("JSON 2D array of values"), mcp.Required()),
			mcp.WithString("input", mcp.Description("Value input option"), mcp.Enum("RAW", "USER_ENTERED"), mcp.DefaultString("USER_ENTERED")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			spreadsheetID, err := requireMCPString(req, "spreadsheet_id")
			if err != nil {
				return nil, err
			}
			rangeSpec, err := requireMCPString(req, "range")
			if err != nil {
				return nil, err
			}
			valuesJSON, err := requireMCPLiteralValuesJSON(req, "values_json")
			if err != nil {
				return nil, err
			}
			input := strings.TrimSpace(req.GetString("input", "USER_ENTERED"))
			if input == "" {
				input = "USER_ENTERED"
			}
			return []string{"sheets", "append", "--values-json", valuesJSON, "--input", input, "--", spreadsheetID, rangeSpec}, nil
		},
	}
}

func mcpSheetsClearTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_clear",
		Service:     "sheets",
		Risk:        mcpRiskWrite,
		Description: "Clear values in a Google Sheets range. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("spreadsheet_id", mcp.Description("Google Sheets spreadsheet ID"), mcp.Required()),
			mcp.WithString("range", mcp.Description("A1 notation or named range"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			spreadsheetID, err := requireMCPString(req, "spreadsheet_id")
			if err != nil {
				return nil, err
			}
			rangeSpec, err := requireMCPString(req, "range")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "sheets", "clear").done(spreadsheetID, rangeSpec)
		},
	}
}

func mcpSheetsCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_create",
		Service:     "sheets",
		Risk:        mcpRiskWrite,
		Description: "Create a new spreadsheet. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Spreadsheet title"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Parent folder ID")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPString(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "sheets", "create").
				str("parent", "--parent").done(title)
		},
	}
}

func mcpSheetsAddTabTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "sheets_add_tab",
		Service:     "sheets",
		Risk:        mcpRiskWrite,
		Description: "Add a new tab/sheet to a spreadsheet. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("spreadsheet_id", mcp.Description("Google Sheets spreadsheet ID"), mcp.Required()),
			mcp.WithString("tab_name", mcp.Description("New tab name"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			spreadsheetID, err := requireMCPString(req, "spreadsheet_id")
			if err != nil {
				return nil, err
			}
			tabName, err := requireMCPString(req, "tab_name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "sheets", "add-tab").done(spreadsheetID, tabName)
		},
	}
}
