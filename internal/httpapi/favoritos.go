package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// listarFavoritos retorna os favoritos do usuário logado.
func (srv *Server) listarFavoritos(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver favoritos")
		return
	}

	favoritos, err := srv.Eventos.ListarFavoritos(r.Context(), user.UID)
	if err != nil {
		srv.Logger.Error("listar favoritos falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"favoritos": favoritos})
}

// salvarFavorito adiciona um produto aos favoritos do usuário.
func (srv *Server) salvarFavorito(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para salvar favoritos")
		return
	}

	var fav store.Favorito
	if err := json.NewDecoder(r.Body).Decode(&fav); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if fav.ProdutoID == "" && fav.Nome == "" {
		writeErr(w, http.StatusBadRequest, "produto_id ou nome é obrigatório")
		return
	}

	fav.OwnerUID = user.UID
	fav.SalvoEm = time.Now().UTC()

	if err := srv.Eventos.SalvarFavorito(r.Context(), fav); err != nil {
		srv.Logger.Error("salvar favorito falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// removerFavorito remove um produto dos favoritos.
func (srv *Server) removerFavorito(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para remover favoritos")
		return
	}

	produtoID := r.URL.Query().Get("produto_id")
	if produtoID == "" {
		writeErr(w, http.StatusBadRequest, "parâmetro 'produto_id' é obrigatório")
		return
	}

	if err := srv.Eventos.RemoverFavorito(r.Context(), user.UID, produtoID); err != nil {
		srv.Logger.Error("remover favorito falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removido"})
}
