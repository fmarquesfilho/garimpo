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
	eventos     []store.Evento
	snapshots   []store.Snapshot
	buscas      []store.Busca
	publicacoes []store.Publicacao
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
func (s *spyStore) Conversoes(_ context.Context, _ int) ([]store.ConversaoResumo, error) {
	return nil, nil
}
func (s *spyStore) SalvarPublicacao(_ context.Context, p store.Publicacao) error {
	// Upsert: se já existe, atualiza; senão, appenda.
	for i, existing := range s.publicacoes {
		if existing.ID == p.ID {
			s.publicacoes[i] = p
			return nil
		}
	}
	s.publicacoes = append(s.publicacoes, p)
	return nil
}
func (s *spyStore) ListarPublicacoes(_ context.Context, status string) ([]store.Publicacao, error) {
	if status == "" {
		return s.publicacoes, nil
	}
	var filtrada []store.Publicacao
	for _, p := range s.publicacoes {
		if p.Status == status {
			filtrada = append(filtrada, p)
		}
	}
	return filtrada, nil
}
func (s *spyStore) AtualizarPublicacao(_ context.Context, id, status, detalhe string) error {
	for i, p := range s.publicacoes {
		if p.ID == id {
			s.publicacoes[i].Status = status
			s.publicacoes[i].Detalhe = detalhe
			break
		}
	}
	return nil
}
func (s *spyStore) Novidades(_ context.Context, buscaID string, dias int) (store.NovidadesLojas, error) {
	return store.NovidadesLojas{BuscaID: buscaID, DiasJanela: dias}, nil
}
func (s *spyStore) EvolucaoLojas(_ context.Context, dias int) (store.EvolucaoLojasResult, error) {
	return store.EvolucaoLojasResult{DiasJanela: dias}, nil
}
func (s *spyStore) SalvarFavorito(_ context.Context, _ store.Favorito) error { return nil }
func (s *spyStore) ListarFavoritos(_ context.Context, _ string) ([]store.Favorito, error) {
	return nil, nil
}
func (s *spyStore) RemoverFavorito(_ context.Context, _, _ string) error { return nil }
func (s *spyStore) EnsureSchema(_ context.Context) error                 { return nil }

// spyRepo wraps spyStore to satisfy store.Repository.
type spyRepo struct {
	sp *spyStore
}

func (r *spyRepo) Eventos() store.EventoRepo          { return r.sp }
func (r *spyRepo) Snapshots() store.SnapshotRepo      { return r.sp }
func (r *spyRepo) Buscas() store.BuscaRepo            { return r.sp }
func (r *spyRepo) Publicacoes() store.PublicacaoRepo  { return r.sp }
func (r *spyRepo) Destinos() store.DestinoRepo        { return store.NovoNopRepository().Destinos() }
func (r *spyRepo) Templates() store.TemplateRepo      { return store.NovoNopRepository().Templates() }
func (r *spyRepo) Favoritos() store.FavoritoRepo      { return r.sp }
func (r *spyRepo) Tenants() store.TenantRepo          { return store.NovoMemTenantRepo() }
func (r *spyRepo) EnsureSchema(context.Context) error { return nil }
func (r *spyRepo) Nome() string                       { return "spy" }

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

