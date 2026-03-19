#!/usr/bin/env bash
set -euo pipefail

APP="bgscan"
DIST="dist"

IPS_DIR="ips"
SETTINGS_DIR="settings"

MAIN_FILE="./cmd/bgscan/main.go"

mkdir -p "$DIST"

targets=(
	"linux/amd64"
	"linux/arm64"
	"windows/amd64"
	"darwin/amd64"
	"darwin/arm64"
	"android/arm64"
)

echo "Building binaries..."

for target in "${targets[@]}"; do
	IFS=/ read -r GOOS GOARCH <<<"$target"

	FOLDER="$APP-$GOOS-$GOARCH"
	OUT_DIR="$DIST/$FOLDER"

	mkdir -p "$OUT_DIR"

	BIN_NAME="$APP"
	if [ "$GOOS" = "windows" ]; then
		BIN_NAME="$APP.exe"
	fi

	echo " -> $FOLDER"

	GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
		go build -trimpath -ldflags="-s -w" -o "$OUT_DIR/$BIN_NAME" "$MAIN_FILE"

	# copy ips folder
	if [ -d "$IPS_DIR" ]; then
		cp -r "$IPS_DIR" "$OUT_DIR/"
	fi

	# copy settings folder
	if [ -d "$SETTINGS_DIR" ]; then
		cp -r "$SETTINGS_DIR" "$OUT_DIR/"
	fi

	# generate checksum inside folder
	(
		cd "$OUT_DIR"
		sha256sum "$BIN_NAME" >"$BIN_NAME.sha256"
	)

done

echo
echo "✅ Build complete"
echo "Artifacts available in ./$DIST/"
