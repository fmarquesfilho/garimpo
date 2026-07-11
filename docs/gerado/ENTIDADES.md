---
title: Modelo de dados (ER)
description: Diagrama entidade-relacionamento gerado do schema BigQuery.
---

:::caution[Arquivo gerado]
Não edite manualmente. Rode `mise run docs:er` para regenerar.
:::

```mermaid
erDiagram
    EVENTOS {
        STRING tipo
        STRING produto_id
        STRING nome
        STRING categoria
        STRING estrategia
        STRING canal
        STRING sub_id
        FLOAT64 comissao
        FLOAT64 preco
        INT64 vendas
        FLOAT64 score
        TIMESTAMP em
    }

    SNAPSHOTS {
        TIMESTAMP coletado_em
        STRING busca_id
        STRING categoria
        STRING keyword
        STRING estrategia
        INT64 posicao
        STRING produto_id
        STRING nome
        FLOAT64 preco
        FLOAT64 comissao
        INT64 vendas
        FLOAT64 nota
        FLOAT64 score
    }

    BUSCAS {
        STRING id
        STRING keywords
        STRING shop_ids
        STRING categoria
        STRING estrategia
        FLOAT64 comissao_min
        INT64 vendas_min
        FLOAT64 nota_min
        INT64 top
        STRING cron
        BOOL ativo
        STRING owner_uid
        STRING rotation_cursor
        STRING full_scan_at
        TIMESTAMP salvo_em
    }

    CONVERSOES {
        STRING conversion_id
        STRING produto_id
        STRING nome
        STRING canal
        STRING estrategia
        FLOAT64 comissao_total
        FLOAT64 preco
        STRING status
        TIMESTAMP clique_em
        TIMESTAMP compra_em
        TIMESTAMP convertido_em
    }

    DESTINOS {
        STRING id
        STRING nome
        STRING tipo
        STRING config
        BOOL ativo
        TIMESTAMP salvo_em
    }

    TEMPLATES {
        STRING id
        STRING nome
        STRING corpo
        BOOL com_foto
        BOOL ativo
        TIMESTAMP salvo_em
    }

    PUBLICACOES {
        STRING id
        STRING produto_id
        STRING nome
        STRING categoria
        FLOAT64 preco
        FLOAT64 comissao
        STRING link
        STRING imagem
        STRING estrategia
        STRING destino_id
        STRING template_id
        STRING agendada_em
        STRING status
        STRING detalhe
        TIMESTAMP criada_em
        STRING enviada_em
        STRING owner_uid
    }

    FAVORITOS {
        STRING produto_id
        STRING nome
        FLOAT64 preco
        FLOAT64 comissao
        STRING link
        STRING imagem
        STRING loja
        STRING categoria
        STRING origem
        BOOL ativo
        STRING owner_uid
        TIMESTAMP salvo_em
    }

    COUPON_SNAPSHOTS {
        STRING coupon_id
        STRING marketplace
        STRING code
        STRING discount_type
        FLOAT64 discount_value
        FLOAT64 min_spend
        TIMESTAMP start_time
        TIMESTAMP end_time
        STRING applicable_categories
        STRING status
        STRING detection_status
        STRING owner_uid
        TIMESTAMP collected_at
    }

    BUSCAS ||--o{ SNAPSHOTS : "gera coletas"
    SNAPSHOTS ||--o{ EVENTOS : "produto selecionado"
    EVENTOS ||--o{ CONVERSOES : "gera conversão"
```

## Particionamento

| Tabela | Partição |
|---|---|
| `eventos` | `DATE(em)` |
| `snapshots` | `DATE(coletado_em)` |
| `buscas` | `DATE(salvo_em)` |
| `conversoes` | `DATE(compra_em)` |
