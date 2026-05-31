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
