# Implementation Plan: Omnibox Smart Search

## Overview

Evolução do Omnibox para padrão Headless UI Controller: a BuscaEngine se torna o controlador
único de toda a UI da página Descobrir. Componentes são renderizadores puros. Comportamento
configurável via JSON, verificável em CI, observável via OTel. Inclui Smart Dropdown com detecção
de intenção, Store Cards enriquecidos, busca local de lojas, e monitoramento inline.
14 tasks em 6 waves.

## Tasks

- [ ] 1. Migration: persistir campos enriquecidos na tabela Lojas
  - Adicionar colunas à entidade `Loja.cs`: ImageUrl, CoverUrl, FollowerCount, ItemCount, RatingStar, ShopLocation, Description (todos nullable)
  - Configurar em `AppDbContext.OnModelCreating` (MaxLength para URLs/strings)
  - Gerar migration EF Core `AddLojaEnrichedFields`
  - Atualizar `POST /api/lojas/resolver` para persistir os campos no upsert
  - Atualizar `GET /api/lojas/registro` para retornar os novos campos
  - `dotnet test` — tudo verde
  - **Requirements: 5.2, 5.3, 6.1, 6.4**

- [ ] 2. Endpoint `GET /api/lojas/buscar` (local-only)
  - Implementar em LojasEndpoints.cs: recebe `q` (min 2 chars) e `marketplace` (opcional)
  - Busca por NomeNormalizado.Contains, retorna campos enriquecidos, max 20
  - HTTP 400 se q < 2 chars
  - Adicionar `buscarLojas(q, marketplace)` em api.js
  - Teste xUnit: param validation, matching, filtro marketplace
  - **Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6**

- [ ] 3. Módulo `omnibox-intencao.js` (detecção de intenção — função pura)
  - Criar com `detectarIntencao(texto, ctx)` → IntencaoOption[]
  - Detecção de URL (isUrl), opções base (produtos + lojas), match de categorias com sufixo contextual
  - Ler config de `rules.omnibox.intencao` (minChars, ordemOpcoes, maxCategorias, urlPatterns)
  - Testes Vitest: URL → resolver_link; texto → produtos + lojas; categoria match; < 2 chars → vazio; contexto marketplace/lojas
  - **Requirements: 1.1, 1.2, 1.3, 1.5, 1.6, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.5, 3.6**

- [ ] 4. Sub-estado `ui` na engine + migrar UI state existente
  - Criar `criarUIInicial()` em busca-engine-state.js com blocos: omnibox, resultados, paineis
  - Adicionar `ui = $state(criarUIInicial())` à BuscaEngine
  - Migrar campos existentes: `filtrosAberto` → `ui.paineis.filtrosAberto`, `buscasPainelAberto` → `ui.paineis.buscasSalvasAberto`, `salvarAberto` → `ui.paineis.salvarAberto`
  - Adicionar getters: `get omnibox()`, `get modoResultados()`, `get resultadosLojas()`
  - Atualizar BuscaUnificada.svelte para ler de `engine.ui.paineis.*` em vez de campos diretos
  - Testes: criarUIInicial retorna shape correto; getters expõem dados reativos
  - **Requirements: 10.1, 10.8**

