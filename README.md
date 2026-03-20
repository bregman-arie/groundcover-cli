# groundcover-cli

CLI for the observability platform Groundcover.

Minimal Groundcover CLI (invoked as `gc`) focused on monitors: list monitors, list monitor issues, and manage silences.

Status: early MVP; flags and output may change.

## Install

From source:

```bash
cd groundcover-cli
go build -o bin/gc ./cmd/gc
./bin/gc --help
```

## Commands

```bash
gc get monitors
gc list monitors

gc list monitor-issues

gc silence list
gc silence create --comment "maintenance" --duration 1h --matcher namespace=payments
```

Run `gc --help` or `gc <command> --help` for flags.

## Auth

Auth can come from flags, environment variables, or a config file.

Config file:

- Default: `~/.config/groundcover-cli/config.json` (override with `--config`)
- Auto-created on first run (interactive) when a command requires auth

Environment variables (either set works):

- `GC_API_KEY` (or `GROUNDCOVER_API_KEY`)
- `GC_BACKEND_ID` (or `GROUNDCOVER_BACKEND_ID`)
- `GC_BASE_URL` (or `GROUNDCOVER_API_URL`) (optional, defaults to `https://api.groundcover.com`)

Precedence (highest to lowest):

1. Flags
2. Environment variables
3. Config file

## Examples

```bash
export GC_API_KEY=...
export GC_BACKEND_ID=...
export GC_BASE_URL=https://api.groundcover.com

go run ./cmd/gc get monitors
go run ./cmd/gc list monitors --output json

go run ./cmd/gc list monitor-issues --env prod --namespace payments
go run ./cmd/gc silence list --active
go run ./cmd/gc silence create \
  --comment "deploy freeze" \
  --duration 45m \
  --matcher namespace=payments \
  --matcher workload~api-.*
```

Config file example:

```json
{
  "api_key": "...",
  "backend_id": "...",
  "base_url": "https://api.groundcover.com"
}
```

## Notes

- `gc list monitor-issues` currently targets `POST /api/monitors/issues/list` and prints a best-effort table (the API response fields can vary).
- `gc silence create` timestamps are in UTC; pass RFC3339 for `--starts-at` / `--ends-at` when needed.
