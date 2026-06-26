# Spec: Rastreamento de Conversões — Fechar o Ciclo

## Problema

Hoje o Garimpei publica ofertas com links de afiliado (via `generateShortLink` da Shopee com `subIds`). Mas não há feedback sobre **quais publicações geraram vendas**. A aba "Desempenho" nas publicações mostra volume de envio por canal — não conversões reais.

Sem isso:
- Mileny não sabe quais produtos convertem
- Não há como comparar performance entre destinos (Telegram vs WhatsApp)
- A estratégia de curadoria não aprende com resultados reais
- A página de Estatísticas não tem métricas de receita

## O que a API da Shopee oferece

### `conversionReport` (Query GraphQL)
Retorna conversões (pedidos realizados via links de afiliado):
- `purchaseTime`, `clickTime`
- `conversionId`
- `totalCommission`
- `utmContent` — **é aqui que o subId volta** (ex: `telegram_nicho_20260622`)
- `orders.items[].itemId`, `itemName`, `itemTotalCommission`
- Status: `UNPAID` → `PENDING` → `COMPLETED` → `CANCELLED`
- Paginação via `scrollId` (válido por 30s)

### `validatedReport` (Query GraphQL)
Conversões **validadas** (após período de devolução):
- Valor final de comissão (inclui `refundAmount`)
- É o que confirma antes do pagamento

## Fluxo proposto

```
Publicar oferta (com subId: canal_estrategia_data)
    ↓
Shopee tracking (7-30 dias)
    ↓
Poll periódico: GET conversionReport (1x/dia via Cloud Scheduler)
    ↓
Cruza subId com publicações no BigQuery
    ↓
Atualiza status da publicação: "converteu" + valor real
    ↓
Dashboard de performance (por destino, por estratégia, por produto)
```

## Implementação

### Fase 1: Poll + armazenamento (MVP)
1. Novo endpoint interno `POST /api/conversoes/sync` (protegido por COLETA_TOKEN)
2. Cloud Scheduler dispara 1x/dia
3. Chama `conversionReport` com filtro dos últimos 7 dias
4. Grava na tabela `conversoes` do BigQuery
5. Cruza `utmContent` com `sub_id` das publicações

### Fase 2: Dashboard
1. Endpoint `GET /api/conversoes/performance` — agrupado por destino/estratégia
2. Frontend: aba "Desempenho" com dados reais (receita por canal, por produto)
3. Métrica-rainha: **receita por publicação** e **ROI por destino**

### Fase 3: Feedback loop
1. Produtos que convertem ganham peso no scoring futuro
2. Destinos com melhor ROI são sugeridos como padrão
3. Alerta quando uma publicação converte (Telegram)

## Schema BigQuery (nova tabela ou expandir existente)

```sql
CREATE TABLE IF NOT EXISTS `garimpo.conversoes_reais` (
  conversion_id   STRING,
  sub_id          STRING,    -- canal_estrategia_data (de utmContent)
  produto_id      STRING,
  nome_produto    STRING,
  comissao_total  FLOAT64,
  status          STRING,    -- PENDING | COMPLETED | CANCELLED
  clique_em       TIMESTAMP,
  compra_em       TIMESTAMP,
  sincronizado_em TIMESTAMP
)
PARTITION BY DATE(compra_em);
```

## Riscos
- A API de conversões pode ter delay de 24-48h (normal da Shopee)
- `scrollId` expira em 30s — precisa paginação rápida
- Volume baixo no início (poucas publicações) — resultados demoram a aparecer

## Decisões a tomar
- [ ] Frequência do poll: 1x/dia? 2x/dia?
- [ ] Período de lookback: últimos 7 dias? 30 dias?
- [ ] Onde mostrar: aba em Publicações? Página separada? Na Oportunidades?
- [ ] Notificar conversão via Telegram (bot de alertas)?
