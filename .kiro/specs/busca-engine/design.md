# Design Document

## Overview

Implementar uma State Machine Engine leve (~200 linhas) que modela toda a lógica da
página Garimpar como uma FSM declarativa. A engine substitui o estado ad hoc atual
(getters/setters em `.svelte.js`) por um modelo com transições explícitas, guards que
impedem estados incoerentes, e efeitos colaterais injetáveis para testabilidade total.

## Components and Interfaces

### busca-engine.js — createMachine

```javascript
/**
 * Factory da máquina de estados. Retorna instância com state reativo e send().
 * @param {MachineConfig} config — estados, transições, guards
 * @param {Effects} effects — funções de efeito colateral (API calls)
 * @returns {Machine} — { state, send }
 */
export function createMachine(config, effects) { ... }
```

**MachineConfig:**
```javascript
{
  initialState: 'idle',
  initialContext: { keyword: '', shopIds: [], comissaoMin: 0.07, ... },
  states: {
    idle: {
      on: {
        DIGITAR:         { target: 'idle', actions: ['atualizarKeyword', 'debounce'] },
        ADICIONAR_LOJA:  { target: 'searching', guard: 'lojaInputValida', effects: ['resolverLoja'] },
        MUDAR_FILTRO:    { target: 'idle', actions: ['atualizarFiltro', 'debounce'] },
        MUDAR_FONTES:    { target: 'idle', actions: ['atualizarFontes', 'debounce'] },
        CARREGAR_SALVA:  { target: 'searching', actions: ['restaurarContexto'], effects: ['executarBusca'] },
        INICIALIZAR:     { target: 'searching', effects: ['inicializar'] },
        DEBOUNCE_FIM:    { target: 'searching', guard: 'temContextoBusca', effects: ['executarBusca'] }
      }
    },
    searching: {
      on: {
        BUSCA_SUCESSO:   { target: 'results', actions: ['setResultados'] },
        BUSCA_ERRO:      { target: 'error', actions: ['setErro'] },
        DIGITAR:         { target: 'idle', actions: ['atualizarKeyword', 'debounce'] },
        CANCELAR:        { target: 'idle' }
      }
    },
    results: {
      on: {
        DIGITAR:         { target: 'idle', actions: ['atualizarKeyword', 'debounce'] },
        ADICIONAR_LOJA:  { target: 'searching', guard: 'lojaInputValida', effects: ['resolverLoja'] },
        REMOVER_LOJA:    { target: 'searching', actions: ['removerLoja'], effects: ['executarBusca'] },
        MUDAR_FILTRO:    { target: 'results', actions: ['atualizarFiltro', 'refiltrar'] },
        MUDAR_FONTES:    { target: 'searching', actions: ['atualizarFontes'], effects: ['executarBusca'] },
        SALVAR:          { target: 'saving', effects: ['salvarBusca'] },
        CARREGAR_SALVA:  { target: 'searching', actions: ['restaurarContexto'], effects: ['executarBusca'] },
        LIMPAR:          { target: 'idle', actions: ['resetContexto'] }
      }
    },
    saving: {
      on: {
        SALVAR_SUCESSO:  { target: 'results', actions: ['adicionarBuscaSalva'] },
        SALVAR_ERRO:     { target: 'results', actions: ['setErro'] }
      }
    },
    error: {
      on: {
        DIGITAR:         { target: 'idle', actions: ['atualizarKeyword', 'debounce'] },
        RETRY:           { target: 'searching', effects: ['executarBusca'] },
        LIMPAR:          { target: 'idle', actions: ['resetContexto'] }
      }
    }
  },
  guards: { ... },
  actions: { ... }
}
```

**Machine instance:**
```javascript
const machine = createMachine(buscaConfig, effects);

// State é um getter reativo (compatível com Svelte 5 $derived)
machine.state;         // { name: 'results', context: {...} }
machine.send(event);   // despacha evento
```

### busca-engine-effects.js — Efeitos Injetáveis

```javascript
export function criarEffects(apiDeps) {
  return {
    async inicializar(ctx) { ... },      // sincroniza store + carrega categorias
    async executarBusca(ctx) { ... },     // chama carregarFontes + montarResultados
    async resolverLoja(ctx, event) { ... }, // chama POST /api/lojas (resolve)
    async salvarBusca(ctx) { ... },       // chama POST /api/buscas
    async carregarCategorias(ctx) { ... } // carrega categorias Shopee (ADR-0019)
  };
}
```

**Para testes:**
```javascript
const mockEffects = {
  async executarBusca(ctx) { return { type: 'BUSCA_SUCESSO', resultados: [...] }; },
  async resolverLoja(ctx, ev) { return { type: 'LOJA_RESOLVIDA', shopId: 123, nome: 'X' }; }
};
const machine = createMachine(buscaConfig, mockEffects);
```

### BuscaUnificada.svelte — View Pura

