package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpCustomTools returns the hand-written tool specs whose request-to-argv
// mapping cannot be expressed with mcp/mcpdesc annotations (conditional
// subcommand selection, cross-field validation). Every other tool is generated
// from the annotated command grammar; see mcp_gen.go.
func mcpCustomTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpGmailReplyTool(),
		mcpDocsGetTool(),
		mcpDocsWriteTool(),
		mcpContactsCreateTool(),
	}
}

// mcpGmailReplyTool stays hand-written: it selects the child subcommand
// (`reply` vs `reply-all`) from the request's reply_all boolean, which the
// annotation grammar cannot express.
func mcpGmailReplyTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_reply",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Reply to a Gmail message. Requires --allow-write; also blocked by --gmail-no-send.",
		Options: []mcp.ToolOption{
			mcp.WithString("message_id", mcp.Description("Message ID to reply to"), mcp.Required()),
			mcp.WithString("body", mcp.Description("Reply body"), mcp.Required()),
			mcp.WithBoolean("reply_all", mcp.Description("Reply to all participants"), mcp.DefaultBool(false)),
			mcp.WithString("from", mcp.Description("Send-as address")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			messageID, err := requireMCPString(req, "message_id")
			if err != nil {
				return nil, err
			}
			body, err := requireMCPText(req, "body")
			if err != nil {
				return nil, err
			}
			sub := "reply"
			if req.GetBool("reply_all", false) {
				sub = "reply-all"
			}
			return mcpCommand(req, "gmail", sub, "--body", body).
				str("from", "--from").done(messageID)
		},
	}
}

// mcpDocsGetTool stays hand-written: tab/all_tabs are mutually exclusive and a
// provided-but-empty tab must fail, cross-field rules the annotation grammar
// cannot express.
func mcpDocsGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_get",
		Service:     "docs",
		Risk:        mcpRiskRead,
		Description: "Read a Google Doc as wrapped text, all tabs, or one tab.",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
			mcp.WithString("tab", mcp.Description("Optional tab title or ID")),
			mcp.WithBoolean("all_tabs", mcp.Description("Read all tabs"), mcp.DefaultBool(false)),
			mcp.WithInteger("max_bytes", mcp.Description("Maximum text bytes, 0 for unlimited"), mcp.DefaultNumber(2000000), mcp.Min(0)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			args := []string{"docs", "cat", "--max-bytes", strconv.Itoa(clampMCPInt(req.GetInt("max_bytes", 2000000), 0, 20_000_000))}
			tab := strings.TrimSpace(req.GetString("tab", ""))
			_, tabProvided := req.GetArguments()["tab"]
			if tabProvided && tab == "" {
				return nil, fmt.Errorf("tab cannot be empty")
			}
			allTabs := req.GetBool("all_tabs", false)
			if tab != "" && allTabs {
				return nil, fmt.Errorf("tab and all_tabs are mutually exclusive")
			}
			if tab != "" {
				args = append(args, "--tab", tab)
			}
			if allTabs {
				args = append(args, "--all-tabs")
			}
			return append(args, "--", docID), nil
		},
	}
}

// mcpContactsCreateTool stays hand-written: the "at least one of given,
// family, or email" cross-field rule cannot be expressed with the annotation
// grammar.
func mcpContactsCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_create",
		Service:     "contacts",
		Risk:        mcpRiskWrite,
		Description: "Create a contact. Provide at least a name or email. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("given", mcp.Description("Given (first) name")),
			mcp.WithString("family", mcp.Description("Family (last) name")),
			mcp.WithString("email", mcp.Description("Email address")),
			mcp.WithString("phone", mcp.Description("Phone number")),
			mcp.WithString("org", mcp.Description("Organization")),
			mcp.WithString("title", mcp.Description("Job title")),
			mcp.WithString("note", mcp.Description("Note")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			given := strings.TrimSpace(req.GetString("given", ""))
			family := strings.TrimSpace(req.GetString("family", ""))
			email := strings.TrimSpace(req.GetString("email", ""))
			if given == "" && family == "" && email == "" {
				return nil, fmt.Errorf("provide at least one of given, family, or email")
			}
			return mcpCommand(req, "contacts", "create").
				str("given", "--given").
				str("family", "--family").
				str("email", "--email").
				str("phone", "--phone").
				str("org", "--org").
				str("title", "--title").
				str("note", "--note").done()
		},
	}
}

// mcpDocsWriteTool stays hand-written: append/replace carry cross-field logic
// with an explicit failure mode the annotation grammar cannot express.
func mcpDocsWriteTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "docs_write",
		Service:     "docs",
		Risk:        mcpRiskWrite,
		Description: "Write text to a Google Doc. Requires --allow-write on the MCP server.",
		Options: []mcp.ToolOption{
			mcp.WithString("document_id", mcp.Description("Google Docs document ID"), mcp.Required()),
			mcp.WithString("text", mcp.Description("Text or markdown to write"), mcp.Required()),
			mcp.WithString("tab", mcp.Description("Optional tab title or ID")),
			mcp.WithBoolean("append", mcp.Description("Append instead of replacing"), mcp.DefaultBool(true)),
			mcp.WithBoolean("replace", mcp.Description("Replace all existing content"), mcp.DefaultBool(false)),
			mcp.WithBoolean("markdown", mcp.Description("Convert markdown to Docs formatting"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			docID, err := requireMCPString(req, "document_id")
			if err != nil {
				return nil, err
			}
			text, err := requireMCPText(req, "text")
			if err != nil {
				return nil, err
			}
			args := []string{"docs", "write", "--text", text}
			reqArgs := req.GetArguments()
			replace := req.GetBool("replace", false)
			appendProvided := false
			if reqArgs != nil {
				_, appendProvided = reqArgs["append"]
			}
			appendMode := req.GetBool("append", true)
			if replace && appendProvided && appendMode {
				return nil, fmt.Errorf("append and replace are mutually exclusive")
			}
			switch {
			case replace:
				args = append(args, "--replace")
			case appendMode:
				args = append(args, "--append")
			default:
				return nil, fmt.Errorf("append=false requires replace=true to avoid implicit document replacement")
			}
			if req.GetBool("markdown", false) {
				args = append(args, "--markdown")
			}
			if tab := strings.TrimSpace(req.GetString("tab", "")); tab != "" {
				args = append(args, "--tab", tab)
			}
			return append(args, "--", docID), nil
		},
	}
}
