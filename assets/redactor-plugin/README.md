# redactor-plugin

A worked Bifrost plugin that redacts secrets and PII **in the request on the way to the provider**
(ingress), **in the response on the way back** (egress), and **in streamed chunks**.

It is the second bundled example for the [`bifrost-plugin-development`](../../SKILL.md) skill, chosen to
exercise hooks the hello-world example doesn't: the two LLM hooks plus the streaming hook.

## Hooks

| Hook | Role |
| --- | --- |
| `PreLLMHook` | redact outgoing request messages (ingress) |
| `PostLLMHook` | redact incoming response messages (egress) |
| `HTTPTransportStreamChunkHook` | redact each streamed chunk, watermark the final one |

`GetName` and `Cleanup` are the required lifecycle symbols; `Init` is optional.

## What it redacts

A small illustrative regex set (`redactionRules` in `main.go`): private-key headers, AWS access-key IDs,
provider API keys (`sk-`/`pk-`/`rk-`/`api_`), bearer tokens, emails, SSNs, and credit-card numbers. Each match
becomes `[REDACTED:<type>]`.

This is a teaching example, **not** production-grade DLP. In particular:

- Regex redaction is best-effort; widen the rules or call a real PII/secret detector.
- **Streaming redaction is per-chunk.** A secret split across two chunks (`"sk-live"` + `"abc123"`) will not
  match and slips through. Fixing that needs a rolling buffer across chunks, which adds latency.

## Build and test

```bash
# from this directory
go test ./...
make build        # -> build/redactor-plugin.so
```

Requires `CGO_ENABLED=1` and `-buildmode=plugin`. The `.so` must match the host Bifrost binary's OS,
architecture, and Go version (no cross-compilation; Linux/macOS only).

## Configure

```json
{
  "plugins": [
    {
      "enabled": true,
      "name": "redactor-plugin",
      "path": "/absolute/path/to/redactor-plugin.so",
      "config": {}
    }
  ]
}
```

`name` must equal `GetName()` (`"redactor-plugin"`).
