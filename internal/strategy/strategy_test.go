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
