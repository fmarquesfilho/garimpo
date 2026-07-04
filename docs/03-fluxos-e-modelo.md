# Fluxos e modelo de dados

## Entidades principais

O diagrama ER completo é gerado automaticamente — ver `docs/gerado/ENTIDADES.md`.

### Tabelas BigQuery (dataset `garimpo`)

| Tabela | O que armazena | Partição |
|---|---|---|
| `eventos` | Seleções e publicações de curadoria | `DATE(em)` |
| `snapshots` | Foto periódica dos top N de uma keyword/categoria | `DATE(coletado_em)` |
| `coupon_snapshots` | Cupons coletados (append-only, 90d TTL) | `DATE(collected_at)` |
| `buscas` | Perfis de coleta (filtros + cron + shop_ids), append-only | `DATE(salvo_em)` |
| `conversoes` | Conversões reais da Shopee (conversionReport) | `DATE(compra_em)` |

### Regras de negócio das entidades

**Busca:**
- `keywords[]` (JSON array de termos)
- `shop_ids[]` (JSON array de IDs de lojas)
- `categorias[]` (plural, filtro OR) — ver [ADR 0006](/docs/decisoes/0006-categorias-plural/)
- `cron` vazio = atalho manual (sem agendamento)
- `ativo = false` = tombstone (soft delete append-only)
- `rotation_cursor` = JSON map shopID→próxima página (rotação de catálogo)
- `full_scan_at` = JSON map shopID→timestamp da última varredura completa
- `fontes[]` = tipos de dados monitorados: `curadoria`, `quedas`, `novos`, `favoritos`
- `origem_padrao` = país padrão herdado por todos os produtos da loja

**Evento:**
- `tipo`: `selecao` ou `publicacao`
- `sub_id`: atribuição no formato `canal_estrategia_AAAAMMDD`
- Registrado automaticamente ao garimpar ou publicar

**Snapshot:**
- Posição (`posicao`) indica ranking no dia (1 = topo)
- Usado para detectar novidades e variações de preço (diff entre coletas)

**Conversão:**
- Status: `PENDING` → `COMPLETED` → `PAID` ou `CANCELLED`
- Vinculada a evento via `sub_id` (utm_content do conversionReport)

## Favoritos

Persistência dual:
- **localStorage** para acesso instantâneo (frontend-first)
- **Sync para BigQuery** como backup servidor

Schema: `produto_id`, `nome`, `preco`, `comissao`, `link`, `imagem`, `loja`,
`categoria`, `origem`, `salvo_em`.

Conflitos resolvidos por last-write-wins (`salvo_em`).

Ver [ADR 0007](/docs/decisoes/0007-persistencia-favoritos/).

## Buscas agendadas

### Fluxo completo

```
[Criar busca] → localStorage (manual imediato)
                └─ sync → POST /api/buscas → BigQuery `buscas`
                                                   │
                                  Cloud Scheduler (1 job por busca com cron)
                                                   │
                                  POST /api/coletar?busca_id=X
                                                   │
                                          BigQuery `snapshots`
                                                   │
                              ┌─────────────────────┼──────────────────────┐
                              ▼                     ▼                      ▼
                     /api/estatisticas      Novidades (diff)      Alertas Telegram
                    (tela Estatísticas)    (produtos novos,        (variação >
                                           variações preço)        threshold)
```

### Fontes de dados da busca

| Fonte | Descrição |
|---|---|
| `curadoria` | Ranking padrão por teor |
| `quedas` | Produtos com variação negativa de preço |
| `novos` | Produtos detectados recentemente (janela `dias_janela`) |
| `favoritos` | Produtos favoritados pelo usuário |

### Monitoramento de lojas

`POST /api/lojas` cria automaticamente uma busca com `shop_ids` e cron padrão
(`0 */4 * * *`). Detecção de:
- **Novos produtos** — não existiam na coleta anterior
- **Variações de preço** — quedas e altas significativas (acima do threshold)

### Alertas de preço

