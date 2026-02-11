#!/bin/bash
set -euo pipefail

# Downloads a static ffmpeg binary for macOS and places it in build/darwin/resources/

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RESOURCES_DIR="$PROJECT_ROOT/build/darwin/resources"
FFMPEG_PATH="$RESOURCES_DIR/ffmpeg"

DOWNLOAD_URL="https://evermeet.cx/ffmpeg/getrelease/ffmpeg/zip"

if [ -f "$FFMPEG_PATH" ]; then
    echo "FFmpeg already exists at $FFMPEG_PATH"
    "$FFMPEG_PATH" -version | head -1
    echo "To re-download, delete the file first: rm $FFMPEG_PATH"
    exit 0
fi

echo "==> Creating resources directory..."
mkdir -p "$RESOURCES_DIR"

TMPDIR_DL="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_DL"' EXIT

echo "==> Downloading static ffmpeg for macOS..."
echo "    Source: $DOWNLOAD_URL"
curl -L -o "$TMPDIR_DL/ffmpeg.zip" "$DOWNLOAD_URL"

echo "==> Extracting..."
unzip -o "$TMPDIR_DL/ffmpeg.zip" -d "$TMPDIR_DL"

# The zip contains a single 'ffmpeg' binary
if [ -f "$TMPDIR_DL/ffmpeg" ]; then
    mv "$TMPDIR_DL/ffmpeg" "$FFMPEG_PATH"
else
    echo "Error: ffmpeg binary not found in archive"
    ls -la "$TMPDIR_DL/"
    exit 1
fi

chmod +x "$FFMPEG_PATH"

echo "==> FFmpeg downloaded successfully!"
"$FFMPEG_PATH" -version | head -1
echo "    Location: $FFMPEG_PATH"
