// Package httpapi expõe o motor de curadoria como uma API HTTP em JSON.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/auth"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/logs"
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/scheduler"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
	"github.com/fmarquesfilho/garimpo/internal/tenant"
)

// Server guarda a configuração e dependências do servidor HTTP.
type Server struct {
	DefaultCSV string
	Fonte      string
	CatID      int
	Categoria  string
	Keyword    string
	VendasMin  int
	NotaMin    float64
	Exploracao float64
	CacheTTL   time.Duration

	Eventos    store.EventoStore
	Publicador publish.Publicador
	Scheduler  scheduler.Scheduler
	Auth       auth.Verifier
	Destinos   publish.DestinoStore
	Templates  publish.TemplateStore
	Tenants    tenant.Store

	FonteFactory func(q url.Values) (source.ProductSource, string)
	Logger       *slog.Logger
	LogBuffer    *logs.Buffer

	mu    sync.Mutex
	cache map[string]*cacheEntry
}

type cacheEntry struct {
	produtos []domain.Product
	err      error
	em       time.Time
}

// Handler monta o mux com todas as rotas, organizadas por domínio.
func (srv *Server) Handler() http.Handler {
	srv.inicializar()

	mux := http.NewServeMux()

	// ── Curadoria ─────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/candidatos", srv.candidatos)
	mux.HandleFunc("GET /api/comparar", srv.comparar)
	mux.HandleFunc("POST /api/eventos", srv.eventos)
	mux.HandleFunc("GET /api/buscas", srv.listarBuscas)
	mux.HandleFunc("POST /api/buscas", srv.salvarBusca)
	mux.HandleFunc("POST /api/resolver-link", srv.resolverLink)
	mux.HandleFunc("GET /api/produto/origem", srv.produtoOrigem)
	mux.HandleFunc("POST /api/produto/origem/batch", srv.produtoOrigemBatch)

	// ── Lojas (monitoramento) ─────────────────────────────────────────────
	mux.HandleFunc("GET /api/lojas", srv.listarLojas)
	mux.HandleFunc("POST /api/lojas", srv.adicionarLoja)
	mux.HandleFunc("DELETE /api/lojas", srv.removerLoja)
	mux.HandleFunc("GET /api/lojas/novidades", srv.novidades)
	mux.HandleFunc("GET /api/lojas/evolucao", srv.evolucaoLojas)

	// ── Alertas ───────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/alertas", srv.alertasConfig)
	mux.HandleFunc("POST /api/alertas/testar", srv.alertasTestar)
	mux.HandleFunc("POST /api/alertas/configurar", srv.alertasAtualizar)

	// ── Publicação ────────────────────────────────────────────────────────
	mux.HandleFunc("POST /api/publicar", srv.publicar)
	mux.HandleFunc("GET /api/publicacoes", srv.listarPublicacoes)
	mux.HandleFunc("POST /api/publicacoes", srv.agendarPublicacao)
	mux.HandleFunc("POST /api/publicar-pendentes", srv.publicarPendentes)

	// ── Destinos e Templates ──────────────────────────────────────────────
	mux.HandleFunc("GET /api/destinos", srv.listarDestinos)
	mux.HandleFunc("POST /api/destinos", srv.salvarDestino)
	mux.HandleFunc("DELETE /api/destinos", srv.deletarDestino)
	mux.HandleFunc("GET /api/templates", srv.listarTemplates)
	mux.HandleFunc("POST /api/templates", srv.salvarTemplate)
	mux.HandleFunc("DELETE /api/templates", srv.deletarTemplate)
	mux.HandleFunc("POST /api/templates/preview", srv.templatePreview)
	mux.HandleFunc("GET /api/whatsapp/grupos", srv.whatsappGrupos)

	// ── Coleta e Análise ──────────────────────────────────────────────────
	mux.HandleFunc("POST /api/coletar", srv.coletar)
	mux.HandleFunc("GET /api/estatisticas", srv.estatisticas)
	mux.HandleFunc("GET /api/coletas", srv.coletas)
	mux.HandleFunc("GET /api/conversoes", srv.conversoes)
	mux.HandleFunc("POST /api/conversoes/sync", srv.syncConversoes)

	// ── Admin ─────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/health", srv.health)
	mux.HandleFunc("GET /api/admin/logs", srv.adminLogs)
	mux.HandleFunc("POST /api/admin/log-level", srv.adminLogLevel)
	mux.HandleFunc("GET /api/admin/me", srv.adminMe)
	mux.HandleFunc("GET /api/admin/shopee-introspect", srv.adminShopeeIntrospect)
	mux.HandleFunc("GET /api/docs", srv.apiDocs)
	mux.HandleFunc("GET /api/openapi.yaml", srv.openapiSpec)

	// ── Onboarding / Tenant ──────────────────────────────────────────────
	mux.HandleFunc("GET /api/onboarding/status", srv.onboardingStatus)
	mux.HandleFunc("POST /api/onboarding/termos", srv.onboardingTermos)
	mux.HandleFunc("POST /api/onboarding/shopee", srv.onboardingShopee)
	mux.HandleFunc("POST /api/onboarding/telegram", srv.onboardingTelegram)
	mux.HandleFunc("POST /api/onboarding/validar", srv.onboardingValidar)
	mux.HandleFunc("POST /api/onboarding/excluir-conta", srv.onboardingExcluirConta)

	// ── Frontend (SPA fallback) ───────────────────────────────────────────
	mux.Handle("/", srv.spaHandler())

	return cors(srv.logRequests(mux))
}

