# Design Document: Omnibox Input

## Overview

O Omnibox substitui o sistema de raias (lanes) da página Descobrir por um input unificado que resolve keywords, lojas, categorias e marketplaces a partir de um único campo de texto. O usuário digita naturalmente e o sistema infere o tipo; prefixos (`@`, `#`, `!`) funcionam como atalhos opcionais para usuários avançados.

A implementação é 100% frontend — não requer novos endpoints ou mudanças no backend. Reutiliza o cache L2 (já implementado) para respostas instantâneas e os dados já disponíveis no cliente (categorias via `/api/categorias`, lojas monitoradas via buscas salvas).

## Architecture

```mermaid
graph TD
    subgraph "Omnibox Component"
        INPUT[Input Field]
        PARSER[parsearInput — tokenizer]
        SUGGEST[gerarSugestoes — inference]
        DROP[Dropdown UI — grouped]
    end

    subgraph "Existing"
        ENGINE[BuscaEngine FSM]
        RULES[busca-rules.json]
        API[/api/candidatos]
        CACHE[Cache L2 — instant]
    end

    INPUT -->|raw text| PARSER
    PARSER -->|tokens[]| SUGGEST
    SUGGEST -->|sugestões[]| DROP
    DROP -->|seleção| ENGINE
    INPUT -->|Enter| ENGINE
    ENGINE -->|debounce 400ms| API
    API --> CACHE
    SUGGEST -.->|lê| RULES
```

## Components and Interfaces

### Component 1: Token Parser (`web/src/lib/omnibox-parser.js`)

**Purpose:** Função pura que tokeniza o texto do input em tokens tipados.

**Interface:**

```javascript
/**
 * @typedef {Object} Token
 * @property {'keyword'|'loja'|'categoria'|'marketplace'} tipo
 * @property {string} valor - texto sem prefixo
 * @property {boolean} completo - true se seguido de espaço ou fim
 */

/** Tokeniza raw text do input → Token[] */
export function parsearInput(raw: string): Token[]

/** Serializa tokens de volta para string (round-trip) */
export function serializarTokens(tokens: Token[]): string

/** Converte tokens resolvidos → contexto para BuscaEngine */
export function tokensParaContexto(tokens: Token[], ctx: ResolvedContext): EngineContext
```

**Gramática:**

```
<input>       ::= <token> (" " <token>)*
<token>       ::= <prefixed> | <keyword>
<prefixed>    ::= ("@" | "#" | "!") <texto>
<keyword>     ::= <texto>   (texto sem prefixo)
<texto>       ::= [^\s]+
```

**Regras de completude:**
- Token seguido de espaço → `completo: true`
- Último token (sem espaço depois) → `completo: false` (é o token ativo para sugestões)

### Component 2: Gerador de Sugestões (`web/src/lib/omnibox-sugestoes.js`)

**Purpose:** Função pura que gera sugestões a partir do último token incompleto e do contexto disponível.

**Interface:**

```javascript
/**
 * @typedef {Object} Sugestao
 * @property {'keyword'|'loja'|'categoria'|'marketplace'|'busca_salva'} tipo
 * @property {string} label - texto para exibir
 * @property {string} valor - texto a inserir no input
 * @property {Object} [meta] - dados extras (shopId, categoriaId, etc.)
 */

/**
 * @typedef {Object} SugestoesContext
 * @property {Array} lojasMonitoradas - [{id, nome, marketplace}]
 * @property {Array} categoriasDisponiveis - [{nome, slug, marketplace}]
 * @property {Array} marketplaces - ["shopee", "amazon", "mercadolivre"]
 * @property {Array} buscasSalvas - configs do BuscaEngine
 */

/** Gera sugestões agrupadas por tipo */
export function gerarSugestoes(
  ultimoToken: Token,
  ctx: SugestoesContext,
  config: { minChars: number, maxSugestoes: number }
): Map<string, Sugestao[]>
```

**Algoritmo de match:**
1. Se `ultimoToken.valor.length < config.minChars` → retorna vazio
2. Se `ultimoToken.tipo !== 'keyword'` → filtra apenas pelo tipo correspondente
3. Se `ultimoToken.tipo === 'keyword'` (sem prefixo) → busca em TODOS os tipos:
   - Lojas monitoradas: `nome.toLowerCase().includes(query)`
   - Categorias: `nome.toLowerCase().includes(query)`
   - Marketplaces: `nome.toLowerCase().startsWith(query)`
   - Buscas salvas: keywords/shopNames matcham parcialmente
4. Limita a `maxSugestoes` por grupo
5. Buscas salvas sempre primeiro (se match)

