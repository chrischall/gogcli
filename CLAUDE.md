# CLAUDE.md

**This repo is a fork, not our project.** `chrischall/gogcli` tracks
[`openclaw/gogcli`](https://github.com/openclaw/gogcli) — Peter Steinberger's
`gog` Google Workspace CLI (Go module path is still `github.com/steipete/gogcli`).
As of 2026-07-19 `origin/main` matched `upstream/main` exactly (both at
`4747fb05`, v0.34.1); apart from this file the fork carries no local commits and
is only a mirror plus a place to push contribution branches.

**This file is fork-local — never include it in an upstream PR.** It is the one
deliberate divergence from `upstream/main`, so syncing needs a rebase rather
than a fast-forward:
`git fetch upstream && git rebase upstream/main`. Keep it that way: one commit
of ours on top of an otherwise untouched mirror. Cut contribution branches from
`upstream/main`, not from local `main`, so this file never rides along. Commit
nothing else to the fork's `main`.

**Upstream's [`AGENTS.md`](AGENTS.md) is the authority** on structure, style,
commit conventions and the review/land flow. Follow it. Do not restyle,
reorganize, or apply our house conventions to upstream's code — changes here are
proposed to someone else's project and get reviewed by its maintainer.

## Build and test

`make` builds `bin/gog`. `make ci` is the gate CI enforces, and three of its
steps fail on things a plain `go build && go test` never surfaces:

- `deadcode` — any unreachable function fails the build, and it runs **twice**:
  natively, then cross-compiled `GOOS=linux GOARCH=amd64`. Code reachable only
  on darwin still trips the linux pass.
- `docs-check` / `agent-skills-check` — `docs/commands/*` and the generated
  agent skills are committed artifacts. Adding or changing a command or flag
  without running `make docs-commands` and `make agent-skills` fails CI even
  though the code is correct.
- `docker-version-check` — `ARG GO_VERSION` in `Dockerfile` must exactly match
  the `go` line in `go.mod` (both `1.26.5` today). Bumping one alone fails.

Live Google API tests are opt-in and need a real account:
`GOG_IT_ACCOUNT=you@gmail.com go test -tags=integration ./internal/integration`.

## The destructive-command confirmation gate

Most destructive subcommands route through `confirmDestructive` /
`dryRunAndConfirmDestructive` (`internal/cmd/confirm.go`, ~52 call sites under
`internal/cmd/`). The non-obvious part, verified live against v0.34.1 on
2026-07-19:

**The gate fires whenever stdin is not a TTY — `--no-input` is not required.**
`confirmDestructiveChecked` refuses on `flags.NoInput || !stdinIsTerminal(ctx)`.
So any spawned or piped caller hits it even with no flags set:

```
$ gog calendar delete-calendar <id> < /dev/null
refusing to permanently delete secondary calendar <id> without --force (non-interactive)
$ echo $?
2
```

Two consequences worth remembering:

- **Exit 2, not a permission code.** The refusal is a `usagef`, so it maps to
  `usage` (2) in `gog schema --json | jq .automation.exit_codes`, alongside
  ordinary bad-flag errors. Callers that classify failures by exit code cannot
  distinguish "needs `--force`" from "typo in a flag" without reading stderr.
- **Probing with fake IDs is unreliable.** The gate normally fires before any
  network call, so `gog <cmd> fakeid` reveals it — but commands that resolve a
  name or ID through the API *first* return a `404 notFound` instead, masking
  whether a gate exists at all (e.g. `sheets delete-dimension`). The
  authoritative check is the `confirmDestructive` call sites in the source, not
  a live probe.

Not every destructive command is gated, and some gates are conditional on flag
values rather than on the subcommand. Read the call site.

This matters downstream: `~/git/gogcli-mcp` wraps this binary and always spawns
it non-interactively, so a tool that omits `--force` passes its mocked unit
tests and fails against the real API. That wrapper owns its own version floor
and PATH handling — see its `CLAUDE.md`, not this file.

<!-- pr-workflow:v3 -->
## Pull requests & release notes

Fleet policy — Conventional-Commit PR titles, labels, the auto-review /
auto-merge ladder, auto-review follow-up issues, PR timing, and release PRs —
lives in `~/.claude/CLAUDE.md`. Don't restate it here; the copies drifted.

Shared technical conventions (publishing, bundling, versioning guards,
write-verification, transport archetypes, testing traps) live in
[`chrischall/workflows`](https://github.com/chrischall/workflows):
`docs/fleet-conventions.md`, plus `README.md` for the CI pipeline contract.

**Repo-specific:** none of that automation exists here. This is a fork, so PRs
go **to `openclaw/gogcli`** (`gh pr create --repo openclaw/gogcli`) from a branch
pushed to `origin`, and are reviewed and merged by upstream's maintainer on
upstream's schedule — there is no auto-review verdict, no `ready-to-merge`
label, no auto-merge, and no release-please. Fifteen such PRs have merged
upstream to date. Versioning and releases are entirely upstream's; never tag or
bump a version in this fork. A handful of PRs have also been opened against
`chrischall/gogcli` itself; treat those as a private staging area only — landing
one changes nothing that ships, since `main` here is just a mirror. The real
target is upstream.
