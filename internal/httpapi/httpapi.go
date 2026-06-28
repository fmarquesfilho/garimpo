// Package httpapi expõe o motor de curadoria como uma API HTTP em JSON.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

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

	// Repo é o ponto de acesso unificado à persistência (novo padrão).
	// Quando presente, tem precedência sobre os campos legados abaixo.
	Repo store.Repository

	// ── Campos legados (mantidos para backward compat durante migração) ──
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

	muNov    sync.Mutex
	cacheNov map[string]*cacheEntryNov

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

// Handler monta o router Chi com todas as rotas, organizadas por domínio.
func (srv *Server) Handler() http.Handler {
	srv.inicializar()

	r := chi.NewRouter()

	// ── Middleware global ──────────────────────────────────────────────────
	r.Use(cors)
	r.Use(srv.logRequests)

	// ── Rotas públicas (sem autenticação) ─────────────────────────────────
	r.Get("/api/health", srv.health)
	r.Get("/api/candidatos", srv.candidatos)
	r.Get("/api/comparar", srv.comparar)
	r.Get("/api/produto/origem", srv.produtoOrigem)
	r.Post("/api/produto/origem/batch", srv.produtoOrigemBatch)
	r.Get("/api/docs", srv.apiDocs)
	r.Get("/api/openapi.yaml", srv.openapiSpec)

	// ── Rotas protegidas por token de coleta (Cloud Scheduler) ────────────
	r.Group(func(r chi.Router) {
		r.Use(srv.requireColetaToken)
		r.Post("/api/coletar", srv.coletar)
		r.Post("/api/conversoes/sync", srv.syncConversoes)
		r.Post("/api/publicar-pendentes", srv.publicarPendentes)
	})

	// ── Rotas autenticadas (usuário Firebase) ─────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(srv.requireAuth)

		// Curadoria
		r.Post("/api/eventos", srv.eventos)
		r.Get("/api/buscas", srv.listarBuscas)
		r.Post("/api/buscas", srv.salvarBusca)
		r.Post("/api/resolver-link", srv.resolverLink)

		// Lojas
		r.Get("/api/lojas", srv.listarLojas)
		r.Post("/api/lojas", srv.adicionarLoja)
		r.Delete("/api/lojas", srv.removerLoja)
		r.Get("/api/lojas/novidades", srv.novidades)
		r.Get("/api/lojas/evolucao", srv.evolucaoLojas)

		// Alertas
		r.Get("/api/alertas", srv.alertasConfig)
		r.Post("/api/alertas/testar", srv.alertasTestar)
		r.Post("/api/alertas/configurar", srv.alertasAtualizar)

		// Favoritos
		r.Get("/api/favoritos", srv.listarFavoritos)
		r.Post("/api/favoritos", srv.salvarFavorito)
		r.Delete("/api/favoritos", srv.removerFavorito)

		// Publicação
		r.Post("/api/publicar", srv.publicar)
		r.Get("/api/publicacoes", srv.listarPublicacoes)
		r.Post("/api/publicacoes", srv.agendarPublicacao)

		// Destinos e Templates
		r.Get("/api/destinos", srv.listarDestinos)
		r.Post("/api/destinos", srv.salvarDestino)
		r.Delete("/api/destinos", srv.deletarDestino)
		r.Get("/api/templates", srv.listarTemplates)
		r.Post("/api/templates", srv.salvarTemplate)
		r.Delete("/api/templates", srv.deletarTemplate)
		r.Post("/api/templates/preview", srv.templatePreview)
		r.Get("/api/whatsapp/grupos", srv.whatsappGrupos)

		// Análise
		r.Get("/api/estatisticas", srv.estatisticas)
		r.Get("/api/coletas", srv.coletas)
		r.Get("/api/conversoes", srv.conversoes)
		r.Get("/api/conversoes/reais", srv.conversoesReais)

		// Onboarding
		r.Get("/api/onboarding/status", srv.onboardingStatus)
		r.Post("/api/onboarding/termos", srv.onboardingTermos)
		r.Post("/api/onboarding/shopee", srv.onboardingShopee)
		r.Post("/api/onboarding/telegram", srv.onboardingTelegram)
		r.Post("/api/onboarding/validar", srv.onboardingValidar)
		r.Post("/api/onboarding/excluir-conta", srv.onboardingExcluirConta)

		// Admin (requer auth + admin)
		r.Group(func(r chi.Router) {
			r.Use(srv.requireAdmin)
			r.Get("/api/admin/logs", srv.adminLogs)
			r.Post("/api/admin/log-level", srv.adminLogLevel)
			r.Get("/api/admin/me", srv.adminMe)
			r.Get("/api/admin/shopee-introspect", srv.adminShopeeIntrospect)
		})
	})

	// ── Documentação (Starlight) ──────────────────────────────────────────
	r.Mount("/docs", srv.docsFileServer())

	// ── Frontend (SPA fallback) ───────────────────────────────────────────
	r.NotFound(srv.spaHandler().ServeHTTP)

	return r
}

