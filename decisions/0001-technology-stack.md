# nsk technology stack decision

Consulted the `modern-go-template` technology radar (`guides/technology-radar.md`, last verified 2026-08-22) at decision
time. Radar candidates are not project dependencies. Do not copy that radar into this repository.

| Field           | Value                                                                                                                                                   |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Status          | accepted                                                                                                                                                |
| Date            | 2026-09-15                                                                                                                                              |
| Decision makers | repo owner; recorded from the stack-decision turn                                                                                                       |
| Project shape   | CLI (local `net/http` Server arrives in PR 5; not an HTTP-API product)                                                                                  |
| Supersedes      | none                                                                                                                                                    |
| Review trigger  | Chrome 124 cannot pass nodeseek.com Cloudflare; Kong cannot express no-args `structure`; a new product job needs persistence, RPC, or a second language |

Do not rewrite this file in place. A stack change adds a new decision and records supersedes / superseded-by on both
documents.

## Decision

`nsk` is a **Go 1.26 CLI**. Quality tooling is inherited wholesale from `modern-go-template`. Runtime packages are added
only for jobs the product already has: Kong for the subcommand tree, pelletier TOML plus explicit `NSK_*` for XDG
config, bogdanfinn Chrome 124 TLS for NodeSeek behind Cloudflare, and goquery for SSR HTML. stdout is slim JSON for
agents; stderr is `log/slog`. There is no TUI, MCP, browser, database, or second language.

## Context

### Core jobs

1. Give an agent one-shot `nsk` subcommands that dump NodeSeek structure, lists, posts, users, and notifications as slim
   JSON.
2. Keep a cookie-only Account (`pjwt`) on one machine, with an optional later Server so other machines do not copy the
   jar.
3. Parse Vue SSR HTML plus a few JSON endpoints without depending on any open-source forum product.

### Constraints

| Constraint                   | Current fact                                                                                                                                  |
| ---------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Go/runtime                   | Minimum Go **1.26.7**, matching `modern-go-template` `go.mod`. `CGO_ENABLED=0`. Do not raise the `go` directive without explicit approval.    |
| Deployment environment       | User-installed single binary. XDG config/state. Optional loopback Server.                                                                     |
| Data ownership and residency | Cookie jar stays on the machine that holds Account. No third-party telemetry.                                                                 |
| Scale and latency            | One agent process, one request at a time. Forum GET timeout 30s; remote Forum 60s.                                                            |
| Availability and durability  | Best-effort CLI. No retry on 429. Warmup failure must not overwrite `CookieFile()`.                                                           |
| Security and compliance      | Cookie-only v1. Strip `cf_*` on import. Reject bind `0.0.0.0` and `::`. Binary will include tls-client **BSD-4-Clause**. Product license MIT. |
| Team and operations          | Same Go + `just` workflow as the template. No on-call.                                                                                        |
| Compatibility boundary       | Closed-source NodeSeek Vue SSR + Cloudflare. Ops shape copied from `ldo`; quality contract is the template, not `ldo`.                        |

### Non-goals

- MCP, Bubble Tea TUI, 摸鱼 REPL, Playwright / chromedp, password + Turnstile.
- Cobra, chi, koanf, ORM, sqlc, Redis, ConnectRPC, OpenTelemetry.
- A second HTTP stack or a browser fallback when Cloudflare returns 403.
- Shipping `ldo`'s Go 1.23, hand-rolled `flag` soup, or Bubble Tea as the nsk baseline.
- Installing a package because it might be useful later.

## Capability decisions

