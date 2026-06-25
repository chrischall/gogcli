package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpSlidesTools returns the Google Slides MCP tool surface (service "slides").
func mcpSlidesTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpSlidesInfoTool(),
		mcpSlidesListSlidesTool(),
		mcpSlidesReadSlideTool(),
		mcpSlidesCreateTool(),
		mcpSlidesCreateFromMarkdownTool(),
		mcpSlidesReplaceTextTool(),
		mcpSlidesDuplicateSlideTool(),
		mcpSlidesDeleteSlideTool(),
	}
}

func mcpSlidesInfoTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_info",
		Service:     "slides",
		Risk:        mcpRiskRead,
		Description: "Get Google Slides presentation metadata.",
		Options: []mcp.ToolOption{
			mcp.WithString("presentation_id", mcp.Description("Google Slides presentation ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			presentationID, err := requireMCPString(req, "presentation_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "info").done(presentationID)
		},
	}
}

func mcpSlidesListSlidesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_list_slides",
		Service:     "slides",
		Risk:        mcpRiskRead,
		Description: "List all slides in a presentation with their object IDs.",
		Options: []mcp.ToolOption{
			mcp.WithString("presentation_id", mcp.Description("Google Slides presentation ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			presentationID, err := requireMCPString(req, "presentation_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "list-slides").done(presentationID)
		},
	}
}

func mcpSlidesReadSlideTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_read_slide",
		Service:     "slides",
		Risk:        mcpRiskRead,
		Description: "Read a slide's text, speaker notes, and images.",
		Options: []mcp.ToolOption{
			mcp.WithString("presentation_id", mcp.Description("Google Slides presentation ID"), mcp.Required()),
			mcp.WithString("slide_id", mcp.Description("Slide object ID"), mcp.Required()),
			mcp.WithBoolean("detail", mcp.Description("Include element-level detail"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			presentationID, err := requireMCPString(req, "presentation_id")
			if err != nil {
				return nil, err
			}
			slideID, err := requireMCPString(req, "slide_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "read-slide").
				flag("detail", "--detail").done(presentationID, slideID)
		},
	}
}

func mcpSlidesCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_create",
		Service:     "slides",
		Risk:        mcpRiskWrite,
		Description: "Create a Google Slides presentation. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Presentation title"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Parent folder ID")),
			mcp.WithString("template", mcp.Description("Template presentation ID to base on")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPString(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "create").
				str("parent", "--parent").
				str("template", "--template").done(title)
		},
	}
}

func mcpSlidesCreateFromMarkdownTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_create_from_markdown",
		Service:     "slides",
		Risk:        mcpRiskWrite,
		Description: "Create a Google Slides presentation from markdown content. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Presentation title"), mcp.Required()),
			mcp.WithString("content", mcp.Description("Markdown content (slides separated by ---)"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Parent folder ID")),
			mcp.WithBoolean("no_notes", mcp.Description("Do not generate speaker notes"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPString(req, "title")
			if err != nil {
				return nil, err
			}
			content, err := requireMCPText(req, "content")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "create-from-markdown", "--content", content).
				str("parent", "--parent").
				flag("no_notes", "--no-notes").done(title)
		},
	}
}

func mcpSlidesReplaceTextTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_replace_text",
		Service:     "slides",
		Risk:        mcpRiskWrite,
		Description: "Find and replace text across a presentation. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("presentation_id", mcp.Description("Google Slides presentation ID"), mcp.Required()),
			mcp.WithString("find", mcp.Description("Text to find"), mcp.Required()),
			mcp.WithString("replacement", mcp.Description("Replacement text"), mcp.Required()),
			mcp.WithBoolean("match_case", mcp.Description("Case-sensitive match"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			presentationID, err := requireMCPString(req, "presentation_id")
			if err != nil {
				return nil, err
			}
			find, err := requireMCPText(req, "find")
			if err != nil {
				return nil, err
			}
			replacement, err := requireMCPText(req, "replacement")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "replace-text").
				flag("match_case", "--match-case").done(presentationID, find, replacement)
		},
	}
}

func mcpSlidesDuplicateSlideTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_duplicate_slide",
		Service:     "slides",
		Risk:        mcpRiskWrite,
		Description: "Duplicate a slide by object ID. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("presentation_id", mcp.Description("Google Slides presentation ID"), mcp.Required()),
			mcp.WithString("slide_id", mcp.Description("Slide object ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			presentationID, err := requireMCPString(req, "presentation_id")
			if err != nil {
				return nil, err
			}
			slideID, err := requireMCPString(req, "slide_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "duplicate-slide").done(presentationID, slideID)
		},
	}
}

func mcpSlidesDeleteSlideTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "slides_delete_slide",
		Service:     "slides",
		Risk:        mcpRiskWrite,
		Description: "Delete a slide by object ID. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("presentation_id", mcp.Description("Google Slides presentation ID"), mcp.Required()),
			mcp.WithString("slide_id", mcp.Description("Slide object ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			presentationID, err := requireMCPString(req, "presentation_id")
			if err != nil {
				return nil, err
			}
			slideID, err := requireMCPString(req, "slide_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "slides", "delete-slide").done(presentationID, slideID)
		},
	}
}
