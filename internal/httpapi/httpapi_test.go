package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/auth"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// --- dublês ---------------------------------------------------------------

type fonteFake struct {
	produtos []domain.Product
	chamadas int
}

func (f *fonteFake) Name() string { return "fake" }
func (f *fonteFake) Fetch() ([]domain.Product, error) {
	f.chamadas++
	return f.produtos, nil
}

type spyStore struct {
	eventos   []store.Evento
	snapshots []store.Snapshot
	buscas    []store.Busca
}

func (s *spyStore) Registrar(_ context.Context, e store.Evento) error {
	s.eventos = append(s.eventos, e)
	return nil
}
func (s *spyStore) RegistrarSnapshot(_ context.Context, snap store.Snapshot) error {
	s.snapshots = append(s.snapshots, snap)
	return nil
}
func (s *spyStore) Estatisticas(_ context.Context, dias int) (store.Estatisticas, error) {
	return store.Estatisticas{
		Fonte: "spy", DiasJanela: dias, TotalAmostras: 3,
		PorCategoria: []store.EstatCategoria{
			{Categoria: "cosméticos", Amostras: 3, ComissaoMedia: 0.12, TeorMedio: 0.5},
		},
	}, nil
}
func (s *spyStore) Nome() string { return "spy" }

func (s *spyStore) SalvarBusca(_ context.Context, b store.Busca) error {
	s.buscas = append(s.buscas, b)
	return nil
}
func (s *spyStore) ListarBuscas(_ context.Context) ([]store.Busca, error) {
	return s.buscas, nil
}
func (s *spyStore) HistoricoColetas(_ context.Context, _ int) ([]store.ColetaResumo, error) {
	return nil, nil
}
func (s *spyStore) EnsureSchema(_ context.Context) error { return nil }

type spyPub struct {
	chamadas int
	ultima   publish.Oferta
}

func (p *spyPub) Nome() string { return "spy" }
func (p *spyPub) Publicar(_ context.Context, o publish.Oferta) (publish.Resultado, error) {
	p.chamadas++
	p.ultima = o
	return publish.Resultado{Canal: "spy", Enviado: true, Mensagem: o.Mensagem(), Detalhe: "spy"}, nil
}

// fakeVerifier aceita qualquer token não-vazio e retorna um usuário fixo.
type fakeVerifier struct{}

func (fakeVerifier) Verify(_ context.Context, token string) *auth.User {
	if token == "" {
		return nil
	}
	return &auth.User{UID: "test-user", Email: "test@test.com"}
}

var amostra = []domain.Product{
	{ID: "P1", Name: "Sérum", Category: "cosméticos", Price: 100, Commission: 0.15, Sales30d: 80, Rating: 4.8},
	{ID: "P2", Name: "Fone", Category: "eletrônicos", Price: 100, Commission: 0.10, Sales30d: 900, Rating: 4.3},
	{ID: "P3", Name: "Creme", Category: "cosméticos", Price: 50, Commission: 0.05, Sales30d: 300, Rating: 4.9}, // 5% -> fora
}

func montar(fonte *fonteFake, ev store.EventoStore, pub publish.Publicador) http.Handler {
	srv := &Server{
		Eventos:    ev,
		Publicador: pub,
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return fonte, "fake|fixo"
		},
	}
	return srv.Handler()
}

