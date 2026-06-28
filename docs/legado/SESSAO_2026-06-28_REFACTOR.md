# Sessão 2026-06-28 — Refactors Completos

## Resumo executivo

Duas mudanças estruturais completas nesta sessão:

1. **T-0008 — Tratamento de erros idiomático** ✅ 100%
2. **Repository Pattern** ✅ 100% (migração completa, sem legado)

---

## 1. Tratamento de Erros (T-0008)

### Métricas

| Métrica | Antes | Depois |
|---------|-------|--------|
| Issues err113 | 29 | 0 |
| Issues wrapcheck | 36 | 0 |
| Issues errorlint | 0 | 0 |
| Exclusões de debt no .golangci.yml | 5 blocos | 0 |
| Erros testáveis com `errors.Is` | 0 | 13 sentinels |
| `catch(() => {})` silenciosos no frontend | 3 | 0 |

### Pacote `internal/apperr`

Erros sentinel do domínio, leaf package sem dependências:

```go
apperr.ErrShopeeAPI         // API de afiliados
apperr.ErrTelegram          // Telegram Bot API
apperr.ErrMaytapi           // Maytapi (WhatsApp)
apperr.ErrInvalidInput      // dado inválido
apperr.ErrNotFound          // recurso não encontrado
apperr.ErrInactive          // recurso desabilitado
apperr.ErrUnauthorized      // sem autenticação
apperr.ErrForbidden         // sem permissão
apperr.ErrCrypto            // falha criptográfica
apperr.ErrIO                // falha de I/O
apperr.ErrNoConfig          // configuração ausente
apperr.ErrTooManyRedirects  // excesso de redirects
apperr.ErrNoProvider        // provedor não registrado
apperr.ErrCSV               // erro no parsing CSV
```

### Padrão de uso

```go
return fmt.Errorf("telegram enviar grupo %s: %w", groupID, apperr.ErrTelegram)

if errors.Is(err, apperr.ErrTelegram) { /* retry */ }
```

### Frontend (`web/src/lib/errors.js`)

- `isAuthError(err)` → 401
- `isRetryable(err)` → retry: true
- `isExternalServiceError(err)` → 502/503
- `mensagemAmigavel(err)` → texto para toast

### Travas no CI

- **golangci-lint** — err113 + wrapcheck + errorlint sem exclusões (bloqueia deploy)
- **`internal/apperr/errors_test.go`** — 3 testes:
  - Sentinels distintos entre si
  - `errors.Is` funciona com wrapping em 1 e 2 níveis
  - Mensagem preserva contexto + sentinel

---

## 2. Repository Pattern (migração completa)

### Antes vs Depois

| Aspecto | Antes | Depois |
|---------|-------|--------|
| Interface principal | `EventoStore` (17 métodos) | 8 interfaces segregadas |
| Agregador | Nenhum — 5 campos no Server | `store.Repository` |
| Server fields | Eventos, Destinos, Templates, Tenants, Publicador | Repo, Publicador |
| Tipos canônicos | Espalhados (publish, tenant) | `store.Destino`, `store.Template`, `store.TenantConfig` |
| Wiring | `criarStoresAuxiliares()` + build tags | `store.NovoRepository()` |
| Crypto | Em `tenant` (causava import cycle) | Em `internal/crypto` (leaf) |
| Bridge adapter | Temporário | **Removido** |

### Arquitetura final

```
store.Repository
├── Eventos()     → EventoRepo       (1 método)
├── Snapshots()   → SnapshotRepo     (5 métodos)
├── Buscas()      → BuscaRepo        (2 métodos)
├── Publicacoes() → PublicacaoRepo   (4 métodos)
├── Destinos()    → DestinoRepo      (4 métodos)
├── Templates()   → TemplateRepo     (4 métodos)
├── Favoritos()   → FavoritoRepo     (3 métodos)
├── Tenants()     → TenantRepo       (3 métodos)
├── EnsureSchema()
└── Nome()
```

### Implementações

| Implementação | Uso | Build tag |
|---|---|---|
| `NopRepository` | Dev/testes (memória) | `!gcp` (default) |
| `BQRepository` | Produção (BigQuery) | `gcp` |

### Como um consumer declara dependência

```go
// Handler recebe apenas o que precisa (ISP)
func (srv *Server) listarFavoritos(w http.ResponseWriter, r *http.Request) {
    favs, err := srv.Repo.Favoritos().ListarFavoritos(r.Context(), uid)
}

// Service de coleta recebe o Repository inteiro
svc := coleta.Novo(coleta.Deps{Repo: srv.Repo, Logger: logger})
```

### Travas no CI

- **arch-go** — 7 regras (100% compliance):
  - `apperr` é leaf (sem imports internos)
  - `store` não importa httpapi/source/engine
  - `domain` não importa ninguém
- **`conformance_test.go`** — compile-time checks garantem que MemDestinoRepo/MemTemplateRepo/etc satisfazem as interfaces

---

## Pacote `internal/crypto` (novo)

Extração de `tenant/crypto.go` para um pacote leaf sem dependências além de `apperr`. Ambos `store` e `tenant` importam sem ciclo.

---

## Arquivos criados/removidos

### Criados
```
internal/apperr/errors.go
internal/apperr/errors_test.go
internal/crypto/crypto.go
internal/store/repository.go
internal/store/memstore.go
internal/store/bqrepository.go
internal/store/destino.go
internal/store/template.go
internal/store/tenant.go
internal/store/conformance_test.go
web/src/lib/errors.js
docs/decisoes/0010-error-handling.md
docs/decisoes/0011-repository-pattern.md
```

### Removidos
```
cmd/garimpo-api/stores_default.go
cmd/garimpo-api/stores_gcp.go
internal/tenant/adapter.go
```

---

## Commits

```
ec6de49 feat(T-0008): tratamento de erros idiomático
b91acd4 refactor(store): adiciona Repository pattern com interfaces segregadas
d592154 refactor(httpapi): bridge adapter conecta Repository ao EventoStore legado
753a4ae refactor(store): testes de conformidade + travas arquiteturais + ADR 0011
2ce3391 test(apperr): testes unitários dos erros sentinel
e9f6a8b docs: documento consolidado da sessão 2026-06-28
13eb78f refactor(store): migração completa — remove campos legados e bridge adapter
```

## Verificação

```bash
go test ./...                  # 13 pacotes OK
golangci-lint run ./...        # 0 issues
arch-go                        # 7/7 regras, 100% compliance
cd web && npx vitest --run     # 109 testes frontend
```
