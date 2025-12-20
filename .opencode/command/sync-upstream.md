---
description: Sync upstream into fork main and open PR into main-dev
subtask: true
---

Run the upstream sync workflow end-to-end on the local clone.

REQUIRED TOOLS: bash only.

## Behavior

- Mirror upstream/main into fork main.
- Create/update a branch sync/upstream and open or update a PR into main-dev.
- If merge conflicts occur at any point: STOP.
  - DO NOT edit files.
  - DO NOT stage files.
  - DO NOT commit.
  - Print `git status` and list conflicted files.
  - Ask the user what to do next and wait.

## Steps

1) Pre-flight
- Show current branch and ensure clean working tree.
- Fetch all remotes.

2) Mirror upstream into fork main
- Checkout main
- Reset hard to upstream/main
- Push origin main (force-with-lease)

3) Create integration branch off main-dev
- Checkout main-dev
- Create or reset branch sync/upstream

4) Merge main into sync/upstream
- First try fast-forward only.
- If that fails, try normal merge.
- If conflicts, STOP.

5) Push branch and create/update PR
- Push origin sync/upstream
- Create PR (base=main-dev, head=sync/upstream) if missing
- Optionally enable auto-merge only if user explicitly requests it the same run

## Commands to run

!`git status -sb`

Run these commands now:

- `git fetch --all --prune`
- `git checkout main`
- `git reset --hard upstream/main`
- `git push origin main --force-with-lease`
- `git checkout main-dev`
- `git checkout -B sync/upstream`
- Try merge FF-only: `git merge --ff-only main`
- If FF-only fails: `git merge main --no-edit`

If merge conflicts happen:
- Run: `git status --porcelain=v1`
- STOP and ask user to resolve.

If merge succeeds:
- `git push origin sync/upstream --force-with-lease`
- Create PR if needed:
  - `gh pr create --base main-dev --head sync/upstream --title "chore: sync upstream" --body "Automated upstream sync."`
  - If PR exists, do nothing.
