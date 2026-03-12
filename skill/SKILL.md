---
name: openmarkers-cli
description: "Use the OpenMarkers CLI to track biomarkers and blood test results. Trigger this skill when the user mentions biomarkers, blood tests, lab results, health markers, biological age, or wants to interact with the OpenMarkers API. Also trigger when the user asks to export/import health data, analyze trends in their blood work, compare test dates, or check their biomarker correlations. The CLI is designed for AI agents — always use --json for structured output."
---

# OpenMarkers CLI

CLI for the OpenMarkers biomarker tracking API at `https://openmarkers.app`.

## Key Principles

- Always use `--json` flag — output is `{"data": ...}` on success, `{"error": {"code": "...", "message": "..."}}` on failure
- Parse exit codes: 0=success, 1=general error, 2=usage error, 3=auth required, 4=not found, 5=server error
- If exit code is 3, tell the user to run `openmarkers auth login`
- Pipe-friendly: commands accept stdin JSON and output to stdout
- Default server is `https://openmarkers.app`, override with `--server` or `OPENMARKERS_SERVER`

## Authentication

```bash
openmarkers auth login                    # Opens browser for OAuth 2.1 PKCE
openmarkers auth status --json            # Check if authenticated
openmarkers auth logout                   # Delete stored credentials
```

Tokens are stored in OS keyring (macOS Keychain, etc.) with file fallback. Token refresh is automatic.

## Global Flags

| Flag | Purpose |
|------|---------|
| `--json` | JSON output (always use this) |
| `--profile <id>` | Default profile ID for commands that need one |
| `--server <url>` | Override server URL |
| `--verbose` | Debug info to stderr |

## Commands

### Profiles

```bash
openmarkers profile list --json
openmarkers profile get <id> --json
openmarkers profile create --name "Name" --dob 1990-01-15 --sex M --json
openmarkers profile update <id> --name "New Name" --json
openmarkers profile update <id> --public --handle my-handle --json
openmarkers profile delete <id> --json
```

### Biomarkers & Categories

```bash
openmarkers biomarker list --json                        # All biomarkers
openmarkers biomarker list --category lipids --json      # Filter by category
openmarkers biomarker get <biomarker_id> --json
openmarkers biomarker create --id custom_test --category custom --unit mg/dL --ref-min 10 --ref-max 100 --json
openmarkers biomarker update <biomarker_id> --unit mmol/L --ref-min 5 --ref-max 50 --json
openmarkers category list --json
```

### Results

```bash
openmarkers result list --profile <id> --json
openmarkers result list --profile <id> --biomarker glucose --date-from 2024-01-01 --date-to 2024-12-31 --json
openmarkers result add --profile <id> --biomarker glucose --date 2024-03-15 --value 95 --json
openmarkers result update <result_id> --value 92 --json
openmarkers result delete <result_id> --json
```

Stdin support for adding results:
```bash
echo '{"profile_id":1,"biomarker_id":"glucose","date":"2024-03-15","value":95}' | openmarkers result add --json
```

Batch add:
```bash
echo '{"profile_id":1,"date":"2024-03-15","entries":[{"biomarker_id":"glucose","value":95},{"biomarker_id":"hdl","value":55}]}' | openmarkers result batch-add --json
```

### Analytics

All analytics commands take profile_id as first positional arg or via `--profile`.

```bash
openmarkers timeline <profile_id> --json                                          # Test dates + counts
openmarkers snapshot <profile_id> --date 2024-03-15 --json                        # All values on a date
openmarkers trends <profile_id> --json                                            # Direction, rate of change, warnings
openmarkers trends <profile_id> --biomarker glucose --json                        # Single biomarker trend
openmarkers trends <profile_id> --category lipids --json                          # Category trends
openmarkers compare <profile_id> --date1 2024-01-01 --date2 2024-06-01 --json     # Compare two dates
openmarkers correlations <profile_id> --json                                      # Biomarker correlations
openmarkers bioage <profile_id> --json                                            # Biological age calculation
openmarkers analysis <profile_id> --json                                          # AI analysis prompt
openmarkers analysis <profile_id> --lang cs --json                                # Analysis in Czech
```

### Import & Export

```bash
openmarkers export <profile_id>                                    # JSON to stdout
openmarkers export <profile_id> > backup.json                      # Save to file
openmarkers import backup.json --confirm --json                    # Import from file
cat backup.json | openmarkers import --confirm --json              # Import from stdin
openmarkers export 1 | openmarkers import --confirm --json         # Round-trip
```

Without `--confirm`, import checks for duplicates first and returns a warning if the profile exists.

### Public (No Auth)

```bash
openmarkers public list --json                # List public profiles
openmarkers public get <handle> --json        # Get public profile data
openmarkers schema                            # Biomarker definitions JSON
```

## Response Shapes

### Profile list
```json
{"data": [{"id": 1, "name": "Name", "dateOfBirth": "1990-01-15", "sex": "M", "isPublic": false, "publicHandle": null}]}
```

### Results
```json
{"data": [{"id": 456, "profile_id": 1, "biomarker_id": "glucose", "date": "2024-03-15", "value": "95", "created_at": "..."}]}
```

### Trends
```json
{"data": [{"biomarkerId": "glucose", "categoryId": "metabolism", "direction": "down", "rateChange": -1.2, "overallChange": -5, "trendWarning": false, "improving": true, "latestValue": 90, "latestDate": "2024-06-01"}]}
```

### Biological Age
```json
{"data": [{"date": "2024-01-15", "phenoAge": 35.2, "chronoAge": 40, "delta": -4.8, "mortalityScore": 0.92, "scores": [...]}]}
```

### Error
```json
{"error": {"code": "not_found", "message": "Profile not found"}}
```

## Common Workflows

**Get a full health overview:**
```bash
PROFILE=1
openmarkers trends $PROFILE --json          # What's changing
openmarkers bioage $PROFILE --json          # Biological age
openmarkers correlations $PROFILE --json    # Related markers
```

**Add new lab results from a blood test:**
```bash
openmarkers result batch-add --json <<'EOF'
{"profile_id":1,"date":"2024-03-15","entries":[
  {"biomarker_id":"glucose","value":95},
  {"biomarker_id":"total_cholesterol","value":185},
  {"biomarker_id":"hdl","value":55},
  {"biomarker_id":"ldl","value":110}
]}
EOF
```

**Compare before and after an intervention:**
```bash
openmarkers compare 1 --date1 2024-01-01 --date2 2024-06-01 --json
```

**Backup and restore:**
```bash
openmarkers export 1 > backup.json
openmarkers import backup.json --confirm --json
```
