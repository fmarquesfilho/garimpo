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

	muNov       sync.Mutex
	cacheNov    map[string]*cacheEntryNov

	// Override de alertas em runtime (evita os.Setenv race condition)
	alertasChatIDOverride       string
	alertasThresholdOverride    float64
	alertasApenasQuedasOverride *bool
}

type cacheEntry struct {
	produtos []domain.Product
	err      error
	em       time.Time
}

type cacheEntryNov struct {
	dados store.NovidadesLojas
	em    time.Time
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

	// ── Favoritos ─────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/favoritos", srv.listarFavoritos)
	mux.HandleFunc("POST /api/favoritos", srv.salvarFavorito)
	mux.HandleFunc("DELETE /api/favoritos", srv.removerFavorito)

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
	mux.HandleFunc("GET /api/conversoes/reais", srv.conversoesReais)
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

// spaHandler serve os arquivos estáticos do frontend (web/build) com fallback SPA.
func (srv *Server) spaHandler() http.Handler {
	dir := os.Getenv("WEB_DIR")
	if dir == "" {
		dir = "web/build"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			writeErr(w, http.StatusNotFound, "rota não encontrada")
			return
		}

		path := r.URL.Path
		if path == "/" {
			path = "/200.html"
		}

		fullPath := dir + path
		if _, err := os.Stat(fullPath); err == nil {
			if len(path) > 17 && path[:17] == "/_app/immutable/" {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if len(path) > 5 && path[:5] == "/_app" {
				w.Header().Set("Cache-Control", "no-cache")
			} else {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			}
			http.ServeFile(w, r, fullPath)
			return
		}

		fallback := dir + "/200.html"
		if _, err := os.Stat(fallback); err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, fallback)
	})
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
