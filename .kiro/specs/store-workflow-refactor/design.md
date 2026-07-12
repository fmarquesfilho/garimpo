# Design Document

## Overview

Este documento descreve o design técnico para a refatoração do subsistema de lojas (store workflow)
da página Descobrir do Garimpei. A refatoração introduz uma entidade `Loja` independente no backend
(PostgreSQL via EF Core), elimina a derivação de lojas a partir de buscas salvas, e reestrutura
o frontend para usar estados de resolução explícitos com normalização robusta de nomes.

**Princípios de design:**
- A BuscaEngine (FSM) permanece estruturalmente intacta — as mudanças afetam handlers internos e o shape do ctx.
- O Omnibox.svelte não muda de interface — apenas consome dados de forma diferente (lojasDisponiveis vem da API).
- Separação clara: Loja é entidade de primeira classe, não embedded em Busca.
- Marketplace é sempre explícito e rastreável — zero fallbacks silenciosos.

## Architecture

### Diagrama de Componentes (Antes vs. Depois)

```
ANTES:
┌───────────────────────────────────────────────────────────────────┐
│ Frontend                                                           │
│                                                                    │
│  Omnibox → BuscaEngine.send(ADICIONAR_LOJA)                       │
│                │                                                   │
│                ├─ event.loja?.id → #adicionarLojaMonitorada (sync) │
│                │   (fonte: lojasDisponiveis = derivado de buscas)  │
│                │                                                   │
│                └─ event.value → #adicionarLoja (async)             │
│                    ctx.lojaResolvendo = true/false (flat bool)     │
│                    ctx.lojaErro = '' | 'msg' (flat string)         │
│                                                                    │
│  listarLojasMonitoradas(buscasSalvas) → lojas para autocomplete   │
│  encontrarLojaPorNome: nome.includes(q) (simple)                  │
└───────────────────────────────────────────────────────────────────┘
         │ POST /api/lojas {input}
         ▼
┌──────────────────────────────────────────────────────────────┐
│ Backend (C# API)                                              │
│  POST /api/lojas → Collector.ResolveShop → cria Busca entity │
│  (Loja = Busca com ShopIds, não entidade separada)           │
└──────────────────────────────────────────────────────────────┘
```

```
DEPOIS:
┌──────────────────────────────────────────────────────────────────────────┐
│ Frontend                                                                  │
│                                                                           │
│  Omnibox → BuscaEngine.send(ADICIONAR_LOJA)                              │
│                │                                                          │
│                ├─ event.loja?.id → #adicionarLojaConhecida (sync)         │
│                │   (fonte: ctx.lojasDisponiveis = GET /api/lojas/registro)│
│                │                                                          │
│                └─ event.value → #resolverLojaRemota (async)               │
│                    ctx.resolucaoLoja = {status, input?, erro?} (typed)    │
│                                                                           │
│  loja-registry.js (novo) — normalizarNome, matchLojas                    │
│  omnibox-sugestoes: matchLojas usa Nome_Normalizado                      │
└──────────────────────────────────────────────────────────────────────────┘
         │ POST /api/lojas/resolver {input, marketplace?}
         │ GET  /api/lojas/registro
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Backend (C# API)                                                     │
│  Entidade Loja (nova tabela) — shopId, nome, marketplace, cron, etc │
│  POST /api/lojas/resolver → Collector.ResolveShop → upsert Loja     │
│  GET /api/lojas/registro → lista todas as Lojas do tenant           │
│  Busca.ShopIds continua existindo (referências para o Scheduler)    │
└─────────────────────────────────────────────────────────────────────┘
```

### Fluxo de Dados Principal

```
1. INICIALIZAR
   Frontend: GET /api/lojas/registro → ctx.lojasDisponiveis (fonte primária)
   Frontend: GET /api/buscas → ctx.buscasSalvas (como antes)

2. ADICIONAR_LOJA (loja conhecida, via Omnibox dropdown)
   event.loja.id existe em ctx.lojasDisponiveis
   → caminho síncrono: popula shopIds/shopNomes/shopMeta do registro
   → ctx.resolucaoLoja permanece {status:'idle'}
   → executarBusca()

3. ADICIONAR_LOJA (loja nova, via URL/ID digitado)
   event.value = "https://s.shopee.com.br/..." ou "lebotanic"
   → ctx.resolucaoLoja = {status:'resolvendo', input: event.value}
   → POST /api/lojas/resolver {input, marketplace?}
   → API: Collector.ResolveShop → upsert na tabela Lojas → retorna Loja
   → ctx.resolucaoLoja = {status:'idle'}
   → popula shopIds/shopNomes/shopMeta
   → executarBusca()
```

