package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"github.com/suryadi346-star/aero-core/internal/cache"
	"github.com/suryadi346-star/aero-core/internal/model"
	"github.com/suryadi346-star/aero-core/internal/session"
)

type ChatHandler struct {
	Model model.Provider
	Store *session.Store
	Cache *cache.SQLiteCache
	Mu    sync.Mutex
}

func NewChatHandler(m model.Provider, s *session.Store, c *cache.SQLiteCache) *ChatHandler {
	return &ChatHandler{Model: m, Store: s, Cache: c}
}

func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	defer r.Body.Close()

	var req struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
		Model     string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSSEError(w, "invalid_payload"); return
	}
	if req.Message == "" { writeSSEError(w, "empty_message"); return }
	if req.SessionID == "" { writeSSEError(w, "missing_session_id"); return }

	modelName := req.Model
	if modelName == "" { modelName = "qwen2.5:1.5b-instruct-q4_k_m" }

	ctx := r.Context()
	sess := h.Store.Get(req.SessionID)
	if sess == nil {
		dbSess, err := h.Cache.LoadSession(req.SessionID)
		if err != nil { slog.Warn("db load failed", "err", err) }
		sess = dbSess
		if sess == nil {
			sess = &session.Session{ID: req.SessionID, Messages: []session.Message{
				{Role: "system", Content: "You are a helpful assistant. Keep responses concise and safe."},
			}}
		}
		h.Store.Save(sess)
	}

	promptHash := cache.PromptHash(sess.Messages[0].Content, req.Message)
	cachedResp, _ := h.Cache.GetCachedResponse(promptHash)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok { http.Error(w, "streaming not supported", http.StatusNotImplemented); return }

	if cachedResp != "" {
		h.streamCached(w, flusher, cachedResp); return
	}

	sess.Messages = append(sess.Messages, session.Message{Role: "user", Content: req.Message})
	h.Store.Save(sess)
	contextMsgs := sess.PrepareContext(2048)

	stream, errCh := h.Model.Stream(ctx, model.ChatRequest{Model: modelName, Messages: contextMsgs})
	var fullResponse strings.Builder
	streamDone := false

	for {
		select {
		case <-ctx.Done():
			writeSSEEvent(w, flusher, "error", `{"reason":"client_disconnect"}`); return
		case err := <-errCh:
			if err != nil { writeSSEEvent(w, flusher, "error", fmt.Sprintf(`{"error":"%s"}`, sanitizeJSON(err.Error()))) }
			return
		case chunk, ok := <-stream:
			if !ok { streamDone = true; break }
			if chunk.Done { streamDone = true; break }
			fullResponse.WriteString(chunk.Content)
			writeSSEEvent(w, flusher, "chunk", fmt.Sprintf(`{"content":"%s"}`, sanitizeJSON(chunk.Content)))
		}
		if streamDone { break }
	}

	sess.Messages = append(sess.Messages, session.Message{Role: "assistant", Content: fullResponse.String()})
	h.Store.Save(sess)

	go func() {
		h.Cache.CacheResponse(promptHash, fullResponse.String())
		h.Mu.Lock(); defer h.Mu.Unlock()
		h.Cache.PruneCache()
	}()
	writeSSEEvent(w, flusher, "done", `{}`)
}

func (h *ChatHandler) streamCached(w http.ResponseWriter, f http.Flusher, resp string) {
	for _, c := range resp {
		writeSSEEvent(w, f, "chunk", fmt.Sprintf(`{"content":"%s"}`, string(c)))
	}
	writeSSEEvent(w, f, "done", `{}`)
}

func writeSSEEvent(w http.ResponseWriter, f http.Flusher, event, data string) {
	fmt.Fprintf(w, "event: %s\n %s\n\n", event, data)
	f.Flush()
}

func writeSSEError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `{"error":"%s"}`, msg)
}

func sanitizeJSON(s string) string {
	b, _ := json.Marshal(s)
	if len(b) < 2 { return "" }
	return string(b[1 : len(b)-1])
}
