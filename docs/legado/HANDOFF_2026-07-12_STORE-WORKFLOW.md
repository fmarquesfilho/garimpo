# Handoff — Store Workflow Refactor (2026-07-12)

> **Status: 📋 SPEC COMPLETA, PRONTA PARA IMPLEMENTAÇÃO.**
> Feature "Store Workflow Refactor" (T-0056). Spec em `.kiro/specs/store-workflow-refactor/`.
> Este doc é o ponto de retomada para o próximo agente/dev.

## Contexto

O Omnibox (T-0055) foi implementado e está em produção. A BuscaEngine (FSM headless)
está sólida. O **subsistema de lojas é o ponto fraco** — feedback do dono do produto.

Problemas atuais:
- Lojas derivadas de buscas salvas (não existe cadastro independente)
- Nomes com espaço ("Glory of Seoul") não casam com tokens sem espaço (`@gloryofseoul`)
- Estados de resolução são flags planas (`lojaResolvendo: bool`, `lojaErro: string`)
- Marketplace tem fallback silencioso para "shopee"
- Sem distinção clara entre loja monitorada (cron) e loja escopada (filtro)

## O que foi feito nesta sessão

Criação completa da spec (requirements → design → tasks):

| Arquivo | Conteúdo |
|---------|----------|
| `.kiro/specs/store-workflow-refactor/requirements.md` | 6 requisitos com critérios de aceitação (EARS) |
| `.kiro/specs/store-workflow-refactor/design.md` | Arquitetura, componentes, data models, contratos API, error handling, correctness properties, testing strategy |
| `.kiro/specs/store-workflow-refactor/tasks.md` | 11 tasks em 7 waves de execução |

### Decisões tomadas com o dono do produto

1. **Registro de Lojas é server-side** — nova tabela PostgreSQL + endpoints REST na API C#.
2. **Sem retrocompatibilidade** — o requisito 7 (migração) foi removido. Pode quebrar dados legados.
   Motivo: apenas 2 usuários (dev + PO), fase de eliminar dívida técnica.
3. **BuscaEngine e Omnibox permanecem estruturalmente intactos** — mudanças são internas.

## Arquitetura resumida (pós-refactor)

```
Frontend:
  Omnibox → BuscaEngine.send(ADICIONAR_LOJA)
    ├─ event.loja.id → #adicionarLojaConhecida (sync, do registro)
    └─ event.value   → #resolverLojaRemota (async, POST /api/lojas/resolver)
       ctx.resolucaoLoja = {status:'idle'|'resolvendo'|'erro', input?, erro?}

  loja-registry.js (NOVO) — normalizarNome(), matchLojas()
  ctx.lojasDisponiveis ← GET /api/lojas/registro (na inicialização)

Backend:
  Entidade Loja (NOVA tabela) — ShopId, Nome, NomeNormalizado, Marketplace, Cron
  GET  /api/lojas/registro  — lista lojas do tenant
  POST /api/lojas/resolver  — Collector.ResolveShop → upsert Loja → retorna registro
```

## Waves de implementação

| Wave | Tasks | Descrição |
|------|-------|-----------|
| 1 | 1, 3, 4 | Entity Loja + DB, loja-registry.js frontend, state refactor (paralelas) |
| 2 | 2, 5 | Endpoints backend, effects frontend |
| 3 | 6 | Engine handlers — task central, mais complexa |
| 4 | 7 | Inicialização da engine (carregar registro em paralelo) |
| 5 | 8, 9 | Omnibox sugestões + visual monitorada/escopada |
| 6 | 10 | Cleanup: remover código legado, atualizar rules/contracts |
| 7 | 11 | Validação E2E + commit |

## Arquivos-chave para a implementação

### Backend (C#)
- `src/Garimpei.Domain/Entities/Loja.cs` — **NOVO** (criar)
- `src/Garimpei.Infrastructure/Persistence/AppDbContext.cs` — adicionar DbSet + config
- `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` — refatorar (novos endpoints)

### Frontend (Svelte/JS)
- `web/src/lib/loja-registry.js` — **NOVO** (criar)
- `web/src/lib/busca-engine.svelte.js` — refatorar `#adicionarLoja`
- `web/src/lib/busca-engine-state.js` — substituir `lojaResolvendo`/`lojaErro`
- `web/src/lib/busca-engine-effects.js` — novo effect, remover legado
- `web/src/lib/omnibox-sugestoes.js` — usar normalização robusta
- `web/src/lib/descobrir-logic.js` — remover `listarLojasMonitoradas`
- `web/src/lib/api.js` — novos endpoints

### Config/Contracts
- `rules/busca-rules.json` — bloco `lojaRegistro`
- `contracts/schemas/lojas.request.json` — formato resolver
- `contracts/schemas/lojas.response.json` — formato registro
- `fixtures/normalizacao-pares.json` — **NOVO** (cross-validation C#/JS)

## Como rodar / verificar

```bash
# Backend
cd src/Garimpei.Api && dotnet run          # API local
dotnet test                                 # testes C#
dotnet ef database update                   # aplicar migrations

# Frontend
cd web && bunx vitest run                   # testes unitários
bun run check && bun run build              # type check + build

# Playtest local (sem backend local — proxy para prod)
cd web && DEV_API_PROXY=https://garimpei.app.br VITE_API_BASE= bun run dev
```

## Referências

- Spec completa: `.kiro/specs/store-workflow-refactor/` (requirements, design, tasks)
- Handoff anterior (Omnibox): `docs/legado/HANDOFF_2026-07-12_OMNIBOX.md`
- ADR-0027 (BuscaEngine) — FSM e regras externas
- ADR-0030 (BuscaContract) — contrato de busca
- Task backlog: `backlog/tasks/T-0056-bugs-cards-busca-v3.yaml` (seção "Atualização 2026-07-12")
