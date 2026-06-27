# Análise Estática — Garimpei

Configurada em 2026-06-27. Todas as ferramentas rodam localmente e no CI a cada push.

---

## Ferramentas

| Ferramenta | Propósito | Bloq. CI | Como rodar local |
|-----------|-----------|:---:|-----------|
| **golangci-lint** | Linting Go (errcheck, gosec, staticcheck, etc.) | ✅ | `golangci-lint run ./...` |
| **govulncheck** | Vulnerabilidades alcançáveis em deps Go | ✅ | `govulncheck ./...` |
| **arch-go** | Regras de dependência entre packages | ✅ | `arch-go` |
| **go test** | Testes unitários e integração | ✅ | `go test ./...` |
| **ESLint** | Linting JS/Svelte | ✅ | `cd web && npm run lint:js` |
| **Stylelint** | CSS/design tokens | ✅ | `cd web && npm run lint:css` |
| **Vitest** | Testes unitários frontend | ✅ | `cd web && npx vitest run` |
| **knip** | Código/deps mortas frontend | — | `cd web && npx knip` |

### Linters adicionais (rodar manualmente, não bloqueiam CI)

Para sugestões de performance — útil antes de otimizar hot paths:
```bash
golangci-lint run --enable=gocritic,prealloc ./...
```

---

## Configuração

### `.golangci.yml` — Linters habilitados no CI

| Linter | O que detecta |
|--------|--------------|
| errcheck | Erros de retorno ignorados (causa de bugs silenciosos) |
| staticcheck | Padrões incorretos, código morto, simplificações |
| govet | Erros detectados pelo `go vet` oficial |
| ineffassign | Atribuições a variáveis nunca lidas |
| typecheck | Erros de tipo |
| unused | Funções/variáveis/tipos não usados |
| gosimple | Simplificações possíveis |
| revive | Regras de estilo configuráveis |
| bodyclose | `http.Response.Body` não fechado (leak) |
| gosec | Problemas de segurança (secrets, crypto fraco) |
| gofmt | Formatação padrão |
| goimports | Imports organizados |

**Exclusões intencionais:**
- `gocritic` — sugestões de performance (hugeParam, rangeValCopy). Não é bug, é otimização incremental.
- `prealloc` — slices sem pre-alocação. Idem.
- `misspell` — falsos positivos com termos em português (ex: "rela" flagged como "real").
- Arquivos `_test.go` — excluídos de errcheck e gosec (testes podem ignorar erros intencionalmente).
- `G404` (gosec) — math/rand usado intencionalmente para exploração (shuffle, não cripto).

### `arch-go.yml` — Regras de Arquitetura

Protege a separação de camadas. Se alguém importar `httpapi` dentro de `domain`, o CI falha.

| Package | Não pode importar |
|---------|------------------|
| `domain` | Nada externo (zero deps internas) |
| `source` | httpapi, store, publish |
| `engine` | httpapi, store |
| `strategy` | httpapi, store, source |
| `store` | httpapi, source, engine |
| `tenant` | httpapi, source, store |

```
domain ← engine ← httpapi → store
  ↑        ↑                   ↑
strategy  source              tenant
```

---

## Resultado Atual (após correções)

### Backend

```
golangci-lint:  0 issues (CI passa) ✅
govulncheck:    0 vulnerabilidades ✅
arch-go:        100% compliance (6/6 regras) ✅
go test:        todos os packages passam ✅
```

### Frontend

```
ESLint:    0 errors ✅ (16 warnings — no-unused-vars em código existente)
Stylelint: 0 errors ✅
knip:      0 issues ✅
Vitest:    10/10 testes passam ✅
```

---

## O que foi corrigido na sessão de 27/06

### Segurança (P0) ✅
- [x] Go 1.26.0 → 1.26.4 (12 CVEs na stdlib resolvidas)
- [x] `http.Server` com ReadTimeout/WriteTimeout/IdleTimeout (antes: sem timeout)

### Bugs (P1) ✅
- [x] 4 `errcheck` — `json.Decode` e `w.Write` com erros ignorados
- [x] 2 funções mortas removidas (`parseShopInput`, `resolveShopSlug`)

### Estilo (P3) ✅
- [x] `gofmt` aplicado em todo o projeto (27 issues de formatação)

### Pendente (melhorias incrementais)
- [ ] `gocritic/hugeParam` (84) — converter structs grandes para ponteiros nos hot paths
- [ ] `prealloc` (3) — pre-alocar slices quando o tamanho é conhecido

---

## Integração com CI

Adicionado ao job `test-go` em `.github/workflows/deploy-gcp.yml`:

```yaml
- name: golangci-lint
  run: |
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    golangci-lint run --timeout=3m --issues-exit-code=1 ./...

- name: govulncheck
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...

- name: arch-go
  run: |
    go install github.com/arch-go/arch-go@latest
    arch-go
```

**Comportamento:** se qualquer linter de bug/segurança falhar, ou se uma vulnerabilidade alcançável for encontrada, ou se uma regra de arquitetura for violada — o deploy é bloqueado.

---

## Como testar localmente antes do push

```bash
# Checagem completa (exatamente o que o CI roda):
go test ./...
golangci-lint run --timeout=3m --issues-exit-code=1 ./...
govulncheck ./...
arch-go

# Frontend:
cd web
npm run lint:js
npm run lint:css
npx vitest run
npm run build
```

Se tudo passar localmente, o CI também vai passar.