// inicializar preenche campos com defaults quando não injetados (dev/local).
func (srv *Server) inicializar() {
	if srv.CacheTTL == 0 {
		srv.CacheTTL = 60 * time.Second
	}
	if srv.Eventos == nil {
		srv.Eventos = store.NopStore{}
	}
	if srv.Publicador == nil {
		srv.Publicador = publish.NovoMock("telegram")
	}
	if srv.Scheduler == nil {
		srv.Scheduler = scheduler.NopScheduler{}
	}
	if srv.Auth == nil {
		srv.Auth = auth.NopVerifier{}
	}
	if srv.Templates == nil {
		srv.Templates = publish.NovoMemTemplateStore()
	}
	if srv.Logger == nil {
		srv.Logger = slog.Default()
	}
	srv.cache = map[string]*cacheEntry{}
}

// spaHandler serve os arquivos estáticos do frontend (web/build) com fallback
// para 200.html (SPA). Em produção, os arquivos são embeddados ou montados em
// /web. Em dev, usa o diretório local web/build.
func (srv *Server) spaHandler() http.Handler {
	// Determina o diretório do frontend
	dir := os.Getenv("WEB_DIR")
	if dir == "" {
		dir = "web/build"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rotas /api/* que chegaram aqui não correspondem a nenhum handler
		// registrado — devolve 404 JSON (não o SPA).
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			writeErr(w, http.StatusNotFound, "rota não encontrada")
			return
		}

		// Tenta servir o arquivo pedido
		path := r.URL.Path
		if path == "/" {
			path = "/200.html"
		}

		// Verifica se o arquivo existe
		fullPath := dir + path
		if _, err := os.Stat(fullPath); err == nil {
			// Cache headers por tipo de asset
			if len(path) > 17 && path[:17] == "/_app/immutable/" {
				// Assets com hash no nome — nunca mudam, cache forever
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if len(path) > 5 && path[:5] == "/_app" {
				// Assets do SvelteKit sem hash — revalidar sempre
				w.Header().Set("Cache-Control", "no-cache")
			} else {
				// HTML, favicon, etc — nunca cachear (Safari iPad cacheia agressivamente)
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			}
			http.ServeFile(w, r, fullPath)
			return
		}

		// Fallback: SPA (200.html)
		fallback := dir + "/200.html"
		if _, err := os.Stat(fallback); err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, fallback)
	})
}

// ── Middleware ─────────────────────────────────────────────────────────────

type respCapturado struct {
	http.ResponseWriter
	status int
}

func (r *respCapturado) WriteHeader(c int) {
	r.status = c
	r.ResponseWriter.WriteHeader(c)
}

func (srv *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rc := &respCapturado{ResponseWriter: w, status: 200}
		next.ServeHTTP(rc, r)

		dur := time.Since(inicio)
		attrs := []any{
			slog.String("metodo", r.Method),
			slog.String("rota", r.URL.Path),
			slog.Int("status", rc.status),
			slog.Duration("dur", dur),
		}

		nivel := "info"
		switch {
		case rc.status >= 500:
			srv.Logger.Error("requisição", attrs...)
			nivel = "error"
		case r.URL.Path == "/api/health":
			srv.Logger.Debug("requisição", attrs...)
			nivel = "debug"
		default:
			srv.Logger.Info("requisição", attrs...)
		}

		if srv.LogBuffer != nil {
			srv.LogBuffer.Push(logs.Entry{
				Nivel: nivel, Msg: "requisição", Metodo: r.Method,
				Rota: r.URL.Path, Status: rc.status,
				DurMs: float64(dur.Microseconds()) / 1000.0, Em: inicio,
			})
		}
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────

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

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeErr retorna um erro no formato RFC 9457 (Problem Details).
// Mantém compatibilidade com o campo "erro" que o frontend já consome.
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

// ── Handlers simples ──────────────────────────────────────────────────────

func (srv *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"fonte":     srv.fonteAtiva(url.Values{}),
		"categoria": srv.Categoria,
		"keyword":   srv.Keyword,
		"store":     srv.Eventos.Nome(),
		"logs":      "stdout → Cloud Logging (Cloud Run) / terminal (local)",
	})
}

func (srv *Server) eventos(w http.ResponseWriter, r *http.Request) {
	var e store.Evento
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if e.Tipo == "" {
		e.Tipo = "selecao"
	}
	if err := srv.Eventos.Registrar(r.Context(), e); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "registrado"})
}

func (srv *Server) publicar(w http.ResponseWriter, r *http.Request) {
	var c struct {
		ID            string  `json:"id"`
		Nome          string  `json:"nome"`
		Categoria     string  `json:"categoria"`
		Preco         float64 `json:"preco"`
		Comissao      float64 `json:"comissao"`
		Link          string  `json:"link"`
		Imagem        string  `json:"imagem"`
		Estrategia    string  `json:"estrategia"`
		DestinoID     string  `json:"destino_id"`
		TemplateID    string  `json:"template_id"`
		LegendaCustom string  `json:"legenda_custom"`
	}
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	oferta := publish.Oferta{
		ProdutoID: c.ID, Nome: c.Nome, Categoria: c.Categoria,
		Preco: c.Preco, Comissao: c.Comissao, Link: c.Link, Imagem: c.Imagem,
		Estrategia: c.Estrategia, DestinoID: c.DestinoID, TemplateID: c.TemplateID,
		LegendaHTML: c.LegendaCustom,
	}

	res, err := srv.Publicador.Publicar(r.Context(), oferta)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	res.SubID = publish.SubID(res.Canal, c.Estrategia, time.Now())

	srv.Logger.Info("publicacao",
		slog.String("canal", res.Canal),
		slog.String("sub_id", res.SubID),
		slog.String("produto", c.ID),
		slog.Bool("enviado", res.Enviado),
	)

	_ = srv.Eventos.Registrar(r.Context(), store.Evento{
		Tipo: "publicacao", Canal: res.Canal, SubID: res.SubID, ProdutoID: c.ID, Nome: c.Nome,
		Categoria: c.Categoria, Estrategia: c.Estrategia, Comissao: c.Comissao, Preco: c.Preco,
	})

	writeJSON(w, http.StatusOK, res)
}
