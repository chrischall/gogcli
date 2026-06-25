package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpGmailTools returns the Gmail MCP tool surface (service "gmail").
func mcpGmailTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpGmailSearchTool(),
		mcpGmailSearchThreadsTool(),
		mcpGmailGetMessageTool(),
		mcpGmailGetThreadTool(),
		mcpGmailListLabelsTool(),
		mcpGmailGetLabelTool(),
		mcpGmailListDraftsTool(),
		mcpGmailGetDraftTool(),
		mcpGmailHistoryTool(),
		mcpGmailSendTool(),
		mcpGmailReplyTool(),
		mcpGmailForwardTool(),
		mcpGmailCreateDraftTool(),
		mcpGmailSendDraftTool(),
		mcpGmailModifyMessageTool(),
		mcpGmailModifyThreadTool(),
		mcpGmailTrashTool(),
		mcpGmailCreateLabelTool(),
	}
}

func mcpGmailSearchThreadsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_search_threads",
		Service:     "gmail",
		Risk:        mcpRiskRead,
		Description: "Search Gmail threads with Gmail query syntax. Returns thread summaries.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Gmail search query, e.g. newer_than:7d from:person@example.com"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(10), mcp.Min(1), mcp.Max(100)),
			mcp.WithBoolean("oldest", mcp.Description("Sort oldest first"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "search").
				num("max", "--max", 10, 1, 100).
				flag("oldest", "--oldest").
				done(query)
		},
	}
}

func mcpGmailListLabelsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_list_labels",
		Service:     "gmail",
		Risk:        mcpRiskRead,
		Description: "List Gmail labels with their IDs.",
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "gmail", "labels", "list").done()
		},
	}
}

func mcpGmailGetLabelTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_get_label",
		Service:     "gmail",
		Risk:        mcpRiskRead,
		Description: "Get one Gmail label (including message counts) by ID or name.",
		Options: []mcp.ToolOption{
			mcp.WithString("label", mcp.Description("Label ID or name"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			label, err := requireMCPString(req, "label")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "labels", "get").done(label)
		},
	}
}

func mcpGmailListDraftsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_list_drafts",
		Service:     "gmail",
		Risk:        mcpRiskRead,
		Description: "List Gmail drafts.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(20), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "gmail", "drafts", "list").
				num("max", "--max", 20, 1, 100).done()
		},
	}
}

func mcpGmailGetDraftTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_get_draft",
		Service:     "gmail",
		Risk:        mcpRiskRead,
		Description: "Get one Gmail draft by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("draft_id", mcp.Description("Gmail draft ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			draftID, err := requireMCPString(req, "draft_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "drafts", "get").done(draftID)
		},
	}
}

func mcpGmailHistoryTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_history",
		Service:     "gmail",
		Risk:        mcpRiskRead,
		Description: "List Gmail history (mailbox changes) since a history ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("since", mcp.Description("Start history ID")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(500)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "gmail", "history").
				str("since", "--since").
				num("max", "--max", 50, 1, 500).done()
		},
	}
}

func mcpGmailSendTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_send",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Send an email. Requires --allow-write; also blocked by --gmail-no-send.",
		Options: []mcp.ToolOption{
			mcp.WithString("to", mcp.Description("Recipient address(es), comma-separated"), mcp.Required()),
			mcp.WithString("subject", mcp.Description("Subject line"), mcp.Required()),
			mcp.WithString("body", mcp.Description("Plain-text body"), mcp.Required()),
			mcp.WithString("cc", mcp.Description("Cc address(es)")),
			mcp.WithString("bcc", mcp.Description("Bcc address(es)")),
			mcp.WithString("from", mcp.Description("Send-as address")),
			mcp.WithString("thread_id", mcp.Description("Thread ID to reply within")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			to, err := requireMCPString(req, "to")
			if err != nil {
				return nil, err
			}
			subject, err := requireMCPText(req, "subject")
			if err != nil {
				return nil, err
			}
			body, err := requireMCPText(req, "body")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "send", "--to", to, "--subject", subject, "--body", body).
				str("cc", "--cc").
				str("bcc", "--bcc").
				str("from", "--from").
				str("thread_id", "--thread-id").done()
		},
	}
}

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

func mcpGmailForwardTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_forward",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Forward a Gmail message. Requires --allow-write; also blocked by --gmail-no-send.",
		Options: []mcp.ToolOption{
			mcp.WithString("message_id", mcp.Description("Message ID to forward"), mcp.Required()),
			mcp.WithString("to", mcp.Description("Recipient address(es)"), mcp.Required()),
			mcp.WithString("note", mcp.Description("Optional note prepended to the forward")),
			mcp.WithString("cc", mcp.Description("Cc address(es)")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			messageID, err := requireMCPString(req, "message_id")
			if err != nil {
				return nil, err
			}
			to, err := requireMCPString(req, "to")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "forward", "--to", to).
				str("note", "--note").
				str("cc", "--cc").done(messageID)
		},
	}
}

func mcpGmailCreateDraftTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_create_draft",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Create a Gmail draft. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("to", mcp.Description("Recipient address(es)")),
			mcp.WithString("subject", mcp.Description("Subject line")),
			mcp.WithString("body", mcp.Description("Plain-text body")),
			mcp.WithString("cc", mcp.Description("Cc address(es)")),
			mcp.WithString("bcc", mcp.Description("Bcc address(es)")),
			mcp.WithString("thread_id", mcp.Description("Thread ID to draft within")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "gmail", "drafts", "create").
				str("to", "--to").
				str("subject", "--subject").
				str("body", "--body").
				str("cc", "--cc").
				str("bcc", "--bcc").
				str("thread_id", "--thread-id").done()
		},
	}
}

func mcpGmailSendDraftTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_send_draft",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Send an existing Gmail draft. Requires --allow-write; also blocked by --gmail-no-send.",
		Options: []mcp.ToolOption{
			mcp.WithString("draft_id", mcp.Description("Draft ID to send"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			draftID, err := requireMCPString(req, "draft_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "drafts", "send").done(draftID)
		},
	}
}

func mcpGmailModifyMessageTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_modify_message",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Add and/or remove labels on a single Gmail message. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("message_id", mcp.Description("Message ID"), mcp.Required()),
			mcp.WithString("add", mcp.Description("Comma-separated labels to add")),
			mcp.WithString("remove", mcp.Description("Comma-separated labels to remove")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			messageID, err := requireMCPString(req, "message_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "messages", "modify").
				str("add", "--add").
				str("remove", "--remove").done(messageID)
		},
	}
}

func mcpGmailModifyThreadTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_modify_thread",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Add and/or remove labels on a Gmail thread. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("thread_id", mcp.Description("Thread ID"), mcp.Required()),
			mcp.WithString("add", mcp.Description("Comma-separated labels to add")),
			mcp.WithString("remove", mcp.Description("Comma-separated labels to remove")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			threadID, err := requireMCPString(req, "thread_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "labels", "modify").
				str("add", "--add").
				str("remove", "--remove").done(threadID)
		},
	}
}

func mcpGmailTrashTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_trash",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Move a Gmail message to Trash (reversible). Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("message_id", mcp.Description("Message ID to trash"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			messageID, err := requireMCPString(req, "message_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "trash").done(messageID)
		},
	}
}

func mcpGmailCreateLabelTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "gmail_create_label",
		Service:     "gmail",
		Risk:        mcpRiskWrite,
		Description: "Create a new Gmail label. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("name", mcp.Description("Label name"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			name, err := requireMCPString(req, "name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "gmail", "labels", "create").done(name)
		},
	}
}