| Capability                          | Status   | Current choice                                                                     | Why chosen                                                                                          | Alternatives and why rejected                                                    | Re-evaluation trigger                                      |
| ----------------------------------- | -------- | ---------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| Template development baseline       | inherit  | `modern-go-template` quality contract                                              | One `just` surface, golangci-lint `all`, 300/1000 source-lines, 100% app coverage, separate `audit` | `ldo`'s thinner `just check` — older Go, weaker lint, not the workspace template | Template baseline changes and nsk deliberately lags        |
| Package interface and distribution  | n/a      | Single `nsk` binary via GoReleaser                                                 | Agent execs a command, does not import a Go API                                                     | Publishing a library — no public import API                                      | Someone needs `nsk` as an importable module                |
| Formal CLI                          | select   | `github.com/alecthomas/kong`                                                       | Radar Adopt: typed subcommands, help, exit codes. Surface is `gh`-style (`list`, `post`), not flags | `flag` recreates `ldo`'s parser. Cobra is only justified by its plugin ecosystem | Kong cannot express no-args = `structure` without a hack   |
| Forum HTTP client                   | select   | `bogdanfinn/tls-client` + `bogdanfinn/fhttp`, `profiles.Chrome_124`                | Cloudflare is a hard constraint; stdlib `net/http` will not pass it. Same fingerprint as `ldo`      | `net/http`; Playwright; reusing a browser `cf_clearance`                         | Chrome 124 gets a CF challenge that a newer profile passes |
| Local Server HTTP                   | deferred | stdlib `net/http` in PR 5                                                          | A handful of `/api/v1` routes                                                                       | chi or any web framework — routing tree is small                                 | Route tree becomes large enough that chi is clearer        |
| HTML parsing                        | select   | `github.com/PuerkitoBio/goquery` on `golang.org/x/net/html`                        | SSR lists and posts are HTML. Selectors stay in `parse_*.go`                                        | Regex over markup; a headless DOM                                                | NodeSeek ships a stable JSON list/post API                 |
| RPC transport                       | n/a      | —                                                                                  | No protobuf API                                                                                     | gRPC / ConnectRPC                                                                | Server protocol changes to protobuf                        |
| Realtime transport                  | n/a      | —                                                                                  | No websocket job                                                                                    | —                                                                                | —                                                          |
| Config and external data validation | select   | XDG TOML via `pelletier/go-toml/v2`; explicit `NSK_*` overlay in `internal/config` | Two sources only: file + a handful of env vars + `--config`                                         | koanf (layered framework); `caarlos0/env` (too few variables)                    | Config gains many sources or a plugin layer                |
| Database and transactions           | n/a      | —                                                                                  | No persistence beyond the cookie file                                                               | —                                                                                | —                                                          |
| Schema migration                    | n/a      | —                                                                                  | —                                                                                                   | —                                                                                | —                                                          |
| Cache or broker                     | n/a      | —                                                                                  | —                                                                                                   | —                                                                                | —                                                          |
| Background tasks                    | n/a      | —                                                                                  | One-shot process                                                                                    | —                                                                                | —                                                          |
| Serialization and API contract      | select   | `encoding/json`                                                                    | stdout protocol is slim JSON                                                                        | protobuf / MessagePack                                                           | A non-JSON agent protocol is required                      |
| Logs, traces, and metrics           | select   | `log/slog` on stderr                                                               | stdout is agent JSON. Never log cookie / token / pjwt                                               | zerolog / zap; OTel (single-process CLI)                                         | A metrics backend is actually operated                     |
| Extra testing capability            | select   | `testing` + `github.com/google/go-cmp/cmp`                                         | Radar Adopt for structural diffs on views and parsed posts                                          | testify                                                                          | —                                                          |
| Performance benchmark               | deferred | `testing.B` only if a hot path appears                                             | No measured bottleneck                                                                              | A PR-blocking bench without a noise baseline                                     | Parse or `--all` is shown to be slow                       |
| Build and deploy                    | select   | `just build`; GoReleaser linux/darwin/windows × amd64/arm64, `CGO_ENABLED=0`       | Template default. Windows is a release artifact only                                                | A second Makefile / Taskfile                                                     | A platform-specific code path is required                  |

## Dependencies added now

These packages may enter `go.mod` in **PR 1**, which is the first call site. They must not be added before that PR.
Versions are pinned by `go.mod` / `go.sum` at implementation time, not in this document.

| Package                           | runtime/dev | Capability served | First call site                           | Version policy            | Owner             |
| --------------------------------- | ----------- | ----------------- | ----------------------------------------- | ------------------------- | ----------------- |
| `github.com/alecthomas/kong`      | runtime     | Formal CLI        | `cmd/nsk` help / `--version` / empty argv | module minor at first use | `cmd/nsk`         |
| `github.com/pelletier/go-toml/v2` | runtime     | Config            | `internal/config` TOML decode             | module minor at first use | `internal/config` |
| `github.com/google/go-cmp`        | dev         | Test assertions   | `internal/config` tests                   | module minor at first use | tests             |

Verified at decision time (2026-09-15):

| Package    | License      | Maintenance signal                  |
| ---------- | ------------ | ----------------------------------- |
| Kong       | MIT          | tag `v1.16.1`; last push 2026-09-15 |
| go-toml v2 | MIT          | release `v2.4.3` (2026-07-05)       |
| go-cmp     | BSD-3-Clause | release `v0.7.0`                    |

