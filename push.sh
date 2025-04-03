#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <branch-name>"
    exit 1
fi

BRANCH_NAME="$1"
git remote set-url origin git@Rocky9991905:Rocky9991905/UmV2aWV3.git

git checkout -b "$BRANCH_NAME"

