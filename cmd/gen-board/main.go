// cmd/gen-board gera docs/gerado/BOARD.md e docs/gerado/ROADMAP.md a partir de backlog/tasks/*.yaml.
//
// Uso: go run ./cmd/gen-board
// Ou via Makefile: make docs-board
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Task struct {
	ID         string   `yaml:"id"`
	Titulo     string   `yaml:"titulo"`
	Epic       string   `yaml:"epic"`
	Status     string   `yaml:"status"`
	Prioridade string   `yaml:"prioridade"`
	Estimativa string   `yaml:"estimativa"`
	Sprint     string   `yaml:"sprint"`
	Valor      string   `yaml:"valor"`
	Criterios  []string `yaml:"criterios"`
	DependeDe  []string `yaml:"depende_de"`
	Tags       []string `yaml:"tags"`
	CriadaEm   string   `yaml:"criada_em"`
}

var statusOrder = map[string]int{
	"blocked": 0, "backlog": 1, "next": 2, "doing": 3, "review": 4, "done": 5,
}

var statusEmoji = map[string]string{
	"backlog": "📋", "next": "⏭️", "doing": "🔨", "review": "👀", "done": "✅", "blocked": "🚫",
}

func main() {
	tasksDir := "backlog/tasks"
	sprintFile := "backlog/sprint.txt"
	outBoard := "docs/gerado/BOARD.md"
	outRoadmap := "docs/gerado/ROADMAP.md"

	sprint := readSprint(sprintFile)
	tasks := loadTasks(tasksDir)

	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "Nenhuma tarefa encontrada em", tasksDir)
		os.Exit(1)
	}

	// Validar status
	validStatus := map[string]bool{"backlog": true, "next": true, "doing": true, "review": true, "done": true, "blocked": true}
	for _, t := range tasks {
		if !validStatus[t.Status] {
			fmt.Fprintf(os.Stderr, "❌ %s: status inválido '%s'\n", t.ID, t.Status)
			os.Exit(1)
		}
	}

	// Validar depende_de aponta para IDs existentes
	ids := map[string]bool{}
	for _, t := range tasks {
		ids[t.ID] = true
	}
	for _, t := range tasks {
		for _, dep := range t.DependeDe {
			if !ids[dep] {
				fmt.Fprintf(os.Stderr, "❌ %s: depende_de referencia ID inexistente '%s'\n", t.ID, dep)
				os.Exit(1)
			}
		}
	}

	generateBoard(outBoard, tasks, sprint)
	generateRoadmap(outRoadmap, tasks)

	fmt.Printf("✅ Board e roadmap gerados (%d tarefas, sprint %s)\n", len(tasks), sprint)
}

func readSprint(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "current"
	}
	return strings.TrimSpace(string(data))
}

func loadTasks(dir string) []Task {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro listando tasks:", err)
		os.Exit(1)
	}
	var tasks []Task
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro lendo %s: %v\n", f, err)
			os.Exit(1)
		}
		var t Task
		if err := yaml.Unmarshal(data, &t); err != nil {
			fmt.Fprintf(os.Stderr, "Erro parsing %s: %v\n", f, err)
			os.Exit(1)
		}
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if statusOrder[tasks[i].Status] != statusOrder[tasks[j].Status] {
			return statusOrder[tasks[i].Status] < statusOrder[tasks[j].Status]
		}
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}

