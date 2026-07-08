# Design Document

## Overview

Substituir os 3 componentes de configuração de busca (FilterBar, FormAdicionarLoja,
GerenciarBuscas) por um único componente `BuscaUnificada` que integra keywords, seleção
de lojas (plural, multi-marketplace), filtros, fontes e agendamento. O componente fica
no topo da página, é colapsável, executa buscas em real-time e permite salvar a
configuração como busca persistente (com ou sem agendamento).

## Components and Interfaces

### Frontend — Componente BuscaUnificada.svelte

```
┌────────────────────────────────────────────────────────────────────┐
│ [🔍 keywords input________________] [⚙️ Filtros (2)] [💾 Salvar] │
│ [🏪 + Loja A ✕] [🏪 + Loja B ✕] [+ adicionar loja]             │
│ ┌─ Filtros (colapsável) ─────────────────────────────────────────┐│
│ │ Comissão: [≥7% ▾]  Vendas: [___]  Categoria: [cosméticos ✕]  ││
│ └────────────────────────────────────────────────────────────────┘│
│ [🔍 Busca] [📉 Quedas] [🆕 Novos] [⭐ Favoritos] [🏪 Lojas]    │
│ ┌─ Buscas salvas ───────────────────────────────────────────────┐│
│ │ [sérum ⏱] [retinol] [perfume + 2 lojas ⏱] [✕]                ││
│ └────────────────────────────────────────────────────────────────┘│
└────────────────────────────────────────────────────────────────────┘
```

**Modo colapsado (compact summary):**
```
┌────────────────────────────────────────────────────────────────────┐
│ 🔍 "sérum" · 2 lojas · 2 filtros · ⏱ a cada 8h          [▼ abrir]│
└────────────────────────────────────────────────────────────────────┘
```

**Props:**
- `onresultados: (resultados[]) => void` — emite resultados filtrados para o grid
- `oncarregando: (boolean) => void` — indica estado de loading
- `onerro: (Error | null) => void` — erros de API

**Sub-componentes reutilizados (não recriados):**
- `TagInput` — keywords e categorias
- `AgendadorBusca` — cron
- `ToggleGroup` (multiple) — fontes
- `Select` — comissão mín
- `Input` — vendas mín
- `Collapsible` — filtros e salvar
- `Badge` — contagens e indicadores
- `Button` — ações

### Frontend — Módulo busca-unificada-logic.ts

Lógica pura extraída para respeitar o limite de 180 linhas no script block:

```typescript
// Estado serializável da busca (salva/carrega)
export interface BuscaConfig {
  keywords: string[];
  shopIds: long[];
  shopNomes: Record<string, string>; // shopId → nome resolvido
  comissaoMin: number;
  vendasMin: number;
  categorias: string[];
  fontes: string[];              // ['curadoria','quedas','novos','lojas','favoritos']
  cron: string | null;
  marketplaces: string[];        // ['shopee','amazon','mercadolivre']
}

// Converte BuscaConfig para o payload do POST /api/buscas
export function configToPayload(config: BuscaConfig): SyncBuscaRequest { ... }

// Converte response do GET /api/buscas para BuscaConfig
export function payloadToConfig(busca: any): BuscaConfig { ... }

// Gera compact summary para modo colapsado
export function gerarResumo(config: BuscaConfig): string { ... }
```

### Backend — SyncBuscaRequest (estendido)

```csharp
public sealed record SyncBuscaRequest
{
    public string? Id { get; init; }
    public string? Keyword { get; init; }
    public string[]? Keywords { get; init; }
    public long[]? ShopIds { get; init; }          // NOVO
    public string? Cron { get; init; }
    public string? SortBy { get; init; }
    public int? Limit { get; init; }
    public decimal? ComissaoMin { get; init; }     // NOVO
    public int? VendasMin { get; init; }           // NOVO
    public string[]? Categorias { get; init; }     // NOVO
    public string[]? Fontes { get; init; }         // NOVO
    public string? Marketplaces { get; init; }     // NOVO (CSV)
}
```

### Backend — Entidade Busca (estendida)

```csharp
public sealed class Busca : IOwnedEntity
{
    // ... campos existentes ...
    
    // NOVOS (nullable, sem migration de dados)
    public decimal? ComissaoMin { get; set; }
    public int? VendasMin { get; set; }
    public string[]? Categorias { get; set; }
    public string[]? Fontes { get; set; }    // ["curadoria","quedas","novos","lojas","favoritos"]
}
```

### Backend — GET /api/buscas (response estendido)

