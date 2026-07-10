# Handoff — Bugs v3 da página Descobrir (2026-07-09)

> Próxima sessão: corrigir 5 bugs identificados na v3 da state machine.
> Task: T-0056. Branch: main (trabalhar direto).

## Estado atual

- **294 unit tests passando** (lógica da engine funciona)
- **24 E2E locais passando** (com mocks — UI antiga, precisa atualizar seletores)
- **15 E2E prod passando** (auth real, APIs reais)
- v3 state machine deployada em produção (modos, duplicatas, marketplaces)
- Usuário de teste: `e2e@garimpei.app.br` (pwd em `web/.env.e2e.local`)

## Bugs a corrigir (T-0056)

### Bug 1: Nomes de loja = IDs numéricos ou keywords

**Onde:** `web/src/lib/busca-unificada-logic.js` linha 35
```js
shopNomes: busca.nome ? { [busca.shop_ids?.[0]]: busca.nome } : {}
```

**Problema:**
- `busca.nome` vem de `b.Keyword` no backend (pode ser "serum", não nome da loja)
- Só mapeia o primeiro shopId → demais mostram ID numérico
- Backend (`BuscasEndpoints.cs` linha ~30): `nome = hasShop ? b.Keyword : null`

**Fix necessário:**
1. Backend: guardar `ShopNames` (dict) na entidade `Busca` ao criar/salvar
2. API: retornar `shop_names: { "920292999": "Le Botanic" }` no GET /api/buscas
3. Frontend: `payloadToConfig` usar `busca.shop_names` diretamente

### Bug 2: Título dos BuscaCard

**Onde:** `web/src/lib/components/BuscaCard.svelte` linha 26
```js
let titulo = $derived(gerarLabelBusca(busca));
```

**Fix:** Remover a div do título bold. Manter apenas as seções + cron badge.

### Bug 3: Cards muito altos

**Onde:** `BuscaCard.svelte` layout vertical com labels de 74px

**Fix:** Compactar — uma linha por seção com inline badges, reduzir gaps.

### Bug 4: Busca sem keyword não retorna nada

**Onde:** `web/src/lib/busca-engine-effects.js` condição:
```js
if (ctx.fontes.curadoria && (ctx.keyword.trim() || ctx.categorias.length > 0 || ctx.shopIds.length > 0))
```
e `web/src/lib/descobrir.js` → `carregarCuradoria` manda `keyword: undefined`

**Fix:** Quando só tem categorias, usar nome da categoria como keyword na API.
Quando só tem loja, usar `FetchShop` (já existe no Collector).

**Investigar:** O que Shopee retorna com keyword vazio? Trending? Recomendados?
Ver `services/collector/server.go` → `Fetch` e `FetchShop`.

### Bug 5: Falta mise run db:reset

**Fix:** Criar `.mise/tasks/db/reset` (truncate PostgreSQL com confirmação).

## Arquivos-chave para a correção

| Arquivo | O quê |
|---------|-------|
| `web/src/lib/busca-unificada-logic.js` | `payloadToConfig` — fix shopNomes |
| `web/src/lib/components/BuscaCard.svelte` | Layout — remover título, compactar |
| `src/Garimpei.Api/Endpoints/BuscasEndpoints.cs` | Retornar shop_names |
| `src/Garimpei.Domain/Entities/Busca.cs` | Campo ShopNames (se necessário) |
| `web/src/lib/busca-engine-effects.js` | Lógica busca sem keyword |
| `web/src/lib/descobrir.js` | `carregarCuradoria` / `buildCuradoriaParams` |
| `services/collector/server.go` | Entender Fetch vs FetchShop |
| `.mise/tasks/db/reset` | Novo script |

## Como verificar

```bash
# Unit tests (devem continuar 294+):
cd web && npx vitest run

# E2E local (seletores podem precisar atualização — ver T-0054):
cd web && npm run test:e2e:local

# E2E prod (15 testes contra produção):
cd web && npm run test:e2e:prod

# Checks completos:
mise run prepush
```

## Decisões a respeitar

- ADR-0027: regras em JSON, funções puras, testes validam contra rules
- v3: modos (explorando/vinculada/editando) — não quebrar
- Engine headless: view burra, lógica na engine
- Código é a correção, não hack de teste
