# msteams-cli Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI that sends a Microsoft Teams 1:1 message to a user by email address, using delegated OAuth2 browser authentication and the MS Graph REST API.

**Architecture:** Single binary with a `send` subcommand. `azidentity.InteractiveBrowserCredential` handles OAuth2 browser login and persistent token caching. All Graph API calls are raw `net/http` against `https://graph.microsoft.com/v1.0`. Config (client ID) is stored in `~/.config/msteams-cli/config.json`.

**Tech Stack:** Go 1.22+, github.com/spf13/cobra, github.com/Azure/azure-sdk-for-go/sdk/azidentity, github.com/Azure/azure-sdk-for-go/sdk/azcore

---

## File Map

| File | Responsibility |
|------|---------------|
| `main.go` | Entry point; calls `cmd.Execute()` |
| `cmd/root.go` | Root cobra command; `--client-id` persistent flag; saves client ID to config |
| `cmd/send.go` | `send` subcommand; wires auth + graph; prints result |
| `cmd/send_test.go` | Tests for `send` command required-flag validation |
| `internal/config/config.go` | Read/write `~/.config/msteams-cli/config.json` |
| `internal/config/config_test.go` | Tests for config Load/Save with temp HOME dir |
| `internal/auth/auth.go` | `NewCredential` + `GetToken` via `InteractiveBrowserCredential` |
| `internal/graph/client.go` | `GetMe`, `GetUserByEmail`, `CreateOrGetChat`, `SendMessage` |
| `internal/graph/client_test.go` | Tests for all four Graph methods using `httptest.NewServer` |

---

### Task 1: Initialize Go module and project structure

**Files:**
- Create: `go.mod`
- Create: `main.go`

- [ ] **Step 1: Initialize the module**

```bash
cd /Users/wstevenson/code/msteams-cli
go mod init github.com/wstevenson/msteams-cli
```

Expected: `go.mod` created containing `module github.com/wstevenson/msteams-cli`.

> Note: if you later push to GitHub under a different username, update the module path in `go.mod` and all import paths.

- [ ] **Step 2: Add dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity@latest
go get github.com/Azure/azure-sdk-for-go/sdk/azcore@latest
```

Expected: `go.mod` and `go.sum` updated with pinned versions.

- [ ] **Step 3: Create directory structure**

```bash
mkdir -p cmd internal/config internal/auth internal/graph
```

- [ ] **Step 4: Create main.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/wstevenson/msteams-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "feat: initialize Go module and project structure"
```

---

### Task 2: Config package

**Files:**
- Create: `internal/config/config_test.go`
- Create: `internal/config/config.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func useTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestLoadMissingFile(t *testing.T) {
	useTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v; want nil", err)
	}
	if cfg.ClientID != "" {
		t.Errorf("ClientID = %q; want empty", cfg.ClientID)
	}
}

func TestSaveAndLoad(t *testing.T) {
	useTempHome(t)
	want := &Config{ClientID: "test-client-id-abc"}
	if err := Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.ClientID != want.ClientID {
		t.Errorf("ClientID = %q; want %q", got.ClientID, want.ClientID)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	useTempHome(t)
	if err := Save(&Config{ClientID: "x"}); err != nil {
		t.Fatal(err)
	}
	dir, _ := Dir()
	info, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("config.json permissions = %04o; want 0600", perm)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/config/...
```

Expected: FAIL — package `config` does not exist yet.

- [ ] **Step 3: Implement config.go**

Create `internal/config/config.go`:

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ClientID string `json:"client_id"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "msteams-cli"), nil
}

func Load() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, json.Unmarshal(data, &cfg)
}

func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0600)
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/config/... -v
```

Expected: `TestLoadMissingFile`, `TestSaveAndLoad`, `TestConfigFilePermissions` all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with load/save"
```

---

### Task 3: Graph API client

**Files:**
- Create: `internal/graph/client_test.go`
- Create: `internal/graph/client.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/graph/client_test.go`:

```go
package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return newWithBaseURL("test-token", srv.URL)
}

func TestGetMe(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("path = %q; want /me", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header missing or wrong")
		}
		json.NewEncoder(w).Encode(User{ID: "sender-id-123"})
	}))

	user, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if user.ID != "sender-id-123" {
		t.Errorf("ID = %q; want sender-id-123", user.ID)
	}
}