## Components and Interfaces

### 1. Backend — Entidade Loja (Domain)

```csharp
// src/Garimpei.Domain/Entities/Loja.cs
namespace Garimpei.Domain.Entities;

public sealed class Loja : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string OwnerUid { get; set; }

    /// <summary>ID numérico da loja no marketplace (ex: 920292999).</summary>
    public required long ShopId { get; init; }

    /// <summary>Nome oficial retornado pelo marketplace.</summary>
    public required string Nome { get; set; }

    /// <summary>Nome normalizado para matching (lowercase, sem espaços/acentos, [a-z0-9]).</summary>
    public required string NomeNormalizado { get; set; }

    /// <summary>Marketplace obrigatório (shopee, mercado_livre, amazon).</summary>
    public required string Marketplace { get; init; }

    /// <summary>Cron de coleta. Null = loja escopada (sem monitoramento).</summary>
    public string? CronExpression { get; set; }

    /// <summary>URL original usada na resolução (preserva tracking de afiliado).</summary>
    public string? SourceUrl { get; set; }

    /// <summary>Origem geográfica padrão (ex: 🇰🇷, 🇧🇷).</summary>
    public string? OrigemPadrao { get; set; }

    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>Gera NomeNormalizado a partir do Nome canônico.</summary>
    public static string Normalizar(string nome)
    {
        // NFD decompose → remove combining marks → keep [a-z0-9] → lowercase
        var normalized = nome.Normalize(System.Text.NormalizationForm.FormD);
        var sb = new System.Text.StringBuilder(normalized.Length);
        foreach (var c in normalized)
        {
            var cat = System.Globalization.CharUnicodeInfo.GetUnicodeCategory(c);
            if (cat == System.Globalization.UnicodeCategory.NonSpacingMark) continue;
            if (char.IsAsciiLetterOrDigit(c)) sb.Append(char.ToLowerInvariant(c));
        }
        return sb.ToString();
    }
}
```

### 2. Backend — Persistência (AppDbContext)

```csharp
// Adição em AppDbContext.OnModelCreating:
modelBuilder.Entity<Loja>(entity =>
{
    entity.HasKey(e => e.Id);
    entity.HasIndex(e => e.OwnerUid);
    entity.HasIndex(e => new { e.ShopId, e.Marketplace, e.OwnerUid }).IsUnique();
    entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
    entity.Property(e => e.NomeNormalizado).HasMaxLength(200);
    entity.Property(e => e.Nome).HasMaxLength(200);
    entity.Property(e => e.Marketplace).HasMaxLength(50);
});

// Novo DbSet:
public DbSet<Loja> Lojas => Set<Loja>();
```

**Migration:** `AddLojaEntity` — cria tabela `Lojas` com índice único composto `(ShopId, Marketplace, OwnerUid)`.

### 3. Backend — Endpoints Novos

