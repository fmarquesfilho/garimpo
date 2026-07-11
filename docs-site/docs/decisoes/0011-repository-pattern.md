# ADR 0011 — Repository Pattern para a camada de persistência

**Status:** aceito  
**Data:** 2026-06-28  

## Contexto

A camada de persistência do Garimpei tinha problemas estruturais:

1. **God Interface** — `EventoStore` com 17 métodos misturando eventos, snapshots,
   buscas, publicações, novidades, evolução e favoritos num único contrato
2. **Interfaces espalhadas** — `publish.DestinoStore`, `publish.TemplateStore` e
   `tenant.Store` viviam em pacotes separados sem relação formal
3. **Tipos de domínio acoplados** — `Destino` e `Template` definidos no pacote
   `publish` em vez de no pacote de persistência
4. **Server com muitos campos** — `httpapi.Server` recebia 5 stores diferentes
   sem agregação

## Decisão

### Interfaces segregadas (Interface Segregation Principle)

Quebramos `EventoStore` em 8 sub-interfaces por domínio:

| Interface | Responsabilidade | Métodos |
|-----------|-----------------|---------|
| `EventoRepo` | Registrar decisões de curadoria | 1 |
| `SnapshotRepo` | Snapshots de mercado, estatísticas, novidades | 5 |
| `BuscaRepo` | Perfis de busca/coleta | 2 |
| `PublicacaoRepo` | Publicações agendadas, conversões | 4 |
| `DestinoRepo` | CRUD de destinos de publicação | 4 |
| `TemplateRepo` | CRUD de templates de mensagem | 4 |
| `FavoritoRepo` | Produtos salvos pelo usuário | 3 |
| `TenantRepo` | Config por tenant (multi-tenancy) | 3 |

### Repository como agregador

```go
type Repository interface {
    Eventos() EventoRepo
    Snapshots() SnapshotRepo
    Buscas() BuscaRepo
    Publicacoes() PublicacaoRepo
    Destinos() DestinoRepo
    Templates() TemplateRepo
    Favoritos() FavoritoRepo
    Tenants() TenantRepo
    EnsureSchema(ctx context.Context) error
    Nome() string
}
```

### Implementações

- `NopRepository` — memória (dev/testes), com `MemDestinoRepo`, `MemTemplateRepo`, etc.
- `BQRepository` — BigQuery (produção), compõe `BigQueryStore` existente

### Tipos canônicos no pacote `store`

- `store.Destino` — antes vivia em `publish`
- `store.Template` — antes vivia em `publish`
- `store.TenantConfig` — antes vivia em `tenant`

### Bridge para migração gradual

Um `repoEventoStoreAdapter` no httpapi faz bridge entre o `Repository` e o
`EventoStore` legado, permitindo migrar handlers um a um sem quebrar nada.

## Consequências

### Positivas

- Cada consumer declara dependência apenas na sub-interface que precisa
  (ex.: `func meuHandler(buscas store.BuscaRepo)`)
- Mocks de teste implementam 2-3 métodos em vez de 17
- Teste de conformidade compile-time garante que implementações satisfazem interfaces
- `arch-go` protege contra drift (impede store → httpapi, etc.)
- Factory com build tags (`NovoRepository`) isola BQ do build padrão

### Negativas

- Período de transição requer atualização dos imports nos consumers
- Tipos temporariamente duplicados (`publish.Destino` + `store.Destino`)
  até remoção dos antigos numa sessão futura

### Removidos na migração

- `store.EventoStore` (god interface) → **removida** do código
- `store.NopStore` → **removido** (substituído por `NopRepository`)
- `store.Novo()` → **removido** (substituído por `store.NovoRepository()`)
- `publish.DestinoStore` / `publish.TemplateStore` → mantidas no publish (uso interno do Dispatcher), mas httpapi usa `store.DestinoRepo` / `store.TemplateRepo`
- `tenant.Store` → mantida para backward compat com testes de tenant, mas httpapi usa `store.TenantRepo`
- `tenant.RepoAdapter` → desnecessário (crypto movido para `internal/crypto`)
- `cmd/garimpo-api/stores_default.go` / `stores_gcp.go` → `store.NovoRepository()`
