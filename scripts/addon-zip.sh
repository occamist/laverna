#!/bin/bash

# Build script for Laverna Anki Addon
# Creates a .ankiaddon package for distribution on AnkiWeb

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ADDON_DIR="$PROJECT_ROOT/addon"

cd "$ADDON_DIR"

echo "Working directory: $ADDON_DIR"

find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
find . -type d -name ".mypy_cache" -exec rm -rf {} + 2>/dev/null || true
find . -type d -name ".ruff_cache" -exec rm -rf {} + 2>/dev/null || true
find . -type f -name "*.pyc" -delete 2>/dev/null || true
rm -f addon.log 2>/dev/null || true

7z a -tzip "$PROJECT_ROOT/laverna-addon.ankiaddon" __init__.py config.json README.md

echo "Done! Package created at: $PROJECT_ROOT/laverna-addon.ankiaddon"
