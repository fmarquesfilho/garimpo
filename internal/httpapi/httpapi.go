// Package httpapi expõe o motor de curadoria como uma API HTTP em JSON.
// É só mais um mecanismo de entrega sobre o mesmo engine — não duplica regra.
package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/auth"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/engine"
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/scheduler"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
	"github.com/fmarquesfilho/garimpo/internal/strategy"
)

// candidatoDTO é o formato JSON consumido pelo frontend (campos em português).
type candidatoDTO struct {
	ID          string             `json:"id"`
	Nome        string             `json:"nome"`
	Categoria   string             `json:"categoria"`
	Preco       float64            `json:"preco"`
	Comissao    float64            `json:"comissao"`
	Vendas      int                `json:"vendas"`
	Avaliacao   float64            `json:"avaliacao"`
	Link        string             `json:"link"`
	Imagem      string             `json:"imagem,omitempty"`
	Score       float64            `json:"score"`
	Componentes map[string]float64 `json:"componentes"`
	Suspeito    bool               `json:"suspeito"`
	Exploracao  bool               `json:"exploracao"`
}

func toDTO(s domain.Scored) candidatoDTO {
	p := s.Product
	return candidatoDTO{
		ID: p.ID, Nome: p.Name, Categoria: p.Category,
		Preco: p.Price, Comissao: p.Commission, Vendas: p.Sales30d,
		Avaliacao: p.Rating, Link: p.Link, Imagem: p.Image,
		Score: s.Score, Componentes: s.Reasons,
		Suspeito: s.Suspeito, Exploracao: s.Exploracao,
	}
}

// Server guarda a configuração padrão da fonte. Quando a query não especifica
// nada, esses padrões valem — é assim que o front (que não manda 'fonte') passa
// a receber dados reais da Shopee só por subir o servidor com -fonte shopee.
type Server struct {
	DefaultCSV string

	Fonte     string // "csv" (padrão) | "shopee"
	CatID     int
	Categoria string
	Keyword   string

	// Pisos de elegibilidade padrão (a query pode sobrescrever).
	VendasMin int
	NotaMin   float64

	// Exploracao é a fração padrão de vagas reservadas para hold-out (0..1).
	// 0 = desligado. A query (?exploracao=) sobrescreve.
	Exploracao float64

	// CacheTTL evita refazer o fetch (lento e sujeito a rate limit) a cada
	// ajuste de estratégia/piso no front e nas duas buscas do modo "comparar".
	CacheTTL time.Duration

	// Eventos registra decisões de curadoria (seleções) para análise no BigQuery.
	// Se nil, vira NopStore (não persiste).
	Eventos store.EventoStore

	// Publicador envia a oferta para um canal (Telegram). Se nil, vira o Mock.
	Publicador publish.Publicador

	// Scheduler cria/atualiza jobs no Cloud Scheduler quando buscas são salvas.
	// Se nil, vira NopScheduler (não cria jobs).
	Scheduler scheduler.Scheduler

	// Auth valida tokens Firebase. Se nil, vira NopVerifier (aceita tudo).
	Auth auth.Verifier

	// Destinos gerencia os canais de publicação cadastrados pela usuária.
	// Se nil, o endpoint /api/destinos retorna erro (não configurado).
	Destinos publish.DestinoStore

	// Templates gerencia os modelos de mensagem cadastrados pela usuária.
	// Se nil, usa MemTemplateStore (com templates padrão embutidos).
	Templates publish.TemplateStore

	// FonteFactory permite injetar a fonte (testes). Se nil, usa buildSource.
	FonteFactory func(q url.Values) (source.ProductSource, string)

	// Logger estruturado por criticidade. Se nil, usa slog.Default().
	Logger *slog.Logger

	mu    sync.Mutex
	cache map[string]*cacheEntry
}

type cacheEntry struct {
	produtos []domain.Product
	err      error
	em       time.Time
}

