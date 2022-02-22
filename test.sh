#!/bin/bash

set -e

### SETUP
DIR="$(mktemp -d)"
cd "$DIR"
git init

### INITIAL ###
cat << CONTENT > file.txt
1 * 2 = ?
2 * 2 = ?
3 * 2 = ?
CONTENT
git add file.txt
git commit -a -m "initial commit"


### BRANCH ###
git switch -c branch-1
# commit A
cat << CONTENT > file.txt
1 * 2 = 2
2 * 2 = 3
3 * 2 = ?
CONTENT
git commit -a -m "change A"
COMMIT_A_HASH=$(git rev-parse HEAD)

# commit B
cat << CONTENT > file.txt
1 * 2 = 2
2 * 2 = 3
3 * 2 = 6
CONTENT
git commit -a -m "change B"


### create PR branch ###
git switch -c pr-a "$COMMIT_A_HASH"


### trunk change ###
git switch trunk
cat << CONTENT > otherfile.txt
hello
CONTENT
git add otherfile.txt
git commit -a -m "added otherfile.txt"


### RESULT ###
cat file.txt
git log --graph --oneline --all
rm -rf "$DIR"
