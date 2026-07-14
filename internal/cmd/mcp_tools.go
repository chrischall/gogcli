package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpAllTools returns every typed MCP tool, grouped by service area. Each area
// contributes its own slice so the surface stays "gated by area": --allow-tool
// selectors such as gmail, drive.*, or calendar map directly onto these groups,
// and --allow-write hides the write tools within each area until requested.
func mcpAllTools() []mcpToolSpec {
	generated, err := mcpGeneratedTools()
	if err != nil {
		// Annotation mistakes are programmer errors caught by tests; a broken
		// grammar must never serve a partial tool surface.
		panic(fmt.Sprintf("invalid mcp annotations: %v", err))
	}
	groups := [][]mcpToolSpec{
		mcpCalendarTools(),
		mcpContactsTools(),
		mcpPeopleTools(),
		mcpTasksTools(),
		mcpChatTools(),
		mcpKeepTools(),
		mcpMeetTools(),
		mcpFormsTools(),
		mcpClassroomTools(),
		mcpPhotosTools(),
		mcpMapsTools(),
		mcpYouTubeTools(),
		mcpSearchConsoleTools(),
		mcpAdminTools(),
		mcpGroupsTools(),
		mcpAppScriptTools(),
		mcpAPITools(),
		mcpCustomTools(),
		generated,
	}
	var all []mcpToolSpec
	for _, group := range groups {
		all = append(all, group...)
	}
	return all
}

// mcpArgs is a small fluent helper for assembling a child gog command's argv
// from a typed MCP request. Optional flags are only appended when present, and
// positional arguments are emitted after a `--` separator so model-supplied
// values can never be parsed as flags.
type mcpArgs struct {
	req mcp.CallToolRequest
	out []string
	err error
}

func mcpCommand(req mcp.CallToolRequest, sub ...string) *mcpArgs {
	return &mcpArgs{req: req, out: append([]string{}, sub...)}
}

// str appends "--flag value" when the named string arg is present and non-empty.
func (a *mcpArgs) str(key, flag string) *mcpArgs {
	if v := strings.TrimSpace(a.req.GetString(key, "")); v != "" {
		a.out = append(a.out, flag, v)
	}
	return a
}

// flag appends a bare "--flag" when the named boolean arg is true.
func (a *mcpArgs) flag(key, name string) *mcpArgs {
	if a.req.GetBool(key, false) {
		a.out = append(a.out, name)
	}
	return a
}

// num appends "--flag N" using the request value (or def), clamped to [1,hi].
//
//nolint:unparam // key: remaining hand-written callers all pass "max"; the helper is deleted once every service is annotation-generated
func (a *mcpArgs) num(key, flag string, def, hi int) *mcpArgs {
	a.out = append(a.out, flag, strconv.Itoa(clampMCPInt(a.req.GetInt(key, def), 1, hi)))
	return a
}

// done finalizes the argv, appending positional arguments after `--`.
func (a *mcpArgs) done(pos ...string) ([]string, error) {
	if a.err != nil {
		return nil, a.err
	}
	out := a.out
	if len(pos) > 0 {
		out = append(out, "--")
		out = append(out, pos...)
	}
	return out, nil
}

func mcpCalendarEventsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_events",
		Service:     "calendar",
		Risk:        mcpRiskRead,
		Description: "List Google Calendar events from primary or selected calendars.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_id", mcp.Description("Calendar ID or selector; default primary")),
			mcp.WithString("from", mcp.Description("Start time: RFC3339, date, or relative value")),
			mcp.WithString("to", mcp.Description("End time: RFC3339, date, or relative value")),
			mcp.WithBoolean("today", mcp.Description("Today only"), mcp.DefaultBool(false)),
			mcp.WithBoolean("tomorrow", mcp.Description("Tomorrow only"), mcp.DefaultBool(false)),
			mcp.WithInteger("days", mcp.Description("Next N days"), mcp.DefaultNumber(0), mcp.Min(0), mcp.Max(31)),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(10), mcp.Min(1), mcp.Max(250)),
			mcp.WithString("query", mcp.Description("Free text search")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			args := []string{"calendar", "events"}
			calendarID := strings.TrimSpace(req.GetString("calendar_id", ""))
			for _, pair := range [][2]string{{"from", "--from"}, {"to", "--to"}, {"query", "--query"}} {
				if v := strings.TrimSpace(req.GetString(pair[0], "")); v != "" {
					args = append(args, pair[1], v)
				}
			}
			if req.GetBool("today", false) {
				args = append(args, "--today")
			}
			if req.GetBool("tomorrow", false) {
				args = append(args, "--tomorrow")
			}
			if days := req.GetInt("days", 0); days > 0 {
				args = append(args, "--days", strconv.Itoa(clampMCPInt(days, 1, 31)))
			}
			args = append(args, "--max", strconv.Itoa(clampMCPInt(req.GetInt("max", 10), 1, 250)))
			if calendarID != "" {
				args = append(args, "--", calendarID)
			}
			return args, nil
		},
	}
}

func requireMCPText(req mcp.CallToolRequest, key string) (string, error) {
	value, err := req.RequireString(key)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("empty %s", key)
	}
	return value, nil
}

func requireMCPLiteralValuesJSON(req mcp.CallToolRequest, key string) (string, error) {
	value, err := requireMCPText(req, key)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "-" || strings.HasPrefix(trimmed, "@") {
		return "", fmt.Errorf("%s must be literal JSON, not stdin or @file input", key)
	}
	var rows [][]any
	dec := json.NewDecoder(bytes.NewReader([]byte(trimmed)))
	dec.UseNumber()
	if unmarshalErr := dec.Decode(&rows); unmarshalErr != nil {
		return "", fmt.Errorf("invalid %s JSON 2D array: %w", key, unmarshalErr)
	}
	var extra any
	if extraErr := dec.Decode(&extra); extraErr != io.EOF {
		return "", fmt.Errorf("invalid %s JSON 2D array: trailing content", key)
	}
	canonical, err := json.Marshal(rows)
	if err != nil {
		return "", fmt.Errorf("canonicalize %s: %w", key, err)
	}
	return string(canonical), nil
}
