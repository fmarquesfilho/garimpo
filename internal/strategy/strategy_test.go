package strategy

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/scoring"
)

func TestEligible(t *testing.T) {
	if Eligible(domain.Product{Commission: 0.05}, 0.07) {
		t.Error("5% não deveria ser elegível com piso de 7%")
	}
	if !Eligible(domain.Product{Commission: 0.07}, 0.07) {
		t.Error("7% deveria ser elegível (piso inclusivo)")
	}
	if !Eligible(domain.Product{Commission: 0.12}, 0.07) {
		t.Error("12% deveria ser elegível")
	}
}

// O scoring é neutro em relação à categoria — dois produtos idênticos em
// categorias diferentes devem ter o mesmo score.
func TestNicheScoringNeutroCategoria(t *testing.T) {
	dentro := domain.Product{ID: "a", Category: "cosméticos", Price: 100, Commission: 0.10, Sales30d: 50, Rating: 4.5}
	fora := domain.Product{ID: "b", Category: "eletrônicos", Price: 100, Commission: 0.10, Sales30d: 50, Rating: 4.5}

	s := scoring.Compute([]domain.Product{dentro, fora})
	n := NewNiche()

	if n.Score(dentro, s).Score != n.Score(fora, s).Score {
		t.Errorf("scoring deveria ser neutro entre categorias: cosméticos=%.4f eletrônicos=%.4f",
			n.Score(dentro, s).Score, n.Score(fora, s).Score)
	}
}

// A diversificada ignora categoria: dois produtos idênticos em tudo menos
// categoria devem empatar.
func TestDiversificadaIgnoraCategoria(t *testing.T) {
	a := domain.Product{ID: "a", Category: "cosméticos", Price: 100, Commission: 0.10, Sales30d: 50, Rating: 4.5}
	b := domain.Product{ID: "b", Category: "eletrônicos", Price: 100, Commission: 0.10, Sales30d: 50, Rating: 4.5}

	s := scoring.Compute([]domain.Product{a, b})
	d := Diversified{}

	if d.Score(a, s).Score != d.Score(b, s).Score {
		t.Error("diversificada não deveria diferenciar por categoria")
	}
}

// ── Testes de mutation testing: verificam valores numéricos e pesos ──────────

func TestDiversificadaPesosCorretos(t *testing.T) {
	// Dois produtos extremos: P é máximo em tudo, P2 é mínimo
	p := domain.Product{ID: "max", Price: 200, Commission: 0.20, Sales30d: 100, Rating: 4.0}
	p2 := domain.Product{ID: "min", Price: 50, Commission: 0.05, Sales30d: 10, Rating: 3.0}
	s := scoring.Compute([]domain.Product{p, p2})

	d := Diversified{}
	scoredMax := d.Score(p, s)
	scoredMin := d.Score(p2, s)

	// Produto máximo deve ter score = 1.0 (pesos somam 0.50+0.30+0.20)
	if abs(scoredMax.Score-1.0) > 0.0001 {
		t.Errorf("produto máximo deveria ter score=1.0, veio %.4f", scoredMax.Score)
	}

	// Produto mínimo deve ter score = 0.0
	if abs(scoredMin.Score) > 0.0001 {
		t.Errorf("produto mínimo deveria ter score=0.0, veio %.4f", scoredMin.Score)
	}

	// Componentes devem existir e somar ao total
	ev := scoredMax.Reasons["valor_esperado"]
	comm := scoredMax.Reasons["comissao"]
	dem := scoredMax.Reasons["demanda"]
	soma := ev + comm + dem

	if abs(soma-scoredMax.Score) > 0.0001 {
		t.Errorf("componentes não somam ao score: %.4f+%.4f+%.4f=%.4f, score=%.4f",
			ev, comm, dem, soma, scoredMax.Score)
	}

	// Verifica que os pesos são 0.50, 0.30, 0.20 (com normalização=1 no max)
	if abs(ev-0.50) > 0.0001 {
		t.Errorf("peso valor_esperado deveria ser 0.50, veio %.4f", ev)
	}
	if abs(comm-0.30) > 0.0001 {
		t.Errorf("peso comissao deveria ser 0.30, veio %.4f", comm)
	}
	if abs(dem-0.20) > 0.0001 {
		t.Errorf("peso demanda deveria ser 0.20, veio %.4f", dem)
	}

	// Score é monotônico: mais comissão → score maior
	pMeio := domain.Product{ID: "mid", Price: 200, Commission: 0.10, Sales30d: 100, Rating: 4.0}
	scoredMeio := d.Score(pMeio, s)
	if scoredMeio.Score >= scoredMax.Score {
		t.Errorf("produto com menos comissão deveria ter score menor: mid=%.4f max=%.4f",
			scoredMeio.Score, scoredMax.Score)
	}
	if scoredMeio.Score <= scoredMin.Score {
		t.Errorf("produto intermediário deveria ter score maior que mínimo: mid=%.4f min=%.4f",
			scoredMeio.Score, scoredMin.Score)
	}
}

