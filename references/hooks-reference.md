# Hooks reference

Exact signatures verified against `core v1.6.3`. All hooks are **package-level functions in `package main`**.
Import `github.com/maximhq/bifrost/core/schemas`.

## Lifecycle

```go
func GetName() string                       // required; must match config.json "name"
func Init(config any) error                 // optional; config from config.json
func Cleanup() error                        // required; release resources
```

`Init` is **not** on any core interface — the loader calls it by name.

## HTTP transport hooks (bifrost-http only; skipped when using the Go SDK)

```go
func HTTPTransportPreHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest) (*schemas.HTTPResponse, error)
func HTTPTransportPostHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, resp *schemas.HTTPResponse) error
func HTTPTransportStreamChunkHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, chunk *schemas.BifrostStreamChunk) (*schemas.BifrostStreamChunk, error)
```

`schemas.HTTPRequest` / `HTTPResponse` expose `Method`, `Path`, `Headers map[string]string`,
`Query map[string]string`, `Body []byte`, `PathParams` (request) and `StatusCode`, `Headers`, `Body`
(response).

- `HTTPTransportPreHook` — mutate `req.Headers` / `Body` / `Query` in place. Return `(nil, nil)` to continue,
  `(*HTTPResponse, nil)` to short-circuit, `(nil, error)` to fail. **On the pinned v1.6.x line this runs
  *before* authentication** — it is the credential / pre-auth seam. From v2.0 the dedicated
  `HTTPTransportPreAuthHook` takes that position and this hook moves to *after* auth. In that case a short-circuit
  reply is `&schemas.HTTPResponse{StatusCode: 403}`.
- `HTTPTransportPostHook` — mutate `resp` in place. Return `nil` to continue, an `error` to stop remaining
  post-hooks. **Not called for streaming responses.**
- `HTTPTransportStreamChunkHook` — return `(chunk, nil)` unchanged, `(nil, nil)` to drop the chunk, a modified
  chunk, or `(nil, error)` to abort the stream. `BifrostStreamChunk` embeds exactly one non-nil typed
  response (chat, text completion, responses, speech, transcription, image generation, or `BifrostError`).

Prefer the case-insensitive readers when reading: `req.CaseInsensitiveHeaderLookup("Content-Type")`,
`req.CaseInsensitiveQueryLookup("api_key")`.

## Core LLM hooks

```go
func PreRequestHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) error
func PreLLMHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error)
func PostLLMHook(ctx *schemas.BifrostContext, resp *schemas.BifrostResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error)
```

- **`PreRequestHook`** — the routing phase, once per request. Provider/model/fallback mutations propagate to
  every attempt; mutations in `PreLLMHook` do **not** apply to provider/model. A returned `error` is
  non-blocking (logged, pipeline continues). Helpers: `req.GetRequestFields()`,
  `req.SetProvider(...)`, `req.SetModel(...)`, `req.SetFallbacks(...)`.
- **`PreLLMHook`** — per attempt. Return `(req, nil, nil)` to continue, or a `*schemas.LLMPluginShortCircuit`
  to skip the provider call:

```go
type LLMPluginShortCircuit struct {
    Response *BifrostResponse        // synthetic response; skips provider
    Stream   chan *BifrostStreamChunk // synthetic stream; skips provider
    Error    *BifrostError           // synthetic error (can set AllowFallbacks)
}
```

- **`PostLLMHook`** — per attempt, reverse order. It receives **both** a response and an error, and must
  tolerate either being nil. Return them (possibly modified) to pass through; to recover, return a response
  with a nil error; to invalidate, return a nil response with an error. Do **not** drop the error — return the
  one you were given unless you intend to change behaviour.

Control fallbacks on a `*schemas.BifrostError` via `AllowFallbacks *bool` (nil = allowed).

## MCP hooks (optional; v1.5.x+)

```go
func PreMCPHook(ctx *schemas.BifrostContext, req *schemas.BifrostMCPRequest) (*schemas.BifrostMCPRequest, *schemas.MCPPluginShortCircuit, error)
func PostMCPHook(ctx *schemas.BifrostContext, resp *schemas.BifrostMCPResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostMCPResponse, *schemas.BifrostError, error)
func PreMCPConnectionHook(ctx *schemas.BifrostContext, req *schemas.BifrostMCPConnectRequest) (*schemas.BifrostMCPConnectRequest, *schemas.MCPConnectionShortCircuit, error)
func PostMCPConnectionHook(ctx *schemas.BifrostContext, resp *schemas.BifrostMCPConnectResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostMCPConnectResponse, *schemas.BifrostError, error)
```

Envelope hooks fire per `ping` / `list_tools` / `execute_tool`; connection hooks fire once per transport
setup. `MCPPluginShortCircuit{Response, Error}` and `MCPConnectionShortCircuit{Response, Error}` are separate
types. See the MCP section of the official writing-go-plugin doc for routing/allowlist recipes.

## Context API (`*schemas.BifrostContext`)

- `Log(level LogLevel, msg string)` — structured, plugin-scoped logging. Levels: `schemas.LogLevelDebug`,
  `LogLevelInfo`, `LogLevelWarn`, `LogLevelError`. Use this in hooks instead of `fmt.Println`.
- `SetValue(key any, value any)` / `Value(key any) any` — share state between hooks. Use typed
  `schemas.BifrostContextKey` constants.
- `GetPluginLogs()` / `DrainPluginLogs()` — read back emitted plugin logs.
- `AppendRoutingEngineLog(engine string, level LogLevel, msg string)` — explain a routing decision in
  observability.

The context is pooled after the request; do not retain it in a goroutine — capture `Root()` if you must.
