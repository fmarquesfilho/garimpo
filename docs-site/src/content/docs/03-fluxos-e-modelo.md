---
title: "Fluxos e modelo de dados"
---


## Entidades principais

O diagrama ER completo é gerado automaticamente — ver `docs/gerado/ENTIDADES.md`.

### Tabelas BigQuery (dataset `garimpo`)

| Tabela | O que armazena | Partição |
|---|---|---|
| `eventos` | Seleções e publicações de curadoria | `DATE(em)` |
| `snapshots` | Foto periódica dos top N de uma keyword/categoria | `DATE(coletado_em)` |
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
