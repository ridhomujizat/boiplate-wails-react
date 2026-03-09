#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

APP_BASENAME="onx-screen-record"
APP_DISPLAY_NAME="ONX Screen Record"
APP_PATH="$ROOT_DIR/build/bin/$APP_BASENAME.app"
RESOURCES_DIR="$APP_PATH/Contents/Resources"
MACOS_DIR="$APP_PATH/Contents/MacOS"
DMG_PATH="$ROOT_DIR/build/bin/$APP_BASENAME.dmg"

resolve_ffmpeg_source() {
    local candidates=()

    if [ -n "${FFMPEG_SRC:-}" ]; then
        candidates+=("$FFMPEG_SRC")
    fi

    candidates+=(
        "$ROOT_DIR/build/darwin/resources/ffmpeg"
        "$ROOT_DIR/build/darwin/ffmpeg"
        "$ROOT_DIR/assets/ffmpeg"
    )

    local candidate
    for candidate in "${candidates[@]}"; do
        if [ -f "$candidate" ]; then
            echo "$candidate"
            return 0
        fi
    done

    return 1
}

ENTITLEMENTS_FILE="$(mktemp /tmp/entitlements.XXXXXX.plist)"
DMG_STAGING_DIR="$(mktemp -d /tmp/onx-dmg.XXXXXX)"

cleanup() {
    rm -f "$ENTITLEMENTS_FILE"
    rm -rf "$DMG_STAGING_DIR"
}
trap cleanup EXIT

echo "Building with Wails..."
(
    cd "$ROOT_DIR"
    wails build -platform darwin/universal
)

if [ ! -d "$APP_PATH" ]; then
    echo "ERROR: app bundle not found at $APP_PATH"
    exit 1
fi

mkdir -p "$RESOURCES_DIR"

if FFMPEG_SOURCE="$(resolve_ffmpeg_source)"; then
    echo "Bundling FFmpeg from $FFMPEG_SOURCE..."
    cp "$FFMPEG_SOURCE" "$RESOURCES_DIR/ffmpeg"
    chmod +x "$RESOURCES_DIR/ffmpeg"
    echo "FFmpeg bundled at $RESOURCES_DIR/ffmpeg"
else
    echo "WARNING: FFmpeg binary not found. FFmpeg will NOT be bundled."
    echo "Expected one of:"
    echo "  - $ROOT_DIR/build/darwin/resources/ffmpeg"
    echo "  - $ROOT_DIR/build/darwin/ffmpeg"
    echo "  - $ROOT_DIR/assets/ffmpeg"
    echo "Or set env var FFMPEG_SRC=/path/to/ffmpeg"
fi

echo "Code signing app and embedded binaries..."

if [ -f "$RESOURCES_DIR/ffmpeg" ]; then
    echo "Signing embedded ffmpeg..."
    codesign --force --sign - "$RESOURCES_DIR/ffmpeg"
fi

codesign --force --sign - "$MACOS_DIR/$APP_BASENAME"

cat > "$ENTITLEMENTS_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.device.audio-input</key>
    <true/>
    <key>com.apple.security.device.camera</key>
    <true/>
</dict>
</plist>
EOF

codesign --force --deep --sign - --entitlements "$ENTITLEMENTS_FILE" "$APP_PATH"
echo "Code signing complete."

echo "Creating DMG..."
rm -f "$DMG_PATH"
cp -R "$APP_PATH" "$DMG_STAGING_DIR/"
ln -s /Applications "$DMG_STAGING_DIR/Applications"

hdiutil create \
    -volname "$APP_DISPLAY_NAME" \
    -srcfolder "$DMG_STAGING_DIR" \
    -ov \
    -format UDZO \
    "$DMG_PATH"

echo "Build complete:"
echo "  App: $APP_PATH"
echo "  DMG: $DMG_PATH"
