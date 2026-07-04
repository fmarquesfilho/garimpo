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
  ├─ frontend: npm ci + build + lint:css + lint:js + vitest + playwright (Firebase Emulator)
  ├─ api-contract: check-api-contract + check-config-consistency + check-schema-sync
  ├─ docker: build all 5 images (validação de Dockerfiles)
  ├─ deploy-web: wrangler pages deploy → Cloudflare Pages (só push main)
  └─ deploy-docs: sync + build + deploy → Cloudflare Pages (só push main, se docs mudam)
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
| C# (Integração) | xUnit | 15 | Onboarding multi-tenant end-to-end |
| Frontend (unit) | Vitest | 108 | Componentes, stores, utils, lógica filtros |
| Frontend (E2E) | Playwright | 24 | Smoke + Descobrir (filtros, fontes, badges) |
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

#### `.mise/tasks/check/api-contract`

Extrai todas as chamadas `/api/*` do frontend (`web/src/lib/api.js`) e verifica
que cada rota tem um endpoint correspondente no C# (`src/Garimpei.Api/Endpoints/`).

**Detecta:** endpoint adicionado no frontend sem implementação no backend (ou rota
removida do backend que o frontend ainda usa).

#### `.mise/tasks/check/config-consistency`

Verifica consistência de configurações compartilhadas entre stacks:

- Nome do dataset BigQuery (`garimpo`, nunca `garimpei`)
- Portas de serviços gRPC (collector=50051, publisher=50052)
- URL do analyzer (porta 8060)
- Tabelas BigQuery documentadas no schema
- URLs de produção hardcoded no frontend

**Detecta:** alguém introduz um typo no nome do dataset, muda uma porta num lugar
e esquece no outro.

#### `.mise/tasks/check/schema-sync`

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
| Testes C# | 38 (10 persistence + 13 arch + 15 integration) | `dotnet test` |
| Testes frontend | 109 | `vitest --run` (<2s) |
| Drift API | 0 rotas faltantes | `check-api-contract.sh` |
| Drift config | 0 inconsistências | `check-config-consistency.sh` |
| Drift schema | 0 desincronizações | `check-schema-sync.sh` |
| Pre-push | 7/7 checks | `pre-push-check.sh` |
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
./.mise/tasks/check/api-contract
./.mise/tasks/check/config-consistency
./.mise/tasks/check/schema-sync

# TUDO de uma vez (mesmo script usado pelo pre-push hook)
./scripts/pre-push-check.sh
```

---

## Pre-push hook (gate local obrigatório)

O projeto inclui um **git pre-push hook** que bloqueia push se qualquer check falhar.
Roda automaticamente antes de cada `git push`:

```
═══════════════════════════════════════════════════════════════
  Pre-push: verificação completa antes de enviar
═══════════════════════════════════════════════════════════════

🔨 C# (build + testes + arquitetura):
  [1] Build... ✓
  [2] Testes (38: persistence + architecture + integration)... ✓

🔨 Go (build + testes):
  [3] Build... ✓
  [4] Testes... ✓

🔍 Drift checks (cross-stack):
  [5] API contract (frontend↔backend)... ✓
  [6] Config consistency (dataset, portas)... ✓
  [7] Schema sync (BQ↔Go↔C#↔Analyzer)... ✓

═══════════════════════════════════════════════════════════════
✅ 7/7 checks passaram. Push liberado.
```

### Instalar

```bash
ln -sf ../../scripts/pre-push-check.sh .git/hooks/pre-push
```

### Quando roda

- **Automaticamente** antes de cada `git push` (bloqueia se falhar)
- **Manualmente** via `./scripts/pre-push-check.sh` (pra validar antes de commitar)

### Estado da arte

Esta abordagem combina 3 técnicas complementares:

| Técnica | Quando roda | O que garante |
|---------|-------------|--------------|
| **Pre-push hook** (local) | Antes do push | Feedback rápido, bloqueia código quebrado |
| **CI (GitHub Actions)** | Após push/PR | Validação em ambiente limpo, Docker builds |
| **Fitness functions** (in-code) | `dotnet test` | Regras arquiteturais como testes |

O pre-push hook é a **primeira linha de defesa**: impede que código quebrado chegue
ao remote. O CI é a segunda (valida em ambiente reprodutível). As fitness functions
são a terceira (regras vivas dentro do código).

### Bypassar (emergência)

```bash
git push --no-verify   # Pula o hook (use com responsabilidade)
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

---

## Gerenciamento de dependências

### Política

Dependências devem ser mantidas sempre atualizadas. Ferramentas com versões
defasadas representam risco de segurança e acumulam dívida técnica que cresce
exponencialmente com o tempo.

**Princípios:**
- Atualizações (patch, minor, major) são auto-merged se o CI passa
- O CI rigoroso (13+ checks, fitness functions, drift checks) é a barreira de segurança
- Se um bump major quebra o CI, o PR fica aberto para intervenção manual
- Vulnerabilidades são priorizadas e auto-merged imediatamente

### Renovate (automação)

O repositório usa [Renovate](https://docs.renovatebot.com/) para monitoramento
automático de dependências. Configuração em `renovate.json`.

| Tipo de update | Ação | Frequência |
|---|---|---|
| Patch (4.1.9→4.1.10) | Auto-merge se CI verde | Semanal (segundas) |
| Minor (4.1→4.2) | Auto-merge se CI verde | Semanal (segundas) |
| Major (4→5) | Auto-merge se CI verde | Semanal (segundas) |
| Vulnerabilidade (CVE) | Auto-merge + label `security` | Imediato |

**Agrupamento por workspace:**

| Workspace | PR | Schedule |
|---|---|---|
| `web/` (frontend) | 1 PR agrupado | Semanal |
| `docs-site/` | 1 PR agrupado | Mensal |
| Go (`go.mod`) | 1 PR agrupado | Semanal |
| GitHub Actions | 1 PR agrupado | Semanal |

**Pacotes ignorados:** `@astrojs/*`, `astro`, `sharp` no docs-site (migrado para Rspress).

### Segurança complementar

| Ferramenta | Função | Ação |
|---|---|---|
| **Renovate** | Detecta + corrige (abre PR com fix) | Auto-merge |
| **Codacy/Trivy** | Detecta + reporta (scan sem fix) | Alerta |
| **npm audit** | Reporta vulnerabilidades JS | Manual |
| **go vuln** | Reporta vulnerabilidades Go | Manual |

Renovate corrige, Codacy/Trivy detecta — camadas complementares.

### Como atualizar manualmente

```bash
# Frontend
cd web && npm update && npm run build && npx vitest run

# Go
go get -u ./... && go mod tidy && go test ./...

# Verificar outdated
cd web && npm outdated
go list -m -u all | grep '\['
```
