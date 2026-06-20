// Package httpapi expõe o motor de curadoria como uma API HTTP em JSON.
// É só mais um mecanismo de entrega sobre o mesmo engine — não duplica regra.
package httpapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/engine"
	"github.com/fmarquesfilho/garimpo/internal/publish"
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
	Score       float64            `json:"score"`
	Componentes map[string]float64 `json:"componentes"`
}

func toDTO(s domain.Scored) candidatoDTO {
	p := s.Product
	return candidatoDTO{
		ID: p.ID, Nome: p.Name, Categoria: p.Category,
		Preco: p.Price, Comissao: p.Commission, Vendas: p.Sales30d,
		Avaliacao: p.Rating, Link: p.Link,
		Score: s.Score, Componentes: s.Reasons,
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

	// CacheTTL evita refazer o fetch (lento e sujeito a rate limit) a cada
	// ajuste de estratégia/piso no front e nas duas buscas do modo "comparar".
	CacheTTL time.Duration

	// Eventos registra decisões de curadoria (seleções) para análise no BigQuery.
	// Se nil, vira NopStore (não persiste).
	Eventos store.EventoStore

	// Publicador envia a oferta para um canal (Telegram). Se nil, vira o Mock.
	Publicador publish.Publicador

	// FonteFactory permite injetar a fonte (testes). Se nil, usa buildSource.
	FonteFactory func(q url.Values) (source.ProductSource, string)

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
	srv.cache = map[string]*cacheEntry{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", srv.health)
	mux.HandleFunc("/api/candidatos", srv.candidatos)
	mux.HandleFunc("/api/comparar", srv.comparar)
	mux.HandleFunc("/api/eventos", srv.eventos)
	mux.HandleFunc("/api/publicar", srv.publicar)
	mux.HandleFunc("/api/coletar", srv.coletar)
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// eventos registra uma decisão de curadoria (ex.: produto selecionado) no store.
func (srv *Server) eventos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
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
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var c struct {
		ID         string  `json:"id"`
		Nome       string  `json:"nome"`
		Categoria  string  `json:"categoria"`
		Preco      float64 `json:"preco"`
		Comissao   float64 `json:"comissao"`
		Link       string  `json:"link"`
		Estrategia string  `json:"estrategia"`
	}
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	res, err := srv.Publicador.Publicar(r.Context(), publish.Oferta{
		ProdutoID: c.ID, Nome: c.Nome, Categoria: c.Categoria,
		Preco: c.Preco, Comissao: c.Comissao, Link: c.Link, Estrategia: c.Estrategia,
	})
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	// Registra a publicação (best-effort) para análise por canal no BigQuery.
	_ = srv.Eventos.Registrar(r.Context(), store.Evento{
		Tipo: "publicacao", Canal: res.Canal, ProdutoID: c.ID, Nome: c.Nome,
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
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
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
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"categoria":  categoria,
		"estrategia": estrategia,
		"coletados":  len(snap.Itens),
		"em":         snap.Em,
	})
}

func (srv *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"fonte":     srv.fonteAtiva(url.Values{}),
		"categoria": srv.Categoria,
		"keyword":   srv.Keyword,
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

// rankear aplica elegibilidade + scoring + ordenação sobre um pool já buscado.
func rankearDTO(produtos []domain.Product, st strategy.Strategy, elig strategy.Elegibilidade, n int) []candidatoDTO {
	scored := engine.Rankear(produtos, st, elig)
	if n > len(scored) {
		n = len(scored)
	}
	out := make([]candidatoDTO, 0, n)
	for _, s := range scored[:n] {
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

	out := rankearDTO(produtos, strategyDe(estrategia), srv.elegibilidade(q), topN(q))
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
		"nicho":         rankearDTO(produtos, strategy.NewNiche(), elig, n),
		"diversificada": rankearDTO(produtos, strategy.Diversified{}, elig, n),
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
