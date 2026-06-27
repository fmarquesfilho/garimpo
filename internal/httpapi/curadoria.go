package httpapi

import (
"math/rand"
"net/http"
"strings"
"time"

"github.com/fmarquesfilho/garimpo/internal/domain"
"github.com/fmarquesfilho/garimpo/internal/engine"
"github.com/fmarquesfilho/garimpo/internal/strategy"
)

type candidatoDTO struct {
	ID          string             `json:"id"`
	Nome        string             `json:"nome"`
	Categoria   string             `json:"categoria"`
	Loja        string             `json:"loja,omitempty"`
	LojaID      string             `json:"loja_id,omitempty"`
	Origem      string             `json:"origem,omitempty"`
	Preco       float64            `json:"preco"`
	Comissao    float64            `json:"comissao"`
	Vendas      int                `json:"vendas"`
	Avaliacao   float64            `json:"avaliacao"`
	Link        string             `json:"link"`
	LinkProduto string             `json:"link_produto,omitempty"`
	Imagem      string             `json:"imagem,omitempty"`
	Score       float64            `json:"score"`
	Componentes map[string]float64 `json:"componentes"`
	Suspeito    bool               `json:"suspeito"`
}

func toDTO(s domain.Scored) candidatoDTO {
	p := s.Product
	lojaID := p.ShopID
	if lojaID == "0" || lojaID == "" {
		// Tenta extrair do ProductLink (formato: -i.SHOPID.ITEMID)
		lojaID = extrairShopIDDoLink(p.Link)
	}
	return candidatoDTO{
		ID: p.ID, Nome: p.Name, Categoria: p.Category, Loja: p.ShopName,
		LojaID: lojaID, Origem: p.Origin,
		Preco: p.Price, Comissao: p.Commission, Vendas: p.Sales30d,
		Avaliacao: p.Rating, Link: p.Link, LinkProduto: p.ProductLink,
		Imagem: p.Image,
		Score: s.Score, Componentes: s.Reasons,
		Suspeito: s.Suspeito,
	}
}

// extrairShopIDDoLink extrai o shopId de uma URL no formato -i.SHOPID.ITEMID
func extrairShopIDDoLink(link string) string {
	// Formato: https://shopee.com.br/Nome-i.SHOPID.ITEMID ou https://shope.ee/...
	idx := strings.LastIndex(link, "-i.")
	if idx < 0 {
		return ""
	}
	rest := link[idx+3:]
	parts := strings.SplitN(rest, ".", 2)
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func rankearDTO(produtos []domain.Product, st strategy.Strategy, pipeline strategy.Pipeline, n int, fracaoExpl float64, r *rand.Rand) []candidatoDTO {
	scored := engine.RankearComPipeline(produtos, st, pipeline)
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

	var pipeline strategy.Pipeline
	if q.Get("sem_filtro") == "true" {
		pipeline = strategy.PipelineMonitoramento()
	} else {
		pipeline = strategy.PipelineCuradoria(srv.elegibilidade(q))
	}

	out := rankearDTO(produtos, strategyDe(estrategia), pipeline, topN(q),
srv.fracaoExploracao(q), rand.New(rand.NewSource(time.Now().UnixNano())))
	writeJSON(w, http.StatusOK, map[string]any{
"fonte":       src.Name(),
		"estrategia":  estrategia,
		"candidatos":  out,
		"total_bruto": len(produtos),
	})
}

func (srv *Server) comparar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	src, chave := srv.fonte(q)
	produtos, err := srv.fetchCacheado(src, chave)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	n := topN(q)
	pipeline := strategy.PipelineCuradoria(srv.elegibilidade(q))
	writeJSON(w, http.StatusOK, map[string]any{
"fonte":         src.Name(),
		"nicho":         rankearDTO(produtos, strategy.NewNiche(), pipeline, n, 0, nil),
		"diversificada": rankearDTO(produtos, strategy.Diversified{}, pipeline, n, 0, nil),
	})
}
