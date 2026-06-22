package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
)

// conversoes retorna o relatório de publicações com atribuição, permitindo
// à usuária ver quais destinos (Telegram, WhatsApp) geraram conversão.
//
// O dado vem dos eventos tipo="publicacao" no BigQuery, cruzado com o sub_id
// que identifica canal+estrategia+data. Para conversão real (comissão paga),
// precisa de um webhook ou poll no conversion API da Shopee — isso é um passo
// futuro; por ora, mostra o histórico de publicações por destino.
//
//	GET /api/conversoes?dias=30
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

	relatorio, err := srv.Eventos.Conversoes(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("conversoes falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversoes":  relatorio,
		"dias_janela": dias,
	})
}