func (srv *Server) Handler() http.Handler {
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

	mux := http.NewServeMux()

	// ── Rotas com método explícito (Go 1.22+ mux patterns) ────────────────
	mux.HandleFunc("GET /api/health", srv.health)
	mux.HandleFunc("GET /api/candidatos", srv.candidatos)
	mux.HandleFunc("GET /api/comparar", srv.comparar)
	mux.HandleFunc("POST /api/eventos", srv.eventos)
	mux.HandleFunc("POST /api/publicar", srv.publicar)
	mux.HandleFunc("POST /api/coletar", srv.coletar)
	mux.HandleFunc("GET /api/estatisticas", srv.estatisticas)
	mux.HandleFunc("GET /api/coletas", srv.coletas)
	mux.HandleFunc("GET /api/conversoes", srv.conversoes)

	// Buscas: GET lista, POST salva/remove
	mux.HandleFunc("GET /api/buscas", srv.listarBuscas)
	mux.HandleFunc("POST /api/buscas", srv.salvarBusca)

	// Destinos: GET lista, POST salva, DELETE remove
	mux.HandleFunc("GET /api/destinos", srv.listarDestinos)
	mux.HandleFunc("POST /api/destinos", srv.salvarDestino)
	mux.HandleFunc("DELETE /api/destinos", srv.deletarDestino)

	// Templates: GET lista, POST salva, DELETE remove, POST preview
	mux.HandleFunc("GET /api/templates", srv.listarTemplates)
	mux.HandleFunc("POST /api/templates", srv.salvarTemplate)
	mux.HandleFunc("DELETE /api/templates", srv.deletarTemplate)
	mux.HandleFunc("POST /api/templates/preview", srv.templatePreview)

	// Publicações: GET lista, POST agenda/envia, POST pendentes (scheduler)
	mux.HandleFunc("GET /api/publicacoes", srv.listarPublicacoes)
	mux.HandleFunc("POST /api/publicacoes", srv.agendarPublicacao)
	mux.HandleFunc("POST /api/publicar-pendentes", srv.publicarPendentes)

	return cors(srv.logRequests(mux))
}

// respCapturado embrulha o ResponseWriter para capturar o status no log.
type respCapturado struct {
	http.ResponseWriter
	status int
}

func (r *respCapturado) WriteHeader(c int) {
	r.status = c
	r.ResponseWriter.WriteHeader(c)
}

// logRequests registra cada requisição com método, rota, status e duração.
// /api/health cai em DEBUG (ruidoso, health checks frequentes); o resto em INFO,
// e respostas 5xx em ERROR.
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
		switch {
		case rc.status >= 500:
			srv.Logger.Error("requisição", attrs...)
		case r.URL.Path == "/api/health":
			srv.Logger.Debug("requisição", attrs...)
		default:
			srv.Logger.Info("requisição", attrs...)
		}
	})
}

// usuarioDoRequest extrai o usuário autenticado do header Authorization.
// Retorna nil se não autenticado (anônimo). Não bloqueia — os endpoints
// decidem individualmente se exigem auth.
func (srv *Server) usuarioDoRequest(r *http.Request) *auth.User {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil
	}
	return srv.Auth.Verify(r.Context(), token)
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

// eventos registra uma decisão de curadoria (ex.: produto selecionado) no store.
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

