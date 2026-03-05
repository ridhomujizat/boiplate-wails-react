#!/bin/bash
set -e

FFMPEG_SRC="build/darwin/resources/ffmpeg"
APP_PATH="build/bin/onx-screen-record.app"
RESOURCES_DIR="$APP_PATH/Contents/Resources"
MACOS_DIR="$APP_PATH/Contents/MacOS"

# Build dengan Wails
echo "Building with Wails..."
wails build -platform darwin/universal

# Copy FFmpeg ke .app bundle
if [ -f "$FFMPEG_SRC" ]; then
    echo "Bundling FFmpeg..."
    chmod +x "$FFMPEG_SRC"
    cp "$FFMPEG_SRC" "$RESOURCES_DIR/ffmpeg"
    echo "FFmpeg bundled successfully at $RESOURCES_DIR/ffmpeg"
else
    echo "WARNING: $FFMPEG_SRC not found. FFmpeg will NOT be bundled."
    echo "To bundle FFmpeg, place a static ffmpeg binary at: $FFMPEG_SRC"
    echo "Example: cp \$(which ffmpeg) $FFMPEG_SRC"
    echo "App will require ffmpeg to be installed on user's system."
fi

# Code sign the app (required for macOS permissions to work)
echo "Code signing the app..."

# Sign embedded binaries first (ffmpeg if bundled)
if [ -f "$RESOURCES_DIR/ffmpeg" ]; then
    echo "Signing embedded ffmpeg..."
    codesign --force --sign - "$RESOURCES_DIR/ffmpeg"
fi

# Sign the main executable
codesign --force --sign - "$MACOS_DIR/onx-screen-record"

# Sign the entire app bundle with entitlements for screen recording & audio
ENTITLEMENTS_FILE=$(mktemp /tmp/entitlements.XXXXXX.plist)
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
rm -f "$ENTITLEMENTS_FILE"

echo "Code signing complete."
echo "Build complete: $APP_PATH"
