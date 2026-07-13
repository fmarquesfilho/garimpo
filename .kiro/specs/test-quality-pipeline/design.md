# Design: Test Quality Pipeline

## Overview

Pipeline de qualidade de testes para o Garimpei, com execução via mise tasks, mutation testing (Stryker), coverage granular, e teste de paridade cross-language. Otimiza o feedback loop para a página Descobrir e o app inteiro.

## Arquitetura

```
┌─────────────────────────────────────────────────────────────────────┐
│ mise tasks (orquestrador)                                           │
│                                                                     │
│  test:fast ──────→ vitest run (unit only) + dotnet test (arch)      │
│  test:unit ──────→ vitest run + dotnet test                         │
│  test:integration → vitest (*.integration.test.js) + dotnet (Integ) │
│  test:regression ─→ vitest (*.regression.test.js)                   │
│  test:parity ────→ vitest (normalizacao) + dotnet (LojaTests)       │
│  test:coverage ──→ vitest --coverage (v8) + thresholds              │
│  test:mutate ────→ StrykerJS (vitest runner, modules core)          │
│  test:mutate-backend → Stryker.NET (Domain project)                 │
│  test:descobrir ─→ vitest (busca-engine-*, omnibox-*, busca-config) │
│  test:pre-push ──→ unit + integration + regression + parity + check │
│  test:ci ────────→ pre-push + coverage                              │
│  test:full ──────→ ci + e2e-local                                   │
│  test:e2e-local ─→ playwright (mocks, sem backend)                  │
│  test:e2e-prod ──→ playwright (API real, manual only)               │
└─────────────────────────────────────────────────────────────────────┘
```

## Classificação de Testes Existentes

### Frontend (Vitest — 419 testes)

| Arquivo | Camada | Justificativa |
|---------|--------|---------------|
| busca-config.test.js | unit | Funções puras de normalização e config |
| busca-duplicata.test.js | unit | Fingerprint e comparação |
| busca-engine.test.js | integration | Engine com mocks de effects |
| busca-engine-cenarios.test.js | integration | Cenários E2E da engine |
| busca-engine-omnibox.test.js | integration | Engine OMNIBOX_* handlers |
| busca-unificada.test.js | integration | Componente + engine |
| omnibox.test.js | integration | Componente + engine mock |
| omnibox-parser.test.js | unit | Parser puro (tokenização) |
| omnibox-sugestoes.test.js | unit | Gerador de sugestões puro |
| omnibox-intencao.test.js | unit | Detecção de intenção puro |
| descobrir.test.js | unit | montarResultados puro |
| descobrir-busca-id.test.js | unit | Derivação de busca ID |
| CandidateCard.test.js | integration | Componente DOM |
| SeletorGrupo.test.js | integration | Componente DOM |
| fixtures-contract.test.js | regression | Contrato de fixtures |
| loading-timeout.test.js | regression | Bug de timeout |
| contrast.test.js | unit | Contraste CSS |
| theme.test.js | unit | Theme tokens |
| favoritos.test.js | integration | Store + API mock |
| oportunidades.test.js | unit | Lógica de oportunidades |
| publicar.test.js | integration | Flow de publicação |

### Backend (xUnit — 94 testes)

| Namespace/Classe | Camada | Justificativa |
|------------------|--------|---------------|
| Architecture/ | unit | Arch rules (NetArchTest) |
| Domain/Entities/LojaTests | unit | Normalizar puro |
| Persistence/MultiTenantQueryFilter | integration | EF Core InMemory |
| Integration/LojasBuscarEndpoint | integration | Query patterns |
| Integration/BuscasAgendadas | integration | Scheduler logic |
| Integration/JsonBinding | integration | Serialização |
| Integration/OnboardingMultiTenant | integration | Multi-tenant flow |
| Integration/PublishFlow | integration | Publicação flow |
| Services/CouponDeduplication | unit | Dedup logic |
| Tenancy/ | unit | TenantContext |
| CollectionKeysTests | unit | Derivação determin. |
| SchedulerJobsTests | unit | Job registration |

### Estratégia de Classificação (sem rename obrigatório)

Em vez de renomear 419 arquivos, a classificação é feita por **glob pattern** no vitest.config.js e nos tasks mise:

