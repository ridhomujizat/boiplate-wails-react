# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ONX Screen Record is a desktop screen recording application built with **Wails v2** (Go backend + React frontend). It supports screen recording with audio (microphone and system audio), activity tracking, and runs on macOS and Windows.

## Development Commands

### Frontend (React + Vite + TypeScript)
Located in `frontend/` directory. Uses **bun** as the package manager.

```bash
# Install dependencies
bun install

# Development server (for frontend-only changes)
cd frontend && bun run dev

# Build frontend
cd frontend && bun run build
```

### Backend (Go + Wails)

```bash
# Run in development mode with hot reload
wails dev

# Build production binary
wails build

# Run Go tests
go test ./...
```

## Architecture

### Backend Structure (`internal/`)

- **`app/`** - Main application logic and Wails bindings
  - `app.go` - Core App struct, startup, and exported methods for frontend
  - `bridge.go` - Settings, permissions, audio devices, and recording controls
  - `activity_bridge.go` - Activity tracking data queries (timeline, top apps, stats)
  - `server.go` - HTTP server (Gin on port 8080) for integration endpoints
  - `tray_*.go` - Platform-specific system tray implementation

- **`service/`** - Business logic layer
  - `setting/` - Settings management with DTOs
  - `activity/` - Activity tracker service (polling-based window tracking)
  - `integration/` - HTTP integration server with Gin

- **`repository/`** - Data access layer (GORM + SQLite)
  - `activity/` - Activity event CRUD and statistics queries
  - `setting/` - App settings persistence

- **`pkg/`** - Platform-specific utilities
  - `recorder/` - Screen and audio recording (ffmpeg on macOS, platform-specific on Windows)
  - `audio/` - Audio device enumeration using `malgo`
  - `activity/` - Active window detection (platform-specific)
  - `tray/` - System tray management
  - `permission/` - Screen recording and accessibility permission checks

- **`common/`** - Shared types
  - `model/` - GORM models (ActivityEvent, AppSettings)
  - `enum/` - Enums and constants
  - `type/` - Common DTO types

### Frontend Structure (`frontend/src/`)

- **`App.tsx`** - Main app with Ant Design ConfigProvider and React Router
- **`pages/`** - Login, Home, Setting, Recording, Activity
- **`layouts/`** - MainLayout wrapper
- **`contexts/`** - AuthContext for authentication state
- **`components/`** - ProtectedRoute, etc.
- Uses **Ant Design** v6, **Tailwind CSS** v3, **React Router** v7

### Key Data Flow

1. **Recording**: Frontend calls `StartRecording()` → Go spawns ffmpeg process → Audio recorders capture mic/system audio → `StopRecording()` muxes video + audio with ffmpeg
2. **Activity Tracking**: Background goroutine polls active window every N seconds → Stores events in SQLite → Frontend queries via `GetActivityTimeline()`, `GetTopApplications()`
3. **Settings**: Stored in SQLite via GORM, retrieved through `setting.IService`

## Platform-Specific Code

Files with `_darwin.go` or `_windows.go` suffix are platform-specific. Wails build tags ensure only the correct platform is compiled:
- macOS: Uses ffmpeg for screen recording, ScreenCaptureKit for system audio
- Windows: Uses different APIs (see `recorder_windows.go`, `window_windows.go`)

## Deep Linking

App handles `onxrecord://` scheme URLs. Deep links are captured in:
1. `main.go` - Initial launch args processed via `SetInitialDeepLink()`
2. `OnSecondInstanceLaunch` - For single-instance handling
3. `app.go:HandleDeepLink()` - Parses and logs deep link data

## Database

SQLite with GORM. Migrations handled in `internal/pkg/db/migrator.go`. Two main tables:
- `activity_events` - Window/activity tracking data
- `app_settings` - Key-value settings storage

## Important Notes

- Frontend uses **bun**, not npm/yarn
- `wails.json` configures frontend commands and app metadata
- App uses **SingleInstanceLock** - second instances trigger deep link handling
- Window close behavior differs by OS (macOS: hide, Windows: minimize to tray)
