#!/bin/bash


if [ ! -f "package.json" ]; then
    echo "Error: package.json not found in current directory"
    exit 1
fi

VERSION=$(node -p "require('./package.json').version")

if [ -z "$VERSION" ]; then
    echo "Error: Could not read version from package.json"
    exit 1
fi

echo "Version: $VERSION"

git add package.json

if git diff --cached --quiet; then
    echo "No changes in package.json to commit"
    exit 0
fi

git commit -m "deploy v$VERSION"

echo "Successfully committed version $VERSION"

git push origin main

echo "Deployed version $VERSION"