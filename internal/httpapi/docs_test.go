package httpapi

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestDocsHandlerServesIndex(t *testing.T) {
	// Cria um diretório temporário simulando docs-site/dist
	dir := t.TempDir()
	os.WriteFile(dir+"/index.html", []byte("<html>docs</html>"), 0o644)
	os.MkdirAll(dir+"/02-arquitetura", 0o755)
	os.WriteFile(dir+"/02-arquitetura/index.html", []byte("<html>arq</html>"), 0o644)

	t.Setenv("DOCS_DIR", dir)

	srv := &Server{}
	srv.inicializar()
	handler := srv.docsHandler()

	tests := []struct {
		path   string
		status int
		body   string
	}{
		{"/docs/", 200, "<html>docs</html>"},
		{"/docs", 200, "<html>docs</html>"},
		{"/docs/02-arquitetura/", 200, "<html>arq</html>"},
		{"/docs/nao-existe", 404, ""},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != tt.status {
			t.Errorf("%s: esperava status %d, obteve %d", tt.path, tt.status, rec.Code)
		}
		if tt.body != "" && rec.Body.String() != tt.body {
			t.Errorf("%s: body inesperado: %q", tt.path, rec.Body.String())
		}
	}
}
