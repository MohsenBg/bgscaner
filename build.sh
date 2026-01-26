#!/bin/bash
set -e

APP=bgscan
DIST=dist

mkdir -p "$DIST"

targets=(
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "darwin/amd64"
  "darwin/arm64"
  "android/arm64"
)

echo "building binaries..."
for target in "${targets[@]}"; do
  IFS=/ read GOOS GOARCH <<<"$target"

  out="$APP-$GOOS-$GOARCH"
  if [ "$GOOS" = "windows" ]; then
    out="$out.exe"
  fi

  echo "  -> $out"
  GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o "$DIST/$out"
done

echo
echo "generating checksums..."

(
  cd "$DIST"
  sha256sum * >checksums.txt
)

echo
echo "done."
echo "artifacts in ./$DIST/"
