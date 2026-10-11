# Versioning

This skill's snapshot is pinned to **`github.com/maximhq/bifrost/core v1.6.3`**, which corresponds to the docs'
**"v1.5.x+"** tab. Bifrost is versioned across lines, and the exported symbol set and signatures differ.

## Symbol set by line

| Line | Symbols |
| --- | --- |
| **v1.3.x** | `TransportInterceptor(ctx, url, headers, body)`; `PreHook`; `PostHook`; `PluginShortCircuit` |
| **v1.4.x** | `HTTPTransportPreHook/PostHook/StreamChunkHook`; `PreLLMHook`; `PostLLMHook` (no `PreRequestHook`) |
| **v1.5.x+** | adds `ctx.Log`; MCP hooks (`Pre/PostMCPHook`, `Pre/PostMCPConnectionHook`); `PostLLMHook` takes `*BifrostError` |
| **v1.6.x** | adds `PreRequestHook` (routing phase); config placement/order |
| **v2.0** | adds `HTTPTransportPreAuthHook` (runs pre-auth); `HTTPTransportPreHook` moves to **after** auth |

`schemas.PluginShortCircuit` still exists in v1.6.3 as a deprecated alias for `LLMPluginShortCircuit` (from
v1.3.x compatibility).

## Hooks newer than the pinned core v1.6.3

The live docs describe a newer line. These appear in current docs but **not** in `core v1.6.3`:

- `HTTPTransportPreAuthHook` (v2.0+) — intercept before authentication; the credential seam.
- `HTTPTransportResponseHeadersHook` — response headers phase.
- `A2APlugin` (`Pre/PostA2AHook`) — agent-to-agent surface.

If the host runs a core newer than v1.6.3, these become available; if it runs v1.6.3, exporting them is
harmless but they will not be called.

## Go version pin — the docs disagree with themselves

- The docs' Prerequisites say **Go 1.27.0**; a note in the same page says pin `go.mod` to **1.26.1**.
- The bundled `assets/example-plugin/` uses **`go 1.26.4`**.

What actually matters is parity: **the `.so` must be built with the same Go version as the host Bifrost
binary.** Pin `go.mod` to the host's version, not to a number copied from the docs.

## Policy

When the host's core version is known, match this skill's signatures to it. When it is unknown, build against
`v1.6.3` (this snapshot) and **verify against the live docs before finishing** (`references/official-docs.md`).
