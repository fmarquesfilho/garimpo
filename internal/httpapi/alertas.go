package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/fmarquesfilho/garimpo/internal/alerts"
)

// alertasConfig retorna a configuração atual de alertas (sem expor o token).
func (srv *Server) alertasConfig(w http.ResponseWriter, r *http.Request) {

	cfg := alerts.ConfigFromEnv()
	// Mostra chat_id mascarado se configurado
	chatID := cfg.TelegramChatID
	mascarado := ""
	if chatID != "" {
		if len(chatID) > 4 {
			mascarado = chatID[:4] + "***"
		} else {
			mascarado = "***"
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ativo":          cfg.Ativo(),
		"chat_id":        mascarado,
		"threshold":      cfg.Threshold,
		"apenas_quedas":  cfg.ApenasQuedas,
		"telegram_token": cfg.TelegramToken != "",
	})
}

// alertasTestar envia um alerta de teste para o grupo configurado.
func (srv *Server) alertasTestar(w http.ResponseWriter, r *http.Request) {

	cfg := alerts.ConfigFromEnv()
	cfg.Logger = srv.Logger
	if !cfg.Ativo() {
		writeErr(w, http.StatusBadRequest, "alertas não configurados. Defina ALERTAS_TELEGRAM_CHAT_ID e TELEGRAM_BOT_TOKEN nas variáveis de ambiente.")
		return
	}

	// Aceita um busca_id opcional para enviar alertas reais de uma loja
	var req struct {
		BuscaID string `json:"busca_id"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	alerter := alerts.Novo(cfg)

	if req.BuscaID != "" {
		// Envia alertas reais da busca
		alerter.VerificarENotificar(r.Context(), srv.Repo.Snapshots(), req.BuscaID)
		alerter.VerificarNovos(r.Context(), srv.Repo.Snapshots(), req.BuscaID)
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "alertas verificados e enviados (se houver variações)",
			"busca":  req.BuscaID,
		})
	} else {
		// Envia mensagem de teste
		err := alerter.EnviarTeste(r.Context())
		if err != nil {
			srv.Logger.Error("teste alerta falhou", slog.String("erro", err.Error()))
			writeErr(w, http.StatusBadGateway, "falha ao enviar teste: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "mensagem de teste enviada"})
	}
}

// alertasAtualizar permite atualizar threshold e apenas_quedas em runtime (via env override).
// Em produção, as env vars são definidas no Cloud Run; isso serve como override temporário para testes.
func (srv *Server) alertasAtualizar(w http.ResponseWriter, r *http.Request) {

	var req struct {
		ChatID       string  `json:"chat_id"`
		Threshold    float64 `json:"threshold"`
		ApenasQuedas *bool   `json:"apenas_quedas"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	if req.ChatID != "" {
		srv.mu.Lock()
		srv.alertasChatIDOverride = req.ChatID
		srv.mu.Unlock()
		os.Setenv("ALERTAS_TELEGRAM_CHAT_ID", req.ChatID)
		srv.Logger.Info("alertas: chat_id atualizado", slog.String("chat_id", req.ChatID))
	}
	if req.Threshold > 0 && req.Threshold < 1 {
		srv.mu.Lock()
		srv.alertasThresholdOverride = req.Threshold
		srv.mu.Unlock()
		os.Setenv("ALERTAS_THRESHOLD", strconv.FormatFloat(req.Threshold, 'f', 2, 64))
		srv.Logger.Info("alertas: threshold atualizado", slog.Float64("threshold", req.Threshold))
	}
	if req.ApenasQuedas != nil {
		srv.mu.Lock()
		srv.alertasApenasQuedasOverride = req.ApenasQuedas
		srv.mu.Unlock()
		val := "false"
		if *req.ApenasQuedas {
			val = "true"
		}
		os.Setenv("ALERTAS_APENAS_QUEDAS", val)
	}

	cfg := alerts.ConfigFromEnv()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "atualizado",
		"ativo":         cfg.Ativo(),
		"threshold":     cfg.Threshold,
		"apenas_quedas": cfg.ApenasQuedas,
	})
}
