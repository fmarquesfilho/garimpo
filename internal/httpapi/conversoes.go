package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
)

// conversoes retorna o relatório de publicações agrupado por canal/sub_id.
func (srv *Server) conversoes(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver conversões")
		return
	}
	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}
	conversoes, err := srv.Eventos.Conversoes(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("conversoes falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"conversoes": conversoes, "dias": dias})
}
