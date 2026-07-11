# Handoff — Página Garimpar / BuscaEngine (2026-07-09)

> ✅ **CONCLUÍDO em 2026-07-10.** Bugs corrigidos no commit `fix(T-0056+T-0057)`. Engine v3 estável.

> Documento de passagem para a próxima sessão (outra IA). Descreve o estado, o
> que já foi resolvido, o que falta, as causas-raiz e como verificar. Leia junto
> com `docs/componentes.md` e `src/lib/busca-config.js`.

## Contexto

A página **Garimpar** (`web/src/routes/+page.svelte`) unificou Descobrir + Lojas.
O estado é controlado pela **BuscaEngine** (`web/src/lib/busca-engine.svelte.js`),
uma FSM Svelte 5 *headless* (classe com runes), com effects injetáveis
(`busca-engine-effects.js`) e lógica pura (`busca-unificada-logic.js`,
`descobrir-logic.js`). O usuário reportou uma lista de bugs e pediu: conta de
teste que pule o Firebase, cenários E2E fim-a-fim, e uma engine que bloqueie
estados incoerentes com regras representáveis via config.

## Branch e commits

Branch: `claude/monitored-stores-refactor-xazzf9` (base = `origin/main`).
Commits desta sessão (do mais antigo ao mais novo):
- `bab81ec` — config declarativa (`busca-config.js`) + engine consome a config.
- `4ed309b` — harness E2E local com bypass de auth.
- `be8d5d0` — fix #2 (escopo de loja) + fix #6/#7 (round-trip de keywords no backend).

## Decisões de arquitetura (respeitar)

- **Engine headless, não componente.** Manter a lógica em `busca-engine.svelte.js`
  (testável com `new BuscaEngine(mockEffects())`, sem DOM). A view
  (`BuscaUnificada.svelte`) é burra: só `engine.send(event)` e lê `engine.ctx`.
- **Config para dados/decisões, código para fluxo.** Regras (defaults,
  normalização, guards `requiresAny`, `INTENT_TABLE`) vivem em `busca-config.js`.
  Sem `json-rules-engine` (overkill). Se a FSM crescer muito, avaliar XState.
- **Ports & Adapters.** `effects.executarBusca(ctx)` é a porta; hoje o adapter
  chama a API C#/Shopee. Backends futuros (Solr/Lucene) entram como novos
  adapters selecionáveis por config — sem tocar engine/view.

## ✅ Já resolvido nesta sessão

- **Harness de teste local (pedido do usuário).** `window.__E2E_AUTH_USER__`
  (injetado pelo Playwright) faz o app subir autenticado sem Firebase; `mockApi()`
  controla `/api/**`. Rodar: `npm run test:e2e:local`
  (`PW_CHROMIUM=/opt/pw-browsers/chromium-1194/chrome-linux/chrome` neste ambiente).
  Specs em `web/tests/local/`. **3/3 passam.**
- **Config declarativa** (`busca-config.js`) + testes (`busca-config.test.js`).
- **#2 le botanic** — busca escopada na loja. `carregarCuradoria` agora recebe
  `shopIds` do contexto; `executarBusca` passa `ctx.shopIds`. Com keyword+loja,
  fetch é `fonte:'shopee-shop'` (não global); keyword filtra client-side.
- **#6/#7 round-trip "(sem keywords)"** — **backend**: `POST /api/buscas`
  (`BuscasEndpoints.cs`) agora persiste `b.Keywords`; o GET de shop-busca voltava
  `keywords=[]`. Conserta o label e o pill que apagava a keyword. **Validar no CI**
  (não há dotnet neste ambiente).

## ⬜ Pendente (com causa-raiz e onde mexer)

### #4 — Filtro de categoria regrediu (ADR-0019)
Antes era um seletor das **categorias de 1º nível da Shopee**; hoje
`BuscaUnificada.svelte` usa um `<TagInput>` de texto livre para `categorias`.
As categorias já são carregadas em `engine.ctx.categoriasDisponiveis`
(`effects.carregarCategorias` → `categorias.js` → `/api/categorias`).
**Fix:** trocar o `TagInput` por um seletor (ex.: `ToggleGroup type="multiple"`
ou `Select`) alimentado por `engine.ctx.categoriasDisponiveis`, emitindo
`MUDAR_FILTRO { categorias }`. Arquivo: `BuscaUnificada.svelte` (bloco de filtros).

### #8 — Novos da loja salva não aparecem (stores fora de sync)
`effects.getBuscasSalvas` lê o store `$buscasSalvas`; mas `#salvar` (na engine)
recarrega só `engine.ctx.buscasSalvas` (via `carregarBuscasSalvas`), **sem
atualizar o store `buscasSalvas`**. Então, após salvar a le botanic, o
`executarBusca` (que monta `buscasComLojas` de `getBuscasSalvas()`) ainda não vê
a loja, e quedas/novos não coletam para ela.
**Fix:** após salvar/remover, sincronizar o store `buscasSalvas`
(`buscasSalvas.sincronizarDoServidor()` ou injetar um effect que atualize o
store), OU unificar a fonte (a engine passar a ser a única dona das buscas
salvas e o effect ler de `engine.ctx`). Cuidado com o cache de 2min em
`carregarOportunidades`/`carregarProdutosLojas` (`descobrir.js`) — pode mascarar.
Arquivos: `busca-engine.svelte.js` (#salvar/#removerSalva), `busca-engine-effects.js`.

### #10 — Testes de filtros, polish e cenários E2E
- **Filtros automatizados (#3):** `montarResultados` (`descobrir-logic.js`) já
  filtra comissão/vendas/categoria client-side. Adicionar unit tests cobrindo
  esses filtros (o usuário não conseguia testar manualmente porque as comissões
  eram todas > 15%). O "7.0000000" **já está resolvido** na main (o filtro virou
  `Select` com rótulos "5%/7%/…" em `BuscaUnificada`); confirmar e fechar.
- **Botão salvar (#5):** o `💾` é confuso — dar label/tooltip ("Salvar busca").
  `BuscaUnificada.svelte`.
- **Cenários E2E fim-a-fim:** em `web/tests/local/`, adicionar (TDD) — escopo de
  loja (regressão do #2: mockar `/api/candidatos` retornando produto diferente
  conforme `shopIds` no request e afirmar que só a loja aparece); salvar+agendar
  (cron avançado `42 21 * * *`) e afirmar o label correto (keywords + loja);
  toggle de novos após salvar loja. Depois, o fluxo coleta→estatísticas.

## Como verificar (neste ambiente)

- Frontend: `cd web && npm run check && npm run lint:js && npm run format:check && npm run test:unit` (208 testes) `&& npm run build`.
- E2E local: `cd web && PW_CHROMIUM=/opt/pw-browsers/chromium-1194/chrome-linux/chrome npm run test:e2e:local`.
- **Backend (dotnet) e E2E com emulador NÃO rodam aqui** — dependem do CI.

## Ressalvas / armadilhas

- `pkill -f "vite preview"` costuma abortar o comando (exit 144) neste ambiente —
  evite; use portas novas ou deixe o Playwright gerenciar o server.
- Há **duas fontes** de buscas salvas (store `buscasSalvas` e `engine.ctx.buscasSalvas`)
  — a raiz do #8. Considerar unificar.
- Ao mexer no backend, lembre que o `check:docs-drift` (job Contracts) regenera
  `docs/gerado/BOARD.md`/`ROADMAP.md` — rode `go run ./cmd/gen-board` se mexer no backlog.
