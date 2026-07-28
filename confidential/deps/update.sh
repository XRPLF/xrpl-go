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
#   bash confidential/deps/update.sh --lockfile conan.lock  # use a pre-resolved Conan graph
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
LOCKFILE=""
FORCE=false

while [[ $# -gt 0 ]]; do
	case "$1" in
	--version)
		VERSION="$2"
		shift 2
		;;
	--platform)
		PLATFORM="$2"
		shift 2
		;;
	--lockfile)
		LOCKFILE="$2"
		shift 2
		;;
	--force)
		FORCE=true
		shift
		;;
	-h | --help)
		sed -n '2,/^$/s/^# \?//p' "$0"
		exit 0
		;;
	*)
		# Bare arg = version (backward compat)
		VERSION="$1"
		shift
		;;
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

if ! command -v python3 &>/dev/null; then
	echo "ERROR: python3 is required to resolve versions and filter headers."
	exit 1
fi

if ! command -v cc &>/dev/null; then
	echo "ERROR: a C compiler is required to resolve the public header dependencies."
	exit 1
fi

if ! command -v shasum &>/dev/null; then
	echo "ERROR: shasum is required to create and verify the header manifest."
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
	x86_64) DETECTED_ARCH="amd64" ;;
	aarch64 | arm64) DETECTED_ARCH="arm64" ;;
	esac
	PLATFORM="${DETECTED_OS}-${DETECTED_ARCH}"
fi

# Map platform to Conan settings
case "$PLATFORM" in
linux-amd64)
	CONAN_OS="Linux"
	CONAN_ARCH="x86_64"
	;;
linux-arm64)
	CONAN_OS="Linux"
	CONAN_ARCH="armv8"
	;;
darwin-amd64)
	CONAN_OS="Macos"
	CONAN_ARCH="x86_64"
	;;
darwin-arm64)
	CONAN_OS="Macos"
	CONAN_ARCH="armv8"
	;;
*)
	echo "ERROR: unsupported platform '$PLATFORM'"
	echo "Supported: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64"
	exit 1
	;;
esac

LIBS_DIR="$SCRIPT_DIR/libs/$PLATFORM"
INCLUDE_DIR="$SCRIPT_DIR/include"

if [ -n "$LOCKFILE" ]; then
	if [ ! -f "$LOCKFILE" ]; then
		echo "ERROR: Conan lockfile not found: $LOCKFILE"
		exit 1
	fi
	LOCKFILE="$(cd "$(dirname "$LOCKFILE")" && pwd)/$(basename "$LOCKFILE")"
fi

# --- Resolve version ---

