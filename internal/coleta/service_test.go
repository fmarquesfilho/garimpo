package coleta

import (
	"context"
	"log/slog"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// ── Mocks ────────────────────────────────────────────────────────────────────

type mockSource struct {
	produtos []domain.Product
	fetched  int
}

func (m *mockSource) Name() string { return "mock" }
func (m *mockSource) Fetch() ([]domain.Product, error) {
	m.fetched++
	return m.produtos, nil
}

type mockStore struct {
	snapshots []store.Snapshot
	buscas    []store.Busca
}

func (m *mockStore) Registrar(context.Context, store.Evento) error { return nil }
func (m *mockStore) RegistrarSnapshot(_ context.Context, s store.Snapshot) error {
	m.snapshots = append(m.snapshots, s)
	return nil
}
func (m *mockStore) Estatisticas(context.Context, int) (store.Estatisticas, error) {
	return store.Estatisticas{}, nil
}
func (m *mockStore) SalvarBusca(_ context.Context, b store.Busca) error {
	m.buscas = append(m.buscas, b)
	return nil
}
func (m *mockStore) ListarBuscas(context.Context) ([]store.Busca, error) { return m.buscas, nil }
func (m *mockStore) HistoricoColetas(context.Context, int) ([]store.ColetaResumo, error) {
	return nil, nil
}
func (m *mockStore) Conversoes(context.Context, int) ([]store.ConversaoResumo, error) {
	return nil, nil
}
func (m *mockStore) SalvarPublicacao(context.Context, store.Publicacao) error { return nil }
func (m *mockStore) ListarPublicacoes(context.Context, string) ([]store.Publicacao, error) {
	return nil, nil
}
func (m *mockStore) AtualizarPublicacao(context.Context, string, string, string) error { return nil }
func (m *mockStore) Novidades(_ context.Context, id string, dias int) (store.NovidadesLojas, error) {
	return store.NovidadesLojas{BuscaID: id, DiasJanela: dias}, nil
}
func (m *mockStore) EvolucaoLojas(context.Context, int) (store.EvolucaoLojasResult, error) {
	return store.EvolucaoLojasResult{}, nil
}
func (m *mockStore) SalvarFavorito(context.Context, store.Favorito) error          { return nil }
func (m *mockStore) ListarFavoritos(context.Context, string) ([]store.Favorito, error) {
	return nil, nil
}
func (m *mockStore) RemoverFavorito(context.Context, string, string) error { return nil }
func (m *mockStore) EnsureSchema(context.Context) error                    { return nil }
func (m *mockStore) Nome() string                                          { return "mock" }

// ── Dados de teste ───────────────────────────────────────────────────────────

var produtosTeste = []domain.Product{
	{ID: "P1", Name: "Sérum", Category: "cosméticos", Price: 100, Commission: 0.15, Sales30d: 80, Rating: 4.8},
	{ID: "P2", Name: "Fone", Category: "eletrônicos", Price: 200, Commission: 0.10, Sales30d: 500, Rating: 4.3},
	{ID: "P3", Name: "Creme", Category: "cosméticos", Price: 50, Commission: 0.12, Sales30d: 300, Rating: 4.9},
}

// ── Testes ───────────────────────────────────────────────────────────────────

func TestExecutarColetaBasica(t *testing.T) {
	st := &mockStore{}
	src := &mockSource{produtos: produtosTeste}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, err := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Categoria:  "cosméticos",
		Keyword:    "sérum",
		Top:        3,
	})

	if err != nil {
		t.Fatal(err)
	}
	if resultado.Coletados != 3 {
		t.Errorf("esperava 3 coletados, veio %d", resultado.Coletados)
	}
	if resultado.Keyword != "sérum" {
		t.Errorf("keyword deveria ser 'sérum', veio %q", resultado.Keyword)
	}
	if len(st.snapshots) != 1 {
		t.Fatal("snapshot não gravado")
	}
	if len(st.snapshots[0].Itens) != 3 {
		t.Errorf("snapshot deveria ter 3 itens, tem %d", len(st.snapshots[0].Itens))
	}
	if src.fetched != 1 {
		t.Errorf("source deveria ter sido chamada 1 vez, foi %d", src.fetched)
	}
}

func TestExecutarColetaComBuscaIDUsaKeywordDaBusca(t *testing.T) {
	st := &mockStore{}
	src := &mockSource{produtos: produtosTeste}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, err := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		BuscaID:    "loja-123456",
		Top:        2,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Sem keyword explícito, deve usar o busca_id
	if resultado.Keyword != "loja-123456" {
		t.Errorf("keyword deveria ser 'loja-123456', veio %q", resultado.Keyword)
	}
	if st.snapshots[0].Keyword != "loja-123456" {
		t.Errorf("snapshot keyword deveria ser 'loja-123456', veio %q", st.snapshots[0].Keyword)
	}
}

func TestExecutarColetaTopLimitaResultados(t *testing.T) {
	st := &mockStore{}
	src := &mockSource{produtos: produtosTeste}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, _ := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "test",
		Top:        1,
	})

	if resultado.Coletados != 1 {
		t.Errorf("top=1 deveria limitar a 1, veio %d", resultado.Coletados)
	}
}

func TestExecutarColetaTopDefaultE20(t *testing.T) {
	// Gera 25 produtos
	muitos := make([]domain.Product, 25)
	for i := range muitos {
		muitos[i] = domain.Product{ID: "P", Name: "Prod", Price: 100, Commission: 0.1, Sales30d: 50, Rating: 4.5}
	}

	st := &mockStore{}
	src := &mockSource{produtos: muitos}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, _ := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "test",
		// Top não definido → default 20
	})

	if resultado.Coletados != 20 {
		t.Errorf("default top deveria ser 20, veio %d", resultado.Coletados)
	}
}

func TestExecutarColetaGravaTimestamp(t *testing.T) {
	st := &mockStore{}
	src := &mockSource{produtos: produtosTeste}
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, _ := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "test",
		Top:        2,
	})

	if resultado.Em.IsZero() {
		t.Error("timestamp não deveria ser zero")
	}
}

func TestExecutarColetaFalhaGraciosamente(t *testing.T) {
	st := &mockStore{}
	src := &mockSource{produtos: nil} // sem produtos
	svc := Novo(Deps{Store: st, Logger: slog.Default()})

	resultado, err := svc.Executar(context.Background(), src, Params{
		Estrategia: "nicho",
		Keyword:    "test",
		Top:        5,
	})

	if err != nil {
		t.Fatal("não deveria errar com lista vazia")
	}
	if resultado.Coletados != 0 {
		t.Errorf("sem produtos, coletados deveria ser 0, veio %d", resultado.Coletados)
	}
}
