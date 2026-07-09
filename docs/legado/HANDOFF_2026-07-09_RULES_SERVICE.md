# Handoff — Rules Service + BuscaEngine fixes (2026-07-09)

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
```

## ✅ Concluído nesta sessão

### Bugs corrigidos
- **#2** — Adicionar loja agora escopa busca (keyword + shopIds combinados)
- **#4** — Categorias como chips selecionáveis (de categoriasDisponiveis), não TagInput livre
- **#6/#7** — Round-trip de keywords no backend (POST /api/buscas persiste b.Keywords)
- **#8** — Sync store `buscasSalvas` após salvar/remover (sincronizarStoreExterno)
- **#10** — Botão salvar com label "💾 Salvar" (não apenas ícone)

### Infraestrutura
- **Config declarativa** (`busca-config.js`) — defaults, normalização, guards, INTENT_TABLE
- **Harness E2E local** — bypass auth via `window.__E2E_AUTH_USER__`, 3/3 testes passando
- **Spec rules-service** — design + requirements + tasks completos para Go + zen-go

### Métricas
- 208 unit tests passando
- 3 E2E locais passando
- svelte-check 0/0, eslint 0 warnings, build ok
- Backend: precisa validar no CI (dotnet não disponível localmente)

## ⬜ Pendente — Rules Service (spec em `.kiro/specs/rules-service/`)

### Task 1: Proto
- Criar `protos/rules/v1/rules.proto`
- `buf generate` para Go e C# stubs

### Task 2: JDM
- Criar `rules/busca-rules.json` (Decision Tables: intent + guards; Expression Nodes: normalização)
- Formato GoRules JDM

### Task 3: Go server
- `services/rules/main.go` + `server.go`
- gorules/zen-go + atomic pointer + health check + SIGHUP handler
- `go build ./services/rules/...`

### Task 4: Go testes
- Table-driven (4 intents, guard consistency, normalização)
- Property-based (determinismo com rapid)
- Reload concorrente

### Task 5: C# proxy
- `POST /api/rules/evaluate` → gRPC localhost:50055
- DI registration + RequireAuthorization

### Task 6: Frontend integração
- `evaluateRules(ctx)` no effects module
- Cache 30s + fallback local
- BuscaEngine chama rules para intent/validation em ADICIONAR_LOJA e SALVAR

### Task 7: Docker + Deploy
- Dockerfile + Cloud Run config (porta 50055, CPU 0.25, RAM 128Mi)
- CI: build image rules-service

### Task 8: Docs + Drift checks
- Atualizar arquitetura, fluxos, contracts registry
- `mise run prepush` verde

## Decisões de arquitetura (RESPEITAR)

1. **Engine headless (classe Svelte 5).** `BuscaEngine` em `busca-engine.svelte.js` é testável com `new BuscaEngine(mockEffects())`. View (`BuscaUnificada.svelte`) é burra.

2. **Sidecar Go separado do scheduler.** Scheduler = QUANDO executar. Rules = O QUÊ decidir. Porta 50055.

3. **gorules/zen-go.** Binding Go do zen-engine Rust. JDM (JSON Decision Model) carregado do disco. Hot-reload via SIGHUP sem downtime.

4. **JDM versionado no git.** Arquivo `rules/busca-rules.json` editável por PR. Decision Tables para intent/guards. Expression Nodes para normalização.

5. **C# API como proxy transparente.** `POST /api/rules/evaluate` → gRPC. Sem lógica no proxy.

6. **Frontend com fallback.** Guards simples (temContextoBusca) ficam locais para zero-latência. Decisões complexas (intent, validation) consultam backend. Cache 30s.

7. **Config (`busca-config.js`) permanece** para guards/defaults locais. O rules service é para decisões complexas e centralizadas. Não substituir tudo — complementar.

8. **Ports & Adapters.** Effects são injetáveis. Mock para testes. Real para produção.

## Como verificar (no ambiente dev)

```bash
# Frontend (tudo deve passar):
cd web && npm run check && npm run lint:js && npm run format:check && npx vitest run && npm run build

# E2E local (3 specs):
cd web && npm run test:e2e:local

# Go (após implementar rules-service):
go build ./services/rules/...
go test ./services/rules/...

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
| `web/src/lib/busca-engine.svelte.js` | FSM (classe Svelte 5 com $state) |
| `web/src/lib/busca-engine-effects.js` | API calls injetáveis |
| `web/src/lib/busca-config.js` | Config declarativa (guards, defaults, intent table) |
| `web/src/lib/components/BuscaUnificada.svelte` | View pura |
| `web/src/lib/busca-unificada-logic.js` | Funções puras (payload, labels) |
| `web/src/lib/descobrir-logic.js` | montarResultados (filtro client-side) |
| `web/src/lib/descobrir.js` | Orquestração de fontes (fetch) |
| `.kiro/specs/rules-service/` | Spec completa (design + requirements + tasks) |
| `web/tests/local/` | E2E com bypass auth |
