# Bifrost Plugin Development

An **agent skill** that teaches an agent how to build a [Bifrost](https://github.com/maximhq/bifrost) gateway
plugin. The repository is the skill; the deliverable is `SKILL.md`. A complete, tested example plugin ships
alongside it under `assets/example-plugin/`.

## Installing this skill

Drop this repository into your skills directory. Copy it, or symlink it so edits stay in sync:

```bash
# Option A — copy
cp -r . ~/.agents/skills/bifrost-plugin-development

# Option B — symlink (edit in place)
ln -s "$(pwd)" ~/.agents/skills/bifrost-plugin-development
```

The skill's name is `bifrost-plugin-development`. Once installed, an agent asked to build, extend, or debug a
Bifrost plugin will load `SKILL.md`.

## Layout

| Path | What it is |
| --- | --- |
| `SKILL.md` | The skill: frontmatter (`name`, `description`) + the teaching body |
| `references/` | On-demand detail: architecture, hooks, lifecycle/config, versioning, gotchas, official docs |
| `assets/example-plugin/` | A minimal, tested Go plugin module — the starter the skill tells agents to copy |
| `assets/redactor-plugin/` | A second worked example: request/response/streaming secret redaction |
| `CONTEXT.md` | Glossary of this repo's domain language |
| `docs/adr/` | Architecture decision records |
| `docs/agents/` | Configuration consumed by the engineering skills |

## The bundled example plugin

`assets/example-plugin/` is a real Go module (`github.com/datamathcode/bifrost-plugin-development`) pinned to
`bifrost/core v1.6.3`. Run Go commands from inside it:

```bash
cd assets/example-plugin
make build     # -> build/bifrost-plugin-development.so
go test ./...  # hook tests + compile-time signature guards
```

See `assets/example-plugin/README.md` for build, configuration, and hook-ordering details.

## License

MIT
