# Practice backlog (layer 2 Evolving candidates)

> These are not a gate before they are enabled, and they are not Invariant. Flow: complete the policy -> write into
> [`contracts.md`](./contracts.md) -> change config -> delete this entry -> `just check`. Add or remove items with
> product architecture; do not open everything early just to be "stricter".

## How to fill this in

| Field            | Meaning                                |
| ---------------- | -------------------------------------- |
| Policy draft     | One-sentence contract                  |
| Trigger          | When a real product need appears       |
| Enforcement      | Linter, tests, build, or human review  |
| Strength         | Intended MUST / SHOULD once enabled    |
| Promotion target | Stable or Evolving in `contracts.md`   |
| Noise risk       | low / med / high                       |
| Status           | `todo` / `ready` / `done -> contracts` |

## Pending list

### B1. Dependency boundary

| Field            | Content                                                                                                               |
| ---------------- | --------------------------------------------------------------------------------------------------------------------- |
| Policy draft     | State which packages may import each other; forbid cross-layer dependencies and uncontrolled third-party dependencies |
| Trigger          | The project grows multiple business layers, adapters, or a stable dependency direction                                |
| Enforcement      | `depguard` config + package-boundary docs; build an allowlist first, then enable                                      |
| Strength         | MUST                                                                                                                  |
| Promotion target | Stable                                                                                                                |
| Noise risk       | med                                                                                                                   |
| Status           | todo                                                                                                                  |

### B2. Complete struct initialization

| Field            | Content                                                                                                            |
| ---------------- | ------------------------------------------------------------------------------------------------------------------ |
| Policy draft     | Selected config, protocol, or immutable-state structs must initialize key fields explicitly                        |
| Trigger          | Zero values no longer represent a safe default, or a missing external-config field would cause a high-cost failure |
| Enforcement      | `exhaustruct` explicit mode with a type/path allowlist; do not enable it on every external option struct           |
| Strength         | MUST (selected types)                                                                                              |
| Promotion target | Evolving                                                                                                           |
| Noise risk       | high                                                                                                               |
| Status           | todo                                                                                                               |

### B3. Experimental or opinionated checkers

| Field            | Content                                                                                                         |
| ---------------- | --------------------------------------------------------------------------------------------------------------- |
| Policy draft     | Enable an aggregate linter's experimental/opinionated checkers only after team consensus                        |
| Trigger          | A checker has a clear payoff against product defects, and a minimal reproduction plus exception policy is ready |
| Enforcement      | Per-checker config on aggregate linters such as `gocritic`; update the linter risk register                     |
| Strength         | SHOULD -> MUST                                                                                                  |
| Promotion target | Evolving                                                                                                        |
| Noise risk       | high                                                                                                            |
| Status           | todo                                                                                                            |

### B4. Race tests in the daily gate

| Field            | Content                                                                                     |
| ---------------- | ------------------------------------------------------------------------------------------- |
| Policy draft     | Once concurrent behavior is a core product path, run the race detector on every daily check |
| Trigger          | A service, worker, cache, or shared state becomes the main downstream behavior              |
| Enforcement      | Fold `just race` into `just check` and adjust the time budget                               |
| Strength         | MUST (concurrent products)                                                                  |
| Promotion target | Evolving                                                                                    |
| Noise risk       | low / cost high                                                                             |
| Status           | todo                                                                                        |

### B5. Integration-test infrastructure

| Field            | Content                                                                                                |
| ---------------- | ------------------------------------------------------------------------------------------------------ |
| Policy draft     | Behavior that depends on a database, queue, or external service must have repeatable integration tests |
| Trigger          | Unit tests cannot cover real protocol or storage semantics                                             |
| Enforcement      | Choose testcontainers, a temporary service, or an explicit CI fixture, and add a dedicated recipe      |
| Strength         | MUST (matching product path)                                                                           |
| Promotion target | Evolving                                                                                               |
| Noise risk       | med                                                                                                    |
| Status           | todo                                                                                                   |

### B6. Stricter generated-code policy

| Field            | Content                                                                          |
| ---------------- | -------------------------------------------------------------------------------- |
| Policy draft     | Generated-code check scope, regeneration, and commit policy must be explicit     |
| Trigger          | The project introduces protobuf, OpenAPI, sqlc, or another stable code generator |
| Enforcement      | Generate commands, version pins, diff checks, and necessary linter exclusions    |
| Strength         | MUST (after a generator is enabled)                                              |
| Promotion target | Stable                                                                           |
| Noise risk       | med                                                                              |
| Status           | todo                                                                             |

### B7. Platform-specific implementations

| Field            | Content                                                                                    |
| ---------------- | ------------------------------------------------------------------------------------------ |
| Policy draft     | Introduce OS/arch-specific implementations and build tags only when the product needs them |
| Trigger          | A platform in the release matrix needs a different system API or behavior                  |
| Enforcement      | Narrow build tags, platform tests, and GoReleaser matrix verification                      |
| Strength         | MUST (matching platforms)                                                                  |
| Promotion target | Evolving                                                                                   |
| Noise risk       | med                                                                                        |
| Status           | todo                                                                                       |

## Explicitly not planned as defaults

These policies conflict with the template's generality or the current Invariant; enabling them requires a contract
change and a stated cost:

- Opening every internal experimental checker in `gocritic`, `revive`, and similar just to make `all` literal;
- Enabling `depguard` without a dependency-boundary policy;
- Forcing `exhaustruct` on every external and option struct;
- Folding `just audit` into routine `just check`, or maintaining a parallel offline Tombi gate;
- Keeping a second Makefile, Taskfile, or undocumented script entrypoint for local convenience.

## Change log

| Date             | Change                                                                                                           |
| ---------------- | ---------------------------------------------------------------------------------------------------------------- |
| Template initial | Establish the Go layer-2 backlog, aligned with `all` + explicit softens, tests, release, and platform boundaries |
