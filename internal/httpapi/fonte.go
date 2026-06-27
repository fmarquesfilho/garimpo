package httpapi

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/strategy"
)

func (srv *Server) fonteAtiva(q url.Values) string {
	if f := q.Get("fonte"); f != "" {
		return f
	}
	if srv.Fonte != "" {
		return srv.Fonte
	}
	return "csv"
}

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

func (srv *Server) fonte(q url.Values) (source.ProductSource, string) {
	if srv.FonteFactory != nil {
		return srv.FonteFactory(q)
	}
	return srv.buildSource(q)
}

func (srv *Server) buildSource(q url.Values) (source.ProductSource, string) {
	switch srv.fonteAtiva(q) {
	case "shopee":
		cat := srv.CatID
		if c := q.Get("cat"); c != "" {
			if v, err := strconv.Atoi(c); err == nil {
				cat = v
			}
		}
		keyword := srv.Keyword
		if v := q.Get("keyword"); v != "" {
			keyword = v
		}
		sh := source.NewShopeeAPISource(os.Getenv("SHOPEE_APP_ID"), os.Getenv("SHOPEE_SECRET"))
		sh.ProductCatID = cat
		sh.Keyword = keyword
		sh.MaxPages = 2 // 100 produtos por busca (2 × 50)
		chave := "shopee|" + strconv.Itoa(cat) + "|" + keyword
		return sh, chave

	case "shopee-shop":
		var shopIDs []int64
		for _, s := range strings.Split(q.Get("shop_ids"), ",") {
			s = strings.TrimSpace(s)
			if v, err := strconv.ParseInt(s, 10, 64); err == nil {
				shopIDs = append(shopIDs, v)
			}
		}
		keyword := q.Get("keyword")
		sh := source.NewShopeeShopSource(os.Getenv("SHOPEE_APP_ID"), os.Getenv("SHOPEE_SECRET"), shopIDs)
		sh.Keyword = keyword
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
	return 20
}
