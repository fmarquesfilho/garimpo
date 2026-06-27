package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// ── Testes de extrairShopIDDoLink ────────────────────────────────────────────

func TestExtrairShopIDDoLink(t *testing.T) {
	cases := []struct {
		name string
		link string
		want string
	}{
		{"link vazio", "", ""},
		{"link sem padrão", "https://shopee.com.br/alguma-coisa", ""},
		{"offerLink curto (shope.ee)", "https://shope.ee/abc123", ""},
		{"productLink padrão", "https://shopee.com.br/Serum-Coreano-SKIN1004-i.785541033.23720954312", "785541033"},
		{"productLink com query params", "https://shopee.com.br/Serum-i.123456.789?sp_atk=xxx", "123456"},
		{"productLink com nome complexo", "https://shopee.com.br/Kit-3-Produtos-Skin-Care-Coreano-Original-i.920174533.19044369472", "920174533"},
		{"productLink só com -i.", "https://shopee.com.br/-i.111.222", "111"},
		{"múltiplos -i. no nome (edge case)", "https://shopee.com.br/Gel-Anti-i.mpurezas-i.555.666", "555"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extrairShopIDDoLink(tc.link)
			if got != tc.want {
				t.Errorf("extrairShopIDDoLink(%q) = %q, want %q", tc.link, got, tc.want)
			}
		})
	}
}

// ── Testes de toDTO com cenários de origem ───────────────────────────────────

func TestToDTO_OrigemELojaID(t *testing.T) {
	cases := []struct {
		name       string
		product    domain.Product
		wantLojaID string
		wantOrigem string
	}{
		{
			name: "shopId preenchido direto pela API",
			product: domain.Product{
				ID: "111", Name: "Sérum", ShopID: "785541033",
				Link: "https://shope.ee/aff123", ProductLink: "https://shopee.com.br/X-i.785541033.111",
			},
			wantLojaID: "785541033",
			wantOrigem: "",
		},
		{
			name: "shopId é 0 — fallback para productLink",
			product: domain.Product{
				ID: "222", Name: "Tônico", ShopID: "0",
				Link: "https://shope.ee/aff456", ProductLink: "https://shopee.com.br/Tonico-i.920174533.222",
			},
			wantLojaID: "920174533",
			wantOrigem: "",
		},
		{
			name: "shopId vazio — fallback para offerLink (sem -i.)",
			product: domain.Product{
				ID: "333", Name: "Creme", ShopID: "",
				Link: "https://shope.ee/aff789", ProductLink: "",
			},
			wantLojaID: "",
			wantOrigem: "",
		},
		{
			name: "shopId vazio — fallback para link com -i.",
			product: domain.Product{
				ID: "444", Name: "Máscara", ShopID: "",
				Link: "https://shopee.com.br/Mascara-i.100200.444", ProductLink: "",
			},
			wantLojaID: "100200",
			wantOrigem: "",
		},
		{
			name: "origin preenchido pelo fallback (origem_padrao da busca)",
			product: domain.Product{
				ID: "555", Name: "Essence", ShopID: "123",
				Origin: "Coreia",
			},
			wantLojaID: "123",
			wantOrigem: "Coreia",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scored := domain.Scored{Product: tc.product, Score: 0.5}
			dto := toDTO(scored)

			if dto.LojaID != tc.wantLojaID {
				t.Errorf("LojaID = %q, want %q", dto.LojaID, tc.wantLojaID)
			}
			if dto.Origem != tc.wantOrigem {
				t.Errorf("Origem = %q, want %q", dto.Origem, tc.wantOrigem)
			}
		})
	}
}

// ── Testes do endpoint /api/candidatos com origem ────────────────────────────

