# Bifrost Plugin Development

The context for this repository, which is an **agent skill**: it teaches an agent how to build a Bifrost
gateway plugin. The repository is not itself a plugin — it contains one as a bundled example.

## Language

**Bifrost plugin**:
A Go shared object (`.so`) that the Bifrost gateway loads at runtime to intercept and modify requests and
responses. It is the artifact an agent builds; it lives under `assets/example-plugin/` here purely as a
starter.
_Avoid_: middleware, extension, add-on

**Hook**:
A named function a plugin exports that Bifrost calls at a specific point in the request lifecycle.
_Avoid_: callback, interceptor, middleware

**Exported symbol**:
A package-level function in the plugin's `package main` that the loader resolves **by name** via
`plugin.Lookup`, as opposed to being an interface method.
_Avoid_: interface, method set

**Delivery mechanism**:
How plugin code reaches Bifrost — native Go `.so` (supported), in-process Go SDK (supported, no `.so`), WASM
(deprecated), webhooks (planned).
_Avoid_: plugin type

**Agent skill**:
The artifact this repository *is*: a `SKILL.md`-rooted bundle that teaches an agent to build Bifrost plugins.
_Avoid_: plugin, package, module