## Explicitly deferred

| Capability                        | Why not adopted now                                        | Re-evaluation trigger             |
| --------------------------------- | ---------------------------------------------------------- | --------------------------------- |
| `bogdanfinn/tls-client` + `fhttp` | No network in PR 1. First call site is PR 2 warmup `GET /` | PR 2 starts                       |
| `PuerkitoBio/goquery`             | First call site is PR 2 `parse_*.go`                       | PR 2 starts                       |
| stdlib `net/http` Server          | Multi-machine Server is PR 5                               | PR 5 starts                       |
| OpenTelemetry                     | Single-process CLI                                         | An operated backend exists        |
| koanf / `caarlos0/env`            | Two config sources, few keys                               | Config layering grows             |
| chi                               | Few routes                                                 | Server routing tree becomes large |
| Database, queue, RPC, websocket   | No product job                                             | A job appears                     |

tls-client extra notes for PR 2: upstream `v1.16.0` (2026-09-02), module `go 1.24.1`, license **BSD-4-Clause**
(advertising clause). goquery `v1.13.0` (2026-08-27), BSD-3-Clause.

## Boundaries

```text
cmd/nsk (Kong)
  -> internal/config (TOML + NSK_* + bind rules)
  -> client.Forum
       -> internal/client  (tls-client, PR 2)
       -> internal/remote  (stdlib HTTP to Server, PR 5)
  -> internal/server (stdlib net/http, PR 5)
```

- Configuration is loaded and validated only in `internal/config` (`[server]` / `[client]` mutex, default
  `127.0.0.1:9200`, reject `0.0.0.0` and `::`).
- Timeouts and the cookie jar belong to the TLS client. Forum domain code does not import `fhttp`.
- CSS selectors exist only in `parse_*.go`.
- External failures become `ErrCloudflare` / `ErrExpiredCookie` / `ErrForbidden` / `ErrRateLimited` / `ErrNotFound` at
  the Forum boundary.
- Kong stays in `cmd/nsk`. `internal/` does not import Kong.
- Domain packages may use the standard library and go-cmp in tests. They do not import TLS, goquery, or Kong.

## Spike evidence

| Assumption                                      | Minimal experiment                                      | Result or artifact                  | Conclusion                                                                |
| ----------------------------------------------- | ------------------------------------------------------- | ----------------------------------- | ------------------------------------------------------------------------- |
| Chrome 124 reaches nodeseek.com past Cloudflare | PR 2: `GET /` with a real `pjwt`, inspect status / body | Not run (needs the operator cookie) | Does not block this decision. Failure → change profile or tls-client only |
| Kong empty argv = `structure`; `--text` global  | PR 1: help, no args, unknown subcommand, exit codes     | Not run (no `go.mod` yet)           | First CLI PR must prove this before more subcommands                      |

## Consequences

### Benefits

- Agents get a typed subcommand tree instead of `ldo`-style flag soup.
- Quality gates match the rest of the Go workspace, so `just check` means the same thing as in the template.
- Cloudflare handling stays a single, already-proven TLS path.
- `go.mod` stays empty of unused frameworks.

### Costs and risks

- tls-client is outside the radar default HTTP client and brings BSD-4-Clause plus a large fingerprint dependency tree.
- NodeSeek HTML can change; the mitigation is full-page fixtures, not a second parser.
- Template 100% coverage and golangci-lint `all` are stricter than `ldo`; PR 1 must copy the template files rather than
  invent a lighter gate.

### Migration and rollback

- No production users yet. Rollback is deleting the binary.
- Changing TLS, CLI, or config later is a new numbered decision.
- PR 1 copies template identity fields (`go.mod` module path `github.com/solanab/nsk`, binary `nsk`) via the template's
  `go-baseline-adoption` skill. It does not `require` the template module.

## Acceptance checklist

- [x] Every immediately added dependency has a named first call site and owner (call sites land in PR 1; none are in the
      tree today)
- [x] At least one seriously considered alternative and rejection reason is recorded
- [x] Go support, license, release status, and maintenance signal were checked for Kong, go-toml, go-cmp, tls-client,
      and goquery
- [ ] Hardest runtime assumption (Chrome 124 vs nodeseek.com) — scheduled for PR 2, not faked here
- [ ] `go.mod` / `go.sum` — PR 1
- [x] Deferred capabilities are not written as v1 commitments
- [ ] `just check` — PR 1
- [ ] `just audit` — when PR 1/PR 2 add modules
