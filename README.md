Idea 1: nodejs server that connects to github webhooks

New idea: local script

>[note: using git worktree might prevent having to stash at all]

commands:
- git-pick <commit(s)> | cherry picks each commit to a new branch
- git-group <commit(s)> | cherry picks all commits to one new branch

scenario:
- branch: `feature/awesomeness` from `main`
- commit: `commit1` rename property
- commit: `commit2` unrelated bugfix
- I want to open a PR for `commit2` asap, but dont want to switch contexts
- git-pick `commit2`
  - make new branch (`commit2-only-branch`) from current tracking (`main`)
  - cherry pick `commit2` (creates `cherrypicked2`)
  - push and PR
  - rebase `commit2-only-branch` onto feature branch
  - drop `commit2` (since we already have `cherrypicked2`)
- commit: `commit3` refactor class
- I want to open a PR for `commit3`, to get quick feedback
- >[note: actually we'd want to open separate PRs for commit 1 & 3 if they're unrelated]
- git-ready `commit3`
  - ```
    you are about to mark these 2 commits as ready:
     - `commit1`: rename property
     - `commit3`: refactor class
    Do you want to continue? (y/n)
    ```
  - make a new branch (`commit1-and-commit2-branch`) from current tracking (`main`)
  - cherry-pick `commit1` (creating `cherrypicked1`)
  - cherry-pick `commit3` (creating `cherrypicked3`)
  - push and pr
