package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

// adminLogs retorna os últimos logs capturados pelo buffer.
// GET /api/admin/logs?n=100&nivel=error
func (srv *Server) adminLogs(w http.ResponseWriter, r *http.Request) {

	if srv.LogBuffer == nil {
		writeJSON(w, http.StatusOK, map[string]any{"logs": []any{}, "stats": map[string]int{}})
		return
	}

	n := 100
	if s := r.URL.Query().Get("n"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			n = v
		}
	}

	entries := srv.LogBuffer.Ultimos(n)

	// Filtro opcional por nível
	nivel := r.URL.Query().Get("nivel")
	if nivel != "" {
		var filtrado []any
		for _, e := range entries {
			if e.Nivel == nivel {
				filtrado = append(filtrado, e)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"logs":  filtrado,
			"stats": srv.LogBuffer.Stats(),
			"total": srv.LogBuffer.Total(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"logs":  entries,
		"stats": srv.LogBuffer.Stats(),
		"total": srv.LogBuffer.Total(),
	})
}

// adminLogLevel permite mudar o nível de log em runtime.
// POST /api/admin/log-level  body: {"nivel": "debug|info|warn|error"}
func (srv *Server) adminLogLevel(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Nivel string `json:"nivel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	// Atualiza o nível do logger via novo handler
	nivel := parseLogLevel(req.Nivel)
	opts := &slog.HandlerOptions{Level: nivel}
	var h slog.Handler
	if os.Getenv("K_SERVICE") != "" || os.Getenv("LOG_FORMAT") == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	newLogger := slog.New(h)
	srv.Logger = newLogger
	slog.SetDefault(newLogger)

	srv.Logger.Info("log-level alterado", slog.String("nivel", req.Nivel))
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "nivel": req.Nivel})
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// adminMe retorna informações do usuário logado (incluindo se é admin).
// GET /api/admin/me
func (srv *Server) adminMe(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"uid":   user.UID,
		"email": user.Email,
		"admin": user.Admin,
	})
}
