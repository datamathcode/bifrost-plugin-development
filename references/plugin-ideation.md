# Plugin ideation

What can you actually build with a Bifrost plugin? This is a menu, not a spec. Capabilities are grouped by the
hook they live in; each entry names the hook(s) it needs and why it can't just be config.

## First, the gap

Much of what you'd reach for first is already first-class Bifrost configuration, **not** plugin territory:
providers, MCP clients, semantic cache, OTel/Datadog/Maxim/Prometheus, governance (virtual keys, budgets, rate
limits), guardrails, routing rules. A plugin earns its place only for the *gap* — org-specific, cross-cutting,
or not-yet-built logic. Before writing one, check the built-ins so you don't reinvent them.

## Capabilities by hook

| Capability | Hook(s) | Notes |
| --- | --- | --- |
| Credential / virtual-key derivation, request shaping | `HTTPTransportPreHook` (pre-auth on v1.6.x) | HTTP transport only; see `versioning.md` for the v2.0 auth move |
| Streaming transform, redaction, filtering, watermarking | `HTTPTransportStreamChunkHook` | per chunk; the hook most people never touch |
| Routing, model aliasing, provider selection, residency | `PreRequestHook` | the only phase whose provider/model mutations commit |
| Request transform / validate / redact; short-circuit (cache hit, refusal) | `PreLLMHook` | can skip the provider entirely by returning a short-circuit |
| Response transform / redact / reject; cost accounting; fallback shaping | `PostLLMHook` | tolerate nil response **and** nil error |
| MCP tool-call governance; connect-time header injection | `Pre/PostMCPHook`, `Pre/PostMCPConnectionHook` | MCP surface |
| Trace fan-out to a niche backend | `ObservabilityPlugin.Inject` | async, after the response is written |

## Candidate plugins

★ = good flagship/demo material. ⚠ = verify the built-in doesn't already cover it.

- ★ **PII / secret redaction (ingress + egress + streaming)** — `PreLLMHook` redacts the outgoing request,
  e.g. `PostLLMHook` the response, `HTTPTransportStreamChunkHook` the stream. See `assets/redactor-plugin/`.
- ★ **Deterministic record-and-replay** — `PostLLMHook` records provider responses keyed by request hash;
  `PreLLMHook` replays them (as a short-circuit) for CI and stable demos. Exercises short-circuit; clear
  latency/cost payoff.
- ★ **Streaming redaction / watermark** — scrub or tag tokens as they stream in `HTTPTransportStreamChunkHook`.
  Demoable live; shows the obscure hook. Hard part: chunk boundaries.
- ★ **Cost / chargeback ledger** — `PostLLMHook` + `ctx.CalculateCost` / `ctx.GetModelInfo` emits per-request
  cost to a webhook/CSV/SQLite keyed by tenant. ⚠ Enterprise has budgets; self-hosted chargeback to *your*
  system is the gap.
- ★ **MCP tool governance** — `PreMCPHook` per-client tool allowlist/denylist and result redaction;
  `PreMCPConnectionHook` injects transport headers.
- **Model aliasing / allowlist** — `PreRequestHook` canonicalises `gpt-4o` → `openai/gpt-4o`, enforces a
  per-team model allowlist. Demos the routing phase.
- **Custom rate limiter** — token-bucket / multi-dimensional limits in `PreLLMHook`. ⚠ Governance covers basic
  virtual-key limits; the gap is bespoke dimensions (per-user, per-prompt-class).
- **Response schema validation + repair** — validate JSON in `PostLLMHook`; reject or re-ask.
- **Fallback shaping by error class / tenant** — `PostLLMHook` sets `AllowFallbacks`.
- **Data-residency routing** — `PreRequestHook` routes EU tenants to EU providers. Strong enterprise story.
- **Prompt-injection / jailbreak gate** — heuristic blocklist in `PreLLMHook` → short-circuit a refusal.
- **Tenant system-prompt injection** — `PreLLMHook` injects a tenant-specific system prompt.
- **Niche trace sink** — `ObservabilityPlugin.Inject` → Loki/syslog/CSV/webhook. ⚠ Log exports may cover S3/GCS.

## How to choose

Prefer capabilities that (a) aren't already config, (b) exercise a hook that's hard to discover (short-circuit,
stream chunk, routing, MCP), and (c) have a visible payoff (cost, latency, safety). The two bundled examples —
`assets/example-plugin/` (the minimal shape) and `assets/redactor-plugin/` (multi-hook redaction) — bracket the
range.
