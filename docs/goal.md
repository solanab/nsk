# nsk Implementation Goal for Issues #1-#12

This goal is written to be used by subagents (via `spawn_subagent` or `execute-plan` workflow) to implement the 12 GitHub issues in vertical slices, with TDD at seams, and git commit after each issue completion.

## Workflow Rules
- Use `tdd` skill for TDD at declared seams.
- After each issue is done (tests green, `just check` green, coverage 100%), commit with:
  ```
  git add -A
  git commit -m "feat(#<issue>): <short description from issue title or body>"
  git push
  ```
  and the commit message should reference Closes #<issue>.
- Use `code-review` skill after implementation.
- Do not add MCP, TUI, REPL.
- Forum seam: TDD at `internal/client` and `internal/remote` for all Forum methods.
- Config seam for #1: TDD at `internal/config/*` (paths, bind, config).

## Issues List (with blocked_by and seams)

**#1 Skeleton, XDG config, Kong help**  
Blocked by: none  
TDD seams: `internal/config/paths.go`, `internal/config/bind.go`, `internal/config/config.go`, `internal/app/app.go` (error paths)  
Acceptance: `just check` green, `nsk help` / `--version` / empty argv work, XDG read, mutex, bind rules.

**#2 whoami + cookie + Chrome 124**  
Blocked by: #1  
TDD seams: `internal/client/client.go`, `internal/config/config.go` (CookieFile), error paths for warmup.  
Acceptance: cookie import, warmup, whoami JSON, Chrome 124 spike.

**#3 structure / cats**  
Blocked by: #2  
TDD seams: `internal/app/app.go` (structure), `internal/client` (Categories).  
Acceptance: `nsk` / `nsk structure` / `nsk cats` JSON.

**#4 list**  
Blocked by: #2  
TDD seams: `internal/client` (LatestPosts, CategoryPosts), HTML parse.  
Acceptance: `nsk list`, `nsk list <slug>`.

**#5 post**  
Blocked by: #2  
TDD seams: `internal/client` (GetPost, GetPostAll, FormatPost), Markdown.  
Acceptance: `nsk post <id> [--page N|--all]`.

**#6 search**  
Blocked by: #2  
TDD seams: `internal/client` (Search), HTML parse.  
Acceptance: `nsk search <q>`.

**#7 user**  
Blocked by: #2  
TDD seams: `internal/client` (GetUser, WhoAmI), JSON.  
Acceptance: `nsk user <id>`.

**#8 notify**  
Blocked by: #2  
TDD seams: `internal/client` (Notifications), JSON.  
Acceptance: `nsk notify`.

**#9 pin Reply POST**  
Blocked by: #2  
TDD seams: `docs/forum-write.md` (gate), `internal/client` if pinned.  
Acceptance: document pinned or reject Reply.

**#10 reply (or close if unpinned)**  
Blocked by: #9  
TDD seams: `internal/client` (Reply) if pinned, or close issue.  
Acceptance: `nsk reply <id> --body FILE`.

**#11 nsk server + remote Forum**  
Blocked by: #3-#8, #10  
TDD seams: `internal/remote`, `internal/server`, mutex.  
Acceptance: `nsk server`, remote Client.

**#12 release**  
Blocked by: #11  
TDD seams: `go.mod`, `.goreleaser.yaml`, docs/install.  
Acceptance: tag v*, GoReleaser binaries, README install.

## Subagent Usage
Use `spawn_subagent` for each issue with prompt containing the issue body, current code, and "Implement per this goal, TDD at seams, commit after done".

Start with subagent for #1, then sequential.

