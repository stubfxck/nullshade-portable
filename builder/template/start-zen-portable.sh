#!/usr/bin/env bash
# Portable launcher for Linux builds of Zen Browser
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP="$ROOT/App/Zen"
PROFILE="$ROOT/Data/profile"
TEMP_DIR="$ROOT/Data/temp"

mkdir -p "$PROFILE" "$TEMP_DIR"

export TMPDIR="$TEMP_DIR"

exec "$APP/zen" -profile "$PROFILE" -no-remote "$@"
