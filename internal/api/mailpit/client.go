package mailpit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailAddress struct {
	Address string `json:"Address"`
	Name    string `json:"Name"`
}

type Message struct {
	ID      string         `json:"ID"`
	From    EmailAddress   `json:"From"`
	To      []EmailAddress `json:"To"`
	Subject string         `json:"Subject"`
	Snippet string         `json:"Snippet"`
}

type MessagesResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
}

type MessageDetail struct {
	ID      string         `json:"ID"`
	From    EmailAddress   `json:"From"`
	To      []EmailAddress `json:"To"`
	Subject string         `json:"Subject"`
	Text    string         `json:"Text"`
}

type deleteRequest struct {
	IDs []string `json:"IDs"`
}

type Client struct {
	BaseURL string
}

// todo: pass logger and correctly randle resp.Body.Close()

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// GetMessages returns all messages from Mailpit
func (c *Client) GetMessages() ([]Message, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/messages", c.BaseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var messagesResp MessagesResponse
	err = json.NewDecoder(resp.Body).Decode(&messagesResp)
	if err != nil {
		return nil, err
	}

	return messagesResp.Messages, nil
}

// GetMessageDetail returns detailed message info by ID
func (c *Client) GetMessageDetail(id string) (*MessageDetail, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/message/%s", c.BaseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var detail MessageDetail
	err = json.NewDecoder(resp.Body).Decode(&detail)
	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (c *Client) DeleteMessages(ids []string) error {
	body, err := json.Marshal(deleteRequest{IDs: ids})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/messages", c.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete messages: status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DeleteAllMessages() error {
	body, err := json.Marshal(deleteRequest{IDs: []string{}})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/messages", c.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete all messages: status %d", resp.StatusCode)
	}

	return nil
}