### Component 3: Omnibox Svelte Component (`web/src/lib/components/Omnibox.svelte`)

**Purpose:** Componente visual que integra parser + sugestões + dropdown + BuscaEngine.

**Props:**

```javascript
let {
  engine,            // BuscaEngine instance (prop do parent)
  lojasMonitoradas,  // derivado das buscas salvas
  placeholder = "Buscar produtos, lojas ou categorias..."
} = $props();
```

**Estado interno:**

```javascript
let inputValue = $state('');       // texto literal
let aberto = $state(false);        // dropdown aberto
let highlightIdx = $state(-1);     // sugestão destacada
let sugestoes = $derived(/* computed from parser + gerarSugestoes */);
```

**Eventos emitidos para BuscaEngine:**

| Ação do usuário | Evento BuscaEngine | Dados |
|---|---|---|
| Enter (keyword) | `DIGITAR` | `{ keyword }` |
| Seleciona loja | `ADICIONAR_LOJA` | `{ loja: { id, nome } }` |
| Seleciona categoria | `ADICIONAR_CATEGORIA` | `{ nome }` |
| Seleciona marketplace | `MUDAR_MARKETPLACES` | `{ marketplaces }` |

**Debounce:** O evento `DIGITAR` usa debounce de 400ms conforme `busca-rules.json`. Seleções de loja/categoria/marketplace são imediatas (sem debounce).

### Component 4: Configuração em `busca-rules.json`

**Adição ao JSON existente:**

```json
{
  "omnibox": {
    "prefixos": {
      "@": { "tipo": "loja", "fonte": "lojasMonitoradas", "campo": "nome" },
      "#": { "tipo": "categoria", "fonte": "categoriasDisponiveis", "campo": "nome" },
      "!": { "tipo": "marketplace", "fonte": "marketplaces.suportados", "campo": "nome" }
    },
    "minChars": 2,
    "maxSugestoes": 7,
    "matchBuscaSalva": true,
    "debounceMs": 400
  }
}
```

## Data Models

### Token

```typescript
interface Token {
  tipo: 'keyword' | 'loja' | 'categoria' | 'marketplace';
  valor: string;      // texto limpo (sem prefixo)
  completo: boolean;  // true se finalizado (espaço após)
}
```

### Sugestão

```typescript
interface Sugestao {
  tipo: 'keyword' | 'loja' | 'categoria' | 'marketplace' | 'busca_salva';
  label: string;         // "Glory of Seoul" ou "Beleza"
  valor: string;         // "@gloryofseoul.br" ou "#beleza"
  icone?: string;        // "🏪", "📂", "🌐", "💾"
  meta?: {
    shopId?: number;
    categoriaId?: number;
    marketplace?: string;
    buscaId?: string;
  };
}
```

### Resolved Context (output de `tokensParaContexto`)

```typescript
interface ResolvedContext {
  keyword: string;                // texto das keywords concatenadas
  shopIds: number[];              // de tokens @loja resolvidos
  categorias: string[];           // de tokens #categoria
  marketplacesFiltro: string[];   // de tokens !marketplace
}
```

## Algorithmic Pseudocode

### Algorithm 1: Token Parser

```
ALGORITHM parsearInput(raw)
INPUT: raw string do input
OUTPUT: Token[]

BEGIN
  tokens ← []
  parts ← split(raw, /\s+/)
  
  FOR i = 0 TO parts.length - 1 DO
    part ← parts[i]
    completo ← (i < parts.length - 1) OR raw.endsWith(' ')
    
    IF part.startsWith('@') THEN
      tokens.push({ tipo: 'loja', valor: part.slice(1), completo })
    ELSE IF part.startsWith('#') THEN
      tokens.push({ tipo: 'categoria', valor: part.slice(1), completo })
    ELSE IF part.startsWith('!') THEN
      tokens.push({ tipo: 'marketplace', valor: part.slice(1), completo })
    ELSE
      tokens.push({ tipo: 'keyword', valor: part, completo })
    END IF
  END FOR
  
  RETURN tokens
END
```

### Algorithm 2: Geração de Sugestões

