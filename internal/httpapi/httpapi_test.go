package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
}

func (s *spyStore) Registrar(_ context.Context, e store.Evento) error {
	s.eventos = append(s.eventos, e)
	return nil
}
func (s *spyStore) RegistrarSnapshot(_ context.Context, snap store.Snapshot) error {
	s.snapshots = append(s.snapshots, snap)
	return nil
}
func (s *spyStore) Nome() string { return "spy" }

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

var amostra = []domain.Product{
	{ID: "P1", Name: "Sérum", Category: "cosméticos", Price: 100, Commission: 0.15, Sales30d: 80, Rating: 4.8},
	{ID: "P2", Name: "Fone", Category: "eletrônicos", Price: 100, Commission: 0.10, Sales30d: 900, Rating: 4.3},
	{ID: "P3", Name: "Creme", Category: "cosméticos", Price: 50, Commission: 0.05, Sales30d: 300, Rating: 4.9}, // 5% -> fora
}

func montar(fonte *fonteFake, ev store.EventoStore, pub publish.Publicador) http.Handler {
	srv := &Server{
		Eventos:    ev,
		Publicador: pub,
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
