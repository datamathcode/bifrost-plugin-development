# AGENTS.md

## What this is

A single-file native Go plugin (`main.go`) for the [Bifrost](https://github.com/maximhq/bifrost) LLM
gateway. It is loaded as a `.so` shared object via Go's `plugin` package — **not** a library and
**not** a standalone binary. Module: `github.com/datamathcode/bifrost-plugin-development`, Go
1.26.4, depends on `github.com/maximhq/bifrost/core v1.6.3`.

## Commands

- `make dev` — build for the host platform → `build/bifrost-plugin-development.so`
- `make build` — optimized build; cross-compile with `make build GOOS=linux GOARCH=amd64`
  (non-Linux targets via Docker)
- `make vet` / `go vet ./...` — clean today
- `make fmt` — `go fmt ./...`
- `make test` — runs `go test ./...`, but there are **no test files**, so it is a no-op that
  always passes. Do not treat it as coverage.
- `make install BIFROST_PLUGIN_DIR=/path/to/plugins` — build + copy

A build **must** use `-buildmode=plugin` with `CGO_ENABLED=1`. A plain `go build` produces a
normal executable, not a loadable plugin.

## Architecture

- Every hook (`Init`, `GetName`, `Cleanup`, and all the `*Hook` funcs) is a **package-level
  function** in `main.go`, not a method on a type. The loader that resolves them by name lives
  in `bifrost-http` and is **not vendored here**; the `core` module only declares the
  `BasePlugin` / `HTTPTransportPlugin` / `LLMPlugin` interfaces in `schemas/plugin.go`.
- Hook order: `HTTPTransportPreHook` → `PreRequestHook` → `PreLLMHook` → provider →
  `PostLLMHook` → `HTTPTransportPostHook` → (streaming only) `HTTPTransportStreamChunkHook`.
  The HTTP post-hooks run **after** the LLM post-hook.
- Hooks share state through `ctx.SetValue` / `ctx.Value` using typed `schemas.BifrostContextKey`
  constants (declared at the top of `main.go`). Keep new keys as typed constants — collisions are
  silent.

## Gotchas

- **`PostLLMHook` (`main.go:97`) does not match `bifrost/core v1.6.3`.** Core requires
  `(ctx, resp, bifrostErr *BifrostError) (*BifrostResponse, *BifrostError, error)`; this code
  uses `bifrostErr error` and returns only two values. Fix this before extending the LLM hook
  surface. All other hook signatures currently match core.
- `value2` in `PostLLMHook` is always `nil`: `transportPostHookKey` is written by
  `HTTPTransportPostHook`, which runs *after* `PostLLMHook`. Hook read/write order matters.
- `Init(config any)` is not part of any core interface (`BasePlugin` is only `GetName` +
  `Cleanup`); the external loader invokes it by name.
- `bifrost-plugin-development.so` (28 MB) is **committed to git** and there is no `.gitignore`.
  It is a build artifact — regenerate with `make dev`, never hand-edit it. Native plugins are
  locked to the host Bifrost binary's OS, architecture, and Go version.
- README's build command outputs to the repo root; `make build` outputs to `build/`. Trust the
  Makefile.

## Not present (do not assume)

No CI, no linter config, no tests, no `.gitignore`, no other instruction files.

## Agent skills

### Issue tracker

Issues and specs live as GitHub issues in `datamathcode/bifrost-plugin-development` (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical roles map 1:1 to labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
