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
