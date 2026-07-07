# Garimpei — Web App C# (ASP.NET Core 10)

API principal do Garimpei. Minimal API com Clean Architecture.

## Estrutura

```
src/
├── Garimpei.Api/             # Minimal API (endpoints, middleware, Program.cs)
│   └── Endpoints/            # Arquivos de endpoint por domínio
├── Garimpei.Application/     # Casos de uso (MediatR handlers, validators)
├── Garimpei.Domain/          # Entidades, interfaces, value objects, services
│   ├── Entities/             # Busca, Product, Favorito, Destino, Template, Publicacao, TenantConfig
│   ├── Interfaces/           # IOwnedEntity, ITenantContext, IRepositories
│   ├── ValueObjects/         # ScoredProduct, ProductCandidate, EligibilityFilter, PoolStats
│   └── Services/             # ScoringService (static, stateless)
├── Garimpei.Infrastructure/  # EF Core (PostgreSQL), gRPC clients, tenancy
│   ├── Persistence/          # AppDbContext, migrations
│   └── Tenancy/              # TenantContext (scoped per-request)
├── Garimpei.Protos/          # Proto-generated gRPC client stubs (Collector, Publisher, Scheduler)
└── Garimpei.Tests/           # xUnit (68 testes: persistence + architecture + integração)
    ├── Persistence/          # Multi-tenant query filter tests
    ├── Integration/          # Onboarding, JSON binding, publish flow, buscas agendadas
    ├── Services/             # CouponDeduplication
    ├── Tenancy/              # TenantContext unit tests
    └── Architecture/         # NetArchTest fitness functions (13 regras)
```

## Endpoints

| Rota | Descrição |
|------|-----------|
| GET `/api/health` | Health check |
| GET `/api/admin/me` | Verifica se é admin |
| GET `/api/candidatos` | Busca + ranking (scoring engine) |
| GET `/api/categorias` | Categorias por marketplace |
| GET/POST `/api/buscas` | Perfis de busca (sync servidor) |
| GET/POST/DELETE `/api/lojas` | Monitoramento de lojas (ResolveShop + Scheduler) |
| GET `/api/lojas/novidades` | Produtos novos + variações (proxy analyzer) |
| GET `/api/lojas/evolucao` | Série temporal preço (proxy analyzer) |
| GET/POST/DELETE `/api/favoritos` | Favoritos |
| GET/POST/DELETE `/api/destinos` | Canais de publicação |
| GET/POST/DELETE `/api/templates` | Templates de mensagem |
| POST `/api/templates/preview` | Preview de template |
| POST `/api/publicar` | Publicar oferta (via gRPC publisher) |
| GET/POST `/api/publicacoes` | Publicações agendadas/enviadas |
| GET `/api/alertas` | Config de alertas |
| POST `/api/alertas/testar` | Teste de alerta |
| POST `/api/alertas/configurar` | Atualizar alertas |
| GET `/api/onboarding/status` | Status do onboarding |
| POST `/api/onboarding/*` | Steps do onboarding (termos, shopee, telegram, whatsapp, validar) |
| POST `/api/onboarding/excluir-conta` | LGPD: excluir dados |
| GET `/api/conversoes/reais` | Conversões Shopee (proxy analyzer) |
| GET `/api/estatisticas` | Dashboard (proxy analyzer) |
| GET `/api/coletas` | Histórico coletas (proxy analyzer) |
| POST `/api/resolver-link` | Resolver link curto Shopee |

## gRPC clients (sidecars)

| Client | Porta | Uso |
|--------|-------|-----|
| `CollectorServiceClient` | 50051 | ResolveShop, Fetch, FetchShop, GenerateAffiliateLink |
| `PublisherServiceClient` | 50052 | Publish, ListGroups |
| `SchedulerServiceClient` | 50054 | SetSchedule (criar/pausar jobs de coleta) |

## Setup local

```bash
# Pré-requisitos: .NET SDK 10.0+, mise (task runner)

# Subir PostgreSQL + API
mise run up

# Testes (aplica migrations automaticamente)
mise run test:csharp

# Ou manualmente:
dotnet restore src/Garimpei.sln
dotnet ef database update --project src/Garimpei.Infrastructure --startup-project src/Garimpei.Api
dotnet run --project src/Garimpei.Api
# → http://localhost:5000 (Development mode = bypass auth)
```

## Multi-tenancy

Toda entidade que implementa `IOwnedEntity` é filtrada automaticamente pelo `owner_uid`
do JWT. O `TenantMiddleware` extrai o claim `user_id` do Firebase JWT e seta o `TenantContext`
(scoped per-request). O `AppDbContext` aplica global query filters em todas as queries.

## Fitness functions

13 testes NetArchTest validam regras de Clean Architecture. Rodam via `dotnet test`.
Veja `Garimpei.Tests/Architecture/ArchitectureTests.cs`.
