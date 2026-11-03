# Session Summary: One-Day Redeploy of the Playground

**Date**: 2026-11-02
**Duration**: ~25 minutes
**Conversation Turns**: 1 user turn
**Estimated Cost**: ~$3
**Model**: Claude Fable 5.1

## Key Actions

- Mason asked to redeploy the playground for a day. Found the first droplet gone (torn down 2026-10-30), no DNS record, and the local `state.sops.json` modified but uncommitted with an empty resource list: the destroy had run from this machine but its state was never committed, so origin still described the old droplet.
- `just infra-apply` on the local (correct) state created a new droplet (137.184.113.37), firewall, A record, and inventory. `just deploy` from scratch: 42 ok, 33 changed, 0 failed, images built. Let's Encrypt cert issued. 32 smoke tests passed, including the multi-run session check.
- Added a p1 Todoist task for 2026-11-03 to tear it down (deadline inside three days per the priority rules), 30 minutes, in the Temporal project.
- Committed the state (both the earlier teardown and today's apply) with the deploy README's new dates on `chore/redeploy-2026-11-02`.

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| Ok, let's redeploy this for a day | Checked state, applied, deployed, smoke-tested, Todoist teardown task, state commit | Live at https://temporal.egger.codes; PR for the state |

## Efficiency Insights

**What went well:**
- The `_tf` wrapper applied straight from the uncommitted local state; no reconciliation needed because the local file was the truth.

**What could improve:**
- The teardown's state change should have been committed on 2026-10-30; the Todoist task says so, but it was easy to skip. The state file is the only copy.

**Course corrections:**
- none

## Process Improvements

- After `just infra-destroy`, commit the state in the same sitting; an uncommitted empty state and a committed non-empty one disagree about whether money is being spent.

## Observations

- A fresh droplet from apply to passing smoke suite is about 20 minutes, most of it image builds.

## Suggested Skills for Next Session

- none