func generateBoard(path string, tasks []Task, sprint string) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Quadro — Sprint %s\n\n", sprint))
	buf.WriteString("> Gerado automaticamente. Não edite — rode `make docs-board`.\n\n")

	// Agrupar por status
	grouped := map[string][]Task{}
	for _, t := range tasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}

	// Colunas do board
	columns := []string{"backlog", "next", "doing", "review", "done"}

	// Tabela
	buf.WriteString("|")
	for _, col := range columns {
		buf.WriteString(fmt.Sprintf(" %s %s |", statusEmoji[col], capitalize(col)))
	}
	buf.WriteString("\n|")
	for range columns {
		buf.WriteString("---|")
	}
	buf.WriteString("\n")

	// Encontrar o máximo de linhas
	maxRows := 0
	for _, col := range columns {
		if len(grouped[col]) > maxRows {
			maxRows = len(grouped[col])
		}
	}

	for row := 0; row < maxRows; row++ {
		buf.WriteString("|")
		for _, col := range columns {
			if row < len(grouped[col]) {
				t := grouped[col][row]
				est := ""
				if t.Estimativa != "" {
					est = " (" + t.Estimativa + ")"
				}
				buf.WriteString(fmt.Sprintf(" %s %s%s |", t.ID, t.Titulo, est))
			} else {
				buf.WriteString(" |")
			}
		}
		buf.WriteString("\n")
	}

	// Bloqueadas
	if blocked := grouped["blocked"]; len(blocked) > 0 {
		buf.WriteString("\n**🚫 Bloqueadas:**\n")
		for _, t := range blocked {
			buf.WriteString(fmt.Sprintf("- %s %s\n", t.ID, t.Titulo))
		}
	}

	// Métricas
	buf.WriteString("\n---\n\n")
	buf.WriteString("**Métricas:**\n")
	for _, col := range append([]string{"blocked"}, columns...) {
		if col == "blocked" && len(grouped[col]) == 0 {
			continue
		}
		buf.WriteString(fmt.Sprintf("- %s %s: %d\n", statusEmoji[col], col, len(grouped[col])))
	}

	doingCount := len(grouped["doing"])
	wipLimit := 2
	if doingCount > wipLimit {
		buf.WriteString(fmt.Sprintf("\n⚠️ **WIP excedido:** %d tarefas em Doing (limite: %d)\n", doingCount, wipLimit))
	}

	writeFile(path, buf.String())
}

func generateRoadmap(path string, tasks []Task) {
	var buf strings.Builder
	buf.WriteString("# Roadmap\n\n")
	buf.WriteString("> Gerado automaticamente. Não edite — rode `make docs-board`.\n\n")

	// Now = doing + review
	// Next = next
	// Later = backlog
	now := filterStatus(tasks, "doing", "review")
	next := filterStatus(tasks, "next")
	later := filterStatus(tasks, "backlog")

	buf.WriteString("## 🔵 Now (em andamento)\n\n")
	if len(now) == 0 {
		buf.WriteString("_Nenhuma tarefa em andamento._\n")
	}
	for _, t := range now {
		buf.WriteString(formatRoadmapLine(t))
	}

	buf.WriteString("\n## 🟡 Next (próximo sprint)\n\n")
	if len(next) == 0 {
		buf.WriteString("_Nenhuma tarefa planejada._\n")
	}
	for _, t := range next {
		buf.WriteString(formatRoadmapLine(t))
	}

	buf.WriteString("\n## ⚪ Later (radar)\n\n")
	if len(later) == 0 {
		buf.WriteString("_Nenhuma tarefa no radar._\n")
	}
	for _, t := range later {
		buf.WriteString(formatRoadmapLine(t))
	}

	// Done recente
	done := filterStatus(tasks, "done")
	if len(done) > 0 {
		buf.WriteString("\n## ✅ Concluídas\n\n")
		for _, t := range done {
			buf.WriteString(formatRoadmapLine(t))
		}
	}

	writeFile(path, buf.String())
}

func filterStatus(tasks []Task, statuses ...string) []Task {
	set := map[string]bool{}
	for _, s := range statuses {
		set[s] = true
	}
	var out []Task
	for _, t := range tasks {
		if set[t.Status] {
			out = append(out, t)
		}
	}
	return out
}

func formatRoadmapLine(t Task) string {
	parts := []string{fmt.Sprintf("- **%s** %s", t.ID, t.Titulo)}
	if t.Epic != "" {
		parts = append(parts, "· "+t.Epic)
	}
	if t.Estimativa != "" {
		parts = append(parts, "· "+t.Estimativa)
	}
	return strings.Join(parts, " ") + "\n"
}

func writeFile(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Erro criando diretório para %s: %v\n", path, err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Erro escrevendo %s: %v\n", path, err)
		os.Exit(1)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
