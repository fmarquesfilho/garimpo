package scoring

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

func TestMinMax(t *testing.T) {
	casos := []struct {
		v, mn, mx, quer float64
	}{
		{5, 0, 10, 0.5},
		{0, 0, 10, 0.0},
		{10, 0, 10, 1.0},
		{5, 5, 5, 0.5}, // sem variação no pool -> neutro
	}
	for _, c := range casos {
		if got := MinMax(c.v, c.mn, c.mx); got != c.quer {
			t.Errorf("MinMax(%v,%v,%v)=%v, quer %v", c.v, c.mn, c.mx, got, c.quer)
		}
	}
}

func TestEV(t *testing.T) {
	p := domain.Product{Commission: 0.10, Price: 100, Sales30d: 50}
	if got := EV(p); got != 500 {
		t.Errorf("EV=%v, quer 500", got)
	}
}

func TestComputeExtremos(t *testing.T) {
	produtos := []domain.Product{
		{Commission: 0.08, Price: 50, Sales30d: 10, Rating: 4.0},
		{Commission: 0.15, Price: 200, Sales30d: 100, Rating: 4.9},
	}
	s := Compute(produtos)
	if s.MinComm != 0.08 || s.MaxComm != 0.15 {
		t.Errorf("comissão min/max errados: %v / %v", s.MinComm, s.MaxComm)
	}
	if s.MinPrice != 50 || s.MaxPrice != 200 {
		t.Errorf("preço min/max errados: %v / %v", s.MinPrice, s.MaxPrice)
	}
}
