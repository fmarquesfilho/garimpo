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

- Período de coexistência: `EventoStore` legado ainda existe durante a migração
- Bridge adapter adiciona uma camada de indireção (será removida ao final)
- Tipos duplicados temporariamente (`publish.Destino` + `store.Destino`)

### Próximos passos

1. Migrar handlers do httpapi para usar `srv.Repo.X()` diretamente
2. Migrar `coleta/service.go` para receber sub-interfaces
3. Remover `publish.DestinoStore`, `publish.TemplateStore`, `tenant.Store`
4. Remover `store.EventoStore` e `store.NopStore`
5. Remover bridge adapter
