# Qualidade e testes

## Pipeline de CI

O workflow `ci.yml` roda em push para `main` e PRs:

```
push main → GitHub Actions (ci.yml)
  │
  ├─ go: build + test + lint + arch-go + docs-check + file-size
  ├─ csharp: restore + build + test (com PostgreSQL service)
  ├─ python: ruff lint + syntax check
  ├─ proto: buf lint + sync check (Go + C# stubs atualizados?)
  ├─ frontend: npm ci + build + lint:css + lint:js + vitest
  ├─ api-contract: check-api-contract + check-config-consistency + check-schema-sync
  ├─ docker: build all 6 images (validação de Dockerfiles)
  └─ docs-deploy: sync + build + deploy Cloudflare Pages (apenas push main)
```

Pushes que só tocam `docs/legado/**`, `docs/meta/**` ou `README.md` são ignorados.

---

## Estratégia de testes

### Cobertura por camada

| Camada | Ferramenta | Testes | Foco |
|--------|-----------|--------|------|
| Go (internal) | go test | ~200 | source 87%, publish 62%, store 36% |
| Go (services) | go test | 12 | Validações + fluxos gRPC |
| C# (Domain + Infra) | xUnit | 10 | Multi-tenant, persistence, isolation |
| C# (Arquitetura) | xUnit + NetArchTest | 13 | Fitness functions (regras Clean Architecture) |
| Frontend (unit) | Vitest | 109 | Componentes, stores, utils |
| Frontend (E2E) | Playwright | ~10 | Fluxos críticos do usuário |
| Cross-stack (drift) | Shell scripts | 3 | API contract, config, schema sync |

### BDD (Behaviour-Driven Development)

Os testes seguem cenários Given/When/Then:

- **Given** — estado inicial (mock de dados, configuração)
- **When** — ação do usuário ou trigger do sistema
- **Then** — resultado esperado (response, efeito colateral)

---

## Fitness functions (testes de arquitetura)

Baseado no livro "Building Evolutionary Architectures" (Ford, Parsons, Kua).
Fitness functions são **testes automatizados que validam propriedades arquiteturais**,
não comportamento funcional.

### C# — NetArchTest.Rules

13 testes que validam Clean Architecture no projeto C#:

| # | Regra | Motivação |
|---|-------|-----------|
| 1 | Domain ≠ Application | Inversão de dependência |
| 2 | Domain ≠ Infrastructure | Domain puro |
| 3 | Domain ≠ Api | Domain não conhece apresentação |
| 4 | Domain ≠ EF Core | Persistence ignorance |
| 5 | Application ≠ Infrastructure | Use cases não conhecem infra |
| 6 | Application ≠ Api | Application não conhece HTTP |
| 7 | Infrastructure ≠ Api | Infra não conhece apresentação |
| 8 | Entities são sealed | Previne herança acidental |
| 9 | Entities implementam IOwnedEntity | Multi-tenancy obrigatória |
| 10 | Interfaces começam com "I" | Naming convention |
| 11 | Interfaces em Domain.Interfaces | Organização |
| 12 | ValueObjects são records | Imutabilidade garantida |
| 13 | Domain Services são static | Stateless |

Rodar: `dotnet test --filter Architecture`

### Go — arch-go

9 regras de dependência entre pacotes Go, configuradas em `arch-go.yml`.

### Scripts de drift (cross-stack)

3 scripts que detectam inconsistências entre os diferentes stacks:

#### `scripts/check-api-contract.sh`

Extrai todas as chamadas `/api/*` do frontend (`web/src/lib/api.js`) e verifica
que cada rota tem um endpoint correspondente no C# (`src/Garimpei.Api/Endpoints/`).

**Detecta:** endpoint adicionado no frontend sem implementação no backend (ou rota
removida do backend que o frontend ainda usa).

#### `scripts/check-config-consistency.sh`

Verifica consistência de configurações compartilhadas entre stacks:

- Nome do dataset BigQuery (`garimpo`, nunca `garimpei`)
- Portas de serviços gRPC (collector=50051, publisher=50052)
- URL do analyzer (porta 8060)
- Tabelas BigQuery documentadas no schema
- URLs de produção hardcoded no frontend

