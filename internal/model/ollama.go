package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaClient struct {
	baseURL string
	client  *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        5,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

type ollamaResp struct {
	Message struct { Content string `json:"content"` } `json:"message"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

func (c *OllamaClient) Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
	out := make(chan StreamChunk, 16)
	errCh := make(chan error, 1)
	payload, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		errCh <- fmt.Errorf("http: %w", err)
		close(out)
		return out, errCh
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		errCh <- fmt.Errorf("ollama http %d", resp.StatusCode)
		close(out)
		return out, errCh
	}

	go func() {
		defer resp.Body.Close()
		defer close(out)
		dec := json.NewDecoder(resp.Body)
		var buf ollamaResp
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}
			if err := dec.Decode(&buf); err != nil {
				if err == io.EOF { return }
				errCh <- fmt.Errorf("decode: %w", err)
				return
			}
			if buf.Error != "" {
				errCh <- fmt.Errorf("ollama: %s", buf.Error)
				return
			}
			if buf.Message.Content != "" {
				select { case out <- StreamChunk{Content: buf.Message.Content, Done: false}: case <-ctx.Done(): return }
			}
			if buf.Done {
				select { case out <- StreamChunk{Done: true}: case <-ctx.Done(): return }
				return
			}
		}
	}()
	return out, errCh
}