func TestGetUserByEmail(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/alice@example.com" {
			t.Errorf("path = %q; want /users/alice@example.com", r.URL.Path)
		}
		json.NewEncoder(w).Encode(User{ID: "recipient-id-456"})
	}))

	user, err := client.GetUserByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}
	if user.ID != "recipient-id-456" {
		t.Errorf("ID = %q; want recipient-id-456", user.ID)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"code":"Request_ResourceNotFound","message":"Resource not found"}}`))
	}))

	_, err := client.GetUserByEmail(context.Background(), "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestCreateOrGetChat(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chats" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["chatType"] != "oneOnOne" {
			t.Errorf("chatType = %v; want oneOnOne", body["chatType"])
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Chat{ID: "chat-id-789"})
	}))

	chat, err := client.CreateOrGetChat(context.Background(), "sender-id-123", "recipient-id-456")
	if err != nil {
		t.Fatalf("CreateOrGetChat() error = %v", err)
	}
	if chat.ID != "chat-id-789" {
		t.Errorf("ID = %q; want chat-id-789", chat.ID)
	}
}

func TestSendMessage(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chats/chat-id-789/messages" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		bodyField := body["body"].(map[string]any)
		if bodyField["content"] != "Hello world" {
			t.Errorf("content = %v; want 'Hello world'", bodyField["content"])
		}
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.SendMessage(context.Background(), "chat-id-789", "Hello world")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/graph/...
```

Expected: FAIL — package `graph` does not exist yet.

- [ ] **Step 3: Implement graph/client.go**

Create `internal/graph/client.go`:

```go
package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const productionBaseURL = "https://graph.microsoft.com/v1.0"

type Client struct {
	http    *http.Client
	token   string
	baseURL string
}

func New(token string) *Client {
	return newWithBaseURL(token, productionBaseURL)
}

func newWithBaseURL(token, baseURL string) *Client {
	return &Client{
		http:    &http.Client{},
		token:   token,
		baseURL: baseURL,
	}
}

type User struct {
	ID string `json:"id"`
}

type Chat struct {
	ID string `json:"id"`
}

func (c *Client) GetMe(ctx context.Context) (*User, error) {
	return c.getUser(ctx, c.baseURL+"/me")
}

func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return c.getUser(ctx, c.baseURL+"/users/"+email)
}

func (c *Client) getUser(ctx context.Context, url string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readGraphError(resp)
	}

	var user User
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

func (c *Client) CreateOrGetChat(ctx context.Context, senderID, recipientID string) (*Chat, error) {
	body := map[string]any{
		"chatType": "oneOnOne",
		"members": []map[string]any{
			{
				"@odata.type":     "#microsoft.graph.aadUserConversationMember",
				"roles":           []string{"owner"},
				"user@odata.bind": fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", senderID),
			},
			{
				"@odata.type":     "#microsoft.graph.aadUserConversationMember",
				"roles":           []string{"owner"},
				"user@odata.bind": fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s", recipientID),
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chats", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, readGraphError(resp)
	}

	var chat Chat
	return &chat, json.NewDecoder(resp.Body).Decode(&chat)
}

func (c *Client) SendMessage(ctx context.Context, chatID, message string) error {
	body := map[string]any{
		"body": map[string]string{
			"content": message,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/chats/%s/messages", c.baseURL, chatID),
		bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return readGraphError(resp)
	}

	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
}

func readGraphError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("graph API error %d: %s", resp.StatusCode, string(body))
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/graph/... -v
```

Expected: All five tests (`TestGetMe`, `TestGetUserByEmail`, `TestGetUserByEmail_NotFound`, `TestCreateOrGetChat`, `TestSendMessage`) PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/graph/
git commit -m "feat: add graph API client with httptest coverage"
```

---

### Task 4: Auth package

**Files:**
- Create: `internal/auth/auth.go`

The interactive browser auth flow cannot be unit-tested without a live Azure tenant. This task implements the package; the integration smoke test in Task 6 covers it end-to-end.

> **Token storage note:** `TokenCachePersistenceOptions` stores tokens in the OS credential store (macOS Keychain, Windows Credential Manager). With `AllowUnencryptedStorage: true` it falls back to a plaintext file if the OS store is unavailable. The cache location is managed by azidentity — you do not need to specify a path.

- [ ] **Step 1: Implement auth.go**

Create `internal/auth/auth.go`:

```go
package auth

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func NewCredential(clientID string) (*azidentity.InteractiveBrowserCredential, error) {
	return azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		ClientID: clientID,
		TenantID: "organizations",
		TokenCachePersistenceOptions: &azidentity.TokenCachePersistenceOptions{
			AllowUnencryptedStorage: true,
		},
	})
}

func GetToken(ctx context.Context, cred *azidentity.InteractiveBrowserCredential) (string, error) {
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return "", err
	}
	return token.Token, nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/auth/...
```

Expected: No output (compiles cleanly).

- [ ] **Step 3: Commit**

```bash
git add internal/auth/
git commit -m "feat: add auth package with interactive browser credential"
```

---

### Task 5: cobra commands

**Files:**
- Create: `cmd/send_test.go`
- Create: `cmd/root.go`
- Create: `cmd/send.go`

- [ ] **Step 1: Write the failing tests**

Create `cmd/send_test.go`:

```go
package cmd

