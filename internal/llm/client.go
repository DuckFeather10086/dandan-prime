//go:build !js && !wasm
// +build !js,!wasm

// Package llm is a minimal OpenAI-compatible chat client.
//
// It exists so the matcher can talk to whatever endpoint is configured —
// agyproxy on this box, DeepSeek, anything speaking /v1/chat/completions —
// without pulling in a vendor SDK. The matcher only ever asks for a small
// JSON object back, so streaming, tools and multi-turn are all out of scope.
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func New(baseURL, apiKey, model string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTP:    &http.Client{Timeout: 180 * time.Second},
	}
}

// Configured reports whether there is enough config to make a call. The
// resolver checks this so a library scan degrades to "leave it unmatched"
// instead of erroring out when no key is present.
func (c *Client) Configured() bool {
	return c != nil && c.BaseURL != "" && c.Model != ""
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	Messages    []message `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

var fenceRe = regexp.MustCompile("(?m)^```(?:json)?\\s*$")

// Chat sends one system+user pair at temperature 0 and returns the assistant
// text with any ``` fence stripped, so the caller can json.Unmarshal it
// directly. Models wrap JSON in a fence often enough that handling it here
// beats handling it at every call site.
func (c *Client) Chat(system, user string) (string, Usage, error) {
	if !c.Configured() {
		return "", Usage{}, fmt.Errorf("llm: not configured")
	}

	payload, err := json.Marshal(chatRequest{
		Model:       c.Model,
		Temperature: 0,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", Usage{}, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", Usage{}, err
	}
	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Usage{}, err
	}

	var parsed chatResponse
	if err := json.Unmarshal(bodyData, &parsed); err != nil {
		return "", Usage{}, fmt.Errorf("llm: status %d: unparseable body: %s", resp.StatusCode, snippet(bodyData))
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", Usage{}, fmt.Errorf("llm: %s", parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", Usage{}, fmt.Errorf("llm: status %d: %s", resp.StatusCode, snippet(bodyData))
	}
	if len(parsed.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("llm: no choices returned")
	}

	usage := Usage{parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens}
	text := strings.TrimSpace(fenceRe.ReplaceAllString(parsed.Choices[0].Message.Content, ""))
	return text, usage, nil
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}