```json
{
  "buscas": [
    {
      "id": "uuid",
      "keywords": ["sérum", "vitamina c"],
      "shop_ids": [920292999, 123456789],
      "nome": null,
      "cron": "0 */8 * * *",
      "comissao_min": 0.07,
      "vendas_min": 50,
      "categorias": ["cosméticos"],
      "fontes": ["curadoria", "quedas", "lojas"],
      "marketplaces": "shopee",
      "ativo": true,
      "criado_em": "2026-07-08T..."
    }
  ]
}
```

## Data Models

### EF Core Migration (AddFiltersToBusca)

```csharp
migrationBuilder.AddColumn<decimal>("ComissaoMin", "Buscas", nullable: true);
migrationBuilder.AddColumn<int>("VendasMin", "Buscas", nullable: true);
migrationBuilder.AddColumn<string[]>("Categorias", "Buscas", nullable: true);
migrationBuilder.AddColumn<string[]>("Fontes", "Buscas", nullable: true);
```

Todas nullable → zero impacto em registros existentes. Sem data migration.

### Fluxo de dados: salvar busca

```
BuscaUnificada (frontend)
  │  configToPayload(currentConfig)
  ▼
POST /api/buscas {keywords, shop_ids, cron, comissao_min, vendas_min, categorias, fontes}
  │
  ├─ Se shop_ids presentes: persiste em Busca.ShopIds
  ├─ Se cron presente: SchedulerJobs.RegisterAsync (keyword_search ou shop_collection)
  ├─ Filtros persistidos nos novos campos
  └─ Responde {id, keywords, cron, status: "salva"}
```

### Fluxo de dados: carregar busca salva

```
BuscaUnificada (onMount)
  │  GET /api/buscas
  ▼
Response {buscas: [{id, keywords, shop_ids, cron, comissao_min, ...}]}
  │  payloadToConfig(busca) para cada item
  ▼
Chips de buscas salvas renderizados
  │  (clique)
  ▼
Restaura estado: busca, lojas, filtros, fontes, cron
  │  (debounce 400ms)
  ▼
Executa busca real-time com a config restaurada
```

## Error Handling

| Cenário | Comportamento |
|---------|--------------|
| Resolver loja falha | Mensagem inline no campo de loja, mantém input para correção |
| Salvar busca falha | Alert de erro, não perde estado |
| Carregar buscas falha | Usa cache local (mesmo pattern existente) |
| API timeout (>25s) | Mensagem "demorou demais" + botão retry |
| Scheduler indisponível ao salvar | Busca persiste no PG, job reconciliado depois |

## Testing Strategy

| Camada | O que testar |
|--------|-------------|
| Unit (vitest) | `configToPayload`, `payloadToConfig`, `gerarResumo` |
| Unit (vitest) | `montarResultados` com novos filtros (backward compat) |
| Component (vitest + testing-library) | BuscaUnificada renderiza sub-componentes |
| E2E (Playwright) | Fluxo: digitar → filtrar → salvar → carregar |
| Backend (xUnit) | POST /api/buscas com novos campos persiste corretamente |

## Summary of Changes

| Arquivo | Tipo | Descrição |
|---------|------|-----------|
| `web/src/lib/components/BuscaUnificada.svelte` | **Novo** | Componente unificado |
| `web/src/lib/busca-unificada-logic.js` | **Novo** | Lógica pura (< 180 linhas no .svelte) |
| `web/src/routes/+page.svelte` | Refactor | Remove FilterBar/FormAdicionarLoja/GerenciarBuscas, usa BuscaUnificada |
| `web/src/lib/components/FilterBar.svelte` | **Removido** | Absorvido por BuscaUnificada |
| `web/src/lib/components/FormAdicionarLoja.svelte` | **Removido** | Absorvido (campo de loja integrado) |
| `web/src/lib/components/GerenciarBuscas.svelte` | **Removido** | Absorvido (buscas salvas integradas) |
| `web/src/lib/components/PainelNovidades.svelte` | Mantido | Exibido ao selecionar busca com resultados |
| `src/Garimpei.Domain/Entities/Busca.cs` | Ajuste | Novos campos nullable |
| `src/Garimpei.Api/Endpoints/BuscasEndpoints.cs` | Ajuste | Aceita e retorna novos campos |
| `src/Garimpei.Infrastructure/Persistence/Migrations/...` | **Novo** | AddFiltersToBusca |
| `web/src/tests/busca-unificada.test.js` | **Novo** | Testes das funções de lógica |

## Non-goals

- Não muda o Scheduler Go (ele já suporta shop_collection e keyword_search)
- Não muda o Collector (ResolveShop e Fetch já existem)
- Não muda o Analyzer Python (queries BigQuery não dependem de filtros da busca)
- Não muda a lógica de `montarResultados` (filtros client-side permanecem)
- Multi-marketplace (Amazon, ML) é suportado pela entidade mas os collectors não existem ainda — campo `Marketplaces` é preparação (ADR-0016)