func montar(fonte *fonteFake, repo store.Repository, pub publish.Publicador) http.Handler {
	srv := &Server{
		Repo:       repo,
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
	// Sempre envia token para o fakeVerifier (simula usuário autenticado),
	// a menos que o caller passe "Authorization" explicitamente no mapa.
	if _, explicit := headers["Authorization"]; !explicit {
		r.Header.Set("Authorization", "Bearer fake-token")
	}
	for k, v := range headers {
		if v != "" {
			r.Header.Set(k, v)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// reqSemAuth faz request sem header Authorization (testa rejeição por middleware).
func reqSemAuth(t *testing.T, h http.Handler, metodo, alvo string, corpo []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Authorization"] = "" // marca presença para o req() não setar default
	var r *http.Request
	if corpo != nil {
		r = httptest.NewRequest(metodo, alvo, bytes.NewReader(corpo))
	} else {
		r = httptest.NewRequest(metodo, alvo, nil)
	}
	for k, v := range headers {
		if v != "" {
			r.Header.Set(k, v)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// --- testes ---------------------------------------------------------------

func TestCandidatosRankeiaEFiltra(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, &spyPub{})
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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "GET", "/api/eventos", nil, nil)
	// Com SPA handler, rotas /api/* sem handler retornam 404 ou 405.
	if rec.Code != 405 && rec.Code != 404 {
		t.Errorf("GET em /api/eventos deveria dar 404 ou 405, veio %d", rec.Code)
	}
}

func TestPublicarChamaPublicadorERegistra(t *testing.T) {
	sp := &spyStore{}
	pub := &spyPub{}
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, pub)
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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, &spyPub{})

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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, &spyPub{})

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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, &spyPub{})

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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	// POST sem auth → 401
	rec := reqSemAuth(t, h, "POST", "/api/buscas", []byte(`{"keywords":["x"]}`), map[string]string{"Content-Type": "application/json"})
	if rec.Code != 401 {
		t.Errorf("busca sem auth deveria dar 401, veio %d", rec.Code)
	}
	// GET sem auth → 401 (middleware rejeita antes do handler)
	rec = reqSemAuth(t, h, "GET", "/api/buscas", nil, nil)
	if rec.Code != 401 {
		t.Errorf("GET sem auth deveria dar 401, veio %d", rec.Code)
	}
}

func TestBuscasExigeAlgumCriterio(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	// sem keywords, lojas, categorias ou fontes → 400
	rec := req(t, h, "POST", "/api/buscas", []byte(`{}`), map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"})
	if rec.Code != 400 {
		t.Errorf("busca sem critérios deveria dar 400, veio %d", rec.Code)
	}
}

func TestBuscasRemoverMarcaInativo(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, &spyPub{})
	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer fake-token"}
	req(t, h, "POST", "/api/buscas?remover", []byte(`{"id":"minha-busca","keywords":["x"]}`), authH)
	if len(sp.buscas) != 1 || sp.buscas[0].Ativo {
		t.Errorf("remover deveria gravar tombstone (ativo=false): %+v", sp.buscas)
	}
}

func TestEstatisticasRetornaResumo(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
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
	h := montar(f, &spyRepo{sp: &spyStore{}}, &spyPub{})
	req(t, h, "GET", "/api/candidatos?estrategia=nicho", nil, nil)
	req(t, h, "GET", "/api/candidatos?estrategia=diversificada", nil, nil)
	req(t, h, "GET", "/api/comparar", nil, nil)
	if f.chamadas != 1 {
		t.Errorf("a fonte deveria ser buscada 1 vez (cache), veio %d", f.chamadas)
	}
}

func TestBuscasMultiKeyword(t *testing.T) {
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: sp}, &spyPub{})

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
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
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

// --- Testes de destinos ---------------------------------------------------