// publicar envia a oferta para o canal (Telegram/Mock) e registra a publicação.
func (srv *Server) publicar(w http.ResponseWriter, r *http.Request) {
	var c struct {
		ID         string  `json:"id"`
		Nome       string  `json:"nome"`
		Categoria  string  `json:"categoria"`
		Preco      float64 `json:"preco"`
		Comissao   float64 `json:"comissao"`
		Link       string  `json:"link"`
		Imagem     string  `json:"imagem"`
		Estrategia string  `json:"estrategia"`
		DestinoID  string  `json:"destino_id"`
		TemplateID string  `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	oferta := publish.Oferta{
		ProdutoID: c.ID, Nome: c.Nome, Categoria: c.Categoria,
		Preco: c.Preco, Comissao: c.Comissao, Link: c.Link, Imagem: c.Imagem,
		Estrategia: c.Estrategia, DestinoID: c.DestinoID, TemplateID: c.TemplateID,
	}

	// Se um template foi escolhido, aplica na oferta (renderiza corpo + decide se envia foto)
	if c.TemplateID != "" && srv.Templates != nil {
		tmpl, err := srv.Templates.Buscar(r.Context(), c.TemplateID)
		if err == nil {
			// Template com foto: mantém imagem; sem foto: remove imagem para forçar sendMessage
			if !tmpl.ComFoto {
				oferta.Imagem = ""
			}
		}
	}

	res, err := srv.Publicador.Publicar(r.Context(), oferta)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	// subId de atribuição (canal_estrategia_data) — pronto para o conversionReport.
	res.SubID = publish.SubID(res.Canal, c.Estrategia, time.Now())

	srv.Logger.Info("publicacao",
		slog.String("canal", res.Canal),
		slog.String("sub_id", res.SubID),
		slog.String("produto", c.ID),
		slog.Bool("enviado", res.Enviado),
	)

	// Registra a publicação (best-effort) para análise por canal no BigQuery.
	_ = srv.Eventos.Registrar(r.Context(), store.Evento{
		Tipo: "publicacao", Canal: res.Canal, SubID: res.SubID, ProdutoID: c.ID, Nome: c.Nome,
		Categoria: c.Categoria, Estrategia: c.Estrategia, Comissao: c.Comissao, Preco: c.Preco,
	})

	writeJSON(w, http.StatusOK, res)
}

// autorizadoColeta protege o endpoint de coleta: se COLETA_TOKEN estiver
// definido, exige o header X-Garimpo-Token igual. Sem token configurado (dev),
// fica liberado. Evita que terceiros disparem coletas e queimem o rate limit.
func (srv *Server) autorizadoColeta(r *http.Request) bool {
	tok := os.Getenv("COLETA_TOKEN")
	if tok == "" {
		return true
	}
	return r.Header.Get("X-Garimpo-Token") == tok
}

// coletar roda a busca de uma categoria e grava um snapshot (top N do momento)
// para análise posterior. Disparado pelo Cloud Scheduler em cron.
func (srv *Server) coletar(w http.ResponseWriter, r *http.Request) {
	if !srv.autorizadoColeta(r) {
		writeErr(w, http.StatusUnauthorized, "token de coleta inválido")
		return
	}

	q := r.URL.Query()
	estrategia := "nicho"
	if v := q.Get("estrategia"); v != "" {
		estrategia = v
	}

	src, chave := srv.fonte(q)
	produtos, err := srv.fetchCacheado(src, chave)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	scored := engine.Rankear(produtos, strategyDe(estrategia), srv.elegibilidade(q))
	n := topN(q)
	if n > len(scored) {
		n = len(scored)
	}

	categoria := q.Get("categoria")
	if categoria == "" {
		categoria = srv.Categoria
	}
	keyword := q.Get("keyword")
	if keyword == "" {
		keyword = srv.Keyword
	}

	snap := store.Snapshot{
		Categoria:  categoria,
		Keyword:    keyword,
		Estrategia: estrategia,
		Em:         time.Now().UTC(),
	}
	for i, s := range scored[:n] {
		p := s.Product
		snap.Itens = append(snap.Itens, store.ItemSnapshot{
			Posicao: i + 1, ProdutoID: p.ID, Nome: p.Name,
			Preco: p.Price, Comissao: p.Commission, Vendas: p.Sales30d,
			Nota: p.Rating, Score: s.Score,
		})
	}

	if err := srv.Eventos.RegistrarSnapshot(r.Context(), snap); err != nil {
		srv.Logger.Error("coleta falhou ao gravar snapshot",
			slog.String("categoria", categoria), slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	srv.Logger.Info("coleta",
		slog.String("categoria", categoria),
		slog.String("keyword", keyword),
		slog.String("estrategia", estrategia),
		slog.Int("coletados", len(snap.Itens)),
		slog.String("store", srv.Eventos.Nome()),
	)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"categoria":  categoria,
		"estrategia": estrategia,
		"coletados":  len(snap.Itens),
		"em":         snap.Em,
	})
}

// listarBuscas devolve os perfis de coleta ativos do usuário.
func (srv *Server) listarBuscas(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	// Sem auth → lista vazia (buscas são privadas)
	if user == nil {
		writeJSON(w, http.StatusOK, map[string]any{"buscas": []store.Busca{}, "store": srv.Eventos.Nome()})
		return
	}
	lista, err := srv.Eventos.ListarBuscas(r.Context())
	if err != nil {
		srv.Logger.Error("listar buscas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	// Filtra por owner
	var filtrada []store.Busca
	for _, b := range lista {
		if b.OwnerUID == "" || b.OwnerUID == user.UID {
			filtrada = append(filtrada, b)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"buscas": filtrada, "store": srv.Eventos.Nome()})
}

// salvarBusca salva/atualiza um perfil de coleta. Com ?remover, grava tombstone.
func (srv *Server) salvarBusca(w http.ResponseWriter, r *http.Request) {
	user := srv.usuarioDoRequest(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "faça login para salvar buscas")
		return
	}
	var b store.Busca
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	b = store.NormalizarBusca(b)
	if b.ID == "" {
		writeErr(w, http.StatusBadRequest, "busca precisa de ao menos uma keyword")
		return
	}
	b.Ativo = !r.URL.Query().Has("remover")
	b.OwnerUID = user.UID

	if err := srv.Eventos.SalvarBusca(r.Context(), b); err != nil {
		srv.Logger.Error("salvar busca falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	// Sincroniza com o Cloud Scheduler (best-effort, não bloqueia o usuário)
	go func() {
		params := scheduler.ColetaParams{
			Categoria:  b.Categoria,
			Estrategia: b.Estrategia,
			Top:        b.Top,
			VendasMin:  b.VendasMin,
			NotaMin:    b.NotaMin,
		}
		var err error
		if b.Ativo {
			err = srv.Scheduler.SyncBusca(context.Background(), b.ID, b.Keywords, b.Cron, params)
		} else {
			err = srv.Scheduler.DeletarBusca(context.Background(), b.ID, b.Keywords)
		}
		if err != nil {
			srv.Logger.Error("scheduler sync falhou", slog.String("busca", b.ID), slog.String("erro", err.Error()))
		} else {
			srv.Logger.Info("scheduler sync", slog.String("busca", b.ID), slog.Bool("ativo", b.Ativo), slog.String("cron", b.Cron))
		}
	}()

	srv.Logger.Info("busca salva", slog.String("id", b.ID), slog.Bool("ativo", b.Ativo))
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "ok", "id": b.ID, "ativo": b.Ativo})
}

// estatisticas devolve o resumo descritivo dos snapshots coletados (por
// categoria) numa janela de `dias` (padrão 30). É o primeiro passo do pipeline
// de análise e a base para um painel no frontend.
func (srv *Server) estatisticas(w http.ResponseWriter, r *http.Request) {
	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}
	est, err := srv.Eventos.Estatisticas(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("estatisticas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, est)
}

// coletas retorna o histórico de coletas executadas (snapshots agrupados por
// execução). Útil para validar que os jobs do scheduler estão funcionando.
func (srv *Server) coletas(w http.ResponseWriter, r *http.Request) {
	dias := 30
	if s := r.URL.Query().Get("dias"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			dias = v
		}
	}
	historico, err := srv.Eventos.HistoricoColetas(r.Context(), dias)
	if err != nil {
		srv.Logger.Error("historico coletas falhou", slog.String("erro", err.Error()))
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"coletas": historico, "dias": dias})
}

func (srv *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"fonte":     srv.fonteAtiva(url.Values{}),
		"categoria": srv.Categoria,
		"keyword":   srv.Keyword,
		"store":     srv.Eventos.Nome(),
		// Logs vão para stdout — no Cloud Run são capturados pelo Cloud Logging.
		// Filtre por severity, rota ou categoria em:
		// https://console.cloud.google.com/logs → recurso "Cloud Run Revision"
		"logs": "stdout → Cloud Logging (Cloud Run) / terminal (local)",
	})
}

func (srv *Server) fonteAtiva(q url.Values) string {
	if f := q.Get("fonte"); f != "" {
		return f
	}
	if srv.Fonte != "" {
		return srv.Fonte
	}
	return "csv"
}

// fracaoExploracao resolve a fração de exploração (query > padrão do servidor).
func (srv *Server) fracaoExploracao(q url.Values) float64 {
	f := srv.Exploracao
	if s := q.Get("exploracao"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			f = v
		}
	}
	if f < 0 {
		f = 0
	}
	if f > 0.9 {
		f = 0.9
	}
	return f
}

// fonte resolve a fonte: usa a injetada (testes) ou a padrão (buildSource).
func (srv *Server) fonte(q url.Values) (source.ProductSource, string) {
	if srv.FonteFactory != nil {
		return srv.FonteFactory(q)
	}
	return srv.buildSource(q)
}

// buildSource monta a fonte (csv ou shopee) usando a query OU os padrões do
// servidor, e devolve também uma chave estável para o cache.
func (srv *Server) buildSource(q url.Values) (source.ProductSource, string) {
	switch srv.fonteAtiva(q) {
	case "shopee":
		cat := srv.CatID
		if c := q.Get("cat"); c != "" {
			if v, err := strconv.Atoi(c); err == nil {
				cat = v
			}
		}
		categoria := srv.Categoria
		if v := q.Get("categoria"); v != "" {
			categoria = v
		}
		keyword := srv.Keyword
		if v := q.Get("keyword"); v != "" {
			keyword = v
		}
		sh := source.NewShopeeAPISource(os.Getenv("SHOPEE_APP_ID"), os.Getenv("SHOPEE_SECRET"))
		sh.ProductCatID = cat
		sh.CategoryLabel = categoria
		sh.Keyword = keyword
		chave := "shopee|" + strconv.Itoa(cat) + "|" + categoria + "|" + keyword
		return sh, chave

	case "shopee-shop":
		// Monitoramento de loja: busca por shopId(s)
		var shopIDs []int64
		for _, s := range strings.Split(q.Get("shop_ids"), ",") {
			s = strings.TrimSpace(s)
			if v, err := strconv.ParseInt(s, 10, 64); err == nil {
				shopIDs = append(shopIDs, v)
			}
		}
		keyword := q.Get("keyword")
		categoria := q.Get("categoria")
		if categoria == "" {
			categoria = srv.Categoria
		}
		sh := source.NewShopeeShopSource(os.Getenv("SHOPEE_APP_ID"), os.Getenv("SHOPEE_SECRET"), shopIDs)
		sh.Keyword = keyword
		sh.CategoryLabel = categoria
		chave := fmt.Sprintf("shop|%v|%s", shopIDs, keyword)
		return sh, chave

	default:
		csv := q.Get("csv")
		if csv == "" {
			csv = srv.DefaultCSV
		}
		return source.NewCSVSource(csv), "csv|" + csv
	}
}

// fetchCacheado busca os produtos da fonte com cache por TTL, de modo que
// trocar estratégia/piso (ou comparar as duas) reaproveite o mesmo fetch.
func (srv *Server) fetchCacheado(src source.ProductSource, chave string) ([]domain.Product, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if e, ok := srv.cache[chave]; ok && time.Since(e.em) < srv.CacheTTL {
		return e.produtos, e.err
	}
	produtos, err := src.Fetch()
	srv.cache[chave] = &cacheEntry{produtos: produtos, err: err, em: time.Now()}
	return produtos, err
}

func strategyDe(nome string) strategy.Strategy {
	if nome == "diversificada" {
		return strategy.Diversified{}
	}
	return strategy.NewNiche()
}

func (srv *Server) elegibilidade(q url.Values) strategy.Elegibilidade {
	e := strategy.Elegibilidade{
		ComissaoMin: strategy.MinCommission,
		VendasMin:   srv.VendasMin,
		NotaMin:     srv.NotaMin,
	}
	if s := q.Get("comissao_min"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			e.ComissaoMin = v
		}
	}
	if s := q.Get("vendas_min"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			e.VendasMin = v
		}
	}
	if s := q.Get("nota_min"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			e.NotaMin = v
		}
	}
	return e
}

func topN(q url.Values) int {
	if s := q.Get("top"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 10
}

// rankearDTO aplica elegibilidade + scoring + ordenação sobre um pool já buscado.
// Se fracaoExpl > 0 e houver rng, reserva parte das vagas para exploração.
func rankearDTO(produtos []domain.Product, st strategy.Strategy, elig strategy.Elegibilidade, n int, fracaoExpl float64, r *rand.Rand) []candidatoDTO {
	scored := engine.Rankear(produtos, st, elig)
	var escolhidos []domain.Scored
	if fracaoExpl > 0 && r != nil {
		escolhidos = engine.SelecionarComExploracao(scored, n, fracaoExpl, r)
	} else {
		if n > len(scored) {
			n = len(scored)
		}
		escolhidos = scored[:n]
	}
	out := make([]candidatoDTO, 0, len(escolhidos))
	for _, s := range escolhidos {
		out = append(out, toDTO(s))
	}
	return out
}

func (srv *Server) candidatos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	estrategia := "nicho"
	if v := q.Get("estrategia"); v != "" {
		estrategia = v
	}

	src, chave := srv.fonte(q)
	produtos, err := srv.fetchCacheado(src, chave)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	out := rankearDTO(produtos, strategyDe(estrategia), srv.elegibilidade(q), topN(q),
		srv.fracaoExploracao(q), rand.New(rand.NewSource(time.Now().UnixNano())))
	writeJSON(w, http.StatusOK, map[string]any{
		"fonte":      src.Name(),
		"estrategia": estrategia,
		"candidatos": out,
	})
}

func (srv *Server) comparar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	src, chave := srv.fonte(q)
	produtos, err := srv.fetchCacheado(src, chave) // busca uma vez, ranqueia duas
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	n := topN(q)
	elig := srv.elegibilidade(q)
	writeJSON(w, http.StatusOK, map[string]any{
		"fonte":         src.Name(),
		"nicho":         rankearDTO(produtos, strategy.NewNiche(), elig, n, 0, nil),
		"diversificada": rankearDTO(produtos, strategy.Diversified{}, elig, n, 0, nil),
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"erro": msg})
}
