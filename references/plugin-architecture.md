# Plugin architecture

Verified against `github.com/maximhq/bifrost/core v1.6.3`. Re-check against the live docs before relying on
version-sensitive details (`references/official-docs.md`).

## Delivery mechanisms

Bifrost can run plugin code in more than one way. **Only native Go `.so` is both supported and current.**

| Mechanism | What it is | Status |
| --- | --- | --- |
| **Native Go `.so`** | A `package main` plugin built with `-buildmode=plugin`, loaded at runtime | **Supported — use this** |
| **In-process Go SDK** | Types registered in `BifrostConfig` when embedding `core` as a library | Supported (no `.so`) |
| **WASM** | WebAssembly plugin | **Deprecated** (docs: "Use Go plugins for new development") |
| **Webhooks** | Webhook-based plugins | Planned, not yet shipped |

### Native `.so` loading

The loader lives in Bifrost's `framework/plugins` module (imported by `transports/bifrost-http`), **not** in
`core`. It:

1. `plugin.Open()`s the `.so`, then
2. `plugin.Lookup`s **package-level symbols** by name — `GetName`, `Cleanup`, and any optional hooks, and
3. wraps what it finds in a dynamic adapter that no-ops the hooks a plugin didn't export.

Because lookup is by name, the plugin exports *functions*, not a type implementing an interface. This is the
single most important structural fact: a hook that is a method on a struct, or misnamed, is invisible to the
loader.

## Interfaces (for in-process Go SDK plugins)

`core/schemas/plugin.go` declares interfaces used when embedding Bifrost. A native `.so` does not implement
these as methods, but they are the canonical definition of the hook set and signatures.

- **`BasePlugin`** — `GetName() string`; `Cleanup() error`.
- **`HTTPTransportPlugin`** (`BasePlugin` +) — `HTTPTransportPreHook`, `HTTPTransportPostHook`,
  `HTTPTransportStreamChunkHook`.
- **`LLMPlugin`** (`BasePlugin` +) — `PreRequestHook`, `PreLLMHook`, `PostLLMHook`.
- **`MCPPlugin`** — `PreMCPHook`, `PostMCPHook`.
- **`MCPConnectionPlugin`** — `PreMCPConnectionHook`, `PostMCPConnectionHook`; `MCPPluginNoOpHooks` supplies
  the pair a plugin doesn't want to implement.
- **`ObservabilityPlugin`** — `Inject(ctx context.Context, trace *Trace) error`, called asynchronously after
  the response is written.
- **`ConfigMarshallerPlugin`** — `MarshalConfigForStorage` / `RedactConfig` for custom config serialization.

## Execution model

- Pre-hooks run in **registration order**; post-hooks run in **reverse** order.
- `HTTPTransportPreHook`, `PreRequestHook`, `HTTPTransportPostHook` run **once per top-level request**.
- `PreLLMHook` / `PostLLMHook` run **once per provider attempt** — the primary call plus each fallback.
- Plugins can short-circuit: `PreLLMHook` may return a synthetic response (skip the provider), and
  `PostLLMHook` may recover from an error or invalidate a response (see `references/hooks-reference.md`).

## `PluginConfig`

The `plugins[]` entry shape (from `config.json`) is `schemas.PluginConfig`:

| Field | Meaning |
| --- | --- |
| `Enabled` | load on startup |
| `Name` | identifier; must match the plugin's `GetName()` |
| `Path` | path to the `.so` |
| `Config` | arbitrary value passed to `Init` |
| `Version` | optional schema version |
| `Placement` | `pre_builtin` / `post_builtin` (default) / `builtin` |
| `Order` | position within the placement group |

See `references/lifecycle-and-config.md` for the `config.json` example and load verification.
