package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDocsFileServer(t *testing.T) {
	// Cria um diretório temporário simulando docs-site/dist
	dir := t.TempDir()
	os.WriteFile(dir+"/index.html", []byte("<html>docs</html>"), 0o644)
	os.MkdirAll(dir+"/01-visao-e-negocio", 0o755)
	os.WriteFile(dir+"/01-visao-e-negocio/index.html", []byte("<html>visao</html>"), 0o644)
	os.MkdirAll(dir+"/_astro", 0o755)
	os.WriteFile(dir+"/_astro/page.js", []byte("//js"), 0o644)
	os.WriteFile(dir+"/404.html", []byte("<html>404</html>"), 0o644)

	t.Setenv("DOCS_DIR", dir)

	srv := &Server{}
	srv.inicializar()

	// Simula o setup real: StripPrefix("/docs") + docsFileServer
	handler := http.StripPrefix("/docs", srv.docsFileServer())

	tests := []struct {
		path         string
		status       int
		bodyContains string
	}{
		{"/docs/", 200, "<html>docs</html>"},
		{"/docs/01-visao-e-negocio/", 200, "<html>visao</html>"},
		{"/docs/01-visao-e-negocio", 200, "<html>visao</html>"},
		{"/docs/_astro/page.js", 200, "//js"},
		{"/docs/nao-existe", 404, "<html>404</html>"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != tt.status {
			t.Errorf("%s: esperava status %d, obteve %d", tt.path, tt.status, rec.Code)
		}
		if tt.bodyContains != "" && rec.Body.String() != tt.bodyContains {
			t.Errorf("%s: body esperado %q, obteve %q", tt.path, tt.bodyContains, rec.Body.String())
		}
	}
}
