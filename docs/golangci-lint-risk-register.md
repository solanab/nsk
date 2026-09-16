# golangci-lint risk register

This repository intentionally uses `linters.default: all` with a documented disable list. That policy exposes new
top-level linters when the pinned `golangci-lint` version changes, but it does not mean that every rule inside every
linter is enabled.

This document explains that distinction and defines how to handle overlapping, incompatible, or version-sensitive checks
without turning the disable list into an unexplained collection of exceptions.

## What `all` means

golangci-lint has two selection layers:

1. **Top-level linters.** `linters.default: all` selects every non-disabled linter known to the pinned golangci-lint
   binary.
2. **Checks inside a linter.** Aggregate linters such as `gocritic`, `govet`, `revive`, and `staticcheck` select their
   own analyzers or rules. Their defaults and their `settings` still apply after the top-level linter has been selected.

For example, this repository enables the top-level `gocritic` linter through `default: all`. In golangci-lint v2.12.2,
`gocritic` still uses its stable default checker set; v2.13.1 keeps that behavior and now resolves the internal checker
set from go-critic's own tag metadata instead of a local copy, so the stable/experimental/opinionated split remains
upstream-controlled. Its experimental and opinionated `unnamedResult` checker is not enabled. Enabling that checker
would require a `gocritic` setting such as an explicit checker, tag, or internal `enable-all` option.

Therefore these statements are both true:

- the repository uses an all-linters baseline with a blacklist;
- some checks inside enabled aggregate linters remain disabled by their upstream defaults.

Do not enable every internal checker merely to make the word `all` literal. Internal defaults often exclude
experimental, opinionated, or mutually incompatible policies. Treat an internal checker opt-in as a separate policy
change that requires review.

## Classify the behavior first

Problems described as "linter conflicts" usually belong to one of four categories:

| Category                    | Meaning                                                                  | Typical response                                                 |
| --------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------- |
| Logical conflict            | No source form can satisfy both rules                                    | Disable or narrow the less valuable rule                         |
| Invalid settings            | One linter rejects a combination of its own options                      | Correct the configuration                                        |
| Overlap                     | Several rules report the same concern without demanding opposite changes | Keep useful coverage; reduce noise only when it has a cost       |
| Fixer or formatter conflict | Rules may be satisfiable, but automatic edits overlap or do not converge | Apply fixes separately and keep one owner per formatting concern |

Threshold interactions are normally overlap, not contradiction. For example, `nlreturn` can add lines that make
`funlen.lines` fail, but the function can still be refactored to satisfy both rules.

## Resolution procedure

When a new rule or upgrade appears to create a conflict:

1. Pin the exact Go and golangci-lint versions used by CI. Do not diagnose formatter behavior with a different
   standalone tool version.
2. Reproduce without `--fix`. Record both the top-level linter name and, for an aggregate linter, its internal analyzer
   or checker.
3. Determine whether both diagnostics can be satisfied by a third source form. A confusing suggested fix is not proof of
   a logical conflict.
4. If both rules are sound, fix or refactor the code. Duplicate diagnostics alone are not a reason to weaken the
   baseline.
5. If one internal checker is unsuitable, disable or configure that checker before disabling the entire top-level
   linter.
6. Disable a top-level linter only when its whole policy is unsuitable or it cannot be narrowed. Add an inline comment
   to `.golangci.yml` explaining the concrete reason and any replacement.
7. For overlapping automatic edits, run one formatter or fixer at a time and inspect the diff. Use the repository's
   `just fmt` pipeline for normal formatting rather than composing standalone formatter versions.
8. Run `golangci-lint config verify`, the focused reproduction, and then `just check`. Run `just audit` separately when
   the change also affects dependencies or the toolchain.

Do not permanently blacklist a rule based only on an old issue title. Confirm that the report applies to the pinned
version and current settings.

## Known cases

The cases below were checked against golangci-lint v2.12.2, then re-verified on the v2.13.1 baseline (2026-08-25) with a
green `just check`. They are examples, not a complete compatibility matrix; golangci-lint does not publish one.

### Logical conflict behind an explicit opt-in

`nonamedreturns` rejects named results. The optional `gocritic.unnamedResult` checker asks some functions with repeated
result types to name those results. For a matching signature such as `func f() (float64, float64)`, no form satisfies
both checks.

This repository is not affected because it does not opt in to `gocritic.unnamedResult`. If a future change enables that
checker, keep `nonamedreturns` and reject the opt-in unless the repository deliberately changes its named-result policy.

Sources:

