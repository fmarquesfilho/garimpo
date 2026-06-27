# Qualidade de Código — Garimpei

Ferramentas, regras e práticas configuradas para manter o código seguro, legível e sustentável à medida que o projeto cresce.

Atualizado em: 2026-06-27

---

## Pipeline de Qualidade (CI)

Todo push na main roda automaticamente:

```
go test → golangci-lint → govulncheck → arch-go → ESLint → Stylelint → Vitest → Playwright
```

Se qualquer step falhar, o deploy é bloqueado.

### Resumo dos gates

| Gate | O que bloqueia | Tempo |
|------|---------------|-------|
| `go test ./...` | Testes quebrados | ~6s |
| `golangci-lint` | Bugs, segurança, funções longas, complexidade alta | ~10s |
| `govulncheck` | CVEs alcançáveis nas dependências | ~5s |
| `arch-go` | Violação de regras de dependência entre packages | ~3s |
| `npm run lint:js` | Erros ESLint no frontend | ~3s |
| `npm run lint:css` | Violações de design tokens | ~2s |
| `npx vitest run` | Testes unitários de componentes | ~2s |
| Playwright E2E | Fluxos de ponta a ponta | ~30s |

---

## Linters Go (`.golangci.yml`)

### Linters habilitados (bloqueiam CI)

| Linter | Categoria | O que detecta |
|--------|-----------|--------------|
| errcheck | Bug | Erros de retorno ignorados |
| staticcheck | Bug | Padrões incorretos, código morto |
| govet | Bug | Erros do `go vet` oficial |
| ineffassign | Bug | Atribuições a variáveis nunca lidas |
| typecheck | Bug | Erros de tipo |
| unused | Bug | Código morto (funções/tipos não usados) |
| gosimple | Qualidade | Simplificações possíveis |
| revive | Qualidade | Regras de estilo configuráveis |
| bodyclose | Bug | `http.Response.Body` não fechado (leak de conexão) |
| gosec | Segurança | Secrets hardcoded, crypto fraco, timeouts ausentes |
| gofmt | Estilo | Formatação padrão Go |
| goimports | Estilo | Imports organizados |
| funlen | Manutenção | Funções > 100 linhas ou > 60 statements |
| cyclop | Manutenção | Complexidade ciclomática > 20 |

### Exclusões intencionais

| Item | Motivo |
|------|--------|
| `gocritic` | Sugestões de performance (hugeParam). Útil localmente, não bloqueia CI. |
| `prealloc` | Slices sem pre-alocação. Otimização incremental. |
| `misspell` | Falsos positivos com termos em português. |
| G404 (gosec) | `math/rand` usado para exploração (shuffle, não cripto). |
| `_test.go` | Excluídos de errcheck, gosec, funlen, cyclop. |
| `cmd/` | Excluído de funlen (main é naturalmente longo). |

### Como rodar localmente

```bash
# Exatamente o que o CI roda:
golangci-lint run --timeout=3m --issues-exit-code=1 ./...

# Para ver também sugestões de performance (não bloqueia CI):
golangci-lint run --enable=gocritic,prealloc ./...
```

---

## Segurança de Dependências (`govulncheck`)

Analisa vulnerabilidades **alcançáveis** (não só presentes) nas dependências Go.

```bash
govulncheck ./...
```

**Estado atual:** 0 vulnerabilidades (Go atualizado para 1.26.4).

**Ação se encontrar CVE:** atualizar a dependência ou Go. O CI bloqueia até resolver.

---

## Regras de Arquitetura (`arch-go.yml`)

Protege a separação de camadas. Se alguém importar `httpapi` dentro de `domain`, o CI falha.

```
domain ← engine ← httpapi → store
  ↑        ↑                   ↑
strategy  source              tenant
```

### Regras configuradas

| Package | Não pode importar | Status |
|---------|------------------|:---:|
| `domain` | Nada externo (zero deps internas) | ✅ |
| `source` | httpapi, store, publish | ✅ |
| `engine` | httpapi, store | ✅ |
| `strategy` | httpapi, store, source | ✅ |
| `store` | httpapi, source, engine | ✅ |
| `tenant` | httpapi, source, store | ✅ |

