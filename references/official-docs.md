# Official documentation

This skill is a verified snapshot. The Bifrost plugin surface evolves, and the live docs may describe a newer
version line than this skill's pinned `core v1.6.3`. **Verify before you finish**: fetch the relevant page(s)
below and reconcile signatures and behaviour.

## Canonical pages

| Topic | URL |
| --- | --- |
| Plugins — getting started, architecture, hook order | https://docs.getbifrost.ai/plugins/getting-started |
| Writing Go plugins — signatures, examples, Makefile, config | https://docs.getbifrost.ai/plugins/writing-go-plugin |
| Writing WASM plugins (deprecated) | https://docs.getbifrost.ai/plugins/writing-wasm-plugin |
| Building a dynamically-linked Bifrost binary | https://docs.getbifrost.ai/plugins/building-dynamic-binary |
| Plugin sequencing (placement / order) | https://docs.getbifrost.ai/plugins/sequencing |
| Plugin architecture (internal design) | https://docs.getbifrost.ai/architecture/core/plugins |
| `config.json` plugins section | https://docs.getbifrost.ai/deployment-guides/config-json/plugins |
| Helm plugins | https://docs.getbifrost.ai/deployment-guides/helm/plugins |

## Canonical source examples (in the Bifrost repo)

- Hello world (the reference the docs build from): https://github.com/maximhq/bifrost/tree/main/examples/plugins/hello-world
- MCP-only (all four MCP hooks): https://github.com/maximhq/bifrost/tree/main/examples/plugins/mcp-only

## Discovery index

The docs publish a machine-readable index you can crawl to find every page:

- https://docs.getbifrost.ai/llms.txt

## How to verify (do this before finishing)

1. Fetch https://docs.getbifrost.ai/plugins/writing-go-plugin and read the tab for the host's version line.
2. Compare each hook signature and return value against what you wrote. A drift here is the single most likely
   reason a plugin fails to register.
3. Check the getting-started page's execution-order section for any ordering changes.
4. If the host pins a specific `core` version, cross-check `references/versioning.md` too.