```csharp
// Novos endpoints no LojasEndpoints.cs (refatorado):

// GET /api/lojas/registro — lista todas as lojas do tenant (para autocomplete)
app.MapGet("/api/lojas/registro", async (AppDbContext db, CancellationToken ct) =>
{
    var lojas = await db.Lojas
        .OrderBy(l => l.Nome)
        .ToListAsync(ct);

    return Results.Ok(new
    {
        lojas = lojas.Select(l => new
        {
            id = l.ShopId.ToString(),
            nome = l.Nome,
            nome_normalizado = l.NomeNormalizado,
            marketplace = l.Marketplace,
            monitorada = l.CronExpression != null,
            cron = l.CronExpression,
            origem = l.OrigemPadrao
        })
    });
}).RequireAuthorization().WithTags("Lojas");

// POST /api/lojas/resolver — resolve URL/ID e persiste no registro
// Retorna a loja registrada (upsert por ShopId+Marketplace+OwnerUid)
app.MapPost("/api/lojas/resolver", async (
    AppDbContext db,
    CollectorServiceClient collectorClient,
    ResolverLojaRequest req,
    CancellationToken ct) =>
{
    // 1. Resolve via Collector gRPC
    var marketplace = MapMarketplace(req.Marketplace ?? "shopee");
    var resolveResp = await collectorClient.ResolveShopAsync(
        new ResolveShopRequest { UsernameOrUrl = req.Input, Marketplace = marketplace },
        cancellationToken: ct);

    if (resolveResp.ShopId <= 0)
        return Results.BadRequest(new { error = "Loja não encontrada." });

    var mktStr = req.Marketplace ?? "shopee";

    // 2. Upsert na tabela Lojas
    var existente = await db.Lojas.FirstOrDefaultAsync(
        l => l.ShopId == resolveResp.ShopId && l.Marketplace == mktStr, ct);

    if (existente is not null)
    {
        existente.Nome = resolveResp.ShopName;
        existente.NomeNormalizado = Loja.Normalizar(resolveResp.ShopName);
        existente.UpdatedAt = DateTime.UtcNow;
        if (req.Input.StartsWith("http", StringComparison.OrdinalIgnoreCase))
            existente.SourceUrl = req.Input;
    }
    else
    {
        existente = new Loja
        {
            OwnerUid = "", // preenchido por SaveChangesAsync
            ShopId = resolveResp.ShopId,
            Nome = resolveResp.ShopName,
            NomeNormalizado = Loja.Normalizar(resolveResp.ShopName),
            Marketplace = mktStr,
            SourceUrl = req.Input.StartsWith("http") ? req.Input : null,
            OrigemPadrao = req.Origem
        };
        db.Lojas.Add(existente);
    }

    await db.SaveChangesAsync(ct);

    return Results.Ok(new
    {
        id = existente.ShopId.ToString(),
        nome = existente.Nome,
        nome_normalizado = existente.NomeNormalizado,
        marketplace = existente.Marketplace,
        monitorada = existente.CronExpression != null,
        cron = existente.CronExpression,
        origem = existente.OrigemPadrao
    });
}).RequireAuthorization().WithTags("Lojas");
```

```csharp
public sealed record ResolverLojaRequest
{
    public required string Input { get; init; }
    public string? Marketplace { get; init; }
    public string? Origem { get; init; }
}
```

### 4. Frontend — Módulo `loja-registry.js` (novo)

Módulo puro (sem runes, sem DOM) que encapsula a lógica de normalização e matching
de lojas no frontend. Espelha o `Loja.Normalizar()` do backend.

```javascript
// web/src/lib/loja-registry.js

/**
 * Normaliza nome para matching: NFD → remove diacríticos → remove non-[a-z0-9] → lowercase.
 * Idêntica à implementação C# `Loja.Normalizar()`.
 * @param {string} nome
 * @returns {string} nome normalizado (apenas [a-z0-9])
 */
export function normalizarNome(nome) {
    if (!nome) return '';
    return nome
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')  // remove combining marks
        .replace(/[^a-z0-9]/gi, '')        // keep only alphanum
        .toLowerCase();
}

/**
 * Faz matching de um input (já normalizado) contra a lista de lojas do registro.
 * Retorna lojas que casam por substring no NomeNormalizado OU no Nome canônico (lowercase).
 *
 * @param {string} inputNormalizado — input do usuário já normalizado
 * @param {Array<{id:string, nome:string, nome_normalizado:string, marketplace:string}>} lojas
 * @param {number} [max=7]
 * @returns {Array} lojas que casam
 */
export function matchLojas(inputNormalizado, lojas, max = 7) {
    if (!inputNormalizado || inputNormalizado.length < 2) return [];
    return lojas
        .filter(l =>
            l.nome_normalizado.includes(inputNormalizado) ||
            l.nome.toLowerCase().includes(inputNormalizado)
        )
        .slice(0, max);
}
```

### 5. Frontend — Novo Estado de Resolução (`busca-engine-state.js`)

Substitui `lojaResolvendo: boolean` e `lojaErro: string` por um sub-estado tipado:

