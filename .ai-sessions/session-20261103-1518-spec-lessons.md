# Session Summary: Fold What We Learned Into the Spec

**Date**: 2026-11-03
**Duration**: ~20 minutes
**Conversation Turns**: 1 user turn
**Estimated Cost**: ~$2
**Model**: Claude Opus 5.5

## Key Actions

- Rewrote `spec.md` in place so it reads as current intent rather than the 2026-10-22 plan.
- Corrected drift: the run limit (20 per minute, burst 20, not 10 and 5), `just` from a pinned release tarball, the build condition (checkout changed or an image missing), nftables instead of a `DOCKER-USER` oneshot in G5, the Terraform token source in the component table, 32 smoke tests instead of 16, the playground being on `main`.
- Recorded lessons where the next reader needs them: why isolation is nftables bridge family (Docker 28+ and `br_netfilter`), why the gateway is recreated after a checkout (stale bind-mount inodes), region size availability, `doctl balance` reading zero under the discount, squash merges ruling out stacked PRs, committing state in the same sitting as apply or destroy.
- Added G8 (session history handoff) and G9 (first-test button, folded sample output), the roadmap's shipped list with PR numbers, the actual costs ($17.67 for October, about $2 for the one-day redeploy), and the on-demand lifetime.
- Moved "persisting across runs" out of Non-goals (G8 does that now); persistence across gateway restarts stays a non-goal, and keeping the droplet up between demos is a new one.
- Split multi-sentence lines to follow the one-sentence-per-line rule.

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| update the spec with what we learned | Compared the spec to what shipped, rewrote stale sections, added G8, G9, costs, lessons | PR opened |

## Efficiency Insights

**What went well:**
- Pulling the October invoice line gave the real cost instead of an estimate.

**What could improve:**
- The spec drifted for twelve days; updating it with each PR would have kept it honest.

**Course corrections:**
- none

## Process Improvements

- When a PR changes a decision the spec records, update the spec in the same PR.

## Observations

- Most drift was in defaults and mechanisms (numbers, install method, firewall layer), not in goals.

## Suggested Skills for Next Session

- none
