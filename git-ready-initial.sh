#!/bin/bash

set -e

PULL_REMOTE="upstream"
PUSH_REMOTE="origin"

DIRNAME=$(dirname "$(readlink -f "$0")")
source "$DIRNAME/includes/include.sh"
source "$DIRNAME/includes/git-utils.sh"

SOURCE_BRANCH="$1"
if [ -z "$SOURCE_BRANCH" ]; then
  echo "Source branch empty"
  exit 1
fi

COMMIT_SHA=$(git rev-parse --verify HEAD)
COMMIT_MSG=$(git_subject "$COMMIT_SHA")
BRANCH_NAME=$(slugify "$COMMIT_MSG")

if branch_exists "$BRANCH_NAME"; then
  echo "Branch exists"
fi
cat <<OUT
${CYN}
┌─────────────────┈┈
│ ℹ️About to cherry pick commit:
│     └▷ $(git_show_oneline "$COMMIT_SHA" || echo_error "commit not found")${CYN}
│ ℹ️Source branch: ${PRP}$PULL_REMOTE/$SOURCE_BRANCH${CYN}
│     └▷ Last commit: $(git_show_oneline "$PULL_REMOTE/$SOURCE_BRANCH" || echo_error "branch not found")${CYN}
│ ℹ️Branch name: ${PRP}$PUSH_REMOTE/$BRANCH_NAME${CYN} $(branch_exists "$BRANCH_NAME" && echo_error "branch exists")${CYN}
└────────────────────────────────────────────────┈┈
${NC}
OUT
echo -n "${GRN}❓ Continue? (y/N) ${NC}"
read -n1 -r -e response && echo
if [[ "$response" =~ ^[yY]$ ]]; then
  echo_green "▶️ continuing"

  echo_green "▶️ Creating temporary directory"
  TMP_WORKTREE="$(mktemp -d)"

  echo_green "▶️ Fetching $PULL_REMOTE"
  git fetch "$PULL_REMOTE"

  echo_green "▶️ Creating temporary worktree in $TMP_WORKTREE"
  git worktree add --no-track -b "$BRANCH_NAME" "$TMP_WORKTREE" "$PULL_REMOTE/$SOURCE_BRANCH"
  cd "$TMP_WORKTREE"

  echo_green "▶️ Cherry picking $COMMIT_SHA"
  git cherry-pick "$COMMIT_SHA"

  echo_green "▶️ Pushing to $PUSH_REMOTE/$BRANCH_NAME"
  git push -u "$PUSH_REMOTE" "$BRANCH_NAME"

  echo_green "▶️ Creating pull request"
  gh pr create --fill --assignee "@me" --base "$SOURCE_BRANCH"

  echo_green "▶️ Cleaning up temporary worktree"
  git worktree remove "$TMP_WORKTREE"
else
  echo_red "❌ cancelling"
  exit 0
fi
