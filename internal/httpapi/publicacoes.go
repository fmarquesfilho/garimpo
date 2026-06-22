package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// listarPublicacoes retorna publicações filtradas por status (?status=agendada|enviada|erro).
func (srv *Server) listarPublicacoes(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para ver publicações")
		return
	}
	status := r.URL.Query().Get("status") // vazio = todas
	lista, err := srv.Eventos.ListarPublicacoes(r.Context(), status)
	if err != nil {
		srv.Logger.Error("listar publicacoes falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"publicacoes": lista})
}

// agendarPublicacao cria uma publicação agendada (ou imediata se agendada_em está vazio).
func (srv *Server) agendarPublicacao(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para agendar publicações")
		return
	}

	var req struct {
		ProdutoID  string  `json:"produto_id"`
		Nome       string  `json:"nome"`
		Categoria  string  `json:"categoria"`
		Preco      float64 `json:"preco"`
		Comissao   float64 `json:"comissao"`
		Link       string  `json:"link"`
		Imagem     string  `json:"imagem"`
		Estrategia string  `json:"estrategia"`
		DestinoID  string  `json:"destino_id"`
		TemplateID string  `json:"template_id"`
		AgendadaEm string  `json:"agendada_em"` // ISO 8601; vazio = enviar agora
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if req.Nome == "" {
		writeErr(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}

	agora := time.Now().UTC()
	pub := store.Publicacao{
		ID:         generateID(req.ProdutoID, agora),
		ProdutoID:  req.ProdutoID,
		Nome:       req.Nome,
		Categoria:  req.Categoria,
		Preco:      req.Preco,
		Comissao:   req.Comissao,
		Link:       req.Link,
		Imagem:     req.Imagem,
		Estrategia: req.Estrategia,
		DestinoID:  req.DestinoID,
		TemplateID: req.TemplateID,
		AgendadaEm: req.AgendadaEm,
		Status:     "agendada",
		CriadaEm:   agora.Format(time.RFC3339),
		OwnerUID:   user.UID,
	}

	// Se não tem agendada_em, publica imediatamente
	if req.AgendadaEm == "" {
		pub.Status = "enviada"
		oferta := publish.Oferta{
			ProdutoID: req.ProdutoID, Nome: req.Nome, Categoria: req.Categoria,
			Preco: req.Preco, Comissao: req.Comissao, Link: req.Link, Imagem: req.Imagem,
			Estrategia: req.Estrategia, DestinoID: req.DestinoID, TemplateID: req.TemplateID,
		}

		// Aplica template (se com_foto=false, remove imagem)
		if req.TemplateID != "" && srv.Templates != nil {
			tmpl, err := srv.Templates.Buscar(r.Context(), req.TemplateID)
			if err == nil && !tmpl.ComFoto {
				oferta.Imagem = ""
			}
		}

		res, err := srv.Publicador.Publicar(r.Context(), oferta)
		if err != nil {
			pub.Status = "erro"
			pub.Detalhe = err.Error()
		} else {
			pub.Detalhe = publish.SubID(res.Canal, req.Estrategia, agora)
			pub.EnviadaEm = agora.Format(time.RFC3339)

			// Registra no BigQuery (best-effort)
			_ = srv.Eventos.Registrar(r.Context(), store.Evento{
				Tipo: "publicacao", Canal: res.Canal, SubID: pub.Detalhe,
				ProdutoID: req.ProdutoID, Nome: req.Nome, Categoria: req.Categoria,
				Estrategia: req.Estrategia, Comissao: req.Comissao, Preco: req.Preco,
			})
		}
	}

	// Persiste a publicação
	_ = srv.Eventos.SalvarPublicacao(r.Context(), pub)

	srv.Logger.Info("publicacao criada",
		slog.String("id", pub.ID),
		slog.String("status", pub.Status),
		slog.String("agendada_em", pub.AgendadaEm),
	)

	writeJSON(w, http.StatusCreated, map[string]any{"publicacao": pub})
}

func generateID(produtoID string, t time.Time) string {
	return produtoID + "-" + t.Format("20060102150405")
}