import (
	"io"
	"testing"
)

func TestSendRequiresFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing both flags", []string{}},
		{"missing --message", []string{"--to", "user@example.com"}},
		{"missing --to", []string{"--message", "hi"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSendCmd()
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SilenceErrors(true)
			cmd.SilenceUsage(true)
			if err := cmd.Execute(); err == nil {
				t.Error("expected error for missing required flags, got nil")
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./cmd/...
```

Expected: FAIL — package `cmd` does not exist yet.

- [ ] **Step 3: Implement cmd/root.go**

Create `cmd/root.go`:

```go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wstevenson/msteams-cli/internal/config"
)

var clientID string

var rootCmd = &cobra.Command{
	Use:   "msteams-cli",
	Short: "Send Microsoft Teams messages from the command line",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if clientID != "" {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.ClientID = clientID
			return config.Save(cfg)
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "Azure app registration client ID (saved on first use)")
	rootCmd.AddCommand(newSendCmd())
}
```

- [ ] **Step 4: Implement cmd/send.go**

Create `cmd/send.go`:

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wstevenson/msteams-cli/internal/auth"
	"github.com/wstevenson/msteams-cli/internal/config"
	"github.com/wstevenson/msteams-cli/internal/graph"
)

func newSendCmd() *cobra.Command {
	var to, message string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a Teams message to a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSend(cmd.Context(), to, message)
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "recipient email address (required)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "message body (required)")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("message")

	return cmd
}

func runSend(ctx context.Context, to, message string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg.ClientID == "" {
		return fmt.Errorf("no client ID configured; run with --client-id <id> on first use")
	}

	cred, err := auth.NewCredential(cfg.ClientID)
	if err != nil {
		return fmt.Errorf("creating credential: %w", err)
	}

	token, err := auth.GetToken(ctx, cred)
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}

	client := graph.New(token)

	me, err := client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	recipient, err := client.GetUserByEmail(ctx, to)
	if err != nil {
		return fmt.Errorf("user not found %q: %w", to, err)
	}

	chat, err := client.CreateOrGetChat(ctx, me.ID, recipient.ID)
	if err != nil {
		return fmt.Errorf("creating chat: %w", err)
	}

	if err := client.SendMessage(ctx, chat.ID, message); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	fmt.Printf("Message sent to %s\n", to)
	return nil
}
```

- [ ] **Step 5: Run tests to confirm they pass**

```bash
go test ./cmd/... -v
```

Expected: All three `TestSendRequiresFlags` subtests PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/
git commit -m "feat: add cobra send command and root command"
```

---

### Task 6: Build, smoke test, and .gitignore

**Files:**
- Create: `.gitignore`

- [ ] **Step 1: Run full test suite**

```bash
go test ./...
```

Expected: All tests PASS, no failures.

- [ ] **Step 2: Build the binary**

```bash
go build -o msteams-cli .
```

Expected: `msteams-cli` binary produced in the project root.

- [ ] **Step 3: Verify --help output**

```bash
./msteams-cli --help
./msteams-cli send --help
```

Expected: Usage info for both commands. `send --help` shows `--to`, `--message`/`-m`, and the inherited `--client-id` flag.

- [ ] **Step 4: Verify missing required flag error**

```bash
./msteams-cli send --to user@example.com
```

Expected: `Error: required flag(s) "message" not set`

- [ ] **Step 5: Verify missing client ID error**

```bash
./msteams-cli send --to user@example.com --message "test"
```

Expected: `Error: no client ID configured; run with --client-id <id> on first use`

- [ ] **Step 6: Integration test (requires Azure app registration)**

Follow the Azure setup steps in [the design spec](../specs/2026-05-31-msteams-cli-design.md#azure-app-registration-one-time-manual-setup) to get a client ID, then:

```bash
./msteams-cli --client-id <YOUR_CLIENT_ID> send \
  --to your-colleague@yourorg.com \
  --message "test message from CLI"
```

Expected: Browser opens for Microsoft sign-in. After login, terminal prints:
```
Message sent to your-colleague@yourorg.com
```

- [ ] **Step 7: Verify token caching**

Run again without `--client-id` (client ID was saved to config on the previous run):

```bash
./msteams-cli send \
  --to your-colleague@yourorg.com \
  --message "second message — no browser prompt this time"
```

Expected: Message sends immediately, no browser prompt.

- [ ] **Step 8: Create .gitignore and final commit**

```bash
cat > .gitignore << 'EOF'
msteams-cli
EOF

git add .gitignore
git commit -m "chore: add .gitignore for compiled binary"
```
