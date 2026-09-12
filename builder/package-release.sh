#!/usr/bin/env bash
# Zen Browser Portable — Linux (Debian/Ubuntu/Mint и т.д.).
# Скачивает официальный релиз Zen Browser (zen.linux-$ARCH.tar.xz),
# распаковывает и собирает portable-папку + tar.gz. Ничего не устанавливает
# в систему. Зеркалит builder/package-release.ps1 — см. его комментарии
# для контекста по каждому шагу, тут только то, что отличается для Linux.
#
# Нужны: curl, jq, tar (с поддержкой xz), go (для сборки лаунчера — если
# нет, соберётся пакет без него, останется только start-zen-portable.sh).
#
# Использование:
#   ./package-release.sh                    # последняя версия, x86_64
#   ./package-release.sh --version 1.22b
#   ./package-release.sh --arch aarch64
set -euo pipefail

VERSION="latest"
ARCH="x86_64"
OUTPUT_DIR="output"
WORK_DIR="work"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --arch) ARCH="$2"; shift 2 ;;
    --output-dir) OUTPUT_DIR="$2"; shift 2 ;;
    --work-dir) WORK_DIR="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

case "$ARCH" in
  x86_64|aarch64) ;;
  *) echo "Unsupported --arch: $ARCH (expected x86_64 or aarch64)" >&2; exit 1 ;;
esac

ROOT="$(pwd)"
WORK="$ROOT/$WORK_DIR"
OUT="$ROOT/$OUTPUT_DIR"
mkdir -p "$WORK" "$OUT"

# --- 1. Определить версию через GitHub API -----------------------------------
echo "== Resolving Zen Browser version..."
if [[ "$VERSION" == "latest" ]]; then
  RELEASE_JSON=$(curl -sL "https://api.github.com/repos/zen-browser/desktop/releases/latest")
else
  RELEASE_JSON=$(curl -sL "https://api.github.com/repos/zen-browser/desktop/releases/tags/$VERSION")
fi
TAG=$(echo "$RELEASE_JSON" | jq -r '.tag_name')
if [[ -z "$TAG" || "$TAG" == "null" ]]; then
  echo "Couldn't resolve Zen version (tag=$VERSION)" >&2
  exit 1
fi
echo "   Version: $TAG"

# --- 2. Скачать официальный tar.xz --------------------------------------------
ASSET_NAME="zen.linux-$ARCH.tar.xz"
ASSET_URL=$(echo "$RELEASE_JSON" | jq -r --arg name "$ASSET_NAME" '.assets[] | select(.name == $name) | .browser_download_url')
if [[ -z "$ASSET_URL" || "$ASSET_URL" == "null" ]]; then
  echo "Asset $ASSET_NAME not found in release $TAG" >&2
  exit 1
fi
TARBALL="$WORK/$ASSET_NAME"
echo "== Downloading $ASSET_NAME..."
curl -sL "$ASSET_URL" -o "$TARBALL"

# --- 3. Распаковать официальный тарбол ----------------------------------------
EXTRACT_DIR="$WORK/extracted"
rm -rf "$EXTRACT_DIR"
mkdir -p "$EXTRACT_DIR"
echo "== Extracting official build..."
tar -xJf "$TARBALL" -C "$EXTRACT_DIR"
# Официальный тарбол разворачивается в один каталог верхнего уровня "zen/"
# (проверено на реальном релизе — содержит zen/zen-bin, zen/omni.ja и т.д.).
CORE_DIR="$EXTRACT_DIR/zen"
if [[ ! -f "$CORE_DIR/zen" ]]; then
  echo "zen binary not found inside extracted tarball (layout changed upstream?)" >&2
  exit 1
fi

# --- 4. Собрать portable-структуру --------------------------------------------
PKG_NAME="ZenBrowserPortable-$TAG-linux-$ARCH"
PKG="$WORK/$PKG_NAME"
rm -rf "$PKG"
echo "== Building portable package..."
mkdir -p "$PKG/App/Zen" "$PKG/Data/profile" "$PKG/Data/cache" "$PKG/Data/temp" "$PKG/Support"

cp -a "$CORE_DIR/." "$PKG/App/Zen/"

# Portable-переопределения — тот же смысл, что и в portable.js для Windows
# (см. package-release.ps1), пути и часть механизмов платформо-специфичны
# (нет реестра/AppData — просто нечего перекрывать для этого на Linux).
cat > "$PKG/App/Zen/defaults/pref/portable.js" <<'EOF'
// Portable build overrides
pref("app.update.service.enabled", false);
pref("app.update.background.scheduling.enabled", false);
pref("app.update.auto", false);
pref("browser.shell.checkDefaultBrowser", false);
pref("toolkit.telemetry.enabled", false);
pref("toolkit.telemetry.unified", false);
pref("datareporting.healthreport.uploadEnabled", false);
pref("datareporting.policy.dataSubmissionEnabled", false);
pref("app.normandy.enabled", false);
pref("browser.crashReports.unsubmittedCheck.autoSubmit2", false);
pref("zen.portable.mode", true);
EOF

