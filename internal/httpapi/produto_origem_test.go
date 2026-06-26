package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizarOrigemProduto(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"Coreia", "Coreia"},
		{"Coréia", "Coreia"},
		{"Korea", "Coreia"},
		{"South Korea", "Coreia"},
		{"Coreia do Sul", "Coreia"},
		{"COREIA", "Coreia"},
		{"kr", "Coreia"},
		{"Japão", "Japão"},
		{"Japan", "Japão"},
		{"jp", "Japão"},
		{"China", "China"},
		{"Mainland China", "China"},
		{"cn", "China"},
		{"Brasil", "Brasil"},
		{"Brazil", "Brasil"},
		{"EUA", "EUA"},
		{"United States", "EUA"},
		{"Taiwan", "Taiwan"},
		{"  coreia  ", "Coreia"},  // trim
		{"França", "França"},
		{"desconhecido", "Desconhecido"}, // capitaliza
	}
	for _, tc := range cases {
		got := NormalizarOrigemProduto(tc.input)
		if got != tc.want {
			t.Errorf("NormalizarOrigemProduto(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestProdutoOrigemEndpoint_MissingParams(t *testing.T) {
	srv := &Server{Eventos: &spyStore{}, Auth: fakeVerifier{}}
	handler := srv.Handler()

	// Sem item_id
	req := httptest.NewRequest("GET", "/api/produto/origem?shop_id=123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, recebeu %d", rec.Code)
	}

	// Sem shop_id
	req = httptest.NewRequest("GET", "/api/produto/origem?item_id=456", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, recebeu %d", rec.Code)
	}
}

func TestProdutoOrigemEndpoint_CacheHit(t *testing.T) {
	// Pre-popula o cache
	salvarOrigemNoCache("999:888", origemCacheEntry{Origem: "Coreia", Marca: "SKIN1004"})

	srv := &Server{Eventos: &spyStore{}, Auth: fakeVerifier{}}
	handler := srv.Handler()

	req := httptest.NewRequest("GET", "/api/produto/origem?item_id=888&shop_id=999", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, recebeu %d", rec.Code)
	}

	var resp origemResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

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

func TestProdutoOrigemBatch_EmptyList(t *testing.T) {
	srv := &Server{Eventos: &spyStore{}, Auth: fakeVerifier{}}
	handler := srv.Handler()

	body := `{"itens": []}`
	req := httptest.NewRequest("POST", "/api/produto/origem/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperado 400 para lista vazia, recebeu %d", rec.Code)
	}
}

func TestProdutoOrigemBatch_WithCache(t *testing.T) {
	// Pre-popula o cache
	salvarOrigemNoCache("100:200", origemCacheEntry{Origem: "Japão", Marca: "Shiseido"})

	srv := &Server{Eventos: &spyStore{}, Auth: fakeVerifier{}}
	handler := srv.Handler()

	body := `{"itens": [{"item_id": "200", "shop_id": "100"}]}`
	req := httptest.NewRequest("POST", "/api/produto/origem/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, recebeu %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Resultados []origemResponse `json:"resultados"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Resultados) != 1 {
		t.Fatalf("esperado 1 resultado, recebeu %d", len(resp.Resultados))
	}
	if resp.Resultados[0].Origem != "Japão" {
		t.Errorf("origem = %q, want 'Japão'", resp.Resultados[0].Origem)
	}
	if resp.Resultados[0].Fonte != "cache" {
		t.Errorf("fonte = %q, want 'cache'", resp.Resultados[0].Fonte)
	}
}
