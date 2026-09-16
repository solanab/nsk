# Engineering docs

This directory records the template's engineering-practice layer. Root `AGENTS.md` owns maintainer gates and
collaboration rules; this directory owns the **current code contract**, **not-yet-enabled practices**, and
version-sensitive special records. Runtime technology choices are not quality contracts; they live in
[`../decisions/`](../decisions/).

## Three-layer model

Split by "is it immutable?" and "must it run now?":

| Layer | Name      | Meaning                                                                           | Landing place                                                          |
| ----- | --------- | --------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| 0     | Invariant | Template-identity principles; do not change them for project convenience          | `AGENTS.md`, `just check`, CI                                          |
| 1     | Stable    | Mandatory contract every downstream project should inherit today                  | [`contracts.md`](./contracts.md), `go.mod`, `justfile`, quality config |
| 2     | Evolving  | Practices not yet general; wait for architecture or product needs to trigger them | [`backlog.md`](./backlog.md)                                           |

Entries may attach `MUST`, `SHOULD`, or `MAY`, and must note enforcement (lint, tests, build, or human review) and noise
risk.

Classification order:

1. If you remove it, is this still a project in this template family? If not, it is Invariant.
2. Does it hold for a pure library, a CLI, and a service? If yes, lean Stable; otherwise it is Evolving.
3. Is there no real need yet? Write it only in the backlog; do not add a gate early for symmetry.

## Three core Markdown files

| File                             | Responsibility                                                       |
| -------------------------------- | -------------------------------------------------------------------- |
| This `README.md`                 | Meta-rules, layer definitions, and evolution flow                    |
| [`contracts.md`](./contracts.md) | Enabled human-readable contract; must match executable configuration |
| [`backlog.md`](./backlog.md)     | Not-yet-enabled candidate practices; they are not a current gate     |

[`golangci-lint-risk-register.md`](./golangci-lint-risk-register.md) is a special-purpose risk register, not a fourth
contract layer. It records only evidenced rule interactions under the current version and configuration; do not expand
it into a global inventory that claims to enumerate every linter combination.

## Evolution flow

```text
Evolving candidate
  -> docs/backlog.md (policy + trigger + enforcement + noise)
  -> on trigger, write into contracts.md (Stable or enabled Evolving)
  -> sync .golangci.yml / justfile / .testcoverage.yml / CI (as needed)
  -> delete the matching backlog entry (Single Path)
  -> just check
```

Principles:

- Write the policy first, then flip the switch; do not enable high-noise rules without that policy.
- Changing Stable must sync docs, config, and gates, and counts as a template behavior change.
- Changing Invariant is refused by default unless the user explicitly asks and the root contract, README, and CI
  narrative stay in sync.
- Docs are not a second `just` or CI entrypoint; command truth remains in the `justfile` and workflows.

## Quality gates

Markdown files are checked by dprint and Rumdl through Prek. After changing engineering docs, run `just check`; for a
fast docs-only check, run `just markdown-format` and `just markdown-lint`.
