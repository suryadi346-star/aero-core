package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

func HealthDeep(db *sql.DB, ollamaURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{"status": "ok"}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil { status["db"] = "error" } else { status["db"] = "ok" }
		resp, err := http.Get(ollamaURL + "/api/tags")
		if err != nil || resp.StatusCode != http.StatusOK {
			status["ollama"] = "error"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			status["ollama"] = "ok"
			resp.Body.Close()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
