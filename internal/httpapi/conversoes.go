package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

// conversoesReais consulta o conversionReport da Shopee e retorna vendas reais.
// Protegido por autenticação normal (Bearer token), não por COLETA_TOKEN.
// GET /api/conversoes/reais?dias=30
func (srv *Server) conversoesReais(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver conversões")
		return
	}

	appID := os.Getenv("SHOPEE_APP_ID")
	secret := os.Getenv("SHOPEE_SECRET")
	if appID == "" || secret == "" {
		writeErr(w, http.StatusServiceUnavailable, "credenciais Shopee não configuradas no servidor")
		return
	}

	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 90 {
			dias = v
		}
	}

	conversoes, err := buscarConversoesShopee(appID, secret, dias)
	if err != nil {
		srv.Logger.Error("conversoes reais falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, "falha ao consultar Shopee: "+err.Error())
		return
	}

	// Calcular totais
	var totalComissao float64
	pendentes := 0
	confirmadas := 0
	for _, c := range conversoes {
		totalComissao += c.TotalCommission
		switch c.Status {
		case "PENDING", "UNPAID":
			pendentes++
		case "COMPLETED", "PAID":
			confirmadas++
		}
	}

	srv.Logger.Info("conversoes reais consultadas",
		slog.Int("total", len(conversoes)),
		slog.Int("dias", dias),
		slog.Float64("comissao_total", totalComissao),
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"dias":           dias,
		"total":          len(conversoes),
		"pendentes":      pendentes,
		"confirmadas":    confirmadas,
		"comissao_total": totalComissao,
		"conversoes":     conversoes,
	})
}

// conversoes retorna o relatório de conversões estimadas (baseado em publicações).
// GET /api/conversoes?dias=30
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
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversoes": conversoes,
		"dias":       dias,
	})
}
