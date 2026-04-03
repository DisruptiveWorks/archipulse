---
name: Git workflow — always use branches and PRs
description: Never push directly to main; always create a feature branch and open a PR
type: feedback
---

Always work on a feature branch and open a PR. Never commit or push directly to main.

**Why:** User explicitly requested this after direct pushes to main happened repeatedly.

**How to apply:**
- Before starting any work, create a branch: `git checkout -b feat/...` or `fix/...`
- Commit to the branch, push with `-u origin <branch>`
- Open a PR with `gh pr create`
- Never run `git push` targeting main directly