Implementados no backend, desabilitados por padrão (aguardando config por usuário).
Quando ativos: variação de preço > threshold → notificação Telegram.

Ver [ADR 0008](/docs/decisoes/0008-alertas-desabilitados/).

## Fluxo de busca de produtos (curadoria)

### Ciclo completo de uma busca

```
Frontend (Svelte)
    │  GET /api/candidatos?keyword=serum&top=20&comissao_min=0.07
    ▼
Cloudflare Worker (proxy)
    │  roteia /api/* → Cloud Run C#
    ▼
API C# (ASP.NET, CompatEndpoints.cs)
    │  1. Recebe keyword, monta FetchRequest
    │  2. Chama collector via gRPC: collector.FetchAsync({keyword, limit})
    ▼
Collector Go (gRPC sidecar, porta 50051 na mesma instância Cloud Run)
    │  1. Resolve marketplace (default: Shopee)
    │  2. Pega o ProductSource via Pipeline
    │  3. Chama ShopeeAdapter.Search(keyword, limit)
    ▼
Shopee Affiliate GraphQL API
    │  POST https://open-api.affiliate.shopee.com.br/graphql
    │  Auth: SHA256(AppId + Timestamp + Body + Secret)
    │  Query: productOfferV2(keyword, limit, sortType)
    ▼
Shopee responde com N produtos
    │  (nome, preço, comissão, vendas, rating, imagem, link afiliado)
    ▼  (volta pelo mesmo caminho)
Collector Go
    │  Mapeia Shopee nodes → domain.Product → proto Product
    │  Retorna FetchResponse {products[], total_found}
    ▼
API C#
    │  1. Proto Product → ProductCandidate (ProductMappings.ToCandidate)
    │  2. ScoringService.Rank(candidates, filter, top):
    │     - Filtra elegíveis (comissão ≥ 7%, vendas ≥ 0)
    │     - Score = 0.45×norm(comissão) + 0.35×norm(EV) + 0.20×norm(rating)
    │       onde EV = comissão × preço × vendas
    │     - Detecta "suspeitos" (comissão alta + vendas 0)
    │  3. Retorna JSON {estrategia, candidatos[], total_bruto}
    ▼
Frontend
    │  Renderiza ProductCards com score, preço, comissão, botão Publicar
```

### Persistência por tipo de dado

| Dado | Onde | Quando |
|------|------|--------|
| Resultado da busca (candidatos) | **Nenhum lugar** | A busca é real-time pass-through — não salva |
| Perfil de busca (keywords, filtros) | **PostgreSQL** (tabela `Buscas`) | Quando o usuário salva uma busca |
| Snapshot de mercado (top produtos) | **BigQuery** (tabela `snapshots`) | Scheduler executa coletas agendadas (cron) |
| Publicações enviadas | **PostgreSQL** (tabela `Publicacoes`) | Ao clicar "Enviar" |
| Favoritos | **PostgreSQL** (tabela `Favoritos`) | Ao clicar ★ |
| Destinos (canais Telegram/WhatsApp) | **PostgreSQL** (tabela `Destinos`) | Configuração em /canais |
| Conversões reais (vendas) | **BigQuery** (tabela `conversoes`) | Scheduler consulta Shopee periodicamente |
| Alertas de preço | **BigQuery** (tabela `snapshots`) → diff | Scheduler compara snapshots |

### Separação de dados

- **PostgreSQL (Neon)**: dados transacionais do app — o que o usuário configura e faz (buscas, destinos, publicações, favoritos, tenant config)
- **BigQuery**: dados analíticos — o que acontece no mercado (snapshots periódicos, conversões, evolução de preço). Somente escrita pelos services Go; leitura pelo Analyzer Python.

A busca (`/api/candidatos`) é **stateless** — chama a Shopee em tempo real via collector e não persiste nada. A persistência só acontece quando o usuário toma uma ação (publicar, favoritar, salvar busca) ou quando o scheduler executa uma coleta agendada.
