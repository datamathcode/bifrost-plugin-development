# Repo is a skill, not a plugin

This repository exists to be an agent skill that teaches an agent how to build a Bifrost gateway plugin, but
its first implementation misread the assignment and produced a reference plugin instead of a skill. We
repurposed the repo so `SKILL.md` is the deliverable and demoted the (now correct and tested) plugin code to a
bundled starter under `assets/example-plugin/` rather than deleting it. We teach only the native Go `.so`
delivery mechanism because the official Bifrost docs now mark WASM plugins deprecated, and we pin the skill's
verified snapshot to `core v1.6.3` while citing the live docs so an agent re-verifies before finishing.
