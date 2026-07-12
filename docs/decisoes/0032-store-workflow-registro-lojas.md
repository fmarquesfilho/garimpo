# ADR-0032: Store Workflow — Registro de Lojas independente + normalização cross-language

**Data:** 2026-07-12
**Status:** Aceita
**Decisores:** Fernando
**Spec:** `.kiro/specs/store-workflow-refactor/`
**Handoff:** `docs/legado/HANDOFF_2026-07-12_STORE-WORKFLOW.md`

## Contexto

Após o Omnibox (T-0055, [ADR-0027](/decisoes/0027-busca-engine-regras-externas) §v4) a
BuscaEngine ficou sólida, mas o **subsistema de lojas era o ponto fraco** da página
Descobrir:

- Lojas eram **derivadas das buscas salvas** (`listarLojasMonitoradas`) — não havia
  cadastro independente. Sem busca salva, não havia loja para autocompletar.
- Nomes com espaço (`Glory of Seoul`) não casavam com tokens sem espaço (`@gloryofseoul`).
- Estado de resolução em flags planas (`lojaResolvendo: bool`, `lojaErro: string`).
- Marketplace com fallback silencioso para `shopee`; sem distinção clara entre loja
  **monitorada** (com cron) e **escopada** (só filtro de busca).

## Decisão

### 1. Registro de Lojas server-side (nova entidade `Loja`)

Tabela PostgreSQL própria (`Loja`), independente de `Busca`, com índice único composto
`(ShopId, Marketplace, OwnerUid)` (multi-tenant; um ShopId pode existir em marketplaces
diferentes). Campos: `Id` (Guid PK), `ShopId` (long), `Nome`, `NomeNormalizado`,
`Marketplace`, `CronExpression` (null = escopada), `SourceUrl`, `OrigemPadrao`.

### 2. Contrato de normalização idêntico C# ↔ JS (validado por fixture)

`Loja.Normalizar` (C#) e `normalizarNome` (`loja-registry.js`) implementam a mesma regra:
**NFD → remove combining marks → mantém só `[a-z0-9]` → lowercase**. A paridade é garantida
por `fixtures/normalizacao-pares.json`, consumido pelos **dois lados** (`LojaTests.cs` +
`loja-registry.test.js`) — se qualquer implementação divergir do `expected`, seu teste
quebra. Cobre acentos, emoji (flags), pontuação e espaços. Assim `@gloryofseoul` casa
"Glory of Seoul" no dropdown E no Enter.

### 3. Endpoints REST

| Endpoint | O quê |
|----------|-------|
| `GET /api/lojas/registro` | Lista o registro central de lojas do tenant |
| `POST /api/lojas/resolver` | Resolve via Collector gRPC → upsert `Loja` → devolve o registro |

Contratos: `contracts/schemas/lojas.request.json` / `lojas.response.json` (registrados em
`contracts/registry.yaml`).

### 4. Estado de resolução explícito (FSM)

`ctx.resolucaoLoja = { status: 'idle' | 'resolvendo' | 'erro', erro? }` substitui as flags
planas. Guard `resolucaoPermitida` bloqueia resoluções concorrentes. Timeout de 10s via
**`AbortController`** (cancela o fetch subjacente, evitando upsert órfão em background).
Dois caminhos em `#adicionarLoja`: **match local exato** no registro (síncrono, sem rede) e
**resolução remota** (assíncrona, `POST /api/lojas/resolver`).

### 5. Tipo de `id` / `shopIds` (a decisão mais robusta, não a do design)

- **API transporta `id` como STRING** (`ShopId.ToString()`): `ShopId` é `long` (64-bit) e
  JS só garante inteiros até 2^53 — string é o padrão à prova de perda de precisão.
- **`ctx.shopIds` é NÚMERO** (o design §Data Models dizia `string[]`): o save vai direto
  para `Busca.ShopIds` = `long[]`, sem coerção nem mudança no backend. Coerção num único
  ponto na entrada (`Number(loja.id)`), válida porque shop_ids reais têm ~9-10 dígitos
  (« 2^53). Se algum marketplace passar a usar IDs > 2^53, só esse ponto muda.
- O `id` da resposta é o **ShopId**, não o Guid PK — é a chave de escopo da busca
  (consumida por `collection_keys`, ver [ADR-0030](/decisoes/0030-busca-contract-unificado)).

### 6. Sem retrocompatibilidade

MVP com 2 usuários (dev + PO). Dados legados podem quebrar; `db:reset` repopula. Migração
(requisito 7 original) foi removida do escopo.

## Consequências

### Positivas
- Lojas passam a existir independentemente de buscas salvas.
- Match robusto a acentos/espaço/emoji, consistente entre dropdown e Enter.
- Estado de resolução testável e sem race (guard + AbortController).
- FSM da BuscaEngine **reusada sem alterações** — valida de novo a arquitetura headless
  ([ADR-0027](/decisoes/0027-busca-engine-regras-externas)).
- `rules.lojaRegistro` (`matchMinChars`) declarativo, consumido pelo código (fonte da verdade).

### Correções da revisão de qualidade (2026-07-12)
A implementação inicial passou nos testes mas tinha 2 bugs **mascarados pelos mocks**:

| Bug | Causa | Fix |
|-----|-------|-----|
| Escopo de busca quebrado em prod | endpoints devolviam `id = l.Id` (Guid) em vez de `l.ShopId.ToString()`; mock usava `id` numérico coincidente | alinhado ao design §11; testes usam `id` string real |
| `TypeError` no match local exato | `matches[0].meta` — o `matchLojas` do registry retorna lojas cruas (sem `.meta`); nenhum teste populava `lojasDisponiveis` | `matches[0]` direto; regressão com registro populado |

### Neutras / Débito conhecido
- `POST /api/lojas` (criação de loja monitorada com cron) permanece — exercido pelos e2e,
  mas ainda sem UI dedicada no frontend.
- Coerção `Number(loja.id)` assume shop_ids « 2^53 (documentado; ponto único de mudança).

## Arquivos-chave

| Arquivo | Papel |
|---------|-------|
| `src/Garimpei.Domain/Entities/Loja.cs` | Entidade + `Normalizar()` |
| `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` | `/api/lojas/registro`, `/api/lojas/resolver` |
| `web/src/lib/loja-registry.js` | `normalizarNome`, `matchLojas` (paridade com C#) |
| `web/src/lib/busca-engine.svelte.js` | `#adicionarLoja` (sync/async), `#resolverLojaRemota` (AbortController) |
| `web/src/lib/busca-engine-state.js` | `resolucaoLoja: {status}`, guard `resolucaoPermitida` |
| `web/src/lib/busca-engine-effects.js` | `carregarRegistroLojas` (substitui `listarLojasMonitoradas`) |
| `fixtures/normalizacao-pares.json` | Fixture de paridade C#/JS |
| `src/Garimpei.Tests/Domain/Entities/LojaTests.cs` | Testa `Normalizar` contra a fixture |
| `web/src/lib/loja-registry.test.js` | Testa `normalizarNome`/`matchLojas` contra a fixture |
| `rules/busca-rules.json` | Bloco `lojaRegistro` (`matchMinChars`, normalização) |
| `contracts/schemas/lojas.{request,response}.json` | Contratos dos endpoints |
