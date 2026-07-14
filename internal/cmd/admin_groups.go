package cmd

import (
	"context"
	"fmt"
	"strings"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/steipete/gogcli/internal/outfmt"
	"github.com/steipete/gogcli/internal/ui"
)

// AdminGroupsCmd manages Workspace groups.
type AdminGroupsCmd struct {
	List    AdminGroupsListCmd    `cmd:"" name:"list" aliases:"ls" help:"List groups in a domain" mcp:"admin_list_groups,read" mcpdesc:"List Workspace groups in a domain (Directory API)."`
	Members AdminGroupsMembersCmd `cmd:"" name:"members" help:"Manage group members"`
}

type AdminGroupsListCmd struct {
	Domain    string `name:"domain" help:"Domain to list groups from (e.g., example.com)" mcp:"domain" mcpdesc:"Domain to list groups for"`
	Max       int64  `name:"max" aliases:"limit" help:"Max results" default:"100" mcp:"max,default=50,min=1,max=200" mcpdesc:"Maximum results"`
	Page      string `name:"page" aliases:"cursor" help:"Page token"`
	All       bool   `name:"all" aliases:"all-pages,allpages" help:"Fetch all pages"`
	FailEmpty bool   `name:"fail-empty" aliases:"non-empty,require-results" help:"Exit with code 3 if no results"`
}

func (c *AdminGroupsListCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	domain := strings.TrimSpace(c.Domain)
	if domain == "" {
		return usage("domain required (e.g., --domain example.com)")
	}
	if c.Max <= 0 {
		return usage("max must be > 0")
	}
	account, err := requireAdminAccount(flags)
	if err != nil {
		return err
	}

	svc, err := adminDirectoryService(ctx, account)
	if err != nil {
		return wrapAdminDirectoryError(err, account)
	}

	fetch := func(pageToken string) ([]*admin.Group, string, error) {
		call := svc.Groups.List().
			Domain(domain).
			MaxResults(c.Max).
			Context(ctx)
		if strings.TrimSpace(pageToken) != "" {
			call = call.PageToken(pageToken)
		}
		resp, fetchErr := call.Do()
		if fetchErr != nil {
			return nil, "", wrapAdminDirectoryError(fetchErr, account)
		}
		return resp.Groups, resp.NextPageToken, nil
	}

	groups, nextPageToken, err := loadPagedItems(c.Page, c.All, fetch)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		type item struct {
			Email              string `json:"email"`
			Name               string `json:"name,omitempty"`
			Description        string `json:"description,omitempty"`
			DirectMembersCount int64  `json:"directMembersCount"`
		}
		items := make([]item, 0, len(groups))
		for _, group := range groups {
			if group == nil {
				continue
			}
			items = append(items, item{
				Email:              group.Email,
				Name:               group.Name,
				Description:        group.Description,
				DirectMembersCount: group.DirectMembersCount,
			})
		}
		if err := outfmt.WriteJSON(ctx, stdoutWriter(ctx), map[string]any{
			"groups":        items,
			"nextPageToken": nextPageToken,
		}); err != nil {
			return err
		}
		if len(items) == 0 {
			return failEmptyExit(c.FailEmpty)
		}
		return nil
	}

	if len(groups) == 0 {
		u.Err().Println("No groups found")
		return failEmptyExit(c.FailEmpty)
	}

	if err := outfmt.WriteTable(ctx, stdoutWriter(ctx), nonNilAdminRows(groups), adminGroupColumns()); err != nil {
		return err
	}
	printNextPageHintWithAll(u, nextPageToken, "--all/--all-pages")
	return nil
}

type AdminGroupsMembersCmd struct {
	List   AdminGroupsMembersListCmd   `cmd:"" name:"list" aliases:"ls" help:"List group members" mcp:"admin_list_group_members,read" mcpdesc:"List members of a Workspace group (Directory API)."`
	Add    AdminGroupsMembersAddCmd    `cmd:"" name:"add" aliases:"invite" help:"Add a member to a group" mcp:"admin_add_group_member,write" mcpdesc:"Add a member to a Workspace group (Directory API). Requires --allow-write."`
	Remove AdminGroupsMembersRemoveCmd `cmd:"" name:"remove" aliases:"rm,del,delete" help:"Remove a member from a group" mcp:"admin_remove_group_member,write" mcpdesc:"Remove a member from a Workspace group (Directory API). Requires --allow-write."`
}

type AdminGroupsMembersListCmd struct {
	GroupEmail string `arg:"" name:"groupEmail" help:"Group email (e.g., engineering@example.com)" mcp:"group_email" mcpdesc:"Group email address"`
	Max        int64  `name:"max" aliases:"limit" help:"Max results" default:"100" mcp:"max,default=50,min=1,max=200" mcpdesc:"Maximum results"`
	Page       string `name:"page" aliases:"cursor" help:"Page token"`
	All        bool   `name:"all" aliases:"all-pages,allpages" help:"Fetch all pages"`
	FailEmpty  bool   `name:"fail-empty" aliases:"non-empty,require-results" help:"Exit with code 3 if no results"`
}