**Detecta:** alguém introduz um typo no nome do dataset, muda uma porta num lugar
e esquece no outro.

#### `scripts/check-schema-sync.sh`

Verifica sincronização de schemas entre os 3 datastores e os componentes:

1. **BigQuery SQL ↔ Go EnsureSchema**: toda tabela criada pelo Go deve estar no SQL schema
2. **Analyzer Python ↔ BigQuery**: tabelas consultadas devem existir no schema
3. **Entities C# ↔ DbSets**: toda entidade deve ter DbSet (e vice-versa)
4. **IOwnedEntity ↔ QueryFilter**: toda entidade multi-tenant deve ter filtro

**Decisão superset/subset:**
> SQL schema (`deploy/bigquery_schema.sql`) = **superset** (fonte de verdade)
> Go `EnsureSchema` = **subset** (só cria tabelas que os microserviços gerenciam)
> Tabelas externas (ex.: `conversoes` via webhook) existem apenas no SQL schema.

---

## Análise estática

| Ferramenta | Stack | O que valida |
|-----------|-------|-------------|
| golangci-lint | Go | 50+ linters (estilo, bugs, performance) |
| arch-go | Go | Dependências entre pacotes (9 regras) |
| buf lint | Proto | Protos seguem STANDARD rules |
| buf breaking | Proto | Breaking changes nos contratos |
| dotnet build (TreatWarningsAsErrors) | C# | Zero warnings |
| NetArchTest | C# | 13 regras de Clean Architecture |
| ruff | Python | Lint rápido (compatible com flake8) |
| eslint | Frontend | Qualidade JS/Svelte |
| stylelint | Frontend | Qualidade CSS |
| Knip | Frontend | Dead code/exports |
| check-file-size | All | Máx 400 linhas por arquivo |

---

## Métricas de qualidade

| Métrica | Alvo | Validação |
|---------|------|-----------|
| Testes Go | ~200 | `go test ./...` |
| Testes C# | 23 (10 persistence + 13 arch) | `dotnet test` |
| Testes frontend | 109 | `vitest --run` (<2s) |
| Drift API | 0 rotas faltantes | `check-api-contract.sh` |
| Drift config | 0 inconsistências | `check-config-consistency.sh` |
| Drift schema | 0 desincronizações | `check-schema-sync.sh` |
| Arquivos > 400 linhas | 0 | CI bloqueia |
| Warnings C# | 0 | TreatWarningsAsErrors |

---

## Como rodar os testes localmente

```bash
# Go
go test ./...

# C#
cd src && dotnet test

# Frontend
cd web && npx vitest run

# Scripts de drift (sem dependências externas)
./scripts/check-api-contract.sh
./scripts/check-config-consistency.sh
./scripts/check-schema-sync.sh
```

---

## Como adicionar uma nova entidade (checklist)

Quando criar uma nova entidade no projeto C#, garantir que:

1. ☐ Arquivo em `src/Garimpei.Domain/Entities/NomeEntity.cs`
2. ☐ Implementa `IOwnedEntity` (se for multi-tenant)
3. ☐ É `sealed class`
4. ☐ DbSet adicionado no `AppDbContext`
5. ☐ `HasQueryFilter` configurado no `OnModelCreating` (se IOwnedEntity)
6. ☐ Migration criada (`dotnet ef migrations add NomeMigration`)
7. ☐ Endpoint criado em `Endpoints/` (se o frontend consumir)
8. ☐ Rota adicionada no `Program.cs`
9. ☐ `dotnet test` passa (inclui fitness functions)
10. ☐ `check-schema-sync.sh` passa

Se a entidade também precisa de tabela no BigQuery:

11. ☐ Tabela adicionada em `deploy/bigquery_schema.sql`
12. ☐ Se Go gerencia: adicionada em `internal/store/bigquery_schema.go`
13. ☐ Se analyzer consulta: rota adicionada em `services/analyzer/routes/`
14. ☐ `check-schema-sync.sh` passa