- [ ] 5. Handlers OMNIBOX_* na engine (controle completo do Omnibox)
  - Implementar `#omniboxInput(event)`: seta ui.omnibox.inputValue, parseia tokens, roteia modo (intencao vs sugestoes), gera opcoes, atualiza ctx.keyword via debounce
  - Implementar `#omniboxKeydown(event)`: ArrowDown/Up (navegação cíclica), Enter (#omniboxExecutar), Escape (fecha)
  - Implementar `#omniboxSelecionar(event)`: executa opcao por índice
  - Implementar `#omniboxBlur()`: fecha dropdown
  - Implementar `#omniboxExecutar()`: resolve highlightIdx → opcao → #executarIntencao ou #executarSugestaoLegado
  - Implementar `#executarIntencao(opcao)`: switch por tipo → DIGITAR, BUSCAR_LOJAS, ADICIONAR_CATEGORIA, ADICIONAR_LOJA
  - Implementar `#gerarSugestoesLegado(ultimoToken)`: reusa gerarSugestoes existente
  - Ler config de rules.omnibox.intencao para: minChars, navegacaoCiclica, enterSemSelecao
  - Testes Vitest: INPUT seta ui.omnibox; KEYDOWN navega; Enter executa; prefixo ativa sugestoes legado; URL ativa resolver_link
  - **Requirements: 1.1, 1.2, 1.4, 1.6, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 9.1, 9.2, 9.3, 9.4, 9.5, 10.2, 10.3, 10.4**

- [ ] 6. Handler BUSCAR_LOJAS + modoResultados
  - Implementar `#buscarLojas(event)`: guard min 2 chars, seta ui.resultados.modo='lojas', status SEARCHING, chama effect buscarLojasPorNome, popula ui.resultados.lojas
  - Atualizar `#digitar`: se modoResultados === 'lojas', restaurar para 'produtos'
  - Adicionar effect `buscarLojasPorNome(termo)` em effects.js
  - Testes: BUSCAR_LOJAS popula lojas; DIGITAR restaura modo; guard < 2 chars
  - **Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6**

- [ ] 7. Handler MONITORAR_LOJA
  - Implementar `#monitorarLoja(event)`: reutiliza lógica de #adicionarLojaConhecida + atualiza ui.resultados.lojas para refletir nova monitorada
  - Não rola página (zero scrollIntoView)
  - Erro inline via resolucaoLoja.erro (Store Card lê)
  - Re-fetch dos resultados atuais após adição
  - Testes: MONITORAR_LOJA adiciona ao escopo; atualiza resultadosLojas; erro não trava
  - **Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6**

- [ ] 8. Reescrever Omnibox.svelte como renderizador puro
  - Remover TODOS os $state locais (inputValue, aberto, highlightIdx)
  - Remover toda lógica de decisão (parsearInput, gerarSugestoes, onInput, selecionar, executarBusca, onKeydown)
  - Componente lê exclusivamente engine.omnibox.* (via getter)
  - Emite apenas: OMNIBOX_INPUT, OMNIBOX_KEYDOWN, OMNIBOX_SELECIONAR, OMNIBOX_BLUR
  - ARIA: combobox, listbox, option, aria-expanded, aria-activedescendant, aria-live (tudo derivado de engine.omnibox)
  - Target: ~50 linhas, zero lógica
  - Teste component: renderiza engine.omnibox corretamente; emite eventos nos callbacks corretos
  - **Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 11.1, 11.2, 11.3, 11.4, 11.5**

- [ ] 9. Componente StoreCard.svelte (configurável por marketplace)
  - Criar com props: loja, engine
  - Ler `storeCard.camposVisiveis[marketplace]` da config para decidir quais campos renderizar
  - Layout: imagem circular (avatar ou fallback), nome + bandeira, info (marketplace, produtos, seguidores, avaliação), botão monitorar/status
  - Emite MONITORAR_LOJA via engine.send()
  - Mobile-first, estilo similar ao ProductCard
  - Testes: renderiza por config Shopee (todos campos); renderiza fallback (campos mínimos); botão emite evento
  - **Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7**

- [ ] 10. Atualizar BuscaUnificada.svelte (renderização condicional + migrar estado)
  - Renderizar StoreCard quando engine.modoResultados === 'lojas'
  - Mensagem "Nenhuma loja encontrada" com sugestão de colar link
  - Migrar uso de engine.filtrosAberto → engine.ui.paineis.filtrosAberto etc
  - Remover qualquer estado local remanescente (TOGGLES pode virar config do JSON)
  - **Requirements: 4.3, 4.4, 7.3, 7.4, 10.8**

- [ ] 11. Observabilidade: OTel spans por evento na engine
  - Importar `@opentelemetry/api` trace (já no projeto)
  - Instrumentar `send()`: span por evento com atributos (type, modo, status)
  - Métricas: engine.events_total (counter), engine.event_duration_ms (histogram)
  - Métrica específica: engine.omnibox.intencao_selecionada (counter por tipo)
  - Teste: mock tracer verifica que spans são criados por tipo
  - **Requirements: 10.6**

- [ ] 12. Expandir rules/busca-rules.json + schema validation
  - Adicionar bloco completo `omnibox.intencao` (habilitado, minChars, ordemOpcoes, maxCategorias, urlPatterns, enterSemSelecao, navegacaoCiclica)
  - Adicionar bloco `omnibox.placeholders` (default, comLoja, comCategoria)
  - Adicionar bloco `storeCard` (camposVisiveis por marketplace, layout, acoes)
  - Adicionar transições: OMNIBOX_INPUT, OMNIBOX_KEYDOWN, OMNIBOX_SELECIONAR, BUSCAR_LOJAS, MONITORAR_LOJA
  - Expandir `busca-rules.schema.json` para validar novos blocos obrigatórios
  - `mise run check:rules-schema` passa
  - **Requirements: 10.5, 10.7**

- [ ] 13. Atualizar contracts e drift checks
  - Criar `contracts/schemas/lojas-buscar.response.json`
  - Atualizar `contracts/schemas/lojas.response.json` com campos enriquecidos
  - Rodar `mise run checks` (api-contract, schema-sync, service-contracts) — tudo verde
  - **Requirements: 5.3**

- [ ] 14. Validação end-to-end e commit
  - Suite completa: `bunx vitest run`, `dotnet test`, `bun run check`, `bun run build`, `mise run checks`
  - Aplicar migration: `dotnet ef database update`
  - Playtest: digitar "glory" → Smart Dropdown com opções; "Pesquisar em Lojas" → Store Cards; colar link → "Resolver Link" → Store Card com imagem; "Monitorar" inline; @prefixo → sugestões legado
  - Verificar OTel spans no console (dev mode)
  - Mobile responsive check
  - Commitar: `feat(omnibox): smart search — headless UI controller, store cards, busca de lojas`
  - **Requirements: todos**

## Task Dependency Graph

```json
{
  "waves": [
    {"tasks": [1, 2, 3]},
    {"tasks": [4, 12]},
    {"tasks": [5, 6, 7]},
    {"tasks": [8, 9, 10]},
    {"tasks": [11, 13]},
    {"tasks": [14]}
  ]
}
```

Wave 1: Backend (migration, endpoint) + módulo puro intencao — independentes.
Wave 2: Estado da engine (sub-estado ui) + config JSON — base para handlers.
Wave 3: Handlers da engine (omnibox, buscar lojas, monitorar) — core da refatoração.
Wave 4: Componentes puros (Omnibox, StoreCard, BuscaUnificada) — consomem engine.
Wave 5: Observabilidade + contracts — polimento.
Wave 6: Validação E2E final.

## Notes

- O proto e o Collector Go já foram atualizados (campos enriquecidos). Falta apenas a migration para persistir (Task 1).
- A mudança mais significativa é a Task 5 + 8: migrar toda a lógica do Omnibox.svelte para dentro da engine. O componente passa de ~220 linhas com lógica para ~50 linhas de renderização pura.
- O padrão Headless UI Controller é o mesmo usado por XState Stately (máquina controla UI state), mas implementado nativamente com Svelte 5 $state (sem dependência externa).
- A observabilidade (Task 11) usa o OTel já configurado no projeto (ADR-0028). Spans do frontend vão para o mesmo collector via OTLP.
- Os TOGGLES hardcoded em BuscaUnificada (fontes novos/quedas/favoritos) são candidatos a ir para o JSON de config em uma futura iteração — por ora mantidos no componente.
- A busca de lojas é local-only (Shopee não expõe busca por nome). O registro cresce conforme o usuário resolve links.
