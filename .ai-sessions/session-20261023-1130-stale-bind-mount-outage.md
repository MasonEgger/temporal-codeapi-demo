# Session Summary: Stale Bind Mount Outage After the Move to main

**Date**: 2026-10-23
**Duration**: ~30 minutes
**Conversation Turns**: 1 user turn
**Estimated Cost**: ~$4
**Model**: Claude Fable 5.1

## Key Actions

- Mason reported "Error reading directory" at the site root right after the droplet moved to main. The gateway log said `readdirent /srv/site/.: no such file or directory`: the container's bind mounts of `web/` and `examples/` pointed at directories the deploy had deleted and recreated. Every page 404'd and the root 500'd while `/v1/exec` kept working, so the smoke suite had passed.
- Root cause from the checkout's reflog: the Ansible git module switched to the local `main` branch, still at the initial commit (no `web/`), then reset to `origin/main`, so the directories were removed and recreated with new inodes.
- Fixed the outage by recreating the gateway container with `docker compose up --force-recreate`; all pages back within a minute of the report.
- Permanent fix: the "Run the gateway" task now passes `recreate: always` whenever the checkout task changed. Validated with a redeploy (gateway recreated, pages fine) and a second no-op redeploy (0 changed).
- Added `TestSite` to the smoke suite: every page, `tutorial.js`, `tutorial.css`, each language's tutorial Markdown, and one example file per language must return 200 through the gateway. Go's FileServer redirects `/index.html` to `/`, so the root path stands in for it.

## Prompt Inventory

| Prompt/Command | Action Taken | Outcome |
|---|---|---|
| Immediately got "error reading directory" on that url | Read the gateway log, confirmed stale mounts, recreated the container, fixed Ansible, added page tests | Site back; redeploys recreate the gateway after a checkout change; 11 new tests |

## Efficiency Insights

**What went well:**
- The gateway's own log line named the failure; the reflog on the droplet explained why the directory had been recreated.

**What could improve:**
- The smoke suite only exercised `/v1/exec`; a page fetch would have caught the outage before Mason did.

**Course corrections:**
- `index.html` dropped from the page list after the first run: the file server answers it with a 301 to `/`.

## Process Improvements

- A smoke suite must touch every path a user hits, not only the one the feature is about.
- When a long-lived container bind-mounts a directory that a deploy rewrites, recreate the container as part of the deploy.

## Observations

- Ansible's git module reaches the target ref via the local branch of that name, so moving to a branch whose local copy is stale passes through an old tree first.

## Suggested Skills for Next Session

- `python:python`: if the smoke suite grows further.
