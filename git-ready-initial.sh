#!/bin/bash

PULL_REMOTE="upstream"
PUSH_REMOTE="origin"

NC=$(tput sgr0)
RED=$(tput setaf 1)
GRN=$(tput setaf 2)
CYN=$(tput setaf 6)
PRP=$(tput setaf 5)
ERR=$(tput setaf 9)
BOLD=$(tput bold)

set -e

function echo_green() {
  echo "${GRN}$*${NC}"
}
function echo_red() {
  echo "${RED}$*${NC}"
}
function echo_cyan() {
  echo "${CYN}$*${NC}"
}
function echo_error() {
  echo "${ERR}${BOLD}❗❗ ${*^^}${NC}"
}

function slugify() {
  echo "$1" | iconv -t ascii//TRANSLIT | sed -r s/[~^]+//g | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr "[:upper:]" "[:lower:]"
}

function branch_exists() {
  git show-ref --verify --quiet "refs/heads/$1"
}
function branch_exists_on_remote() {
  git show-ref --verify --quiet "refs/remotes/$1"
}
function git_show() {
  git show -q "$1" \
    --date=human \
    --color \
    --pretty=format:"%C(yellow)%h %C(blue)%ad %C(green)%aN %C(reset)%s" 2>/dev/null
}

SOURCE_BRANCH="$1"
if [ -z "$SOURCE_BRANCH" ]; then
  echo "Source branch empty"
  exit 1
fi

COMMIT_SHA=$(git rev-parse --verify HEAD)
COMMIT_MSG=$(git show -s --format=%s "$COMMIT_SHA")
BRANCH_NAME=$(slugify "$COMMIT_MSG")

if branch_exists "$BRANCH_NAME"; then
  echo "Branch exists"
fi
cat << OUT
${CYN}
┌─────────────────┈┈
│ ℹ️About to cherry pick commit:
│     └▷ $(git_show "$COMMIT_SHA" || echo_error "commit not found")${CYN}
│ ℹ️Source branch: ${PRP}$PULL_REMOTE/$SOURCE_BRANCH${CYN}
│     └▷ Last commit: $(git_show "$PULL_REMOTE/$SOURCE_BRANCH" || echo_error "branch not found")${CYN}
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