```
ALGORITHM gerarSugestoes(ultimoToken, ctx, config)
INPUT: último token (incompleto), contexto (lojas, categorias, etc.), config
OUTPUT: Map<tipo, Sugestao[]>

BEGIN
  IF ultimoToken.valor.length < config.minChars THEN
    RETURN empty map
  END IF
  
  query ← ultimoToken.valor.toLowerCase()
  result ← new Map()
  
  -- Determinar quais tipos buscar
  tiposFiltro ← IF ultimoToken.tipo ≠ 'keyword' THEN [ultimoToken.tipo]
                 ELSE ['loja', 'categoria', 'marketplace', 'busca_salva']
  
  IF 'busca_salva' IN tiposFiltro AND config.matchBuscaSalva THEN
    matches ← ctx.buscasSalvas.filter(b => 
      b.keywords?.some(k => k.includes(query)) OR
      Object.values(b.shopNames ?? {}).some(n => n.toLowerCase().includes(query)))
    result.set('busca_salva', matches.slice(0, config.maxSugestoes))
  END IF
  
  IF 'loja' IN tiposFiltro THEN
    matches ← ctx.lojasMonitoradas.filter(l => l.nome.toLowerCase().includes(query))
    result.set('loja', matches.slice(0, config.maxSugestoes).map(toSugestaoLoja))
  END IF
  
  IF 'categoria' IN tiposFiltro THEN
    matches ← ctx.categoriasDisponiveis.filter(c => c.nome.toLowerCase().includes(query))
    result.set('categoria', matches.slice(0, config.maxSugestoes).map(toSugestaoCategoria))
  END IF
  
  IF 'marketplace' IN tiposFiltro THEN
    matches ← ctx.marketplaces.filter(m => m.toLowerCase().startsWith(query))
    result.set('marketplace', matches.slice(0, config.maxSugestoes).map(toSugestaoMarketplace))
  END IF
  
  RETURN result
END
```

### Algorithm 3: Handling de Seleção

```
ALGORITHM onSugestaoSelecionada(sugestao, tokens, engine)
INPUT: sugestão selecionada, tokens atuais, BuscaEngine
OUTPUT: side-effects (emite evento, atualiza input)

BEGIN
  -- Substituir último token incompleto pelo valor da sugestão
  tokens[tokens.length - 1] ← { tipo: sugestao.tipo, valor: sugestao.valor, completo: true }
  inputValue ← serializarTokens(tokens) + ' '  -- espaço para próximo token
  
  -- Emitir evento imediato para BuscaEngine
  CASE sugestao.tipo
    WHEN 'loja':
      engine.send({ type: 'ADICIONAR_LOJA', loja: sugestao.meta })
    WHEN 'categoria':
      engine.send({ type: 'ADICIONAR_CATEGORIA', nome: sugestao.valor })
    WHEN 'marketplace':
      marketplaces ← [...engine.ctx.marketplacesFiltro, sugestao.valor]
      engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces })
    WHEN 'busca_salva':
      engine.send({ type: 'CARREGAR_SALVA', config: sugestao.meta.config })
    WHEN 'keyword':
      -- Não emite nada — keyword só dispara no Enter
  END CASE
  
  fecharDropdown()
END
```

## Key Functions with Formal Specifications

### Function: parsearInput

**Preconditions:**
- `raw` é string (pode ser vazia)

**Postconditions:**
- Retorna array de Token (pode ser vazio se raw é vazio)
- Cada token tem `tipo`, `valor` e `completo` preenchidos
- `valor` nunca contém o caractere de prefixo
- Round-trip: `parsearInput(serializarTokens(parsearInput(raw)))` == `parsearInput(raw)`

### Function: gerarSugestoes

**Preconditions:**
- `ultimoToken.valor.length >= 0`
- `ctx` contém arrays (podem ser vazios)
- `config.minChars >= 1` e `config.maxSugestoes >= 1`

**Postconditions:**
- Retorna Map onde cada valor tem `.length <= config.maxSugestoes`
- Se `ultimoToken.valor.length < config.minChars` → retorna Map vazio
- Match é sempre case-insensitive
- Buscas salvas aparecem primeiro (se matcham)

## Error Handling

| Cenário | Comportamento |
|---|---|
| `/api/categorias` falha | Omnibox funciona sem sugestões de categorias |
| 0 lojas monitoradas | Sugestões de loja não aparecem; keyword funciona |
| Token com prefixo inválido (`$texto`) | Tratado como keyword |
| Input vazio + Enter | Nenhum evento emitido (guard `podeBuscar` falha) |
| Rede lenta | Sugestões locais (lojas, categorias) são instantâneas; keyword busca usa debounce + cache L2 |

## Testing Strategy

### Unit Tests (`web/src/tests/omnibox-parser.test.js`)

