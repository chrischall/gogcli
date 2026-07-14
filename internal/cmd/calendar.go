package cmd

type CalendarCmd struct {
	Calendars       CalendarCalendarsCmd       `cmd:"" name:"calendars" help:"List calendars" mcp:"calendar_list_calendars,read" mcpdesc:"List the calendars on the account."`
	Subscribe       CalendarSubscribeCmd       `cmd:"" name:"subscribe" aliases:"sub,add-calendar" help:"Add a calendar to your calendar list"`
	Unsubscribe     CalendarUnsubscribeCmd     `cmd:"" name:"unsubscribe" aliases:"unsub" help:"Remove a calendar from your calendar list"`
	CreateCalendar  CalendarCreateCalendarCmd  `cmd:"" name:"create-calendar" aliases:"new-calendar" help:"Create a new secondary calendar"`
	DeleteCalendar  CalendarDeleteCalendarCmd  `cmd:"" name:"delete-calendar" help:"Delete an owned secondary calendar"`
	ACL             CalendarAclCmd             `cmd:"" name:"acl" aliases:"permissions,perms" help:"List calendar ACL"`
	Alias           CalendarAliasCmd           `cmd:"" name:"alias" help:"Manage calendar aliases"`
	Events          CalendarEventsCmd          `cmd:"" name:"events" aliases:"list,ls" help:"List events from a calendar or all calendars" mcp:"calendar_events,read" mcpdesc:"List Google Calendar events from primary or selected calendars."`
	Event           CalendarEventCmd           `cmd:"" name:"event" aliases:"get,info,show" help:"Get event" mcp:"calendar_get_event,read" mcpdesc:"Get a single calendar event by ID."`
	Raw             CalendarRawCmd             `cmd:"" name:"raw" help:"Dump raw Google Calendar API response as JSON (Events.Get; lossless; for scripting and LLM consumption)"`
	Create          CalendarCreateCmd          `cmd:"" name:"create" aliases:"add,new" help:"Create an event" mcp:"calendar_create_event,write" mcpdesc:"Create a calendar event. Requires --allow-write."`
	Update          CalendarUpdateCmd          `cmd:"" name:"update" aliases:"edit,set" help:"Update an event" mcp:"calendar_update_event,write" mcpdesc:"Update a calendar event. Requires --allow-write."`
	Move            CalendarMoveCmd            `cmd:"" name:"move" aliases:"transfer" help:"Move an event to another calendar"`
	Delete          CalendarDeleteCmd          `cmd:"" name:"delete" aliases:"rm,del,remove" help:"Delete an event" mcp:"calendar_delete_event,write" mcpdesc:"Delete a calendar event. Requires --allow-write."`
	FreeBusy        CalendarFreeBusyCmd        `cmd:"" name:"freebusy" help:"Get free/busy" mcp:"calendar_freebusy,read" mcpdesc:"Get free/busy information for one or more calendars."`
	Respond         CalendarRespondCmd         `cmd:"" name:"respond" aliases:"rsvp,reply" help:"Respond to an event invitation" mcp:"calendar_respond,write" mcpdesc:"Respond to a calendar event invitation. Requires --allow-write."`
	ProposeTime     CalendarProposeTimeCmd     `cmd:"" name:"propose-time" help:"Generate URL to propose a new meeting time (browser-only feature)"`
	Colors          CalendarColorsCmd          `cmd:"" name:"colors" help:"Show calendar colors"`
	Conflicts       CalendarConflictsCmd       `cmd:"" name:"conflicts" help:"Find busy-time overlaps across calendars"`
	Changed         CalendarChangedCmd         `cmd:"" name:"changed" help:"List most recently changed events (including deletions)"`
	Search          CalendarSearchCmd          `cmd:"" name:"search" aliases:"find,query" help:"Search events" mcp:"calendar_search,read" mcpdesc:"Search calendar events by free text."`
	Time            CalendarTimeCmd            `cmd:"" name:"time" help:"Show server time"`
	Users           CalendarUsersCmd           `cmd:"" name:"users" help:"List workspace users (use their email as calendar ID)"`
	Team            CalendarTeamCmd            `cmd:"" name:"team" help:"Show events for Workspace group members (service account, direct token, or ADC)"`
	FocusTime       CalendarFocusTimeCmd       `cmd:"" name:"focus-time" aliases:"focus" help:"Create a Focus Time block"`
	OOO             CalendarOOOCmd             `cmd:"" name:"out-of-office" aliases:"ooo" help:"Create an Out of Office event"`
	WorkingLocation CalendarWorkingLocationCmd `cmd:"" name:"working-location" aliases:"wl" help:"Set working location (home/office/custom)"`
}