```javascript
// Em criarContextoInicial():
export function criarContextoInicial() {
    return {
        // ... campos existentes ...
        lojasDisponiveis: [],  // agora populado via GET /api/lojas/registro

        // NOVO: sub-estado de resolução (substitui lojaResolvendo + lojaErro)
        resolucaoLoja: { status: 'idle' },
        // Formato: { status: 'idle' }
        //        | { status: 'resolvendo', input: string }
        //        | { status: 'sucesso' }  (transiciona para idle imediatamente)
        //        | { status: 'erro', input: string, erro: string }

        // REMOVIDOS: lojaResolvendo, lojaErro
        // ... demais campos ...
    };
}
```

### 6. Frontend — Refatoração do `#adicionarLoja` na BuscaEngine

```javascript
// busca-engine.svelte.js — novos handlers

/** Caminho síncrono: loja conhecida (do registro). Req 6.1 */
#adicionarLojaConhecida(loja) {
    const { id, nome, nome_normalizado, marketplace, monitorada, cron, origem } = loja;
    if (this.ctx.shopIds.includes(id)) return;

    this.ctx.shopIds = [...this.ctx.shopIds, id];
    this.ctx.shopNomes = { ...this.ctx.shopNomes, [id]: nome };
    this.ctx.shopMeta = {
        ...this.ctx.shopMeta,
        [id]: {
            marketplace,
            origem: origem ?? null,
            monitorada: Boolean(monitorada),
            cron: cron ?? '',
            tipo: monitorada ? 'monitorada' : 'escopada'  // Req 2
        }
    };
    return this.#executarBusca();
}

/** Caminho assíncrono: loja desconhecida, resolve via API. Req 4, 6.2 */
async #resolverLojaRemota(input) {
    // Guard: já resolvendo (Req 4.5)
    if (this.ctx.resolucaoLoja.status === 'resolvendo') return;

    // Guard: input vazio (Req 4.9)
    if (!input?.trim()) return;

    this.ctx.resolucaoLoja = { status: 'resolvendo', input };

    try {
        const r = await Promise.race([
            this.#effects.resolverLoja(input),
            new Promise((_, rej) =>
                setTimeout(() => rej(new Error('Timeout: loja não respondeu em 10s')), 10000)
            )
        ]);

        // Validar marketplace (Req 5.4, 5.6)
        if (!r.marketplace || !MARKETPLACES_SUPORTADOS.includes(r.marketplace)) {
            this.ctx.resolucaoLoja = {
                status: 'erro',
                input,
                erro: `Marketplace "${r.marketplace ?? '?'}" não suportado.`
            };
            return;
        }

        // Sucesso: adiciona ao contexto
        this.ctx.shopIds = [...this.ctx.shopIds, r.id];
        this.ctx.shopNomes = { ...this.ctx.shopNomes, [r.id]: r.nome };
        this.ctx.shopMeta = {
            ...this.ctx.shopMeta,
            [r.id]: {
                marketplace: r.marketplace,
                origem: r.origem ?? null,
                monitorada: Boolean(r.monitorada),
                cron: r.cron ?? '',
                tipo: r.monitorada ? 'monitorada' : 'escopada'
            }
        };

        // Atualiza lojasDisponiveis com a nova loja (Req 1.5)
        if (!this.ctx.lojasDisponiveis.find(l => l.id === r.id)) {
            this.ctx.lojasDisponiveis = [...this.ctx.lojasDisponiveis, r];
        }

        this.ctx.resolucaoLoja = { status: 'idle' };  // Req 4.8
        await this.#executarBusca();
    } catch (e) {
        this.ctx.resolucaoLoja = {
            status: 'erro',
            input,
            erro: e?.message ?? 'Falha ao resolver loja'
        };
    }
}

/** Handler público: roteia para o caminho correto (Req 6) */
async #adicionarLoja(event) {
    // Caminho 1: loja já conhecida (seleção do dropdown)
    if (event.loja?.id) {
        return this.#adicionarLojaConhecida(event.loja);
    }

    // Caminho 2: input textual (Req 6.2)
    if (typeof event.value === 'string' && event.value.trim()) {
        // Checar se o input casa com loja existente no registro (Req 4.10)
        const match = this.ctx.lojasDisponiveis.find(l =>
            l.id === event.value || l.nome_normalizado === normalizarNome(event.value)
        );
        if (match) return this.#adicionarLojaConhecida(match);

        return this.#resolverLojaRemota(event.value);
    }

    // Payload inválido (Req 6.3)
    this.ctx.resolucaoLoja = {
        status: 'erro',
        input: event.value ?? '',
        erro: 'Evento ADICIONAR_LOJA inválido: informe loja.id ou value.'
    };
}
```