func TestCandidatosRetornaLojaIDELinkProduto(t *testing.T) {
	produtos := []domain.Product{
		{
			ID: "P1", Name: "Sérum Coreano", Category: "cosméticos",
			Price: 100, Commission: 0.15, Sales30d: 80, Rating: 4.8,
			ShopID: "785541033", ShopName: "SKIN1004",
			Link:        "https://shope.ee/aff111",
			ProductLink: "https://shopee.com.br/Serum-i.785541033.111",
		},
	}

	fonte := &fonteFake{produtos: produtos}
	srv := &Server{
		Eventos: &spyStore{},
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return fonte, "fake|fixo"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "GET", "/api/candidatos?estrategia=nicho&sem_filtro=true", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Candidatos []struct {
			ID          string `json:"id"`
			LojaID      string `json:"loja_id"`
			LinkProduto string `json:"link_produto"`
			Loja        string `json:"loja"`
			Origem      string `json:"origem"`
		} `json:"candidatos"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Candidatos) == 0 {
		t.Fatal("esperava pelo menos 1 candidato")
	}
	c := resp.Candidatos[0]
	if c.LojaID != "785541033" {
		t.Errorf("loja_id = %q, want '785541033'", c.LojaID)
	}
	if c.LinkProduto != "https://shopee.com.br/Serum-i.785541033.111" {
		t.Errorf("link_produto = %q, want URL com -i.", c.LinkProduto)
	}
	if c.Loja != "SKIN1004" {
		t.Errorf("loja = %q, want 'SKIN1004'", c.Loja)
	}
}

// ── Testes do endpoint /api/produto/origem com mock da API pública ───────────

func TestProdutoOrigemComMockShopee(t *testing.T) {
	// Mock da API pública v4 da Shopee
	mockShopee := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simula resposta com atributos de origem
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"item": map[string]any{
					"brand": "SKIN1004",
					"attributes": []map[string]string{
						{"name": "País de Origem", "value": "Coreia do Sul"},
						{"name": "Marca", "value": "SKIN1004"},
					},
				},
			},
		})
	}))
	defer mockShopee.Close()

	// Para testar, precisamos que buscarOrigemProdutoShopee use o mock.
	// Como a função usa URL hardcoded, testamos via cache + endpoint diretamente.
	// Primeiro, teste com cache pre-populado (já testado acima).
	// Aqui testamos que o endpoint responde corretamente com cache.

	salvarOrigemNoCache("999:777", origemCacheEntry{Origem: "Coreia", Marca: "SKIN1004"})

	srv := &Server{Eventos: &spyStore{}, Auth: fakeVerifier{}}
	h := srv.Handler()

	rec := httptest.NewRequest("GET", "/api/produto/origem?item_id=777&shop_id=999", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, rec)

	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	var resp origemResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Origem != "Coreia" {
		t.Errorf("origem = %q, want 'Coreia'", resp.Origem)
	}
	if resp.Marca != "SKIN1004" {
		t.Errorf("marca = %q, want 'SKIN1004'", resp.Marca)
	}
	if resp.Fonte != "cache" {
		t.Errorf("fonte = %q, want 'cache'", resp.Fonte)
	}
}

func TestProdutoOrigemSemShopIdRetorna400(t *testing.T) {
	srv := &Server{Eventos: &spyStore{}, Auth: fakeVerifier{}}
	h := srv.Handler()

	// Sem item_id
	rec := httptest.NewRequest("GET", "/api/produto/origem?shop_id=123", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, rec)
	if w.Code != 400 {
		t.Errorf("sem item_id deveria dar 400, veio %d", w.Code)
	}

	// Sem shop_id
	rec = httptest.NewRequest("GET", "/api/produto/origem?item_id=456", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rec)
	if w.Code != 400 {
		t.Errorf("sem shop_id deveria dar 400, veio %d", w.Code)
	}

	// Ambos vazios
	rec = httptest.NewRequest("GET", "/api/produto/origem", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, rec)
	if w.Code != 400 {
		t.Errorf("sem params deveria dar 400, veio %d", w.Code)
	}
}

// ── Teste de NormalizarOrigemProduto (reforço) ───────────────────────────────

func TestNormalizarOrigemProduto_VariacoesReais(t *testing.T) {
	// Valores que a API pública real pode retornar (baseado no screenshot da Mileny)
	cases := []struct {
		input string
		want  string
	}{
		{"Coreia", "Coreia"},
		{"Coréia", "Coreia"},
		{"Coreia do Sul", "Coreia"},
		{"Coréia do Sul", "Coreia"},     // acento variante
		{"  Coreia do Sul  ", "Coreia"}, // espaços
		{"COREIA DO SUL", "Coreia"},     // uppercase
		{"Korea", "Coreia"},
		{"South Korea", "Coreia"},
		{"Japão", "Japão"},
		{"Japan", "Japão"},
		{"China", "China"},
		{"Mainland China", "China"},
		{"Brasil", "Brasil"},
		{"", ""},
		{"Tailândia", "Tailândia"},
		{"Thailand", "Tailândia"},
		{"Indonesia", "Indonésia"},
	}
	for _, tc := range cases {
		got := NormalizarOrigemProduto(tc.input)
		if got != tc.want {
			t.Errorf("NormalizarOrigemProduto(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── Teste de integração: busca por keyword enriquece origem via loja monitorada ──

func TestCandidatosEnriqueceOrigemDaLojaMonitorada(t *testing.T) {
	// Cenário: loja 785541033 está monitorada com origem_padrao="Coreia"
	// Produto da busca por keyword pertence a essa loja (shopId=785541033)
	// → badge deve aparecer mesmo sem coleta service

	sp := &spyStore{
		buscas: []store.Busca{
			{
				ID:           "loja-785541033",
				ShopIDs:      []int64{785541033},
				OrigemPadrao: "Coreia",
				Ativo:        true,
				OwnerUID:     "test-user",
			},
		},
	}

	produtos := []domain.Product{
		{
			ID: "ITEM1", Name: "Sérum Coreano SKIN1004", Category: "beleza",
			Price: 80, Commission: 0.12, Sales30d: 200, Rating: 4.9,
			ShopID: "785541033", ShopName: "SKIN1004 Official",
			Link:        "https://shope.ee/aff111",
			ProductLink: "https://shopee.com.br/Serum-i.785541033.111",
		},
		{
			ID: "ITEM2", Name: "Fone Bluetooth Genérico", Category: "eletrônicos",
			Price: 50, Commission: 0.10, Sales30d: 500, Rating: 4.2,
			ShopID: "999999999", ShopName: "Loja Aleatória",
			Link:        "https://shope.ee/aff222",
			ProductLink: "https://shopee.com.br/Fone-i.999999999.222",
		},
	}

	fonte := &fonteFake{produtos: produtos}
	srv := &Server{
		Eventos: sp,
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return fonte, "fake|fixo"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "GET", "/api/candidatos?sem_filtro=true", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Candidatos []struct {
			ID     string `json:"id"`
			LojaID string `json:"loja_id"`
			Origem string `json:"origem"`
			Loja   string `json:"loja"`
		} `json:"candidatos"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Candidatos) < 2 {
		t.Fatalf("esperava 2 candidatos, veio %d", len(resp.Candidatos))
	}

	// Encontra cada produto por ID
	var coreano, generico *struct {
		ID     string `json:"id"`
		LojaID string `json:"loja_id"`
		Origem string `json:"origem"`
		Loja   string `json:"loja"`
	}
	for i := range resp.Candidatos {
		if resp.Candidatos[i].ID == "ITEM1" {
			coreano = &resp.Candidatos[i]
		}
		if resp.Candidatos[i].ID == "ITEM2" {
			generico = &resp.Candidatos[i]
		}
	}

	if coreano == nil {
		t.Fatal("produto ITEM1 não encontrado na resposta")
	}
	if generico == nil {
		t.Fatal("produto ITEM2 não encontrado na resposta")
	}

	// Produto da loja monitorada deve ter origem "Coreia"
	if coreano.Origem != "Coreia" {
		t.Errorf("ITEM1 (loja monitorada coreana) deveria ter origem='Coreia', veio %q", coreano.Origem)
	}

	// Produto de loja não monitorada NÃO deve ter origem
	if generico.Origem != "" {
		t.Errorf("ITEM2 (loja não monitorada) deveria ter origem vazia, veio %q", generico.Origem)
	}
}

