#!/bin/bash
set -euo pipefail

# macOS build script: builds the Wails app, bundles FFmpeg, signs, and creates a DMG.
#
# Usage:
#   ./scripts/build-macos.sh                                                    # ad-hoc signing
#   ./scripts/build-macos.sh --sign "Developer ID Application: Name (TEAMID)"  # distribution signing
#   ./scripts/build-macos.sh --sign "..." --notarize                            # + notarization

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

APP_NAME="ONX Screen Record"
OUTPUT_NAME="onx-screen-record"
BUILD_DIR="$PROJECT_ROOT/build/bin"
APP_BUNDLE="$BUILD_DIR/${APP_NAME}.app"
DMG_OUTPUT="$BUILD_DIR/${OUTPUT_NAME}.dmg"
ENTITLEMENTS="$PROJECT_ROOT/build/darwin/entitlements.plist"
FFMPEG_SOURCE="$PROJECT_ROOT/build/darwin/resources/ffmpeg"

SIGN_IDENTITY="-"  # ad-hoc by default
NOTARIZE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --sign)
            SIGN_IDENTITY="$2"
            shift 2
            ;;
        --notarize)
            NOTARIZE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--sign <identity>] [--notarize]"
            exit 1
            ;;
    esac
done

echo "=========================================="
echo "  macOS Build Pipeline"
echo "=========================================="
echo "  Sign identity: $SIGN_IDENTITY"
echo "  Notarize: $NOTARIZE"
echo ""

# ──────────────────────────────────────────────
# Step 1: Check prerequisites
# ──────────────────────────────────────────────
echo "==> Step 1: Checking prerequisites..."

if ! command -v wails &>/dev/null; then
    echo "Error: wails CLI not found. Install it: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

if ! command -v codesign &>/dev/null; then
    echo "Error: codesign not found. Install Xcode Command Line Tools: xcode-select --install"
    exit 1
fi

if [ ! -f "$ENTITLEMENTS" ]; then
    echo "Error: Entitlements file not found at $ENTITLEMENTS"
    exit 1
fi

echo "    All prerequisites met."

# ──────────────────────────────────────────────
# Step 2: Build with Wails
# ──────────────────────────────────────────────
echo "==> Step 2: Building with Wails..."

cd "$PROJECT_ROOT"
wails build -platform darwin/universal -clean

if [ ! -d "$APP_BUNDLE" ]; then
    echo "Error: App bundle not found at $APP_BUNDLE"
    echo "Available files in build/bin:"
    ls -la "$BUILD_DIR/" 2>/dev/null || true
    exit 1
fi

echo "    App bundle created: $APP_BUNDLE"

# ──────────────────────────────────────────────
# Step 3: Bundle FFmpeg
# ──────────────────────────────────────────────
echo "==> Step 3: Bundling FFmpeg..."

if [ -f "$FFMPEG_SOURCE" ]; then
    RESOURCES_DIR="$APP_BUNDLE/Contents/Resources"
    mkdir -p "$RESOURCES_DIR"
    cp "$FFMPEG_SOURCE" "$RESOURCES_DIR/ffmpeg"
    chmod +x "$RESOURCES_DIR/ffmpeg"
    echo "    FFmpeg bundled into app."
else
    echo "    Warning: FFmpeg not found at $FFMPEG_SOURCE"
    echo "    Run ./scripts/download-ffmpeg-macos.sh first to bundle FFmpeg."
    echo "    Continuing without bundled FFmpeg..."
fi

# ──────────────────────────────────────────────
# Step 4: Code sign
# ──────────────────────────────────────────────
echo "==> Step 4: Code signing..."

# Sign the bundled ffmpeg binary first (if present)
if [ -f "$APP_BUNDLE/Contents/Resources/ffmpeg" ]; then
    codesign --force --options runtime \
        --sign "$SIGN_IDENTITY" \
        --entitlements "$ENTITLEMENTS" \
        "$APP_BUNDLE/Contents/Resources/ffmpeg"
    echo "    Signed bundled FFmpeg."