mkdir -p "$PKG/App/Zen/distribution"
cat > "$PKG/App/Zen/distribution/policies.json" <<'EOF'
{
  "policies": {
    "DisableAppUpdate": true,
    "DisableTelemetry": true,
    "DontCheckDefaultBrowser": true,
    "DisableFirefoxStudies": true
  }
}
EOF

TEMPLATE_DIR="$(dirname "$0")/template"
SUPPORT_DIR="$PKG/Support"
cp "$TEMPLATE_DIR/start-zen-portable.sh" "$SUPPORT_DIR/"
chmod +x "$SUPPORT_DIR/start-zen-portable.sh"
cp "$TEMPLATE_DIR/README-PORTABLE.md" "$PKG/"
cp "$TEMPLATE_DIR/README-PORTABLE_RU.md" "$PKG/"

BUILT_AT=$(date -u +"%Y-%m-%d %H:%M:%S")
cat > "$SUPPORT_DIR/VERSION.txt" <<EOF
Zen Browser Portable
Version: $TAG
Arch: linux-$ARCH
Mode: release-repackage
Built: $BUILT_AT
Source: https://github.com/zen-browser/desktop/releases/tag/$TAG
EOF

BUILT_AT_ISO=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
cat > "$SUPPORT_DIR/version.json" <<EOF
{
  "portableTag": "portable-$TAG",
  "zenTag": "$TAG",
  "arch": "linux-$ARCH",
  "builtAt": "$BUILT_AT_ISO"
}
EOF

# --- 5. Лаунчер (если доступен Go; иначе останется только .sh) ---------------
LAUNCHER_SRC="$(dirname "$0")/../launcher"
LAUNCHER_BIN="$PKG/ZenBrowserPortable"
if command -v go >/dev/null 2>&1 && [[ -f "$LAUNCHER_SRC/main.go" ]]; then
  echo "== Building ZenBrowserPortable launcher..."
  GOARCH="amd64"
  [[ "$ARCH" == "aarch64" ]] && GOARCH="arm64"
  ( cd "$LAUNCHER_SRC" && GOOS=linux GOARCH="$GOARCH" CGO_ENABLED=0 go build -ldflags "-s -w" -o "$LAUNCHER_BIN" . )
  chmod +x "$LAUNCHER_BIN"
  if [[ ! -f "$LAUNCHER_BIN" ]]; then
    echo "Launcher build failed" >&2
    exit 1
  fi
else
  echo "== Go not found: skipping launcher, start-zen-portable.sh will be used"
fi

# --- 6. Права на исполняемые файлы (обязательно ДО архивации) ----------------
# tar сохраняет unix-права как есть — если не проставить +x здесь, zen после
# распаковки архива у пользователя не запустится вообще без ручного chmod.
chmod +x "$PKG/App/Zen/zen" "$PKG/App/Zen/zen-bin" 2>/dev/null || true

# --- 7. Проверки пакета (smoke tests) -----------------------------------------
echo "== Validating package..."
REQUIRED=(
  "App/Zen/zen"
  "App/Zen/omni.ja"
  "Support/start-zen-portable.sh"
  "Support/VERSION.txt"
  "Support/version.json"
  "README-PORTABLE.md"
  "README-PORTABLE_RU.md"
)
for r in "${REQUIRED[@]}"; do
  if [[ ! -e "$PKG/$r" ]]; then
    echo "VALIDATION FAILED: missing $r" >&2
    exit 1
  fi
done
ZEN_SIZE=$(stat -c%s "$PKG/App/Zen/zen")
if [[ "$ZEN_SIZE" -lt 300000 ]]; then
  echo "VALIDATION FAILED: zen binary suspiciously small ($ZEN_SIZE bytes)" >&2
  exit 1
fi
if [[ ! -x "$PKG/App/Zen/zen" ]]; then
  echo "VALIDATION FAILED: zen binary is not executable after packaging" >&2
  exit 1
fi
APP_SIZE=$(du -sb "$PKG/App/Zen" | cut -f1)
echo "   OK: zen present and executable, App/Zen = $((APP_SIZE / 1024 / 1024)) MB"

# --- 8. tar.gz ------------------------------------------------------------------
TARBALL_OUT="$OUT/$PKG_NAME.tar.gz"
rm -f "$TARBALL_OUT"
echo "== Creating $TARBALL_OUT ..."
tar -czf "$TARBALL_OUT" -C "$WORK" "$PKG_NAME"

echo
echo "DONE:"
echo "  Folder: $PKG"
echo "  Archive: $TARBALL_OUT"
