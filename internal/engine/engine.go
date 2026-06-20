// Package engine orquestra o fluxo: fonte -> elegibilidade -> scoring -> ranking.
// Depende apenas das PORTAS (source.ProductSource e strategy.Strategy), nunca
// de implementações concretas. É o coração testável da prova de conceito.
package engine

import (
	"sort"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/scoring"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/strategy"
)

type Engine struct {
	Source        source.ProductSource
	Strategy      strategy.Strategy
	Elegibilidade strategy.Elegibilidade
}

func New(src source.ProductSource, st strategy.Strategy, elig strategy.Elegibilidade) *Engine {
	return &Engine{Source: src, Strategy: st, Elegibilidade: elig}
}

// Rankear aplica elegibilidade + scoring + ordenação sobre um pool de produtos
// JÁ buscado. Separar isto do fetch permite buscar a fonte uma vez (cache) e
// ranquear de várias formas (estratégias diferentes, comparação).
func Rankear(produtos []domain.Product, st strategy.Strategy, elig strategy.Elegibilidade) []domain.Scored {
	// 1. Elegibilidade: comissão (+ vendas/nota, se configurados).
	elegiveis := make([]domain.Product, 0, len(produtos))
	for _, p := range produtos {
		if elig.Atende(p) {
			elegiveis = append(elegiveis, p)
		}
	}

	// 2. Estatísticas do pool (normalização relativa aos candidatos do dia).
	stats := scoring.Compute(elegiveis)

	// 3. Scoring de cada elegível pela estratégia escolhida.
	scored := make([]domain.Scored, 0, len(elegiveis))
	for _, p := range elegiveis {
		scored = append(scored, st.Score(p, stats))
	}

	// 4. Ranking decrescente, estável (empates preservam ordem de entrada).
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	return scored
}

// Rank devolve os candidatos elegíveis ordenados do melhor para o pior.
func (e *Engine) Rank() ([]domain.Scored, error) {
	produtos, err := e.Source.Fetch()
	if err != nil {
		return nil, err
	}
	return Rankear(produtos, e.Strategy, e.Elegibilidade), nil
}

// Top devolve os n melhores candidatos.
func (e *Engine) Top(n int) ([]domain.Scored, error) {
	scored, err := e.Rank()
	if err != nil {
		return nil, err
	}
	if n > len(scored) {
		n = len(scored)
	}
	return scored[:n], nil
}
