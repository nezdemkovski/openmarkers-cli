# OpenMarkers CLI

Command-line interface for [OpenMarkers](https://github.com/nezdemkovski/openmarkers) — an open-source biomarker and blood test tracker.

Built **AI-agent first**: structured JSON output, meaningful exit codes, and stdin/pipe support. Interactive TUI for humans as a secondary mode.

## Install

### Homebrew

```bash
brew install nezdemkovski/tap/openmarkers
```

### From source

```bash
go install github.com/openmarkers/openmarkers-cli@latest
```

### Binary releases

Download from [GitHub Releases](https://github.com/nezdemkovski/openmarkers-cli/releases).

## Quick start

```bash
# Authenticate (opens browser for OAuth)
openmarkers auth login

# List your profiles
openmarkers profile list

# View biomarker trends
openmarkers trends 1

# Export a profile as JSON
openmarkers export 1 > backup.json

# Import it back
openmarkers import backup.json --confirm
```

## Commands

### Authentication

```bash
openmarkers auth login      # OAuth 2.1 PKCE flow (opens browser)
openmarkers auth logout     # Delete stored credentials
openmarkers auth status     # Check authentication status
```

### Profiles

```bash
openmarkers profile list
openmarkers profile get <id>
openmarkers profile create --name "Name" --dob 1990-01-15 --sex M
openmarkers profile update <id> --public --handle my-handle
openmarkers profile delete <id>
```

### Biomarkers & Categories

```bash
openmarkers biomarker list [--category lipids]
openmarkers biomarker get <id>
openmarkers biomarker create --id custom_test --category custom --unit mg/dL
openmarkers category list
```

### Results

```bash
openmarkers result list --profile 1 [--biomarker glucose] [--date-from 2024-01-01]
openmarkers result add --profile 1 --biomarker glucose --date 2024-03-15 --value 95
openmarkers result batch-add --profile 1 --date 2024-03-15 --file results.json
openmarkers result update <id> --value 92
openmarkers result delete <id>
```

### Analytics

```bash
openmarkers timeline <profile_id>
openmarkers snapshot <profile_id> --date 2024-03-15
openmarkers trends <profile_id> [--biomarker glucose] [--category lipids]
openmarkers compare <profile_id> --date1 2024-01-01 --date2 2024-06-01
openmarkers correlations <profile_id>
openmarkers bioage <profile_id>
openmarkers analysis <profile_id> [--lang en]
```

### Import & Export

```bash
openmarkers export <profile_id>                    # JSON to stdout
openmarkers export 1 > profile.json                # Save to file
openmarkers import profile.json --confirm           # Import from file
cat profile.json | openmarkers import --confirm     # Import from stdin
```

### Public profiles

```bash
openmarkers public list            # No auth required
openmarkers public get <handle>
```

### Schema

```bash
openmarkers schema                 # Biomarker definitions (no auth required)
```

## Output formats

The CLI defaults to JSON when piped and table format in a terminal.

```bash
# Force JSON (for scripting / AI agents)
openmarkers profile list --json

# Force table
openmarkers profile list --output table

# Force plain text
openmarkers profile list --output text
```

### JSON envelope

Success:
```json
{
  "data": [ ... ]
}
```

Error:
```json
{
  "error": {
    "code": "not_found",
    "message": "Profile not found"
  }
}
```

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error |
| 3 | Authentication required |
| 4 | Not found |
| 5 | Server error |

## Configuration

| Source | Example |
|--------|---------|
| Flag | `--server https://custom.example.com` |
| Env | `OPENMARKERS_SERVER=https://...` |
| Config | `~/.config/openmarkers/config.json` |
| Default | `https://openmarkers.app` |

Resolution order: flag > env > config file > default.

### Token storage

Credentials are stored securely via OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service) with automatic fallback to an encrypted config file.

## Shell completions

```bash
# Bash
openmarkers completion bash > /etc/bash_completion.d/openmarkers

# Zsh
openmarkers completion zsh > "${fpath[1]}/_openmarkers"

# Fish
openmarkers completion fish > ~/.config/fish/completions/openmarkers.fish
```

## Development

```bash
go build -o openmarkers .
go test ./...
go vet ./...
```

## License

MIT
