# Design Document

## Overview

Conectar o fluxo de ponta a ponta para que uma busca por palavras-chave criada na UI
seja efetivamente agendada no Scheduler e seus resultados visíveis ao usuário. O backend
(C# API + Scheduler Go) já está implementado; o gap está no frontend que precisa:
1. Enviar o campo `cron` corretamente ao backend via `sincronizarBusca`
2. Exibir o estado do agendamento e permitir visualizar novidades coletadas

## Architecture

### Diagrama de sequência — Criação de busca agendada

```
┌──────────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  GerenciarBuscas │     │  api.js      │     │  C# API      │     │  Scheduler   │
│  (Svelte)        │     │  (frontend)  │     │  /api/buscas  │     │  (Go gRPC)   │
└────────┬─────────┘     └──────┬───────┘     └──────┬───────┘     └──────┬───────┘
         │                       │                     │                     │
         │ salvar({keywords,cron})                     │                     │
         ├──────────►            │                     │                     │
         │           sincronizarBusca(busca)           │                     │
         │           ──────────────────────►           │                     │
         │                       │    POST /api/buscas │                     │
         │                       │    {keywords,cron}  │                     │
         │                       │    ─────────────────►                     │
         │                       │                     │  SetSchedule(       │
         │                       │                     │   type=keyword_search,
         │                       │                     │   keywords, cron)   │
         │                       │                     ├────────────────────►│
         │                       │                     │                     │
         │                       │                     │◄───── OK ──────────┤
         │                       │◄──── {id, cron}─────┤                     │
         │◄──── update store ────┤                     │                     │
```

### Diagrama de sequência — Visualização de resultados

```
┌──────────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  GerenciarBuscas │     │  api.js      │     │  C# API      │     │  Analyzer    │
│  (Svelte)        │     │  (frontend)  │     │  /api/lojas/  │     │  (Python)    │
└────────┬─────────┘     └──────┬───────┘     │  novidades   │     └──────┬───────┘
         │                       │             └──────┬───────┘             │
         │ selecionar busca(id)  │                     │                     │
         ├──────────►            │                     │                     │
         │           buscarNovidades({buscaId})        │                     │
         │           ──────────────────────►           │                     │
         │                       │  GET /api/lojas/    │                     │
         │                       │  novidades?busca_id │                     │
         │                       │  ─────────────────►│                     │
         │                       │                     │  GET /novidades?    │
         │                       │                     │  busca_id=X&dias=7  │
         │                       │                     ├────────────────────►│
         │                       │                     │                     │
         │                       │                     │◄── {produtos_novos, │
         │                       │                     │     variacoes}──────┤
         │                       │◄──── response ──────┤                     │
         │◄──── render cards ────┤                     │                     │
```

### Componentes envolvidos

| Camada | Componente | Estado | Mudança necessária |
|--------|-----------|--------|-------------------|
| Backend C# | `BuscasEndpoints.cs` | ✅ Implementado | Nenhuma — já aceita `Cron`, persiste e chama `SchedulerJobs.RegisterAsync` |
| Backend C# | `SchedulerJobs.cs` | ✅ Implementado | Nenhuma — já monta `keyword_search` para buscas sem ShopIds |
| Backend Go | `scheduler/jobs.go` | ✅ Implementado | Nenhuma — branch `default` em `executeJob` já executa `Fetch(keyword)` |
| Backend C# | `GET /api/lojas/novidades` | ✅ Implementado | Nenhuma — já faz proxy para Analyzer com `busca_id` independente de loja |
| Frontend | `api.js::sincronizarBusca` | ✅ Implementado | Nenhuma — já envia o objeto completo (inclui `cron`) |
| Frontend | `buscas.js::salvar` | ✅ Implementado | Nenhuma — já inclui `cron` no payload |
| Frontend | `GerenciarBuscas.svelte` | ⚠️ Parcial | Precisa: botão "ver resultados" e painel de novidades |
| Frontend | `BuscaCard.svelte` | ⚠️ Parcial | Precisa: indicador de frequência legível e ação de visualizar |

## Detailed Component Design

### 1. GerenciarBuscas.svelte — Painel de resultados

**Situação atual:** O formulário já envia keywords + cron ao backend corretamente.
A listagem mostra buscas do servidor (via `sincronizarDoServidor`). Falta a
visualização de novidades coletadas.

**Design:**

```svelte
<!-- Dentro de GerenciarBuscas, após a lista de BuscaCards -->
{#if buscaSelecionada}
  <PainelNovidades buscaId={buscaSelecionada.id} keywords={buscaSelecionada.keywords} />
{/if}
```

Novo componente `PainelNovidades.svelte`:
- Props: `buscaId: string`, `keywords: string[]`
- Chama `buscarNovidades({ buscaId, dias: 7 })` (já existe em `api.js`)
- Renderiza `ProductCard` (layout "compact") para novos e variações
- Exibe estado vazio se nunca coletou ("Aguardando primeira coleta...")
- Loading state enquanto busca

**Interação:** Clicar em uma keyword dentro do `BuscaCard` seleciona a busca e
abre o painel de resultados abaixo.

### 2. BuscaCard.svelte — Indicador de frequência

**Situação atual:** Já exibe badge `⏱ agendada` quando `busca.cron` é truthy.

**Melhoria:** Mostrar a frequência de forma legível:

```javascript
function cronLabel(cron) {
  if (!cron) return null;
  if (cron === '0 */8 * * *') return 'a cada 8h';
  if (cron === '0 */12 * * *') return 'a cada 12h';
  if (cron === '0 9 * * *') return 'diária 9h';
  return cron; // fallback: mostra o cron raw
}
```

Badge atualizado: `⏱ {cronLabel(busca.cron)}` em vez de apenas `⏱ agendada`.

### 3. Fluxo de dados (stores)

```
buscasSalvas (writable store)
    │
    ├─ .salvar(busca)   → localStorage + sincronizarBusca(POST /api/buscas)
    │                      Backend registra no Scheduler se cron presente
    │
    ├─ .remover(id)     → localStorage + sincronizarBusca(?remover)
    │                      Backend pausa job via SchedulerJobs.PauseAsync
    │
    └─ .sincronizarDoServidor() → GET /api/buscas → substitui store local
                                   (inclui cron e keywords reais do servidor)
```

A store já funciona corretamente. O campo `cron` viaja no payload do `salvar`
→ `sincronizarBusca` → `POST /api/buscas` → backend registra no Scheduler.

### 4. Modelo de dados — payload POST /api/buscas

```typescript
// Request (SyncBuscaRequest no C#)
interface SyncBuscaRequest {
  id?: string;
  keyword?: string;     // keyword única (formato legado)
  keywords?: string[];  // array de keywords (formato preferido)
  cron?: string;        // expressão cron — null/vazio = manual only
  sort_by?: string;
  limit?: number;
}

// Response (sucesso)
interface BuscaCriada {
  id: string;           // UUID da busca persistida
  keywords: string[];   // keywords normalizadas
  cron: string | null;  // cron efetivo (null se manual)
  status: "salva";
}
```

### 5. Scheduler job params — keyword_search

Quando o C# registra um job `keyword_search` via `SchedulerJobs.BuildRequest`:

```json
{
  "job_id": "busca-{uuid}",
  "cron_expression": "0 */8 * * *",
  "enabled": true,
  "params": {
    "owner_uid": "firebase-uid",
    "type": "keyword_search",
    "keywords": "serum,skin1004,vitamina c"
  }
}
```

No Scheduler Go, `dispatchJob` roteia para `executeJob` (não é `shop_collection`
nem `coupon_collection` nem `scheduled_publish`), que cai no branch `default`:

```go
// Branch default em executeJob:
keyword = params["keyword"]  // keyword única (legado)
// OU se o job tem params["keywords"]:
// Scheduler itera sobre cada keyword separada por vírgula
```

**Nota:** O Scheduler hoje no branch `default` usa `params["keyword"]` (singular).
O `SchedulerJobs.BuildRequest` monta `params["keywords"]` (plural). Isso já funciona
porque o `keyword_search` type não é `shop_collection`, então cai no `default`
que lê `params["keyword"]`. Se `params["keyword"]` está vazio, usa `job.name`.

**Ajuste necessário no Scheduler:** O branch `default` precisa ler `params["keywords"]`
(plural, comma-separated) e iterar, como o `shop_collection` faz quando tem keywords.
Atualmente ele só lê `params["keyword"]` (singular). Sem esse ajuste, o job `keyword_search`
com múltiplas keywords só executaria a primeira ou usaria o job name.

### 6. Ajuste no Scheduler Go — suporte a keywords plural no default branch

```go
// Em executeJob, branch default:
default:
    keywords := params["keywords"]
    if keywords != "" {
        // Múltiplas keywords: itera como shop_collection faz
        for _, kw := range strings.Split(keywords, ",") {
            kw = strings.TrimSpace(kw)
            if kw == "" { continue }
            s.logger.Info("executing keyword job", slog.String("job", job.name), slog.String("keyword", kw))
            resp, err := s.collector.Fetch(ctx, &collectorpb.FetchRequest{
                Keyword:     kw,
                Limit:       50,
                Marketplace: collectorpb.Marketplace_MARKETPLACE_SHOPEE,
                OwnerUid:    params["owner_uid"],
            })
            if err != nil {
                s.logger.Error("keyword job falhou", slog.String("keyword", kw), slog.String("erro", err.Error()))
                continue
            }
            totalFound += resp.GetTotalFound()
        }
        keyword = keywords
    } else {
        // Keyword única (legado)
        keyword = params["keyword"]
        if keyword == "" { keyword = job.name }
        // ... fetch como antes
    }
```

## Summary of Changes

| Arquivo | Tipo | Descrição |
|---------|------|-----------|
| `services/scheduler/jobs.go` | Ajuste | Branch `default` ler `params["keywords"]` (plural) e iterar |
| `web/src/lib/components/PainelNovidades.svelte` | Novo | Painel de resultados (novidades + variações) para busca selecionada |
| `web/src/lib/components/GerenciarBuscas.svelte` | Ajuste | Estado `buscaSelecionada` + renderizar `PainelNovidades` |
| `web/src/lib/components/BuscaCard.svelte` | Ajuste | `cronLabel()` legível + ação de selecionar busca |

## Non-goals (já implementado, não precisa mudar)

- POST /api/buscas (já aceita `cron`, já chama `SchedulerJobs.RegisterAsync`)
- GET /api/buscas (já retorna `cron` e `keywords` reais)
- SchedulerJobs.cs (já monta `keyword_search` para buscas sem ShopIds)
- api.js::sincronizarBusca (já envia objeto completo)
- buscas.js::salvar (já inclui `cron`)
- GET /api/lojas/novidades (já funciona com qualquer `busca_id`)
