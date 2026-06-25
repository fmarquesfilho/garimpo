package httpapi

import (
	"net/http"
	"os"
)

// apiDocs serve uma página HTML com Swagger UI apontando para o openapi.yaml.
func (srv *Server) apiDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerHTML))
}

// openapiSpec serve o arquivo openapi.yaml.
func (srv *Server) openapiSpec(w http.ResponseWriter, r *http.Request) {
	// Tenta ler de docs/openapi.yaml (relativo ao binário)
	paths := []string{"docs/openapi.yaml", "../docs/openapi.yaml", "../../docs/openapi.yaml"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write(data)
			return
		}
	}
	writeErr(w, http.StatusNotFound, "openapi.yaml não encontrado")
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Garimpei API — Documentação</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
  <style>
    body { margin: 0; background: #fafafa; }
    .topbar { display: none !important; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '/api/openapi.yaml',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: 'BaseLayout'
    });
  </script>
</body>
</html>`
