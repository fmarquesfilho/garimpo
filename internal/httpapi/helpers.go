package httpapi

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/fmarquesfilho/garimpo/internal/auth"
)

// ── Serialização ─────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeErr retorna um erro no formato RFC 9457 (Problem Details).
func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   problemTypeFromStatus(status),
		"title":  problemTitleFromStatus(status),
		"status": status,
		"detail": msg,
		"erro":   msg, // compatibilidade com frontend existente
	})
}

func problemTypeFromStatus(status int) string {
	switch {
	case status == 400:
		return "https://garimpei.app.br/problemas/entrada-invalida"
	case status == 401:
		return "https://garimpei.app.br/problemas/nao-autenticado"
	case status == 403:
		return "https://garimpei.app.br/problemas/sem-permissao"
	case status == 404:
		return "https://garimpei.app.br/problemas/nao-encontrado"
	case status == 409:
		return "https://garimpei.app.br/problemas/conflito"
	case status == 502:
		return "https://garimpei.app.br/problemas/servico-externo"
	case status == 503:
		return "https://garimpei.app.br/problemas/indisponivel"
	default:
		return "about:blank"
	}
}

func problemTitleFromStatus(status int) string {
	switch status {
	case 400:
		return "Dados inválidos"
	case 401:
		return "Não autenticado"
	case 403:
		return "Acesso negado"
	case 404:
		return "Não encontrado"
	case 409:
		return "Conflito"
	case 502:
		return "Serviço externo indisponível"
	case 503:
		return "Serviço temporariamente indisponível"
	case 500:
		return "Erro interno"
	default:
		return http.StatusText(status)
	}
}

// ── Auth helpers ─────────────────────────────────────────────────────────────

func (srv *Server) usuarioDoRequest(r *http.Request) *auth.User {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil
	}
	return srv.Auth.Verify(r.Context(), token)
}

func (srv *Server) autorizadoColeta(r *http.Request) bool {
	tok := os.Getenv("COLETA_TOKEN")
	if tok == "" {
		return true
	}
	return r.Header.Get("X-Garimpo-Token") == tok
}
