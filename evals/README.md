# Evals

Validation prompts for the `bifrost-plugin-development` skill, in the [`skill-creator`](../../) eval format.

`evals.json` defines two prompts:

1. **scaffold-plugin** — an agent is asked to scaffold a buildable Bifrost plugin that adds a request header
   and logs token usage. Tests whether the skill produces the required shape (`package main`, package-level
   symbols, `-buildmode=plugin`) without the agent having to rediscover the API.
2. **which-hook** — an agent must choose the right hooks for provider routing and streamed-chunk rewriting,
   and quote the exact signatures. Tests the ordering/routing semantics the skill documents.

## Running them

Follow the `skill-creator` workflow: spawn a with-skill and a without-skill run per eval, grade each against the
`expectations` in `evals.json`, then aggregate:

```bash
# from a checkout of the skill-creator skill
python3 -m scripts.aggregate_benchmark <skill>-workspace/iteration-1 --skill-name bifrost-plugin-development
python3 eval-viewer/generate_review.py <skill>-workspace/iteration-1 --skill-name bifrost-plugin-development \
  --benchmark <skill>-workspace/iteration-1/benchmark.json --static <skill>-workspace/iteration-1/review.html
```

Run outputs live outside the repo, in `<skill>-workspace/`, not here — only the prompt set is version-controlled.

## Known result (iteration 1)

Both configurations passed **100%** (delta +0.00): a capable agent with access to the pinned
`core v1.6.3` module cache and the live docs reconstructs the same answers without the skill. The current
expectations therefore do **not** discriminate. The grader suggested sharper, skill-specific probes for the
next iteration — docs-vs-pin drift, HTTP-transport-only caveats, Go-version/host parity, and drop-chunk
`(nil, nil)` semantics — and a harder prompt whose naive answer is wrong (e.g. "add an auth header *before*
Bifrost authenticates", where v1.6.3 has no `HTTPTransportPreAuthHook`).
