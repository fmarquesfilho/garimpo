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
	"github.com/fmarquesfilho/garimpo/internal/source"
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
	srv.cache = map[string]*cacheEntry{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", srv.health)
	mux.HandleFunc("/api/candidatos", srv.candidatos)
	mux.HandleFunc("/api/comparar", srv.comparar)
	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
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
func rankear(produtos []domain.Product, st strategy.Strategy, elig strategy.Elegibilidade, n int) []candidatoDTO {
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

	src, chave := srv.buildSource(q)
	produtos, err := srv.fetchCacheado(src, chave)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	out := rankear(produtos, strategyDe(estrategia), srv.elegibilidade(q), topN(q))
	writeJSON(w, http.StatusOK, map[string]any{
		"fonte":      src.Name(),
		"estrategia": estrategia,
		"candidatos": out,
	})
}

func (srv *Server) comparar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	src, chave := srv.buildSource(q)
	produtos, err := srv.fetchCacheado(src, chave) // busca uma vez, ranqueia duas
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	n := topN(q)
	elig := srv.elegibilidade(q)
	writeJSON(w, http.StatusOK, map[string]any{
		"fonte":         src.Name(),
		"nicho":         rankear(produtos, strategy.NewNiche(), elig, n),
		"diversificada": rankear(produtos, strategy.Diversified{}, elig, n),
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
