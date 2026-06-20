// Package scoring concentra a matemática neutra de pontuação: cálculo do valor
// esperado e normalização min-max do pool de candidatos. É deliberadamente
// "burra" e sem opinião — a OPINIÃO (que peso dar a quê) mora em strategy.
package scoring

import "github.com/fmarquesfilho/garimpo/internal/domain"

// EV é um proxy do valor esperado bruto de comissão de um produto:
//
//	comissão (taxa) * preço * vendas no período.
//
// Aproxima "quanto de comissão esse produto tende a gerar" se anunciado.
func EV(p domain.Product) float64 {
	return p.Commission * p.Price * float64(p.Sales30d)
}

// MinMax normaliza v para [0,1] dado o intervalo [min,max] do pool.
// Sem variação no pool (min==max), devolve 0.5 (neutro) para não dividir por zero.
func MinMax(v, mn, mx float64) float64 {
	if mx == mn {
		return 0.5
	}
	return (v - mn) / (mx - mn)
}

// Stats guarda os extremos do pool, usados para normalizar cada componente.
type Stats struct {
	MinPrice, MaxPrice   float64
	MinComm, MaxComm     float64
	MinSales, MaxSales   float64
	MinRating, MaxRating float64
	MinEV, MaxEV         float64
}

// Compute extrai os extremos do conjunto de candidatos ELEGÍVEIS.
// A normalização é sempre relativa ao pool do dia — um produto é "bom" em
// comparação aos outros candidatos daquele momento, não em escala absoluta.
func Compute(products []domain.Product) Stats {
	if len(products) == 0 {
		return Stats{}
	}
	p0 := products[0]
	s := Stats{
		MinPrice: p0.Price, MaxPrice: p0.Price,
		MinComm: p0.Commission, MaxComm: p0.Commission,
		MinSales: float64(p0.Sales30d), MaxSales: float64(p0.Sales30d),
		MinRating: p0.Rating, MaxRating: p0.Rating,
		MinEV: EV(p0), MaxEV: EV(p0),
	}
	for _, p := range products {
		s.MinPrice = min(s.MinPrice, p.Price)
		s.MaxPrice = max(s.MaxPrice, p.Price)
		s.MinComm = min(s.MinComm, p.Commission)
		s.MaxComm = max(s.MaxComm, p.Commission)
		sales := float64(p.Sales30d)
		s.MinSales = min(s.MinSales, sales)
		s.MaxSales = max(s.MaxSales, sales)
		s.MinRating = min(s.MinRating, p.Rating)
		s.MaxRating = max(s.MaxRating, p.Rating)
		ev := EV(p)
		s.MinEV = min(s.MinEV, ev)
		s.MaxEV = max(s.MaxEV, ev)
	}
	return s
}
