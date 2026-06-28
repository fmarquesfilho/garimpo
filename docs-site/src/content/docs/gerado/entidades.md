---
title: Modelo de dados (ER)
description: Diagrama entidade-relacionamento gerado do schema BigQuery.
---

:::caution[Arquivo gerado]
Não edite manualmente. Rode `make docs-er` para regenerar.
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
        STRING estrategia
        FLOAT64 comissao_total
        STRING status
        TIMESTAMP clique_em
        TIMESTAMP compra_em
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
