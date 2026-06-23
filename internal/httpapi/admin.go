package httpapi

import (
	"net/http"
	"strconv"
)

// adminLogs retorna os últimos logs capturados pelo buffer.
// GET /api/admin/logs?n=100&nivel=error
func (srv *Server) adminLogs(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para acessar admin")
		return
	}

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
