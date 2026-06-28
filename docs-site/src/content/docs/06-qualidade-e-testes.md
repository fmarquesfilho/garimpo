---
title: "Qualidade e testes"
---


## Pipeline de CI

O workflow `deploy-gcp.yml` roda em push para `main`:

1. **go test** — ~200 testes unitários e de integração
2. **golangci-lint** — estilo e bugs Go
3. **arch-go** — restrições de dependência entre pacotes
4. **vitest** — 109 testes frontend (<2s)
5. **eslint + stylelint** — qualidade do frontend
6. **check-file-size** — bloqueia arquivos > 400 linhas
7. **build + deploy** — Docker → Artifact Registry → Cloud Run

Pushes que só tocam `docs/**` são ignorados (`paths-ignore`).

## Estratégia de testes

### BDD (Behaviour-Driven Development)

Os testes seguem cenários Given/When/Then:

- **Given** — estado inicial (mock de dados, configuração)
- **When** — ação do usuário ou trigger do sistema
- **Then** — resultado esperado (response, efeito colateral)

### Cobertura por camada

| Camada | Ferramenta | Foco |
|---|---|---|
| Domain/Engine | `go test` | Ranking, score, filtros, exploração |
| HTTP API | `go test` + httptest | Handlers, validação, auth |
| Coleta | `go test` | Integração Shopee mock, rotação, throttling |
| Frontend | Vitest | Componentes, stores, utils |
| E2E | Playwright | Fluxos críticos do usuário |

### Padrão de teste Go

```go
func TestAlgo(t *testing.T) {
    // Given
    dado := prepararFixture()

    // When
    resultado := funcaoSobTeste(dado)

    // Then
    if resultado != esperado {
        t.Errorf("esperava %v, obteve %v", esperado, resultado)
    }
}
```

## Análise estática

### golangci-lint

Configurado em `.golangci.yml`. Linters ativos:
`errcheck`, `govet`, `staticcheck`, `unused`, `gofmt`, `goimports`

### arch-go

Configurado em `arch-go.yml`. Garante:
- `internal/domain` não importa nada de infra
- `internal/httpapi` não acessa BigQuery diretamente
- `cmd/` só importa `internal/`

### Frontend

- **eslint** — qualidade JS/Svelte (`eslint.config.js`)
- **stylelint** — qualidade CSS (`.stylelintrc.json`)
- **knip** — detecta dead code (`knip.json`)

## docs-check (CI de documentação)

```bash
make docs-check
```

Valida que docs geradas (`docs/gerado/`) estão atualizadas.
Falha se `git diff --exit-code docs/gerado` mostrar diferença.

## Métricas de qualidade

| Métrica | Alvo |
|---|---|
| Testes Go | ~200, `go test ./...` |
| Testes frontend | 109, `vitest --run` (<2s) |
| Arquivos > 400 linhas | 0 (CI bloqueia) |
| Cobertura | Não há meta numérica; foco em cenários críticos |
