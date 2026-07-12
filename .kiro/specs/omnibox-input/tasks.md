# Implementation Plan — Omnibox Input

> Gerado a partir de `design.md` + `requirements.md`. Substituição direta do
> `BuscaUnificada.svelte` (sem migração gradual, sem feature flag).

## Decisão de componente (pesquisa concluída)

**Bits UI Combobox (v2.18.1, já instalado) suporta groups com headers** — reexporta
`Group` + `GroupHeading` do Select, é compatível com Svelte 5 (peer `^5.33.0`, instalado
`5.46.4`) e traz ARIA/teclado nativos. Portanto **`cmdk-sv` não é necessário** (zero deps novas).

**Porém**, o `inputValue` do Bits UI Combobox é read-only e acoplado ao modelo de seleção
(`value` single/multiple) — digitar filtra, selecionar autopreenche o input com o label do item.
Isso conflita com o requisito do omnibox: **texto literal multi-token** (`serum @lebotanic #beleza`),
completar apenas o token final e Enter-para-buscar sem seleção (Req 1.3, 5.1, 5.3). O próprio
codebase já **hand-rolla** `ui/Combobox.svelte` por esse mesmo motivo.

**Decisão:** dropdown hand-rolled seguindo o padrão de `ui/Combobox.svelte`, estendido com
seções agrupadas + headers + ARIA combobox completo. Zero deps novas.

**Nota de UX (seleção):** ao selecionar loja/categoria/marketplace, o token ativo é **removido**
do input (não reinserido como `@nome`) e a seleção aparece como card abaixo do omnibox — igual ao
comportamento atual do `ui/Combobox.svelte` (limpa a query no select). Evita o bug de nomes com
espaço (`Glory of Seoul` → `@Glory of Seoul` quebraria a tokenização) e mantém "sem chips no campo".

## Wave 1 — Config declarativa

- [ ] **1.1** Adicionar bloco `omnibox` em `rules/busca-rules.json`
  (prefixos `@`/`#`/`!`, `minChars: 2`, `maxSugestoes: 7`, `matchBuscaSalva: true`, `debounceMs: 400`).
  _Req 6.5, 8.5; design "Component 4"._
- [ ] **1.2** Estender `rules/busca-rules.schema.json` com a definição de `omnibox`
  (top-level tem `additionalProperties: false` → sem isso o `check:rules-schema` falha).
- [ ] **1.3** Exportar `OMNIBOX` em `web/src/lib/busca-config.js` (re-export de `rules.omnibox`).

## Wave 2 — Parser puro (`web/src/lib/omnibox-parser.js`)

- [ ] **2.1** `parsearInput(raw) → Token[]` — tokeniza por espaço, mapeia prefixos, marca `completo`.
  _Req 7.1–7.4._
- [ ] **2.2** `serializarTokens(tokens) → string` — round-trip inverso.
  _Req 7.5 (propriedade round-trip)._
- [ ] **2.3** `tokensParaContexto(tokens, ctx) → { keyword, shopIds, categorias, marketplacesFiltro, lojasResolvidas }`
  — resolve tokens contra lojas monitoradas/categorias/marketplaces disponíveis. _Req 1.4._
- [ ] **2.4** Testes `web/src/tests/omnibox-parser.test.js` — casos da tabela do design + round-trip + vazio.

## Wave 3 — Gerador de sugestões (`web/src/lib/omnibox-sugestoes.js`)

- [ ] **3.1** `gerarSugestoes(ultimoToken, ctx, config) → Map<tipo, Sugestao[]>`.
  - `< minChars` → Map vazio. _Req 8.1._
  - Token com prefixo → só o tipo correspondente. _Req 8.2._
  - Token sem prefixo (keyword) → busca em loja/categoria/marketplace/busca_salva. _Req 8.3._
  - Buscas salvas primeiro (ordem de inserção do Map). _Req 4.3._
  - Limita a `maxSugestoes` por grupo. _Req 8.5._
  - Match case-insensitive: `includes` (loja/categoria/busca_salva), `startsWith` (marketplace). _Req 4.4._
- [ ] **3.2** Testes `web/src/tests/omnibox-sugestoes.test.js` — casos da tabela do design + estado zero (Req 8.4).

## Wave 4 — Componente (`web/src/lib/components/Omnibox.svelte`)

- [ ] **4.1** Props `{ engine, lojasMonitoradas, placeholder }`; estado `inputValue`, `aberto`, `highlightIdx`.
- [ ] **4.2** Derivados: `tokens`, `ultimoToken`, `sugestoesMap`, `grupos`, `flat` (para nav por índice global).
- [ ] **4.3** Input: `oninput` → atualiza `inputValue`, envia `DIGITAR` com keyword-only (engine debouncia 400ms). _Req 6.3, 6.5._
- [ ] **4.4** Dropdown agrupado (headers por tipo, ícones), `role=listbox`/`group`/`option`. _Req 4.1–4.2, 10.3._
- [ ] **4.5** Teclado: ArrowUp/Down navega `flat`, Enter seleciona destaque ou busca, Esc fecha, blur fecha. _Req 5.1–5.5, 10.4._
- [ ] **4.6** Seleção → emite `ADICIONAR_LOJA` / `ADICIONAR_CATEGORIA` / `MUDAR_MARKETPLACES` / `CARREGAR_SALVA`
  e remove o token ativo. _Req 6.1–6.4._
- [ ] **4.7** ARIA combobox: `role=combobox`, `aria-expanded`, `aria-controls`, `aria-activedescendant`, `aria-live` de contagem. _Req 10.1–10.2._
- [ ] **4.8** Degradação graceful: 0 lojas / categorias ausentes → só keyword; prefixo inválido → keyword. _Req 9.1–9.4._
- [ ] **4.9** Testes `web/src/tests/omnibox.test.js` (@testing-library/svelte) — render, digitar→dropdown, selecionar→evento, Esc, ArrowDown.

## Wave 5 — Substituição direta

- [ ] **5.1** Reescrever `web/src/lib/components/BuscaUnificada.svelte`: dono da engine, renderiza
  `<Omnibox>` no topo + filtros numéricos (fontes, comissão, vendas, marketplaces) + cards de escopo
  (LojaCard/CategoriaCard) + `BuscasSalvasPanel`. Mantém props `onresultados/oncarregando/onerro`
  (→ `+page.svelte` não muda). Sem raias. Limite 400 linhas. _Design "Migration Strategy"._
- [ ] **5.2** Deletar `web/src/lib/components/Lane.svelte` (só era usado por BuscaUnificada) e remover do índice se listado.
- [ ] **5.3** Verificação: `bunx vitest run` + `bun run check` + `bun run build` verdes; `mise run test:e2e-novos` (pipeline backend, não afetado).

## Notas

- Specs Playwright locais/prod que dirigem as raias (`web/tests/local/busca-rules.spec.js`,
  `web/tests/prod/descobrir.spec.js`) ficam obsoletas com a remoção das lanes — fora do escopo dos
  gates exigidos (vitest/check/build/e2e-novos), a atualizar em sessão separada.
- Sem novos endpoints; reusa cache L2 via `DIGITAR`/`executarBusca` existentes.