func req(t *testing.T, h http.Handler, metodo, alvo string, corpo []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if corpo != nil {
		r = httptest.NewRequest(metodo, alvo, bytes.NewReader(corpo))
	} else {
		r = httptest.NewRequest(metodo, alvo, nil)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// --- testes ---------------------------------------------------------------

func TestCandidatosRankeiaEFiltra(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/candidatos?estrategia=nicho", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var resp struct {
		Estrategia string `json:"estrategia"`
		Candidatos []struct {
			ID string `json:"id"`
		} `json:"candidatos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Candidatos) != 2 { // P3 (5%) cai no piso
		t.Fatalf("esperava 2 candidatos, veio %d", len(resp.Candidatos))
	}
	if resp.Candidatos[0].ID != "P1" { // nicho prioriza cosmético de alta comissão
		t.Errorf("esperava P1 no topo, veio %q", resp.Candidatos[0].ID)
	}
}

func TestCandidatosFiltroVendasMin(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/candidatos?estrategia=diversificada&vendas_min=100", nil, nil)
	var resp struct {
		Candidatos []struct {
			ID string `json:"id"`
		} `json:"candidatos"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Candidatos) != 1 || resp.Candidatos[0].ID != "P2" {
		t.Errorf("com vendas_min=100 só P2 passa; veio %+v", resp.Candidatos)
	}
}

func TestCompararTrazAsDuas(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/comparar", nil, nil)
	var resp struct {
		Nicho         []map[string]any `json:"nicho"`
		Diversificada []map[string]any `json:"diversificada"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Nicho) == 0 || len(resp.Diversificada) == 0 {
		t.Errorf("comparar deveria trazer as duas listas: nicho=%d div=%d", len(resp.Nicho), len(resp.Diversificada))
	}
}

func TestEventosRegistra(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})
	corpo := []byte(`{"tipo":"selecao","id":"P1","nome":"Sérum","estrategia":"nicho"}`)
	rec := req(t, h, "POST", "/api/eventos", corpo, map[string]string{"Content-Type": "application/json"})
	if rec.Code != 202 {
		t.Fatalf("status %d", rec.Code)
	}
	if len(sp.eventos) != 1 || sp.eventos[0].ProdutoID != "P1" || sp.eventos[0].Tipo != "selecao" {
		t.Errorf("evento não registrado corretamente: %+v", sp.eventos)
	}
}

func TestEventosRecusaGET(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	if rec := req(t, h, "GET", "/api/eventos", nil, nil); rec.Code != 405 {
		t.Errorf("GET em /api/eventos deveria dar 405, veio %d", rec.Code)
	}
}

func TestPublicarChamaPublicadorERegistra(t *testing.T) {
	sp := &spyStore{}
	pub := &spyPub{}
	h := montar(&fonteFake{produtos: amostra}, sp, pub)
	corpo := []byte(`{"id":"P1","nome":"Sérum","preco":100,"comissao":0.15,"link":"http://l","estrategia":"nicho"}`)
	rec := req(t, h, "POST", "/api/publicar", corpo, map[string]string{"Content-Type": "application/json"})
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var res struct {
		Canal   string `json:"canal"`
		Enviado bool   `json:"enviado"`
	}
	json.Unmarshal(rec.Body.Bytes(), &res)
	if !res.Enviado || res.Canal != "spy" {
		t.Errorf("resultado inesperado: %+v", res)
	}
	if pub.chamadas != 1 || pub.ultima.Nome != "Sérum" {
		t.Errorf("publicador não chamado certo: chamadas=%d ultima=%+v", pub.chamadas, pub.ultima)
	}
	if len(sp.eventos) != 1 || sp.eventos[0].Tipo != "publicacao" || sp.eventos[0].Canal != "spy" {
		t.Errorf("publicação não registrada: %+v", sp.eventos)
	}
}

func TestColetarGravaSnapshotComToken(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})

	// sem token -> 401
	if rec := req(t, h, "POST", "/api/coletar", nil, nil); rec.Code != 401 {
		t.Fatalf("sem token deveria dar 401, veio %d", rec.Code)
	}
	// com token -> 202 e snapshot gravado
	rec := req(t, h, "POST", "/api/coletar?categoria=cosméticos", nil, map[string]string{"X-Garimpo-Token": "segredo"})
	if rec.Code != 202 {
		t.Fatalf("com token deveria dar 202, veio %d", rec.Code)
	}
	if len(sp.snapshots) != 1 || len(sp.snapshots[0].Itens) == 0 {
		t.Errorf("snapshot não gravado: %+v", sp.snapshots)
	}
}

func TestBuscasSalvaEListagem(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})

	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"}

	// envia no novo formato: keywords[] + id implícito via slug
	corpo := []byte(`{"keywords":["perfume"],"categoria":"perfumaria","estrategia":"nicho","cron":"0 8 * * *","top":20}`)
	rec := req(t, h, "POST", "/api/buscas", corpo, authH)
	if rec.Code != 202 {
		t.Fatalf("POST status %d — body: %s", rec.Code, rec.Body.String())
	}
	if len(sp.buscas) != 1 || sp.buscas[0].ID != "perfume" || !sp.buscas[0].Ativo {
		t.Fatalf("busca não salva certo: %+v", sp.buscas)
	}
	if sp.buscas[0].OwnerUID != "test-user" {
		t.Errorf("owner_uid deveria ser test-user, veio %q", sp.buscas[0].OwnerUID)
	}

	rec = req(t, h, "GET", "/api/buscas", nil, map[string]string{"Authorization": "Bearer fake-token"})
	var resp struct {
		Buscas []struct {
			ID   string `json:"id"`
			Cron string `json:"cron"`
		} `json:"buscas"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Buscas) != 1 || resp.Buscas[0].Cron != "0 8 * * *" {
		t.Errorf("listagem inesperada: %+v", resp.Buscas)
	}
}

func TestBuscasSalvaCompatibilidadeLegada(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})

	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"}

	// formato antigo: campo "nome" + "keyword" (string) — deve ser normalizado
	corpo := []byte(`{"nome":"perfumaria diária","keyword":"perfume","categoria":"perfumaria","cron":"0 8 * * *","top":20}`)
	rec := req(t, h, "POST", "/api/buscas", corpo, authH)
	if rec.Code != 202 {
		t.Fatalf("POST legado status %d — body: %s", rec.Code, rec.Body.String())
	}
	if len(sp.buscas) != 1 || sp.buscas[0].ID == "" {
		t.Fatalf("busca legada não normalizada: %+v", sp.buscas)
	}
	if len(sp.buscas[0].Keywords) == 0 || sp.buscas[0].Keywords[0] != "perfume" {
		t.Errorf("keyword legada não migrada para Keywords[]: %+v", sp.buscas[0])
	}
}

func TestBuscasExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	// POST sem auth → 401
	rec := req(t, h, "POST", "/api/buscas", []byte(`{"keywords":["x"]}`), map[string]string{"Content-Type": "application/json"})
	if rec.Code != 401 {
		t.Errorf("busca sem auth deveria dar 401, veio %d", rec.Code)
	}
	// GET sem auth → 200 mas lista vazia
	rec = req(t, h, "GET", "/api/buscas", nil, nil)
	if rec.Code != 200 {
		t.Errorf("GET sem auth deveria dar 200, veio %d", rec.Code)
	}
}

func TestBuscasExigeKeywords(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	// sem keywords e sem keyword legado → 400
	rec := req(t, h, "POST", "/api/buscas", []byte(`{}`), map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"})
	if rec.Code != 400 {
		t.Errorf("busca sem keywords deveria dar 400, veio %d", rec.Code)
	}
}

func TestBuscasRemoverMarcaInativo(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})
	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"}
	req(t, h, "POST", "/api/buscas?remover", []byte(`{"id":"minha-busca","keywords":["x"]}`), authH)
	if len(sp.buscas) != 1 || sp.buscas[0].Ativo {
		t.Errorf("remover deveria gravar tombstone (ativo=false): %+v", sp.buscas)
	}
}

func TestEstatisticasRetornaResumo(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/estatisticas?dias=15", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var resp struct {
		Fonte        string `json:"fonte"`
		DiasJanela   int    `json:"dias_janela"`
		PorCategoria []struct {
			Categoria string `json:"categoria"`
		} `json:"por_categoria"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.DiasJanela != 15 {
		t.Errorf("dias_janela esperado 15, veio %d", resp.DiasJanela)
	}
	if len(resp.PorCategoria) != 1 || resp.PorCategoria[0].Categoria != "cosméticos" {
		t.Errorf("por_categoria inesperado: %+v", resp.PorCategoria)
	}
}

func TestCacheBuscaUmaVez(t *testing.T) {
	f := &fonteFake{produtos: amostra}
	h := montar(f, &spyStore{}, &spyPub{})
	req(t, h, "GET", "/api/candidatos?estrategia=nicho", nil, nil)
	req(t, h, "GET", "/api/candidatos?estrategia=diversificada", nil, nil)
	req(t, h, "GET", "/api/comparar", nil, nil)
	if f.chamadas != 1 {
		t.Errorf("a fonte deveria ser buscada 1 vez (cache), veio %d", f.chamadas)
	}
}

func TestBuscasMultiKeyword(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})

	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"}
	corpo := []byte(`{"keywords":["kenzo","shiseido","issey miyake"],"categoria":"perfumaria","estrategia":"ambas","cron":"0 8,18 * * *","top":12}`)
	rec := req(t, h, "POST", "/api/buscas", corpo, authH)
	if rec.Code != 202 {
		t.Fatalf("POST status %d — body: %s", rec.Code, rec.Body.String())
	}
	if len(sp.buscas) != 1 {
		t.Fatalf("esperava 1 busca, veio %d", len(sp.buscas))
	}
	b := sp.buscas[0]
	if b.ID != "kenzo" {
		t.Errorf("ID deveria ser slug da primeira keyword (kenzo), veio %q", b.ID)
	}
	if len(b.Keywords) != 3 {
		t.Errorf("esperava 3 keywords, veio %d", len(b.Keywords))
	}
	if b.Estrategia != "ambas" {
		t.Errorf("estrategia deveria ser 'ambas', veio %q", b.Estrategia)
	}
	if b.Cron != "0 8,18 * * *" {
		t.Errorf("cron deveria ser '0 8,18 * * *', veio %q", b.Cron)
	}
}

func TestHealthRetornaStoreELogs(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/health", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["store"] != "spy" {
		t.Errorf("health deveria expor store, veio %v", resp["store"])
	}
	if resp["logs"] == nil || resp["logs"] == "" {
		t.Errorf("health deveria expor info de logs")
	}
}
