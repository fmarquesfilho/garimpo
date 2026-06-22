package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/publish"
)

// templates gerencia os modelos de mensagem (GET lista, POST salva, DELETE remove).
//
//	GET    /api/templates         -> lista templates ativos
//	POST   /api/templates         -> cria/atualiza um template
//	DELETE /api/templates?id=xxx  -> remove um template
//	POST   /api/templates/preview -> renderiza um template com dados fictícios
func (srv *Server) templates(w http.ResponseWriter, r *http.Request) {
	// Preview não exige auth (útil para testar)
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/preview") {
		srv.templatePreview(w, r)
		return
	}

	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para gerenciar templates")
		return
	}

	switch r.Method {
	case http.MethodGet:
		lista, err := srv.Templates.Listar(r.Context())
		if err != nil {
			srv.Logger.Error("listar templates falhou", slog.String("erro", err.Error()))
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"templates": lista})

	case http.MethodPost:
		var t publish.Template
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeErr(w, http.StatusBadRequest, "json inválido")
			return
		}
		if t.Nome == "" {
			writeErr(w, http.StatusBadRequest, "nome é obrigatório")
			return
		}
		if t.Corpo == "" {
			writeErr(w, http.StatusBadRequest, "corpo é obrigatório")
			return
		}
		if t.ID == "" {
			t.ID = slugificarTemplate(t.Nome)
		}
		t.Ativo = true
		if t.CriadoEm == "" {
			t.CriadoEm = time.Now().UTC().Format(time.RFC3339)
		}

		if err := srv.Templates.Salvar(r.Context(), t); err != nil {
			srv.Logger.Error("salvar template falhou", slog.String("erro", err.Error()))
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}
		srv.Logger.Info("template salvo", slog.String("id", t.ID))
		writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "template": t})

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeErr(w, http.StatusBadRequest, "informe ?id=")
			return
		}
		if err := srv.Templates.Deletar(r.Context(), id); err != nil {
			srv.Logger.Error("deletar template falhou", slog.String("erro", err.Error()))
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}
		srv.Logger.Info("template removido", slog.String("id", id))
		writeJSON(w, http.StatusOK, map[string]any{"status": "removido", "id": id})

	default:
		writeErr(w, http.StatusMethodNotAllowed, "use GET, POST ou DELETE")
	}
}

// templatePreview renderiza um template com dados de exemplo.
func (srv *Server) templatePreview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TemplateID string `json:"template_id"`
		Corpo      string `json:"corpo"` // alternativa: corpo inline (sem salvar)
		ComFoto    bool   `json:"com_foto"`
		// Dados do produto para preview
		Nome       string  `json:"nome"`
		Preco      float64 `json:"preco"`
		Categoria  string  `json:"categoria"`
		Estrategia string  `json:"estrategia"`
		Link       string  `json:"link"`
		Imagem     string  `json:"imagem"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	var tmpl publish.Template
	if req.TemplateID != "" {
		var err error
		tmpl, err = srv.Templates.Buscar(r.Context(), req.TemplateID)
		if err != nil {
			writeErr(w, http.StatusNotFound, "template não encontrado")
			return
		}
	} else {
		tmpl = publish.Template{Corpo: req.Corpo, ComFoto: req.ComFoto}
	}

	oferta := publish.Oferta{
		Nome:       req.Nome,
		Preco:      req.Preco,
		Categoria:  req.Categoria,
		Estrategia: req.Estrategia,
		Link:       req.Link,
		Imagem:     req.Imagem,
	}

	// Dados de exemplo se não fornecidos
	if oferta.Nome == "" {
		oferta.Nome = "Sérum Vitamina C 30ml"
	}
	if oferta.Preco == 0 {
		oferta.Preco = 49.90
	}
	if oferta.Categoria == "" {
		oferta.Categoria = "Beleza"
	}
	if oferta.Estrategia == "" {
		oferta.Estrategia = "nicho"
	}

	rendered := tmpl.Renderizar(oferta)
	writeJSON(w, http.StatusOK, map[string]any{
		"preview":  rendered,
		"com_foto": tmpl.ComFoto,
		"imagem":   oferta.Imagem,
	})
}

func slugificarTemplate(s string) string {
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
		return "template"
	}
	return result
}
