package llm

import (
	"context"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Client wraps an OpenAI-compatible API client.
type Client struct {
	client *openai.Client
	model  string
}

// NewDeepSeekClient creates a client configured for DeepSeek's API.
func NewDeepSeekClient(apiKey string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com/v1"
	return &Client{
		client: openai.NewClientWithConfig(config),
		model:  "deepseek-chat",
	}
}

// ChatMessage represents a message in the conversation.
type ChatMessage struct {
	Role       string // system, user, assistant, tool
	Content    string
	ToolCallID string     // for tool role messages
	ToolCalls  []ToolCall // for assistant messages with tool calls
}

// ToolCall represents a function call from the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // raw JSON
}

// Tool defines a function the model can call.
type Tool struct {
	Name        string
	Description string
	Parameters  jsonschema.Definition
}

// ChatResponse holds the model's response.
type ChatResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// Chat sends a conversation to the API and returns the response.
func (c *Client) Chat(ctx context.Context, messages []ChatMessage, tools []Tool) (*ChatResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Convert messages
	var msgs []openai.ChatCompletionMessage
	for _, m := range messages {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				})
			}
		}
		msgs = append(msgs, msg)
	}

	// Convert tools
	var oaiTools []openai.Tool
	for _, t := range tools {
		oaiTools = append(oaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: msgs,
	}
	if len(oaiTools) > 0 {
		req.Tools = oaiTools
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	choice := resp.Choices[0]
	result := &ChatResponse{
		Content: choice.Message.Content,
	}

	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}
