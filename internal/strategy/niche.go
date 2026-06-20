package strategy

import (
	"strings"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/scoring"
)

// CategoriasNicho é o foco editorial da operação: cosméticos, perfumaria e
// bem-estar da mulher. Chaves em minúsculo para casar sem depender de acento/caixa.
var CategoriasNicho = map[string]bool{
	"cosméticos": true,
	"cosmeticos": true,
	"perfumaria": true,
	"bem-estar":  true,
	"bem estar":  true,
}

// Niche prioriza o nicho e valoriza COMISSÃO + AVALIAÇÃO (curadoria e confiança
// da audiência) acima do volume bruto. A tese é de marca: construir uma audiência
// que confia nas recomendações compõe retorno ao longo do tempo (cauda longa).
type Niche struct {
	BoostNoNicho   float64 // multiplicador para produtos do nicho
	BoostForaNicho float64 // multiplicador (penalização) para fora do nicho
}

// NewNiche traz padrões sensatos: nicho vale 1.5x, fora do nicho 0.5x.
func NewNiche() Niche {
	return Niche{BoostNoNicho: 1.5, BoostForaNicho: 0.5}
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

	mult := n.BoostForaNicho
	if CategoriasNicho[strings.ToLower(strings.TrimSpace(p.Category))] {
		mult = n.BoostNoNicho
	}
	reasons["multiplicador_nicho"] = mult

	return domain.Scored{Product: p, Score: base * mult, Reasons: reasons}
}