**Compliance:** 100%

---

## Cobertura de Testes

### Backend (Go)

| Package | Cobertura | Responsabilidade |
|---------|:---------:|-----------------|
| logs | 90.9% | Logging estruturado |
| scoring | 89.7% | Cálculo de score |
| source | 86.7% | Adaptadores Shopee API |
| engine | 85.4% | Motor de ranking |
| strategy | 85.2% | Estratégias de curadoria |
| tenant | 81.3% | Multi-tenancy + crypto |
| coleta | 65.8% | Service de coleta periódica |
| publish | 63.7% | Publicação Telegram/WhatsApp |
| store | 57.1% | BigQuery persistence |
| httpapi | 49.2% | HTTP handlers |

**Packages sem testes (infraestrutura pura):** alerts, auth, scheduler, problem — dependem de serviços externos.

### Frontend (Svelte/Vitest)

| Componente | Testes |
|-----------|:------:|
| SeletorGrupo | 10 |
| CandidateCard | 16 |
| **Total** | **26** |

Cobertura: renderização, badges (origem, desconto, expiração, suspeito), interações.

---

## Tratamento de Erros (RFC 9457)

Todos os endpoints retornam erros no formato **Problem Details** ([RFC 9457](https://www.rfc-editor.org/rfc/rfc9457)):

```json
{
  "type": "https://garimpei.app.br/problemas/servico-externo",
  "title": "Serviço externo indisponível",
  "status": 502,
  "detail": "shopee api: timeout após 20s",
  "erro": "shopee api: timeout após 20s"
}
```

O frontend parseia o `detail` e mostra mensagem amigável + botão "Tentar novamente" quando `retry: true`.

---

## Organização do Código (httpapi)

O package `httpapi` está dividido em arquivos por domínio:

| Arquivo | Responsabilidade | Linhas |
|---------|-----------------|:------:|
| httpapi.go | Server struct, Handler, rotas, SPA | 279 |
| curadoria.go | /candidatos, /comparar, enriquecerOrigem | ~170 |
| lojas.go | /lojas (CRUD, resolução de shopId) | ~340 |
| publicacoes.go | /publicacoes, agendamento | ~200 |
| alertas.go | /alertas (config, teste, update) | ~120 |
| onboarding.go | /onboarding (multi-tenant) | ~320 |
| produto_origem.go | /produto/origem (cache, normalização) | ~170 |
| introspect.go | /admin/shopee-introspect | ~130 |
| helpers.go | writeJSON, writeErr, auth helpers | 92 |
| middleware.go | logRequests, CORS | 70 |

---

## Correções realizadas (sessão 27/06)

### Segurança
- [x] Go 1.26.0 → 1.26.4 (12 CVEs resolvidas)
- [x] `http.Server` com timeouts (era sem timeout)
- [x] Race condition em `os.Setenv` → protegido por mutex

### Bugs
- [x] 4 errcheck corrigidos (json.Decode, w.Write sem tratar)
- [x] 2 funções mortas removidas
- [x] shopType na query causava 502 (removido)
- [x] shopId fallback usava campo errado (offerLink vs productLink)
- [x] carregarBusca só rodava para ShopeeShopSource (origem_padrao não aplicava)

### Manutenibilidade
- [x] Split httpapi.go (429→279 + helpers + middleware)
- [x] funlen + cyclop habilitados (previne crescimento futuro)
- [x] arch-go com 6 regras de dependência
- [x] 16 testes novos no CandidateCard
- [x] 4 testes novos no coleta service (origem)
- [x] 7 testes novos no tenant (config, store, crypto)

---

## Como testar tudo localmente antes do push

```bash
# Backend
go test ./...
golangci-lint run --timeout=3m --issues-exit-code=1 ./...
govulncheck ./...
arch-go

# Frontend
cd web
npm run lint:js
npm run lint:css
npx vitest run
npm run build
```

Se tudo passar localmente, o CI também passa.
