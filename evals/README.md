# Evals

Validation prompts for the `bifrost-plugin-development` skill, in the [`skill-creator`](../../) eval format.

`evals.json` defines four prompts:

1. **scaffold-plugin** — scaffold a buildable plugin that adds a request header and logs token usage. Tests the
   required shape (`package main`, package-level symbols, `-buildmode=plugin`) and the host-parity caveat.
2. **which-hook** — choose the right hooks for provider routing and streamed-chunk rewriting, and quote exact
   signatures. Tests the ordering/routing semantics.
3. **preauth-version-trap** — a trap whose naive answer is wrong: set a credential *before auth* while pinned
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

| Iteration | with_skill | without_skill | Delta | Setup |
| --- | --- | --- | --- | --- |
| 1 | 100% | 100% | +0.00 | Source + web available |
| 2 | 18/19 | 18/19 | +0.00 | Added the trap eval; **found a real skill defect** |
| 3 | 4/4 | 4/4 | +0.00 | Targeted non-obvious probe |
| 4 | 100% | 16.7% | **+0.83** | **Air-gapped** — web + module cache denied to both |

**Iteration 2 earned its keep.** The trap surfaced (a) a **false premise in the eval** — expectation 4 claimed
`HTTPTransportPreHook` runs "after auth" on v1.6.3, when it runs *before*; and (b) a **real defect in the
skill** — `SKILL.md` carried that same "after auth" label. Both were corrected.

### Iteration 4 — the discriminating round

Same three knowledge evals (2–4), but both configurations were barred from the web and the Go module cache, so
the only difference between them was the skill:

| Eval | with_skill | without_skill | Delta |
| --- | --- | --- | --- |
| which-hook | 100% | 0% | +100 pp |
| preauth-version-trap | 100% | 0% | +100 pp |
| debug-nonobvious-traps | 100% | 50% | +50 pp |
| **Overall** | **100%** | **16.7%** | **+83 pp** |

The baseline failed for real reasons: it chose `PreLLMHook` for routing (and believed its provider/model
mutations stick), never named `HTTPTransportStreamChunkHook`, invented a
`PreRequestHook(ctx, *http.Request)` signature, and misdiagnosed the nil-read as a context-scope bug rather
than an ordering one. The skill answered all of it from its own references.

## Finding

Iterations 1–3 were non-discriminating because the executor could read the pinned `core v1.6.3` from the Go
module cache and the live docs — so they measured whether a strong agent *can* rebuild the knowledge (it can),
not whether the skill *helps*. Iteration 4 removed those affordances and the skill's contribution became large
and clear: **+83 pp overall**. The skill's value is exactly the non-obvious, hard-won facts — hook selection,
ordering traps, the drop-chunk contract, docs-vs-pin drift, host parity.

The harness default (source + web available) therefore *understates* the value of a knowledge skill like this
one; the air-gapped configuration is the meaningful measurement.
