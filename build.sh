#!/usr/bin/env bash
#
# bgscan build script
# ----------------------------
# This script builds bgscan for the local platform or for all supported targets.
# It verifies Go availability, checks platform-specific assets, injects version
# information into the binary, and copies all required resources into the final
# build directory.
#
# Author: MohsenBg
# Project: bgscan
#

set -e

# ---------------------------------------------------------------------------
# Supported build targets (GOOS/GOARCH)
# ---------------------------------------------------------------------------
targets=(
	"linux/amd64"
	"linux/arm64"
	"windows/amd64"
	"darwin/amd64"
	"darwin/arm64"
	"android/arm64"
)

APP="bgscan"

IPS_DIR="ips"
SETTINGS_DIR="settings"
ASSETS_DIR="assets"

MAIN_FILE="./cmd/bgscan/main.go"

# ---------------------------------------------------------------------------
# Auto-detect project version from Git
# ---------------------------------------------------------------------------
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

# ---------------------------------------------------------------------------
# Parse flags
# ---------------------------------------------------------------------------

all=false

while [[ $# -gt 0 ]]; do
	case "$1" in
	--all)
		all=true
		shift
		;;
	--version)
		VERSION="$2"
		shift 2
		;;
	*)
		echo "Unknown argument: $1"
		exit 1
		;;
	esac
done

echo "Detected version: $VERSION"
echo

# ---------------------------------------------------------------------------
# Detect local OS
# ---------------------------------------------------------------------------
case "$(uname -s)" in
Linux) OS="linux" ;;
Darwin) OS="darwin" ;;
MINGW* | MSYS* | CYGWIN*) OS="windows" ;;
*)
	echo "Unsupported OS"
	exit 1
	;;
esac

# ---------------------------------------------------------------------------
# Detect local architecture
# ---------------------------------------------------------------------------
case "$(uname -m)" in
x86_64) ARCH="amd64" ;;
aarch64 | arm64) ARCH="arm64" ;;
*)
	echo "Unsupported architecture"
	exit 1
	;;
esac

USER_TARGET="$OS/$ARCH"

# ---------------------------------------------------------------------------
# Intro message
# ---------------------------------------------------------------------------
echo "Thank you for using bgscan."
echo "Building this project helps support free and open access to the internet."
echo "Your contribution is appreciated."
echo

# ---------------------------------------------------------------------------
# Check for Go installation
# ---------------------------------------------------------------------------
echo "[CHECK] Go installation..."

if ! command -v go >/dev/null 2>&1; then
	echo "[ERROR] Go is not installed. Please install Golang first."
	exit 1
fi

echo "[OK] Go is installed."
echo

# ---------------------------------------------------------------------------
# Ensure dist directory exists
# ---------------------------------------------------------------------------
DIST="dist/$VERSION"
mkdir -p "$DIST"

# ---------------------------------------------------------------------------
# Asset dependency checker
# ---------------------------------------------------------------------------

check_dep() {
	local name="$1"
	local target="$2"

	IFS=/ read -r GOOS GOARCH <<<"$target"

	local folder="$ASSETS_DIR/$name/$name-$GOOS-$GOARCH"

	# Pick correct binary name
	if [ "$GOOS" = "windows" ]; then
		bin="$folder/$name.exe"
	else
		bin="$folder/$name"
	fi

	if [ ! -f "$bin" ]; then
		echo "[ERROR] Missing dependency: $bin"
		echo "Please download required assets from GitHub releases:"
		echo "    https://github.com/MohsenBg/bgscaner/releases"
		exit 1
	fi
}

check_dependencies_for_target() {
	local target="$1"
	echo "[CHECK] Dependencies for $target"

	check_dep "dnstt-client" "$target"
	check_dep "slipstream-client" "$target"
	check_dep "xray" "$target"

	echo "[OK] All dependencies exist for $target"
	echo
}