// inicializar preenche campos com defaults quando não injetados (dev/local).
func (srv *Server) inicializar() {
	if srv.CacheTTL == 0 {
		srv.CacheTTL = 60 * time.Second
	}

	// Se Repo está presente, preenche campos legados a partir dele (bridge).
	// Isso permite migração gradual: handlers ainda leem os campos antigos,
	// mas a source of truth é o Repository.
	if srv.Repo != nil {
		if srv.Eventos == nil {
			srv.Eventos = &repoEventoStoreAdapter{repo: srv.Repo}
		}
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

// repoEventoStoreAdapter faz bridge entre Repository e o EventoStore legado.
// Delega para as sub-interfaces do Repo, mantendo compatibilidade com código
// que ainda usa srv.Eventos.
type repoEventoStoreAdapter struct {
	repo store.Repository
}

func (a *repoEventoStoreAdapter) Registrar(ctx context.Context, e store.Evento) error {
	return a.repo.Eventos().Registrar(ctx, e)
}
func (a *repoEventoStoreAdapter) RegistrarSnapshot(ctx context.Context, s store.Snapshot) error {
	return a.repo.Snapshots().RegistrarSnapshot(ctx, s)
}
func (a *repoEventoStoreAdapter) Estatisticas(ctx context.Context, dias int) (store.Estatisticas, error) {
	return a.repo.Snapshots().Estatisticas(ctx, dias)
}
func (a *repoEventoStoreAdapter) SalvarBusca(ctx context.Context, b store.Busca) error {
	return a.repo.Buscas().SalvarBusca(ctx, b)
}
func (a *repoEventoStoreAdapter) ListarBuscas(ctx context.Context) ([]store.Busca, error) {
	return a.repo.Buscas().ListarBuscas(ctx)
}
func (a *repoEventoStoreAdapter) HistoricoColetas(ctx context.Context, dias int) ([]store.ColetaResumo, error) {
	return a.repo.Snapshots().HistoricoColetas(ctx, dias)
}
func (a *repoEventoStoreAdapter) Conversoes(ctx context.Context, dias int) ([]store.ConversaoResumo, error) {
	return a.repo.Publicacoes().Conversoes(ctx, dias)
}
func (a *repoEventoStoreAdapter) SalvarPublicacao(ctx context.Context, p store.Publicacao) error {
	return a.repo.Publicacoes().SalvarPublicacao(ctx, p)
}
func (a *repoEventoStoreAdapter) ListarPublicacoes(ctx context.Context, status string) ([]store.Publicacao, error) {
	return a.repo.Publicacoes().ListarPublicacoes(ctx, status)
}
func (a *repoEventoStoreAdapter) AtualizarPublicacao(ctx context.Context, id, status, detalhe string) error {
	return a.repo.Publicacoes().AtualizarPublicacao(ctx, id, status, detalhe)
}
func (a *repoEventoStoreAdapter) Novidades(ctx context.Context, buscaID string, dias int) (store.NovidadesLojas, error) {
	return a.repo.Snapshots().Novidades(ctx, buscaID, dias)
}
func (a *repoEventoStoreAdapter) EvolucaoLojas(ctx context.Context, dias int) (store.EvolucaoLojasResult, error) {
	return a.repo.Snapshots().EvolucaoLojas(ctx, dias)
}
func (a *repoEventoStoreAdapter) SalvarFavorito(ctx context.Context, f store.Favorito) error {
	return a.repo.Favoritos().SalvarFavorito(ctx, f)
}
func (a *repoEventoStoreAdapter) ListarFavoritos(ctx context.Context, ownerUID string) ([]store.Favorito, error) {
	return a.repo.Favoritos().ListarFavoritos(ctx, ownerUID)
}
func (a *repoEventoStoreAdapter) RemoverFavorito(ctx context.Context, ownerUID, produtoID string) error {
	return a.repo.Favoritos().RemoverFavorito(ctx, ownerUID, produtoID)
}
func (a *repoEventoStoreAdapter) EnsureSchema(ctx context.Context) error {
	return a.repo.EnsureSchema(ctx)
}
func (a *repoEventoStoreAdapter) Nome() string {
	return a.repo.Nome()
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
