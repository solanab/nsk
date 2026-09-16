# Enabled code contract

> **Layer 1 Stable** (currently MUST) plus a few layer-2 constraints that may be promoted later. Changing any section:
> sync executable config and the necessary backlog -> `just check`. This document must not form a second rule set beside
> `.golangci.yml`, `justfile`, or CI.

**Layer 0 Invariant** is not restated at length here; see the maintainer contract in root `AGENTS.md`.

## Stable (layer 1: currently MUST)

### S1. Module and Go version

| Policy                                                                                                                                    | Enforcement                                 |
| ----------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| The `go` directive in `go.mod` is the minimum supported language and standard-library version; do not raise it casually to use new syntax | `go.mod`, code review, CI Go version matrix |
| Runtime dependencies are added only for product behavior; tool dependencies must be used by checked-in config or scripts                  | `go mod tidy -diff`, `just check`           |
| Downloaded tools pin versions and verify upstream checksums                                                                               | `scripts/install-tools.sh`                  |

### S2. Command entrypoint and code boundary

- Human-facing tasks have only `just` entrypoints; complex shell orchestration lives in `scripts/`.
- `cmd/<binary>/main.go` is only the process entrypoint; application behavior lives in `internal/`.
- Only an intentionally exposed public Go import API may leave `internal/`.
- Do not keep a parallel Makefile, Taskfile, or legacy compatibility entrypoint.

Enforcement: directory convention, `justfile-check`, `just check`, and human review.

### S3. Go quality baseline

The policy is `golangci-lint` top-level `all` plus the commented explicit disables in `.golangci.yml`:

- Prefer fixing findings in code; do not silently miss a rule because a preset omitted it.
- `depguard`, `exhaustruct`, `exhaustruct_v5`, `noinlineerr`, and deprecated old linters may be disabled only with a
  recorded reason. `exhaustruct_v5` is the successor of `exhaustruct`; both stay off for the same intentional-zero
  policy.
- `gofumpt` and `goimports` own Go formatting; use the pinned versions through `just fmt`.
- Checkers inside aggregate linters still follow their own defaults; `default: all` is not an internal `enable-all`.
- Rule overlap, threshold interactions, and fixer risk are recorded in
  [`golangci-lint-risk-register.md`](./golangci-lint-risk-register.md).

### Varnamelen

Keep `varnamelen` enabled. Short variable names should grow with scope and use distance; that principle comes from
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments#variable-names) and
[Google Go Style Decisions](https://google.github.io/styleguide/go/decisions.html#variable-names). Comments in the
config point to this section so candidate exceptions are not hidden in YAML.

| Config or candidate                    | Template status                | Decision boundary                                                                                                             |
| -------------------------------------- | ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| `max-distance: 5`                      | Enabled                        | Short names apply only to a visible local use window; do not default-relax medium scopes of 8--15 lines.                      |
| `ignore-type-assert-ok`                | Enabled                        | Exempt only the Go comma-ok form in type assertions.                                                                          |
| `ignore-map-index-ok`                  | Enabled                        | Exempt only the Go comma-ok form in map indexes.                                                                              |
| `ignore-chan-recv-ok`                  | Enabled                        | Exempt only the Go comma-ok form in channel receives.                                                                         |
| `w io.Writer`, `r io.Reader`           | Downstream adopt with evidence | Familiar Go abbreviations, but whether they may exceed this baseline's distance limit depends on the actual code.             |
| `tx *sql.Tx`                           | Downstream adopt with evidence | A database transaction is not a domain concept every project has.                                                             |
| `p []byte`                             | Not recommended                | The type does not say whether `p` is payload, path, password, or something else.                                              |
| `_test.go` exclusion or `ignore-names` | Not enabled                    | Do not hide naming problems behind whole-file or whole-name exemptions; use a narrow, reasoned project exception when needed. |

Enforcement: `.golangci.yml`, `just lint`, `just fmt-check`, `just check`.

### S4. Tests, coverage, and static checks

| Policy                                                                                               | Enforcement                                |
| ---------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| Application behavior must have tests; coverage target is 100% of file, package, and total statements | `scripts/coverage.sh`, `.testcoverage.yml` |
| Every Go package passes ordinary tests and `go vet`                                                  | `just test`, `just vet`                    |
| Race checks are a CI and pre-release gate; they are not folded into every `just check` by default    | `just race`, CI                            |
| Online vulnerability-database scanning is independent of the routine gate                            | `just audit`, `just check-online`          |

Thin `cmd/` entrypoints are not an application coverage target; testable behavior must live in a non-`main` package.

### S5. Format, docs, and config

- Go, shell, TOML, YAML, and Markdown each have one primary formatting chain.
- Tombi owns TOML format and lint; `yamlfmt` checks YAML; Markdown is checked by dprint/Rumdl through Prek. dprint
  Markdown uses `textWrap: always` at 120 columns.
- In `just check`, Tombi uses normal online schema resolution and promotes every warning to an error;
  vulnerability-database scanning still belongs to `just audit`.
- source-lines is the only repository source-size gate: Go files default to at most 300 effective code lines and 1000
  total lines; total lines include effective code, comment-only lines, and blank lines. 300/1000 is this template's
  governance guardrail, not a universally proven defect threshold from a paper; crossing it first triggers a split or
  refactor review.
- A file may rise to 500 effective code lines and 1600 total lines only with an exact-path reason in source-lines.toml;
  do not pre-relax tests or examples. The reason should record the observable cost of splitting and the measurement
  basis, such as co-change history, dependency or fixture hand-off, duplicated code, or a complexity/churn baseline and
  result. "The test is too long", a directory category, or a subjective "keep it cohesive" is not itself a reason;
  repeated exception requests should drive a module-boundary refactor.
- When adding or renaming a recipe, script, or quality tool, sync `scripts/README.md`, the README, and any necessary
  help text.

### S6. Release and platform boundary

- Local binaries use `just build`; release snapshots and formal releases use GoReleaser.
- The release matrix covers Linux, macOS, and Windows; do not add platform-specific runtime implementations until the
  product needs them.
- Release commands, version injection, and artifact paths must stay reproducible; cleanup uses `trash`.

### S7. Downstream copy contract

Template-maintainer rules stay in root `AGENTS.md`; downstream projects copy `AGENTS.md.template` and then add
project-specific rules. A downstream contract must not reference maintainer-local paths, sibling repositories, or
implied tools outside this template.

## Evolving (layer 2: not promoted by default)

There are currently no Evolving entries that have been promoted and are in force. Candidate practices live in
[`backlog.md`](./backlog.md); after enabling one, move the policy into this section, sync config, and delete the
original backlog entry.

## Change checklist

- [ ] Matching Stable or enabled Evolving section in this file is updated
- [ ] `.golangci.yml`, `justfile`, CI, or other related config is synced
- [ ] Promoted backlog entries are deleted
- [ ] Project taste is not miswritten as Invariant
- [ ] `just check` passes; for online changes also run `just audit` or `just check-online`
