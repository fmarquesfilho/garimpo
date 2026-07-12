# Design Document

## Overview

Este documento descreve o design técnico da evolução "Smart Search" do Omnibox, com uma mudança
arquitetural significativa: a **BuscaEngine se torna o controlador único de toda a UI da página
Descobrir**, incluindo o estado do Omnibox (input, dropdown, navegação por teclado).

O padrão adotado é **Headless UI Controller + Configuration-Driven Behavior**:
- Componentes são renderizadores puros — não decidem nada, só emitem eventos brutos e renderizam estado derivado.
- Toda lógica de decisão (qual dropdown mostrar, quais opções, qual ação executar) vive na engine.
- O comportamento é declarativo — configurado via `rules/busca-rules.json`, verificável em build-time.
- O estado é observável — monitorável via OpenTelemetry (spans por evento, métricas de uso).

**Princípios de design:**
1. **Single Source of Truth** — A engine é o único dono de estado. Nenhum componente mantém estado significativo local.
2. **Events In, State Out** — Componentes enviam eventos brutos (INPUT, KEYDOWN, CLICK). A engine retorna estado derivado reativo.
3. **Configuration-Driven** — Comportamento do Omnibox (prefixos, minChars, ordem de opções, teclado) é dado declarativo em JSON.
4. **Compile-Time Verifiable** — O schema JSON valida completude das regras em CI. Svelte check valida que os componentes consomem o estado correto.
5. **Runtime Observable** — Cada evento processado gera um span OTel com tipo, duração, e resultado. Métricas de uso por tipo de intenção.

## Architecture

### Padrão: Headless UI Controller

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ rules/busca-rules.json (configuração declarativa — source of truth)          │
│   ├─ omnibox: { prefixos, intencao, teclado, placeholders }                 │
│   ├─ transicoes: { OMNIBOX_INPUT, OMNIBOX_SELECIONAR, BUSCAR_LOJAS, ... }   │
│   ├─ modos, guards, intent, defaults, marketplaces                          │
│   └─ storeCard: { camposVisiveis, layout }                                  │
└─────────────────────────────────────────────────────────────────────────────┘
         │ importado em build-time
         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ BuscaEngine (Headless UI Controller)                                         │
│                                                                              │
│ Estado reativo ($state):                                                     │
│   ctx.* ─── estado de domínio (keyword, shopIds, resultados, fontes...)      │
│   ui.omnibox ─── estado de apresentação do Omnibox                           │
│   ui.resultados ─── modo de exibição (produtos vs lojas)                     │
│   ui.paineis ─── estado de paineis colapsáveis                               │
│                                                                              │
│ Getters derivados ($derived via class getter):                               │
│   engine.omnibox → { inputValue, aberto, opcoes, highlightIdx, placeholder } │
│   engine.lojaCards, engine.categoriaCards, engine.resultados                  │
│                                                                              │
│ Eventos aceitos:                                                             │
│   OMNIBOX_INPUT, OMNIBOX_KEYDOWN, OMNIBOX_SELECIONAR, OMNIBOX_BLUR          │
│   BUSCAR_LOJAS, DIGITAR, ADICIONAR_LOJA, ADICIONAR_CATEGORIA, ...            │
│   MONITORAR_LOJA (novo — inline do Store Card)                               │
│                                                                              │
│ Observabilidade:                                                             │
│   Cada send() → span OTel (tipo, duração, resultado)                         │
│   Métricas: eventos/minuto por tipo, latência de effects                     │
└─────────────────────────────────────────────────────────────────────────────┘
         │ estado derivado (getters reativos)
         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Componentes Svelte (renderizadores puros — zero lógica de decisão)           │
│                                                                              │
│ Omnibox.svelte:                                                              │
│   Lê: engine.omnibox.inputValue, .aberto, .opcoes, .highlightIdx             │
│   Emite: engine.send({type:'OMNIBOX_INPUT', value})                          │
│          engine.send({type:'OMNIBOX_KEYDOWN', key})                          │
│          engine.send({type:'OMNIBOX_SELECIONAR', indice})                    │
│                                                                              │
│ BuscaUnificada.svelte:                                                       │
│   Lê: engine.ctx.resultados, engine.ui.resultados.modo, engine.lojaCards     │
│   Emite: engine.send({type:'MUDAR_FILTRO', ...})                             │
│                                                                              │
│ StoreCard.svelte:                                                            │
│   Lê: loja (prop, derivado da engine)                                        │
│   Emite: engine.send({type:'MONITORAR_LOJA', loja})                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Contraste: Antes vs. Depois

