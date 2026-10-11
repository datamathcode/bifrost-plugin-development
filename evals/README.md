# Evals

Validation prompts for the `bifrost-plugin-development` skill, in the [`skill-creator`](../../) eval format.

`evals.json` defines four prompts:

1. **scaffold-plugin** — scaffold a buildable plugin that adds a request header and logs token usage. Tests the
   required shape (`package main`, package-level symbols, `-buildmode=plugin`) and the host-parity caveat.
2. **which-hook** — choose the right hooks for provider routing and streamed-chunk rewriting, and quote exact
   signatures. Tests the ordering/routing semantics.
3. **pre auth-version-trap** — a trap whose naive answer is wrong: set a credential *before auth* while pinned
   to `core v1.6.3`. The live docs describe a newer line, so an agent that trusts them reaches for
   `HTTPTransportPreAuthHook` (v2.0+); the correct answer for the pin is `HTTPTransportPreHook`.
4. **debug-nonobvious-traps** — three non-obvious outcomes at once: the `HTTPTransportPostHook`-after-
   `PostLLMHook` context trap, the `(nil, nil)` drop-chunk signal, and Go-version host parity.

## Running them

```bash
# from a checkout of the skill-creator skill
python3 -m scripts.aggregate_benchmark <skill>-workspace/iteration-N --skill-name bifrost-plugin-development
python3 eval-viewer/generate_review.py <skill>-workspace/iteration-N --skill-name bifrost-plugin-development \
  --benchmark <skill>-workspace/iteration-N/benchmark.json --static <skill>-workspace/iteration-N/review.html
```

Run outputs live outside the repo, in `<skill>-workspace/`, not here — only the prompt set is version-controlled.

## Results

| Iteration | with_skill | without_skill | Delta | Notes |
| --- | --- | --- | --- | --- |
| 1 | 100% | 100% | +0.00 | Two evals, first pass |
| 2 | 18/19 | 18/19 | +0.00 | Added the trap eval; **found a real skill defect** |
| 3 | 4/4 | 4/4 | +0.00 | Targeted non-obvious probe (`debug-nonobvious-traps`) |

Every round was **non-discriminating**.

**Iteration 2 earned its keep anyway.** The trap surfaced (a) a **false premise in the eval** — expectation 4
claimed `HTTPTransportPreHook` runs "after auth" on v1.6.3, when it runs *before*; and (b) a **real defect in
the skill** — `SKILL.md` carried that same "after auth" label. Both were corrected.

## Finding: these evals measure the harness, not the skill

Across three rounds, every expectation — the trap and the targeted non-obvious probes included — passed in
both configurations. The cause is the test setup, not the skill: the executor has the pinned `core v1.6.3` in
the Go module cache and the live docs on the web, so a capable agent reconstructs the answers unaided. The
evals therefore measure whether a strong agent *can* rebuild the knowledge (yes), not whether the skill *helps*.

To measure lift, run the baseline **air-gapped** — no network and an empty module cache — so the agent cannot
read `schemas/plugin.go` or the docs. Only then do the skill's baked-in facts (ordering traps, drop-chunk
semantics, host parity, docs-vs-pin drift) become the difference. That is the proposed iteration 4.
