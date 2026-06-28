package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// listarDestinos devolve os destinos ativos.
func (srv *Server) listarDestinos(w http.ResponseWriter, r *http.Request) {
	lista, err := srv.Repo.Destinos().ListarDestinos(r.Context())
	if err != nil {
		srv.Logger.Error("listar destinos falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"destinos": lista})
}

// salvarDestino cria ou atualiza um destino de publicação.
func (srv *Server) salvarDestino(w http.ResponseWriter, r *http.Request) {
	var d store.Destino
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if d.Config == "" {
		writeErr(w, http.StatusBadRequest, "config é obrigatório (chat_id, telefone, etc.)")
		return
	}
	if d.Nome == "" {
		writeErr(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}
	if d.Tipo == "" {
		d.Tipo = "telegram"
	}
	if d.ID == "" {
		d.ID = slugificarDestino(d.Nome)
	}
	d.Ativo = true

	if err := srv.Repo.Destinos().SalvarDestino(r.Context(), d); err != nil {
		srv.Logger.Error("salvar destino falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	srv.Logger.Info("destino salvo", slog.String("id", d.ID), slog.String("tipo", d.Tipo))
	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "destino": d})
}

// deletarDestino remove um destino por ID (?id=xxx).
func (srv *Server) deletarDestino(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "informe ?id=")
		return
	}
	if err := srv.Repo.Destinos().DeletarDestino(r.Context(), id); err != nil {
		srv.Logger.Error("deletar destino falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	srv.Logger.Info("destino removido", slog.String("id", id))
	writeJSON(w, http.StatusOK, map[string]any{"status": "removido", "id": id})
}

// slugificarDestino gera um ID slug simples a partir do nome.
func slugificarDestino(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var out []rune
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			out = append(out, r)
		case r == ' ' || r == '_':
			out = append(out, '-')
		}
	}
	result := strings.Trim(string(out), "-")
	if result == "" {
		return "destino"
	}
	return result
}
