# Sessão 2026-06-28 — Refactors Completos: Erros + Repository

---

## 1. Tratamento de Erros (T-0008)

### Problema

O projeto tinha 65+ violações de lint (err113, wrapcheck) indicando:
- Erros criados como strings dinâmicas (`fmt.Errorf("falhou")`) — impossível testar com `errors.Is`
- Erros de pacotes externos (net/http, io, crypto) retornados sem wrapping — sem contexto de onde vieram
- Frontend com `catch(() => {})` silenciosos — erros engolidos sem feedback ao usuário

### Solução

#### Backend: pacote `internal/apperr`

13 erros sentinel, agrupados por domínio:

| Sentinel | Quando usar |
|----------|------------|
| `ErrShopeeAPI` | Qualquer falha da API de afiliados |
| `ErrTelegram` | Falha no Telegram Bot API |
| `ErrMaytapi` | Falha na Maytapi (WhatsApp) |
| `ErrInvalidInput` | Dado inválido do usuário |
| `ErrNotFound` | Recurso não existe |
| `ErrInactive` | Recurso desabilitado |
| `ErrUnauthorized` | Sem autenticação |
| `ErrForbidden` | Sem permissão |
| `ErrCrypto` | Falha criptográfica (AES, GCM) |
| `ErrIO` | Falha de I/O genérica |
| `ErrNoConfig` | Credenciais/config ausentes |
| `ErrTooManyRedirects` | Excesso de redirects |
| `ErrNoProvider` | Provedor de envio não registrado |
| `ErrCSV` | Erro no parsing de CSV |

**Padrão de uso:**
```go
// Cria erro com contexto + sentinel na raiz
return fmt.Errorf("telegram enviar grupo %s: %w", groupID, apperr.ErrTelegram)

// Testa em qualquer camada
if errors.Is(err, apperr.ErrTelegram) { /* retry ou fallback */ }
```

#### Frontend: `web/src/lib/errors.js`

Classificação unificada de erros HTTP:

```javascript
import { isAuthError, isRetryable, mensagemAmigavel } from '$lib/errors.js';

try {
  await buscarCandidatos();
} catch (err) {
  if (isAuthError(err)) goto('/login');
  if (isRetryable(err)) mostrarRetry();
  toast.error(mensagemAmigavel(err));
}
```

Helpers disponíveis:
- `isAuthError(err)` — 401
- `isForbiddenError(err)` — 403
- `isNotFoundError(err)` — 404
- `isValidationError(err)` — 400
- `isExternalServiceError(err)` — 502/503
- `isRetryable(err)` — backend indica `retry: true`
- `isNetworkError(err)` — fetch falhou (sem rede)
- `mensagemAmigavel(err)` — texto para toast/alert
- `codigoErro(err)` — código interno (ex: `"servico_externo"`)

---

## 2. Repository Pattern

### Problema

- **God interface** (`EventoStore`) com 17 métodos misturando 7 domínios
- **Interfaces espalhadas** em 3 pacotes sem relação formal
- **Server com 5 campos** de store sem agregação
- **Mocks de teste** precisavam implementar 17 métodos para testar 1 handler

### Solução: Interfaces Segregadas + Repository

```
store.Repository
├── Eventos()     → EventoRepo       (Registrar)
├── Snapshots()   → SnapshotRepo     (RegistrarSnapshot, Estatisticas, HistoricoColetas, Novidades, EvolucaoLojas)
├── Buscas()      → BuscaRepo        (SalvarBusca, ListarBuscas)
├── Publicacoes() → PublicacaoRepo   (SalvarPublicacao, ListarPublicacoes, AtualizarPublicacao, Conversoes)
├── Destinos()    → DestinoRepo      (ListarDestinos, BuscarDestino, SalvarDestino, DeletarDestino)
├── Templates()   → TemplateRepo     (ListarTemplates, BuscarTemplate, SalvarTemplate, DeletarTemplate)
├── Favoritos()   → FavoritoRepo     (SalvarFavorito, ListarFavoritos, RemoverFavorito)
├── Tenants()     → TenantRepo       (BuscarTenant, SalvarTenant, ExcluirTenant)
├── EnsureSchema(ctx)
└── Nome()
```

