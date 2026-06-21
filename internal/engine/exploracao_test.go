package engine

import (
	"math/rand"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func poolScored(n int) []domain.Scored {
	s := make([]domain.Scored, n)
	for i := 0; i < n; i++ {
		s[i] = domain.Scored{
			Product: domain.Product{ID: string(rune('A' + i))},
			Score:   float64(n - i), // já ordenado desc
		}
	}
	return s
}

func TestExploracaoReservaVagas(t *testing.T) {
	scored := poolScored(20) // pool maior que n, para haver cauda
	r := rand.New(rand.NewSource(1))
	sel := SelecionarComExploracao(scored, 10, 0.2, r) // 10 vagas, 20% exploração

	if len(sel) != 10 {
		t.Fatalf("esperava 10 selecionados, veio %d", len(sel))
	}
	var explor int
	for _, s := range sel {
		if s.Exploracao {
			explor++
		}
	}
	if explor != 2 {
		t.Errorf("esperava 2 vagas de exploração (20%% de 10), veio %d", explor)
	}
}

func TestExploracaoPegaForaDoTopo(t *testing.T) {
	scored := poolScored(20)
	r := rand.New(rand.NewSource(42))
	sel := SelecionarComExploracao(scored, 10, 0.2, r)

	// os marcados como exploração devem vir de FORA do top-8 (posições >= 8 no ranking)
	top8 := map[string]bool{}
	for _, s := range scored[:8] {
		top8[s.Product.ID] = true
	}
	for _, s := range sel {
		if s.Exploracao && top8[s.Product.ID] {
			t.Errorf("produto de exploração %q veio do topo, deveria vir da cauda", s.Product.ID)
		}
	}
}

func TestExploracaoZeroDevolveTopo(t *testing.T) {
	scored := poolScored(10)
	sel := SelecionarComExploracao(scored, 5, 0, nil)
	if len(sel) != 5 {
		t.Fatalf("esperava 5, veio %d", len(sel))
	}
	for _, s := range sel {
		if s.Exploracao {
			t.Error("sem exploração, nenhum item deveria estar marcado")
		}
	}
}

func TestExploracaoPoolMenorQueN(t *testing.T) {
	scored := poolScored(3)
	sel := SelecionarComExploracao(scored, 10, 0.2, rand.New(rand.NewSource(1)))
	if len(sel) != 3 {
		t.Errorf("com pool de 3 e n=10, esperava 3, veio %d", len(sel))
	}
}