```
ANTES (Omnibox com lógica própria):
  Omnibox.svelte:
    inputValue = $state('')        ← estado LOCAL
    aberto = $state(false)         ← estado LOCAL
    highlightIdx = $state(-1)      ← estado LOCAL
    tokens = $derived(parse(...))  ← lógica LOCAL
    sugestoes = $derived(gerar())  ← lógica LOCAL
    onKeydown() { decide... }      ← DECISÃO local sobre qual evento enviar
    selecionar() { decide... }     ← DECISÃO local sobre qual ação tomar

DEPOIS (Omnibox como renderizador puro):
  Omnibox.svelte:
    Lê engine.omnibox.*            ← TUDO vem da engine
    oninput → engine.send(OMNIBOX_INPUT)    ← evento bruto
    onkeydown → engine.send(OMNIBOX_KEYDOWN)← evento bruto
    onclick item → engine.send(OMNIBOX_SELECIONAR) ← evento bruto
    // ZERO lógica de decisão no componente
```

## Components and Interfaces

### 1. Sub-estado `ui` na Engine (novo)

Separação explícita entre estado de domínio (`ctx`) e estado de apresentação (`ui`):

```javascript
// busca-engine-state.js — novo bloco ui
export function criarUIInicial() {
    return {
        omnibox: {
            inputValue: '',
            aberto: false,
            highlightIdx: -1,
            modo: 'intencao',       // 'intencao' | 'sugestoes' (prefixos)
            opcoes: [],             // IntencaoOption[] ou Sugestao[] — o que renderizar
            placeholder: 'Buscar produtos, lojas ou categorias…'
        },
        resultados: {
            modo: 'produtos',       // 'produtos' | 'lojas'
            lojas: []               // LojaResult[] quando modo === 'lojas'
        },
        paineis: {
            buscasSalvasAberto: false,
            filtrosAberto: false
        }
    };
}
```

A engine expõe via getters:
```javascript
get omnibox() { return this.ui.omnibox; }
get modoResultados() { return this.ui.resultados.modo; }
get resultadosLojas() { return this.ui.resultados.lojas; }
```

### 2. Handlers de Omnibox na Engine

```javascript
// Novos handlers no #handlers map:
OMNIBOX_INPUT: (e) => this.#omniboxInput(e),
OMNIBOX_KEYDOWN: (e) => this.#omniboxKeydown(e),
OMNIBOX_SELECIONAR: (e) => this.#omniboxSelecionar(e),
OMNIBOX_BLUR: () => this.#omniboxBlur(),
BUSCAR_LOJAS: (e) => this.#buscarLojas(e),
MONITORAR_LOJA: (e) => this.#monitorarLoja(e),
```

#### `#omniboxInput(event)` — processa cada keystroke

```javascript
#omniboxInput(event) {
    const value = event.value ?? '';
    this.ui.omnibox.inputValue = value;
    this.ui.omnibox.highlightIdx = -1;
    this.ui.omnibox.aberto = true;

    // Parsing e roteamento (lógica que antes vivia no componente)
    const tokens = parsearInput(value);
    const ultimoToken = tokens[tokens.length - 1];

    if (ultimoToken && ultimoToken.tipo !== 'keyword') {
        // Token com prefixo → sistema legado de sugestões
        this.ui.omnibox.modo = 'sugestoes';
        this.ui.omnibox.opcoes = this.#gerarSugestoesLegado(ultimoToken);
    } else {
        // Texto livre → detecção de intenção
        this.ui.omnibox.modo = 'intencao';
        this.ui.omnibox.opcoes = detectarIntencao(value, this.#intencaoCtx);
    }

    // Keyword para a engine (debounce) — só tokens keyword
    const kw = tokens.filter(t => t.tipo === 'keyword').map(t => t.valor).join(' ');
    this.ctx.keyword = kw;
    this.#debounce();
}
```

#### `#omniboxKeydown(event)` — navegação e execução

