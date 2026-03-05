# Repository Guidelines

## Project Structure & Module Organization
- `main.go` and `wails.json` define the Wails app entry point and build/runtime configuration.
- `internal/` contains the Go backend. Core app logic and bindings live under `internal/app/`, with services in `internal/service/`, data access in `internal/repository/`, shared types in `internal/common/`, and platform helpers in `internal/pkg/`.
- `frontend/` hosts the React + Vite + TypeScript UI. Source code lives in `frontend/src/`, and Wails-generated bindings are in `frontend/wailsjs/` (do not edit by hand).
- `assets/` stores app assets used by the desktop build.

## Build, Test, and Development Commands
- `wails dev` runs the full app with live reload (Go backend + Vite dev server).
- `wails build` builds a production desktop bundle.
- `cd frontend && bun run dev` runs the frontend only (useful for UI work without Go changes).
- `cd frontend && bun run build` type-checks and builds the frontend.
- `go test ./...` runs Go tests (none currently present).

## Coding Style & Naming Conventions
- Go code should be formatted with `gofmt` and follow standard Go naming (exported `PascalCase`, unexported `camelCase`).
- Frontend uses TypeScript + React; follow existing component and file naming patterns in `frontend/src/` (e.g., `App.tsx`, `pages/`, `components/`).
- Avoid editing auto-generated files under `frontend/wailsjs/`.

## Testing Guidelines
- No test framework or test files are currently in the repo. If adding tests, follow Go’s `*_test.go` convention and colocate with the package under test.
- Frontend tests are not configured; add a tool (e.g., Vitest) only if required by a change.

## Commit & Pull Request Guidelines
- Commit history mixes short messages and Conventional Commit prefixes (e.g., `feat:`, `refactor:`, `style:`). Prefer clear, scoped messages; use Conventional Commit style when possible.
- PRs should include a summary, rationale, and any relevant screenshots for UI changes. Link related issues or tasks if they exist.

## Configuration & Runtime Notes
- `wails.json` controls build commands, output, and app metadata.
- The app uses SQLite via GORM; see `internal/pkg/db/` for migration behavior.
