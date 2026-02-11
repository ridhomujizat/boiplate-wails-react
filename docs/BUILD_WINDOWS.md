# Building Windows Executable

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)
- [Bun](https://bun.sh/) (frontend package manager)
- [NSIS](https://nsis.sourceforge.io/Download) (only for installer builds)
- Windows 10/11 (or cross-compile setup)

### Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Verify environment

```bash
wails doctor
```

---

## Build Options

### 1. Standard EXE (no installer)

```bash
wails build
```

Output: `build/bin/onx-screen-record.exe`

### 2. EXE with NSIS Installer

```bash
wails build --target windows/amd64 --nsis
```

Output: `build/bin/onx-screen-record-amd64-installer.exe`

### 3. Debug build (with dev tools)

```bash
wails build -debug
```

---

## Bundling FFmpeg

FFmpeg is required for screen recording. The installer can bundle it automatically.

### Step 1: Download FFmpeg

Run the PowerShell script to download FFmpeg into the installer resources:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/download-ffmpeg.ps1
```

This downloads `ffmpeg.exe` to `build/windows/installer/resources/`.

### Step 2: Build the installer

```bash
wails build --target windows/amd64 --nsis
```

The NSIS installer will include `ffmpeg.exe` alongside the application. If FFmpeg is not present in the resources folder, the installer still builds successfully (FFmpeg just won't be bundled).

---

## Installer Features

The NSIS installer includes these optional components:

| Component | Default | Description |
|-----------|---------|-------------|
| Application (required) | Always | Main app + bundled FFmpeg |
| Auto-start with Windows | Optional | Adds registry key to start app on login |
| Add firewall rule | Optional | Allows the integration server through Windows Firewall |

---

## Full Build Steps (from scratch)

```powershell
# 1. Install frontend dependencies
cd frontend
bun install
cd ..

# 2. Download FFmpeg for bundling
powershell -ExecutionPolicy Bypass -File scripts/download-ffmpeg.ps1

# 3. Build with NSIS installer
wails build --target windows/amd64 --nsis
```

The final installer will be at: `build/bin/onx-screen-record-amd64-installer.exe`

---

## Troubleshooting

### "ffmpeg not found" at runtime
- If FFmpeg was not bundled in the installer, install it manually and add to PATH
- Or re-build the installer after running `scripts/download-ffmpeg.ps1`

### NSIS not found
- Install NSIS from https://nsis.sourceforge.io/Download
- Ensure `makensis` is in your PATH

### Wails build fails
- Run `wails doctor` to check dependencies
- Ensure Go, Node/Bun, and WebView2 runtime are installed
- Check that `bun install` was run in the `frontend/` directory