```svelte
<script>
  import { createMachine } from '$lib/busca-engine.js';
  import { criarEffects } from '$lib/busca-engine-effects.js';

  const effects = criarEffects({ buscasSalvas, favoritos, ... });
  const machine = createMachine(buscaConfig, effects);

  // Derivados para a UI (zero lógica de negócio aqui)
  let ctx = $derived(machine.state.context);
  let stateName = $derived(machine.state.name);
  let loading = $derived(stateName === 'searching');
</script>

<!-- Template renderiza ctx e despacha events -->
<input oninput={(e) => machine.send({ type: 'DIGITAR', value: e.target.value })} />
```

### Diagrama de estados

```
                    ┌──────────────────────────────────────────────┐
                    │                                              │
    INICIALIZAR     ▼                                              │
  ─────────────► [idle] ──DEBOUNCE_FIM──► [searching]             │
                    │                         │    │               │
                    │ CARREGAR_SALVA          │    │               │
                    ├─────────────────────────┘    │               │
                    │                              │               │
                    │                    BUSCA_SUCESSO              │
                    │                         │                    │
                    │                         ▼                    │
                    │                    [results] ──SALVAR──► [saving]
                    │                         │                    │
                    │                         │ MUDAR_FILTRO       │
                    │                         └───────┐            │
                    │                                 │ (refiltrar │
                    │                                 │  local)    │
                    │                                 │            │
                    │                    BUSCA_ERRO    │    SALVAR_SUCESSO
                    │                         │       │            │
                    │                         ▼       ▼            │
                    │                      [error] ◄──────────────┘
                    │                         │
                    │         RETRY           │
                    └─────────────────────────┘
```

### Guards

```javascript
const guards = {
  // Busca só executa se tem contexto (keyword ou lojas)
  temContextoBusca: (ctx) => ctx.keyword.trim().length > 0 || ctx.shopIds.length > 0,

  // Loja input é válida (não vazia)
  lojaInputValida: (ctx, event) => (event.value ?? '').trim().length > 0,

  // Comissão está no range válido
  comissaoValida: (ctx, event) => {
    const v = event.value;
    return typeof v === 'number' && v >= 0 && v <= 1;
  },

  // Pode salvar (tem pelo menos keyword ou loja)
  podesSalvar: (ctx) => ctx.keyword.trim().length > 0 || ctx.shopIds.length > 0
};
```

### Actions (mutações puras de contexto)

```javascript
const actions = {
  atualizarKeyword: (ctx, event) => ({ ...ctx, keyword: event.value }),

  atualizarFiltro: (ctx, event) => {
    const updates = { ...ctx };
    if ('comissaoMin' in event) {
      let v = event.comissaoMin;
      if (v > 1) v = v / 100;           // Normaliza 7 → 0.07
      updates.comissaoMin = Math.round(v * 10000) / 10000; // Max 4 decimais
    }
    if ('vendasMin' in event) updates.vendasMin = Math.max(0, Math.floor(event.vendasMin));
    if ('categorias' in event) updates.categorias = event.categorias;
    return updates;
  },

  atualizarFontes: (ctx, event) => ({ ...ctx, fontes: event.fontes }),

  setResultados: (ctx, event) => ({
    ...ctx,
    resultados: event.resultados,
    contagens: event.contagens,
    loading: false,
    error: null
  }),

  setErro: (ctx, event) => ({ ...ctx, error: event.error, loading: false }),

  restaurarContexto: (ctx, event) => ({
    ...ctx,
    keyword: event.config.keyword ?? ctx.keyword,
    shopIds: event.config.shopIds ?? [],
    shopNomes: event.config.shopNomes ?? {},
    comissaoMin: event.config.comissaoMin ?? 0.07,
    vendasMin: event.config.vendasMin ?? 0,
    categorias: event.config.categorias ?? [],
    fontes: event.config.fontes ?? ctx.fontes,
    cron: event.config.cron ?? null
  }),

  removerLoja: (ctx, event) => {
    const shopIds = ctx.shopIds.filter(id => id !== event.shopId);
    const shopNomes = { ...ctx.shopNomes };
    delete shopNomes[event.shopId];
    return { ...ctx, shopIds, shopNomes };
  },

  adicionarBuscaSalva: (ctx, event) => ({
    ...ctx,
    buscasSalvas: [event.busca, ...ctx.buscasSalvas]
  }),

  resetContexto: () => INITIAL_CONTEXT,

  // MUDAR_FILTRO em results: refiltrar localmente sem re-fetch
  refiltrar: (ctx) => {
    // Aplica filtros client-side sobre dadosBrutos (armazenados no ctx)
    const resultados = montarResultados({
      fontes: ctx.fontes,
      dadosCuradoria: ctx.dadosBrutos.curadoria,
      dadosQuedas: ctx.dadosBrutos.quedas,
      dadosNovos: ctx.dadosBrutos.novos,
      dadosLojas: ctx.dadosBrutos.lojas,
      favoritos: ctx.dadosBrutos.favoritos,
      busca: ctx.keyword,
      categorias: ctx.categorias,
      comissaoMin: ctx.comissaoMin,
      vendasMin: ctx.vendasMin
    });
    return { ...ctx, resultados, contagens: computeContagens(resultados) };
  }
};
```

## Data Models

### Machine_Context (forma completa)

