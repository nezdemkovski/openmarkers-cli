# OpenMarkers CLI

Go CLI for the OpenMarkers biomarker tracking API built with Cobra + Bubbletea.

## Build & Test

```bash
go build -o openmarkers .
go test ./...
go vet ./...
```

## Architecture (Clean Architecture)

Strict unidirectional dependency flow: `cmd → presentation → infrastructure → shared`

- `cmd/` — Thin Cobra command orchestration. Each file registers subcommands in `init()`.
- `internal/infrastructure/api/` — HTTP client with auto token refresh.
- `internal/infrastructure/auth/` — OAuth 2.1 PKCE flow + token persistence.
- `internal/infrastructure/config/` — Server URL, default profile, config resolution.
- `internal/presentation/` — Bubbletea TUI models and views.
- `internal/shared/constants/` — App name, config dir name, default server.
- `internal/shared/models/` — Pure data structs with JSON tags matching API responses. Zero internal deps.
- `internal/shared/output/` — Formatters (JSON, text, table). Commands never print directly.
- `internal/shared/ui/` — Centralized Lipgloss styles, color palette, status symbols.

No backward dependencies. `shared/` has zero internal imports. `infrastructure/` only imports `shared/`.

## Conventions

- AI-agent first: structured JSON output (`--json`), meaningful exit codes, stdin/pipe support.
- Exit codes: 0=success, 1=general, 2=usage, 3=auth required, 4=not found, 5=server error.
- Commands use `cmdContext` pattern — shared struct with Client, Formatter, Config. No globals.
- Verbose/debug output goes to stderr only.
- All commands work non-interactively. TUI is opt-in (TTY + no --json/--output flags).
- All styles centralized in `shared/ui/style.go` — no inline Lipgloss in other packages.
- Status symbols (no emojis): • done, → active, ◉ running, ○ pending, ! warning, ✗ error, · info.

## Server

Default: `https://openmarkers.app` (configurable via `--server` or `OPENMARKERS_SERVER`).