func TestDestinosCRUD(t *testing.T) {
	srv := &Server{
		Repo:       store.NovoNopRepository(),
		Publicador: &spyPub{},
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()
	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer tok"}

	// POST cria destino
	corpo := []byte(`{"nome":"Ofertas Beleza","tipo":"telegram","config":"@beleza"}`)
	rec := req(t, h, "POST", "/api/destinos", corpo, authH)
	if rec.Code != 201 {
		t.Fatalf("POST /api/destinos esperava 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	// GET lista destinos
	rec = req(t, h, "GET", "/api/destinos", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("GET /api/destinos esperava 200, veio %d", rec.Code)
	}
	var resp struct {
		Destinos []publish.Destino `json:"destinos"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Destinos) != 1 || resp.Destinos[0].Config != "@beleza" {
		t.Errorf("listagem inesperada: %+v", resp.Destinos)
	}

	// DELETE remove destino
	rec = req(t, h, "DELETE", "/api/destinos?id=ofertas-beleza", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("DELETE esperava 200, veio %d: %s", rec.Code, rec.Body.String())
	}

	// GET após delete → vazio
	rec = req(t, h, "GET", "/api/destinos", nil, map[string]string{"Authorization": "Bearer tok"})
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Destinos) != 0 {
		t.Errorf("após delete deveria estar vazio: %+v", resp.Destinos)
	}
}

func TestDestinosExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := reqSemAuth(t, h, "GET", "/api/destinos", nil, nil)
	if rec.Code != 401 {
		t.Errorf("GET sem auth deveria dar 401, veio %d", rec.Code)
	}
}

// --- Testes de templates ---------------------------------------------------

func TestTemplatesCRUD(t *testing.T) {
	srv := &Server{
		Repo:       store.NovoNopRepository(),
		Publicador: &spyPub{},
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()
	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer tok"}

	// GET lista templates (padrão embutidos)
	rec := req(t, h, "GET", "/api/templates", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("GET esperava 200, veio %d", rec.Code)
	}
	var resp struct {
		Templates []publish.Template `json:"templates"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Templates) < 2 {
		t.Errorf("deveria ter ao menos 2 templates padrão: %d", len(resp.Templates))
	}

	// POST cria template
	corpo := []byte(`{"nome":"Promoção Flash","corpo":"🔥 <b>{{nome}}</b> por {{preco}}!","com_foto":true}`)
	rec = req(t, h, "POST", "/api/templates", corpo, authH)
	if rec.Code != 201 {
		t.Fatalf("POST esperava 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	// DELETE remove
	rec = req(t, h, "DELETE", "/api/templates?id=promocao-flash", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("DELETE esperava 200, veio %d", rec.Code)
	}
}

func TestTemplatePreview(t *testing.T) {
	srv := &Server{
		Repo:       store.NovoNopRepository(),
		Publicador: &spyPub{},
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	corpo := []byte(`{"template_id":"padrao","nome":"Sérum Vitamina C","preco":89.90,"categoria":"Beleza"}`)
	rec := req(t, h, "POST", "/api/templates/preview", corpo, map[string]string{"Content-Type": "application/json"})
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Preview string `json:"preview"`
		ComFoto bool   `json:"com_foto"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Preview == "" {
		t.Error("preview não deveria ser vazio")
	}
	if resp.ComFoto {
		t.Error("template 'padrao' não deveria ter com_foto=true")
	}
}

// --- Testes de publicações -------------------------------------------------

func TestPublicacoesAgendarImediato(t *testing.T) {
	sp := &spyStore{}
	pub := &spyPub{}
	srv := &Server{
		Repo:       &spyRepo{sp: sp},
		Publicador: pub,
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()
	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer tok"}

	corpo := []byte(`{"nome":"Sérum","preco":100,"comissao":0.15,"link":"http://l","estrategia":"nicho"}`)
	rec := req(t, h, "POST", "/api/publicacoes", corpo, authH)
	if rec.Code != 201 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Publicacao struct {
			Status  string `json:"status"`
			Detalhe string `json:"detalhe"`
		} `json:"publicacao"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Publicacao.Status != "enviada" {
		t.Errorf("sem agendada_em deveria enviar imediatamente, status=%q", resp.Publicacao.Status)
	}
	if pub.chamadas != 1 {
		t.Errorf("publicador deveria ter sido chamado 1 vez, veio %d", pub.chamadas)
	}
}

func TestPublicacoesAgendar(t *testing.T) {
	sp := &spyStore{}
	pub := &spyPub{}
	srv := &Server{
		Repo:       &spyRepo{sp: sp},
		Publicador: pub,
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()
	authH := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer tok"}

	corpo := []byte(`{"nome":"Sérum","preco":100,"estrategia":"nicho","agendada_em":"2026-12-25T10:00:00Z"}`)
	rec := req(t, h, "POST", "/api/publicacoes", corpo, authH)
	if rec.Code != 201 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Publicacao struct {
			Status     string `json:"status"`
			AgendadaEm string `json:"agendada_em"`
		} `json:"publicacao"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Publicacao.Status != "agendada" {
		t.Errorf("com agendada_em deveria ficar agendada, status=%q", resp.Publicacao.Status)
	}
	if pub.chamadas != 0 {
		t.Errorf("agendada NÃO deveria chamar o publicador, chamadas=%d", pub.chamadas)
	}
}

func TestPublicacoesExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := reqSemAuth(t, h, "GET", "/api/publicacoes", nil, nil)
	if rec.Code != 401 {
		t.Errorf("GET sem auth deveria dar 401, veio %d", rec.Code)
	}
	rec = reqSemAuth(t, h, "POST", "/api/publicacoes", []byte(`{"nome":"x"}`),
		map[string]string{"Content-Type": "application/json"})
	if rec.Code != 401 {
		t.Errorf("POST sem auth deveria dar 401, veio %d", rec.Code)
	}
}

// --- Testes de conversões --------------------------------------------------

func TestConversoesExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := reqSemAuth(t, h, "GET", "/api/conversoes", nil, nil)
	if rec.Code != 401 {
		t.Errorf("GET sem auth deveria dar 401, veio %d", rec.Code)
	}
}

func TestConversoesComAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "GET", "/api/conversoes", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Testes de publicar com template e imagem ------------------------------

func TestPublicarComDestinoIDETemplateID(t *testing.T) {
	pub := &spyPub{}
	repo := store.NovoNopRepository()
	_ = repo.Destinos().SalvarDestino(context.Background(), store.Destino{
		ID: "beleza", Nome: "Beleza", Tipo: "telegram", Config: "@beleza", Ativo: true,
	})

	srv := &Server{
		Repo:       repo,
		Publicador: pub,
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	corpo := []byte(`{"id":"P1","nome":"Sérum","preco":100,"link":"http://l","estrategia":"nicho","destino_id":"beleza","template_id":"padrao","imagem":"http://img.jpg","legenda_custom":"<b>Oferta!</b>"}`)
	rec := req(t, h, "POST", "/api/publicar", corpo,
		map[string]string{"Content-Type": "application/json"})
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	// Imagem é sempre mantida (user decide no frontend)
	if pub.ultima.Imagem != "http://img.jpg" {
		t.Errorf("imagem deveria ser mantida, veio: %q", pub.ultima.Imagem)
	}
	if pub.ultima.DestinoID != "beleza" {
		t.Errorf("destino_id deveria ser 'beleza', veio %q", pub.ultima.DestinoID)
	}
	// Legenda custom é enviada
	if pub.ultima.LegendaHTML != "<b>Oferta!</b>" {
		t.Errorf("legenda_custom deveria ser '<b>Oferta!</b>', veio %q", pub.ultima.LegendaHTML)
	}
}

func TestPublicarComTemplateFotoMantemImagem(t *testing.T) {
	pub := &spyPub{}
	srv := &Server{
		Repo:       store.NovoNopRepository(),
		Publicador: pub,
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	corpo := []byte(`{"id":"P1","nome":"Sérum","preco":100,"link":"http://l","estrategia":"nicho","template_id":"foto","imagem":"http://img.jpg"}`)
	rec := req(t, h, "POST", "/api/publicar", corpo,
		map[string]string{"Content-Type": "application/json"})
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	// Template "foto" tem com_foto=true → imagem deve ser mantida
	if pub.ultima.Imagem != "http://img.jpg" {
		t.Errorf("template com foto deveria manter imagem, mas ficou: %q", pub.ultima.Imagem)
	}
}

func TestPublicarPendentesExecutaAgendadasVencidas(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")

	sp := &spyStore{
		publicacoes: []store.Publicacao{
			{ID: "pub-1", Nome: "Produto A", Status: "agendada", AgendadaEm: "2020-01-01T10:00:00Z", Estrategia: "nicho"},
			{ID: "pub-2", Nome: "Produto B", Status: "agendada", AgendadaEm: "2099-12-31T23:59:59Z", Estrategia: "nicho"},
			{ID: "pub-3", Nome: "Produto C", Status: "enviada", AgendadaEm: "2020-01-01T10:00:00Z", Estrategia: "nicho"},
		},
	}
	pub := &spyPub{}
	srv := &Server{
		Repo:       &spyRepo{sp: sp},
		Publicador: pub,
		Auth:       fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "POST", "/api/publicar-pendentes", nil, map[string]string{"X-Garimpo-Token": "segredo"})
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Enviadas int `json:"enviadas"`
		Erros    int `json:"erros"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// pub-1: agendada no passado → deve ser publicada
	// pub-2: agendada no futuro → NÃO deve ser publicada
	// pub-3: já enviada → não aparece no filtro "agendada"
	if resp.Enviadas != 1 {
		t.Errorf("esperava 1 enviada, veio %d", resp.Enviadas)
	}
	if pub.chamadas != 1 {
		t.Errorf("publicador deveria ser chamado 1 vez, veio %d", pub.chamadas)
	}

	// Verifica que o status foi atualizado
	for _, p := range sp.publicacoes {
		if p.ID == "pub-1" && p.Status != "enviada" {
			t.Errorf("pub-1 deveria ter status=enviada, veio %q", p.Status)
		}
		if p.ID == "pub-2" && p.Status != "agendada" {
			t.Errorf("pub-2 deveria continuar agendada, veio %q", p.Status)
		}
	}
}

func TestPublicarPendentesExigeToken(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "POST", "/api/publicar-pendentes", nil, nil)
	if rec.Code != 401 {
		t.Errorf("sem token deveria dar 401, veio %d", rec.Code)
	}
}

// --- Testes de lojas/novidades e sem_filtro --------------------------------

func TestCandidatosComSemFiltro(t *testing.T) {
	// Sem sem_filtro: aplica piso de comissão 7% (P3 com 5% cai fora)
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "GET", "/api/candidatos?estrategia=nicho", nil, nil)
	var resp struct {
		Candidatos []struct{ ID string } `json:"candidatos"`
		TotalBruto int                   `json:"total_bruto"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	semFiltroCount := len(resp.Candidatos)

	// Com sem_filtro=true: todos passam
	rec = req(t, h, "GET", "/api/candidatos?estrategia=nicho&sem_filtro=true", nil, nil)
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Candidatos) <= semFiltroCount {
		t.Errorf("sem_filtro=true deveria retornar mais candidatos: sem=%d com=%d", semFiltroCount, len(resp.Candidatos))
	}
	if resp.TotalBruto != 3 {
		t.Errorf("total_bruto deveria ser 3 (todos da amostra), veio %d", resp.TotalBruto)
	}
}

func TestCandidatosTotalBruto(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "GET", "/api/candidatos?estrategia=nicho", nil, nil)
	var resp struct {
		TotalBruto int `json:"total_bruto"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.TotalBruto != 3 {
		t.Errorf("total_bruto deveria ser 3, veio %d", resp.TotalBruto)
	}
}

func TestNovidadesExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := reqSemAuth(t, h, "GET", "/api/lojas/novidades", nil, nil)
	if rec.Code != 401 {
		t.Errorf("sem auth deveria dar 401, veio %d", rec.Code)
	}
}

func TestNovidadesComAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "GET", "/api/lojas/novidades?busca_id=teste&dias=7", nil,
		map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("com auth deveria dar 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		BuscaID    string `json:"busca_id"`
		DiasJanela int    `json:"dias_janela"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.BuscaID != "teste" {
		t.Errorf("busca_id deveria ser 'teste', veio %q", resp.BuscaID)
	}
	if resp.DiasJanela != 7 {
		t.Errorf("dias_janela deveria ser 7, veio %d", resp.DiasJanela)
	}
}

func TestCompararRetornaDuasListas(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyRepo{sp: &spyStore{}}, &spyPub{})
	rec := req(t, h, "GET", "/api/comparar?top=5", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var resp struct {
		Nicho         []map[string]any `json:"nicho"`
		Diversificada []map[string]any `json:"diversificada"`
		Fonte         string           `json:"fonte"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Fonte != "fake" {
		t.Errorf("fonte deveria ser 'fake', veio %q", resp.Fonte)
	}
	if len(resp.Nicho) == 0 {
		t.Error("nicho não deveria estar vazio")
	}
	if len(resp.Diversificada) == 0 {
		t.Error("diversificada não deveria estar vazio")
	}
}
