package strategy

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func TestElegibilidadeAtende(t *testing.T) {
	base := domain.Product{Commission: 0.10, Sales30d: 50, Rating: 4.5}

	casos := []struct {
		nome string
		e    Elegibilidade
		p    domain.Product
		quer bool
	}{
		{"comissao_ok", Elegibilidade{ComissaoMin: 0.07}, base, true},
		{"comissao_abaixo", Elegibilidade{ComissaoMin: 0.15}, base, false},
		{"vendas_off_ignora", Elegibilidade{ComissaoMin: 0.07, VendasMin: 0}, base, true},
		{"vendas_abaixo", Elegibilidade{ComissaoMin: 0.07, VendasMin: 100}, base, false},
		{"vendas_ok", Elegibilidade{ComissaoMin: 0.07, VendasMin: 50}, base, true},
		{"nota_abaixo", Elegibilidade{ComissaoMin: 0.07, NotaMin: 4.8}, base, false},
		{"nota_ok", Elegibilidade{ComissaoMin: 0.07, NotaMin: 4.5}, base, true},
		{"fantasma_comissao_alta_zero_venda", Elegibilidade{ComissaoMin: 0.07, VendasMin: 1},
			domain.Product{Commission: 0.83, Sales30d: 0, Rating: 0}, false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := c.e.Atende(c.p); got != c.quer {
				t.Errorf("Atende()=%v, quer %v", got, c.quer)
			}
		})
	}
}

func TestElegibilidadePadrao(t *testing.T) {
	e := ElegibilidadePadrao()
	if e.ComissaoMin != MinCommission || e.VendasMin != 0 || e.NotaMin != 0 {
		t.Errorf("padrão inesperado: %+v", e)
	}
}
