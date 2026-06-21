package scoring

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func TestSuspeitoFantasma(t *testing.T) {
	// Pool: dois normais e um "fantasma" (comissão altíssima, zero venda/nota).
	pool := []domain.Product{
		{ID: "ok1", Commission: 0.12, Sales30d: 200, Rating: 4.7},
		{ID: "ok2", Commission: 0.15, Sales30d: 150, Rating: 4.5},
		{ID: "fantasma", Commission: 0.83, Sales30d: 0, Rating: 0},
	}
	s := Compute(pool)

	fantasma := pool[2]
	if !Suspeito(fantasma, s) {
		t.Errorf("comissão 83%% com zero venda/nota deveria ser suspeito; P75=%.2f", s.CommissionP75)
	}
	for _, bom := range pool[:2] {
		if Suspeito(bom, s) {
			t.Errorf("produto com tração (%s) não deveria ser suspeito", bom.ID)
		}
	}
}

func TestSuspeitoExigeComissaoAlta(t *testing.T) {
	// Zero venda mas comissão baixa (abaixo do P75) -> NÃO é fantasma de comissão alta.
	pool := []domain.Product{
		{ID: "a", Commission: 0.50, Sales30d: 100},
		{ID: "b", Commission: 0.60, Sales30d: 100},
		{ID: "baixa_sem_venda", Commission: 0.08, Sales30d: 0},
	}
	s := Compute(pool)
	if Suspeito(pool[2], s) {
		t.Error("comissão baixa não deveria disparar a flag de produto-fantasma")
	}
}

func TestCommissionP75(t *testing.T) {
	pool := []domain.Product{
		{Commission: 0.10}, {Commission: 0.20}, {Commission: 0.30}, {Commission: 0.40},
	}
	s := Compute(pool)
	if s.CommissionP75 < 0.30 || s.CommissionP75 > 0.40 {
		t.Errorf("P75 fora do esperado: %.2f", s.CommissionP75)
	}
}
