#!/usr/bin/env bash
# Fallback launcher for Linux builds of Zen Browser Portable — used when the
# Go launcher binary isn't present (build ran without Go available). Does the
# same env/dir setup as launcher/platform_linux.go's runtimeEnvOverrides, but
# skips self-update and AppData-style leftover cleanup — those only live in
# the real launcher.
#
# Lives in Support/ next to this package's root; resolve paths relative to it.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP="$ROOT/App/Zen"
PROFILE="$ROOT/Data/profile"
CACHE_DIR="$ROOT/Data/cache"
TEMP_DIR="$ROOT/Data/temp"

mkdir -p "$PROFILE" "$CACHE_DIR" "$TEMP_DIR"

export TMPDIR="$TEMP_DIR"
export XDG_CACHE_HOME="$CACHE_DIR"
export MOZ_CRASHREPORTER_DISABLE=1

exec "$APP/zen" -profile "$PROFILE" -no-remote "$@"
