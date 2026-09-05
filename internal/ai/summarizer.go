package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"not-jira/internal/config"

	"golang.org/x/net/proxy"
)

type Summarizer struct {
	cfg    config.AIConfig
	client *http.Client
}

type SummarizeResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionRequest struct {
	Model          string                  `json:"model"`
	Messages       []chatCompletionMessage `json:"messages"`
	Temperature    float64                 `json:"temperature"`
	ResponseFormat *responseFormat         `json:"response_format,omitempty"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func New(cfg config.AIConfig, proxyURL string) (*Summarizer, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if cfg.UseProxy && proxyURL != "" {
		pURL, err := url.Parse(proxyURL)
		if err == nil && pURL.Scheme == "socks5" {
			dialer, err := proxy.FromURL(pURL, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("failed to create socks5 dialer for AI: %w", err)
			}
			transport.DialContext = dialer.(proxy.ContextDialer).DialContext
		} else if err == nil && (pURL.Scheme == "http" || pURL.Scheme == "https") {
			transport.Proxy = http.ProxyURL(pURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   25 * time.Second,
	}

	return &Summarizer{
		cfg:    cfg,
		client: client,
	}, nil
}

func (s *Summarizer) Summarize(ctx context.Context, taskType string, text string) (*SummarizeResult, error) {
	if !s.cfg.Enabled || s.cfg.APIKey == "" {
		return nil, fmt.Errorf("AI summarizer is disabled or API key is not set")
	}

	systemPrompt := fmt.Sprintf(`Ты ассистент-трекер задач not-jira. Пользователь описывает %s.
Твоя задача — сжать входящий текст в лаконичную задачу.
Ответь ТОЛЬКО валидным JSON-объектом без лишних символов и без markdown-разметки:
{
  "title": "Краткий заголовок задачи (до 60 символов, без точки в конце)",
  "description": "Емкое, четкое описание сути проблемы или предложения в 1-3 предложениях"
}`, taskType)

	reqBody := chatCompletionRequest{
		Model: s.cfg.Model,
		Messages: []chatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
		Temperature:    0.2,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	apiEndpoint := strings.TrimRight(s.cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send AI request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read AI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("AI returned error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("AI returned empty choices")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	// Remove markdown code blocks if model wrapped it in ```json ... ```
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result SummarizeResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Fallback: parse first line as title, rest as description
		lines := strings.SplitN(content, "\n", 2)
		result.Title = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			result.Description = strings.TrimSpace(lines[1])
		}
	}

	if result.Title == "" {
		result.Title = "Задача"
	}
	if result.Description == "" {
		result.Description = text
	}

	return &result, nil
}