func TestCandidatosNaoSobrescreveOrigemExistente(t *testing.T) {
	// Se o produto já veio com origem (ex: do coleta service), não sobrescrever
	sp := &spyStore{
		buscas: []store.Busca{
			{
				ID: "loja-123", ShopIDs: []int64{123},
				OrigemPadrao: "China", Ativo: true,
			},
		},
	}

	produtos := []domain.Product{
		{
			ID: "P1", Name: "Produto", Price: 50, Commission: 0.10,
			Sales30d: 100, Rating: 4.5, ShopID: "123",
			Origin: "Japão", // já tem origem (veio do coleta service)
		},
	}

	fonte := &fonteFake{produtos: produtos}
	srv := &Server{
		Eventos: sp,
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return fonte, "fake|fixo"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "GET", "/api/candidatos?sem_filtro=true", nil, nil)
	var resp struct {
		Candidatos []struct {
			Origem string `json:"origem"`
		} `json:"candidatos"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Candidatos) == 0 {
		t.Fatal("esperava candidatos")
	}
	// Deve manter "Japão" (não sobrescrever com "China" da loja)
	if resp.Candidatos[0].Origem != "Japão" {
		t.Errorf("origem deveria ser 'Japão' (preservada), veio %q", resp.Candidatos[0].Origem)
	}
}

// ── Teste de integração: busca → candidato com loja_id correto ───────────────

func TestCandidatosShopID_FallbackParaProductLink(t *testing.T) {
	// Simula cenário real: shopId retorna "0" (busca por keyword, não por loja)
	// mas productLink contém o shopId
	produtos := []domain.Product{
		{
			ID: "ITEM1", Name: "Produto Teste", Category: "beleza",
			Price: 50, Commission: 0.10, Sales30d: 100, Rating: 4.5,
			ShopID:      "0", // API retornou 0 para busca por keyword
			ShopName:    "Loja ABC",
			Link:        "https://shope.ee/shortened123", // offerLink sem -i.
			ProductLink: "https://shopee.com.br/Produto-Teste-i.456789.111222",
		},
	}

	fonte := &fonteFake{produtos: produtos}
	srv := &Server{
		Eventos: &spyStore{},
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return fonte, "fake|fixo"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "GET", "/api/candidatos?sem_filtro=true", nil, nil)
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}

	var resp struct {
		Candidatos []struct {
			ID          string `json:"id"`
			LojaID      string `json:"loja_id"`
			LinkProduto string `json:"link_produto"`
		} `json:"candidatos"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Candidatos) == 0 {
		t.Fatal("esperava candidatos")
	}

	c := resp.Candidatos[0]
	// shopId=0, mas productLink tem -i.456789.111222 → loja_id deve ser "456789"
	if c.LojaID != "456789" {
		t.Errorf("loja_id deveria ser '456789' (extraído do productLink), veio %q", c.LojaID)
	}
	if c.LinkProduto == "" {
		t.Error("link_produto não deveria estar vazio")
	}
}
