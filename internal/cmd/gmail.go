package cmd

type GmailCmd struct {
	Search     GmailSearchCmd     `cmd:"" name:"search" aliases:"find,query,ls,list" group:"Read" help:"Search threads using Gmail query syntax" mcp:"gmail_search_threads,read" mcpdesc:"Search Gmail threads with Gmail query syntax. Returns thread summaries."`
	Messages   GmailMessagesCmd   `cmd:"" name:"messages" aliases:"message,msg,msgs" group:"Read" help:"Message operations"`
	Thread     GmailThreadCmd     `cmd:"" name:"thread" aliases:"threads,read" group:"Organize" help:"Thread operations (get, modify)"`
	Get        GmailGetCmd        `cmd:"" name:"get" aliases:"info,show" group:"Read" help:"Get a message (full|metadata|raw)" mcp:"gmail_get_message,read" mcpdesc:"Get one Gmail message by ID. Sanitized content is enabled by default."`
	Raw        GmailRawCmd        `cmd:"" name:"raw" group:"Read" help:"Dump raw Gmail API response as JSON (Users.Messages.Get; lossless; for scripting and LLM consumption)"`
	Attachment GmailAttachmentCmd `cmd:"" name:"attachment" group:"Read" help:"Download a single attachment"`
	URL        GmailURLCmd        `cmd:"" name:"url" group:"Read" help:"Print Gmail web URLs for threads"`
	History    GmailHistoryCmd    `cmd:"" name:"history" group:"Read" help:"Gmail history" mcp:"gmail_history,read" mcpdesc:"List Gmail history (mailbox changes) since a history ID."`

	Labels  GmailLabelsCmd   `cmd:"" name:"labels" aliases:"label" group:"Organize" help:"Label operations"`
	Batch   GmailBatchCmd    `cmd:"" name:"batch" group:"Organize" help:"Batch operations (permanent delete requires broader Gmail scope; use gmail trash for normal trashing)"`
	Archive GmailArchiveCmd  `cmd:"" name:"archive" group:"Organize" help:"Archive messages or explicit threads (remove from inbox)"`
	Read    GmailReadCmd     `cmd:"" name:"mark-read" aliases:"read-messages" group:"Organize" help:"Mark messages as read"`
	Unread  GmailUnreadCmd   `cmd:"" name:"unread" aliases:"mark-unread" group:"Organize" help:"Mark messages as unread"`
	Trash   GmailTrashMsgCmd `cmd:"" name:"trash" group:"Organize" help:"Move messages to trash" mcp:"gmail_trash,write" mcpdesc:"Move a Gmail message to Trash (reversible). Requires --allow-write."`

	Send      GmailSendCmd      `cmd:"" name:"send" group:"Write" help:"Send an email" mcp:"gmail_send,write" mcpdesc:"Send an email. Requires --allow-write; also blocked by --gmail-no-send."`
	Reply     GmailReplyCmd     `cmd:"" name:"reply" group:"Write" help:"Reply to a message"`
	ReplyAll  GmailReplyAllCmd  `cmd:"" name:"reply-all" aliases:"replyall" group:"Write" help:"Reply to all message participants"`
	Forward   GmailForwardCmd   `cmd:"" name:"forward" aliases:"fwd" group:"Write" help:"Forward a message to new recipients" mcp:"gmail_forward,write" mcpdesc:"Forward a Gmail message. Requires --allow-write; also blocked by --gmail-no-send."`
	AutoReply GmailAutoReplyCmd `cmd:"" name:"autoreply" group:"Write" help:"Reply once to matching messages"`
	Track     GmailTrackCmd     `cmd:"" name:"track" group:"Write" help:"Email open tracking"`
	Drafts    GmailDraftsCmd    `cmd:"" name:"drafts" aliases:"draft" group:"Write" help:"Draft operations"`

	Settings GmailSettingsCmd `cmd:"" name:"settings" group:"Admin" help:"Settings and admin"`

	Watch       GmailWatchCmd       `cmd:"" name:"watch" hidden:"" help:"Manage Gmail watch"`
	AutoForward GmailAutoForwardCmd `cmd:"" name:"autoforward" hidden:"" help:"Auto-forwarding settings"`
	Delegates   GmailDelegatesCmd   `cmd:"" name:"delegates" hidden:"" help:"Delegate operations"`
	Filters     GmailFiltersCmd     `cmd:"" name:"filters" hidden:"" help:"Filter operations"`
	Forwarding  GmailForwardingCmd  `cmd:"" name:"forwarding" hidden:"" help:"Forwarding addresses"`
	SendAs      GmailSendAsCmd      `cmd:"" name:"sendas" hidden:"" help:"Send-as settings"`
	Vacation    GmailVacationCmd    `cmd:"" name:"vacation" hidden:"" help:"Vacation responder"`
}

type GmailSettingsCmd struct {
	Filters     GmailFiltersCmd     `cmd:"" name:"filters" group:"Organize" help:"Filter operations"`
	Delegates   GmailDelegatesCmd   `cmd:"" name:"delegates" group:"Admin" help:"Delegate operations"`
	Forwarding  GmailForwardingCmd  `cmd:"" name:"forwarding" group:"Admin" help:"Forwarding addresses"`
	AutoForward GmailAutoForwardCmd `cmd:"" name:"autoforward" group:"Admin" help:"Auto-forwarding settings"`
	SendAs      GmailSendAsCmd      `cmd:"" name:"sendas" group:"Admin" help:"Send-as settings"`
	Vacation    GmailVacationCmd    `cmd:"" name:"vacation" group:"Admin" help:"Vacation responder"`
	Watch       GmailWatchCmd       `cmd:"" name:"watch" group:"Admin" help:"Manage Gmail watch"`
}