```javascript
// vitest.config.js — workspaces por camada
export default defineConfig({
  test: {
    // Default: roda tudo
    include: ['src/tests/**/*.test.js'],
    // Projects por camada (filtráveis via --project)
    workspace: [
      { name: 'unit', include: ['src/tests/{busca-config,busca-duplicata,omnibox-parser,omnibox-sugestoes,omnibox-intencao,descobrir,descobrir-busca-id,contrast,theme,oportunidades}.test.js'] },
      { name: 'integration', include: ['src/tests/{busca-engine,busca-engine-*,busca-unificada,omnibox,CandidateCard,SeletorGrupo,favoritos,publicar}.test.js'] },
      { name: 'regression', include: ['src/tests/{fixtures-contract,loading-timeout}.test.js'] }
    ]
  }
});
```

Para o backend, o filtro é por namespace:
```bash
dotnet test --filter "FullyQualifiedName~Architecture|FullyQualifiedName~Domain|FullyQualifiedName~CollectionKeys|FullyQualifiedName~Scheduler|FullyQualifiedName~Tenancy|FullyQualifiedName~Services"  # unit
dotnet test --filter "FullyQualifiedName~Integration|FullyQualifiedName~Persistence"  # integration
```

## Mutation Testing

### StrykerJS (Frontend)

```json
// stryker.config.json
{
  "testRunner": "vitest",
  "vitest": { "configFile": "vitest.config.js" },
  "mutate": [
    "src/lib/busca-engine.svelte.js",
    "src/lib/busca-engine-state.js",
    "src/lib/busca-config.js",
    "src/lib/omnibox-intencao.js",
    "src/lib/omnibox-parser.js",
    "src/lib/omnibox-sugestoes.js",
    "src/lib/loja-registry.js"
  ],
  "thresholds": { "high": 80, "low": 70, "break": 60 },
  "reporters": ["html", "clear-text", "progress"],
  "htmlReporter": { "fileName": "../reports/mutation/index.html" },
  "incremental": true
}
```

### Stryker.NET (Backend)

```json
// stryker-config.json (na raiz de src/)
{
  "stryker-config": {
    "project": "Garimpei.Domain/Garimpei.Domain.csproj",
    "test-projects": ["Garimpei.Tests/Garimpei.Tests.csproj"],
    "reporters": ["html", "progress"],
    "thresholds": { "high": 80, "low": 70, "break": 60 },
    "mutation-level": "Standard"
  }
}
```

## Coverage com Thresholds

```javascript
// vitest.config.js (bloco coverage)
coverage: {
  provider: 'v8',
  reporter: ['text', 'html', 'lcov'],
  reportsDirectory: '../reports/coverage',
  thresholds: {
    // Core modules
    'src/lib/busca-engine.svelte.js': { lines: 80, branches: 75 },
    'src/lib/busca-config.js': { lines: 85, branches: 80 },
    'src/lib/omnibox-intencao.js': { lines: 85, branches: 80 },
    'src/lib/omnibox-parser.js': { lines: 90, branches: 85 },
    'src/lib/loja-registry.js': { lines: 80, branches: 75 },
    // Components (lower threshold)
    'src/lib/components/**': { lines: 60, branches: 50 }
  }
}
```

## Paridade Cross-Language

O fixture `fixtures/normalizacao-pares.json` é a fonte da verdade:

```json
[
  {"input": "Glory of Seoul", "expected": "gloryofseoul"},
  {"input": "Café Bonito 🇧🇷", "expected": "cafebonito"},
  {"input": "  ", "expected": ""},
  {"input": "SKIN1004", "expected": "skin1004"},
  {"input": "日本語", "expected": ""},
  {"input": "résumé", "expected": "resume"},
  {"input": "123 Shop!", "expected": "123shop"}
]
```

Ambos os testes leem o mesmo arquivo:
- C#: `LojaTests.Normalizar_DeveProcessarParesParametrizadosCorretamente` (já existe)
- JS: `normalizacao-parity.test.js` (a criar) importa `normalizarNome` de `loja-registry.js`

## Tasks mise — Implementação

