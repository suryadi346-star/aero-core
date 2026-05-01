package model

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Options  map[string]any `json:"options,omitempty"`
}

type StreamChunk struct {
	Content string
	Done    bool
}

type Provider interface {
	Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error)
}
