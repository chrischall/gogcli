package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpCustomTools returns the hand-written tool specs whose request-to-argv
// mapping cannot be expressed with mcp/mcpdesc annotations (conditional
// subcommand selection, cross-field validation). Every other tool is generated
// from the annotated command grammar; see mcp_gen.go.
func mcpCustomTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpGmailReplyTool(),
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
