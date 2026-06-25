package httpapi

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/scheduler"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// --- Testes de coleta com rotação e lojas monitoradas ─────────────────────

func TestColetarComBuscaIDAplicaRotacao(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")
	sp := &spyStore{
		buscas: []store.Busca{
			{
				ID:       "loja-99999",
				ShopIDs:  []int64{99999},
				Ativo:    true,
				OwnerUID: "test-user",
				RotationCursor: map[int64]int{99999: 3}, // deve começar na página 3
			},
		},
	}

	// Precisamos de uma fonte shopee-shop real (com mock HTTP) para testar rotação.
	// Aqui testamos que o handler aceita busca_id e não falha.
	srv := &Server{
		Eventos:   sp,
		Auth:      fakeVerifier{},
		Scheduler: nopSched{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake|shop"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "POST", "/api/coletar?fonte=shopee-shop&shop_ids=99999&busca_id=loja-99999&top=5",
		nil, map[string]string{"X-Garimpo-Token": "segredo"})
	if rec.Code != 202 {
		t.Fatalf("esperava 202, veio %d: %s", rec.Code, rec.Body.String())
	}

	// Snapshot deve ter sido gravado
	if len(sp.snapshots) != 1 {
		t.Errorf("esperava 1 snapshot, veio %d", len(sp.snapshots))
	}
}

func TestColetarSemBuscaIDFuncionaNormalmente(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")
	sp := &spyStore{}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})

	rec := req(t, h, "POST", "/api/coletar?top=3", nil, map[string]string{"X-Garimpo-Token": "segredo"})
	if rec.Code != 202 {
		t.Fatalf("esperava 202, veio %d: %s", rec.Code, rec.Body.String())
	}
	if len(sp.snapshots) != 1 {
		t.Errorf("esperava 1 snapshot, veio %d", len(sp.snapshots))
	}
}

