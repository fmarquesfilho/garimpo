# Handoff — Rules Service + BuscaEngine fixes (2026-07-09)

> ✅ **CONCLUÍDO / DESCARTADO.** A abordagem rules-service não será implementada. Regras ficam em JSON no frontend (ADR-0027).

> Documento de passagem para a próxima sessão. Descreve o estado completo, o que
> foi feito, o que falta, decisões de arquitetura a respeitar, e como verificar.

## Branch

`claude/monitored-stores-refactor-xazzf9` (base = `origin/main`)

## Commits desta sessão (cronológicos)

```
bab81ec refactor(busca): regras da engine em config declarativa versionada
4ed309b test(e2e): harness local com bypass de auth
be8d5d0 fix(busca): escopo de loja (#2) + round-trip de keywords no backend (#6/#7)
346dd96 docs: handoff da página Garimpar/BuscaEngine + arquitetura em componentes.md
f80c7db fix: categorias como seletor (#4), sync store após salvar (#8), label salvar (#10)
7b65558 spec: rules-service (Go + zen-go + JDM versionado)
bc41aed docs: handoff sessão rules-service + estado completo da branch
f9dbf6d feat: regras de busca como JSON externo + E2E validando contra rules
159cff0 docs: sessão regras externas + atualizar handoff com decisão final
78c537f test: expandir suite — 35 unit tests (cenários doc) + 15 E2E locais
5dbb7f8 docs(adr): 0027 — BuscaEngine headless + regras externas como JSON
d8f03a1 fix(e2e): corrigir mocks para refletir fluxo real da app
1f7bdd4 fix: corrigir 3 bugs reais da engine (store desync, lojas ctx, erro engolido)
cd35bf7 docs(adr): atualizar 0027 com fixes reais + fluxo de validação regras→UI
```

## ✅ Concluído nesta sessão

### Bugs corrigidos (UI)
- **#2** — Adicionar loja agora escopa busca (keyword + shopIds combinados)
- **#4** — Categorias como chips selecionáveis (de categoriasDisponiveis), não TagInput livre
- **#6/#7** — Round-trip de keywords no backend (POST /api/buscas persiste b.Keywords)
- **#8** — Sync store `buscasSalvas` após salvar/remover (sincronizarStoreExterno)
- **#10** — Botão salvar com label "💾 Salvar" (não apenas ícone)

### Bugs corrigidos (arquitetura)
- **Store desync** — `INICIALIZAR` agora chama `sincronizarStoreExterno()` antes de `executarBusca`
- **Lojas ctx ignoradas** — `buildBuscasComLojas()` combina store + ctx.shopIds (lojas não salvas participam de quedas/novos)
- **Erro engolido** — `carregarCuradoria` propaga erros HTTP ≥400 via `isServerError()` (antes mostrava "0 resultados")

### Infraestrutura
- **Config declarativa** (`busca-config.js`) — defaults, normalização, guards, INTENT_TABLE
- **Harness E2E local** — bypass auth via `window.__E2E_AUTH_USER__`, 3/3 testes passando
- **Spec rules-service** — design + requirements + tasks completos para Go + zen-go

### Métricas
- 243 unit tests passando
- 24 E2E locais passando (0 skips)
- svelte-check 0/0, eslint 0 warnings, build ok
- Backend: precisa validar no CI (dotnet não disponível localmente)

## ⬜ Pendente — Rules Service (spec em `.kiro/specs/rules-service/`)

> **DECISÃO:** O rules-service como sidecar gRPC foi descartado por over-engineering.
> Substituído por `rules/busca-rules.json` (JSON declarativo externo).
> Ver `docs/legado/SESSAO_2026-07-09_REGRAS_EXTERNAS.md` para detalhes.
>
> O spec permanece em `.kiro/specs/rules-service/` como referência futura caso a
> complexidade justifique (multi-tenant com regras por cliente, operadores não-dev).

### ✅ Implementado (abordagem simplificada)