func (c *AdminGroupsMembersListCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	groupEmail := strings.TrimSpace(c.GroupEmail)
	if groupEmail == "" {
		return usage("group email required")
	}
	if c.Max <= 0 {
		return usage("max must be > 0")
	}
	account, err := requireAdminAccount(flags)
	if err != nil {
		return err
	}

	svc, err := adminDirectoryService(ctx, account)
	if err != nil {
		return wrapAdminDirectoryError(err, account)
	}

	fetch := func(pageToken string) ([]*admin.Member, string, error) {
		call := svc.Members.List(groupEmail).
			MaxResults(c.Max).
			Context(ctx)
		if strings.TrimSpace(pageToken) != "" {
			call = call.PageToken(pageToken)
		}
		resp, fetchErr := call.Do()
		if fetchErr != nil {
			return nil, "", wrapAdminDirectoryError(fetchErr, account)
		}
		return resp.Members, resp.NextPageToken, nil
	}

	members, nextPageToken, err := loadPagedItems(c.Page, c.All, fetch)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		type item struct {
			Email string `json:"email"`
			Role  string `json:"role"`
			Type  string `json:"type"`
		}
		items := make([]item, 0, len(members))
		for _, member := range members {
			if member == nil {
				continue
			}
			items = append(items, item{
				Email: member.Email,
				Role:  member.Role,
				Type:  member.Type,
			})
		}
		if err := outfmt.WriteJSON(ctx, stdoutWriter(ctx), map[string]any{
			"members":       items,
			"nextPageToken": nextPageToken,
		}); err != nil {
			return err
		}
		if len(items) == 0 {
			return failEmptyExit(c.FailEmpty)
		}
		return nil
	}

	if len(members) == 0 {
		u.Err().Println("No members found")
		return failEmptyExit(c.FailEmpty)
	}

	if err := outfmt.WriteTable(ctx, stdoutWriter(ctx), nonNilAdminRows(members), adminMemberColumns()); err != nil {
		return err
	}
	printNextPageHintWithAll(u, nextPageToken, "--all/--all-pages")
	return nil
}

type AdminGroupsMembersAddCmd struct {
	GroupEmail  string `arg:"" name:"groupEmail" help:"Group email" mcp:"group_email" mcpdesc:"Group email address"`
	MemberEmail string `arg:"" name:"memberEmail" help:"Member email to add" mcp:"member_email" mcpdesc:"Member email address"`
	Role        string `name:"role" help:"Member role (MEMBER, MANAGER, OWNER)" default:"MEMBER" mcp:"role,enum=MEMBER|MANAGER|OWNER" mcpdesc:"Membership role"`
}

func (c *AdminGroupsMembersAddCmd) Run(ctx context.Context, flags *RootFlags) error {
	u := ui.FromContext(ctx)
	groupEmail := strings.TrimSpace(c.GroupEmail)
	memberEmail := strings.TrimSpace(c.MemberEmail)
	if groupEmail == "" || memberEmail == "" {
		return usage("group email and member email required")
	}
	if err := validatePlainEmail("member email", memberEmail); err != nil {
		return err
	}

	role := strings.ToUpper(c.Role)
	if role != adminRoleMember && role != adminRoleManager && role != adminRoleOwner {
		return usage("role must be MEMBER, MANAGER, or OWNER")
	}

	member := &admin.Member{
		Email: memberEmail,
		Role:  role,
	}

	if dryRunErr := dryRunExit(ctx, flags, "admin.groups.members.add", map[string]any{
		"group":  groupEmail,
		"member": member,
	}); dryRunErr != nil {
		return dryRunErr
	}

	account, err := requireAdminAccount(flags)
	if err != nil {
		return err
	}

	svc, err := adminDirectoryService(ctx, account)
	if err != nil {
		return wrapAdminDirectoryError(err, account)
	}

	created, err := svc.Members.Insert(groupEmail, member).Context(ctx).Do()
	if err != nil {
		return wrapAdminDirectoryError(err, account)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, stdoutWriter(ctx), map[string]any{
			"email": created.Email,
			"role":  created.Role,
		})
	}

	u.Out().Linef("Added %s to %s as %s", created.Email, groupEmail, created.Role)
	return nil
}

type AdminGroupsMembersRemoveCmd struct {
	GroupEmail  string `arg:"" name:"groupEmail" help:"Group email" mcp:"group_email" mcpdesc:"Group email address"`
	MemberEmail string `arg:"" name:"memberEmail" help:"Member email to remove" mcp:"member_email" mcpdesc:"Member email address"`
}

func (c *AdminGroupsMembersRemoveCmd) Run(ctx context.Context, flags *RootFlags) error {
	groupEmail := strings.TrimSpace(c.GroupEmail)
	memberEmail := strings.TrimSpace(c.MemberEmail)
	if groupEmail == "" || memberEmail == "" {
		return usage("group email and member email required")
	}

	if confirmErr := dryRunAndConfirmDestructive(ctx, flags, "admin.groups.members.remove", map[string]any{
		"group":  groupEmail,
		"member": memberEmail,
	}, fmt.Sprintf("remove %s from %s", memberEmail, groupEmail)); confirmErr != nil {
		return confirmErr
	}

	account, err := requireAdminAccount(flags)
	if err != nil {
		return err
	}

	svc, err := adminDirectoryService(ctx, account)
	if err != nil {
		return wrapAdminDirectoryError(err, account)
	}

	if err := svc.Members.Delete(groupEmail, memberEmail).Context(ctx).Do(); err != nil {
		return wrapAdminDirectoryError(err, account)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(ctx, stdoutWriter(ctx), map[string]any{
			"removed": true,
			"email":   memberEmail,
			"group":   groupEmail,
		})
	}

	u := ui.FromContext(ctx)
	u.Out().Linef("Removed %s from %s", memberEmail, groupEmail)
	return nil
}
