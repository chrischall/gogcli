---
title: MCP server
description: "Expose typed, allowlisted gog tools to MCP clients without a generic command runner."
---

# MCP server

`gog mcp` runs a Model Context Protocol server over stdio. It is for agent
clients that need Google Workspace tools but should not receive a generic shell
or arbitrary `gog` command bridge.

The server registers a comprehensive set of typed tools such as `gmail_search`,
`docs_get`, and `sheets_read_range`, spanning every major Google Workspace area:
Gmail, Drive, Docs, Sheets, Slides, Calendar, Contacts, People, Tasks, Chat,
Keep, Meet, Forms, Classroom, Photos, Maps, YouTube, Search Console, Admin
(Directory), Cloud Identity Groups, Apps Script, and the Discovery API. Each
tool has a fixed schema, maps to one specific `gog` operation, and returns a
structured result containing the tool name, service, risk level, exit code,
parsed stdout, and stderr.

The surface is gated by area: every tool carries a `service` (such as `gmail`,
`slides`, or `sheets`), and `--allow-tool` selectors filter on it. For
role-oriented bundles that span several services, `--tool-suite` selects curated
suites like `developer`, `admin`, or `workspace`. Start a server scoped to just
what an agent needs, and add `--allow-write` only when write tools should be
exposed.

## Quick start

Start a read-only server for one account:

```bash
gog --account you@example.com mcp
```

List the tools this server would expose and exit:

```bash
gog --account you@example.com mcp --list-tools
```

Limit the server to Gmail search and Docs reads:

```bash
gog --account you@example.com mcp \
  --allow-tool gmail_search,docs_get
```

Expose Docs read/write tools:

```bash
gog --account you@example.com mcp \
  --allow-write \
  --allow-tool 'docs.*'
```

`--allow-write` is always required for write tools. A write tool that matches
`--allow-tool` is still hidden until `--allow-write` is present.

The exception is an explicit persistent MCP policy. It can authorize a narrow
write surface without repeating `--allow-write` in every client definition;
runtime flags can only reduce that configured surface.

## Why this is not `gog_exec`

MCP clients are often LLM-driven. A generic "run this command" tool would expose
every current and future CLI behavior through one broad capability, including
commands that were not reviewed for MCP use.

`gog mcp` uses a narrower contract:

- no generic command execution tool
- no model-supplied argv passthrough
- fixed tool schemas validated before command execution, including required
  fields, types, and rejection of unknown fields
- read-only tools by default
- write tools require explicit server startup flags
- existing `gog` account, auth, dry-run, no-input, and command safety flags are
  preserved

This keeps MCP useful for agents while making the permission surface visible at
server startup.

## Tool selection

By default, all read tools are registered and write tools are hidden.

Use `--allow-tool` to narrow the registered set. Values can be comma-separated
or repeated:

```bash
gog mcp --allow-tool gmail_search --allow-tool docs_get
gog mcp --allow-tool gmail_search,docs_get
```

Accepted selectors:

| Selector | Meaning |
| --- | --- |
| `gmail_search` | One exact tool |
| `gmail` | All Gmail tools allowed by risk mode |
| `gmail.*` | All Gmail tools allowed by risk mode |
| `read` | All read tools |
| `write` | All write tools, only when `--allow-write` is also set |
| `*` or `all` | All tools allowed by risk mode |

Examples:

```bash
# Read-only Gmail tools.
gog mcp --allow-tool gmail

# Only Docs tools, including writes.
gog mcp --allow-write --allow-tool 'docs.*'

# Read-only server, but only Calendar and Sheets reads.
gog mcp --allow-tool calendar,sheets

# All current write tools. Read tools are not included unless also selected.
gog mcp --allow-write --allow-tool write
```

## Persistent capability policy

For several MCP clients or accounts, put the maximum registered tool surface in
`config.json` instead of duplicating capability arguments. Without an `mcp`
block, behavior is unchanged: all read tools are available, writes require
`--allow-write`, and `--allow-tool` filters the result.

```json5
{
  "mcp": {
    "allow_tools": ["read"],
    "allow_write": false,
    "accounts": {
      "personal@example.com": {
        "allow_tools": ["read", "docs.*", "calendar.*"],
        "allow_write": true
      },
      "work@example.com": {
        "allow_tools": ["read"],
        "allow_write": false
      }
    }
  }
}
```

