# msteams-cli Design Spec

**Date:** 2026-05-31

## Overview

A Go-based CLI tool that sends a Microsoft Teams direct message to a user by email address, using the MS Graph REST API with delegated (interactive browser) OAuth2 authentication.

## Goals

- Send a 1:1 Teams message to any user by email address
- Authenticate interactively via browser (OAuth2 delegated flow) on first run
- Cache the token between runs — no re-authentication until the refresh token expires
- Simple, single-purpose CLI with one primary subcommand: `send`

## Non-Goals

- Sending to channels or group chats
- Reading messages or listing chats
- App-only / service principal authentication
- Message formatting beyond plain text

## CLI Usage

```
msteams-cli send --to user@example.com --message "Hello from the CLI"
```

Flags:
- `--to` (required): recipient's email address
- `--message` / `-m` (required): message body (plain text)
- `--client-id`: Azure app registration client ID (required on first run; saved to config)

## Project Structure

```
msteams-cli/
├── main.go
├── cmd/
│   ├── root.go       # cobra root command, config loading
│   └── send.go       # send subcommand
├── internal/
│   ├── auth/
│   │   └── auth.go   # InteractiveBrowserCredential + token cache
│   └── graph/
│       └── client.go # raw HTTP calls to MS Graph REST API
└── go.mod
```

## Authentication

**Provider:** `azidentity.InteractiveBrowserCredential` from `github.com/Azure/azure-sdk-for-go/sdk/azidentity`

**Flow:**
1. On first run (or expired refresh token), the system browser opens to Microsoft's login page
2. User signs in with their Microsoft 365 account
3. Token (including refresh token) is cached to `~/.config/msteams-cli/token_cache.json`
4. On subsequent runs, the cached token is used and refreshed automatically

**Config file:** `~/.config/msteams-cli/config.json` stores the Azure client ID so it does not need to be passed on every invocation.

### Azure App Registration (one-time manual setup)

The user must create an Azure app registration before first use:

1. Go to [portal.azure.com](https://portal.azure.com) → Azure Active Directory → App registrations → New registration
2. Set platform to **Mobile and desktop applications**
3. Add redirect URI: `http://localhost`
4. Add the following **delegated** API permissions (none require admin consent):
   - `User.Read` — read the authenticated user's own profile
   - `User.ReadBasic.All` — resolve recipient email to user ID
   - `Chat.Create` — create a 1:1 chat
   - `ChatMessage.Send` — post a message to the chat
5. Copy the **Application (client) ID** — pass it via `--client-id` on first run

## Graph API Flow

Three sequential calls to `https://graph.microsoft.com/v1.0`:

| Step | Method | Endpoint | Purpose |
|------|--------|----------|---------|
| 1 | GET | `/me` | Get the authenticated sender's user ID |
| 2 | GET | `/users/{email}` | Resolve recipient email → user ID |
| 3 | POST | `/chats` | Create 1:1 chat (returns existing if already present) |
| 4 | POST | `/chats/{chatId}/messages` | Send the message |

### Step 2 request body
```json
{
  "chatType": "oneOnOne",
  "members": [
    { "@odata.type": "#microsoft.graph.aadUserConversationMember", "roles": ["owner"], "user@odata.bind": "https://graph.microsoft.com/v1.0/users/{senderId}" },
    { "@odata.type": "#microsoft.graph.aadUserConversationMember", "roles": ["owner"], "user@odata.bind": "https://graph.microsoft.com/v1.0/users/{recipientId}" }
  ]
}
```

### Step 3 request body
```json
{
  "body": { "content": "Hello from the CLI" }
}
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Recipient email not found | Print clear error: "User not found: {email}" |
| Auth failure / expired token | Delete cached token, re-trigger browser login |
| Missing permission | Surface the Graph error message and remind user to check app registration |
| Network / API error | Print the HTTP status and Graph error response body |

## Dependencies

- `github.com/spf13/cobra` — CLI parsing
- `github.com/Azure/azure-sdk-for-go/sdk/azidentity` — OAuth2 browser auth + token cache
- `github.com/Azure/azure-sdk-for-go/sdk/azcore` — HTTP pipeline / token provider interface

## Testing Approach

- Unit tests for Graph API request building (mock HTTP responses)
- Manual integration test against a real tenant for end-to-end validation
- No mocking of the auth layer — test auth flows manually
