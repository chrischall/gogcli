package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpAdminTools returns the Google Workspace Admin (Directory API) MCP tool
// surface (service "admin"). These require domain-wide delegation.
func mcpAdminTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpAdminListUsersTool(),
		mcpAdminGetUserTool(),
		mcpAdminListGroupsTool(),
		mcpAdminListGroupMembersTool(),
		mcpAdminListOrgUnitsTool(),
		mcpAdminGetOrgUnitTool(),
		mcpAdminAddGroupMemberTool(),
		mcpAdminRemoveGroupMemberTool(),
		mcpAdminSuspendUserTool(),
	}
}

func mcpAdminListUsersTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_list_users",
		Service:     "admin",
		Risk:        mcpRiskRead,
		Description: "List Workspace users in a domain (Directory API; requires domain-wide delegation).",
		Options: []mcp.ToolOption{
			mcp.WithString("domain", mcp.Description("Domain to list users for")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "admin", "users", "list").
				str("domain", "--domain").
				num("max", "--max", 50, 200).done()
		},
	}
}

func mcpAdminGetUserTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_get_user",
		Service:     "admin",
		Risk:        mcpRiskRead,
		Description: "Get Workspace user details by email (Directory API).",
		Options: []mcp.ToolOption{
			mcp.WithString("user_email", mcp.Description("User email address"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			email, err := requireMCPString(req, "user_email")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "admin", "users", "get").done(email)
		},
	}
}

func mcpAdminListGroupsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_list_groups",
		Service:     "admin",
		Risk:        mcpRiskRead,
		Description: "List Workspace groups in a domain (Directory API).",
		Options: []mcp.ToolOption{
			mcp.WithString("domain", mcp.Description("Domain to list groups for")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "admin", "groups", "list").
				str("domain", "--domain").
				num("max", "--max", 50, 200).done()
		},
	}
}

func mcpAdminListGroupMembersTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_list_group_members",
		Service:     "admin",
		Risk:        mcpRiskRead,
		Description: "List members of a Workspace group (Directory API).",
		Options: []mcp.ToolOption{
			mcp.WithString("group_email", mcp.Description("Group email address"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			group, err := requireMCPString(req, "group_email")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "admin", "groups", "members", "list").
				num("max", "--max", 50, 200).done(group)
		},
	}
}

func mcpAdminListOrgUnitsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_list_orgunits",
		Service:     "admin",
		Risk:        mcpRiskRead,
		Description: "List organizational units (Directory API).",
		Options: []mcp.ToolOption{
			mcp.WithString("parent", mcp.Description("Parent org unit path")),
			mcp.WithString("type", mcp.Description("Org unit type filter"), mcp.Enum("all", "children")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "admin", "orgunits", "list").
				str("parent", "--parent").
				str("type", "--type").done()
		},
	}
}

func mcpAdminGetOrgUnitTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_get_orgunit",
		Service:     "admin",
		Risk:        mcpRiskRead,
		Description: "Get organizational unit details by path (Directory API).",
		Options: []mcp.ToolOption{
			mcp.WithString("path", mcp.Description("Org unit path"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			path, err := requireMCPString(req, "path")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "admin", "orgunits", "get").done(path)
		},
	}
}

func mcpAdminAddGroupMemberTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_add_group_member",
		Service:     "admin",
		Risk:        mcpRiskWrite,
		Description: "Add a member to a Workspace group (Directory API). Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("group_email", mcp.Description("Group email address"), mcp.Required()),
			mcp.WithString("member_email", mcp.Description("Member email address"), mcp.Required()),
			mcp.WithString("role", mcp.Description("Membership role"), mcp.Enum("MEMBER", "MANAGER", "OWNER")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			group, err := requireMCPString(req, "group_email")
			if err != nil {
				return nil, err
			}
			member, err := requireMCPString(req, "member_email")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "admin", "groups", "members", "add").
				str("role", "--role").done(group, member)
		},
	}
}

func mcpAdminRemoveGroupMemberTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_remove_group_member",
		Service:     "admin",
		Risk:        mcpRiskWrite,
		Description: "Remove a member from a Workspace group (Directory API). Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("group_email", mcp.Description("Group email address"), mcp.Required()),
			mcp.WithString("member_email", mcp.Description("Member email address"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			group, err := requireMCPString(req, "group_email")
			if err != nil {
				return nil, err
			}
			member, err := requireMCPString(req, "member_email")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "admin", "groups", "members", "remove").done(group, member)
		},
	}
}

func mcpAdminSuspendUserTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "admin_suspend_user",
		Service:     "admin",
		Risk:        mcpRiskWrite,
		Description: "Suspend a Workspace user account (Directory API; reversible via the admin console). Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("user_email", mcp.Description("User email address"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			email, err := requireMCPString(req, "user_email")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "admin", "users", "suspend").done(email)
		},
	}
}

// mcpGroupsTools returns the Cloud Identity Groups MCP tool surface (service
// "groups").
func mcpGroupsTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpGroupsListTool(),
		mcpGroupsMembersTool(),
	}
}

func mcpGroupsListTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "groups_list",
		Service:     "groups",
		Risk:        mcpRiskRead,
		Description: "List Cloud Identity groups you belong to.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "groups", "list").
				num("max", "--max", 50, 200).done()
		},
	}
}

func mcpGroupsMembersTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "groups_members",
		Service:     "groups",
		Risk:        mcpRiskRead,
		Description: "List members of a Cloud Identity group.",
		Options: []mcp.ToolOption{
			mcp.WithString("group_email", mcp.Description("Group email address"), mcp.Required()),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(200)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			group, err := requireMCPString(req, "group_email")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "groups", "members").
				num("max", "--max", 50, 200).done(group)
		},
	}
}