```typescript
interface MachineContext {
  // Busca
  keyword: string;
  shopIds: number[];
  shopNomes: Record<number, string>;

  // Filtros
  comissaoMin: number;       // sempre decimal (0.07, nunca 7)
  vendasMin: number;         // sempre inteiro ≥ 0
  categorias: string[];
  categoriasDisponiveis: Category[]; // autocomplete (ADR-0019)

  // Fontes
  fontes: { curadoria: boolean, quedas: boolean, novos: boolean, lojas: boolean, favoritos: boolean };

  // Agendamento
  cron: string | null;

  // Resultados
  resultados: Product[];
  contagens: { curadoria: number, quedas: number, novos: number, lojas: number };
  dadosBrutos: { curadoria: [], quedas: [], novos: [], lojas: [], favoritos: [] };

  // Buscas salvas
  buscasSalvas: SavedSearch[];

  // UI state
  loading: boolean;
  error: string | null;
  lojaResolvendo: boolean;
  lojaErro: string | null;
  filtrosAberto: boolean;
  salvarAberto: boolean;
  colapsado: boolean;
}
```

### Debounce interno

O debounce é implementado dentro da engine como um timer interno:

```javascript
// Dentro de createMachine:
let debounceTimer = null;

function handleDebounce(ctx) {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    send({ type: 'DEBOUNCE_FIM' });
  }, 400);
}
```

Transições com `actions: ['debounce']` chamam `handleDebounce`. O event
`DEBOUNCE_FIM` é gerado internamente pelo timer e processado normalmente.

## Error Handling

| Cenário | State transition | Resultado |
|---------|-----------------|-----------|
| API candidatos timeout | searching → error | ctx.error = "A busca demorou demais" |
| Loja não encontrada | permanece no estado atual | ctx.lojaErro = "Loja não encontrada" |
| POST /api/buscas falha | saving → results | ctx.error = "Falha ao salvar" (toast) |
| Categorias fetch falha | permanece | ctx.categoriasDisponiveis = [] (fallback: free text) |

## Testing Strategy

### Unit tests (busca-engine.test.js)

```javascript
describe('Machine — cenário: busca com loja', () => {
  const machine = createMachine(buscaConfig, mockEffects);

  it('DIGITAR "serum" atualiza keyword', () => {
    machine.send({ type: 'DIGITAR', value: 'serum' });
    expect(machine.state.context.keyword).toBe('serum');
  });

  it('DEBOUNCE_FIM dispara busca', () => {
    machine.send({ type: 'DEBOUNCE_FIM' });
    expect(machine.state.name).toBe('searching');
  });

  it('BUSCA_SUCESSO mostra resultados', () => {
    machine.send({ type: 'BUSCA_SUCESSO', resultados: [...] });
    expect(machine.state.name).toBe('results');
  });

  it('ADICIONAR_LOJA busca serum DENTRO da loja', () => {
    machine.send({ type: 'ADICIONAR_LOJA', value: 'https://s.shopee.com.br/...' });
    // Effect executarBusca deve receber ctx com keyword="serum" E shopIds=[...]
    expect(mockEffects.executarBusca).toHaveBeenCalledWith(
      expect.objectContaining({ keyword: 'serum', shopIds: [920292999] })
    );
  });
});
```

### E2E tests (Playwright)

Cenários baseados nos bugs reportados:
1. Busca "serum" → adiciona Le Botanic → resultados são da loja
2. Filtro comissão nunca mostra float cru
3. Salvar → chip correto → clicar chip restaura tudo
4. Agendar → POST inclui cron → badge ⏱

### Auth bypass

```javascript
// Em +layout.svelte ou hook:
if (import.meta.env.DEV && (url.searchParams.has('test') || env.SKIP_AUTH)) {
  user = { uid: 'test-user', email: 'test@garimpo.dev' };
}
```

## Summary of Changes

| Arquivo | Tipo | Descrição |
|---------|------|-----------|
| `web/src/lib/busca-engine.js` | **Novo** | createMachine + config + guards + actions |
| `web/src/lib/busca-engine-effects.js` | **Novo** | Efeitos colaterais (API calls) injetáveis |
| `web/src/lib/components/BuscaUnificada.svelte` | Reescrito | View pura (renderiza state, despacha events) |
| `web/src/lib/components/BuscaUnificada.svelte.js` | **Deletado** | Substituído pela engine |
| `web/src/tests/busca-engine.test.js` | **Novo** | Testes da engine (cenários reais) |
| `web/src/lib/busca-unificada-logic.js` | Mantido | Funções puras reutilizadas pela engine |
| `web/src/routes/+page.svelte` | Ajuste mínimo | Continua consumindo BuscaUnificada via callbacks |
| `web/src/routes/+layout.svelte` | Ajuste | Auth bypass para testes |
| `web/tests/descobrir-engine.spec.js` | **Novo** | E2E baseados nos cenários reais |

## Non-goals

- Não instalar XState (engine leve própria)
- Não mudar backend (endpoints permanecem)
- Não reescrever ProductCard ou outros componentes (apenas BuscaUnificada)
- Não mudar a lógica de `montarResultados` (continua como filtro client-side)