- [`nonamedreturns` analyzer](https://github.com/firefart/nonamedreturns/blob/v1.0.6/analyzer/analyzer.go#L16-L19)
- [`gocritic.unnamedResult` checker](https://github.com/go-critic/go-critic/blob/v0.14.3/checkers/unnamedResult_checker.go#L13-L27)
- [gocritic default selection in golangci-lint](https://github.com/golangci/golangci-lint/blob/v2.12.2/pkg/golinters/gocritic/gocritic_settings.go#L425-L430)

### Invalid settings detected by the tool

Examples include:

- `sloglint.kv-only: true` together with `attr-only: true`;
- `ireturn.allow` together with `ireturn.reject`;
- incompatible `gocritic` internal selection modes;
- placing the same `godoclint` rule in both its enable and disable lists.

These are configuration errors, not reasons to disable the top-level linter. `golangci-lint config verify` or linter
initialization should reject them.

### Historical fixer and formatter reports

Reports involving old `wsl` with `gofumpt`, or old `gci` with `goimports`, show why automatic fixes need care. They do
not establish a current logical conflict:

- the old `wsl` example had a third layout that satisfied both tools, and `wsl_v5` has since replaced it;
- golangci-lint v2 runs formatters through a fixed pipeline;
- historical import-fixer bugs cannot be assumed to reproduce in v2.12.2.

See [issue #1510](https://github.com/golangci/golangci-lint/issues/1510) and
[issue #1490](https://github.com/golangci/golangci-lint/issues/1490).

## Project interaction register

Maintain this table as a project-local risk register. It is intentionally smaller than a pairwise linter matrix: add an
entry only when an interaction is relevant to the current configuration, has been reproduced, or has credible
version-specific evidence.

Status values:

- **Active:** both sides are enabled and the interaction is relevant, but the current policy is satisfiable.
- **Dormant:** the interaction requires an option that is not currently enabled.
- **Historical:** an older version demonstrated the risk, but it has not been reproduced with the pinned version.

| Interaction                                           | Class                           | Trigger                                            | Status     | Current decision                                                                      | Revisit when                                                                 |
| ----------------------------------------------------- | ------------------------------- | -------------------------------------------------- | ---------- | ------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `nonamedreturns` / `gocritic.unnamedResult`           | Logical conflict                | Opting in to `unnamedResult`                       | Dormant    | Keep `nonamedreturns`; do not opt in to the checker                                   | Named-result policy changes or gocritic changes its rule                     |
| `nlreturn` / `wsl_v5`                                 | Whitespace overlap              | Return or branch layout                            | Active     | Find a layout that satisfies both; refactor before disabling either                   | A pinned-version example cannot converge                                     |
| `nlreturn` or `wsl_v5` / `funlen.lines`               | Threshold interaction           | Required blank lines push a function over 60 lines | Active     | Refactor the function; change the threshold only as a repository policy               | Repeated cases show the threshold measures formatting rather than complexity |
| `gofumpt` / `goimports`                               | Formatter overlap               | Import or source formatting                        | Active     | Use the pinned `just fmt` pipeline and require idempotent output                      | The pipeline changes a clean file on a second run                            |
| `cyclop`, `gocyclo`, `gocognit`, `maintidx`, `funlen` | Metric overlap                  | A large or complex function                        | Active     | Treat reports as multiple signals and refactor; do not suppress duplicates by default | A metric repeatedly adds no distinct information                             |
| `paralleltest` / `tparallel`                          | Complementary overlap           | Parallel test declarations                         | Active     | Satisfy both missing-use and correctness checks                                       | A test cannot safely run in parallel under the pinned rules                  |
| `exhaustive` / `gochecksumtype`                       | Coverage overlap                | Enum switches or sum-type handling                 | Active     | Keep both scopes unless diagnostics are exact duplicates                              | A repeated duplicate has no additional coverage value                        |
| `wrapcheck` / `errorlint`                             | Configuration-sensitive overlap | Custom `wrapcheck` ignore signatures               | Active     | Preserve applicable default signatures and use `%w` wrapping                          | Either linter's settings are customized                                      |
| old `wsl` / `gofumpt`; old import fixers              | Fixer or formatter history      | Tool downgrade or formatter pipeline change        | Historical | Do not infer a current conflict; reproduce on the pinned binary                       | Formatter versions or execution order change                                 |
| `exhaustruct_v5` / test struct literals               | Coverage expansion              | exhaustruct v5 flags literals that omit any field  | Active     | List every field in test fixtures; refactor before considering exclude patterns       | Exclude regexes accumulate beyond a documented handful                       |

v2.13.1 note: exhaustruct moved to v5 semantics and now reports unexported fixture structs in `internal/` tests; four
literals in `internal/app/app_test.go` were completed instead of excluded. Under this configuration v2.13.1 enables 109
top-level linters.

The matrix records decisions, not automatic disables. An active row means the combination is deliberately retained and
has a known response. Move a row to a logical conflict only after demonstrating that no source form satisfies both
rules.

Update the matrix when:

- the pinned Go or golangci-lint version changes;
- `.golangci.yml` enables, disables, or reconfigures a linter or internal checker;
- a formatter or fixer is no longer idempotent;
- a minimal reproduction proves a new logical conflict;
- repeated overlap produces measurable maintenance cost.

## Why there is no exhaustive list

The golangci-lint maintainers explicitly state that the project does not track linter or rule overlap and leaves
combination selection to users. Version v2.12.2 exposes 114 top-level linters, which already produces 6,441 unordered
pairs before considering internal rules, settings, plugins, Go versions, source shapes, and generated edits.

Fixer conflicts are especially dynamic: golangci-lint detects overlapping edits produced for the current run rather than
consulting static conflict metadata.

Sources:

- [official discussion on duplication](https://github.com/golangci/golangci-lint/discussions/4661#discussioncomment-9158902)
- [v2.12.2 linter metadata](https://github.com/golangci/golangci-lint/blob/v2.12.2/pkg/lint/linter/config.go#L32-L48)
- [v2.12.2 fixer conflict handling](https://github.com/golangci/golangci-lint/blob/v2.12.2/pkg/result/processors/fixer.go#L124-L169)

## Upgrade checklist

For every golangci-lint upgrade:

1. Review added and removed top-level linters.
2. Review changed defaults in aggregate linters and formatters.
3. Run `just check` without automatic fixes first.
4. Classify new diagnostics and update the project interaction register.
5. Prefer code fixes, then narrow internal configuration, and use a documented top-level disable only as the last policy
   option.
6. Update this document when a conflict is reproducible and version-specific.

The current policy remains `default: all` plus the short, documented disable list in `.golangci.yml`. The purpose of
that baseline is early discovery of new checks, not blind activation of every experimental sub-rule.
