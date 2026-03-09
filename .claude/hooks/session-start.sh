#!/bin/bash
set -euo pipefail

# Only run in remote (Claude Code on the web) environments
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

RUBY_VERSION="4.0.1"

# Install Ruby if not already available
if ! rbenv versions --bare | grep -qx "$RUBY_VERSION"; then
  echo "Installing Ruby $RUBY_VERSION..."
  rbenv install "$RUBY_VERSION"
  rbenv rehash
fi

# Ensure the correct Ruby is active
eval "$(rbenv init -)"

# Install gems
cd "$CLAUDE_PROJECT_DIR"
bundle check || bundle install
