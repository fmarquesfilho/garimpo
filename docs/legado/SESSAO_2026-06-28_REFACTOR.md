# Sessão 2026-06-28 — Refactors: Tratamento de Erros + Repository Pattern

## Resumo executivo

Duas mudanças estruturais grandes foram aplicadas nesta sessão:

1. **T-0008 — Tratamento de erros idiomático** (concluído 100%)
2. **Repository Pattern** (estrutura no lugar, migração de consumers em andamento)

---

## 1. Tratamento de Erros (T-0008)

### O que foi feito

| Item | Antes | Depois |
|------|-------|--------|
| Issues err113 | 29 | 0 |
| Issues wrapcheck | 36 | 0 |
| Issues errorlint | 0 | 0 |
| Exclusões de debt no .golangci.yml | 5 blocos | 0 |
| Erros testáveis com `errors.Is` | 0 | 13 sentinels |
| `catch(() => {})` silenciosos no frontend | 3 | 0 |

### Pacote `internal/apperr`

Erros sentinel do domínio, sem dependências:

```go
apperr.ErrShopeeAPI         // falha na API de afiliados
apperr.ErrTelegram          // falha no Telegram Bot API
apperr.ErrMaytapi           // falha na Maytapi (WhatsApp)
apperr.ErrInvalidInput      // dado inválido do usuário
apperr.ErrNotFound          // recurso não encontrado
apperr.ErrInactive          // recurso desabilitado
apperr.ErrUnauthorized      // falta autenticação
apperr.ErrForbidden         // falta permissão
apperr.ErrCrypto            // falha criptográfica
apperr.ErrIO                // falha de I/O
apperr.ErrNoConfig          // configuração ausente
apperr.ErrTooManyRedirects  // excesso de redirects
apperr.ErrNoProvider        // provedor não registrado
apperr.ErrCSV               // erro no parsing de CSV
```

### Padrão de uso

```go
// Wrapping com contexto (identificar ONDE falhou)
return fmt.Errorf("telegram enviar grupo %s: %w", groupID, apperr.ErrTelegram)

// Teste em camadas superiores
if errors.Is(err, apperr.ErrTelegram) {
    // retry ou fallback
}
```

### Frontend (`web/src/lib/errors.js`)

Classificação por status HTTP com helpers:
- `isAuthError(err)` — 401
- `isRetryable(err)` — backend indica `retry: true`
- `isExternalServiceError(err)` — 502/503
- `mensagemAmigavel(err)` — texto para toast

### Travas no CI

- `golangci-lint` com err113 + wrapcheck + errorlint **sem exclusões** — bloqueia deploy
- Teste `internal/apperr/errors_test.go`:
  - Sentinels são distintos entre si
  - `errors.Is` funciona com wrapping (1 e 2 níveis)
  - Mensagem preserva contexto textual

### Arquivos modificados (15 Go + 2 frontend)

```
internal/apperr/errors.go          ← NOVO
internal/apperr/errors_test.go     ← NOVO
web/src/lib/errors.js              ← NOVO
.golangci.yml                      ← removidas exclusões
internal/alerts/alerts.go
internal/coleta/service.go
internal/engine/engine.go
internal/httpapi/conversoes_sync.go
internal/httpapi/introspect.go
internal/httpapi/onboarding.go
internal/httpapi/resolver.go
internal/httpapi/shopee_resolver.go
internal/httpapi/whatsapp.go
internal/publish/canais.go
internal/publish/dispatcher.go
internal/publish/telegram.go
internal/publish/template.go
internal/publish/whatsapp.go
internal/source/csv.go
internal/source/flex.go
internal/source/shopee.go
internal/source/shopee_shop.go
internal/tenant/crypto.go
web/src/lib/api.js
docs/decisoes/0010-error-handling.md
```

---

## 2. Repository Pattern

### O que foi feito

| Item | Antes | Depois |
|------|-------|--------|
| Interface principal | `EventoStore` (17 métodos, god interface) | 8 interfaces segregadas |
| Agregador | Nenhum (5 campos no Server) | `store.Repository` |
| Tipos canônicos | Espalhados (publish, tenant) | Centralizados no `store` |
| Factories | `store.Novo()` retorna EventoStore | `store.NovoRepository()` retorna Repository |
| Conformidade | Nenhum | Compile-time checks + arch-go |

### Interfaces segregadas

```
store.Repository
├── Eventos()     → EventoRepo (1 método)
├── Snapshots()   → SnapshotRepo (5 métodos)
├── Buscas()      → BuscaRepo (2 métodos)
├── Publicacoes() → PublicacaoRepo (4 métodos)
├── Destinos()    → DestinoRepo (4 métodos)
├── Templates()   → TemplateRepo (4 métodos)
├── Favoritos()   → FavoritoRepo (3 métodos)
├── Tenants()     → TenantRepo (3 métodos)
├── EnsureSchema()
└── Nome()
```

### Implementações

- **NopRepository** — dev/testes, tudo em memória
- **BQRepository** — produção, compõe BigQueryStore + BQDestinoStore + BQTemplateStore

### Migração gradual (em andamento)

O `repoEventoStoreAdapter` no httpapi faz bridge: handlers continuam usando
`srv.Eventos` (interface antiga), mas por baixo delegam para o Repository.

### O que falta (próxima sessão)

1. Migrar cada handler para `srv.Repo.Destinos()`, `srv.Repo.Templates()`, etc.
2. Migrar `coleta/service.go` para receber sub-interfaces
3. Remover `publish.DestinoStore`, `publish.TemplateStore`, `tenant.Store`
4. Remover `store.EventoStore`, `store.NopStore`
5. Remover bridge adapter

### Travas no CI

- `arch-go` — 7 regras de dependência (100% compliance)
  - `domain` não importa ninguém
  - `apperr` não importa ninguém
  - `store` não importa httpapi/source/engine
  - `source` não importa store/publish/httpapi
  - `engine` não importa httpapi/store
  - `strategy` não importa httpapi/store/source
  - `tenant` não importa httpapi/source/engine
- `conformance_test.go` — compile-time interface checks

### Arquivos criados

```
internal/store/repository.go       ← interfaces segregadas
internal/store/memstore.go         ← NopRepository + Mem*Repo
internal/store/bqrepository.go     ← BQRepository (gcp)
internal/store/destino.go          ← store.Destino
internal/store/template.go         ← store.Template
internal/store/tenant.go           ← store.TenantConfig
internal/store/conformance_test.go ← compile-time checks
internal/tenant/adapter.go         ← bridge tenant.Store → store.TenantRepo
docs/decisoes/0011-repository-pattern.md
```

---

## Commits desta sessão

```
ec6de49 feat(T-0008): tratamento de erros idiomático — sentinels, wrapping, Problem Details
b91acd4 refactor(store): adiciona Repository pattern com interfaces segregadas
d592154 refactor(httpapi): bridge adapter conecta Repository ao EventoStore legado
753a4ae refactor(store): testes de conformidade + travas arquiteturais + ADR 0011
2ce3391 test(apperr): testes unitários dos erros sentinel
```

## Como verificar que está tudo verde

```bash
go test ./...                  # testes unitários
golangci-lint run ./...        # lint (err113, wrapcheck, errorlint)
arch-go                        # regras de arquitetura
cd web && npx vitest --run     # testes frontend
make docs-check                # docs geradas atualizadas
```
