# Bifrost Plugin Development

A native Go plugin (`.so` shared object) for the Bifrost LLM gateway, built as a reference implementation demonstrating the Bifrost plugin architecture.

## Overview

This plugin demonstrates all Bifrost hook types:
- **HTTP Transport hooks**: `HTTPTransportPreHook`, `HTTPTransportPostHook`, `HTTPTransportStreamChunkHook`
- **Core lifecycle hooks**: `PreRequestHook`, `PreLLMHook`, `PostLLMHook`
- **Plugin lifecycle**: `Init`, `GetName`, `Cleanup`

## Building

```bash
# Requires Go 1.26+ (matching Bifrost's Go version)
# Requires CGO_ENABLED=1
# A native .so plugin must match the host Bifrost binary's OS, arch, and Go version.

make dev        # development build -> build/bifrost-plugin-development.so
make build      # optimized build  -> build/bifrost-plugin-development.so
```

Or build directly, outputting to the same `build/` directory the Makefile uses:

```bash
go mod tidy
mkdir -p build
CGO_ENABLED=1 go build -buildmode=plugin -o build/bifrost-plugin-development.so main.go
```

## Configuration

Add to your Bifrost `config.json`:

```json
{
  "plugins": [
    {
      "name": "bifrost-plugin-development",
      "enabled": true,
      "path": "/path/to/bifrost-plugin-development.so",
      "config": {}
    }
  ]
}
```

## Requirements

- Bifrost built with dynamic linking (`make build DYNAMIC=1`)
- Linux or macOS (Darwin) - no Windows support
- Same Go version as Bifrost binary
- Same architecture (amd64/arm64) as Bifrost binary

## Hook Execution Order

1. `HTTPTransportPreHook` - HTTP only, before Bifrost core
2. `PreRequestHook` - Once per request, routing decisions
3. `PreLLMHook` - Per provider attempt, before provider call
4. Provider call
5. `PostLLMHook` - Per provider attempt, after provider call (reverse order)
6. `HTTPTransportPostHook` - HTTP only, after Bifrost core (reverse order)
7. `HTTPTransportStreamChunkHook` - Per streaming chunk (HTTP only)

## Context Passing

Hooks share state via `BifrostContext.SetValue()` and `BifrostContext.Value()` using typed constants as keys to avoid collisions.

## License

MIT