package cmd

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpCalendarTools returns the Google Calendar MCP tool surface (service "calendar").
func mcpCalendarTools() []mcpToolSpec {
	return []mcpToolSpec{
		mcpCalendarEventsTool(),
		mcpCalendarListCalendarsTool(),
		mcpCalendarGetEventTool(),
		mcpCalendarSearchTool(),
		mcpCalendarFreeBusyTool(),
		mcpCalendarCreateEventTool(),
		mcpCalendarUpdateEventTool(),
		mcpCalendarDeleteEventTool(),
		mcpCalendarRespondTool(),
	}
}

func mcpCalendarListCalendarsTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_list_calendars",
		Service:     "calendar",
		Risk:        mcpRiskRead,
		Description: "List the calendars on the account.",
		Options: []mcp.ToolOption{
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(50), mcp.Min(1), mcp.Max(250)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			return mcpCommand(req, "calendar", "calendars").
				num("max", "--max", 50, 1, 250).done()
		},
	}
}

func mcpCalendarGetEventTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_get_event",
		Service:     "calendar",
		Risk:        mcpRiskRead,
		Description: "Get a single calendar event by ID.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_id", mcp.Description("Calendar ID"), mcp.Required()),
			mcp.WithString("event_id", mcp.Description("Event ID"), mcp.Required()),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			calendarID, err := requireMCPString(req, "calendar_id")
			if err != nil {
				return nil, err
			}
			eventID, err := requireMCPString(req, "event_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "event").done(calendarID, eventID)
		},
	}
}

//nolint:dupl // coincidental structural twin of searchconsole_query; unrelated domains, no shared helper warranted
func mcpCalendarSearchTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_search",
		Service:     "calendar",
		Risk:        mcpRiskRead,
		Description: "Search calendar events by free text.",
		Options: []mcp.ToolOption{
			mcp.WithString("query", mcp.Description("Free-text search query"), mcp.Required()),
			mcp.WithString("calendar", mcp.Description("Calendar ID or selector; default primary")),
			mcp.WithString("from", mcp.Description("Start time: RFC3339, date, or relative value")),
			mcp.WithString("to", mcp.Description("End time: RFC3339, date, or relative value")),
			mcp.WithInteger("max", mcp.Description("Maximum results"), mcp.DefaultNumber(20), mcp.Min(1), mcp.Max(250)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			query, err := requireMCPString(req, "query")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "search").
				str("calendar", "--calendar").
				str("from", "--from").
				str("to", "--to").
				num("max", "--max", 20, 1, 250).done(query)
		},
	}
}

func mcpCalendarFreeBusyTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_freebusy",
		Service:     "calendar",
		Risk:        mcpRiskRead,
		Description: "Get free/busy information for one or more calendars.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_ids", mcp.Description("Calendar ID (or space-separated IDs)"), mcp.Required()),
			mcp.WithString("from", mcp.Description("Start time: RFC3339, date, or relative value")),
			mcp.WithString("to", mcp.Description("End time: RFC3339, date, or relative value")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			calendarIDs, err := requireMCPString(req, "calendar_ids")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "freebusy").
				str("from", "--from").
				str("to", "--to").done(calendarIDs)
		},
	}
}

func mcpCalendarCreateEventTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_create_event",
		Service:     "calendar",
		Risk:        mcpRiskWrite,
		Description: "Create a calendar event. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_id", mcp.Description("Calendar ID"), mcp.Required()),
			mcp.WithString("summary", mcp.Description("Event title"), mcp.Required()),
			mcp.WithString("from", mcp.Description("Start time: RFC3339, date, or relative value")),
			mcp.WithString("to", mcp.Description("End time: RFC3339, date, or relative value")),
			mcp.WithBoolean("all_day", mcp.Description("All-day event"), mcp.DefaultBool(false)),
			mcp.WithString("description", mcp.Description("Event description")),
			mcp.WithString("location", mcp.Description("Event location")),
			mcp.WithString("attendees", mcp.Description("Comma-separated attendee emails")),
			mcp.WithBoolean("with_meet", mcp.Description("Attach a Google Meet conference"), mcp.DefaultBool(false)),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			calendarID, err := requireMCPString(req, "calendar_id")
			if err != nil {
				return nil, err
			}
			summary, err := requireMCPText(req, "summary")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "create", "--summary", summary).
				str("from", "--from").
				str("to", "--to").
				flag("all_day", "--all-day").
				str("description", "--description").
				str("location", "--location").
				str("attendees", "--attendees").
				flag("with_meet", "--with-meet").done(calendarID)
		},
	}
}

func mcpCalendarUpdateEventTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_update_event",
		Service:     "calendar",
		Risk:        mcpRiskWrite,
		Description: "Update a calendar event. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_id", mcp.Description("Calendar ID"), mcp.Required()),
			mcp.WithString("event_id", mcp.Description("Event ID"), mcp.Required()),
			mcp.WithString("summary", mcp.Description("New event title")),
			mcp.WithString("from", mcp.Description("New start time")),
			mcp.WithString("to", mcp.Description("New end time")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("location", mcp.Description("New location")),
			mcp.WithString("add_attendee", mcp.Description("Attendee email to add")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			calendarID, err := requireMCPString(req, "calendar_id")
			if err != nil {
				return nil, err
			}
			eventID, err := requireMCPString(req, "event_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "update").
				str("summary", "--summary").
				str("from", "--from").
				str("to", "--to").
				str("description", "--description").
				str("location", "--location").
				str("add_attendee", "--add-attendee").done(calendarID, eventID)
		},
	}
}

func mcpCalendarDeleteEventTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_delete_event",
		Service:     "calendar",
		Risk:        mcpRiskWrite,
		Description: "Delete a calendar event. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_id", mcp.Description("Calendar ID"), mcp.Required()),
			mcp.WithString("event_id", mcp.Description("Event ID"), mcp.Required()),
			mcp.WithString("send_updates", mcp.Description("Notify guests"), mcp.Enum("all", "externalOnly", "none")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			calendarID, err := requireMCPString(req, "calendar_id")
			if err != nil {
				return nil, err
			}
			eventID, err := requireMCPString(req, "event_id")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "delete").
				str("send_updates", "--send-updates").done(calendarID, eventID)
		},
	}
}

func mcpCalendarRespondTool() mcpToolSpec {
	return mcpToolSpec{
		Name:        "calendar_respond",
		Service:     "calendar",
		Risk:        mcpRiskWrite,
		Description: "Respond to a calendar event invitation. Requires --allow-write.",
		Options: []mcp.ToolOption{
			mcp.WithString("calendar_id", mcp.Description("Calendar ID"), mcp.Required()),
			mcp.WithString("event_id", mcp.Description("Event ID"), mcp.Required()),
			mcp.WithString("status", mcp.Description("Response status"), mcp.Enum("accepted", "declined", "tentative"), mcp.Required()),
			mcp.WithString("comment", mcp.Description("Optional response comment")),
		},
		BuildArgs: func(req mcp.CallToolRequest) ([]string, error) {
			calendarID, err := requireMCPString(req, "calendar_id")
			if err != nil {
				return nil, err
			}
			eventID, err := requireMCPString(req, "event_id")
			if err != nil {
				return nil, err
			}
			status, err := requireMCPString(req, "status")
			if err != nil {
				return nil, err
			}
			return mcpCommand(req, "calendar", "respond", "--status", status).
				str("comment", "--comment").done(calendarID, eventID)
		},
	}
}