An account entry is a complete replacement for the global policy, not a partial
merge. Account keys are matched case-insensitively after aliases and automatic
account selection are resolved, then that resolved account is pinned for every
MCP child command. Per-account policies require stored account credentials;
direct access tokens and ADC can use only the global policy because an account
label does not prove the authenticated principal. An omitted `allow_tools` value defaults to
`["read"]`; an explicitly empty list is rejected. `allow_write: true` requires
an explicit tool list so a typo cannot accidentally expose every write tool.

The configured policy is a ceiling. `--allow-tool` can intersect it with a
smaller runtime set, `--readonly` removes all writes, and `--allow-write` cannot
widen a read-only policy. Baked safety profiles remain the outer immutable
ceiling. Unknown configured selectors and attempted write widening fail before
the MCP server starts. Use `gog mcp --list-tools` with the same account and flags
to inspect the final registered surface.

## Tool suites

`--tool-suite` (alias `--suite`) exposes curated, cross-service bundles in one
flag, so a client can request a whole role's worth of tools without naming each
service. Values are comma-separated or repeated.

| Suite | Services |
| --- | --- |
| `workspace` | gmail, calendar, drive, docs, sheets, slides, contacts, people, tasks, chat, keep, meet, forms |
| `developer` | appscript, api |
| `admin` | admin, groups |
| `education` | classroom |
| `media` | photos, youtube |
| `insights` | searchconsole |

```bash
# Developer suite, read-only.
gog mcp --tool-suite developer

# Admin suite with writes (Directory API needs domain-wide delegation).
gog mcp --tool-suite admin --allow-write

# Multiple suites at once.
gog mcp --tool-suite developer,admin

# Print the suite-to-service map and exit.
gog mcp --list-suites
```

A suite acts as a service-level filter. It composes with the other flags by
intersection: `--allow-write` still gates write tools, and `--allow-tool`
narrows further *within* the suite. An unknown suite name fails at startup.

```bash
# Within the workspace suite, only Slides tools.
gog mcp --tool-suite workspace --allow-tool slides
```

Every service is still independently selectable with `--allow-tool` regardless
of suite membership; suites are a convenience layer, not a separate permission
system. `--tool-suite admin` and `--tool-suite developer` expose the Admin
(Directory), Cloud Identity Groups, Apps Script, and Discovery API tools, which
are not part of the default read-only surface unless selected.

## Tools by area

Tools are grouped by service area. Read tools are registered by default; write
tools (marked _write_) are hidden unless `--allow-write` is also set. This list
is a summary; run `gog mcp --allow-write --list-tools` for the authoritative
set with full schemas.

| Area | Read tools | Write tools |
| --- | --- | --- |
| `gmail` | `gmail_search`, `gmail_search_threads`, `gmail_get_message`, `gmail_get_thread`, `gmail_list_labels`, `gmail_get_label`, `gmail_list_drafts`, `gmail_get_draft`, `gmail_history` | `gmail_send`, `gmail_reply`, `gmail_forward`, `gmail_create_draft`, `gmail_send_draft`, `gmail_modify_message`, `gmail_modify_thread`, `gmail_trash`, `gmail_create_label` |
| `drive` | `drive_search`, `drive_get`, `drive_list`, `drive_list_drives`, `drive_permissions`, `drive_list_comments`, `drive_list_revisions`, `drive_tree` | `drive_create_folder`, `drive_copy`, `drive_move`, `drive_rename`, `drive_trash`, `drive_share`, `drive_create_comment` |
| `docs` | `docs_get`, `docs_info`, `docs_list_tabs`, `docs_list_comments` | `docs_write`, `docs_create`, `docs_find_replace`, `docs_add_comment` |
| `sheets` | `sheets_read_range`, `sheets_metadata`, `sheets_list_tables`, `sheets_list_named_ranges` | `sheets_update_range`, `sheets_append`, `sheets_clear`, `sheets_create`, `sheets_add_tab` |
| `slides` | `slides_info`, `slides_list_slides`, `slides_read_slide` | `slides_create`, `slides_create_from_markdown`, `slides_replace_text`, `slides_duplicate_slide`, `slides_delete_slide` |
| `calendar` | `calendar_events`, `calendar_list_calendars`, `calendar_get_event`, `calendar_search`, `calendar_freebusy` | `calendar_create_event`, `calendar_update_event`, `calendar_delete_event`, `calendar_respond` |
| `contacts` | `contacts_list`, `contacts_get`, `contacts_search`, `contacts_directory_search` | `contacts_create`, `contacts_update`, `contacts_delete` |
| `people` | `people_me`, `people_get`, `people_search` | — |
| `tasks` | `tasks_lists`, `tasks_list`, `tasks_get` | `tasks_add`, `tasks_complete`, `tasks_update`, `tasks_delete`, `tasks_create_list` |
| `chat` | `chat_list_spaces`, `chat_find_spaces`, `chat_list_messages`, `chat_list_threads` | `chat_send_message`, `chat_send_dm` |
| `keep` | `keep_list`, `keep_get`, `keep_search` | `keep_create` |
| `meet` | `meet_get`, `meet_history`, `meet_participants` | `meet_create` |
| `forms` | `forms_get`, `forms_list_responses`, `forms_get_response` | `forms_create`, `forms_update`, `forms_add_question` |
| `classroom` | `classroom_list_courses`, `classroom_get_course`, `classroom_list_coursework`, `classroom_get_coursework`, `classroom_list_students`, `classroom_list_teachers`, `classroom_roster`, `classroom_list_announcements`, `classroom_list_submissions`, `classroom_list_topics` | — |
| `photos` | `photos_list`, `photos_get`, `photos_search` | — |
| `maps` | `maps_geocode`, `maps_reverse_geocode`, `maps_directions`, `maps_distance`, `maps_places_search`, `maps_place_details` | — |
| `youtube` | `youtube_search`, `youtube_list_videos`, `youtube_list_channels`, `youtube_list_playlists`, `youtube_list_playlist_items`, `youtube_list_comments` | `youtube_create_playlist`, `youtube_add_to_playlist` |
| `searchconsole` | `searchconsole_list_sites`, `searchconsole_query`, `searchconsole_list_sitemaps` | — |
| `admin` | `admin_list_users`, `admin_get_user`, `admin_list_groups`, `admin_list_group_members`, `admin_list_orgunits`, `admin_get_orgunit` | `admin_add_group_member`, `admin_remove_group_member`, `admin_suspend_user` |
| `groups` | `groups_list`, `groups_members` | — |
| `appscript` | `appscript_get`, `appscript_content` | `appscript_create`, `appscript_run` |
| `api` | `api_list`, `api_describe` | `api_call` |