| Teste | Input | Expected |
|---|---|---|
| Keyword simples | `"serum"` | `[{tipo:'keyword', valor:'serum', completo:false}]` |
| Multi-keyword | `"serum vitamina "` | `[{tipo:'keyword', valor:'serum', completo:true}, {tipo:'keyword', valor:'vitamina', completo:true}]` |
| Prefixo loja | `"@glory"` | `[{tipo:'loja', valor:'glory', completo:false}]` |
| Prefixo categoria | `"#beleza"` | `[{tipo:'categoria', valor:'beleza', completo:false}]` |
| Misto | `"serum @lebotanic #beleza"` | 3 tokens com tipos corretos |
| Round-trip | qualquer input válido | `parse(serialize(parse(x))) === parse(x)` |
| Vazio | `""` | `[]` |

### Unit Tests (`web/src/tests/omnibox-sugestoes.test.js`)

| Teste | Cenário |
|---|---|
| Match loja | `"glo"` → sugere "Glory of Seoul" |
| Match categoria | `"bel"` → sugere "Beleza" |
| Match marketplace | `"sho"` → sugere "shopee" |
| Prefixo filtra | `"@glo"` → apenas lojas |
| Min chars | `"s"` (1 char) → vazio |
| Sem lojas | ctx vazio → só keyword |
| Max 7 | 10 categorias matcham → retorna 7 |

### Integration Tests (Vitest + @testing-library/svelte)

- Renderiza Omnibox → digita `"ser"` → dropdown aparece
- Seleciona loja → evento `ADICIONAR_LOJA` emitido
- Enter sem sugestão → evento `DIGITAR` emitido
- Esc fecha dropdown
- ArrowDown navega sugestões

## Performance Considerations

| Aspecto | Decisão |
|---|---|
| Sugestões locais (lojas, categorias) | Filtradas no cliente — O(n) sobre arrays pequenos (~50 categorias, ~20 lojas). Instantâneo. |
| Keyword search | Debounce 400ms + Cache L2 (TTL 30min). ~5ms em hit, ~500ms em miss. |
| Re-render do dropdown | Svelte 5 fine-grained reactivity — só re-renderiza items que mudam. |
| Tamanho do componente | ~200 linhas (Omnibox.svelte) + ~60 linhas (parser) + ~80 linhas (sugestões). Dentro do limite de 400 linhas. |

## Migration Strategy

O Omnibox **substitui diretamente** o `BuscaUnificada.svelte` na rota principal (`/`). Não há rota alternativa — a hipótese é testada em produção com os 2 usuários atuais.

1. Implementar `Omnibox.svelte` + parser + sugestões
2. Substituir `BuscaUnificada.svelte` na página principal
3. Remover lanes mortas (Lane.svelte, autocompletes separados de loja/categoria)
4. Manter filtros numéricos (comissão, vendas, fontes) como controles separados abaixo do omnibox

## Pesquisa de Componentes Existentes (FAZER ANTES DE IMPLEMENTAR)

Antes de construir o dropdown/combobox do zero, a próxima sessão DEVE pesquisar componentes de UI compatíveis com Svelte 5 que já implementem o padrão combobox com agrupamento:

### Candidatos a avaliar

| Componente | Pacote | Por que considerar |
|---|---|---|
| **Bits UI Combobox** | `bits-ui` (já no projeto) | Já usado para lojas/categorias. Suporta groups, keyboard nav, ARIA. Verificar se aceita grupos múltiplos com headers. |
| **cmdk-sv** | `cmdk-sv` | Port do cmdk (React) para Svelte. Command palette com fuzzy search e groups nativos. Pode ser ideal para omnibox. |
| **Melt UI Combobox** | `@melt-ui/svelte` | Headless, ARIA compliant, suporta groups. Verificar compatibilidade Svelte 5. |

### Critérios de decisão

1. **Compatível com Svelte 5 runes** (não usar stores/writable)
2. **Suporta groups nativo** (Lojas, Categorias, Marketplaces como seções com headers)
3. **Keyboard navigation** (Arrow, Tab, Enter, Esc) built-in
4. **ARIA combobox** (role, aria-expanded, aria-activedescendant) built-in
5. **Já está no projeto** (Bits UI tem prioridade — zero dep nova)
6. **Customizável com Tailwind** (sem CSS conflitante)

### Decisão preferencial

**Bits UI Combobox** já está no projeto e resolve 90% do caso. Se faltar apenas agrupamento visual com headers, um wrapper simples resolve. Só adicionar dep nova se Bits UI não suportar o padrão de groups com múltiplos tipos.

Se `cmdk-sv` for a melhor opção, verificar antes:
- Data do último release (regra: < 3 meses)
- Compatibilidade Svelte 5
- Tamanho do bundle