```javascript
#omniboxKeydown(event) {
    const { key } = event;
    const opcoes = this.ui.omnibox.opcoes;
    const n = opcoes.length;

    if (key === 'ArrowDown') {
        this.ui.omnibox.aberto = true;
        this.ui.omnibox.highlightIdx = n ? (this.ui.omnibox.highlightIdx + 1) % n : -1;
    } else if (key === 'ArrowUp') {
        this.ui.omnibox.highlightIdx = n ? (this.ui.omnibox.highlightIdx - 1 + n) % n : -1;
    } else if (key === 'Enter') {
        this.#omniboxExecutar();
    } else if (key === 'Escape') {
        this.ui.omnibox.aberto = false;
        this.ui.omnibox.highlightIdx = -1;
    }
}
```

#### `#omniboxExecutar()` — executa a opção selecionada ou primeira

```javascript
#omniboxExecutar() {
    const { opcoes, highlightIdx, modo } = this.ui.omnibox;
    if (!opcoes.length) return;

    const idx = highlightIdx >= 0 ? highlightIdx : 0;
    const opcao = opcoes[idx];

    this.ui.omnibox.aberto = false;
    this.ui.omnibox.highlightIdx = -1;

    if (modo === 'intencao') {
        this.#executarIntencao(opcao);
    } else {
        this.#executarSugestaoLegado(opcao);
    }
}
```

#### `#executarIntencao(opcao)` — roteia por tipo

```javascript
#executarIntencao(opcao) {
    switch (opcao.tipo) {
        case 'produtos':
            this.ctx.keyword = opcao.payload.keyword;
            this.ui.resultados.modo = 'produtos';
            this.#executarBusca();
            break;
        case 'lojas':
            this.#buscarLojas({ termo: opcao.payload.termo });
            break;
        case 'categoria':
            this.#adicionarCategoria({ nome: opcao.payload.categoria });
            this.ctx.keyword = '';
            this.ui.omnibox.inputValue = '';
            this.#executarBusca();
            break;
        case 'resolver_link':
            this.#adicionarLoja({ value: opcao.payload.url });
            this.ui.resultados.modo = 'lojas';
            break;
    }
}
```

### 3. Configuração Declarativa Expandida (`rules/busca-rules.json`)

```json
{
  "omnibox": {
    "prefixos": {
      "@": { "tipo": "loja", "fonte": "lojasDisponiveis", "campo": "nome" },
      "#": { "tipo": "categoria", "fonte": "categoriasDisponiveis", "campo": "nome" },
      "!": { "tipo": "marketplace", "fonte": "marketplaces.suportados", "campo": "nome" }
    },
    "intencao": {
      "habilitado": true,
      "minChars": 2,
      "ordemOpcoes": ["produtos", "lojas", "categorias"],
      "maxCategorias": 3,
      "urlPatterns": ["^https?://"],
      "enterSemSelecao": "primeira_opcao",
      "navegacaoCiclica": true
    },
    "placeholders": {
      "default": "Buscar produtos, lojas ou categorias…",
      "comLoja": "Buscar dentro de {loja}…",
      "comCategoria": "Buscar em {categoria}…"
    },
    "maxSugestoes": 7,
    "debounceMs": 400
  },
  "storeCard": {
    "camposVisiveis": {
      "shopee": ["imagem", "nome", "origem", "seguidores", "total_produtos", "avaliacao", "monitorada"],
      "mercado_livre": ["imagem", "nome", "origem", "monitorada"],
      "amazon": ["imagem", "nome", "origem", "monitorada"],
      "_fallback": ["nome", "marketplace", "monitorada"]
    },
    "layout": "horizontal",
    "acoes": ["monitorar"]
  },
  "transicoes": {
    "OMNIBOX_INPUT": { "refetch": false, "imediato": false },
    "OMNIBOX_KEYDOWN": { "refetch": false, "imediato": true },
    "OMNIBOX_SELECIONAR": { "refetch": false, "imediato": true },
    "BUSCAR_LOJAS": { "refetch": true, "imediato": true },
    "MONITORAR_LOJA": { "refetch": true, "imediato": true },
    "DIGITAR": { "refetch": true, "imediato": false },
    "ADICIONAR_LOJA": { "refetch": true, "imediato": true }
  }
}
```

