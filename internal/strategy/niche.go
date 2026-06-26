package strategy

import (
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/scoring"
)

// Niche prioriza COMISSÃO + VALOR ESPERADO + AVALIAÇÃO.
// Ranking neutro em relação à categoria — todos os produtos são tratados
// igualmente. O multiplicador de nicho foi removido (simplificação).
type Niche struct{}

// NewNiche cria a estratégia padrão.
func NewNiche() Niche {
	return Niche{}
}

func (Niche) Name() string { return "nicho" }

func (n Niche) Score(p domain.Product, s scoring.Stats) domain.Scored {
	nComm := scoring.MinMax(p.Commission, s.MinComm, s.MaxComm)
	nEV := scoring.MinMax(scoring.EV(p), s.MinEV, s.MaxEV)
	nRating := scoring.MinMax(p.Rating, s.MinRating, s.MaxRating)

	reasons := map[string]float64{
		"comissao":       0.45 * nComm,
		"valor_esperado": 0.35 * nEV,
		"avaliacao":      0.20 * nRating,
	}
	base := reasons["comissao"] + reasons["valor_esperado"] + reasons["avaliacao"]

	return domain.Scored{Product: p, Score: base, Reasons: reasons}
}
