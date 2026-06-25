package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpDriveTools returns the Google Drive MCP tool surface (service "drive").
func mcpDriveTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpDriveSearchTool(),
		mcpDriveGetTool(),
		mcpDriveListTool(),
		mcpDriveListDrivesTool(),
		mcpDrivePermissionsTool(),
		mcpDriveListCommentsTool(),
		mcpDriveListRevisionsTool(),
		mcpDriveTreeTool(),
		mcpDriveCreateFolderTool(),
		mcpDriveCopyTool(),
		mcpDriveMoveTool(),
		mcpDriveRenameTool(),
		mcpDriveTrashTool(),
		mcpDriveShareTool(),
		mcpDriveCreateCommentTool(),
	}
}

func mcpDriveListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_list",
		Service:     "drive",
		Risk:        mcpRiskRead,
		Description: "List files in a Drive folder (default: root).",
		Options: []mcp.ToolOption{
			mcp.WithString("parent", mcp.Description("Parent folder or shared drive ID")),
			mcp.WithString("query", mcp.Description("Optional Drive query filter")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "drive", "ls").
				str("parent", "--parent").
				str("query", "--query").
				num("max", "--max", 50, 200).done()
		},
	}
}

func mcpDriveListDrivesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_list_drives",
		Service:     "drive",
		Risk:        mcpRiskRead,
		Description: "List shared drives (Team Drives).",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Optional name filter")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "drive", "drives").
				str("query", "--query").
				num("max", "--max", 50, 100).done()
		},
	}
}

func mcpDrivePermissionsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_permissions",
		Service:     "drive",
		Risk:        mcpRiskRead,
		Description: "List permissions on a Drive file or folder.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("Drive file ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "permissions").done(fileID)
		},
	}
}

func mcpDriveListCommentsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_list_comments",
		Service:     "drive",
		Risk:        mcpRiskRead,
		Description: "List comments on a Drive file.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("Drive file ID"), mcp.Required()),
			mcp.WithBoolean("include_quoted", mcp.Description("Include quoted file content"), mcp.DefaultBool(false)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "comments", "list").
				flag("include_quoted", "--include-quoted").
				num("max", "--max", 50, 200).done(fileID)
		},
	}
}

func mcpDriveListRevisionsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_list_revisions",
		Service:     "drive",
		Risk:        mcpRiskRead,
		Description: "List revisions for a Drive file.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("Drive file ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "revisions", "list").
				num("max", "--max", 50, 200).done(fileID)
		},
	}
}

func mcpDriveTreeTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_tree",
		Service:     "drive",
		Risk:        mcpRiskRead,
		Description: "Print a read-only Drive folder tree.",
		Options: []mcp.ToolOption{
			mcp.WithString("parent", mcp.Description("Parent folder ID (default: root)")),
			mcp.WithInteger("depth", mcp.Description("Maximum depth"), mcp.DefaultNumber(3), mcp.Min(1), mcp.Max(20)),
			mcp.WithInteger("max", mcp.Description("Maximum entries"), mcp.DefaultNumber(200), mcp.Min(1), mcp.Max(2000)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "drive", "tree").
				str("parent", "--parent").
				num("depth", "--depth", 3, 20).
				num("max", "--max", 200, 2000).done()
		},
	}
}

func mcpDriveCreateFolderTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_create_folder",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Create a Drive folder. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("name", mcp.Description("Folder name"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Parent folder ID")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			name, err := requireMCPString(req, "name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "mkdir").
				str("parent", "--parent").done(name)
		},
	}
}

func mcpDriveCopyTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_copy",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Copy a Drive file. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("Source file ID"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Name for the copy"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Destination folder ID")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			name, err := requireMCPString(req, "name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "copy").
				str("parent", "--parent").done(fileID, name)
		},
	}
}

func mcpDriveMoveTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_move",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Move a Drive file to another folder. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("File ID to move"), mcp.Required()),
			mcp.WithString("parent", mcp.Description("Destination folder ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			parent, err := requireMCPString(req, "parent")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "move", "--parent", parent).done(fileID)
		},
	}
}

func mcpDriveRenameTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_rename",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Rename a Drive file or folder. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("File ID"), mcp.Required()),
			mcp.WithString("new_name", mcp.Description("New name"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			newName, err := requireMCPString(req, "new_name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "rename").done(fileID, newName)
		},
	}
}

func mcpDriveTrashTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_trash",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Move a Drive file to Trash (reversible; never permanently deletes). Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("File ID to trash"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "delete").done(fileID)
		},
	}
}

func mcpDriveShareTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_share",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Share a Drive file or folder with a user or domain. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("File ID to share"), mcp.Required()),
			mcp.WithString("email", mcp.Description("User or group email to share with")),
			mcp.WithString("domain", mcp.Description("Domain to share with")),
			mcp.WithString("role", mcp.Description("Permission role"), mcp.Enum("reader", "commenter", "writer", "fileOrganizer", "organizer", "owner")),
			mcp.WithBoolean("notify", mcp.Description("Send notification email"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "share").
				str("email", "--email").
				str("domain", "--domain").
				str("role", "--role").
				flag("notify", "--notify").done(fileID)
		},
	}
}

func mcpDriveCreateCommentTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "drive_create_comment",
		Service:     "drive",
		Risk:        mcpRiskWrite,
		Description: "Create a comment on a Drive file. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("file_id", mcp.Description("Drive file ID"), mcp.Required()),
			mcp.WithString("content", mcp.Description("Comment content"), mcp.Required()),
			mcp.WithString("quoted", mcp.Description("Optional quoted text the comment anchors to")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			fileID, err := requireMCPString(req, "file_id")
			if err != nil {
				return nil, err
			}
			content, err := requireMCPText(req, "content")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "drive", "comments", "create").
				str("quoted", "--quoted").done(fileID, content)
		},
	}
}