**Verificação em build-time:** O `busca-rules.schema.json` é expandido para validar:
- Presença de todos os campos obrigatórios em `omnibox.intencao`
- `storeCard.camposVisiveis` contém ao menos `_fallback`
- Todas as transições referenciadas nos handlers existem no mapa
- CI (`mise run check:rules-schema`) falha se schema diverge

### 4. Omnibox.svelte Refatorado (Renderizador Puro)

```svelte
<script>
    /** Omnibox — renderizador puro. Zero lógica de decisão.
     *  Lê estado derivado de engine.omnibox.
     *  Emite eventos brutos para a engine.
     */
    let { engine } = $props();

    // Tudo vem da engine — nenhum $state local
    let om = $derived(engine.omnibox);
    let inputEl;
</script>

<div class="relative" onfocusout={(e) => {
    if (!e.currentTarget.contains(e.relatedTarget))
        engine.send({type: 'OMNIBOX_BLUR'});
}}>
    <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 opacity-50">🔍</span>
    <input
        bind:this={inputEl}
        type="text"
        value={om.inputValue}
        placeholder={om.placeholder}
        autocomplete="off"
        spellcheck="false"
        role="combobox"
        aria-expanded={om.aberto && om.opcoes.length > 0}
        aria-controls="omnibox-listbox"
        aria-autocomplete="list"
        aria-activedescendant={om.highlightIdx >= 0 ? `omnibox-opt-${om.highlightIdx}` : undefined}
        class="w-full rounded-sm border border-input bg-background py-2.5 pl-9 pr-4 text-base"
        oninput={(e) => engine.send({type: 'OMNIBOX_INPUT', value: e.currentTarget.value})}
        onfocus={() => engine.send({type: 'OMNIBOX_INPUT', value: om.inputValue})}
        onkeydown={(e) => {
            if (['Enter','ArrowDown','ArrowUp','Escape'].includes(e.key)) {
                e.preventDefault();
                engine.send({type: 'OMNIBOX_KEYDOWN', key: e.key});
            }
        }}
    />

    <span class="sr-only" aria-live="polite">
        {om.aberto && om.opcoes.length > 0
            ? `${om.opcoes.length} ${om.opcoes.length === 1 ? 'opção' : 'opções'}`
            : ''}
    </span>

    {#if om.aberto && om.opcoes.length > 0}
        <ul id="omnibox-listbox" role="listbox" aria-label="Opções de busca"
            class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 max-h-80
                   overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md">
            {#each om.opcoes as opcao, i (opcao.tipo + ':' + i)}
                <li id={`omnibox-opt-${i}`} role="option"
                    aria-selected={om.highlightIdx === i}
                    aria-label={opcao.labelAcessivel ?? opcao.label}>
                    <button type="button" tabindex="-1"
                        class="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-sm
                               transition-colors {om.highlightIdx === i
                                   ? 'bg-accent text-accent-foreground' : 'hover:bg-accent'}"
                        onmouseenter={() => engine.send({type:'OMNIBOX_KEYDOWN', key:'highlight', idx: i})}
                        onclick={() => engine.send({type:'OMNIBOX_SELECIONAR', indice: i})}>
                        <span aria-hidden="true">{opcao.icone}</span>
                        <span class="flex-1 truncate">{opcao.label}</span>
                        {#if opcao.tipo === 'produtos' || opcao.tipo === 'lojas' || opcao.tipo === 'resolver_link'}
                            <kbd class="text-xs text-muted-foreground">↵</kbd>
                        {/if}
                    </button>
                </li>
            {/each}
        </ul>
    {/if}
</div>
```

**Nota:** O componente tem ~50 linhas. Zero `$state` local. Zero `if/else` de decisão. Lê
`engine.omnibox` e emite 4 tipos de eventos brutos. Toda a inteligência está na engine.

### 5. Store Card com Campos Configuráveis

O `storeCard.camposVisiveis` no JSON determina quais campos mostrar por marketplace:

