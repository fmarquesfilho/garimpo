package strategy

import (
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/scoring"
)

// Diversified é AGNÓSTICA a categoria. Persegue o maior valor esperado imediato:
// comissão alta sobre produtos que já vendem em volume. É a tese de "pegar a
// onda" — boa para picos de receita, fraca para construir audiência fiel.
type Diversified struct{}

func (Diversified) Name() string { return "diversificada" }

func (Diversified) Score(p domain.Product, s scoring.Stats) domain.Scored {
	nEV := scoring.MinMax(scoring.EV(p), s.MinEV, s.MaxEV)
	nComm := scoring.MinMax(p.Commission, s.MinComm, s.MaxComm)
	nSales := scoring.MinMax(float64(p.Sales30d), s.MinSales, s.MaxSales)

	reasons := map[string]float64{
		"valor_esperado": 0.50 * nEV,
		"comissao":       0.30 * nComm,
		"demanda":        0.20 * nSales,
	}
	total := reasons["valor_esperado"] + reasons["comissao"] + reasons["demanda"]
	return domain.Scored{Product: p, Score: total, Reasons: reasons}
}