### 7. Frontend — Atualização do `omnibox-sugestoes.js`

A função `matchLojas` passa a usar normalização robusta em vez de `.includes()`:

```javascript
// omnibox-sugestoes.js — matchLojas refatorado
import { normalizarNome, matchLojas as matchLojasRegistry } from './loja-registry.js';

function matchLojas(lojas, q, max) {
    const qNorm = normalizarNome(q);
    return matchLojasRegistry(qNorm, lojas ?? [], max)
        .map(l => ({
            tipo: 'loja',
            label: l.nome,
            valor: '@' + l.nome,
            icone: ICONE.loja,
            meta: l
        }));
}
```

Isso resolve o bug central: `@gloryofseoul` → `normalizarNome('gloryofseoul')` = `"gloryofseoul"` →
substring match em `nome_normalizado` de "Glory of Seoul" = `"gloryofseoul"` ✓.

### 8. Frontend — Atualização do `busca-engine-effects.js`

```javascript
// busca-engine-effects.js — novos/alterados effects

export function criarEffects({ getBuscasSalvas, getFavoritos, sincronizarStore }) {
    return {
        // NOVO: carrega registro de lojas da API
        async carregarRegistroLojas() {
            const resp = await pegar('/api/lojas/registro');
            return resp?.lojas ?? [];
        },

        // ALTERADO: resolve via novo endpoint (retorna Loja completa)
        async resolverLoja(input) {
            return postar('/api/lojas/resolver', { input });
        },

        // REMOVIDO: listarLojasMonitoradas() — substituído por carregarRegistroLojas()
        // ... demais effects inalterados ...
    };
}
```

### 9. Frontend — Inicialização da Engine (refatorada)

```javascript
async #inicializar() {
    this.status = STATES.SEARCHING;
    try {
        const [buscas, categorias, lojas] = await Promise.all([
            this.#effects.carregarBuscasSalvas(),
            this.#effects.carregarCategorias(),
            this.#effects.carregarRegistroLojas()  // NOVO
        ]);
        this.ctx.buscasSalvas = (buscas ?? []).map(payloadToConfig);
        this.ctx.categoriasDisponiveis = categorias ?? [];
        this.ctx.lojasDisponiveis = lojas ?? [];  // fonte primária: API

        await this.#effects.sincronizarStoreExterno();

        if (this.ctx.keyword.trim() || this.ctx.shopIds.length > 0 || this.ctx.categorias.length > 0) {
            await this.#executarBusca();
        } else {
            this.status = STATES.IDLE;
        }
    } catch (e) {
        this.ctx.error = e?.message ?? 'Falha ao inicializar';
        this.status = STATES.ERROR;
    }
}
```

### 10. Frontend — Distinção Visual Monitorada vs. Escopada

O getter `lojaCards` na engine já retorna `shopMeta[id]`. A view usa o novo campo `tipo`:

```javascript
// Em BuscaEngine:
get lojaCards() {
    return this.ctx.shopIds.map(id => ({
        id,
        nome: this.ctx.shopNomes[id] || id,
        ...(this.ctx.shopMeta[id] ?? {
            marketplace: 'shopee', origem: null,
            monitorada: false, cron: '', tipo: 'escopada'
        })
    }));
}
```

Na view (BuscaUnificada.svelte), o card de loja mostra badge:

```svelte
{#each engine.lojaCards as loja}
    <div class="flex items-center gap-1.5 rounded bg-muted/50 px-2 py-1 text-sm">
        <span>{loja.nome}</span>
        {#if loja.tipo === 'monitorada'}
            <span class="text-xs text-green-600" title="Coleta automática: {loja.cron}">⏱</span>
        {/if}
        <button onclick={() => engine.send({type:'REMOVER_LOJA', shopId: loja.id})}>×</button>
    </div>
{/each}
```