```svelte
<!-- StoreCard.svelte — renderizador configurável -->
<script>
    import { MARKETPLACES } from '$lib/busca-config.js';
    import { STORE_CARD_CONFIG } from '$lib/busca-config.js';

    let { loja, engine } = $props();

    const mktIcone = MARKETPLACES?.icones?.[loja.marketplace] ?? '🛒';
    const campos = STORE_CARD_CONFIG?.camposVisiveis?.[loja.marketplace]
        ?? STORE_CARD_CONFIG?.camposVisiveis?._fallback
        ?? ['nome', 'marketplace', 'monitorada'];
    const monitorando = $derived(engine.ctx.resolucaoLoja.status === 'resolvendo');
</script>

<div class="flex items-center gap-3 rounded-md border border-border bg-card p-3">
    {#if campos.includes('imagem')}
        {#if loja.imagem}
            <img src={loja.imagem} alt={loja.nome} class="h-10 w-10 rounded-full object-cover" />
        {:else}
            <span class="flex h-10 w-10 items-center justify-center rounded-full bg-muted text-lg">{mktIcone}</span>
        {/if}
    {/if}

    <div class="flex-1 min-w-0">
        <div class="flex items-center gap-1.5">
            <p class="truncate font-medium text-sm">{loja.nome}</p>
            {#if campos.includes('origem') && loja.origem}
                <span class="text-xs" title="Origem">{loja.origem}</span>
            {/if}
        </div>
        <p class="text-xs text-muted-foreground">
            {loja.marketplace}
            {#if campos.includes('total_produtos') && loja.total_produtos}
                · {loja.total_produtos} produtos
            {/if}
            {#if campos.includes('seguidores') && loja.seguidores}
                · {loja.seguidores.toLocaleString()} seguidores
            {/if}
            {#if campos.includes('avaliacao') && loja.avaliacao}
                · ⭐ {loja.avaliacao.toFixed(1)}
            {/if}
        </p>
    </div>

    {#if campos.includes('monitorada')}
        {#if loja.monitorada}
            <span class="text-green-600 text-sm" title="Monitorada">⏱ ✓</span>
        {:else}
            <button
                class="rounded-sm bg-primary/10 px-2 py-1 text-xs font-medium text-primary
                       hover:bg-primary/20 disabled:opacity-50"
                disabled={monitorando}
                onclick={() => engine.send({type: 'MONITORAR_LOJA', loja})}
                aria-label="Monitorar loja {loja.nome}"
            >
                {monitorando ? '…' : '+ Monitorar'}
            </button>
        {/if}
    {/if}
</div>
```

### 6. Observabilidade (OpenTelemetry)

Cada `send()` da engine gera um span:

```javascript
// busca-engine.svelte.js — instrumentação
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('busca-engine');

async send(event) {
    const span = tracer.startSpan(`engine.${event.type}`, {
        attributes: {
            'engine.event_type': event.type,
            'engine.modo': this.ctx.modo,
            'engine.status': this.status
        }
    });
    try {
        // ... handler execution ...
        span.setStatus({ code: 0 }); // OK
    } catch (e) {
        span.setStatus({ code: 2, message: e.message }); // ERROR
        throw e;
    } finally {
        span.end();
    }
}
```

Métricas exportáveis:
- `engine.events_total` (counter, por tipo) — uso de cada feature
- `engine.event_duration_ms` (histogram, por tipo) — latência de processamento
- `engine.omnibox.intencao_selecionada` (counter, por tipo de intenção) — quais opções os usuários escolhem
- `engine.store_card.monitorar_total` (counter) — conversão de descoberta → monitoramento

### 7. Endpoint Backend `GET /api/lojas/buscar` (local-only)

```csharp
app.MapGet("/api/lojas/buscar", async (AppDbContext db, string? q, string? marketplace, CancellationToken ct) =>
{
    if (string.IsNullOrWhiteSpace(q) || q.Length < 2)
        return Results.BadRequest(new { error = "Parâmetro 'q' é obrigatório (mín. 2 caracteres)." });

    var termoNorm = Loja.Normalizar(q);
    var query = db.Lojas.AsQueryable();
    if (!string.IsNullOrWhiteSpace(marketplace))
        query = query.Where(l => l.Marketplace == marketplace.ToLowerInvariant());

    var lojas = await query
        .Where(l => l.NomeNormalizado.Contains(termoNorm))
        .OrderBy(l => l.Nome)
        .Take(20)
        .Select(l => new {
            id = l.ShopId.ToString(),
            nome = l.Nome,
            nome_normalizado = l.NomeNormalizado,
            marketplace = l.Marketplace,
            monitorada = !string.IsNullOrEmpty(l.CronExpression),
            origem = l.OrigemPadrao,
            imagem = l.ImageUrl,
            seguidores = l.FollowerCount,
            total_produtos = l.ItemCount,
            avaliacao = l.RatingStar
        })
        .ToListAsync(ct);

    return Results.Ok(new { lojas, total = lojas.Count });
}).RequireAuthorization().WithTags("Lojas");
```

