---
title: "ADR 0010 — Tratamento de erros idiomático"
---

**Status:** aceito  
**Data:** 2026-06-28  
**Tarefa:** T-0008

## Contexto

O projeto acumulou 65+ issues de lint (err113, wrapcheck) indicando erros
sem contexto (strings dinâmicas) e erros de pacotes externos retornados sem
wrapping. Isso impedia:

- Testar erros com `errors.Is` / `errors.As`
- Criar dashboards por tipo de falha
- Rastrear a cadeia de causas em logs estruturados
- Preparar para OpenTelemetry (spans com error classification)

## Decisão

### Backend (Go)

1. **Pacote `internal/apperr`** com erros sentinel do domínio:
   - Serviços externos: `ErrShopeeAPI`, `ErrTelegram`, `ErrMaytapi`
   - Validação: `ErrInvalidInput`, `ErrNotFound`, `ErrInactive`
   - Infra: `ErrCrypto`, `ErrIO`, `ErrTooManyRedirects`
   - Config: `ErrNoConfig`, `ErrNoProvider`
   - Dados: `ErrCSV`

2. **Padrão de wrapping:** `fmt.Errorf("contexto operação: %w", apperr.ErrX)`
   - O sentinel fica na raiz (testável com `errors.Is`)
   - O contexto textual indica *onde* falhou

3. **Linters bloqueantes em CI** (sem exclusões de debt):
   - `err113` — erros devem ser sentinel vars
   - `wrapcheck` — erros externos devem ser wrapped
   - `errorlint` — comparações usam `errors.Is`/`errors.As`

### Frontend (Svelte)

1. Backend retorna **Problem Details (RFC 9457)** com campos:
   `{ type, title, status, detail, code, retry }`

2. `src/lib/errors.js` classifica erros por status/tipo:
   - `isAuthError(err)` → 401, redireciona para login
   - `isRetryable(err)` → mostra botão "tentar novamente"
   - `mensagemAmigavel(err)` → texto para toast/alert

3. Eliminados `catch(() => {})` silenciosos nas funções DELETE
   (agora usam `parseProblem` como GET/POST).

## Métricas

| Métrica | Antes | Depois |
|---------|-------|--------|
| Issues err113 | 29 | 0 |
| Issues wrapcheck | 36 | 0 |
| Issues errorlint | 0 | 0 |
| Exclusões de debt no .golangci.yml | 5 blocos | 0 |
| Erros testáveis com `errors.Is` | 0 | 13 sentinels |
| `catch(() => {})` silenciosos no frontend | 3 | 0 |

## Consequências

### Positivas

- Todo erro retornado pode ser testado com `errors.Is(err, apperr.ErrX)`
- Logs estruturados mostram a cadeia completa de contexto
- Frontend distingue 401/403/404/5xx e reage adequadamente
- Pronto para OpenTelemetry: `span.RecordError(err)` com classification
- CI bloqueia regressões — novo código já nasce correto

### Negativas

- Mensagens de erro agora incluem o sentinel no `Unwrap()`, o que pode
  expor nomes internos em logs (mitigado: sentinels são genéricos)
- Dependência circular potencial se outro pacote importar `apperr` — mitigado
  porque `apperr` não importa nada do projeto

## Preparação para OpenTelemetry

Com sentinels, futuramente poderemos:

```go
span.SetAttributes(attribute.String("error.type", "shopee_api"))
span.RecordError(err)
```

E criar dashboards como:
- Taxa de erro por sentinel (Shopee vs Telegram vs Crypto)
- Latência por tipo de falha
- Alertas quando ErrShopeeAPI ultrapassa threshold
