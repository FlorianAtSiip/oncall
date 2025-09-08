# Oncall TUI

Terminal dashboard for Sentry issues, API health latency, Kubernetes pods, and live pod logs, built with Bubble Tea, Bubbles, and Lip Gloss.

## Features

- **Sentry errors (multiple projects)**: Lists recent unresolved issues for `siip-ticketing` and `siip-iam-service` using `sentry-cli`.
- **Analytics pane**:
  - **Sentry totals**: Counts of current issues per project.
  - **API latency**: Measures response times for `https://ticketing.siip.io/health` and `https://iam.siip.io/health` via `curl`.
  - **IAM health details**: Parses status and lists group statuses by querying subgroup endpoints.
- **Kubernetes pods overview**:
  - Fetches `kubectl get pods` and colorizes pod rows by status (Running/Pending/Error states).
  - Shows the current kube context in the pane title.
  - Navigate the list and open logs for the selected pod.
- **Live pod log viewer**:
  - Streams the last lines (`kubectl logs --tail=500`) for the selected pod.
  - Scroll with arrow keys or mouse wheel, press `Esc` to return.
- **UX details**:
  - Splash screen on startup with version.
  - Auto-refresh of panes on a 15s tick, with Sentry errors refreshed at least every 60s.
  - Clean keybindings for navigation across panes and within pods.

## Requirements

- **Go**: 1.24+
- **External tools in PATH**:
  - `sentry-cli`
  - `kubectl`
  - `curl`

## Setup: External Tools

### sentry-cli

- Install:
  - macOS (Homebrew):
    ```bash
    brew install getsentry/tools/sentry-cli
    ```
  - macOS/Linux (curl script): see `sentry-cli` install instructions.
- Authenticate:
  - Interactive login:
    ```bash
    sentry-cli login
    ```
  - Or set a token via environment variable:
    ```bash
    export SENTRY_AUTH_TOKEN=YOUR_TOKEN
    ```
- Verify:
  ```bash
  sentry-cli info
  sentry-cli issues list --org siip --project siip-ticketing | head -n 5 | cat
  ```

### kubectl

- Install (macOS):
  ```bash
  brew install kubectl
  ```
- Ensure you have a valid kubeconfig/current context and permissions.
- Verify:
  ```bash
  kubectl config current-context | cat
  kubectl get pods | head -n 10 | cat
  ```

### curl

- Preinstalled on macOS/Linux. Verify:
  ```bash
  curl --version | cat
  ```

## Build

```bash
# Resolve modules
go mod tidy

# Local build (macOS)
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

- **Global**: `q` or `Ctrl+C` to quit, `Tab` / `Shift+Tab` to switch panes
- **Pods pane**: `↑/k` and `↓/j` to move selection, `l` to view logs, `Esc` to return

## Configuration

- Sentry org/projects/queries are configured in code: `getSentryErrorLogsCmd` and `getSentryStatsCmd` in `sentry.go`.
- API endpoints for latency checks are configured in `getApiResponseTimesCmd` in `health.go`.
- Kubernetes context is read from your current `kubectl` context; switch via `kubectl config use-context`.

## Notes

- The app makes shell calls via `os/exec` to `sentry-cli`, `kubectl`, and `curl`; ensure these are accessible and authenticated where needed.
- Pod coloring heuristics cover common statuses: Running (green), Pending/Initializing (yellow), Error/CrashLoopBackOff/ImagePullBackOff (red).
- A `.gitignore` is included to avoid committing build artifacts, logs, and OS/editor files.

## Troubleshooting

- **Sentry panes empty**: Verify `sentry-cli login` or `SENTRY_AUTH_TOKEN` and project access.
- **Kubernetes pane errors**: Verify kube context (`kubectl config current-context`) and cluster RBAC.
- **API latency errors**: Ensure `curl` is installed and the endpoints are reachable from your network.

## License

This project is licensed under the GNU GPL v3.0. See [LICENSE](LICENSE).
