# macOS Build Guide

This guide covers building ONX Screen Record as a signed `.dmg` for macOS distribution.

## Prerequisites

- **Go** 1.21+
- **Wails CLI** v2 — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Bun** — frontend package manager
- **Xcode Command Line Tools** — `xcode-select --install`
- **create-dmg** (optional, for nicer DMG layout) — `brew install create-dmg`

## Quick Start

```bash
# 1. Download FFmpeg static binary
./scripts/download-ffmpeg-macos.sh

# 2. Build the app and create DMG (ad-hoc signing)
./scripts/build-macos.sh
```

The DMG will be at `build/bin/onx-screen-record.dmg`.

## Build Options

### Ad-hoc Signing (local testing)

```bash
./scripts/build-macos.sh
```

The app will be signed with an ad-hoc identity (`-`). This works for local testing but macOS Gatekeeper will block it for other users.

### Developer ID Signing (distribution)

```bash
./scripts/build-macos.sh --sign "Developer ID Application: Your Name (TEAMID)"
```

### Signing + Notarization (full distribution)

```bash
# Set up credentials (one-time)
export APPLE_ID=your@email.com
export APPLE_TEAM_ID=YOURTEAMID
xcrun notarytool store-credentials 'notarytool-profile' \
    --apple-id $APPLE_ID --team-id $APPLE_TEAM_ID

# Build with signing and notarization
./scripts/build-macos.sh --sign "Developer ID Application: Your Name (TEAMID)" --notarize
```

## FFmpeg Bundling

The app bundles a static FFmpeg binary inside the `.app` bundle at `Contents/Resources/ffmpeg`.

**Download script**: `scripts/download-ffmpeg-macos.sh`
- Downloads from [evermeet.cx](https://evermeet.cx/ffmpeg/) (trusted macOS static builds)
- Places the binary at `build/darwin/resources/ffmpeg`
- The x86_64 binary runs on Apple Silicon via Rosetta 2

**Path resolution** (in `internal/pkg/ffmpeg/ffmpeg.go`):
1. Resolves the app executable path
2. Navigates from `.app/Contents/MacOS/binary` to `.app/Contents/Resources/ffmpeg`
3. Falls back to system PATH if bundled binary is not found

## Entitlements

The file `build/darwin/entitlements.plist` declares these entitlements for hardened runtime:

| Entitlement | Purpose |
|---|---|
| `com.apple.security.device.audio-input` | Microphone recording |
| `com.apple.security.device.camera` | Screen capture on newer macOS |
| `com.apple.security.cs.allow-unsigned-executable-memory` | CGo + ScreenCaptureKit bridging |
| `com.apple.security.cs.allow-jit` | WebKit/WebView JS execution |
| `com.apple.security.cs.disable-library-validation` | Loading bundled ffmpeg binary |

**Note:** Screen recording permission is controlled by TCC via `NSScreenCaptureUsageDescription` in Info.plist, not via entitlements.

## Info.plist Usage Descriptions

Both `Info.plist` and `Info.dev.plist` include:

- `NSScreenCaptureUsageDescription` — Screen recording permission prompt
- `NSMicrophoneUsageDescription` — Microphone permission prompt
- `NSAccessibilityUsageDescription` — Accessibility permission for window tracking

## Build Pipeline Steps

The `scripts/build-macos.sh` script performs these steps:

1. **Check prerequisites** — verifies wails, codesign, and entitlements file
2. **Build with Wails** — `wails build -platform darwin/universal`
3. **Bundle FFmpeg** — copies static binary into `.app/Contents/Resources/`
4. **Code sign** — signs FFmpeg first, then the app bundle with entitlements
5. **Create DMG** — uses `create-dmg` (with drag-to-Applications layout) or falls back to `hdiutil`
6. **Notarize** (optional) — submits to Apple and staples the ticket

## Troubleshooting

### "FFmpeg not found" in the app
- Ensure you ran `./scripts/download-ffmpeg-macos.sh` before building
- Verify the binary exists: `ls -la build/darwin/resources/ffmpeg`

### Gatekeeper blocks the app
- Ad-hoc signed apps are blocked by default on other machines
- Use Developer ID signing for distribution: `--sign "Developer ID Application: ..."`
- For local testing: System Settings → Privacy & Security → Open Anyway

### Permission prompts don't appear
- Ensure the app is properly code signed with entitlements
- Check Info.plist contains the usage description keys
- Reset TCC database for testing: `tcsutil reset All com.wails.onx-screen-record`

### Notarization fails
- Ensure `APPLE_ID` and `APPLE_TEAM_ID` environment variables are set
- Ensure keychain profile is stored: `xcrun notarytool store-credentials`
- Check notarization log: `xcrun notarytool log <submission-id> --apple-id $APPLE_ID --team-id $APPLE_TEAM_ID`
