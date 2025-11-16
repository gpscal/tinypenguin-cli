package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultTinyllamaURL = "http://localhost:11434/v1"
	DefaultTimeout      = 30 * time.Second
)

// TinyllamaClient handles communication with the tinyllama API
type TinyllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model    string      `json:"model"`
	Messages []Message   `json:"messages"`
	Stream   bool        `json:"stream,omitempty"`
	Tools    []Tool      `json:"tools,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string     `json:"role"`
	Content string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Tool represents a function tool definition
type Tool struct {
	Type     string     `json:"type"`
	Function Function   `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call from the model
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int           `json:"index"`
	Message      Message       `json:"message"`
	Logprobs     interface{}   `json:"logprobs,omitempty"`
	FinishReason string        `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewTinyllamaClient creates a new tinyllama client
func NewTinyllamaClient(baseURL string) *TinyllamaClient {
	if baseURL == "" {
		baseURL = DefaultTinyllamaURL
	}
	
	return &TinyllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// Chat creates a chat completion
func (c *TinyllamaClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &chatResp, nil
}

// Generate creates a text generation
type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream,omitempty"`
}

type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Context   []int  `json:"context,omitempty"`
	TotalDuration     int64 `json:"total_duration"`
	LoadDuration      int64 `json:"load_duration"`
	PromptEvalCount   int   `json:"prompt_eval_count"`
	PromptEvalDuration int64 `json:"prompt_eval_duration"`
	EvalCount          int   `json:"eval_count"`
	EvalDuration       int64 `json:"eval_duration"`
}

// Generate creates a text generation
func (c *TinyllamaClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	url := fmt.Sprintf("%s/generate", c.baseURL)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &genResp, nil
}

// ListModels lists available models
type ModelList struct {
	Models []ModelInfo `json:"models"`
}

type ModelInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

func (c *TinyllamaClient) ListModels(ctx context.Context) (*ModelList, error) {
	url := fmt.Sprintf("%s/models", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var modelList ModelList
	if err := json.NewDecoder(resp.Body).Decode(&modelList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &modelList, nil
}

// CreateToolDefinition converts a tool definition to the format expected by tinyllama
func CreateToolDefinition(name, description string, parameters map[string]interface{}) Tool {
	return Tool{
		Type: "function",
		Function: Function{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}

// CreateToolCall creates a tool call from the model response
func CreateToolCall(id, name, arguments string) ToolCall {
	return ToolCall{
		ID:   id,
		Type: "function",
		Function: FunctionCall{
			Name:      name,
			Arguments: arguments,
		},
	}
}