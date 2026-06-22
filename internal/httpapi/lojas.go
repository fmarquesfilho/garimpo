package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
)

// novidades compara os últimos dois snapshots de uma busca com lojas e retorna:
// - produtos novos (presentes no último snapshot mas não no anterior)
// - variações de preço (mesmo produto_id com preço diferente)
func (srv *Server) novidades(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver novidades")
		return
	}

	buscaID := r.URL.Query().Get("busca_id")
	dias := 7
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}

	novidades, err := srv.Eventos.Novidades(r.Context(), buscaID, dias)
	if err != nil {
		srv.Logger.Error("novidades falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, novidades)
}
