package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpDocsTools returns the Google Docs MCP tool surface (service "docs").
func mcpDocsTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpDocsGetTool(),
		mcpDocsInfoTool(),
		mcpDocsListTabsTool(),
		mcpDocsListCommentsTool(),
		mcpDocsWriteTool(),
		mcpDocsCreateTool(),
		mcpDocsFindReplaceTool(),
		mcpDocsAddCommentTool(),
	}
}

func mcpDocsInfoTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_info",
		Service:     "docs",
		Risk:        mcpRiskRead,
		Description: "Get Google Doc metadata (title, revision, tabs).",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "docs", "info").done(docID)
		},
	}
}

func mcpDocsListTabsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_list_tabs",
		Service:     "docs",
		Risk:        mcpRiskRead,
		Description: "List all tabs in a Google Doc with their IDs.",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "docs", "list-tabs").done(docID)
		},
	}
}

func mcpDocsListCommentsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_list_comments",
		Service:     "docs",
		Risk:        mcpRiskRead,
		Description: "List comments on a Google Doc.",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
			mcp.WithBoolean("include_resolved", mcp.Description("Include resolved comments"), mcp.DefaultBool(false)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "docs", "comments", "list").
				flag("include_resolved", "--include-resolved").
				num("max", "--max", 50, 1, 200).done(docID)
		},
	}
}

func mcpDocsCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_create",
		Service:     "docs",
		Risk:        mcpRiskWrite,
		Description: "Create a new Google Doc. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Document title"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Parent folder ID")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPString(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "docs", "create").
				str("parent", "--parent").done(title)
		},
	}
}

func mcpDocsFindReplaceTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_find_replace",
		Service:     "docs",
		Risk:        mcpRiskWrite,
		Description: "Find and replace text in a Google Doc. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
			mcp.WithString("find", mcp.Description("Text to find"), mcp.Required()),
			mcp.WithString("replace", mcp.Description("Replacement text"), mcp.Required()),
			mcp.WithBoolean("match_case", mcp.Description("Case-sensitive match"), mcp.DefaultBool(false)),
			mcp.WithBoolean("first", mcp.Description("Replace only the first occurrence"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			find, err := requireMCPText(req, "find")
			if err != nil {
				return nil, err
			}
			replace, err := requireMCPText(req, "replace")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "docs", "find-replace").
				flag("match_case", "--match-case").
				flag("first", "--first").done(docID, find, replace)
		},
	}
}

func mcpDocsAddCommentTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_add_comment",
		Service:     "docs",
		Risk:        mcpRiskWrite,
		Description: "Add a comment to a Google Doc. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
			mcp.WithString("content", mcp.Description("Comment content"), mcp.Required()),
			mcp.WithString("quoted", mcp.Description("Optional quoted text the comment anchors to")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			content, err := requireMCPText(req, "content")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "docs", "comments", "add").
				str("quoted", "--quoted").done(docID, content)
		},
	}
}
