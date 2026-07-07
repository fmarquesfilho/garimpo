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
  ├─ contracts: service-contracts + api-contract + config-consistency + schema-sync + data-ownership
  ├─ security: Semgrep SAST (JavaScript + TypeScript)
  ├─ deploy-backend: build Docker (5 imgs) + migrations + Cloud Run [se backend mudou]
  └─ deploy-web: build + Cloudflare Pages [se frontend mudou]
```

**Otimizações de path filtering:**

Pushes que tocam apenas estes caminhos **não disparam CI**:
- `docs/`, `docs-site/`, `backlog/`, `.kiro/`, `.vscode/`
- `*.md`, `LICENSE`, `.gitignore`, `.codacy.yml`, `.semgrepignore`, `renovate.json`

**Deploy condicional (monorepo-aware):**
- `deploy-backend` detecta via `git diff HEAD~1` se houve mudança em `src/`, `services/`, `protos/`, `deploy/`, `go.*`, `internal/`, ou `contracts/registry`. Se não, **pula o deploy** (~4min economizados).
- `deploy-web` roda apenas quando o frontend ou contratos mudam.

---

## Estratégia de testes

### Cobertura por camada

| Camada | Ferramenta | Testes | Foco |
|--------|-----------|--------|------|
| Go (internal) | go test | ~200 | source 87%, publish 62%, store 36% |
| Go (services) | go test | 12 | Validações + fluxos gRPC |
| C# (Domain + Infra) | xUnit | 10 | Multi-tenant, persistence, isolation |
| C# (Arquitetura) | xUnit + NetArchTest | 13 | Fitness functions (regras Clean Architecture) |
| C# (Integração) | xUnit | 38 | Onboarding, JSON binding, dedup, publish flow |
| Frontend (unit) | Vitest | ~109 | Componentes, stores, utils, lógica filtros |
| Frontend (E2E) | Playwright | ~36 | Smoke + Descobrir + Lojas/Preços + ResolveShop |
| Cross-stack (drift) | Shell scripts (mise) | 7 | Contracts, ownership, stale refs, schema sync |

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
| Testes C# | 61 (multi-tenant + arch + integration + JSON) | `dotnet test` |
| Testes frontend | ~109 unitários + ~36 E2E | `vitest --run` + `playwright test` |
| Drift API | 0 rotas faltantes | `mise run check:api-contract` |
| Drift config | 0 inconsistências | `mise run check:config-consistency` |
| Drift schema | 0 desincronizações | `mise run check:schema-sync` |
| Pre-push | 9/9 checks | `mise run prepush` |
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

## Testes opcionais (não rodam no CI)

Testes que requerem serviços externos ou são informativos (não bloqueiam push):

### E2E completo (frontend + Firebase Emulator)

Testa fluxos autenticados no browser com Playwright + Firebase Auth Emulator:

```bash
mise run test:e2e
```

**Requer:** Firebase CLI instalada (`npm i -g firebase-tools`).
**Cenários:** login → descobrir → filtros → lojas → preços → publicar.

### E2E ResolveShop (integração real com Collector + Shopee)

Testa o fluxo de adição de loja **sem mocks** — valida a comunicação real entre
Frontend → C# API → Collector gRPC → Shopee API v4 → PostgreSQL.

```bash
# Pré-requisitos: API + Collector + Firebase Emulator rodando
mise run up                                    # sobe API + PG + Analyzer
cd services/collector && go run . &            # sobe collector na porta 50051
FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 npx playwright test lojas-resolve-shop
```

**Arquivo:** `web/tests/lojas-resolve-shop.spec.js`

**Cenários testados:**

| # | Cenário | Input | Validação |
|---|---------|-------|-----------|
| 1 | Link direto `shopee.com.br/{username}` | `https://shopee.com.br/belezanaweb_oficial` | shop_id=1674883556, name="Beleza na Web Oficial" |
| 2 | Link curto `s.shopee.com.br/{hash}` | `https://s.shopee.com.br/70IKp57jnV` | Segue redirect → shop_id=920292999, name="Glory of Seoul" |
| 3 | Username puro | `gloryofseoul.br` | shop_id=920292999 |
| 4 | Listagem com shop_ids | GET /api/lojas | Lojas resolvidas aparecem com `shop_ids[]` preenchido |
| 5 | Link inválido | `loja_que_nao_existe_xyz_999` | HTTP 400 com mensagem "não encontrada" |

**Data ownership validada no teste:**
- C# API persiste em PostgreSQL (dono exclusivo do PG) ✅
- Collector Go faz I/O externo (Shopee API v4) sem tocar PG ✅
- Comunicação via gRPC (contrato tipado) ✅

**Nota:** Este teste depende da API pública da Shopee e pode falhar se houver rate
limiting ou instabilidade na API. Por isso não roda no CI por padrão — é um teste
de integração manual para validação pré-deploy.

#### Boas Práticas e Resolução de Problemas E2E

1. **Paridade de Mocks da API:**
   Ao criar testes E2E que não dependem do backend real (mockando chamadas de rede), certifique-se de que a rota interceptada pelo Playwright (`page.route()`) corresponde **exatamente** à URL chamada pelo `$lib/api.js`. Por exemplo, o frontend pode listar lojas usando `GET /api/buscas`, mas adicionar usando `POST /api/lojas`. Mocar a rota incorreta causará timeouts no frontend.
2. **Estabilidade de Seletores:**
   Devido à adoção do `shadcn-svelte` e ao estilo utilitário do Tailwind, **não use classes CSS (ex: `.marca`, `.hero-features`)** como seletores no Playwright. 
   Prefira seletores resilientes a mudanças de estilo:
   - Atributos ARIA: `page.locator('[aria-label="Abrir menu"]')`
   - Texto explícito: `page.locator('text=Garimpei')`
   - Tags semânticas contextuais: `page.locator('h1')`
3. **Timeouts Inesperados:**
   Falhas de timeout no login de teste (`typeof window.__TEST_SIGN_IN__`) frequentemente indicam que a página travou antes do Firebase carregar (ex: o `npm run preview` devolveu um erro HTTP 500 no console JS ou a porta do emulator Firebase, 9099, já estava em uso por um processo zumbi).

### Qualidade de comentários

Detecta anti-patterns em comentários: código morto comentado, TODOs sem issue,
comentários óbvios, comentários de 1 palavra. Cross-stack (Go, C#, Python, Svelte).

```bash
mise run check:comment-quality              # arquivos alterados vs main
mise run check:comment-quality -- --all     # todos os arquivos
mise run check:comment-quality -- --strict  # exit 1 se houver warnings
```

**Anti-patterns detectados:**
- Blocos de 3+ linhas de código comentado (dead code)
- `TODO`/`FIXME` sem referência a task (`T-NNNN`) ou issue
- Comentários triviais que repetem o que o código diz
- Comentários de 1 palavra em funções (sem contexto útil)

**Não bloqueia push** — é informativo. Use `--strict` se quiser enforcement.

### Teste de alertas (Telegram real)

Testa o fluxo de alertas com envio real para Telegram:

```bash
./scripts/test-alerts.sh local     # lógica sem envio (analyzer mock)
./scripts/test-alerts.sh telegram  # envio real (requer tokens)
./scripts/test-alerts.sh prod      # Cloud Task em produção
```

**Requer:** `TELEGRAM_BOT_TOKEN` (de `gcloud secrets`) para modo telegram/prod.

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
