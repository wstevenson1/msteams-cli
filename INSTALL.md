# Installation & Setup

## Prerequisites

- Go 1.22+
- A Microsoft 365 work or school account
- An Azure app registration (one-time setup — see below)

## Build

```bash
git clone https://github.com/wstevenson/msteams-cli
cd msteams-cli
go build -o msteams-cli .
```

Move the binary somewhere on your PATH:

```bash
mv msteams-cli /usr/local/bin/
```

## Azure App Registration (one-time)

1. Go to [portal.azure.com](https://portal.azure.com) → **Azure Active Directory** → **App registrations** → **New registration**
2. Give it a name (e.g. `msteams-cli`)
3. Under **Supported account types**, select **Accounts in any organizational directory**
4. Under **Platform**, choose **Mobile and desktop applications**
5. Add redirect URI: `http://localhost`
6. Click **Register** and copy the **Application (client) ID**

### API Permissions

In your new app registration → **API permissions** → **Add a permission** → **Microsoft Graph** → **Delegated permissions**:

| Permission | Purpose |
|---|---|
| `User.Read` | Read your own profile |
| `User.ReadBasic.All` | Look up recipient by email |
| `Chat.Create` | Create 1:1 chat |
| `ChatMessage.Send` | Send the message |

Click **Grant admin consent** if prompted (or ask your tenant admin).

## First Use

```bash
msteams-cli --client-id <YOUR_CLIENT_ID> send \
  --to colleague@yourorg.com \
  --message "Hello from the CLI"
```

Your browser will open for Microsoft sign-in. After authenticating, the client ID is saved to `~/.config/msteams-cli/config.json` and the token is cached — you won't be prompted again until the refresh token expires.

## Subsequent Use

```bash
msteams-cli send --to colleague@yourorg.com --message "Hello again"
```

## Configuration

Config is stored at `~/.config/msteams-cli/config.json`:

```json
{
  "client_id": "your-azure-client-id"
}
```

Token cache is managed by the Azure Identity library (macOS Keychain / OS credential store).