### 11. Contrato da Resposta `/api/lojas/registro`

```json
{
    "lojas": [
        {
            "id": "920292999",
            "nome": "Glory of Seoul",
            "nome_normalizado": "gloryofseoul",
            "marketplace": "shopee",
            "monitorada": true,
            "cron": "0 */8 * * *",
            "origem": "🇰🇷"
        },
        {
            "id": "282170857",
            "nome": "Le Botanic",
            "nome_normalizado": "lebotanic",
            "marketplace": "shopee",
            "monitorada": true,
            "cron": "0 */6 * * *",
            "origem": null
        }
    ]
}
```

### 12. Contrato da Resposta `/api/lojas/resolver`

```json
{
    "id": "920292999",
    "nome": "Glory of Seoul",
    "nome_normalizado": "gloryofseoul",
    "marketplace": "shopee",
    "monitorada": false,
    "cron": null,
    "origem": "🇰🇷"
}
```

### 13. Atualização das Regras (`rules/busca-rules.json`)

Adição de marketplaces suportados como array validável:

```json
{
    "marketplaces": {
        "suportados": ["shopee", "mercado_livre", "amazon"],
        ...
    },
    "lojaRegistro": {
        "normalizacao": "NFD_strip_nonalphanum_lowercase",
        "matchMinChars": 2,
        "fonte": "api"
    }
}
```

### 14. Guard de Marketplace Explícito

No `busca-engine-state.js`, o guard `lojaInputValida` é ampliado:

```javascript
export const guards = {
    // ... existentes ...
    lojaInputValida: (_ctx, event) => {
        // Path 1: loja conhecida (precisa ter marketplace)
        if (event.loja?.id) return Boolean(event.loja.marketplace);
        // Path 2: input textual
        return (event.value ?? '').trim().length > 0;
    },
    resolucaoPermitida: (ctx) => ctx.resolucaoLoja.status !== 'resolvendo'
};
```

## Data Models

### Tabela `Lojas` (PostgreSQL)

| Coluna | Tipo | Constraints |
|--------|------|-------------|
| Id | uuid | PK, default newid |
| OwnerUid | varchar | NOT NULL, indexed |
| ShopId | bigint | NOT NULL |
| Nome | varchar(200) | NOT NULL |
| NomeNormalizado | varchar(200) | NOT NULL |
| Marketplace | varchar(50) | NOT NULL |
| CronExpression | varchar(50) | nullable |
| SourceUrl | text | nullable |
| OrigemPadrao | varchar(20) | nullable |
| CreatedAt | timestamp | NOT NULL, default now() |
| UpdatedAt | timestamp | NOT NULL, default now() |

**Índice único:** `(ShopId, Marketplace, OwnerUid)` — garante no-duplication por tenant.

### Contexto da BuscaEngine (frontend, shape parcial)

```javascript
{
    shopIds: string[],           // IDs de lojas no escopo atual
    shopNomes: Record<string, string>,  // id → nome canônico
    shopMeta: Record<string, {
        marketplace: string,
        origem: string | null,
        monitorada: boolean,
        cron: string,
        tipo: 'monitorada' | 'escopada'
    }>,
    lojasDisponiveis: Array<{    // do GET /api/lojas/registro
        id: string,
        nome: string,
        nome_normalizado: string,
        marketplace: string,
        monitorada: boolean,
        cron: string | null,
        origem: string | null
    }>,
    resolucaoLoja:
        | { status: 'idle' }
        | { status: 'resolvendo', input: string }
        | { status: 'erro', input: string, erro: string }
}
```

## Decisões de Design

