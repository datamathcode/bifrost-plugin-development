# AGENTS.md

## What this repo is

An **agent skill** that teaches an agent how to build a Bifrost gateway plugin. The repo is **not** a plugin
itself — the deliverable is `SKILL.md` at the root. A complete, tested plugin lives under
`assets/example-plugin/` as the starter the skill tells agents to copy. See `CONTEXT.md` for the vocabulary
(note the deliberate distinction between a *Bifrost plugin* and an *agent skill*).

## Layout

- `SKILL.md` — the skill: YAML frontmatter (`name`, `description`) + the teaching body.
- `references/` — on-demand docs: `plugin-architecture.md`, `hooks-reference.md`,
  `lifecycle-and-config.md`, `versioning.md`, `gotchas.md`, `official-docs.md`.
- `assets/example-plugin/` — the example plugin; **its own Go module**.
- `CONTEXT.md` — glossary (glossary only; no implementation detail).
- `docs/agents/*` — config consumed by the engineering skills; `docs/adr/` — decisions.

## Working on the example plugin

- It is a **separate Go module**. Run Go commands from `assets/example-plugin/`, not the repo root.
- Build: `make build` (or `make dev`). A build **must** use `-buildmode=plugin` with `CGO_ENABLED=1`; a plain
  `go build` produces a normal binary that Bifrost cannot load.
- Test: `go test ./...` from that directory. The suite covers each hook and includes compile-time signature
  guards pinning every hook to its `core` interface method.
- Pinned to `github.com/maximhq/bifrost/core v1.6.3`. The `.so` must match the host Bifrost binary's OS,
  architecture, and Go version; there is no cross-compilation (Docker for Linux targets).
- `PostLLMHook` takes `(ctx, resp, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse,
  *schemas.BifrostError, error)` — do not regress it to the older two-value form.

## Working on the skill

- Keep the `SKILL.md` body under ~500 lines; push detail into `references/` and point to it.
- Any signature or behaviour claim should be verified against `core v1.6.3` and the live docs
  (`references/official-docs.md`). The official docs describe a newer version line than the pin — flag drift,
  don't silently copy it.
- `CONTEXT.md` is a glossary only. Implementation decisions belong in `docs/adr/`, not there.

## Not present (do not assume)

No CI, no linter config, and **no root Go module** — the module is under `assets/example-plugin/`. The built
`.so` and `build/` directories are git-ignored.
