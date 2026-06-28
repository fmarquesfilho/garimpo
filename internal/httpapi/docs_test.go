package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestDocsFileServer(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/index.html", []byte("<html>docs</html>"), 0o600)
	os.MkdirAll(dir+"/01-visao-e-negocio", 0o755)
	os.WriteFile(dir+"/01-visao-e-negocio/index.html", []byte("<html>visao</html>"), 0o600)
	os.MkdirAll(dir+"/_astro", 0o755)
	os.WriteFile(dir+"/_astro/page.js", []byte("//js"), 0o600)
	os.WriteFile(dir+"/404.html", []byte("<html>404</html>"), 0o600)

	t.Setenv("DOCS_DIR", dir)

	srv := &Server{}
	srv.inicializar()

	// Monta com Chi exatamente como em produção
	r := chi.NewRouter()
	r.Mount("/docs", srv.docsFileServer())

	tests := []struct {
		path         string
		status       int
		bodyContains string
	}{
		{"/docs/", http.StatusOK, "<html>docs</html>"},
		{"/docs", http.StatusOK, "<html>docs</html>"},
		{"/docs/01-visao-e-negocio/", http.StatusOK, "<html>visao</html>"},
		{"/docs/01-visao-e-negocio", http.StatusOK, "<html>visao</html>"},
		{"/docs/_astro/page.js", http.StatusOK, "//js"},
		{"/docs/nao-existe", http.StatusNotFound, "<html>404</html>"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != tt.status {
			t.Errorf("%s: esperava status %d, obteve %d", tt.path, tt.status, rec.Code)
		}
		if tt.bodyContains != "" && rec.Body.String() != tt.bodyContains {
			t.Errorf("%s: body esperado %q, obteve %q", tt.path, tt.bodyContains, rec.Body.String())
		}
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	srv := &Server{}
	srv.inicializar()
	handler := srv.Handler()

	// POST em rota que só aceita GET deve retornar 405
	req := httptest.NewRequest("POST", "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("/api/health POST: esperava 405, obteve %d", rec.Code)
	}
}

func TestRouterCORS(t *testing.T) {
	srv := &Server{}
	srv.inicializar()
	handler := srv.Handler()

	// OPTIONS em qualquer rota deve retornar headers CORS
	req := httptest.NewRequest("OPTIONS", "/api/candidatos", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS: Access-Control-Allow-Origin ausente")
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS: Access-Control-Allow-Methods ausente")
	}
}
