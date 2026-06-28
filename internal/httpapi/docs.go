package httpapi

import (
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
)

// docsFileServer serve o site de documentação (Starlight).
// Chi Mount("/docs", ...) remove o prefixo; usamos chi.RouteContext para o path real.
func (srv *Server) docsFileServer() http.Handler {
	dir := os.Getenv("DOCS_DIR")
	if dir == "" {
		dir = "docs-site/dist"
	}

	sub := chi.NewRouter()
	sub.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		relPath := chi.URLParam(r, "*")
		if relPath == "" {
			relPath = "index.html"
		}

		// Tentar servir o arquivo diretamente
		fullPath := path.Join(dir, relPath)
		if info, err := os.Stat(fullPath); err == nil {
			if info.IsDir() {
				indexPath := path.Join(fullPath, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					http.ServeFile(w, r, indexPath)
					return
				}
			} else {
				http.ServeFile(w, r, fullPath)
				return
			}
		}

		// Starlight gera /page/index.html para /page/
		dirPath := path.Join(dir, relPath, "index.html")
		if _, err := os.Stat(dirPath); err == nil {
			http.ServeFile(w, r, dirPath)
			return
		}

		// Fallback: 404 do Starlight
		notFound := path.Join(dir, "404.html")
		if _, err := os.Stat(notFound); err == nil {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, notFound)
			return
		}

		http.NotFound(w, r)
	})
	return sub
}

// apiDocs serve a referência de API (Scalar/Swagger UI).
// GET /api/docs
func (srv *Server) apiDocs(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html><head>
<title>Garimpei API</title>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest/dist/style.min.css"/>
</head><body>
<script id="api-reference" data-url="/api/openapi.yaml"></script>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest/dist/browser/standalone.min.js"></script>
</body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// openapiSpec serve o arquivo openapi.yaml.
// GET /api/openapi.yaml
func (srv *Server) openapiSpec(w http.ResponseWriter, r *http.Request) {
	specPath := "api/openapi.yaml"
	if _, err := os.Stat(specPath); err != nil {
		// Fallback para localização antiga
		specPath = "docs/openapi.yaml"
	}
	w.Header().Set("Content-Type", "application/yaml")
	http.ServeFile(w, r, specPath)
}
