#!/bin/bash

# Script to push frontend code to GitHub repository
# Usage: ./push-frontend.sh [commit-message]

set -e  # Exit on error

FRONTEND_DIR="frontend"
REPO_URL="https://github.com/Johnson-f/Zaned-frontend.git"

# Get commit message from argument or use default
COMMIT_MSG="${1:-Update frontend code}"

# Change to frontend directory
cd "$(dirname "$0")/$FRONTEND_DIR" || exit 1

echo "📦 Preparing to push frontend code..."
echo "📍 Repository: $REPO_URL"
echo "📝 Commit message: $COMMIT_MSG"
echo ""

# Check if there are any changes
if [ -z "$(git status --porcelain)" ]; then
    echo "✅ No changes to commit. Working directory is clean."
    exit 0
fi

# Show status
echo "📊 Current status:"
git status --short
echo ""

# Add all changes
echo "➕ Staging all changes..."
git add .

# Commit changes
echo "💾 Committing changes..."
git commit -m "$COMMIT_MSG"

# Push to remote
echo "🚀 Pushing to origin/master..."
git push origin master

echo ""
echo "✅ Successfully pushed to $REPO_URL"

