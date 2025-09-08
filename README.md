# Oncall TUI

Terminal dashboard for Sentry issues, API health latency, and Kubernetes pods, built with Bubble Tea, Bubbles, and Lip Gloss.

## Features

- Recent Sentry errors from multiple projects
- Analytics pane with Sentry issue counts and API latency (curl)
- Kubernetes pods overview with status colors and selection
- Pod log viewer (`l` on selected pod, `Esc` to exit)
- Pane navigation: `Tab` / `Shift+Tab`, pod navigation: `↑/k` and `↓/j`

## Requirements

- Go 1.20+
- `sentry-cli` in PATH and authenticated (`sentry-cli login`)
- `kubectl` in PATH and a current context
- `curl` in PATH

## Setup

```bash
# install deps (modules auto-resolve on build)
go mod tidy
```

## Build

```bash
# macOS (local)
go build -o oncall .

# Linux amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o oncall-linux .

# Windows amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o oncall.exe .
```

## Run

```bash
# from source
go run main.go

# or built binary
./oncall
```

## Keybindings

- Global: `q` or `Ctrl+C` to quit, `Tab`/`Shift+Tab` to switch panes
- Pods pane: `↑/k` and `↓/j` to move selection, `l` to view logs, `Esc` to return

## Configuration

- Sentry org/projects/queries are configured in code (see `getSentryErrorLogsCmd` and `getSentryStatsCmd`).
- API endpoints for latency checks are configured in `getApiResponseTimesCmd`.

## Notes

- The app makes shell calls via `os/exec` to `sentry-cli`, `kubectl`, and `curl`.
- The log viewer uses Bubble’s `viewport` for smooth scrolling.

## License

This project is licensed under the GNU GPL v3.0. See [LICENSE](LICENSE).