| # | Decisão | Justificativa |
|---|---------|---------------|
| 1 | Loja como entidade separada (não embedded em Busca) | Desacopla lifecycle: loja persiste independente de buscas. Permite autocomplete sem buscas salvas. |
| 2 | Chave composta (ShopId + Marketplace + OwnerUid) | Um shopId pode existir em marketplaces diferentes. Multi-tenant por design. |
| 3 | `NomeNormalizado` computado no backend E espelhado no frontend | Garante consistency: mesmo algoritmo em C# e JS. O backend é source of truth; frontend replica para UX instantânea. |
| 4 | Sub-estado tipado em vez de flags | Elimina estados inválidos (ex: `lojaResolvendo=true` + `lojaErro='msg'`). Pattern amplamente testável. |
| 5 | Timeout de 10s hardcoded no frontend | A resolução chama Collector gRPC que pode travar. 10s é generoso para ResolveShop (tipicamente <2s). |
| 6 | POST /api/lojas/resolver faz upsert | Evita duplicatas. Se a loja já existe, atualiza o nome (marketplaces mudam nomes de lojas). |
| 7 | GET /api/lojas/registro carregado no INICIALIZAR em paralelo | Não adiciona latência extra (já faz 2 calls em paralelo; vira 3). |
| 8 | `loja-registry.js` como módulo puro separado | Testável em isolamento (Vitest, sem DOM). Reusável pelo omnibox-sugestoes e pelo omnibox-parser. |
| 9 | Campo `tipo` ('monitorada'/'escopada') no shopMeta | Permite a view renderizar diferente sem lógica extra. Derivado do cron no momento da adição. |
| 10 | Eliminar `listarLojasMonitoradas()` de descobrir-logic.js | Função legada que derivava lojas de buscas. Substituída por dados do registro via API. |

## Impacto em Arquivos Existentes

| Arquivo | Mudança |
|---------|---------|
| `src/Garimpei.Domain/Entities/Loja.cs` | **NOVO** — entidade de domínio |
| `src/Garimpei.Infrastructure/Persistence/AppDbContext.cs` | Adicionar DbSet + config |
| `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` | Refatorar: novos endpoints registro/resolver |
| `web/src/lib/loja-registry.js` | **NOVO** — normalização + matching |
| `web/src/lib/busca-engine.svelte.js` | Refatorar `#adicionarLoja` → 2 caminhos explícitos |
| `web/src/lib/busca-engine-state.js` | Substituir `lojaResolvendo`/`lojaErro` por `resolucaoLoja` |
| `web/src/lib/busca-engine-effects.js` | Novo effect `carregarRegistroLojas`, remover `listarLojasMonitoradas` |
| `web/src/lib/omnibox-sugestoes.js` | `matchLojas` usa `loja-registry.js` |
| `web/src/lib/descobrir-logic.js` | Remover `listarLojasMonitoradas` (+ `encontrarLojaPorNome`) |
| `web/src/lib/api.js` | Adicionar `listarRegistroLojas()`, alterar `adicionarLoja` → `resolverLoja` |
| `web/src/lib/components/BuscaUnificada.svelte` | Loja cards com badge monitorada/escopada |
| `rules/busca-rules.json` | Adicionar bloco `lojaRegistro` |
| `contracts/schemas/lojas.request.json` | Atualizar para novo formato resolver |
| `contracts/schemas/lojas.response.json` | Atualizar para formato do registro |

## Rastreabilidade de Requisitos

| Requisito | Componente(s) de Design |
|-----------|------------------------|
| Req 1 — Registro Independente | §1 (Entidade Loja), §2 (Persistência), §3 (GET /api/lojas/registro), §9 (Inicialização) |
| Req 2 — Monitorada vs. Escopada | §6 (campo `tipo` no shopMeta), §10 (badge visual) |
| Req 3 — Normalização e Matching | §1 (`Loja.Normalizar`), §4 (`loja-registry.js`), §7 (omnibox-sugestoes) |
| Req 4 — Estados de Resolução | §5 (`resolucaoLoja` sub-estado), §6 (`#resolverLojaRemota`) |
| Req 5 — Scoping de Marketplace | §3 (campo obrigatório), §6 (validação), §14 (guard) |
| Req 6 — Separação de Caminhos | §6 (`#adicionarLojaConhecida` vs `#resolverLojaRemota`), §14 (guards) |

## Riscos e Mitigações

