#!/usr/bin/env bash
# Maintainer script: fetches the latest mpt-crypto from the XRPLF Conan remote
# and copies static libraries + headers into the vendored deps directory.
#
# Prerequisites:
#   - conan >= 2.0 (pip install conan)
#   - xrplf remote: conan remote add --index 0 xrplf https://conan.ripplex.io
#
# Usage:
#   bash confidential/deps/update.sh                        # fetch latest, current platform
#   bash confidential/deps/update.sh --version 0.2.0-rc1    # fetch specific version
#   bash confidential/deps/update.sh --platform linux-arm64 # target specific platform
#   bash confidential/deps/update.sh --force                # ignore VERSION file, re-fetch
#
# Supported platforms: linux-amd64, linux-arm64, darwin-arm64, darwin-amd64
#
# For multi-platform builds, run on each platform natively (e.g., CI matrix)
# or use Conan cross-compilation profiles.

set -euo pipefail

# --- Argument parsing ---

VERSION=""
PLATFORM=""
FORCE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)   VERSION="$2"; shift 2 ;;
        --platform)  PLATFORM="$2"; shift 2 ;;
        --force)     FORCE=true; shift ;;
        -h|--help)
            sed -n '2,/^$/s/^# \?//p' "$0"
            exit 0
            ;;
        *)
            # Bare arg = version (backward compat)
            VERSION="$1"; shift ;;
    esac
done

# --- Preflight checks ---

if ! command -v conan &>/dev/null; then
    echo "ERROR: conan is not installed."
    echo ""
    echo "Install it with:"
    echo "  pip install conan"
    echo ""
    echo "Then set up a default profile:"
    echo "  conan profile detect"
    exit 1
fi

if ! conan remote list 2>/dev/null | grep -q "xrplf"; then
    echo "ERROR: the 'xrplf' Conan remote is not configured."
    echo ""
    echo "Add it with:"
    echo "  conan remote add --index 0 xrplf https://conan.ripplex.io"
    exit 1
fi

# --- Configuration ---

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION_FILE="$SCRIPT_DIR/VERSION"

# Detect or validate platform
if [ -z "$PLATFORM" ]; then
    DETECTED_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    DETECTED_ARCH=$(uname -m)
    case "$DETECTED_ARCH" in
        x86_64)        DETECTED_ARCH="amd64" ;;
        aarch64|arm64) DETECTED_ARCH="arm64" ;;
    esac
    PLATFORM="${DETECTED_OS}-${DETECTED_ARCH}"
fi

# Map platform to Conan settings
case "$PLATFORM" in
    linux-amd64)  CONAN_OS="Linux";  CONAN_ARCH="x86_64" ;;
    linux-arm64)  CONAN_OS="Linux";  CONAN_ARCH="armv8" ;;
    darwin-amd64) CONAN_OS="Macos";  CONAN_ARCH="x86_64" ;;
    darwin-arm64) CONAN_OS="Macos";  CONAN_ARCH="armv8" ;;
    *)
        echo "ERROR: unsupported platform '$PLATFORM'"
        echo "Supported: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64"
        exit 1
        ;;
esac

LIBS_DIR="$SCRIPT_DIR/libs/$PLATFORM"
INCLUDE_DIR="$SCRIPT_DIR/include"

# --- Resolve version ---

if [ -z "$VERSION" ]; then
    echo "==> Querying latest mpt-crypto version from xrplf remote..."
    VERSION=$(conan list 'mpt-crypto/*' -r xrplf 2>/dev/null \
        | grep 'mpt-crypto/' \
        | sed 's/.*mpt-crypto\///' \
        | sort -V \
        | tail -1)
    if [ -z "$VERSION" ]; then
        echo "ERROR: could not find any mpt-crypto versions on the xrplf remote"
        exit 1
    fi
    echo "==> Latest version: $VERSION"
else
    echo "==> Using specified version: $VERSION"
fi

# --- Check if already up to date ---

