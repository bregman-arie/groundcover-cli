# groundcover-cli

CLI for the observability platform Groundcover, invoked as `gc`.

This is an intentionally small CLI focused on day-2 monitoring workflows:

- List monitors
- List monitor issues
- List/create silences

## Install

Build and install to a common PATH directory:

```bash
go build -o "$HOME/.local/bin/gc" ./cmd/gc

# ensure it's on your PATH (add to your shell rc if needed)
export PATH="$HOME/.local/bin:$PATH"

gc --help
```

System-wide install (may require sudo):

```bash
go build -o /usr/local/bin/gc ./cmd/gc
gc --help
```

Run without building:

```bash
go run ./cmd/gc --help
```

## First Run (Auth Setup)

Commands that call the API will look for config in this order:

1. Flags
2. Environment variables
3. Config file

If the config file does not exist, `gc` will offer an interactive first-run setup and write:

`~/.config/groundcover-cli/config.json`

You can also override the path with `--config`.

## Command Tour

The CLI is small on purpose; these are all the top-level commands you can use today.

### Monitors

List monitors (same behavior; `list` is an alias):

```bash
gc get monitors
gc list monitors

# common flags
gc get monitors --limit 50
gc get monitors --output json
```

### Monitor Issues

List monitor issues (supports filters and pagination):

```bash
gc list monitor-issues

gc list monitor-issues --env prod --namespace payments
gc list monitor-issues --cluster gke-prod-1 --workload api
gc list monitor-issues --monitor-id <uuid> --silenced false

gc list monitor-issues --limit 100 --skip 100
gc list monitor-issues --output json
```

### Silences

List silences:

```bash
gc silence list
gc silence list --active
gc silence list --output json
```

Create a silence:

```bash
# simplest: "now" + duration
gc silence create --comment "maintenance" --duration 1h --matcher namespace=payments

# multiple matchers
gc silence create \
  --comment "deploy freeze" \
  --duration 45m \
  --matcher namespace=payments \
  --matcher workload~api-.*

# explicit times (RFC3339)
gc silence create \
  --comment "incident" \
  --starts-at 2026-03-20T14:00:00Z \
  --ends-at 2026-03-20T15:00:00Z \
  --matcher service=checkout
```

Matcher syntax:

- `name=value` (equals)
- `name!=value` (not equals)
- `name~regex` (regex match)
- `name!~regex` (regex not match)

## Global Flags

Most commands accept these flags:

```text
--api-key       Groundcover API key
--backend-id    Groundcover backend ID
--base-url      Groundcover API base URL (default: https://api.groundcover.com)
--config        Config file path (default: ~/.config/groundcover-cli/config.json)
--timeout       Request timeout (default: 30s)
--output        Output format: table|json (default: table)
```

Run `gc --help` or `gc <command> --help` for the full flag set.

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
