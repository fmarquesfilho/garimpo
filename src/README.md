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
├── Garimpei.Protos/          # Proto-generated gRPC client stubs (Collector, Publisher)
└── Garimpei.Tests/           # xUnit (persistence + architecture fitness functions)
    ├── Persistence/          # Multi-tenant query filter tests
    ├── Tenancy/              # TenantContext unit tests
    └── Architecture/         # NetArchTest fitness functions (13 regras)
```

## Endpoints

### Compatibilidade (`/api/*`) — formato usado pelo frontend

| Rota | Descrição |
|------|-----------|
| GET `/api/health` | Health check |
| GET `/api/admin/me` | Verifica se é admin |
| GET `/api/candidatos` | Busca + ranking (scoring engine) |
| GET/POST `/api/buscas` | Perfis de busca |
| GET/POST/DELETE `/api/lojas` | Monitoramento de lojas |
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
| POST `/api/onboarding/termos` | Step 1: aceitar termos |
| POST `/api/onboarding/shopee` | Step 2: credenciais Shopee |
| POST `/api/onboarding/telegram` | Step 3: bot Telegram |
| POST `/api/onboarding/validar` | Step 4: validar credenciais |
| POST `/api/onboarding/excluir-conta` | LGPD: excluir dados |
| GET `/api/conversoes` | Relatório conversões |
| GET `/api/conversoes/reais` | Conversões Shopee (proxy analyzer) |
| GET `/api/estatisticas` | Dashboard (proxy analyzer) |
| GET `/api/coletas` | Histórico coletas (proxy analyzer) |
| POST `/api/resolver-link` | Resolver link curto Shopee |

### V2 (`/api/v2/*`) — formato nativo C#

| Rota | Descrição |
|------|-----------|
| GET `/api/v2/buscas` | CRUD buscas (EF Core) |
| GET `/api/v2/curadoria/ranking` | Ranking por keyword |
| GET `/api/v2/curadoria/ranking/shop` | Ranking por loja |
| GET `/api/v2/curadoria/quedas` | Quedas de preço (proxy analyzer) |
| GET `/api/v2/curadoria/novos` | Produtos novos (proxy analyzer) |
| GET `/api/v2/curadoria/favoritos` | Favoritos do tenant |
| POST `/api/v2/publicar` | Publicar (gRPC publisher) |
| GET `/api/v2/publicar/destinos` | Destinos disponíveis |

## Setup local

```bash
# Pré-requisitos: .NET SDK 10.0+, PostgreSQL 17+ (ou Docker)

# Subir PostgreSQL
docker compose up -d postgres

# Restore + migrations
dotnet restore
dotnet ef database update --project Garimpei.Infrastructure --startup-project Garimpei.Api

# Rodar (Development mode = bypass auth para teste local)
dotnet run --project Garimpei.Api
# → http://localhost:5000

# Testes
dotnet test
```

## Multi-tenancy

Toda entidade que implementa `IOwnedEntity` é filtrada automaticamente pelo `owner_uid`
do JWT. O `TenantMiddleware` extrai o claim `user_id` do Firebase JWT e seta o `TenantContext`
(scoped per-request). O `AppDbContext` aplica global query filters em todas as queries.

## Fitness functions

13 testes NetArchTest validam regras de Clean Architecture. Rodam via `dotnet test`.
Veja `Garimpei.Tests/Architecture/ArchitectureTests.cs`.
