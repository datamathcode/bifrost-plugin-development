# Changelog

All notable changes to this project are documented here. Versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-10-11

The repository is repurposed from a single reference plugin into a portable **agent skill** that teaches an
agent how to build a Bifrost gateway plugin. Recorded in `docs/adr/0001`.

### Added

- `SKILL.md` and `references/` — the skill plus on-demand docs: `plugin-architecture.md`,
  `hooks-reference.md`, `lifecycle-and-config.md`, `versioning.md`, `gotchas.md`, `official-docs.md`.
- `assets/example-plugin/` — the working, tested plugin, demoted to a bundled starter the skill points to.
- `CONTEXT.md` glossary and root `README.md` with install steps.
- `evals/` — validation prompts and their results.

### Changed

- Repositioned the plugin code under `assets/example-plugin/` (preserved via git rename).
- Rewrote `AGENTS.md` for the skill repository.

### Fixed

- Corrected `PostLLMHook` to `bifrost/core v1.6.3` — three-value, `*BifrostError` signature — with
  compile-time signature guards.
- Removed the read-before-write of a context key in `PostLLMHook`.
- Corrected `HTTPTransportPreHook` auth ordering for the v1.6.x line (it runs *before* authentication; the
  move to after-auth is v2.0).
- Untracked the committed 28 MB `.so` and added `.gitignore`; aligned the Makefile build recipe and README.

## [0.1.0] - 2026-10-11

### Added

- Initial release: a native Go `.so` reference plugin for the Bifrost LLM gateway demonstrating the hook
  surface.

[0.2.0]: https://github.com/datamathcode/bifrost-plugin-development/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/datamathcode/bifrost-plugin-development/releases/tag/v0.1.0