### Implementações

| | Dev/Testes | Produção |
|---|---|---|
| Build tag | `!gcp` (default) | `gcp` |
| Tipo | `NopRepository` | `BQRepository` |
| Destinos | `MemDestinoRepo` (map) | `BQDestinoStore` (BigQuery) |
| Templates | `MemTemplateRepo` (map) | `BQTemplateStore` (BigQuery) |
| Tenants | `MemTenantRepo` (map) | Injetável (Firestore, etc.) |
| Outros | Nop (descarta) | `BigQueryStore` |

### Como consumers usam

```go
// Handler declara dependência apenas no que precisa
func (srv *Server) listarFavoritos(w http.ResponseWriter, r *http.Request) {
    favs, err := srv.Repo.Favoritos().ListarFavoritos(r.Context(), uid)
}

// Service recebe o Repository inteiro
svc := coleta.Novo(coleta.Deps{Repo: srv.Repo, Logger: logger})

// Alerter recebe apenas SnapshotRepo (ISP)
alerter.VerificarENotificar(ctx, srv.Repo.Snapshots(), buscaID)
```

---

## 3. Barreiras contra drift (testes + CI)

### Barreiras existentes (implementadas nesta sessão)

| Barreira | O que protege | Onde |
|----------|--------------|------|
| `golangci-lint` err113 | Erros devem ser sentinel vars | CI: deploy-gcp.yml |
| `golangci-lint` wrapcheck | Erros externos devem ser wrapped | CI: deploy-gcp.yml |
| `golangci-lint` errorlint | Comparações usam `errors.Is/As` | CI: deploy-gcp.yml |
| `apperr/errors_test.go` | Sentinels distintos + unwrappable | `go test` |
| `store/conformance_test.go` | Implementações satisfazem interfaces | `go test` (compile-time) |
| `arch-go` regra apperr | `apperr` é leaf (zero imports internos) | CI: arch-go |
| `arch-go` regra store | `store` não importa httpapi/source/engine | CI: arch-go |
| `arch-go` regra domain | `domain` não importa ninguém | CI: arch-go |
| `arch-go` regra source | `source` não importa store/publish/httpapi | CI: arch-go |
| `arch-go` regra engine | `engine` não importa httpapi/store | CI: arch-go |
| `arch-go` regra strategy | `strategy` não importa httpapi/store/source | CI: arch-go |
| `arch-go` regra tenant | `tenant` não importa httpapi/source/engine | CI: arch-go |

### Barreiras recomendadas (ainda não implementadas)

| Barreira | O que protegeria | Esforço |
|----------|-----------------|---------|
| **Teste de contrato BQ** | Valida que queries BQ rodam contra schema real (sandbox) | Médio — precisa de dataset de teste no CI |
| **Proibir `store.EventoStore` em código novo** | Impede que alguém use a god interface legada (que ainda existe) | Baixo — adicionar `staticcheck` deprecation ou remover a interface |
| **Lint de imports `publish.DestinoStore`** | Impede uso das interfaces antigas de publish | Baixo — marcar como deprecated ou remover |
| **Teste de coverage mínimo por sub-interface** | Garante que cada Repo tem pelo menos 1 teste | Baixo — `go test -coverprofile` + script |
| **Regra arch-go: crypto é leaf** | `internal/crypto` não pode importar nada interno exceto `apperr` | Trivial |

---

## 4. O que é possível agora que não era antes

### Tratamento de erros

