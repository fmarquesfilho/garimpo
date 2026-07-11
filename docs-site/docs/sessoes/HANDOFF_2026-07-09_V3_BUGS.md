# Handoff — Bugs v3 + Fixtures cross-stack (2026-07-09)

> ✅ **CONCLUÍDO em 2026-07-10.** Todos os 5 bugs corrigidos + fixtures criados.
> Commit: `fix(T-0056+T-0057): shop_names contrato + fixtures cross-stack`
> Branch: main.

## Estado atual (pós-fix)

- **298 unit tests passando** (4 novos: contract test + test atualizado)
- **72 testes C# passando** (39 unit + 33 integration, com ShopNames)
- **24 E2E locais** (seletores podem precisar atualização — ver T-0054)
- **15 E2E prod passando** (auth real, APIs reais)
- v3 state machine deployada com shop_names
- Migration `AddBuscaShopNames` pendente de apply em produção
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

## T-0057: Fixtures compartilhados (fazer junto com T-0056)

A ideia central: **um mesmo dado de teste percorre toda a stack** (Collector → API →
Frontend). Isso prova que o mapeamento está correto em cada fronteira.

```
fixtures/
  lojas.json               ← 3 lojas reais (Le Botanic, Glory of Seoul, COSRX)
  produtos.json            ← 10 produtos determinísticos
  buscas.json              ← 5 buscas (keyword, loja, categoria, mista, agendada)
  respostas/
    collector-fetch.json   ← golden: Collector.Fetch("serum")
    collector-fetchshop.json
    api-candidatos.json    ← golden: GET /api/candidatos
    api-buscas.json        ← golden: GET /api/buscas (com shop_names!)
    api-novidades.json
    frontend-ctx.json      ← golden: engine.ctx após CARREGAR_SALVA
```

**Ordem sugerida:**
1. Criar `fixtures/` com dados das 3 lojas de teste
2. Corrigir backend (shop_names) — grava o formato correto no fixture
3. Atualizar `payloadToConfig` — valida contra `frontend-ctx.json`
4. Gravar golden files (curl produção → salvar)
5. Contract tests (Go + C# + Frontend leem do mesmo fixture)
6. Drift check: `mise run check:fixtures-contract`
7. Bugs restantes (título, compactação, busca sem keyword)
8. `mise run db:reset`

**Lojas de teste (usar nos fixtures):**

| URL | Loja | Shop ID |
|-----|------|---------|
| `https://s.shopee.com.br/70IKp57jnV` | Glory of Seoul | 920292999 |
| `https://s.shopee.com.br/8fQYnxWQqu` | Le Botanic | — |
| `https://s.shopee.com.br/1gGoSgfopD` | — | — |
