// Package strategy define a PORTA de decisão (a curadoria). Cada estratégia é
// uma forma diferente de responder "qual produto vale mais a pena anunciar
// hoje?". Trocar de estratégia troca a curadoria sem tocar no motor.
package strategy

import (
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/scoring"
)

// Strategy é a porta: recebe um produto + as estatísticas do pool e devolve
// o produto pontuado (com a decomposição do score em Reasons).
type Strategy interface {
	Name() string
	Score(p domain.Product, s scoring.Stats) domain.Scored
}

// MinCommission é a regra de negócio da sua esposa: abaixo de 7%, não anuncia.
// É um piso de ELEGIBILIDADE, comum às duas estratégias.
const MinCommission = 0.07

// Eligible aplica o piso de comissão. Produtos abaixo do piso são descartados
// antes do scoring (não competem nem entram na normalização do pool).
func Eligible(p domain.Product, minCommission float64) bool {
	return p.Commission >= minCommission
}

// Elegibilidade reúne os pisos que um candidato precisa cruzar para entrar no
// ranking. Além da comissão (regra dela), os dados reais da Shopee mostraram
// que comissão alta + zero venda costuma ser produto-fantasma; por isso há
// também pisos opcionais de vendas e nota — credibilidade, não só comissão.
type Elegibilidade struct {
	ComissaoMin float64 // fração (ex.: 0.07)
	VendasMin   int     // 0 = sem filtro
	NotaMin     float64 // 0 = sem filtro
}

// Atende diz se o produto cruza todos os pisos configurados.
func (e Elegibilidade) Atende(p domain.Product) bool {
	if p.Commission < e.ComissaoMin {
		return false
	}
	if e.VendasMin > 0 && p.Sales30d < e.VendasMin {
		return false
	}
	if e.NotaMin > 0 && p.Rating < e.NotaMin {
		return false
	}
	return true
}

// ElegibilidadePadrao usa só o piso de comissão (compatível com o início do projeto).
func ElegibilidadePadrao() Elegibilidade {
	return Elegibilidade{ComissaoMin: MinCommission}
}
