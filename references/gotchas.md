# Gotchas

Hard-won traps when writing native Bifrost plugins. Pinned to `core v1.6.3`.

## Structural

- **Hooks are package-level functions, not methods.** The loader resolves them by name via `plugin.Lookup`.
  A hook defined as a method on a struct, or under the wrong name, silently never fires.
- **`Init` is not on an interface** — it is called by name. Don't look for it in `BasePlugin`.
- **Only `GetName` and `Cleanup` are required.** Every hook is optional; the loader no-ops the ones you omit.

## Behaviour

- **Nil and error are both possible.** `PostLLMHook` must tolerate `resp == nil` *and* `bifrostErr == nil`.
  Guard both; never assume either is present.
- **Read/write ordering.** Post-hooks run in reverse of pre-hooks, and `HTTPTransportPostHook` runs *after*
  `PostLLMHook`. A context value written in an HTTP post-hook is invisible to the LLM hooks. Only read keys a
  strictly-earlier phase wrote.
- **`HTTPTransportPostHook` is skipped for streaming**; only `HTTPTransportStreamChunkHook` sees streamed
  chunks. Don't put non-streaming response logic in the only hook that never runs for streams.
- **Auth ordering is defined by the transport, not core, and it changed in v2.0.** On the pinned v1.6.x line
  `HTTPTransportPreHook` runs *before* authentication (it is the credential seam); v2.0 adds the dedicated
  `HTTPTransportPreAuthHook` and moves `HTTPTransportPreHook` to *after* auth. The live docs describe the v2.0
  arrangement — don't carry it back to a v1.6.x host. Verify against the transport version you actually run.
- **HTTP transport hooks don't run under the Go SDK.** If the plugin must work both ways, don't rely on them.
- **Per-attempt vs per-request.** `PreLLMHook`/`PostLLMHook` run again on every fallback; `PreRequestHook` is
  the only phase whose provider/model mutations stick. Put routing in `PreRequestHook`.

## Concurrency and lifetime

- **`HTTPRequest` / `HTTPResponse` are pooled.** Use the acquire/release helpers; never retain a reference
  past the hook.
- **`*BifrostContext` is pooled after the request.** Don't keep it in a goroutine; capture `ctx.Root()`.
- **`*Trace` (observability) must not outlive `Inject`.**

## Build

- **`-buildmode=plugin` + `CGO_ENABLED=1`,** or the output isn't loadable. A plain `go build` produces a
  normal executable.
- **Exact parity required:** OS, CPU architecture, and Go version must match the host Bifrost binary.
  `plugin.Open` fails with a version mismatch otherwise.
- **No cross-compilation.** Build on the target platform; Docker works for Linux targets.
- **Linux / macOS only** — Go plugins don't support Windows.
- **Host must be a dynamic build** of Bifrost.

## Documentation drift

- **The live docs track a newer line than `core v1.6.3`** (`HTTPTransportPreAuthHook`, response-headers hook,
  A2A). Re-verify signatures against the live docs; see `references/versioning.md`.
- **WASM plugins are deprecated** — don't start new work on them.

## Practical

- Read headers/query case-insensitively: `req.CaseInsensitiveHeaderLookup(...)`.
- Log through `ctx.Log(...)`, not `fmt.Println`, so entries are scoped and retrievable.
- Use typed `schemas.BifrostContextKey` constants for context keys — string collisions are silent.
- Keep a compile-time guard per exported hook so a `core` upgrade that changes a signature fails the build
  (see the bundled example's `main_test.go` / `main.go`).