fi

# Sign the app bundle
codesign --force --deep --options runtime \
    --sign "$SIGN_IDENTITY" \
    --entitlements "$ENTITLEMENTS" \
    "$APP_BUNDLE"

echo "    App bundle signed."

# Verify signature
codesign --verify --verbose=2 "$APP_BUNDLE" 2>&1 || {
    echo "Error: Code signature verification failed!"
    exit 1
}
echo "    Signature verified."

# ──────────────────────────────────────────────
# Step 5: Create DMG
# ──────────────────────────────────────────────
echo "==> Step 5: Creating DMG..."

# Remove existing DMG
rm -f "$DMG_OUTPUT"

if command -v create-dmg &>/dev/null; then
    # Use create-dmg for a nice drag-to-Applications layout
    create-dmg \
        --volname "$APP_NAME" \
        --volicon "$PROJECT_ROOT/build/darwin/appicon.icns" \
        --window-pos 200 120 \
        --window-size 600 400 \
        --icon-size 100 \
        --icon "$APP_NAME.app" 150 190 \
        --hide-extension "$APP_NAME.app" \
        --app-drop-link 450 190 \
        "$DMG_OUTPUT" \
        "$APP_BUNDLE"
    echo "    DMG created with drag-to-Applications layout."
else
    echo "    create-dmg not found, using hdiutil fallback."
    echo "    (Install create-dmg for a nicer DMG: brew install create-dmg)"

    TMPDIR_DMG="$(mktemp -d)"
    trap 'rm -rf "$TMPDIR_DMG"' EXIT

    cp -R "$APP_BUNDLE" "$TMPDIR_DMG/"
    ln -s /Applications "$TMPDIR_DMG/Applications"

    hdiutil create -volname "$APP_NAME" \
        -srcfolder "$TMPDIR_DMG" \
        -ov -format UDZO \
        "$DMG_OUTPUT"

    echo "    DMG created with hdiutil."
fi

echo "    Output: $DMG_OUTPUT"

# ──────────────────────────────────────────────
# Step 6: Notarize (optional)
# ──────────────────────────────────────────────
if [ "$NOTARIZE" = true ]; then
    echo "==> Step 6: Notarizing..."

    if [ "$SIGN_IDENTITY" = "-" ]; then
        echo "Error: Notarization requires a valid Developer ID signing identity."
        echo "Use: $0 --sign 'Developer ID Application: Name (TEAMID)' --notarize"
        exit 1
    fi

    if [ -z "${APPLE_ID:-}" ] || [ -z "${APPLE_TEAM_ID:-}" ]; then
        echo "Error: Notarization requires APPLE_ID and APPLE_TEAM_ID environment variables."
        echo "  export APPLE_ID=your@email.com"
        echo "  export APPLE_TEAM_ID=YOURTEAMID"
        echo ""
        echo "You also need an app-specific password stored in the keychain:"
        echo "  xcrun notarytool store-credentials 'notarytool-profile' \\"
        echo "    --apple-id \$APPLE_ID --team-id \$APPLE_TEAM_ID"
        exit 1
    fi

    xcrun notarytool submit "$DMG_OUTPUT" \
        --apple-id "$APPLE_ID" \
        --team-id "$APPLE_TEAM_ID" \
        --keychain-profile "notarytool-profile" \
        --wait

    xcrun stapler staple "$DMG_OUTPUT"

    echo "    Notarization complete."
else
    echo "==> Step 6: Skipping notarization (use --notarize to enable)."
fi

# ──────────────────────────────────────────────
# Done
# ──────────────────────────────────────────────
echo ""
echo "=========================================="
echo "  Build complete!"
echo "=========================================="
echo "  App: $APP_BUNDLE"
echo "  DMG: $DMG_OUTPUT"
echo ""
