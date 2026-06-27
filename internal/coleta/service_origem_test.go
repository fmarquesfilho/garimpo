package coleta

import (
	"context"
	"log/slog"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

func TestExecutarColetaAplicaOrigemPadrao(t *testing.T) {
	// Loja monitorada com origem_padrao — produtos devem herdar
	st := &mockStore{
		buscas: []store.Busca{
			{ID: "loja-999", ShopIDs: []int64{999}, OrigemPadrao: "Coreia", Ativo: true},
		},
	}
	produtos := []domain.Product{
		{ID: "P1", Name: "Sérum", Price: 80, Commission: 0.12, Sales30d: 100, Rating: 4.5, Origin: ""},
		{ID: "P2", Name: "Tônico", Price: 50, Commission: 0.10, Sales30d: 80, Rating: 4.3, Origin: ""},
	}
	src := &mockSource{produtos: produtos}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	_, err := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "skincare",
		BuscaID:    "loja-999",
		Top:        2,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verifica que o snapshot gravou com origem preenchida
	if len(st.snapshots) == 0 {
		t.Fatal("snapshot não gravado")
	}
	for _, item := range st.snapshots[0].Itens {
		if item.Origin != "Coreia" {
			t.Errorf("item %s deveria ter origin='Coreia', veio %q", item.ProdutoID, item.Origin)
		}
	}
}

func TestExecutarColetaNaoSobrescreveOrigemDaAPI(t *testing.T) {
	// Se a API já trouxe origem, não sobrescrever com origem_padrao
	st := &mockStore{
		buscas: []store.Busca{
			{ID: "loja-888", ShopIDs: []int64{888}, OrigemPadrao: "China", Ativo: true},
		},
	}
	produtos := []domain.Product{
		{ID: "P1", Name: "Produto", Price: 100, Commission: 0.15, Sales30d: 50, Rating: 4.0, Origin: "Japão"},
	}
	src := &mockSource{produtos: produtos}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	_, err := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "test",
		BuscaID:    "loja-888",
		Top:        1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Origin "Japão" (da API) deve ser preservada, não sobrescrita por "China"
	if st.snapshots[0].Itens[0].Origin != "Japão" {
		t.Errorf("origin deveria ser 'Japão' (preservada), veio %q", st.snapshots[0].Itens[0].Origin)
	}
}

func TestExecutarColetaSemOrigemPadraoNaoAlteraProdutos(t *testing.T) {
	st := &mockStore{
		buscas: []store.Busca{
			{ID: "loja-777", ShopIDs: []int64{777}, OrigemPadrao: "", Ativo: true},
		},
	}
	produtos := []domain.Product{
		{ID: "P1", Name: "Produto", Price: 100, Commission: 0.15, Sales30d: 50, Rating: 4.0, Origin: ""},
	}
	src := &mockSource{produtos: produtos}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	_, err := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "test",
		BuscaID:    "loja-777",
		Top:        1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if st.snapshots[0].Itens[0].Origin != "" {
		t.Errorf("origin deveria estar vazia, veio %q", st.snapshots[0].Itens[0].Origin)
	}
}

func TestExecutarColetaComElegibilidade(t *testing.T) {
	st := &mockStore{}
	produtos := []domain.Product{
		{ID: "P1", Name: "Bom", Price: 100, Commission: 0.15, Sales30d: 100, Rating: 4.8},
		{ID: "P2", Name: "Fraco", Price: 50, Commission: 0.03, Sales30d: 5, Rating: 2.0}, // abaixo dos pisos
	}
	src := &mockSource{produtos: produtos}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, _ := svc.Executar(context.Background(), src, Params{
		Estrategia:  "nicho",
		Keyword:     "test",
		Top:         10,
		ComissaoMin: 0.07,
		VendasMin:   10,
		NotaMin:     3.0,
	})

	// P2 deve ser filtrado por comissão (3% < 7%)
	if resultado.Coletados != 1 {
		t.Errorf("com filtros, deveria coletar 1 (P1), veio %d", resultado.Coletados)
	}
}