# ---------------------------------------------------------------------------
# Copy the correct assets into the build folder
# ---------------------------------------------------------------------------
copy_assets_for_target() {
	local target="$1"
	local OUT_DIR="$2"
	IFS=/ read -r GOOS GOARCH <<<"$target"

	echo "[COPY] Assets for $target"

	# -----------------------------
	# Xray
	# -----------------------------
	local XRAY_SRC="$ASSETS_DIR/xray/xray-$GOOS-$GOARCH"
	local XRAY_DST="$OUT_DIR/assets/xray"

	mkdir -p "$XRAY_DST"
	cp "$XRAY_SRC/xray"* "$XRAY_DST/" 2>/dev/null || true

	# Create destination folder
	mkdir -p "$XRAY_DST/outbounds"

	# Copy only *.example
	find "$ASSETS_DIR/xray/outbounds" -maxdepth 1 -type f -name "*.example" \
		-exec cp {} "$XRAY_DST/outbounds/" \;

	# -----------------------------
	# dnstt-client
	# -----------------------------
	local DNSTT_SRC="$ASSETS_DIR/dnstt-client/dnstt-client-$GOOS-$GOARCH"
	local DNSTT_DST="$OUT_DIR/assets/dnstt-client"

	mkdir -p "$DNSTT_DST"
	cp "$DNSTT_SRC/"* "$DNSTT_DST/"

	# -----------------------------
	# slipstream-client
	# -----------------------------
	local SLIP_SRC="$ASSETS_DIR/slipstream-client/slipstream-client-$GOOS-$GOARCH"
	local SLIP_DST="$OUT_DIR/assets/slipstream-client"

	mkdir -p "$SLIP_DST"
	cp "$SLIP_SRC/"* "$SLIP_DST/"

	echo "[OK] Assets copied for $target"
	echo
}

# ---------------------------------------------------------------------------
# Process .default files: rename them by removing .default suffix
# ---------------------------------------------------------------------------
process_default_files() {
	local DIR="$1"

	if [ ! -d "$DIR" ]; then
		return
	fi

	echo "[PROCESS] Checking for .default files in $DIR"

	# Find all .default files recursively
	find "$DIR" -type f -name "*.default" | while read -r default_file; do
		# Remove .default suffix to get target filename
		target_file="${default_file%.default}"

		# Move/rename the file
		mv "$default_file" "$target_file"
		echo "  → Renamed: $(basename "$default_file") → $(basename "$target_file")"
	done
}

# ---------------------------------------------------------------------------
# Build a single target
# ---------------------------------------------------------------------------
build_target() {
	local target="$1"

	IFS=/ read -r GOOS GOARCH <<<"$target"

	local FOLDER="$APP-$GOOS-$GOARCH"
	local OUT_DIR="$DIST/$FOLDER"

	mkdir -p "$OUT_DIR"

	local BIN_NAME="$APP"
	[ "$GOOS" = "windows" ] && BIN_NAME="$APP.exe"

	echo "[BUILD] $FOLDER"

	GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
		go build \
		-trimpath \
		-ldflags="-s -w -X 'main.Version=$VERSION'" \
		-o "$OUT_DIR/$BIN_NAME" \
		"$MAIN_FILE"

	# Copy only .default files from ips
	if [ -d "$IPS_DIR" ]; then
		mkdir -p "$OUT_DIR/$IPS_DIR"
		for f in "$IPS_DIR"/*.default; do
			[ -e "$f" ] || continue
			cp "$f" "$OUT_DIR/$IPS_DIR/$(basename "${f%.default}")"
		done
	fi

	# Copy only .default files from settings
	if [ -d "$SETTINGS_DIR" ]; then
		mkdir -p "$OUT_DIR/$SETTINGS_DIR"
		for f in "$SETTINGS_DIR"/*.default; do
			[ -e "$f" ] || continue
			cp "$f" "$OUT_DIR/$SETTINGS_DIR/$(basename "${f%.default}")"
		done
	fi

	# Process .default files in copied directories
	process_default_files "$OUT_DIR/$IPS_DIR"
	process_default_files "$OUT_DIR/$SETTINGS_DIR"

	# Copy platform assets
	copy_assets_for_target "$target" "$OUT_DIR"

	# Generate checksum
	(
		cd "$OUT_DIR" || exit
		sha256sum "$BIN_NAME" >"$FOLDER.sha256"
	)

	echo "[DONE] Built $FOLDER"
	echo
}

# ---------------------------------------------------------------------------
# Build logic (single or all)
# ---------------------------------------------------------------------------
if [ "$all" = true ]; then
	echo "[MODE] Building for ALL targets"
	echo
	for target in "${targets[@]}"; do
		check_dependencies_for_target "$target"
		build_target "$target"
	done
else
	echo "[MODE] Building only for detected target: $USER_TARGET"
	echo
	check_dependencies_for_target "$USER_TARGET"
	build_target "$USER_TARGET"
fi

echo "---------------------------------------------"
echo " Build completed successfully (version $VERSION)"
echo " Output directory: $DIST/"
echo "---------------------------------------------"
