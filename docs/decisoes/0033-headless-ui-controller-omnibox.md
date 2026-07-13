# ADR 0033 — Headless UI Controller: Omnibox como renderizador puro

**Status:** aceite  
**Data:** 2026-07-12  
**Supersede:** Complementa ADR-0027 (BuscaEngine headless)  
**Impacto:** Alto — elimina estado disperso em componentes, unifica controle de UI na engine

## Contexto

O Omnibox (T-0055) foi implementado com lógica própria: `$state` local para inputValue,
aberto, highlightIdx; parsing de tokens; geração de sugestões; interpretação de teclas.
Embora funcional, esse design criava dois centros de decisão (engine + omnibox), dificultando:

- Testar o comportamento sem DOM (precisa renderizar o componente)
- Rastrear o estado real do dropdown (está na engine ou no componente?)
- Adicionar funcionalidades (Smart Search) sem inflar o componente
- Observar o uso via OTel (eventos processados no componente não geravam spans)

## Decisão

### Padrão: Headless UI Controller

Todo estado de UI (incluindo o Omnibox) vive na engine. Componentes são renderizadores puros.

```
┌─────────────────────────────────────────┐
│ BuscaEngine (Headless UI Controller)     │
│                                          │
│ ui.omnibox = {                           │
│   inputValue, aberto, highlightIdx,      │
│   modo, opcoes, placeholder              │
│ }                                        │
│                                          │
│ Handlers:                                │
│   OMNIBOX_INPUT → parseia, gera opcoes   │
│   OMNIBOX_KEYDOWN → navega, executa      │
│   OMNIBOX_SELECIONAR → executa opcao     │
│   OMNIBOX_BLUR → fecha                   │
└────────────────────┬────────────────────┘
                     │ estado derivado (getters)
                     ▼
┌─────────────────────────────────────────┐
│ Omnibox.svelte (~65 linhas)              │
│                                          │
│ ZERO $state local                        │
│ Lê: engine.omnibox.*                     │
│ Emite: engine.send(OMNIBOX_INPUT, ...)   │
│ Renderiza: input + dropdown              │
└─────────────────────────────────────────┘
```

### Principios

1. **Events In, State Out** — o componente emite eventos brutos (INPUT, KEYDOWN, CLICK). A engine retorna estado derivado reativo.
2. **Single Source of Truth** — em qualquer instante, todo estado visivel na UI vem de `engine.ctx` + `engine.ui`. Nenhum componente mantém `$state` significativo.
3. **Configuration-Driven** — comportamento do Omnibox (minChars, ordem, maxCategorias, teclado) é dado declarativo em `rules/busca-rules.json`.
4. **Testable Without DOM** — toda lógica pode ser exercitada via `engine.send()` em testes Vitest (sem jsdom, sem Playwright).

### Arquitetura dos módulos

```
busca-engine.svelte.js (orquestrador, ~540 linhas)
├── busca-engine-state.js        — estado inicial + guards
├── busca-engine-omnibox.js      — smart search (input, keyboard, intencao)
├── busca-engine-lojas.js        — adicionar/remover/resolver lojas
├── busca-engine-persistencia.js — salvar/carregar/editar buscas
├── busca-engine-effects.js      — side effects injetaveis (API calls)
├── omnibox-intencao.js          — detecção de intenção (função pura)
├── omnibox-parser.js            — tokenização de input
└── omnibox-sugestoes.js         — geração de sugestões por prefixo
```

Cada módulo exporta funções puras que recebem dados e retornam dados.
A engine chama essas funções e aplica as mutações no `$state` reativo.

### Por que esse pattern (e não alternativas)

| Alternativa | Por que descartada |
|-------------|-------------------|
| XState/Stately | Dependência externa pesada para o que $state do Svelte 5 já faz nativamente |
| Stores separados | Fragmenta estado, cria race conditions entre stores |
| Componente com lógica | Não testável sem DOM, mistura apresentação com decisão |
| Context API | Svelte 5 não tem Context reativo; class com $state é idiomático |

## Consequências

### Positivas

- Omnibox.svelte passou de ~220 linhas com lógica para ~65 linhas de renderização pura
- Smart Search (detecção de intenção, busca de lojas, Store Cards) foi adicionado SEM tocar o componente
- 22 testes unitários exercitam todos os handlers via `engine.send()` — sem DOM
- OTel spans cobrem cada evento processado (observabilidade de uso)
- Qualquer novo modo de busca (ex: busca por imagem) = novo handler na engine + opção no JSON

### Negativas

- Engine tem 540 linhas (próximo do limite) — mitigado pela separação em módulos de domínio
- Componentes não podem reagir independentemente ao estado — precisam sempre da engine
- Debugging requer inspecionar `engine.ui.omnibox` (DevTools do Svelte ajudam)

## Validação

- 457 testes Vitest + 94 xUnit + 20 E2E Playwright = todos verdes
- svelte-check: 0 erros
- Mutation testing baseline disponível via `mise run test:mutate`
- Coverage thresholds por módulo (engine ≥70%, intencao ≥85%)