if [ "$FORCE" = false ]; then
    CURRENT_VERSION=""
    if [ -f "$VERSION_FILE" ]; then
        CURRENT_VERSION=$(cat "$VERSION_FILE")
    fi

    if [ "$CURRENT_VERSION" = "$VERSION" ] && [ -d "$LIBS_DIR" ]; then
        echo "==> Already at mpt-crypto/$VERSION for $PLATFORM. Nothing to do."
        echo "    Use --force to re-fetch."
        exit 0
    fi

    if [ -n "$CURRENT_VERSION" ] && [ "$CURRENT_VERSION" != "$VERSION" ]; then
        echo "==> Updating: $CURRENT_VERSION -> $VERSION"
    else
        echo "==> Installing mpt-crypto/$VERSION for $PLATFORM"
    fi
fi

# --- Fetch via Conan ---

WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT

cat > "$WORK_DIR/conanfile.txt" <<EOF
[requires]
mpt-crypto/$VERSION

[generators]
CMakeDeps
CMakeToolchain
EOF

cd "$WORK_DIR"
echo "==> Running conan install for $PLATFORM (this may build from source on first run)..."
conan install . --build=missing \
    -o "*:shared=False" \
    -o "*:fPIC=True" \
    -s "os=$CONAN_OS" \
    -s "arch=$CONAN_ARCH"

# --- Copy libraries ---

echo "==> Copying static libraries to $LIBS_DIR..."
mkdir -p "$LIBS_DIR"

copy_lib() {
    local name="$1"
    local src
    # Use stat for sorting by modification time (portable across Linux and macOS).
    if [[ "$(uname -s)" == "Darwin" ]]; then
        src=$(find ~/.conan2/p/ -name "$name" -type f -exec stat -f '%m %N' {} \; 2>/dev/null | sort -rn | head -1 | cut -d' ' -f2-)
    else
        src=$(find ~/.conan2/p/ -name "$name" -type f -printf '%T@ %p\n' 2>/dev/null | sort -rn | head -1 | cut -d' ' -f2-)
    fi
    if [ -z "$src" ]; then
        echo "ERROR: $name not found in Conan cache"
        exit 1
    fi
    cp "$src" "$LIBS_DIR/"
    echo "  Copied: $name ($(du -h "$LIBS_DIR/$name" | cut -f1))"
}

copy_lib "libmpt-crypto.a"
copy_lib "libsecp256k1.a"
copy_lib "libcrypto.a"

# --- Copy headers (only once, they're platform-independent) ---

echo "==> Copying headers to $INCLUDE_DIR..."
mkdir -p "$INCLUDE_DIR/utility"

copy_header() {
    local name="$1"
    local dest="$2"
    local src
    src=$(find ~/.conan2/p/ -path "*/p/include/$name" -type f 2>/dev/null | head -1)
    if [ -z "$src" ]; then
        src=$(find ~/.conan2/p/ -name "$(basename "$name")" -path "*/include/*" -type f 2>/dev/null | head -1)
    fi
    if [ -z "$src" ]; then
        echo "ERROR: header $name not found in Conan cache"
        exit 1
    fi
    cp "$src" "$dest"
    echo "  Copied: $name"
}

copy_header "secp256k1_mpt.h" "$INCLUDE_DIR/"
copy_header "utility/mpt_utility.h" "$INCLUDE_DIR/utility/"
copy_header "secp256k1.h" "$INCLUDE_DIR/"

# --- Save version ---

echo "$VERSION" > "$VERSION_FILE"

# --- Summary ---

echo ""
echo "==> Done! Vendored mpt-crypto/$VERSION for $PLATFORM."
echo ""
echo "Libraries:"
ls -lh "$LIBS_DIR/"
echo ""
echo "Headers:"
ls "$INCLUDE_DIR"/secp256k1.h "$INCLUDE_DIR"/secp256k1_mpt.h "$INCLUDE_DIR"/utility/mpt_utility.h
echo ""
echo "Next steps:"
echo "  git add -f confidential/deps/libs/ confidential/deps/include/ confidential/deps/VERSION"
echo "  git commit -m 'feat(confidential): update vendored mpt-crypto to $VERSION'"