## Data Models

### Estado da Engine (shape completo pós-refator)

```javascript
{
    // ── Domínio (ctx) ──
    keyword: string,
    shopIds: number[],
    shopNomes: Record<number, string>,
    shopMeta: Record<number, ShopMeta>,
    comissaoMin: number,
    vendasMin: number,
    categorias: string[],
    categoriasDisponiveis: CategoriaDisponivel[],
    lojasDisponiveis: LojaRegistro[],
    marketplacesFiltro: string[],
    fontes: Record<string, boolean>,
    resultados: Produto[],
    buscasSalvas: BuscaSalva[],
    modo: 'explorando' | 'vinculada' | 'editando',
    resolucaoLoja: ResolucaoState,
    error: string | null,

    // ── Apresentação (ui) ── NOVO
    ui: {
        omnibox: {
            inputValue: string,
            aberto: boolean,
            highlightIdx: number,
            modo: 'intencao' | 'sugestoes',
            opcoes: (IntencaoOption | Sugestao)[],
            placeholder: string
        },
        resultados: {
            modo: 'produtos' | 'lojas',
            lojas: LojaResult[]
        },
        paineis: {
            buscasSalvasAberto: boolean,
            filtrosAberto: boolean
        }
    }
}
```

### IntencaoOption

```typescript
interface IntencaoOption {
    tipo: 'produtos' | 'lojas' | 'categoria' | 'resolver_link';
    label: string;
    labelAcessivel: string;
    icone: string;
    payload: {
        keyword?: string;
        termo?: string;
        categoria?: string;
        url?: string;
        marketplaces?: string[];
    };
}
```

### LojaResult (resultado de busca por loja)

```typescript
interface LojaResult {
    id: string;
    nome: string;
    nome_normalizado: string;
    marketplace: string;
    monitorada: boolean;
    origem: string | null;      // bandeira (🇰🇷) — origem_padrao manual
    imagem: string | null;      // URL avatar
    seguidores: number | null;
    total_produtos: number | null;
    avaliacao: number | null;   // 0-5
}
```

### Contrato `GET /api/lojas/buscar` (response)

```json
{
    "lojas": [LojaResult],
    "total": number
}
```

### Contrato `POST /api/lojas/resolver` (response atualizado)

```json
{
    "id": "920292999",
    "nome": "Glory of Seoul",
    "nome_normalizado": "gloryofseoul",
    "marketplace": "shopee",
    "monitorada": false,
    "cron": null,
    "origem": "🇰🇷",
    "imagem": "https://down-br.img.susercontent.com/abc123",
    "capa": "https://down-br.img.susercontent.com/cover456",
    "seguidores": 12450,
    "total_produtos": 342,
    "avaliacao": 4.8,
    "localizacao": "São Paulo"
}
```

## Decisões de Design

| # | Decisão | Justificativa |
|---|---------|---------------|
| 1 | Engine como Headless UI Controller (todo estado na engine) | Elimina estado disperso em componentes. Single source of truth. Componentes se tornam renderizadores puros, trivialmente testáveis. |
| 2 | Sub-estado `ui` separado de `ctx` | Separação de concerns: domínio (busca, lojas, filtros) vs. apresentação (dropdown, painéis). Permite resetar UI sem perder domínio. |
| 3 | Eventos brutos do Omnibox (INPUT, KEYDOWN, SELECIONAR, BLUR) | O componente não interpreta teclas — repassa. A engine decide o que ArrowDown, Enter, Escape significam baseado no contexto. |
| 4 | Configuração declarativa do comportamento do Omnibox | `rules.omnibox.intencao.*` governa tudo: minChars, ordem, maxCategorias, teclado. Mudança de comportamento = editar JSON, zero code change. |
| 5 | `storeCard.camposVisiveis` por marketplace | Cada marketplace tem dados diferentes disponíveis. O JSON declara quais campos mostrar; o componente renderiza condicionalmente. Extensível sem código novo. |
| 6 | OpenTelemetry spans por evento | Observabilidade de uso real: quais intenções são selecionadas, qual a latência, quais erros ocorrem. Dados para decisões de produto. |
| 7 | Busca de lojas local-only (sem Collector SearchShops) | Shopee não expõe busca por nome via API. Registro enriquece naturalmente via resolução de links. |
| 8 | Proto enriquecido (imagem, seguidores, avaliação) | Dados já disponíveis no `get_shop_detail`, apenas não eram parseados. Store Card rico sem chamada extra. |
| 9 | Bandeira de origem = `origem_padrao` manual | API da Shopee não expõe país de origem. Solução pragmática que funciona porque lojas coreanas vendem produtos coreanos. |
| 10 | Placeholder contextual (config-driven) | O placeholder muda conforme o contexto (com loja ativa, com categoria) — definido no JSON, derivado pela engine. |
| 11 | Schema validation em CI para completude de regras | Se `busca-rules.json` não tem campo obrigatório ou transição declarada, o CI quebra. Verificação compile-time. |