func TestColetarBuscaIDInexistenteNaoQuebraRotacao(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")
	// Busca não existe no store — rotação não deve ser aplicada, mas não deve crashar
	sp := &spyStore{}
	srv := &Server{
		Eventos:   sp,
		Auth:      fakeVerifier{},
		Scheduler: nopSched{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake|shop"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "POST", "/api/coletar?fonte=shopee-shop&shop_ids=11111&busca_id=naoexiste&top=3",
		nil, map[string]string{"X-Garimpo-Token": "segredo"})
	if rec.Code != 202 {
		t.Fatalf("esperava 202, veio %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Testes de publicações preservam título ───────────────────────────────

func TestPublicacaoAgendadaPreservaTituloAoEnviar(t *testing.T) {
	t.Setenv("COLETA_TOKEN", "segredo")
	sp := &spyStore{
		publicacoes: []store.Publicacao{
			{
				ID:         "pub-titulo-1",
				ProdutoID:  "P1",
				Nome:       "Sérum Vitamina C",
				Preco:      89.90,
				Comissao:   0.15,
				Link:       "http://l",
				Estrategia: "nicho",
				Status:     "agendada",
				AgendadaEm: "2020-01-01T10:00:00Z", // no passado → será enviada
				OwnerUID:   "test-user",
			},
		},
	}
	pub := &spyPub{}
	srv := &Server{
		Eventos:    sp,
		Publicador: pub,
		Auth:       fakeVerifier{},
		Scheduler:  nopSched{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "POST", "/api/publicar-pendentes", nil, map[string]string{"X-Garimpo-Token": "segredo"})
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	// Após publicar, a publicação deve manter o nome
	for _, p := range sp.publicacoes {
		if p.ID == "pub-titulo-1" {
			if p.Nome != "Sérum Vitamina C" {
				t.Errorf("nome deveria ser 'Sérum Vitamina C' após envio, veio %q", p.Nome)
			}
			if p.Status != "enviada" {
				t.Errorf("status deveria ser 'enviada', veio %q", p.Status)
			}
			if p.Preco != 89.90 {
				t.Errorf("preco deveria ser 89.90, veio %.2f", p.Preco)
			}
			break
		}
	}
}

// --- Testes do endpoint /api/coletas ──────────────────────────────────────

func TestColetasExigeNada(t *testing.T) {
	// /api/coletas é público (não exige auth)
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/coletas?dias=7", nil, nil)
	if rec.Code != 200 {
		t.Errorf("esperava 200, veio %d", rec.Code)
	}
}

// --- Teste de ListarBuscas com ShopIDs (cenário do bug) ──────────────────

func TestListarBuscasRetornaShopIDs(t *testing.T) {
	sp := &spyStore{
		buscas: []store.Busca{
			{ID: "loja-123", Keywords: []string{}, ShopIDs: []int64{123456}, Ativo: true, OwnerUID: "test-user"},
			{ID: "perfume", Keywords: []string{"perfume"}, Ativo: true, OwnerUID: "test-user"},
		},
	}
	h := montar(&fonteFake{produtos: amostra}, sp, &spyPub{})
	rec := req(t, h, "GET", "/api/buscas", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	var resp struct {
		Buscas []struct {
			ID      string  `json:"id"`
			ShopIDs []int64 `json:"shop_ids"`
		} `json:"buscas"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Buscas) != 2 {
		t.Fatalf("esperava 2 buscas, veio %d", len(resp.Buscas))
	}
	// Busca com ShopIDs deve retornar os IDs
	found := false
	for _, b := range resp.Buscas {
		if b.ID == "loja-123" {
			found = true
			if len(b.ShopIDs) != 1 || b.ShopIDs[0] != 123456 {
				t.Errorf("shop_ids da loja-123 deveria ser [123456], veio %v", b.ShopIDs)
			}
		}
	}
	if !found {
		t.Error("loja-123 não encontrada nas buscas retornadas")
	}
}

// --- Testes de alertas ────────────────────────────────────────────────────

func TestAlertasConfigExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/alertas", nil, nil)
	if rec.Code != 401 {
		t.Errorf("sem auth deveria dar 401, veio %d", rec.Code)
	}
}

func TestAlertasConfigComAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/alertas", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != 200 {
		t.Fatalf("com auth deveria dar 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Ativo     bool    `json:"ativo"`
		Threshold float64 `json:"threshold"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	// Sem ALERTAS_TELEGRAM_CHAT_ID, ativo deve ser false
	if resp.Ativo {
		t.Error("sem env vars, alertas deveria estar inativo")
	}
	if resp.Threshold != 0.15 {
		t.Errorf("threshold default deveria ser 0.15, veio %f", resp.Threshold)
	}
}

func TestAlertasTestarSemConfigRetornaErro(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "POST", "/api/alertas/testar", nil,
		map[string]string{"Authorization": "Bearer tok", "Content-Type": "application/json"})
	if rec.Code != 400 {
		t.Errorf("sem config deveria dar 400, veio %d", rec.Code)
	}
}

func TestAlertasConfigurarAtualizaThreshold(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	corpo := []byte(`{"threshold":0.20,"apenas_quedas":false}`)
	rec := req(t, h, "POST", "/api/alertas/configurar", corpo,
		map[string]string{"Authorization": "Bearer tok", "Content-Type": "application/json"})
	if rec.Code != 200 {
		t.Fatalf("esperava 200, veio %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Threshold    float64 `json:"threshold"`
		ApenasQuedas bool    `json:"apenas_quedas"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Threshold != 0.20 {
		t.Errorf("threshold deveria ser 0.20, veio %f", resp.Threshold)
	}
	if resp.ApenasQuedas {
		t.Error("apenas_quedas deveria ser false após update")
	}
}

func TestAlertasConfigurarExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "POST", "/api/alertas/configurar", []byte(`{}`),
		map[string]string{"Content-Type": "application/json"})
	if rec.Code != 401 {
		t.Errorf("sem auth deveria dar 401, veio %d", rec.Code)
	}
}

// --- helper ──────────────────────────────────────────────────────────────

// nopSched é um scheduler que não faz nada (para testes que não precisam do scheduler mock).
type nopSched = scheduler.NopScheduler
