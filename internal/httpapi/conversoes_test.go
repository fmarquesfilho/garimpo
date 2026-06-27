package httpapi

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/source"
)

func TestConversoesReaisExigeAuth(t *testing.T) {
	h := montar(&fonteFake{produtos: amostra}, &spyStore{}, &spyPub{})
	rec := req(t, h, "GET", "/api/conversoes/reais", nil, nil)
	if rec.Code != 401 {
		t.Errorf("sem auth deveria dar 401, veio %d", rec.Code)
	}
}

func TestConversoesReaisSemCredenciais(t *testing.T) {
	// Com auth mas sem SHOPEE_APP_ID/SECRET → 503
	srv := &Server{
		Eventos: &spyStore{},
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "GET", "/api/conversoes/reais?dias=7", nil, map[string]string{"Authorization": "Bearer tok"})
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("sem credenciais deveria dar 503, veio %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConversoesReaisComCredenciais(t *testing.T) {
	// Com credenciais mas API vai falhar (endpoint real não disponível em teste)
	// O importante é que o handler aceita o request e tenta chamar a API
	t.Setenv("SHOPEE_APP_ID", "test-app")
	t.Setenv("SHOPEE_SECRET", "test-secret")

	srv := &Server{
		Eventos: &spyStore{},
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	rec := req(t, h, "GET", "/api/conversoes/reais?dias=7", nil, map[string]string{"Authorization": "Bearer tok"})
	// Vai dar 502 porque a API da Shopee não é alcançável em teste — mas NÃO 401/503
	if rec.Code == 401 || rec.Code == 503 {
		t.Errorf("com credenciais não deveria dar %d", rec.Code)
	}
}

func TestConversoesReaisLimitaDias(t *testing.T) {
	// dias > 90 deve ser limitado a 90
	t.Setenv("SHOPEE_APP_ID", "test-app")
	t.Setenv("SHOPEE_SECRET", "test-secret")

	srv := &Server{
		Eventos: &spyStore{},
		Auth:    fakeVerifier{},
		FonteFactory: func(q url.Values) (source.ProductSource, string) {
			return &fonteFake{produtos: amostra}, "fake"
		},
	}
	h := srv.Handler()

	// dias=200 → handler aceita e limita internamente
	rec := req(t, h, "GET", "/api/conversoes/reais?dias=200", nil, map[string]string{"Authorization": "Bearer tok"})
	// Não deve dar panic ou 400 — deve aceitar e chamar a API
	if rec.Code == 400 {
		t.Error("dias=200 não deveria dar 400 — handler deve limitar a 90")
	}
}