## Impacto em Arquivos Existentes

| Arquivo | Mudança |
|---------|---------|
| `web/src/lib/omnibox-intencao.js` | **NOVO** — detectarIntencao (função pura) |
| `web/src/lib/components/StoreCard.svelte` | **NOVO** — card configurável por marketplace |
| `web/src/lib/components/Omnibox.svelte` | **REESCRITO** — renderizador puro (~50 linhas, zero lógica) |
| `web/src/lib/components/BuscaUnificada.svelte` | Simplificado — lê engine.ui.*, remove estado local |
| `web/src/lib/busca-engine.svelte.js` | Novos handlers: OMNIBOX_INPUT/KEYDOWN/SELECIONAR/BLUR, BUSCAR_LOJAS, MONITORAR_LOJA. Sub-estado `ui`. OTel spans. |
| `web/src/lib/busca-engine-state.js` | `criarUIInicial()`, remover UI state que era da engine (filtrosAberto, etc → ui.paineis) |
| `web/src/lib/busca-engine-effects.js` | Novo effect `buscarLojasPorNome` |
| `web/src/lib/busca-config.js` | Exportar `STORE_CARD_CONFIG`, `INTENCAO_CONFIG` do rules JSON |
| `web/src/lib/api.js` | Nova função `buscarLojas(q, marketplace)` |
| `protos/collector/v1/collector.proto` | ✅ Já atualizado (ResolveShopResponse enriquecido) |
| `services/collector/server_shop.go` | ✅ Já atualizado (parseia campos extras) |
| `src/Garimpei.Domain/Entities/Loja.cs` | Novos campos: ImageUrl, CoverUrl, FollowerCount, ItemCount, RatingStar, ShopLocation |
| `src/Garimpei.Infrastructure/Persistence/AppDbContext.cs` | Config dos novos campos |
| `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` | ✅ Resolver já retorna enriquecido + novo GET /api/lojas/buscar |
| `rules/busca-rules.json` | Expandir omnibox (intencao, placeholders, teclado), adicionar storeCard, adicionar transições novas |
| `rules/busca-rules.schema.json` | Expandir schema para validar novos blocos |
| `contracts/schemas/lojas-buscar.response.json` | **NOVO** |
| `contracts/schemas/lojas.response.json` | Campos enriquecidos |

## Rastreabilidade de Requisitos

| Requisito | Componente(s) de Design |
|-----------|------------------------|
| Req 1 — Detecção de Intenção | §2 (`#omniboxInput` com `detectarIntencao`), §3 (config `omnibox.intencao`) |
| Req 2 — Contexto Marketplace/Lojas | §2 (`#intencaoCtx` derivado), `omnibox-intencao.js` (`buildCategoriaOption`) |
| Req 3 — Detecção de Links | §2 (`#omniboxInput` detecta URL), `omnibox-intencao.js` (`isUrl`) |
| Req 4 — Busca de Lojas por Nome | §2 (`#buscarLojas` handler), §7 (endpoint local-only) |
| Req 5 — Endpoint Backend | §7 (`GET /api/lojas/buscar`) |
| Req 6 — Store Card | §5 (`StoreCard.svelte` configurável), §3 (`storeCard.camposVisiveis`) |
| Req 7 — Monitoramento via Card | §2 (`MONITORAR_LOJA` handler), §5 (botão no card) |
| Req 8 — Teclado | §2 (`#omniboxKeydown`, `#omniboxExecutar`), §3 (`omnibox.intencao.teclado`) |
| Req 9 — Coexistência Prefixos | §2 (`#omniboxInput` roteamento por tipo do último token) |
| Req 10 — Acessibilidade | §4 (ARIA no Omnibox puro), §5 (aria-label no StoreCard) |