if [ -z "$VERSION" ]; then
	echo "==> Querying latest mpt-crypto version from xrplf remote..."
	VERSION=$(conan list 'mpt-crypto/*' -r xrplf --format=json |
		python3 -c '
import json
import re
import sys

packages = json.load(sys.stdin)
versions = [
    reference.split("/", 1)[1]
    for remote in packages.values()
    for reference in remote
    if reference.startswith("mpt-crypto/")
]
if not versions:
    raise SystemExit("no mpt-crypto versions found")

def version_key(version):
    match = re.fullmatch(r"(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?", version)
    if not match:
        raise SystemExit(f"unsupported mpt-crypto version format: {version}")
    prerelease = match.group(4)
    prerelease_key = tuple(
        (0, int(part)) if part.isdigit() else (1, part)
        for part in re.split(r"[.-]", prerelease or "")
    )
    return (*(int(match.group(index)) for index in range(1, 4)), prerelease is None, prerelease_key)

print(max(versions, key=version_key))
')
	echo "==> Latest version: $VERSION"
else
	echo "==> Using specified version: $VERSION"
fi

if [[ ! "$VERSION" =~ ^[0-9A-Za-z][0-9A-Za-z._+-]*$ ]]; then
	echo "ERROR: invalid mpt-crypto version '$VERSION'"
	exit 1
fi

# --- Check if already up to date ---

bundle_is_complete() {
	local required
	for required in \
		"$LIBS_DIR/libmpt-crypto.a" \
		"$LIBS_DIR/libsecp256k1.a" \
		"$LIBS_DIR/libcrypto.a" \
		"$INCLUDE_DIR/mpt_protocol.h" \
		"$INCLUDE_DIR/secp256k1_mpt.h" \
		"$INCLUDE_DIR/secp256k1.h" \
		"$INCLUDE_DIR/utility/mpt_utility.h"; do
		[ -s "$required" ] || return 1
	done
}

if [ "$FORCE" = false ]; then
	CURRENT_VERSION=""
	if [ -f "$VERSION_FILE" ]; then
		CURRENT_VERSION=$(cat "$VERSION_FILE")
	fi

	if [ "$CURRENT_VERSION" = "$VERSION" ] && bundle_is_complete; then
		echo "==> Already at mpt-crypto/$VERSION for $PLATFORM. Nothing to do."
		echo "    Use --force to re-fetch."
		exit 0
	fi

	if [ "$CURRENT_VERSION" = "$VERSION" ]; then
		echo "==> Repairing incomplete mpt-crypto/$VERSION bundle for $PLATFORM"
	elif [ -n "$CURRENT_VERSION" ]; then
		echo "==> Updating: $CURRENT_VERSION -> $VERSION"
	else
		echo "==> Installing mpt-crypto/$VERSION for $PLATFORM"
	fi
fi

# --- Fetch via Conan ---

WORK_DIR=$(mktemp -d)
HEADER_MANIFEST="$WORK_DIR/MANIFEST.sha256"
LIBS_NEXT="${LIBS_DIR}.next.$$"
INCLUDE_NEXT="${INCLUDE_DIR}.next.$$"
VERSION_NEXT="${VERSION_FILE}.next.$$"

cleanup() {
	rm -rf "$WORK_DIR" "$LIBS_NEXT" "$INCLUDE_NEXT" "$VERSION_NEXT"
}
trap cleanup EXIT

cat >"$WORK_DIR/conanfile.txt" <<EOF
[requires]
mpt-crypto/$VERSION

[generators]
CMakeDeps
CMakeToolchain
EOF

cd "$WORK_DIR"
echo "==> Running conan install for $PLATFORM (this may build from source on first run)..."
CONAN_INSTALL_ARGS=(
	--build=missing
	-o "*:shared=False"
	-o "*:fPIC=True"
	-s "os=$CONAN_OS"
	-s "arch=$CONAN_ARCH"
	--deployer=full_deploy
	--deployer-folder "$WORK_DIR/deploy"
)
if [ -n "$LOCKFILE" ]; then
	CONAN_INSTALL_ARGS+=(--lockfile "$LOCKFILE")
fi
conan install . "${CONAN_INSTALL_ARGS[@]}"

# --- Stage the resolved CGO bundle ---

DEPLOY_ROOT="$WORK_DIR/deploy/full_deploy/host"
STAGE_LIBS="$WORK_DIR/stage/libs"
AVAILABLE_INCLUDE="$WORK_DIR/available-include"
STAGE_INCLUDE="$WORK_DIR/stage/include"
HEADER_PROBE="$WORK_DIR/header_probe.c"
HEADER_DEPS="$WORK_DIR/header_probe.d"

if [ ! -d "$DEPLOY_ROOT" ]; then
	echo "ERROR: Conan full deploy output not found at $DEPLOY_ROOT"
	exit 1
fi

find_unique_archive() {
	local archive="$1"
	local matches
	local count

	matches=$(find "$DEPLOY_ROOT" -type f -path "*/lib/$archive" -print)
	if [ -z "$matches" ]; then
		echo "ERROR: $archive not found in the resolved Conan graph" >&2
		return 1
	fi

	count=$(printf '%s\n' "$matches" | grep -c .)
	if [ "$count" -ne 1 ]; then
		echo "ERROR: expected one $archive in the resolved Conan graph, found $count:" >&2
		printf '%s\n' "$matches" >&2
		return 1
	fi

	printf '%s\n' "$matches"
}

merge_include_tree() {
	local package_root="$1"
	local package_name="$2"
	local source_dir="$package_root/include"
	local source_file
	local relative_path
	local destination_file
	local header_count

	if [ ! -d "$source_dir" ]; then
		echo "ERROR: public include tree not found at $source_dir"
		exit 1
	fi

	header_count=$(find "$source_dir" -type f -name '*.h' -print | wc -l | tr -d ' ')
	if [ "$header_count" -eq 0 ]; then
		echo "ERROR: no public headers found at $source_dir"
		exit 1
	fi

	while IFS= read -r source_file; do
		relative_path=${source_file#"$source_dir/"}
		destination_file="$AVAILABLE_INCLUDE/$relative_path"
		if [ -e "$destination_file" ] && ! cmp -s "$source_file" "$destination_file"; then
			echo "ERROR: conflicting public header: $relative_path"
			exit 1
		fi
	done < <(find "$source_dir" -type f -print)

	cp -R "$source_dir/." "$AVAILABLE_INCLUDE/"
	echo "  Discovered: $header_count headers from $package_name"
}

MPT_LIB=$(find_unique_archive "libmpt-crypto.a")
SECP256K1_LIB=$(find_unique_archive "libsecp256k1.a")
CRYPTO_LIB=$(find_unique_archive "libcrypto.a")
MPT_PACKAGE_ROOT=${MPT_LIB%/lib/libmpt-crypto.a}
SECP256K1_PACKAGE_ROOT=${SECP256K1_LIB%/lib/libsecp256k1.a}

mkdir -p "$STAGE_LIBS" "$AVAILABLE_INCLUDE" "$STAGE_INCLUDE"
cp "$MPT_LIB" "$SECP256K1_LIB" "$CRYPTO_LIB" "$STAGE_LIBS/"
merge_include_tree "$MPT_PACKAGE_ROOT" "mpt-crypto"
merge_include_tree "$SECP256K1_PACKAGE_ROOT" "secp256k1"

# Resolve the downstream header closure from the same root and include paths
# used by the CGO preamble. -MM excludes system headers and fails if a package
# header references an unavailable dependency.
cat >"$HEADER_PROBE" <<'EOF'
#include "mpt_utility.h"
EOF
cc \
	-I"$AVAILABLE_INCLUDE" \
	-I"$AVAILABLE_INCLUDE/utility" \
	-MM \
	-MF "$HEADER_DEPS" \
	"$HEADER_PROBE"

python3 - "$AVAILABLE_INCLUDE" "$STAGE_INCLUDE" "$HEADER_DEPS" <<'PY'
from pathlib import Path
import shlex
import shutil
import sys

source_root = Path(sys.argv[1]).resolve()
destination_root = Path(sys.argv[2])
dependency_file = Path(sys.argv[3])
dependency_text = dependency_file.read_text().replace("\\\n", " ")
try:
    dependency_values = shlex.split(dependency_text.split(":", 1)[1])
except IndexError as error:
    raise SystemExit(f"invalid compiler dependency output: {dependency_text}") from error

copied = set()
for dependency_value in dependency_values:
    dependency = Path(dependency_value).resolve()
    try:
        relative_path = dependency.relative_to(source_root)
    except ValueError:
        continue
    if relative_path in copied:
        continue
    if not dependency.is_file():
        raise SystemExit(f"resolved header is not a file: {dependency}")
    destination = destination_root / relative_path
    destination.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(dependency, destination)
    copied.add(relative_path)

if not copied:
    raise SystemExit("compiler did not resolve any vendored headers")

print(f"  Filtered: {len(copied)} downstream headers")
for relative_path in sorted(copied):
    print(f"    {relative_path}")
PY

(
	cd "$STAGE_INCLUDE"
	find . -type f -print |
		LC_ALL=C sort |
		while IFS= read -r header; do shasum -a 256 "$header"; done
) >"$HEADER_MANIFEST"

for required in \
	"$STAGE_LIBS/libmpt-crypto.a" \
	"$STAGE_LIBS/libsecp256k1.a" \
	"$STAGE_LIBS/libcrypto.a" \
	"$STAGE_INCLUDE/mpt_protocol.h" \
	"$STAGE_INCLUDE/secp256k1_mpt.h" \
	"$STAGE_INCLUDE/secp256k1.h" \
	"$STAGE_INCLUDE/utility/mpt_utility.h"; do
	if [ ! -s "$required" ]; then
		echo "ERROR: incomplete CGO bundle; missing $required"
		exit 1
	fi
done

if ! (cd "$STAGE_INCLUDE" && shasum -a 256 -c "$HEADER_MANIFEST" >/dev/null); then
	echo "ERROR: generated header manifest failed verification"
	exit 1
fi

# Replace rather than overlay so removed upstream files cannot remain stale.
mkdir -p "$(dirname "$LIBS_DIR")" "$(dirname "$INCLUDE_DIR")"
rm -rf "$LIBS_NEXT" "$INCLUDE_NEXT"
cp -R "$STAGE_LIBS" "$LIBS_NEXT"
cp -R "$STAGE_INCLUDE" "$INCLUDE_NEXT"
rm -rf "$LIBS_DIR" "$INCLUDE_DIR"
mv "$LIBS_NEXT" "$LIBS_DIR"
mv "$INCLUDE_NEXT" "$INCLUDE_DIR"
printf '%s\n' "$VERSION" >"$VERSION_NEXT"
mv "$VERSION_NEXT" "$VERSION_FILE"

for archive in "$LIBS_DIR"/*.a; do
	echo "  Copied: $(basename "$archive") ($(du -h "$archive" | cut -f1))"
done

# --- Summary ---

echo ""
echo "==> Done! Vendored mpt-crypto/$VERSION for $PLATFORM."
echo ""
echo "Libraries:"
ls -lh "$LIBS_DIR/"
echo ""
echo "Headers:"
find "$INCLUDE_DIR" -type f -print | sort
echo ""
echo "Next steps:"
echo "  git add -f confidential/deps/libs/ confidential/deps/include/ confidential/deps/VERSION"
echo "  git commit -m 'feat(confidential): update vendored mpt-crypto to $VERSION'"
