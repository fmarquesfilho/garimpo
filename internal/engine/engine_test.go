package engine

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/strategy"
)

// fakeSource é um adaptador de teste — prova que o motor não depende de CSV
// nem de API, só da porta source.ProductSource.
type fakeSource struct{ produtos []domain.Product }

func (f fakeSource) Name() string                     { return "fake" }
func (f fakeSource) Fetch() ([]domain.Product, error) { return f.produtos, nil }

func TestRankFiltraInelegiveisEOrdena(t *testing.T) {
	produtos := []domain.Product{
		{ID: "barato", Category: "cosméticos", Price: 50, Commission: 0.05, Sales30d: 100, Rating: 4.0}, // 5% -> fora
		{ID: "bom", Category: "cosméticos", Price: 120, Commission: 0.15, Sales30d: 80, Rating: 4.8},
		{ID: "medio", Category: "perfumaria", Price: 80, Commission: 0.08, Sales30d: 40, Rating: 4.2},
	}
	eng := New(fakeSource{produtos}, strategy.NewNiche(), strategy.Elegibilidade{ComissaoMin: 0.07})

	ranked, err := eng.Rank()
	if err != nil {
		t.Fatal(err)
	}
	if len(ranked) != 2 {
		t.Fatalf("esperava 2 elegíveis (o de 5%% sai), veio %d", len(ranked))
	}
	if ranked[0].Product.ID != "bom" {
		t.Errorf("esperava 'bom' em primeiro, veio %q", ranked[0].Product.ID)
	}
}

func TestRankearFiltraEOrdenaSemFonte(t *testing.T) {
	produtos := []domain.Product{
		{ID: "barato", Category: "cosméticos", Price: 50, Commission: 0.05, Sales30d: 100, Rating: 4.0}, // 5% -> fora
		{ID: "bom", Category: "cosméticos", Price: 120, Commission: 0.15, Sales30d: 80, Rating: 4.8},
		{ID: "medio", Category: "perfumaria", Price: 80, Commission: 0.08, Sales30d: 40, Rating: 4.2},
	}
	ranked := Rankear(produtos, strategy.NewNiche(), strategy.Elegibilidade{ComissaoMin: 0.07})
	if len(ranked) != 2 {
		t.Fatalf("esperava 2 elegíveis, veio %d", len(ranked))
	}
	if ranked[0].Product.ID != "bom" {
		t.Errorf("esperava 'bom' em primeiro, veio %q", ranked[0].Product.ID)
	}
}

func TestTopLimita(t *testing.T) {
	produtos := []domain.Product{
		{ID: "a", Category: "cosméticos", Price: 100, Commission: 0.10, Sales30d: 50, Rating: 4.5},
		{ID: "b", Category: "perfumaria", Price: 100, Commission: 0.12, Sales30d: 60, Rating: 4.6},
		{ID: "c", Category: "bem-estar", Price: 100, Commission: 0.09, Sales30d: 70, Rating: 4.4},
	}
	eng := New(fakeSource{produtos}, strategy.Diversified{}, strategy.Elegibilidade{ComissaoMin: 0.07})

	top, err := eng.Top(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 2 {
		t.Fatalf("Top(2) deveria devolver 2, veio %d", len(top))
	}
}
func TestRankearPoolVazio(t *testing.T) {
	// Todos abaixo do piso -> ranking vazio, sem panic na normalização.
	produtos := []domain.Product{
		{ID: "a", Commission: 0.03, Sales30d: 10, Rating: 4.0},
		{ID: "b", Commission: 0.05, Sales30d: 20, Rating: 4.5},
	}
	got := Rankear(produtos, strategy.NewNiche(), strategy.Elegibilidade{ComissaoMin: 0.07})
	if len(got) != 0 {
		t.Fatalf("esperava 0 elegíveis, veio %d", len(got))
	}
}

func TestRankearUmProdutoNaoQuebra(t *testing.T) {
	// Pool de 1 -> min==max em todas as métricas (MinMax devolve 0.5). Não pode panic.
	produtos := []domain.Product{
		{ID: "unico", Category: "cosméticos", Price: 80, Commission: 0.12, Sales30d: 40, Rating: 4.6},
	}
	got := Rankear(produtos, strategy.NewNiche(), strategy.Elegibilidade{ComissaoMin: 0.07})
	if len(got) != 1 || got[0].Product.ID != "unico" {
		t.Fatalf("esperava o único produto, veio %+v", got)
	}
}
