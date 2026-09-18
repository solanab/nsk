# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Before exploring, read these

- **`CONTEXT.md`** at the repo root — Account / Server / Client / Forum / Post / Floor, XDG paths, `NSK_*`, the agent
  command surface.
- **`decisions/`** — accepted stack and architecture decisions (currently `0001-technology-stack.md`). Treat these as
  ADRs; do not rewrite an accepted decision in place.
- **`docs/design.md`** — product spec (Forum methods, cookie import, command surface, PR plan). Load before changing
  those.
- **`docs/forum-write.md`** — the Reply POST gate. Load before implementing `Reply`.

If a listed file does not exist, **proceed silently**. Don't flag its absence; don't suggest creating it upfront. The
`domain-modeling` skill creates glossary and ADR files lazily when terms or decisions actually get resolved.

## File structure

Single-context repo:

```text
/
├── CONTEXT.md
├── AGENTS.md
├── decisions/
│   └── 0001-technology-stack.md
└── docs/
    ├── design.md
    ├── forum-write.md
    └── agents/
```

This repo does not use `CONTEXT-MAP.md` or per-package `CONTEXT.md`. Stack decisions live in `decisions/`, not
`docs/adr/`.

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the
term as defined in `CONTEXT.md`. Don't drift to synonyms the glossary explicitly avoids (`auth`, `session`, `gateway`,
`Topic`, `PostStream`).

If the concept you need isn't in the glossary yet, that's a signal — either you're inventing language the project
doesn't use (reconsider) or there's a real gap (note it for `domain-modeling`).

## Flag ADR conflicts

If your output contradicts an existing decision, surface it explicitly rather than silently overriding:

> *Contradicts decisions/0001-technology-stack.md (Kong + Chrome 124) — but worth reopening because…*
