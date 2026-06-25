package cmd

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpContactsTools returns the Google Contacts MCP tool surface (service "contacts").
func mcpContactsTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpContactsListTool(),
		mcpContactsGetTool(),
		mcpContactsSearchTool(),
		mcpContactsDirectorySearchTool(),
		mcpContactsCreateTool(),
		mcpContactsUpdateTool(),
		mcpContactsDeleteTool(),
	}
}

func mcpContactsListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_list",
		Service:     "contacts",
		Risk:        mcpRiskRead,
		Description: "List the account's contacts.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "contacts", "list").
				num("max", "--max", 50, 1, 200).done()
		},
	}
}

func mcpContactsGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_get",
		Service:     "contacts",
		Risk:        mcpRiskRead,
		Description: "Get a contact by resource name (e.g. people/c123).",
		Options: []mcp.ToolOption{
			mcp.WithString("resource_name", mcp.Description("Contact resource name"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			resourceName, err := requireMCPString(req, "resource_name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "contacts", "get").done(resourceName)
		},
	}
}

func mcpContactsSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_search",
		Service:     "contacts",
		Risk:        mcpRiskRead,
		Description: "Search contacts by name, email, or phone.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "contacts", "search").
				num("max", "--max", 25, 1, 100).done(query)
		},
	}
}

func mcpContactsDirectorySearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_directory_search",
		Service:     "contacts",
		Risk:        mcpRiskRead,
		Description: "Search people in the Workspace directory.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "contacts", "directory", "search").
				num("max", "--max", 25, 1, 100).done(query)
		},
	}
}

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

func mcpContactsUpdateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_update",
		Service:     "contacts",
		Risk:        mcpRiskWrite,
		Description: "Update a contact by resource name. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("resource_name", mcp.Description("Contact resource name"), mcp.Required()),
			mcp.WithString("given", mcp.Description("Given (first) name")),
			mcp.WithString("family", mcp.Description("Family (last) name")),
			mcp.WithString("email", mcp.Description("Email address")),
			mcp.WithString("phone", mcp.Description("Phone number")),
			mcp.WithString("org", mcp.Description("Organization")),
			mcp.WithString("title", mcp.Description("Job title")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			resourceName, err := requireMCPString(req, "resource_name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "contacts", "update").
				str("given", "--given").
				str("family", "--family").
				str("email", "--email").
				str("phone", "--phone").
				str("org", "--org").
				str("title", "--title").done(resourceName)
		},
	}
}

func mcpContactsDeleteTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "contacts_delete",
		Service:     "contacts",
		Risk:        mcpRiskWrite,
		Description: "Delete a contact by resource name. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("resource_name", mcp.Description("Contact resource name"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			resourceName, err := requireMCPString(req, "resource_name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "contacts", "delete").done(resourceName)
		},
	}
}

// mcpPeopleTools returns the Google People MCP tool surface (service "people").
func mcpPeopleTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpPeopleMeTool(),
		mcpPeopleGetTool(),
		mcpPeopleSearchTool(),
	}
}

func mcpPeopleMeTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "people_me",
		Service:     "people",
		Risk:        mcpRiskRead,
		Description: "Show the authenticated user's profile.",
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "people", "me").done()
		},
	}
}

func mcpPeopleGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "people_get",
		Service:     "people",
		Risk:        mcpRiskRead,
		Description: "Get a user profile by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("user_id", mcp.Description("User ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			userID, err := requireMCPString(req, "user_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "people", "get").done(userID)
		},
	}
}

func mcpPeopleSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "people_search",
		Service:     "people",
		Risk:        mcpRiskRead,
		Description: "Search the Workspace directory.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "people", "search").
				num("max", "--max", 25, 1, 100).done(query)
		},
	}
}

// mcpTasksTools returns the Google Tasks MCP tool surface (service "tasks").
func mcpTasksTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpTasksListsTool(),
		mcpTasksListTool(),
		mcpTasksGetTool(),
		mcpTasksAddTool(),
		mcpTasksCompleteTool(),
		mcpTasksUpdateTool(),
		mcpTasksDeleteTool(),
		mcpTasksCreateListTool(),
	}
}

func mcpTasksListsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_lists",
		Service:     "tasks",
		Risk:        mcpRiskRead,
		Description: "List the account's task lists.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "tasks", "lists", "list").
				num("max", "--max", 50, 1, 100).done()
		},
	}
}

func mcpTasksListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_list",
		Service:     "tasks",
		Risk:        mcpRiskRead,
		Description: "List tasks in a task list.",
		Options: []mcp.ToolOption{
			mcp.WithString("tasklist_id", mcp.Description("Task list ID"), mcp.Required()),
			mcp.WithBoolean("show_completed", mcp.Description("Include completed tasks"), mcp.DefaultBool(false)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			tasklistID, err := requireMCPString(req, "tasklist_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "list").
				flag("show_completed", "--show-completed").
				num("max", "--max", 50, 1, 100).done(tasklistID)
		},
	}
}

func mcpTasksGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_get",
		Service:     "tasks",
		Risk:        mcpRiskRead,
		Description: "Get a task by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("tasklist_id", mcp.Description("Task list ID"), mcp.Required()),
			mcp.WithString("task_id", mcp.Description("Task ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			tasklistID, err := requireMCPString(req, "tasklist_id")
			if err != nil {
				return nil, err
			}
			taskID, err := requireMCPString(req, "task_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "get").done(tasklistID, taskID)
		},
	}
}

func mcpTasksAddTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_add",
		Service:     "tasks",
		Risk:        mcpRiskWrite,
		Description: "Add a task to a task list. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("tasklist_id", mcp.Description("Task list ID"), mcp.Required()),
			mcp.WithString("title", mcp.Description("Task title"), mcp.Required()),
			mcp.WithString("notes", mcp.Description("Task notes")),
			mcp.WithString("due", mcp.Description("Due date (RFC3339 or date)")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			tasklistID, err := requireMCPString(req, "tasklist_id")
			if err != nil {
				return nil, err
			}
			title, err := requireMCPText(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "add", "--title", title).
				str("notes", "--notes").
				str("due", "--due").done(tasklistID)
		},
	}
}

func mcpTasksCompleteTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_complete",
		Service:     "tasks",
		Risk:        mcpRiskWrite,
		Description: "Mark a task completed. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("tasklist_id", mcp.Description("Task list ID"), mcp.Required()),
			mcp.WithString("task_id", mcp.Description("Task ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			tasklistID, err := requireMCPString(req, "tasklist_id")
			if err != nil {
				return nil, err
			}
			taskID, err := requireMCPString(req, "task_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "done").done(tasklistID, taskID)
		},
	}
}

func mcpTasksUpdateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_update",
		Service:     "tasks",
		Risk:        mcpRiskWrite,
		Description: "Update a task. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("tasklist_id", mcp.Description("Task list ID"), mcp.Required()),
			mcp.WithString("task_id", mcp.Description("Task ID"), mcp.Required()),
			mcp.WithString("title", mcp.Description("New title")),
			mcp.WithString("notes", mcp.Description("New notes")),
			mcp.WithString("due", mcp.Description("New due date")),
			mcp.WithString("status", mcp.Description("Status"), mcp.Enum("needsAction", "completed")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			tasklistID, err := requireMCPString(req, "tasklist_id")
			if err != nil {
				return nil, err
			}
			taskID, err := requireMCPString(req, "task_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "update").
				str("title", "--title").
				str("notes", "--notes").
				str("due", "--due").
				str("status", "--status").done(tasklistID, taskID)
		},
	}
}

func mcpTasksDeleteTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_delete",
		Service:     "tasks",
		Risk:        mcpRiskWrite,
		Description: "Delete a task. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("tasklist_id", mcp.Description("Task list ID"), mcp.Required()),
			mcp.WithString("task_id", mcp.Description("Task ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			tasklistID, err := requireMCPString(req, "tasklist_id")
			if err != nil {
				return nil, err
			}
			taskID, err := requireMCPString(req, "task_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "delete").done(tasklistID, taskID)
		},
	}
}

func mcpTasksCreateListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "tasks_create_list",
		Service:     "tasks",
		Risk:        mcpRiskWrite,
		Description: "Create a new task list. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Task list title"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			title, err := requireMCPString(req, "title")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "tasks", "lists", "create").done(title)
		},
	}
}

// mcpChatTools returns the Google Chat MCP tool surface (service "chat").
func mcpChatTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpChatListSpacesTool(),
		mcpChatFindSpacesTool(),
		mcpChatListMessagesTool(),
		mcpChatListThreadsTool(),
		mcpChatSendMessageTool(),
		mcpChatSendDMTool(),
	}
}

func mcpChatListSpacesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "chat_list_spaces",
		Service:     "chat",
		Risk:        mcpRiskRead,
		Description: "List Google Chat spaces.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "chat", "spaces", "list").
				num("max", "--max", 50, 1, 100).done()
		},
	}
}

func mcpChatFindSpacesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "chat_find_spaces",
		Service:     "chat",
		Risk:        mcpRiskRead,
		Description: "Find Google Chat spaces by display name.",
		Options: []mcp.ToolOption{
			mcp.WithString("display_name", mcp.Description("Space display name"), mcp.Required()),
			mcp.WithBoolean("exact", mcp.Description("Exact match only"), mcp.DefaultBool(false)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			displayName, err := requireMCPString(req, "display_name")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "chat", "spaces", "find").
				flag("exact", "--exact").
				num("max", "--max", 25, 1, 100).done(displayName)
		},
	}
}

func mcpChatListMessagesTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "chat_list_messages",
		Service:     "chat",
		Risk:        mcpRiskRead,
		Description: "List messages in a Google Chat space.",
		Options: []mcp.ToolOption{
			mcp.WithString("space", mcp.Description("Space name or ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			space, err := requireMCPString(req, "space")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "chat", "messages", "list").
				num("max", "--max", 50, 1, 100).done(space)
		},
	}
}

func mcpChatListThreadsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "chat_list_threads",
		Service:     "chat",
		Risk:        mcpRiskRead,
		Description: "List threads in a Google Chat space.",
		Options: []mcp.ToolOption{
			mcp.WithString("space", mcp.Description("Space name or ID"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			space, err := requireMCPString(req, "space")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "chat", "threads", "list").
				num("max", "--max", 50, 1, 100).done(space)
		},
	}
}

func mcpChatSendMessageTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "chat_send_message",
		Service:     "chat",
		Risk:        mcpRiskWrite,
		Description: "Send a message to a Google Chat space. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("space", mcp.Description("Space name or ID"), mcp.Required()),
			mcp.WithString("text", mcp.Description("Message text"), mcp.Required()),
			mcp.WithString("thread", mcp.Description("Thread key to reply within")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			space, err := requireMCPString(req, "space")
			if err != nil {
				return nil, err
			}
			text, err := requireMCPText(req, "text")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "chat", "messages", "send", "--text", text).
				str("thread", "--thread").done(space)
		},
	}
}

func mcpChatSendDMTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "chat_send_dm",
		Service:     "chat",
		Risk:        mcpRiskWrite,
		Description: "Send a Google Chat direct message to a user by email. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("email", mcp.Description("Recipient email"), mcp.Required()),
			mcp.WithString("text", mcp.Description("Message text"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			email, err := requireMCPString(req, "email")
			if err != nil {
				return nil, err
			}
			text, err := requireMCPText(req, "text")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "chat", "dm", "send", "--text", text).done(email)
		},
	}
}

// mcpKeepTools returns the Google Keep MCP tool surface (service "keep").
func mcpKeepTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpKeepListTool(),
		mcpKeepGetTool(),
		mcpKeepSearchTool(),
		mcpKeepCreateTool(),
	}
}

func mcpKeepListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "keep_list",
		Service:     "keep",
		Risk:        mcpRiskRead,
		Description: "List Google Keep notes (Workspace only).",
		Options: []mcp.ToolOption{
			mcp.WithString("filter", mcp.Description("Optional Keep filter expression")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "keep", "list").
				str("filter", "--filter").
				num("max", "--max", 50, 1, 100).done()
		},
	}
}

func mcpKeepGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "keep_get",
		Service:     "keep",
		Risk:        mcpRiskRead,
		Description: "Get a Google Keep note by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("note_id", mcp.Description("Note ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			noteID, err := requireMCPString(req, "note_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "keep", "get").done(noteID)
		},
	}
}

func mcpKeepSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "keep_search",
		Service:     "keep",
		Risk:        mcpRiskRead,
		Description: "Search Google Keep notes by text.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "keep", "search").
				num("max", "--max", 25, 1, 100).done(query)
		},
	}
}

func mcpKeepCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "keep_create",
		Service:     "keep",
		Risk:        mcpRiskWrite,
		Description: "Create a Google Keep note. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("title", mcp.Description("Note title")),
			mcp.WithString("text", mcp.Description("Note body text"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			text, err := requireMCPText(req, "text")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "keep", "create", "--text", text).
				str("title", "--title").done()
		},
	}
}

// mcpMeetTools returns the Google Meet MCP tool surface (service "meet").
func mcpMeetTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpMeetGetTool(),
		mcpMeetHistoryTool(),
		mcpMeetParticipantsTool(),
		mcpMeetCreateTool(),
	}
}

func mcpMeetGetTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "meet_get",
		Service:     "meet",
		Risk:        mcpRiskRead,
		Description: "Get a Google Meet space by meeting code.",
		Options: []mcp.ToolOption{
			mcp.WithString("meeting_code", mcp.Description("Meeting code"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			code, err := requireMCPString(req, "meeting_code")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "meet", "get").done(code)
		},
	}
}

func mcpMeetHistoryTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "meet_history",
		Service:     "meet",
		Risk:        mcpRiskRead,
		Description: "List past calls for a Google Meet space.",
		Options: []mcp.ToolOption{
			mcp.WithString("meeting_code", mcp.Description("Meeting code"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(25), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			code, err := requireMCPString(req, "meeting_code")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "meet", "history").
				num("max", "--max", 25, 1, 100).done(code)
		},
	}
}

func mcpMeetParticipantsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "meet_participants",
		Service:     "meet",
		Risk:        mcpRiskRead,
		Description: "List participants from the latest Google Meet call.",
		Options: []mcp.ToolOption{
			mcp.WithString("meeting_code", mcp.Description("Meeting code"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(100)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			code, err := requireMCPString(req, "meeting_code")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "meet", "participants").
				num("max", "--max", 50, 1, 100).done(code)
		},
	}
}

func mcpMeetCreateTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "meet_create",
		Service:     "meet",
		Risk:        mcpRiskWrite,
		Description: "Create a Google Meet space. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("access", mcp.Description("Access type"), mcp.Enum("OPEN", "TRUSTED", "RESTRICTED")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "meet", "create").
				str("access", "--access").done()
		},
	}
}
