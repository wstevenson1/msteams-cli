# AI Context — msteams-cli

Use this file to quickly brief a new AI session on this project.

## What This Is

A Go CLI that sends a Microsoft Teams 1:1 message to any user by email address, using delegated OAuth2 (interactive browser) authentication against the MS Graph REST API.

```
msteams-cli send --to user@example.com --message "Hello"
```

## Architecture

Single binary. Four packages:

| Package | File | What it does |
|---|---|---|
| `main` | `main.go` | Bootstraps cobra, prints errors to stderr |
| `cmd` | `cmd/root.go`, `cmd/send.go` | cobra root + `send` subcommand; wires auth + graph |
| `internal/config` | `internal/config/config.go` | Read/write `~/.config/msteams-cli/config.json` |
| `internal/auth` | `internal/auth/auth.go` | `azidentity.InteractiveBrowserCredential` + token cache |
| `internal/graph` | `internal/graph/client.go` | Raw HTTP client for MS Graph v1.0 |

## Key Design Decisions

- **Delegated auth, not app-only.** The message appears as sent from the user's own Teams account. No admin consent required for the four permissions used.
- **azidentity over raw OAuth2.** Handles token refresh and OS keychain caching transparently.
- **Raw HTTP over Graph SDK.** The Microsoft Graph SDK for Go (msgraph-sdk-go) is heavily abstracted for simple use cases. Raw `net/http` calls are easier to read and debug.
- **cobra over plain flags.** Better `--help` output and subcommand extensibility.

## Graph API Flow

1. `GET /v1.0/me` → sender's user ID
2. `GET /v1.0/users/{email}` → recipient's user ID (email is `url.PathEscape`'d)
3. `POST /v1.0/chats` — create or retrieve existing 1:1 chat (`chatType: "oneOnOne"`)
4. `POST /v1.0/chats/{chatId}/messages` — send message body

## Config & Auth

- Azure client ID stored in `~/.config/msteams-cli/config.json` (0600)
- Pass `--client-id <id>` on first run; it's saved and reused automatically
- Token stored in OS keychain via `azidentity.TokenCachePersistenceOptions`
- Scope: `https://graph.microsoft.com/.default`
- Tenant: `organizations` (any work/school account)

## Testing

```bash
# Requires GOROOT set for MacPorts Go
GOROOT=/opt/local/lib/go go test ./...
```

Unit tests use `httptest.NewServer` to mock Graph API responses. Auth cannot be unit-tested (requires live Azure tenant) — tested manually.

## Dependencies

```
github.com/spf13/cobra
github.com/Azure/azure-sdk-for-go/sdk/azidentity
github.com/Azure/azure-sdk-for-go/sdk/azcore
```

## Module Path

`github.com/wstevenson/msteams-cli`

## Spec & Plan

- Design spec: `docs/superpowers/specs/2026-05-31-msteams-cli-design.md`
- Implementation plan: `docs/superpowers/plans/2026-05-31-msteams-cli.md`
