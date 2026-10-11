---
name: bifrost-plugin-development
description: Build, extend, or debug a custom Bifrost gateway plugin (native Go `.so`): implementing hooks like `HTTPTransportPreHook`/`PreLLMHook`/`PostLLMHook`, wiring `Init`/`GetName`/`Cleanup`, configuring `config.json`, and building with `-buildmode=plugin`. Use whenever a task mentions Bifrost plugins, Bifrost request/response interception, or the Bifrost plugin hooks, even if not called a "plugin".
---

# Building Bifrost Plugins

A Bifrost plugin is a **Go shared object (`.so`)** that the [Bifrost](https://github.com/maximhq/bifrost) LLM
gateway loads at runtime to intercept and modify requests and responses. This skill teaches you to build one
from scratch, wire it into Bifrost, and verify it.

The authoritative surface is defined by the `github.com/maximhq/bifrost/core/schemas` package. This skill is a
verified snapshot pinned to **`core v1.6.3`**. Before you finish any real task, **verify signatures against the
live docs** — see `references/official-docs.md` (Bifrost ships newer hooks than v1.6.3).

## 1. What a plugin is, and when to reach for this

Use a plugin when you need to sit in the request lifecycle: transform requests/responses, enforce validation,
add caching, emit telemetry, or govern routing. The supported delivery mechanism is a **native Go `.so`**.
(In-process Go SDK plugins and deprecated WASM plugins also exist; see `references/plugin-architecture.md`.)

## 2. The required shape

A plugin is `package main` with **package-level functions**. The loader resolves them by **name** via Go's
`plugin.Lookup` — the hooks are *not* interface methods on a type. Every symbol is optional except two:

| Symbol | Required? | Purpose |
| --- | --- | --- |
| `GetName() string` | **required** | Unique plugin name; must match `name` in `config.json` |
| `Cleanup() error` | **required** | Release resources on shutdown |
| `Init(config any) error` | optional | Receive config from `config.json` at load |
| `HTTPTransportPreHook` | optional | HTTP transport: before core, after auth |
| `HTTPTransportPostHook` | optional | HTTP transport: after core (**not** for streaming) |
| `HTTPTransportStreamChunkHook` | optional | HTTP transport: per streaming chunk |
| `PreRequestHook` | optional | Once per request: routing (provider/model/fallbacks) |
| `PreLLMHook` | optional | Per provider attempt, before the call; can short-circuit |
| `PostLLMHook` | optional | Per provider attempt, after the call; modify response/error |
| `PreMCPHook` / `PostMCPHook` | optional | Per MCP tool call |
| `PreMCPConnectionHook` / `PostMCPConnectionHook` | optional | Per MCP transport connect |

The loader tolerates missing hooks — a plugin that exports only `GetName` and `Cleanup` loads fine. Implement
only the hooks you need. Full signatures and return-value contracts are in `references/hooks-reference.md`.

## 3. Quickstart

1. **Start from the bundled example.** Copy `assets/example-plugin/` to a new directory — it is a complete,
   tested, buildable plugin (`main.go`, `main_test.go`, `go.mod`, `Makefile`, `README.md`). Rename the module
   in `go.mod` and update `GetName()` to your plugin's name.
2. **Depend on core.** `go get github.com/maximhq/bifrost/core@latest` (or pin `@v1.6.3` to match this skill).
3. **Implement your hooks.** Delete the demo hooks you don't need; keep `GetName` and `Cleanup`. See §4–§5.
4. **Build.** `make build` (or `CGO_ENABLED=1 go build -buildmode=plugin -o build/my-plugin.so main.go`).
   The output **must** be a plugin; a plain `go build` produces a normal executable that Bifrost cannot load.
5. **Configure.** Add an entry to Bifrost's `config.json` (§6 of `references/lifecycle-and-config.md`).
6. **Verify.** Start `bifrost-http` and confirm `Init called` + hook log lines while exercising a request.

## 4. Hook catalogue and execution order

Order for a v1.5.x+/v1.6.x request:

```
HTTPTransportPreHook → PreRequestHook → PreLLMHook → provider →
PostLLMHook → HTTPTransportPostHook → (streaming only) HTTPTransportStreamChunkHook
```

Rules that bite:

- **Post-hooks run in reverse** of pre-hooks.
- `HTTPTransportPostHook` runs **after** `PostLLMHook` — so values a hook writes there are **not** readable in
  `PostLLMHook` or `PreLLMHook`. Only read keys written by an earlier phase.
- `HTTPTransportPostHook` is **not called for streaming responses** — use `HTTPTransportStreamChunkHook`.
- The HTTP transport hooks are **not invoked at all** when Bifrost is used as a Go SDK.
- `PreLLMHook`/`PostLLMHook` run **per provider attempt** (again on each fallback); `PreRequestHook` runs
  **once per request** and is the only phase whose provider/model mutations propagate.

## 5. Which hooks do I need?

| Goal | Hooks |
| --- | --- |
| Modify incoming HTTP request / headers | `HTTPTransportPreHook` |
| Modify non-streaming HTTP response / headers | `HTTPTransportPostHook` |
| Transform streamed chunks | `HTTPTransportStreamChunkHook` |
| Choose provider/model/fallbacks | `PreRequestHook` |
| Transform or validate an LLM request | `PreLLMHook` |
| Cache-hit / synthetic response without a provider call | `PreLLMHook` returning a short-circuit |
| Transform a response, or recover/invalidate an error | `PostLLMHook` |
| Govern MCP tool calls / connections | `PreMCPHook` / `PostMCPHook` / `Pre/PostMCPConnectionHook` |

## 6. Build and platform constraints

- Build with **`-buildmode=plugin`** and **`CGO_ENABLED=1`**.
- The `.so` must match the host Bifrost binary's **OS, architecture, and Go version**.
- **No cross-compilation.** Build on the target platform (Docker works for Linux targets).
- **Linux and macOS only** (Go plugins do not support Windows).
- Bifrost must be a dynamically-linked build to load plugins.
- Details and gotchas: `references/gotchas.md`.

## 7. Testing hooks

Hooks are ordinary package-level functions over exported `schemas` values, so test them directly in
`package main` — construct a `*schemas.BifrostContext` with `schemas.NewBifrostContext`, build the request/
response literals, call the hook, and assert on returned values and in-place mutations. The bundled example's
`main_test.go` is a working template. Add a compile-time guard for every hook you export so a future `core`
upgrade that changes a signature fails the build:

```go
var _ func(*schemas.BifrostContext, *schemas.BifrostResponse, *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) = PostLLMHook
```

## 8. Version compatibility

This skill targets **`core v1.6.3`** (the "v1.5.x+" docs tab). Bifrost has other lines with different symbol
sets — v1.3.x uses `TransportInterceptor`/`PreHook`/`PostHook`; v2.0 adds `HTTPTransportPreAuthHook` and runs
`HTTPTransportPreHook` after auth. If the host pins a different core version, read
`references/versioning.md` and re-verify against the live docs before writing code.

## 9. Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `plugin.Open: ... plugin was built with a different version of package` | Go version / arch / OS mismatch with the host binary |
| Plugin loads but no hooks fire | Hook not exported as a **package-level** function, or wrong name |
| Plugin does not load at all | Built without `-buildmode=plugin`, or Bifrost is not a dynamic build |
| `Init` not called | Symbol missing or misnamed; config entry `enabled` false |
| Nil panic in a hook | Response **and** error can both be nil — guard them |
| Post-hook context value is `nil` | Reading a key written by a later phase (see §4) |

## References

Read these on demand:

- `references/plugin-architecture.md` — interfaces, delivery mechanisms, how loading works.
- `references/hooks-reference.md` — every hook's exact signature, return values, and examples.
- `references/lifecycle-and-config.md` — `Init`/`GetName`/`Cleanup`, `config.json`, install, load, verify.
- `references/versioning.md` — v1.3 / v1.4 / v1.5+ / v2.0 deltas vs the pinned `v1.6.3`.
- `references/gotchas.md` — nil handling, ordering, pooling/aliasing, build traps.
- `references/official-docs.md` — canonical doc URLs and the "verify against latest" step.
