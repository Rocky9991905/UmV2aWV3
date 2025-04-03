#!/bin/bash

# Check if a branch name is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <branch-name>"
    exit 1
fi

BRANCH_NAME="$1"
git remote set-url origin git@Per0x1de-1337:Per0x1de-1337/UmV2aWV3.git

# Create and switch to the new branch
git checkout -b "$BRANCH_NAME"

# Modify nothing.go with random content
echo "$(cat a.txt) $(tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 10)" > a.txt

# Add, commit, and push changes
git add a.txt push.sh
git commit -m "add random"
git push --set-upstream origin "$BRANCH_NAME"