- `rules/busca-rules.json` — fonte de verdade (intent, guards, normalize, defaults)
- `rules/busca-rules.schema.json` — JSON Schema
- `busca-config.js` importa do JSON externo
- `.mise/tasks/check/rules-schema` — drift check no CI
- E2E tests validam UI contra o JSON (`web/tests/local/busca-rules.spec.js`)
- Documentação completa atualizada

## Decisões de arquitetura (RESPEITAR)

1. **Engine headless (classe Svelte 5).** `BuscaEngine` em `busca-engine.svelte.js` é testável com `new BuscaEngine(mockEffects())`. View (`BuscaUnificada.svelte`) é burra.

2. **Regras como JSON externo.** `rules/busca-rules.json` é a fonte de verdade. Frontend importa em build-time. E2E testam contra o JSON. Sem engine externo, sem sidecar.

3. **Sem rules-service sidecar.** Complexidade desproporcional para 4 intents e 2 guards. Se no futuro multi-tenant com regras por cliente justificar, o spec em `.kiro/specs/rules-service/` serve de referência.

4. **JDM descartado. JSON puro.** O formato GoRules JDM não é necessário — JSON com dados puros é suficiente e legível por qualquer linguagem.

5. **Config (`busca-config.js`) permanece como adapter.** Importa do JSON externo e re-exporta no formato da engine. Funções puras (normalização, guards, intent) operam sobre os dados importados.

6. **Ports & Adapters.** Effects são injetáveis. Mock para testes. Real para produção.

7. **CI valida regras.** `mise run check:rules-schema` verifica completude da intent table, consistência dos guards, e formato do JSON.

## Como verificar (no ambiente dev)

```bash
# Frontend (tudo deve passar):
cd web && npm run check && npm run lint:js && npm run format:check && npx vitest run && npm run build

# E2E local (3 specs originais + 5 rules):
cd web && npm run test:e2e:local

# Drift check rules:
.mise/tasks/check/rules-schema

# Go (sem regressão):
go build ./... && go test ./...

# Backend C# (requer Docker + dotnet):
cd src && dotnet build && dotnet test

# Tudo junto (pre-push):
mise run prepush
```

## Armadilhas do ambiente

- `dotnet` e E2E com Firebase Emulator dependem do CI (não disponíveis localmente em todos os ambientes)
- `pkill -f "vite preview"` pode abortar o comando (exit 144) — evitar
- Duas fontes de buscas salvas: store `buscasSalvas` E `engine.ctx.buscasSalvas` — fix #8 sincroniza ambas via `sincronizarStoreExterno()`
- Cache de 2min em `carregarOportunidades`/`carregarProdutosLojas` (`descobrir.js`) pode mascarar dados novos
- ESLint: `.svelte.js` tem limit 150 linhas/função (não 80)

## Arquivos-chave

| Arquivo | Papel |
|---------|-------|
| `rules/busca-rules.json` | **Fonte de verdade** — regras externas (intent, guards, normalize, defaults, transições) |
| `rules/busca-rules.schema.json` | JSON Schema para validação |
| `web/src/lib/busca-engine.svelte.js` | FSM (classe Svelte 5 com $state) |
| `web/src/lib/busca-engine-effects.js` | API calls injetáveis |
| `web/src/lib/busca-config.js` | Adapter: importa JSON externo → re-exporta para engine |
| `web/src/lib/components/BuscaUnificada.svelte` | View pura |
| `web/src/lib/busca-unificada-logic.js` | Funções puras (payload, labels) |
| `web/src/lib/descobrir-logic.js` | montarResultados (filtro client-side) |
| `web/src/lib/descobrir.js` | Orquestração de fontes (fetch) |
| `.mise/tasks/check/rules-schema` | Drift check para regras |
| `web/tests/local/busca-rules.spec.js` | E2E contra regras externas |
| `web/tests/local/garimpar.spec.js` | E2E harness original |
| `.kiro/specs/rules-service/` | Spec original (referência futura, não implementar) |
