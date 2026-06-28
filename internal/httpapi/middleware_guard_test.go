package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoManualAuthInHandlers garante que handlers NÃO fazem verificação de auth manualmente.
// Auth deve ser tratada exclusivamente via middlewares (requireAuth, requireAdmin, requireColetaToken).
// Se este teste falhar, significa que alguém adicionou boilerplate de auth dentro de um handler.
func TestNoManualAuthInHandlers(t *testing.T) {
	// Padrões proibidos em arquivos que NÃO são middleware ou teste
	forbidden := []string{
		`srv.usuarioDoRequest(r)`,
		`srv.autorizadoColeta(r)`,
	}

	// Arquivos que podem conter esses padrões (middlewares, helpers, testes)
	allowed := map[string]bool{
		"middleware_auth.go": true,
		"middleware_log.go":  true,
		"helpers.go":         true, // define usuarioDoRequest (chamado pelo middleware)
	}

	files, _ := filepath.Glob("*.go")
	// Se rodando de dentro do pacote, os arquivos estão no diretório atual
	if len(files) == 0 {
		// Tenta path relativo ao root do projeto
		files, _ = filepath.Glob("../../internal/httpapi/*.go")
	}

	for _, f := range files {
		base := filepath.Base(f)
		if allowed[base] || strings.HasSuffix(base, "_test.go") {
			continue
		}

		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		content := string(data)

		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				t.Errorf("%s contém '%s' — auth deve ser feita via middleware, não no handler", base, pattern)
			}
		}
	}
}

// TestNoManualLoggingInHandlers garante que handlers NÃO fazem logging de request manualmente.
// Logging de request (método, rota, status, duração) é responsabilidade do middleware logRequests.
// Handlers podem logar eventos de negócio (ex: "publicacao", "coleta") mas NÃO o request em si.
func TestNoManualLoggingInHandlers(t *testing.T) {
	// Padrões que indicam logging manual de request (não de negócio)
	forbidden := []string{
		`"requisição"`, // a string usada pelo middleware — se aparecer num handler, é duplicação
	}

	allowed := map[string]bool{
		"middleware_log.go":  true,
		"middleware_auth.go": true,
	}

	files, _ := filepath.Glob("*.go")
	if len(files) == 0 {
		files, _ = filepath.Glob("../../internal/httpapi/*.go")
	}

	for _, f := range files {
		base := filepath.Base(f)
		if allowed[base] || strings.HasSuffix(base, "_test.go") {
			continue
		}

		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		content := string(data)

		for _, pattern := range forbidden {
			if strings.Contains(content, pattern) {
				t.Errorf("%s contém '%s' — logging de request deve ser feito via middleware logRequests", base, pattern)
			}
		}
	}
}
