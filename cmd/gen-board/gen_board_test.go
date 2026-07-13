package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestGenBoardProducesValidOutput verifica que gen-board executa sem erro
// e produz os arquivos markdown esperados em docs/gerado/.
func TestGenBoardProducesValidOutput(t *testing.T) {
	root := "../../"

	cmd := exec.Command("go", "run", "./cmd/gen-board")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gen-board falhou: %v\nOutput: %s", err, string(out))
	}

	// Verifica que os arquivos foram gerados
	expectedFiles := []string{"docs/gerado/BOARD.md", "docs/gerado/ROADMAP.md"}
	for _, f := range expectedFiles {
		content, err := os.ReadFile(root + f)
		if err != nil {
			t.Errorf("arquivo esperado nao encontrado: %s", f)
			continue
		}
		if len(content) < 50 {
			t.Errorf("%s muito curto (%d bytes)", f, len(content))
		}
	}
}

// TestGenBoardValidatesTasks verifica que todas as tarefas tem campos obrigatorios.
func TestGenBoardValidatesTasks(t *testing.T) {
	tasks := loadTasks("../../backlog/tasks")
	if len(tasks) == 0 {
		t.Skip("nenhuma tarefa encontrada")
	}

	validStatus := map[string]bool{
		"backlog": true, "next": true, "doing": true,
		"review": true, "done": true, "blocked": true, "ready": true,
	}

	for _, task := range tasks {
		if task.ID == "" {
			t.Errorf("tarefa sem ID")
		}
		if task.Titulo == "" {
			t.Errorf("%s: titulo vazio", task.ID)
		}
		if !validStatus[task.Status] {
			t.Errorf("%s: status invalido '%s'", task.ID, task.Status)
		}
		if task.Prioridade == "" {
			t.Errorf("%s: prioridade vazia", task.ID)
		}
	}
}

// TestGenBoardNoBrokenDependencies verifica que depende_de referencia IDs existentes.
func TestGenBoardNoBrokenDependencies(t *testing.T) {
	tasks := loadTasks("../../backlog/tasks")
	ids := map[string]bool{}
	for _, task := range tasks {
		ids[task.ID] = true
	}

	for _, task := range tasks {
		for _, dep := range task.DependeDe {
			if !ids[dep] {
				t.Errorf("%s: depende_de referencia ID inexistente '%s'", task.ID, dep)
			}
		}
	}
}