Select a whole area with `--allow-tool <area>` (for example `--allow-tool
slides` or `--allow-tool 'docs.*'`), or list individual tool names. `gmail_send`,
`gmail_reply`, `gmail_forward`, and `gmail_send_draft` are additionally blocked
by `--gmail-no-send`.

The generated command reference for the server itself is
[`gog mcp`](commands/gog-mcp.md).

MCP clients discover the registered surface through the protocol's standard
`tools/list` request. For shell-side inspection before starting the server, use
`gog mcp --list-tools`; no model-callable discovery tool is added.

Tool definitions are generated from annotations on the CLI command grammar, so
each tool's schema (types, defaults, enums, bounds) cannot drift from the
command it invokes. A handful of tools with cross-field rules remain
hand-written. The registered surface is unchanged by this mechanism; it remains
fixed, typed, and allowlisted as described above.

## Client configuration

MCP clients usually need a command and an argument list. Put account selection
and safety policy on the server command, not inside tool calls.

Minimal stdio configuration:

```json
{
  "command": "gog",
  "args": ["--account", "you@example.com", "mcp"]
}
```

Read-only Docs and Sheets configuration:

```json
{
  "command": "gog",
  "args": [
    "--account", "you@example.com",
    "--enable-commands-exact", "mcp,docs.cat,sheets.get",
    "mcp",
    "--allow-tool", "docs_get,sheets_read_range"
  ]
}
```

Docs read/write configuration:

```json
{
  "command": "gog",
  "args": [
    "--account", "you@example.com",
    "--enable-commands-exact", "mcp,docs.cat,docs.write",
    "--no-input",
    "mcp",
    "--allow-write",
    "--allow-tool", "docs.*"
  ]
}
```

For headless services, set `GOG_KEYRING_BACKEND=file` and
`GOG_KEYRING_PASSWORD` on the MCP client process or service unit. A successful
interactive shell check does not prove the MCP client inherited those
variables; verify through the same process manager that launches the server.

## mcporter examples

List registered tools and their schemas:

```bash
mcporter list \
  --stdio gog \
  --stdio-arg --account \
  --stdio-arg you@example.com \
  --stdio-arg mcp \
  --stdio-arg --allow-tool \
  --stdio-arg 'docs.*' \
  --schema \
  --json
```

Dry-run a Docs write through MCP:

```bash
mcporter call \
  --stdio gog \
  --stdio-arg --account \
  --stdio-arg you@example.com \
  --stdio-arg --dry-run \
  --stdio-arg mcp \
  --stdio-arg --allow-write \
  --stdio-arg --allow-tool \
  --stdio-arg docs_write \
  docs_write \
  '{"document_id":"DOCUMENT_ID","text":"MCP smoke test\n","append":true}'
```

