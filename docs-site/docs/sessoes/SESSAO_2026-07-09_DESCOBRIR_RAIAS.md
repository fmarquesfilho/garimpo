# Sessão 09/Julho 2026 — Redesign da página Descobrir em raias

## Objetivo

Reorganizar a página principal (Descobrir) em **raias horizontais** (metáfora de piscina),
uma por tipo de configuração de busca, e estender a BuscaEngine/regras para suportar os
fluxos que faltavam: multi-marketplace, busca só-categorias, autocomplete de lojas
(monitoradas + resolver nova por link/ID), bandeira de origem e monitoramento no card de
loja, e **edit mode** nas buscas salvas.

Ver **ADR-0004** (layout) e **ADR-0027** (regras/engine v2).

## As quatro raias

1. **Console superior** — input de palavras-chave + 3 botões de grupo (Filtros/Lojas/Buscas)
   com contador cada, + colapsar tudo + limpar tudo.
2. **Filtros** — 2 sub-raias: fontes (Novos/Quedas/Favoritos) + quantitativos (comissão,
   vendas) em cima; categorias (autocomplete com marketplaces + cards) embaixo.
3. **Lojas** — autocomplete de lojas monitoradas + entrada livre "↳ resolver e adicionar
   loja" (link/ID). Cards com marketplace, bandeira de origem e monitoramento.
4. **Buscas** — cards de busca salva em seções (palavras/categorias/lojas/marketplaces) +
   agenda; rodar / editar (edit mode) / remover.

## Mudanças

### Regras (v2.0.0) — `rules/busca-rules.json`

- Guards `temContextoBusca`/`podeSalvar` passam a aceitar `categorias` (busca só-categorias).
- Novo bloco `contextoCategorias.sources` (sources globais quando há categorias sem
  keyword/loja) — avaliado por `sourcesBusca(ctx)`.
- Novo bloco `marketplaces` (`suportados` + `default`).
- Novas transições: `ADICIONAR_CATEGORIA`, `REMOVER_CATEGORIA`, `MUDAR_MARKETPLACES`.
- Schema e drift check (`.mise/tasks/check/rules-schema`) atualizados (+ `marketplaces.default ∈ suportados`).

### Engine / lógica

- `busca-engine.svelte.js`: novos campos de `ctx` (`categoriaMeta`, `shopMeta`,
  `marketplacesFiltro`, `editandoId`, `lojasDisponiveis`), eventos `ADICIONAR_CATEGORIA`/
  `REMOVER_CATEGORIA`/`MUDAR_MARKETPLACES`/`EDITAR_SALVA`, `ADICIONAR_LOJA` aceita loja
  monitorada (por objeto) ou resolução por `value`, **update in-place** em `SALVAR` via
  `editandoId`, e getters `categoriaCards`/`lojaCards`/`contadorFiltros|Lojas|Buscas`.
  `send` virou mapa de handlers (complexidade).
- `busca-config.js`: re-exporta `MARKETPLACES`/`CONTEXTO_CATEGORIAS` + `sourcesBusca`.
- `descobrir-logic.js`: `agruparCategoriasPorMarketplace`, `listarLojasMonitoradas`.
- `busca-engine-effects.js`: `carregarCategorias` agrupa por marketplace;
  `listarLojasMonitoradas` deriva das buscas salvas (sem endpoint novo).
- `busca-unificada-logic.js`: `configToPayload` carrega `id`; `gerarLabelBusca` trata
  categorias-only.

### Componentes

- Novos: `ui/Combobox.svelte` (autocomplete-add com entrada livre), `Lane.svelte`,
  `CategoriaCard.svelte`, `LojaCard.svelte`.
- Reescritos: `BuscaCard.svelte` (busca salva em 4 seções + edit/rodar/remover),
  `BuscaUnificada.svelte` (composição das 4 raias).

## Verificação

```bash
cd web && npm run check && npm run lint:js && npm run lint:css && npx vitest run && npm run build
# → 0 erros/0 warnings, 249 unit tests, build ok
.mise/tasks/check/rules-schema   # → ✅ válido e consistente
```

Render verificado via harness Playwright local (bypass de auth + `mockApi`): raias,
multi-marketplace nos cards de categoria, dropdown de loja com "resolver e adicionar", e
dark mode — todos OK.

## Pendências

- **E2E locais** (`tests/local/busca-rules.spec.js`, `descobrir-cenarios.spec.js`,
  `garimpar.spec.js`) miram a UI antiga e precisam ser reescritos. Não rodam na CI, então
  ficaram para depois.

## Estado da branch

Branch: `feat/descobrir-raias` (base `main`). Sem commit/push — aguardando confirmação.
