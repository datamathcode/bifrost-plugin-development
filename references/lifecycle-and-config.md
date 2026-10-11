# Lifecycle and configuration

## Lifecycle

1. **Load** — Bifrost `plugin.Open()`s the `.so`.
2. **Initialize** — `Init(config any)` is called with the plugin's `config` value.
3. **Hooks** — hooks fire as requests flow (see `references/hooks-reference.md`).
4. **Cleanup** — `Cleanup()` at shutdown.

`GetName()` and `Cleanup()` are required. `Init` is optional and is resolved by name.

### `Init(config any)`

`config` is the raw value from the plugin's `config` object in `config.json` — typically a
`map[string]interface{}`. Type-assert defensively and return an error on bad input:

```go
func Init(config any) error {
    cfg, ok := config.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid config format")
    }
    apiKey, _ := cfg["api_key"].(string)
    _ = apiKey
    return nil
}
```

## `config.json`

Add the plugin to Bifrost's `config.json`:

```json
{
  "plugins": [
    {
      "enabled": true,
      "name": "my-plugin",
      "path": "/absolute/path/to/my-plugin.so",
      "config": { "api_key": "your-api-key", "custom_setting": "value" }
    }
  ]
}
```

- `name` **must** equal the plugin's `GetName()`.
- `config` is passed verbatim to `Init`.
- Optional `placement` (`pre_builtin` / `post_builtin` / `builtin`, default `post_builtin`) and `order`
  control where the plugin sits relative to built-in plugins.
- Plugins may also be configured via Helm (`deployment-guides/helm/plugins`) or the config-JSON plugin docs.

## Build

```bash
# from the plugin directory
CGO_ENABLED=1 go build -buildmode=plugin -o build/my-plugin.so main.go
```

Requirements:

- `-buildmode=plugin` and `CGO_ENABLED=1`. A plain `go build` yields a normal binary Bifrost cannot load.
- The host Bifrost binary must be built with dynamic linking.
- The `.so` must match the host's **OS, architecture, and Go version**.
- Linux and macOS only.

## Verify it loaded

Start `bifrost-http`; you should see the `Init` output and a plugin-loaded line. Then issue a request and
watch for hook log lines:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"Hello!"}]}'
```

Non-streaming: `HTTPTransportPreHook → PreLLMHook → PostLLMHook → HTTPTransportPostHook`.
Streaming: the stream-chunk hook fires per chunk instead of `HTTPTransportPostHook`.
