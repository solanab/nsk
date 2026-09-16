# Scripts

The root `justfile` is the only task entrypoint. Non-trivial shell behavior lives here so recipes remain easy to audit.

## Tool installation

`install-tools.sh` installs the external tools that should not be built through `go install`:

| Tool            | Version | Reason                                                     |
| --------------- | ------- | ---------------------------------------------------------- |
| `golangci-lint` | 2.13.1  | Official builds control the Go version and enabled linters |
| ShellCheck      | 0.11.0  | Native release binary, independent of host packages        |
| GoReleaser      | 2.18.0  | Reproducible local release snapshots                       |
| Tombi           | 1.4.1   | TOML formatting, linting, and schema-aware diagnostics     |
| `source-lines`  | 0.2.0   | Effective and total Go line policy                         |

The script supports Linux and macOS on amd64 and arm64. Every archive uses an upstream SHA-256 checksum and installs
into ignored `runtime/tools/bin/` paths. `just install` is idempotent. Lightweight Go tools (`actionlint`,
`go-test-coverage`, `govulncheck`, `shfmt`, and `yamlfmt`) are pinned in `go.mod` and invoked with `go tool`.

Prek is intentionally installed outside this script because it is an optional Git-hook runner rather than a Go build
dependency. It provides isolated dprint and Rumdl environments for Markdown and reuses the repository's `just` recipes
for language-specific checks.

`install-source-lines.sh` fetches the pinned private GitHub Release through an authenticated `gh` session or CI-provided
`GH_TOKEN`. The upstream installer script checksum is pinned in the consumer script; the upstream installer then
verifies the archive checksum and binary version before writing `runtime/tools/bin/source-lines`.

## `coverage.sh`

`coverage.sh` discovers every non-main Go package, runs its tests with atomic whole-module instrumentation, and applies
the 100% file, package, and total thresholds from `.testcoverage.yml`. Thin `cmd/` process entrypoints are omitted;
application behavior belongs in covered non-main packages.

## `background.sh`

`background.sh` manages named, trusted local development commands. Each task has one PID file, one command file, one
process-group mode file, and one log file under ignored `runtime/` paths. Names are restricted to letters, digits, dots,
underscores, and hyphens. On systems with `setsid`, the task gets its own process group so `just` cannot clean it up
when the recipe exits; stop operations target only that recorded PID/group and never search by port or process name.

Do not use this helper for production services. Use a dedicated system or user service with explicit ownership, restart,
logging, and permission policy.

## Conventions

1. Keep complex orchestration in `scripts/`; keep `justfile` recipes declarative.
2. Keep `just check` aligned with CI; use `just audit` alone or `just check-online` for the current vulnerability
   database.
3. Pin downloaded tools and verify their upstream checksums.
4. Update `help.txt` and this document when adding or renaming a recipe or script.
5. Use `trash` for user-requested cleanup recipes; temporary installer directories may be removed directly by their
   owning script.