| Risco | Probabilidade | Mitigação |
|-------|---------------|-----------|
| Divergência entre normalização C# e JS | Baixa | Implementação espelhada + test parametrizado com os mesmos inputs nos dois lados |
| Latência extra do GET /api/lojas/registro na inicialização | Baixa | Carregado em paralelo (Promise.all) — não adiciona latência ao critical path |
| Testes existentes quebram com remoção de `listarLojasMonitoradas` | Média | Migrar testes para usar mock de `carregarRegistroLojas`; testes de `descobrir-logic.js` que testam a função removida são deletados/reescritos |
| Tabela Lojas vazia no primeiro uso (sem buscas salvas para popular) | N/A | Registro vazio é válido (Req 1 critério 4). Lojas populam conforme o usuário resolve novas. |

## Error Handling

| Cenário | Comportamento | UX |
|---------|---------------|----|
| Collector gRPC `NotFound` / `InvalidArgument` | `resolucaoLoja.erro` = "Loja não encontrada ou link inválido" | Toast + campo com borda vermelha |
| Collector gRPC `Unimplemented` | `resolucaoLoja.erro` = "Resolução não suportada para este marketplace" | Toast informativo |
| Collector timeout (>10s) | `resolucaoLoja.erro` = "Timeout: loja não respondeu em 10s" | Toast + botão retry |
| Marketplace inválido/ausente na resposta | `resolucaoLoja.erro` = "Marketplace não suportado" | Rejeita adição |
| GET /api/lojas/registro falha na inicialização | `ctx.lojasDisponiveis = []` (graceful degradation) | Autocomplete vazio mas funcional |
| Evento ADICIONAR_LOJA durante resolução em andamento | Evento descartado silenciosamente | Input permanece; loading spinner visível |
| Payload inválido (sem loja.id nem value) | `resolucaoLoja.erro` descritivo | Sem loading, erro imediato |

## Correctness Properties

### Property 1: Idempotência de resolução
Resolver a mesma loja N vezes produz exatamente 1 registro na tabela Lojas (upsert por chave composta).

**Validates: Requirements 1.2, 1.5**

### Property 2: Consistência de normalização
`normalizarNome(x)` em JavaScript === `Loja.Normalizar(x)` em C# para todo input UTF-8. Validado por testes parametrizados cross-stack.

**Validates: Requirements 3.1, 3.6**

### Property 3: Invariante de shopMeta
Para todo `id` em `ctx.shopIds`, existe `ctx.shopNomes[id]` (string não-vazia) e `ctx.shopMeta[id]` com `marketplace` definido.

**Validates: Requirements 5.5, 6.4**

### Property 4: Estado de resolução bem-formado
Em qualquer instante, `ctx.resolucaoLoja` é exatamente um dos 3 shapes tipados (idle, resolvendo, erro). Não existem combinações inválidas.

**Validates: Requirements 4.7**

### Property 5: Guard de concorrência
Durante `status === 'resolvendo'`, é impossível disparar nova resolução. A engine descarta o evento.

**Validates: Requirements 4.5**

### Property 6: Unicidade no contexto
`ctx.shopIds` nunca contém duplicatas (checked antes de push em ambos os caminhos).

**Validates: Requirements 1.2, 6.4**

## Testing Strategy

| Camada | Tipo | Escopo |
|--------|------|--------|
| `loja-registry.js` | Unit (Vitest) | `normalizarNome`: pares parametrizados (nomes com espaço, acentos, emojis, etc.). `matchLojas`: substring matching, max, empty input. |
| `busca-engine.svelte.js` | Unit (Vitest) | ADICIONAR_LOJA: caminho sync (loja conhecida), caminho async (mock resolverLoja), timeout, guard de concorrência, payload inválido, marketplace rejection. |
| `omnibox-sugestoes.js` | Unit (Vitest) | matchLojas com lojasDisponiveis usando nome_normalizado. Edge cases: "@glory" match "Glory of Seoul", "@gloryofseoul" idem. |
| `Loja.Normalizar()` | Unit (xUnit) | Mesmos pares parametrizados do frontend (cross-validation). |
| `LojasEndpoints` | Integration (xUnit + TestServer) | POST /api/lojas/resolver mock do Collector → verify upsert. GET /api/lojas/registro → formato de resposta. |
| `EF Core Migration` | Smoke | Verify migration applies cleanly em banco local. |
| Cross-stack normalization | Contract test | JSON fixture com pares `{input, expected}` consumida tanto por Vitest quanto por xUnit. |
