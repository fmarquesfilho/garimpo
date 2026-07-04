package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	publisherpb "github.com/fmarquesfilho/garimpo/gen/go/publisher/v1"
	"github.com/fmarquesfilho/garimpo/internal/taskqueue"
)

// errAnalyzerStatus is returned when the analyzer responds with non-200.
var errAnalyzerStatus = errors.New("analyzer non-200 status")

// ─── HTTP handler for Cloud Tasks alert processing ──────────────────────────

// StartHTTP starts an HTTP server on the given port for receiving Cloud Tasks.
// This runs alongside the gRPC server.
func (s *SchedulerServer) StartHTTP(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /process-alert", s.handleProcessAlert)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.logger.Info("scheduler HTTP listening", slog.String("port", port))
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("HTTP server error", slog.String("error", err.Error()))
	}
}

// handleProcessAlert is called by Cloud Tasks to process a price alert.
// Flow: receive payload → call analyzer /quedas → format message → call publisher gRPC.
func (s *SchedulerServer) handleProcessAlert(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var payload taskqueue.AlertPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if payload.Keyword == "" || payload.ChatID == "" {
		http.Error(w, "keyword and chat_id required", http.StatusBadRequest)
		return
	}

	s.logger.Info("processing alert",
		slog.String("keyword", payload.Keyword),
		slog.String("chat_id", payload.ChatID),
		slog.Float64("threshold", payload.Threshold))

	// 1. Call analyzer to get price drops
	quedas, err := s.fetchQuedas(payload.Keyword, payload.Threshold)
	if err != nil {
		s.logger.Warn("analyzer unavailable for alert", slog.String("error", err.Error()))
		// Return 200 so Cloud Tasks doesn't retry (analyzer issue, not transient)
		writeJSON(w, http.StatusOK, map[string]any{"alerts_sent": 0, "reason": "analyzer_unavailable"})
		return
	}

	if len(quedas) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"alerts_sent": 0, "keyword": payload.Keyword})
		return
	}

	// 2. Format alert message
	msg := formatAlertMessage(payload.Keyword, quedas)

	// 3. Send via publisher gRPC
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	resp, err := s.publisher.Publish(ctx, &publisherpb.PublishRequest{
		Channel: "telegram",
		GroupId: payload.ChatID,
		Content: &publisherpb.PublishContent{
			Title:       msg,
			Description: "", // HTML is in Title
		},
	})
	if err != nil {
		s.logger.Error("publisher failed", slog.String("error", err.Error()))
		// Return 500 so Cloud Tasks retries
		http.Error(w, "publisher error", http.StatusInternalServerError)
		return
	}

	s.logger.Info("alert sent",
		slog.String("keyword", payload.Keyword),
		slog.Int("drops", len(quedas)),
		slog.Bool("success", resp.GetSuccess()))

	writeJSON(w, http.StatusOK, map[string]any{
		"alerts_sent": 1,
		"drops":       len(quedas),
		"keyword":     payload.Keyword,
		"message_id":  resp.GetMessageId(),
	})
}

// ─── Analyzer client ────────────────────────────────────────────────────────

type quedaItem struct {
	ProdutoID     string  `json:"produto_id"`
	Nome          string  `json:"nome"`
	PrecoAnterior float64 `json:"preco_anterior"`
	PrecoAtual    float64 `json:"preco_atual"`
	Variacao      float64 `json:"variacao"`
}

type quedasResponse struct {
	Quedas []quedaItem `json:"quedas"`
}

func (s *SchedulerServer) fetchQuedas(keyword string, threshold float64) ([]quedaItem, error) {
	url := fmt.Sprintf("%s/quedas?dias=2&threshold=%.2f&limit=10", s.analyzerURL, threshold)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("analyzer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %w", resp.StatusCode, errAnalyzerStatus)
	}

	var result quedasResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Quedas, nil
}

// ─── Message formatting ─────────────────────────────────────────────────────

func formatAlertMessage(keyword string, quedas []quedaItem) string {
	var b strings.Builder
	b.WriteString("🔔 <b>Alerta de Preço</b>\n")
	b.WriteString(fmt.Sprintf("🏪 <code>%s</code>\n\n", keyword))

	limit := 10
	if len(quedas) < limit {
		limit = len(quedas)
	}

	for i := range limit {
		q := quedas[i]
		pct := math.Abs(q.Variacao * 100)
		nome := q.Nome
		if len(nome) > 40 {
			nome = nome[:39] + "…"
		}
		b.WriteString(fmt.Sprintf("📉 <b>%s</b>\n", nome))
		b.WriteString(fmt.Sprintf("   R$ %.2f → R$ %.2f (↓%.1f%%)\n\n", q.PrecoAnterior, q.PrecoAtual, pct))
	}

	if len(quedas) > 10 {
		b.WriteString(fmt.Sprintf("<i>...e mais %d quedas</i>\n", len(quedas)-10))
	}

	b.WriteString(fmt.Sprintf("⏰ %s", time.Now().Format("02/01 15:04")))
	return b.String()
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck // best-effort response
}