## Error Handling

| Cenário | Comportamento | UX |
|---------|---------------|----|
| BUSCAR_LOJAS falha (rede) | `ctx.error` setado, `status = ERROR` | Mensagem de erro + retry |
| MONITORAR_LOJA falha | `resolucaoLoja = {status:'erro', erro}` | Feedback inline no Store Card |
| Resolver link falha | `resolucaoLoja = {status:'erro', erro}` | Mensagem no painel de resultados |
| GET /api/lojas/buscar retorna 0 | `ui.resultados.lojas = []`, status RESULTS | "Nenhuma loja encontrada. Tente colar o link." |
| Evento OMNIBOX_INPUT com texto < 2 chars | `ui.omnibox.opcoes = []` | Dropdown fecha (sem opções) |

## Correctness Properties

### Property 1: Single Source of Truth
Em qualquer instante, TODO estado visível na UI é derivado exclusivamente de `engine.ctx` e `engine.ui`. Nenhum componente mantém $state local significativo.

**Validates: Requirements 1.4, 8.1, 8.2, 9.1, 9.2, 9.3**

### Property 2: Exclusividade de modo
`ui.omnibox.modo` é exatamente 'intencao' ou 'sugestoes', nunca ambos. Opções no dropdown correspondem exclusivamente ao modo ativo.

**Validates: Requirements 1.4, 9.5**

### Property 3: Determinismo configurável
Dado o mesmo `rules/busca-rules.json` e o mesmo input, `detectarIntencao` produz as mesmas opções. O comportamento é reproduzível e testável contra fixture.

**Validates: Requirements 1.5, 2.1, 2.2, 2.3, 2.4, 2.5**

### Property 4: Evento bruto → ação determinística
O mapeamento OMNIBOX_KEYDOWN(Enter) → execução da opção correta é determinístico: sempre executa `opcoes[highlightIdx >= 0 ? highlightIdx : 0]`.

**Validates: Requirements 8.1, 8.2**

### Property 5: Observabilidade completa
Todo send() que modifica estado gera um span OTel. É impossível ter mudança de estado sem trace.

**Validates: Requirements 1.1, 4.1, 7.1**

### Property 6: Verificabilidade compile-time
Se `busca-rules.json` viola o schema (campo ausente, transição não declarada), CI falha antes de deploy.

**Validates: Requirements 1.5, 5.4**

## Testing Strategy

| Camada | Tipo | Escopo |
|--------|------|--------|
| `omnibox-intencao.js` | Unit (Vitest) | Função pura: texto → opções determinísticas. URL detection, category match, contexto marketplace/lojas. |
| `busca-engine.svelte.js` (handlers omnibox) | Unit (Vitest) | OMNIBOX_INPUT seta ui.omnibox corretamente; OMNIBOX_KEYDOWN navega; OMNIBOX_SELECIONAR/Enter executa; BUSCAR_LOJAS seta modo; MONITORAR_LOJA adiciona. |
| `Omnibox.svelte` | Component (Vitest + testing-library) | Renderiza engine.omnibox.*; emite eventos corretos ao digitar/teclar/clicar. Sem lógica a testar — apenas binding. |
| `StoreCard.svelte` | Component (Vitest + testing-library) | Renderiza campos conforme config; mostra/esconde por marketplace; botão emite evento correto. |
| `rules/busca-rules.json` | Schema validation (CI) | `busca-rules.schema.json` valida completude de omnibox, storeCard, transições. |
| `LojasEndpoints` | Integration (xUnit) | GET /api/lojas/buscar: param validation, matching, filtro marketplace. |
| Observabilidade | Integration | Verificar que spans são gerados por tipo de evento (mock OTel exporter). |
| Cross-browser | Manual | Mobile Safari/Chrome: dropdown não quebra com teclado virtual. |