| Antes | Agora |
|-------|-------|
| `if err != nil` sem saber o tipo | `if errors.Is(err, apperr.ErrShopeeAPI)` — decisões por tipo |
| Logs com `err.Error()` opaco | Cadeia completa: `"shopee api erro 10020 invalid signature: shopee api"` |
| Frontend mostra "Erro 502" genérico | Frontend distingue 401/403/404/5xx com mensagens amigáveis |
| Impossível criar dashboards por tipo | Pronto para OTel: `span.SetAttributes(attribute.String("error.type", "shopee_api"))` |
| Testes verificam apenas se err != nil | Testes verificam `errors.Is(err, apperr.ErrX)` — pegam regressões de tipo |
| Retry manual no frontend | `isRetryable(err)` decide automaticamente se mostra botão retry |

### Repository Pattern

| Antes | Agora |
|-------|-------|
| Mock precisa implementar 17 métodos | Mock implementa 2-3 métodos (da sub-interface que o teste precisa) |
| Adicionar campo ao Server requer tocar 5 stores | Adiciona método na sub-interface correta |
| Não dá pra injetar tenant repo diferente (Firestore vs mem) | `TenantRepo` é interface — injeta qualquer implementação |
| Handler acoplado ao BigQuery (via EventoStore) | Handler depende de `SnapshotRepo` — pode testar com mock limpo |
| Publicador precisa de DestinoStore da publish | Destino é tipo canônico do `store` — qualquer repo pode fornecer |
| Trocar persistência requer mudar 15 arquivos | Implementa nova struct que satisfaz Repository — 1 arquivo |
| Sem garantia de que implement satisfaz interface | Compile-time checks em conformance_test.go |
| Regras de camada são informais (README) | arch-go bloqueia CI se alguém importar httpapi dentro de store |

### Crypto isolado

| Antes | Agora |
|-------|-------|
| Encrypt/Decrypt no pacote `tenant` (cycle risk) | Pacote `internal/crypto` leaf — importável por qualquer camada |
| Só tenant podia criptografar | Qualquer pacote pode usar crypto.Encrypt/Decrypt |

---

## 5. Estrutura de arquivos final

```
internal/
├── apperr/
│   ├── errors.go          ← 13 sentinels (leaf, zero imports internos)
│   └── errors_test.go     ← testes de conformidade
├── crypto/
│   └── crypto.go          ← AES-256-GCM (leaf, importa apenas apperr)
├── store/
│   ├── repository.go      ← interfaces segregadas + Repository
│   ├── memstore.go        ← NopRepository + Mem*Repo (dev/testes)
│   ├── bqrepository.go    ← BQRepository (gcp)
│   ├── bigquery_*.go      ← implementações BQ existentes
│   ├── destino.go         ← store.Destino (tipo canônico)
│   ├── template.go        ← store.Template (tipo canônico + Renderizar)
│   ├── tenant.go          ← store.TenantConfig (crypto methods)
│   ├── conformance_test.go← compile-time interface checks
│   └── ...
├── tenant/
│   ├── crypto.go          ← delegata para internal/crypto
│   ├── config.go          ← tenant.Config (legado, para backward compat)
│   └── ...
└── httpapi/
    └── httpapi.go         ← Server com Repo store.Repository (único campo de persistência)
```

---

## 6. Como verificar

```bash
go test ./...                  # 13 pacotes, todos OK
golangci-lint run ./...        # 0 issues (err113, wrapcheck, errorlint ativos)
arch-go                        # 7 regras, 100% compliance
cd web && npx vitest --run     # 109 testes frontend
```

---

## 7. Próximos passos sugeridos

1. **Remover `store.EventoStore` e `store.NopStore`** — a god interface ainda existe no código mas não é mais usada pelo httpapi ou coleta. Removê-la elimina a tentação de usá-la.
2. **Remover `publish.DestinoStore` e `publish.TemplateStore`** — substituídas por `store.DestinoRepo`/`store.TemplateRepo`. O pacote publish ainda as declara.
3. **Remover `tenant.Store`** — substituída por `store.TenantRepo`.
4. **Adicionar regra arch-go para `internal/crypto`** — garantir que é leaf.
5. **Implementar tenant via Firestore** — agora que `TenantRepo` é interface, basta criar `FirestoreTenantRepo`.