```
.mise/tasks/test/
├── fast          # unit only (<10s)
├── unit          # [existente, atualizar] unit front + back
├── integration   # integration front + back
├── regression    # regression front
├── parity        # cross-language normalização
├── coverage      # vitest --coverage
├── mutate        # StrykerJS (front, módulos core)
├── mutate-backend # Stryker.NET (Domain)
├── descobrir     # apenas testes da página Descobrir
├── pre-push      # unit + integration + regression + parity + check + build
├── ci            # pre-push + coverage (GitHub Actions)
├── full          # ci + e2e-local
├── all           # [existente, atualizar] = full
├── e2e-local     # [existente] playwright local
└── e2e-prod      # [existente] playwright prod (manual)
```

## Ferramenta de Qualidade: Stryker Mutator

**Por que Stryker em 2026:**
- Único framework de mutation testing com suporte nativo a Vitest + Svelte ([stryker-mutator.io](https://stryker-mutator.io/docs/stryker-js/guides/svelte/))
- Stryker.NET é recomendado pela [Microsoft Learn](https://learn.microsoft.com/en-za/dotnet/core/testing/mutation-testing) como ferramenta oficial para .NET
- Suporta execução incremental (só muta arquivos alterados)
- Gera relatório HTML interativo com cada mutante e seu status
- Target: ≥70% mutation score para módulos core (threshold quebra build se cair abaixo de 60%)

**Alternativas descartadas:**
- Coverage-only (@vitest/coverage-v8): necessário mas insuficiente — 100% coverage pode ter 4% mutation score (paper arxiv)
- PIT: JVM-only (não aplicável)
- mutmut: Python-only

## Decisões de Design

| # | Decisão | Justificativa |
|---|---------|---------------|
| 1 | Classificação por glob (não rename) | Retrocompatível, zero breaking change. Patterns vitest workspace filtra sem mover arquivos. |
| 2 | Stryker (não alternativa) | Único com suporte Vitest + Svelte + .NET. Microsoft-endorsed. |
| 3 | Mutation incremental (default) | Stryker completo pode levar minutos. Incremental permite uso diário. |
| 4 | Fixture compartilhado C#/JS | Fonte única de verdade. Adicionar par → ambos detectam. |
| 5 | Thresholds por módulo (não global) | Módulos core exigem mais que utilitários. Evita threshold baixo global. |
| 6 | CI roda mutation apenas no merge | Pesado para todo PR. Label opt-in permite rodar quando necessário. |
| 7 | `test:descobrir` dedicado | Feedback loop rápido para a feature central. Dev não precisa rodar 419 testes. |
| 8 | mise tasks (não scripts npm) | Orquestra front+back+cross-language em um só lugar. Composable via depends. |

## Impacto em Arquivos

| Arquivo | Mudança |
|---------|---------|
| `.mise/tasks/test/fast` | **NOVO** — unit only |
| `.mise/tasks/test/unit` | **ATUALIZAR** — adicionar filtro xUnit unit |
| `.mise/tasks/test/integration` | **NOVO** |
| `.mise/tasks/test/regression` | **NOVO** |
| `.mise/tasks/test/parity` | **NOVO** |
| `.mise/tasks/test/coverage` | **NOVO** |
| `.mise/tasks/test/mutate` | **NOVO** |
| `.mise/tasks/test/mutate-backend` | **NOVO** |
| `.mise/tasks/test/descobrir` | **NOVO** |
| `.mise/tasks/test/pre-push` | **NOVO** |
| `.mise/tasks/test/ci` | **NOVO** |
| `.mise/tasks/test/full` | **NOVO** |
| `web/vitest.config.js` | Expandir com workspace projects + coverage thresholds |
| `web/stryker.config.json` | **NOVO** — config StrykerJS |
| `src/stryker-config.json` | **NOVO** — config Stryker.NET |
| `web/package.json` | Adicionar devDeps: `@vitest/coverage-v8`, `@stryker-mutator/core`, `@stryker-mutator/vitest-runner` |
| `fixtures/normalizacao-pares.json` | Expandir com edge cases (emoji, CJK, whitespace) |
| `web/src/tests/normalizacao-parity.test.js` | **NOVO** — teste JS contra fixture |
| `.github/workflows/ci.yml` | Adicionar step mutation (condicional) |
| `.gitignore` | Adicionar `reports/` |
