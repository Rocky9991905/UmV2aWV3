#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <branch-name>"
    exit 1
fi


BRANCH_NAME="$1"

git add push.sh second.sh .github/workflows/code-review.yml
git commit -m "add random"
git push --set-upstream origin "$BRANCH_NAME"

