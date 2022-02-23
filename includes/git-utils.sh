#!/usr/bin/env bash

set -e

function branch_exists() {
  git show-ref --verify --quiet "refs/heads/$1"
}

function branch_exists_on_remote() {
  git show-ref --verify --quiet "refs/remotes/$1"
}

function git_show_oneline() {
  git show -q "$1" \
    --date=human \
    --color \
    --pretty=format:"%C(yellow)%h %C(blue)%ad %C(green)%aN %C(reset)%s" 2>/dev/null
}

function git_subject() {
  git show -s --format=%s "$@"
}
