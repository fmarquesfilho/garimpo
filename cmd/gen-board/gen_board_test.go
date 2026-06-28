package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGenBoardMatchesCommitted verifica que o board gerado bate com o commitado.
// Se falhar: rode `make docs-board` e commite.
func TestGenBoardMatchesCommitted(t *testing.T) {
	root := "../../"
	boardPath := root + "docs/gerado/BOARD.md"

	committed, err := os.ReadFile(boardPath)
	if err != nil {
		t.Skip("docs/gerado/BOARD.md não existe, pulando teste de drift")
	}

	cmd := exec.Command("go", "run", "./cmd/gen-board")
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		t.Fatalf("gen-board falhou: %v", err)
	}

	regenerated, err := os.ReadFile(boardPath)
	if err != nil {
		t.Fatalf("não conseguiu ler BOARD.md após regenerar: %v", err)
	}

	// Comparar ignorando a linha "Gerado em" (varia com a data)
	committedLines := stripDateLines(string(committed))
	regeneratedLines := stripDateLines(string(regenerated))

	if committedLines != regeneratedLines {
		t.Errorf("docs/gerado/BOARD.md está desatualizado (diff ignorando datas).\n" +
			"Rode: make docs-board && git add docs/gerado/")
	}

	// Restaurar o arquivo original para não sujar o working tree
	os.WriteFile(boardPath, committed, 0o644)
}

// TestGenBoardValidatesTasks verifica que todas as tarefas têm campos obrigatórios.
func TestGenBoardValidatesTasks(t *testing.T) {
	tasks := loadTasks("../../backlog/tasks")
	if len(tasks) == 0 {
		t.Skip("nenhuma tarefa encontrada")
	}

	validStatus := map[string]bool{
		"backlog": true, "next": true, "doing": true,
		"review": true, "done": true, "blocked": true,
	}

	for _, task := range tasks {
		if task.ID == "" {
			t.Errorf("tarefa sem ID")
		}
		if task.Titulo == "" {
			t.Errorf("%s: título vazio", task.ID)
		}
		if !validStatus[task.Status] {
			t.Errorf("%s: status inválido '%s'", task.ID, task.Status)
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

// stripDateLines remove linhas que contêm "Gerado em" para comparação estável.
func stripDateLines(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, "Gerado em") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
