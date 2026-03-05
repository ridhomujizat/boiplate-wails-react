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
  - `app.go` - Core `App` struct, `Startup()`, auth (Login/Logout), MQTT connection, deep link auth, and file upload
  - `bridge.go` - Settings CRUD, permissions, audio devices, recording controls, and MQTT status
  - `activity_bridge.go` - Activity tracking data queries (timeline, top apps, stats)
  - `server.go` - HTTP server (Gin on port 8080) for integration endpoints
  - `tray_*.go` - Platform-specific system tray implementation

- **`service/`** - Business logic layer
  - `setting/` - Settings management with DTOs (general, audio, activity, recording, upload)
  - `activity/` - Activity tracker service (polling-based window tracking)
  - `mqtt/` - MQTT client (eclipse/paho) with states: disconnected/connecting/connected/reconnecting
  - `auth/` - HTTP-based auth (Login, Logout, DeepLinkAuth) calling backend REST API
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
  - `helper/` - HTTP request helper, device ID, JSON utilities

- **`common/`** - Shared types
  - `model/` - GORM models (ActivityEvent, AppSettings)
  - `enum/` - Enums and constants (HTTP methods, file types, etc.)
  - `type/` - Common DTO types (RecordMQTTPayload, BufferedFile for multipart upload)

### Frontend Structure (`frontend/src/`)

- **`App.tsx`** - Main app with Ant Design ConfigProvider and React Router
- **`pages/`** - Login, Home, Setting, Recording, Activity
- **`layouts/`** - MainLayout wrapper
- **`contexts/`** - AuthContext for authentication state
- **`components/`** - ProtectedRoute, etc.
- **`frontend/wailsjs/go/app/App.js`** - Auto-generated Wails bindings (do not edit manually)
- Uses **Ant Design** v6, **Tailwind CSS** v3, **React Router** v7

### Key Data Flow

1. **Recording**: Frontend calls `StartRecording()` → Go spawns ffmpeg process → Audio recorders capture mic/system audio → `StopRecording()` muxes video + audio with ffmpeg
2. **MQTT-triggered recording**: MQTT message with `action: "start"` → `startRecordingWithSession(sessionId)` → on stop, automatically uploads via `uploadRecording()`
3. **File upload**: After recording stops (via MQTT), Go reads file bytes, sends multipart POST to `{BaseUrl}/api/upload/file` with `folder`, `caption` (sessionId), then optionally deletes local file based on `DeleteAfterUpload` setting
4. **Activity Tracking**: Background goroutine polls active window every N seconds → Stores events in SQLite → Frontend queries via `GetActivityTimeline()`, `GetTopApplications()`
5. **Settings**: Stored in SQLite via GORM, split into groups: general (BaseUrl, TenantCode, MqttBroker), audio, activity, recording, upload

### Wails Runtime Events (Go → Frontend)

Events emitted via `runtime.EventsEmit()` that the frontend can listen to:
- `mqtt-message` - Incoming MQTT message `{topic, payload}`
- `recording-state` - Recording state change `{state, session_id}` or `{state, file_path}`
- `recording-uploaded` - Upload result `{success, filePath, sessionId, statusCode}`
- `deep-link-auth-success` - Deep link auth result `{message, user}`

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
