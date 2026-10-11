# Evals

Validation prompts for the `bifrost-plugin-development` skill, in the [`skill-creator`](../../) eval format.

`evals.json` defines three prompts:

1. **scaffold-plugin** — an agent is asked to scaffold a buildable Bifrost plugin that adds a request header
   and logs token usage. Tests whether the skill produces the required shape (`package main`, package-level
   symbols, `-buildmode=plugin`) and the host-parity caveat.
2. **which-hook** — an agent must choose the right hooks for provider routing and streamed-chunk rewriting,
   and quote the exact signatures. Tests the ordering/routing semantics the skill documents.
3. **preauth-version-trap** — a trap prompt whose naive answer is wrong: the user asks to set a credential
   *before Bifrost authenticates* while pinned to `core v1.6.3`. The live docs describe a newer line, so an
   agent that trusts them reaches for `HTTPTransportPreAuthHook` (a v2.0+ symbol). The correct answer for the
   pin is `HTTPTransportPreHook`, which runs before auth on v1.6.x. Tests docs-vs-pin discipline.

## Running them

Follow the `skill-creator` workflow: spawn a with-skill and a without-skill run per eval, grade each against
the `expectations` in `evals.json`, then aggregate:

```bash
# from a checkout of the skill-creator skill
python3 -m scripts.aggregate_benchmark <skill>-workspace/iteration-N --skill-name bifrost-plugin-development
python3 eval-viewer/generate_review.py <skill>-workspace/iteration-N --skill-name bifrost-plugin-development \
  --benchmark <skill>-workspace/iteration-N/benchmark.json --static <skill>-workspace/iteration-N/review.html
```

Run outputs live outside the repo, in `<skill>-workspace/`, not here — only the prompt set is version-controlled.

## Results

**Iteration 1** — with-skill 100%, without-skill 100% (delta +0.00). Non-discriminating.

**Iteration 2** — added the trap eval and host-parity/transport-only expectations. Result: with-skill
**18/19**, without-skill **18/19** (delta +0.00). Still non-discriminating — a capable agent with the pinned
module cache reconstructs the answers unaided.

**But iteration 2 earned its keep.** The trap surfaced two things:

- A **false premise in the eval itself**: expectation 4 of `preauth-version-trap` claimed
  `HTTPTransportPreHook` runs "after auth" on v1.6.3. The runs (and the skill's own `versioning.md`) showed the
  opposite — v1.6.3 runs it *before* auth; the "after auth" move is v2.0. Expectation corrected above.
- A **real defect in the skill**: `SKILL.md` labelled `HTTPTransportPreHook` as "after auth" for the pinned
  version. The with-skill run explicitly flagged that the skill's table was wrong and overrode it from the
  source. The skill has since been corrected (`SKILL.md`, `references/hooks-reference.md`,
  `references/gotchas.md`).

## Next iteration

The expectations still do not discriminate. To separate the skill from a capable baseline, the next round
should target only skill-specific, non-obvious outcomes and run a prompt whose naive answer is wrong. Candidates
already collected: the drop-chunk `(nil, nil)` semantics, the `HTTPTransportPostHook`-after-`PostLLMHook`
context-visibility trap, and Go-module-version pinning (docs say `1.27.0`/`1.26.1`; the rule is host parity).
