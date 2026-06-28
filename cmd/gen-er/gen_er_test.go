package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGenERMatchesCommitted verifica que a saída de gen-er bate com o arquivo commitado.
// Se este teste falhar, significa que o schema mudou sem regenerar o ENTIDADES.md.
// Resolução: rode `make docs-er` e commite o resultado.
func TestGenERMatchesCommitted(t *testing.T) {
	// Navega para a raiz do projeto (2 níveis acima de cmd/gen-er)
	root := "../../"

	committed, err := os.ReadFile(root + "docs/gerado/ENTIDADES.md")
	if err != nil {
		t.Skip("docs/gerado/ENTIDADES.md não existe, pulando teste de drift")
	}

	cmd := exec.Command("go", "run", "./cmd/gen-er")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("gen-er falhou: %v\n%s", err, output)
	}

	if strings.TrimSpace(string(output)) != strings.TrimSpace(string(committed)) {
		t.Errorf("docs/gerado/ENTIDADES.md está desatualizado em relação ao schema.\n" +
			"Rode: make docs-er && git add docs/gerado/ENTIDADES.md")
	}
}
