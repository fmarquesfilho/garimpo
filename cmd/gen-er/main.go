// cmd/gen-er gera docs/gerado/ENTIDADES.md (Mermaid ER) a partir de deploy/bigquery_schema.sql.
//
// Uso: go run ./cmd/gen-er > docs/gerado/ENTIDADES.md
// Ou via Makefile: make docs-er
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type column struct {
	Name string
	Type string
}

type table struct {
	Name    string
	Columns []column
}

var (
	// Formato: `PROJECT.dataset.tabela` (tudo dentro de um par de backticks)
	reCreateTable = regexp.MustCompile("(?i)CREATE\\s+TABLE\\s+IF\\s+NOT\\s+EXISTS\\s+`[^`]*\\.([^`]+)`")
	reColumn      = regexp.MustCompile(`^\s+(\w+)\s+(STRING|INT64|FLOAT64|BOOL|TIMESTAMP|DATE|BYTES)`)
)

func main() {
	schemaPath := "deploy/bigquery_schema.sql"
	if len(os.Args) > 1 {
		schemaPath = os.Args[1]
	}

	f, err := os.Open(schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir %s: %v\n", schemaPath, err)
		os.Exit(1)
	}
	defer f.Close()

	var tables []table
	var current *table

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := reCreateTable.FindStringSubmatch(line); m != nil {
			tables = append(tables, table{Name: strings.ToUpper(m[1])})
			current = &tables[len(tables)-1]
			continue
		}

		if current != nil {
			if m := reColumn.FindStringSubmatch(line); m != nil {
				current.Columns = append(current.Columns, column{Name: m[1], Type: m[2]})
			}
			if strings.Contains(line, ")") && !strings.Contains(line, "(") {
				current = nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro lendo schema: %v\n", err)
		os.Exit(1)
	}

	// Emitir markdown
	fmt.Println("---")
	fmt.Println("title: Modelo de dados (ER)")
	fmt.Println("description: Diagrama entidade-relacionamento gerado do schema BigQuery.")
	fmt.Println("---")
	fmt.Println()
	fmt.Println(":::caution[Arquivo gerado]")
	fmt.Println("Não edite manualmente. Rode `make docs-er` para regenerar.")
	fmt.Println(":::")
	fmt.Println()
	fmt.Println("```mermaid")
	fmt.Println("erDiagram")

	for _, t := range tables {
		fmt.Printf("    %s {\n", t.Name)
		for _, c := range t.Columns {
			fmt.Printf("        %s %s\n", c.Type, c.Name)
		}
		fmt.Println("    }")
		fmt.Println()
	}

	// Relacionamentos conhecidos
	fmt.Println("    BUSCAS ||--o{ SNAPSHOTS : \"gera coletas\"")
	fmt.Println("    SNAPSHOTS ||--o{ EVENTOS : \"produto selecionado\"")
	fmt.Println("    EVENTOS ||--o{ CONVERSOES : \"gera conversão\"")
	fmt.Println("```")
	fmt.Println()
	fmt.Println("## Particionamento")
	fmt.Println()
	fmt.Println("| Tabela | Partição |")
	fmt.Println("|---|---|")
	fmt.Println("| `eventos` | `DATE(em)` |")
	fmt.Println("| `snapshots` | `DATE(coletado_em)` |")
	fmt.Println("| `buscas` | `DATE(salvo_em)` |")
	fmt.Println("| `conversoes` | `DATE(compra_em)` |")
}
