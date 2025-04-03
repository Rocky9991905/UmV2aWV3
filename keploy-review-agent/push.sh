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

# Find the last comment and append gibberish
if grep -q "//" go.mod; then
    # Get the last comment
    LAST_COMMENT=$(grep "//" go.mod | tail -n1)
    
    # Generate gibberish
    GIBBERISH=$(tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 10)
    
    # Replace the last occurrence of the comment with itself + gibberish
    sed -i "\$s|$LAST_COMMENT|$LAST_COMMENT $GIBBERISH|" go.mod
else
    # If no comment exists, just append a new one
    echo "// $(tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 10)" >> go.mod
fi

# Add, commit, and push changes
git add go.mod push.sh
git commit -m "Modify last comment with gibberish"
git push --set-upstream origin "$BRANCH_NAME"