Read a Sheet range:

```bash
mcporter call \
  --stdio gog \
  --stdio-arg --account \
  --stdio-arg you@example.com \
  --stdio-arg mcp \
  --stdio-arg --allow-tool \
  --stdio-arg sheets_read_range \
  sheets_read_range \
  '{"spreadsheet_id":"SPREADSHEET_ID","range":"Sheet1!A1:C10"}'
```

Update a Sheet range:

```bash
mcporter call \
  --stdio gog \
  --stdio-arg --account \
  --stdio-arg you@example.com \
  --stdio-arg mcp \
  --stdio-arg --allow-write \
  --stdio-arg --allow-tool \
  --stdio-arg sheets_update_range \
  sheets_update_range \
  '{"spreadsheet_id":"SPREADSHEET_ID","range":"Sheet1!A1:B1","values_json":"[[\"status\",\"ok\"]]","input":"RAW"}'
```

`sheets_update_range.values_json` must be literal JSON. MCP rejects `@file`,
`@-`, and `-` expansion forms so a model cannot cause the server process to
read arbitrary local files or stdin.

## Safety model

Tool calls run as subprocesses of the same `gog` executable. The server adds a
non-interactive, agent-oriented root context to every child command:

- `--json`
- `--wrap-untrusted`
- `--no-input`
- `--color=never`

The server also preserves selected parent root flags:

- `--account`
- `--client`
- `--home`
- `--dry-run`
- `--results-only`
- `--select`
- direct access tokens

And it preserves command safety flags:

- `--gmail-no-send`
- `--enable-commands`
- `--enable-commands-exact`
- `--disable-commands`

Use both MCP tool allowlists and command allowlists when the server is exposed
to an untrusted or semi-trusted agent:

```bash
gog --account you@example.com \
  --enable-commands-exact mcp,docs.cat,docs.write \
  --disable-commands gmail.send,gmail.drafts.send \
  --gmail-no-send \
  mcp \
  --allow-write \
  --allow-tool 'docs.*'
```

If a tool maps to a disabled command, the tool call returns a non-zero exit code
and the child command error in `stderr`.

## Output shape

Successful calls return structured MCP content shaped like:

```json
{
  "tool": "docs_get",
  "service": "docs",
  "risk": "read",
  "exit_code": 0,
  "stdout": {
    "documentId": "..."
  },
  "stderr": ""
}
```

If a child command prints valid JSON, `stdout` is parsed as JSON with numeric
literals preserved. Otherwise `stdout` is returned as a string. Empty stdout is
omitted.

If the child command exits non-zero, the MCP result is marked as an error and
includes the same structured fields with `exit_code` and `stderr`.

## Limits and timeouts

Each tool call has a subprocess timeout and bounded stdout/stderr capture:

```bash
gog mcp --timeout-seconds 30 --max-output-bytes 262144
```

Defaults:

- timeout: 60 seconds
- max captured stdout/stderr: 102400 bytes each

Use command-specific limits too. For example, `docs_get` has a `max_bytes`
argument, and search tools have `max` arguments.

## Authentication

The MCP server uses normal `gog` auth. Before wiring a client, verify the same
account and scopes from a shell:

```bash
gog --account you@example.com auth doctor --check
gog --account you@example.com mcp --list-tools
```

Then verify through the MCP client entrypoint. In services and desktop MCP
clients, most auth failures are environment inheritance problems: missing
`GOG_ACCOUNT`, missing file-keyring password, different `GOG_HOME`, or a
different OAuth client selected by `--client`.

## Troubleshooting

`no MCP tools enabled`

: Your `--allow-tool` filters excluded everything, or you selected only write
  tools without `--allow-write`.

`command "..." is disabled`

: The MCP tool was registered, but the child `gog` command was blocked by
  `--enable-commands`, `--enable-commands-exact`, `--disable-commands`, or a
  baked safety profile.

Tool missing in the client

: Run `gog mcp --list-tools` with the same flags. If the tool is not listed,
  fix `--allow-tool` or add `--allow-write` for write tools. If it is listed,
  refresh or restart the MCP client.

Auth works in Terminal but not in the MCP client

: Compare `--account`, `--client`, `--home`, `GOG_HOME`,
  `GOG_KEYRING_BACKEND`, and `GOG_KEYRING_PASSWORD` in the process that starts
  the MCP server.

Large output is truncated

: Increase `--max-output-bytes`, narrow the request, or use tool arguments such
  as `max`, `max_bytes`, date ranges, or Drive field masks.