func TestNichePesosCorretos(t *testing.T) {
	p := domain.Product{ID: "max", Price: 100, Commission: 0.20, Sales30d: 200, Rating: 5.0}
	p2 := domain.Product{ID: "min", Price: 50, Commission: 0.05, Sales30d: 10, Rating: 2.0}
	s := scoring.Compute([]domain.Product{p, p2})

	n := NewNiche()
	scored := n.Score(p, s)
	scored2 := n.Score(p2, s)

	// P deve ter score máximo (1.0)
	if abs(scored.Score-1.0) > 0.0001 {
		t.Errorf("produto máximo nicho deveria ter score=1.0, veio %.4f", scored.Score)
	}

	// P2 deve ter score mínimo (0.0)
	if abs(scored2.Score) > 0.0001 {
		t.Errorf("produto mínimo nicho deveria ter score=0.0, veio %.4f", scored2.Score)
	}

	// Componentes devem existir e somar ao total
	soma := scored.Reasons["comissao"] + scored.Reasons["valor_esperado"] + scored.Reasons["avaliacao"]
	if abs(soma-scored.Score) > 0.0001 {
		t.Errorf("componentes nicho não somam ao score: %.4f vs %.4f", soma, scored.Score)
	}

	// Monotonia: mais comissão → score maior
	pMeio := domain.Product{ID: "mid", Price: 100, Commission: 0.10, Sales30d: 200, Rating: 5.0}
	scoredMeio := n.Score(pMeio, s)
	if scoredMeio.Score >= scored.Score {
		t.Error("produto com menos comissão deveria ter score menor")
	}
}

func TestElegibilidadeComPisos(t *testing.T) {
	eleg := Elegibilidade{ComissaoMin: 0.07, VendasMin: 10, NotaMin: 3.0}
	pipeline := PipelineCuradoria(eleg)

	cases := []struct {
		nome        string
		produto     domain.Product
		passaFiltro bool
	}{
		{"abaixo comissao", domain.Product{Commission: 0.05, Sales30d: 100, Rating: 4.0}, false},
		{"no piso comissao", domain.Product{Commission: 0.07, Sales30d: 100, Rating: 4.0}, true},
		{"abaixo vendas", domain.Product{Commission: 0.10, Sales30d: 5, Rating: 4.0}, false},
		{"no piso vendas", domain.Product{Commission: 0.10, Sales30d: 10, Rating: 4.0}, true},
		{"abaixo nota", domain.Product{Commission: 0.10, Sales30d: 100, Rating: 2.5}, false},
		{"no piso nota", domain.Product{Commission: 0.10, Sales30d: 100, Rating: 3.0}, true},
		{"tudo acima", domain.Product{Commission: 0.15, Sales30d: 200, Rating: 4.8}, true},
		{"tudo no limite", domain.Product{Commission: 0.07, Sales30d: 10, Rating: 3.0}, true},
	}

	for _, tc := range cases {
		t.Run(tc.nome, func(t *testing.T) {
			resultado := pipeline.Aplicar([]domain.Product{tc.produto})
			passou := len(resultado) > 0
			if passou != tc.passaFiltro {
				t.Errorf("produto %s: esperava passa=%v, veio passa=%v", tc.nome, tc.passaFiltro, passou)
			}
		})
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
